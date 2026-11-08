package blobdata

import (
	"errors"
	"io"
	"os"
	"strings"

	"railkeeper/backend/internal/backup"
	"railkeeper/backend/internal/media"
)

type LegacyFiles struct{ root string }

func NewLegacyFiles(root string) *LegacyFiles { return &LegacyFiles{root: root} }
func (f *LegacyFiles) Read(path string) ([]byte, error) {
	data, err := ReadConfinedDataFile(f.root, path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, media.ErrSourceMissing
	}
	return data, err
}

// Normalize only the read path; stored historical Windows text remains unchanged.
func portableMediaPath(value string) (string, error) {
	value = strings.ReplaceAll(value, `\`, "/")
	if err := backup.ValidateBackupFilePath(value); err != nil {
		return "", err
	}
	return value, nil
}
func ReadConfinedDataFile(dataDir, relativePath string) ([]byte, error) {
	relativePath, err := portableMediaPath(relativePath)
	if err != nil {
		return nil, err
	}
	root, err := os.OpenRoot(dataDir)
	if err != nil {
		return nil, err
	}
	defer func() { _ = root.Close() }()
	info, err := root.Lstat("uploads")
	if err != nil {
		return nil, err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil, backup.ErrBackupPath
	}
	uploads, err := root.OpenRoot("uploads")
	if err != nil {
		return nil, err
	}
	defer func() { _ = uploads.Close() }()
	file, err := uploads.Open(strings.TrimPrefix(relativePath, "uploads/"))
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, errors.New("file blob source is empty")
	}
	return data, nil
}
