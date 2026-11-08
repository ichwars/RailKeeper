package backupdata

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"railkeeper/backend/internal/backup"
	"railkeeper/backend/internal/credentials"
)

func (s *Store) ExportFiles() ([]BackupFile, error) {
	uploadsDir := filepath.Join(s.dataDir, "uploads")
	if _, err := os.Stat(uploadsDir); errors.Is(err, os.ErrNotExist) {
		return []BackupFile{}, nil
	}

	files := []BackupFile{}
	if err := filepath.WalkDir(uploadsDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(s.dataDir, path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		if err := backup.ValidateBackupFilePath(relative); err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(data)
		files = append(files, BackupFile{
			Path:          relative,
			SizeBytes:     int64(len(data)),
			SHA256:        hex.EncodeToString(sum[:]),
			ContentBase64: base64.StdEncoding.EncodeToString(data),
		})
		return nil
	}); err != nil {
		return nil, fmt.Errorf("export backup files: %w", err)
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return files, nil
}

type stagedRestoreFiles struct {
	root          string
	uploadsDir    string
	restoredFiles int
}

func (s *Store) stageRestoreFiles(files []BackupFile) (*stagedRestoreFiles, error) {
	if err := os.MkdirAll(s.dataDir, 0o755); err != nil {
		return nil, fmt.Errorf("prepare restore data directory: %w", err)
	}
	root, err := os.MkdirTemp(s.dataDir, ".restore-staging-")
	if err != nil {
		return nil, fmt.Errorf("create restore staging directory: %w", err)
	}
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.RemoveAll(root)
		}
	}()
	uploadsDir := filepath.Join(root, "uploads")
	if err := os.MkdirAll(uploadsDir, 0o755); err != nil {
		return nil, fmt.Errorf("create restore staging uploads: %w", err)
	}
	uploads, err := os.OpenRoot(uploadsDir)
	if err != nil {
		return nil, fmt.Errorf("open restore staging uploads: %w", err)
	}
	defer func() { _ = uploads.Close() }()

	staged := &stagedRestoreFiles{root: root, uploadsDir: uploadsDir}
	for _, file := range files {
		if err := writeStagedRestoreFile(uploads, file); err != nil {
			return nil, err
		}
		staged.restoredFiles++
	}
	cleanup = false
	return staged, nil
}

func writeStagedRestoreFile(uploads *os.Root, file BackupFile) error {
	if err := backup.ValidateBackupFilePath(file.Path); err != nil {
		return err
	}
	data, err := base64.StdEncoding.DecodeString(file.ContentBase64)
	if err != nil {
		return ErrBackupInvalid
	}
	sum := sha256.Sum256(data)
	if file.SHA256 != "" && !strings.EqualFold(file.SHA256, hex.EncodeToString(sum[:])) {
		return ErrBackupInvalid
	}
	relative := strings.TrimPrefix(backup.PathClean(file.Path), "uploads/")
	// Keep both directory creation and file writes confined even if a symlink changes.
	if err := uploads.MkdirAll(filepath.Dir(relative), 0o755); err != nil {
		return fmt.Errorf("create restore staging directory: %w", err)
	}
	if err := uploads.WriteFile(relative, data, 0o600); err != nil {
		return fmt.Errorf("stage restore file: %w", err)
	}
	return nil
}

type uploadsRestoreSwap struct {
	uploadsDir string
	backupDir  string
	hadUploads bool
	replaced   bool
}

func (s *Store) replaceUploadsWithStaged(staged *stagedRestoreFiles) (*uploadsRestoreSwap, error) {
	uploadsDir := filepath.Join(s.dataDir, "uploads")
	swap := &uploadsRestoreSwap{
		uploadsDir: uploadsDir,
		backupDir:  filepath.Join(s.dataDir, ".restore-uploads-backup-"+credentials.RandomID()),
	}

	if _, err := os.Stat(uploadsDir); err == nil {
		swap.hadUploads = true
		if err := os.Rename(uploadsDir, swap.backupDir); err != nil {
			return nil, fmt.Errorf("stage current uploads: %w", err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("inspect uploads directory: %w", err)
	}

	if err := os.Rename(staged.uploadsDir, uploadsDir); err != nil {
		if swap.hadUploads {
			_ = os.Rename(swap.backupDir, uploadsDir)
		}
		return nil, fmt.Errorf("activate restored uploads: %w", err)
	}
	swap.replaced = true
	return swap, nil
}

func (s *uploadsRestoreSwap) rollback() {
	if s == nil || !s.replaced {
		return
	}
	_ = os.RemoveAll(s.uploadsDir)
	if s.hadUploads {
		_ = os.Rename(s.backupDir, s.uploadsDir)
	}
}

func (s *uploadsRestoreSwap) cleanup() {
	if s == nil || !s.hadUploads {
		return
	}
	_ = os.RemoveAll(s.backupDir)
}
