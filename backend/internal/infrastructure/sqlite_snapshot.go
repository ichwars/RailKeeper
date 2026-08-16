package infrastructure

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type sqliteArtifactDigest struct {
	name   string
	digest [sha256.Size]byte
}

func CreateSQLiteSnapshot(ctx context.Context, source *sql.DB, targetPath string) error {
	if source == nil {
		return errors.New("snapshot source database is required")
	}
	if targetPath == "" {
		return errors.New("snapshot target path is required")
	}
	if _, err := os.Stat(targetPath); err == nil {
		return fmt.Errorf("snapshot target already exists: %s", targetPath)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect snapshot target: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o700); err != nil {
		return fmt.Errorf("create snapshot directory: %w", err)
	}
	if _, err := source.ExecContext(ctx, `PRAGMA synchronous=FULL`); err != nil {
		return fmt.Errorf("configure snapshot durability: %w", err)
	}
	if _, err := source.ExecContext(ctx, `VACUUM INTO ?`, targetPath); err != nil {
		_ = os.Remove(targetPath)
		return fmt.Errorf("create SQLite snapshot: %w", err)
	}
	if err := os.Chmod(targetPath, 0o600); err != nil {
		_ = os.Remove(targetPath)
		return fmt.Errorf("secure SQLite snapshot: %w", err)
	}
	if err := ValidateSQLiteSnapshot(ctx, targetPath); err != nil {
		_ = os.Remove(targetPath)
		return err
	}
	return nil
}

func CreateSQLiteSnapshotFromPath(ctx context.Context, sourcePath, targetPath string) error {
	if sourcePath == "" {
		return errors.New("snapshot source path is required")
	}
	workingDir, err := os.MkdirTemp("", ".railkeeper-snapshot-source-")
	if err != nil {
		return fmt.Errorf("create snapshot working directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(workingDir) }()

	before, err := sqliteArtifactManifest(sourcePath)
	if err != nil {
		return fmt.Errorf("inspect snapshot source: %w", err)
	}
	workingPath := filepath.Join(workingDir, "railkeeper.db")
	if err := copySQLiteArtifact(sourcePath, workingPath); err != nil {
		return fmt.Errorf("copy snapshot source database: %w", err)
	}
	walPath := sourcePath + "-wal"
	if artifactDigestExists(before, filepath.Base(walPath)) {
		if err := copySQLiteArtifact(walPath, workingPath+"-wal"); err != nil {
			return fmt.Errorf("copy snapshot source WAL: %w", err)
		}
	}
	after, err := sqliteArtifactManifest(sourcePath)
	if err != nil {
		return fmt.Errorf("verify snapshot source: %w", err)
	}
	if !slices.Equal(before, after) {
		return errors.New("snapshot source changed while creating working copy")
	}

	source, err := sql.Open("sqlite", readWriteSQLiteDSN(workingPath))
	if err != nil {
		return fmt.Errorf("open snapshot working copy: %w", err)
	}
	source.SetMaxOpenConns(1)
	defer func() { _ = source.Close() }()
	if err := source.PingContext(ctx); err != nil {
		return fmt.Errorf("ping snapshot working copy: %w", err)
	}
	return CreateSQLiteSnapshot(ctx, source, targetPath)
}

func ValidateSQLiteSnapshot(ctx context.Context, snapshotPath string) error {
	if snapshotPath == "" {
		return errors.New("snapshot path is required")
	}
	db, err := sql.Open("sqlite", readOnlySQLiteDSN(snapshotPath))
	if err != nil {
		return fmt.Errorf("open SQLite snapshot: %w", err)
	}
	db.SetMaxOpenConns(1)
	defer func() { _ = db.Close() }()

	rows, err := db.QueryContext(ctx, `PRAGMA integrity_check`)
	if err != nil {
		return fmt.Errorf("validate SQLite snapshot: %w", err)
	}
	defer func() { _ = rows.Close() }()
	results := []string{}
	for rows.Next() {
		var result string
		if err := rows.Scan(&result); err != nil {
			return fmt.Errorf("read SQLite integrity result: %w", err)
		}
		results = append(results, result)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate SQLite integrity result: %w", err)
	}
	if len(results) != 1 || results[0] != "ok" {
		return fmt.Errorf("SQLite snapshot integrity check failed: %v", results)
	}
	return nil
}

func SQLiteSnapshotsEquivalent(ctx context.Context, leftPath, rightPath, tempDir string) (bool, error) {
	comparisonDir, err := os.MkdirTemp(tempDir, ".railkeeper-db-compare-")
	if err != nil {
		return false, fmt.Errorf("create database comparison directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(comparisonDir) }()

	leftSnapshot := filepath.Join(comparisonDir, "left.db")
	rightSnapshot := filepath.Join(comparisonDir, "right.db")
	if err := CreateSQLiteSnapshotFromPath(ctx, leftPath, leftSnapshot); err != nil {
		return false, fmt.Errorf("snapshot left database: %w", err)
	}
	if err := CreateSQLiteSnapshotFromPath(ctx, rightPath, rightSnapshot); err != nil {
		return false, fmt.Errorf("snapshot right database: %w", err)
	}
	leftHash, err := normalizedSQLiteFileSHA256(leftSnapshot)
	if err != nil {
		return false, err
	}
	rightHash, err := normalizedSQLiteFileSHA256(rightSnapshot)
	if err != nil {
		return false, err
	}
	return leftHash == rightHash, nil
}

func readOnlySQLiteDSN(path string) string {
	return sqliteFileDSN(path, "ro")
}

func readWriteSQLiteDSN(path string) string {
	return sqliteFileDSN(path, "rw")
}

func sqliteFileDSN(path, mode string) string {
	absolute, err := filepath.Abs(path)
	if err != nil {
		absolute = path
	}
	slashPath := filepath.ToSlash(absolute)
	if filepath.VolumeName(absolute) != "" && !strings.HasPrefix(slashPath, "/") {
		slashPath = "/" + slashPath
	}
	location := &url.URL{Scheme: "file", Path: slashPath}
	query := location.Query()
	query.Set("mode", mode)
	query.Set("_foreign_keys", "on")
	location.RawQuery = query.Encode()
	return location.String()
}

func sqliteArtifactManifest(sourcePath string) ([]sqliteArtifactDigest, error) {
	artifacts := []string{sourcePath, sourcePath + "-wal", sourcePath + "-shm"}
	manifest := make([]sqliteArtifactDigest, 0, len(artifacts))
	for _, path := range artifacts {
		info, err := os.Stat(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		if !info.Mode().IsRegular() {
			return nil, fmt.Errorf("SQLite artifact is not a regular file: %s", path)
		}
		digest, err := fileSHA256(path)
		if err != nil {
			return nil, err
		}
		manifest = append(manifest, sqliteArtifactDigest{name: filepath.Base(path), digest: digest})
	}
	if len(manifest) == 0 || manifest[0].name != filepath.Base(sourcePath) {
		return nil, fmt.Errorf("snapshot source database does not exist: %s", sourcePath)
	}
	return manifest, nil
}

func artifactDigestExists(manifest []sqliteArtifactDigest, name string) bool {
	return slices.ContainsFunc(manifest, func(artifact sqliteArtifactDigest) bool {
		return artifact.name == name
	})
}

func copySQLiteArtifact(sourcePath, targetPath string) error {
	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer func() { _ = source.Close() }()
	target, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	if _, err := io.Copy(target, source); err != nil {
		_ = target.Close()
		return err
	}
	if err := target.Sync(); err != nil {
		_ = target.Close()
		return err
	}
	return target.Close()
}

func fileSHA256(path string) ([sha256.Size]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return [sha256.Size]byte{}, fmt.Errorf("read file for checksum: %w", err)
	}
	return sha256.Sum256(data), nil
}

func normalizedSQLiteFileSHA256(path string) ([sha256.Size]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return [sha256.Size]byte{}, fmt.Errorf("read SQLite file for checksum: %w", err)
	}
	if len(data) < 100 || string(data[:16]) != "SQLite format 3\x00" {
		return [sha256.Size]byte{}, fmt.Errorf("invalid SQLite header: %s", path)
	}
	// VACUUM INTO may advance these administrative counters even when schema and
	// user data are unchanged. The schema b-tree and all content pages remain hashed.
	for _, bounds := range [][2]int{{24, 28}, {40, 44}, {92, 96}} {
		clear(data[bounds[0]:bounds[1]])
	}
	return sha256.Sum256(data), nil
}
