package infrastructure

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

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
	source, err := sql.Open("sqlite", readOnlySQLiteDSN(sourcePath))
	if err != nil {
		return fmt.Errorf("open snapshot source: %w", err)
	}
	source.SetMaxOpenConns(1)
	defer func() { _ = source.Close() }()
	if err := source.PingContext(ctx); err != nil {
		return fmt.Errorf("ping snapshot source: %w", err)
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
	leftHash, err := fileSHA256(leftSnapshot)
	if err != nil {
		return false, err
	}
	rightHash, err := fileSHA256(rightSnapshot)
	if err != nil {
		return false, err
	}
	return leftHash == rightHash, nil
}

func readOnlySQLiteDSN(path string) string {
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
	query.Set("mode", "ro")
	query.Set("_foreign_keys", "on")
	location.RawQuery = query.Encode()
	return location.String()
}

func fileSHA256(path string) ([sha256.Size]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return [sha256.Size]byte{}, fmt.Errorf("read file for checksum: %w", err)
	}
	return sha256.Sum256(data), nil
}
