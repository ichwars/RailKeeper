package blobdata

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"railkeeper/backend/internal/media"
)

type Store struct{ db *sql.DB }

func New(db *sql.DB) *Store { return &Store{db: db} }

type Executor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

func (s *Store) Insert(ctx context.Context, blob media.Blob) error { return Insert(ctx, s.db, blob) }

// Insert also supports caller-owned decoder snapshot transactions.
func Insert(ctx context.Context, writer Executor, blob media.Blob) error {
	if _, err := writer.ExecContext(ctx, `
INSERT INTO file_blobs(id, original_size, compressed_size, compression, sha256, data, created_at)
VALUES(?, ?, ?, ?, ?, ?, ?)
`, blob.ID, blob.OriginalSize, blob.CompressedSize, blob.Compression, blob.SHA256, blob.Data, blob.CreatedAt); err != nil {
		return fmt.Errorf("store file blob: %w", err)
	}
	return nil
}

func (s *Store) Read(ctx context.Context, id string) (media.StoredBlob, error) {
	var compression string
	var compressed []byte
	if err := s.db.QueryRowContext(ctx, `
SELECT compression, data
FROM file_blobs
WHERE id=?
`, id).Scan(&compression, &compressed); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return media.StoredBlob{}, media.ErrBlobNotFound
		}
		return media.StoredBlob{}, err
	}
	return media.StoredBlob{Compression: compression, Data: compressed}, nil
}
func (s *Store) DeleteIfUnreferenced(ctx context.Context, id string) error {
	if _, err := s.db.ExecContext(ctx, `
DELETE FROM file_blobs WHERE id=?
AND NOT EXISTS (SELECT 1 FROM vehicle_images WHERE blob_id=? OR thumbnail_blob_id=?)
AND NOT EXISTS (SELECT 1 FROM vehicle_attachments WHERE blob_id=?)
AND NOT EXISTS (SELECT 1 FROM vehicle_cv_files WHERE blob_id=?)
AND NOT EXISTS (SELECT 1 FROM accessory_documents WHERE file_blob_id=?)
AND NOT EXISTS (SELECT 1 FROM vehicle_sets WHERE set_image_blob_id=? OR set_image_thumbnail_blob_id=?)
AND NOT EXISTS (SELECT 1 FROM decoder_cv_snapshot_sources WHERE blob_id=?)
`, id, id, id, id, id, id, id, id, id); err != nil {
		return fmt.Errorf("delete unreferenced file blob: %w", err)
	}
	return nil
}

func targetColumns(target media.LegacyTarget) (table, blobColumn, pathColumn string, err error) {
	switch target {
	case media.ImageOriginal:
		return "vehicle_images", "blob_id", "storage_path", nil
	case media.ImageThumbnail:
		return "vehicle_images", "thumbnail_blob_id", "thumbnail_path", nil
	case media.VehicleAttachment:
		return "vehicle_attachments", "blob_id", "storage_path", nil
	case media.VehicleCVSource:
		return "vehicle_cv_files", "blob_id", "storage_path", nil
	default:
		return "", "", "", errors.New("unsupported legacy file-blob target")
	}
}
func quoteIdentifier(value string) string { return `"` + strings.ReplaceAll(value, `"`, `""`) + `"` }
func (s *Store) PendingLegacy(ctx context.Context, target media.LegacyTarget) ([]media.LegacyPath, error) {
	table, blobColumn, pathColumn, err := targetColumns(target)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, fmt.Sprintf(`
SELECT id, COALESCE(%s, '')
FROM %s
WHERE COALESCE(%s, '')='' AND COALESCE(%s, '')<>''
`, quoteIdentifier(pathColumn), quoteIdentifier(table), quoteIdentifier(blobColumn), quoteIdentifier(pathColumn)))
	if err != nil {
		return nil, fmt.Errorf("list filesystem blobs for %s: %w", table, err)
	}
	defer func() { _ = rows.Close() }()

	pending := []media.LegacyPath{}
	for rows.Next() {
		var item media.LegacyPath
		if err := rows.Scan(&item.RowID, &item.Path); err != nil {
			return nil, fmt.Errorf("scan filesystem blob for %s: %w", table, err)
		}
		pending = append(pending, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate filesystem blobs for %s: %w", table, err)
	}

	return pending, nil
}
func (s *Store) BindLegacy(ctx context.Context, target media.LegacyTarget, rowID, blobID string) error {
	table, blobColumn, _, err := targetColumns(target)
	if err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, fmt.Sprintf(
		`UPDATE %s SET %s=? WHERE id=?`,
		quoteIdentifier(table),
		quoteIdentifier(blobColumn),
	), blobID, rowID); err != nil {
		return fmt.Errorf("update blob reference for %s: %w", table, err)
	}
	return nil
}
