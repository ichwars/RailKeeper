package application

import (
	"bytes"
	"compress/zlib"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FileBlobService struct {
	db      *sql.DB
	dataDir string
}

type FileBlob struct {
	ID             string
	OriginalSize   int64
	CompressedSize int64
	Compression    string
	SHA256         string
	Data           []byte
	CreatedAt      string
}

func NewFileBlobService(db *sql.DB, dataDir string) *FileBlobService {
	return &FileBlobService{db: db, dataDir: dataDir}
}

func (s *FileBlobService) Store(ctx context.Context, data []byte) (string, error) {
	if s == nil || s.db == nil {
		return "", errors.New("file blob service is not configured")
	}
	if len(data) == 0 {
		return "", errors.New("file blob is empty")
	}
	compressed, err := compressBlobData(data)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	blob := FileBlob{
		ID:             randomID(),
		OriginalSize:   int64(len(data)),
		CompressedSize: int64(len(compressed)),
		Compression:    "zlib",
		SHA256:         hex.EncodeToString(sum[:]),
		Data:           compressed,
		CreatedAt:      time.Now().UTC().Format(time.RFC3339),
	}
	if _, err := s.db.ExecContext(ctx, `
INSERT INTO file_blobs(id, original_size, compressed_size, compression, sha256, data, created_at)
VALUES(?, ?, ?, ?, ?, ?, ?)
`, blob.ID, blob.OriginalSize, blob.CompressedSize, blob.Compression, blob.SHA256, blob.Data, blob.CreatedAt); err != nil {
		return "", fmt.Errorf("store file blob: %w", err)
	}
	return blob.ID, nil
}

func (s *FileBlobService) Load(ctx context.Context, id string) ([]byte, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("file blob service is not configured")
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, sql.ErrNoRows
	}
	var compression string
	var compressed []byte
	if err := s.db.QueryRowContext(ctx, `
SELECT compression, data
FROM file_blobs
WHERE id=?
`, id).Scan(&compression, &compressed); err != nil {
		return nil, err
	}
	data, err := decompressBlobData(compressed, compression)
	if err != nil {
		return nil, fmt.Errorf("load file blob: %w", err)
	}
	return data, nil
}

func (s *FileBlobService) DeleteIfUnreferenced(ctx context.Context, id string) error {
	if s == nil || s.db == nil {
		return nil
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return nil
	}
	var count int
	if err := s.db.QueryRowContext(ctx, `
SELECT
  (SELECT COUNT(*) FROM vehicle_images WHERE blob_id=? OR thumbnail_blob_id=?) +
  (SELECT COUNT(*) FROM vehicle_attachments WHERE blob_id=?) +
  (SELECT COUNT(*) FROM vehicle_cv_files WHERE blob_id=?)
`, id, id, id, id).Scan(&count); err != nil {
		return fmt.Errorf("count file blob references: %w", err)
	}
	if count > 0 {
		return nil
	}
	if _, err := s.db.ExecContext(ctx, `DELETE FROM file_blobs WHERE id=?`, id); err != nil {
		return fmt.Errorf("delete file blob: %w", err)
	}
	return nil
}

func (s *FileBlobService) MigrateFilesystemBlobs(ctx context.Context) error {
	if s == nil || s.db == nil || strings.TrimSpace(s.dataDir) == "" {
		return nil
	}
	paths, err := s.migrateTablePaths(ctx, "vehicle_images", "blob_id", "storage_path")
	if err != nil {
		return err
	}
	more, err := s.migrateTablePaths(ctx, "vehicle_images", "thumbnail_blob_id", "thumbnail_path")
	if err != nil {
		return err
	}
	paths = append(paths, more...)
	more, err = s.migrateTablePaths(ctx, "vehicle_attachments", "blob_id", "storage_path")
	if err != nil {
		return err
	}
	paths = append(paths, more...)
	more, err = s.migrateTablePaths(ctx, "vehicle_cv_files", "blob_id", "storage_path")
	if err != nil {
		return err
	}
	paths = append(paths, more...)

	for _, relativePath := range uniqueStrings(paths) {
		if s.filesystemPathStillReferenced(ctx, relativePath) {
			continue
		}
		if fullPath, err := confinedDataPath(s.dataDir, relativePath); err == nil {
			_ = os.Remove(fullPath)
		}
	}
	if err := s.cleanupUnneededUploadFiles(ctx); err != nil {
		return err
	}
	return nil
}

func (s *FileBlobService) migrateTablePaths(ctx context.Context, table, blobColumn, pathColumn string) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, fmt.Sprintf(`
SELECT id, COALESCE(%s, '')
FROM %s
WHERE COALESCE(%s, '')='' AND COALESCE(%s, '')<>''
`, quoteIdentifier(pathColumn), quoteIdentifier(table), quoteIdentifier(blobColumn), quoteIdentifier(pathColumn)))
	if err != nil {
		return nil, fmt.Errorf("list filesystem blobs for %s: %w", table, err)
	}
	defer func() { _ = rows.Close() }()

	type pendingBlob struct {
		rowID string
		path  string
	}
	pending := []pendingBlob{}
	for rows.Next() {
		var item pendingBlob
		if err := rows.Scan(&item.rowID, &item.path); err != nil {
			return nil, fmt.Errorf("scan filesystem blob for %s: %w", table, err)
		}
		pending = append(pending, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate filesystem blobs for %s: %w", table, err)
	}

	migratedPaths := []string{}
	blobByPath := map[string]string{}
	for _, item := range pending {
		blobID := blobByPath[item.path]
		if blobID == "" {
			data, err := s.readConfinedDataFile(item.path)
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			if err != nil {
				return nil, err
			}
			blobID, err = s.Store(ctx, data)
			if err != nil {
				return nil, err
			}
			blobByPath[item.path] = blobID
			migratedPaths = append(migratedPaths, item.path)
		}
		if _, err := s.db.ExecContext(ctx, fmt.Sprintf(
			`UPDATE %s SET %s=? WHERE id=?`,
			quoteIdentifier(table),
			quoteIdentifier(blobColumn),
		), blobID, item.rowID); err != nil {
			return nil, fmt.Errorf("update blob reference for %s: %w", table, err)
		}
	}
	return migratedPaths, nil
}

func (s *FileBlobService) readConfinedDataFile(relativePath string) ([]byte, error) {
	fullPath, err := confinedDataPath(s.dataDir, relativePath)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, errors.New("file blob source is empty")
	}
	return data, nil
}

func (s *FileBlobService) filesystemPathStillReferenced(ctx context.Context, relativePath string) bool {
	var count int
	if err := s.db.QueryRowContext(ctx, `
SELECT
  (SELECT COUNT(*) FROM vehicle_images WHERE (storage_path=? AND blob_id='') OR (thumbnail_path=? AND thumbnail_blob_id='')) +
  (SELECT COUNT(*) FROM vehicle_attachments WHERE storage_path=? AND blob_id='') +
  (SELECT COUNT(*) FROM vehicle_cv_files WHERE storage_path=? AND blob_id='')
`, relativePath, relativePath, relativePath, relativePath).Scan(&count); err != nil {
		return true
	}
	return count > 0
}

func (s *FileBlobService) cleanupUnneededUploadFiles(ctx context.Context) error {
	uploadsDir := filepath.Join(s.dataDir, "uploads")
	if _, err := os.Stat(uploadsDir); errors.Is(err, os.ErrNotExist) {
		return nil
	}
	dirs := []string{}
	if err := filepath.WalkDir(uploadsDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			dirs = append(dirs, path)
			return nil
		}
		relative, err := filepath.Rel(s.dataDir, path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		if s.filesystemPathStillReferenced(ctx, relative) {
			return nil
		}
		legacyRelative := filepath.FromSlash(relative)
		if legacyRelative != relative && s.filesystemPathStillReferenced(ctx, legacyRelative) {
			return nil
		}
		return os.Remove(path)
	}); err != nil {
		return fmt.Errorf("cleanup migrated upload files: %w", err)
	}
	for i := len(dirs) - 1; i >= 0; i-- {
		if dirs[i] == uploadsDir {
			continue
		}
		_ = os.Remove(dirs[i])
	}
	return nil
}

func compressBlobData(data []byte) ([]byte, error) {
	var buffer bytes.Buffer
	writer := zlib.NewWriter(&buffer)
	if _, err := writer.Write(data); err != nil {
		_ = writer.Close()
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func decompressBlobData(data []byte, compression string) ([]byte, error) {
	switch strings.ToLower(strings.TrimSpace(compression)) {
	case "", "none":
		return data, nil
	case "zlib":
		reader, err := zlib.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		defer func() { _ = reader.Close() }()
		return io.ReadAll(reader)
	default:
		return nil, fmt.Errorf("unsupported blob compression %q", compression)
	}
}

func uniqueStrings(values []string) []string {
	seen := map[string]struct{}{}
	out := []string{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func confinedDataPath(dataDir, relativePath string) (string, error) {
	base, err := filepath.Abs(dataDir)
	if err != nil {
		return "", err
	}
	target := filepath.Join(base, filepath.FromSlash(filepath.ToSlash(relativePath)))
	absTarget, err := filepath.Abs(target)
	if err != nil {
		return "", err
	}
	if absTarget != base && !strings.HasPrefix(absTarget, base+string(os.PathSeparator)) {
		return "", errors.New("path escapes data directory")
	}
	return absTarget, nil
}
