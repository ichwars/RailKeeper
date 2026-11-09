package startup

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func buildLegacyManifest(root string, includeDatabaseArtifacts bool) ([]legacyManifestEntry, error) {
	entries := []legacyManifestEntry{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if relative == "." {
			return nil
		}
		if !pathWithin(root, path) {
			return fmt.Errorf("legacy path escapes data directory: %s", path)
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 || entry.Type()&os.ModeSymlink != 0 ||
			isReparsePoint(info) {
			return fmt.Errorf("legacy data contains unsupported link: %s", relative)
		}
		relative = filepath.ToSlash(relative)
		if !includeDatabaseArtifacts && isSQLiteArtifact(relative) {
			return nil
		}
		if relative == legacyMigrationReceiptName {
			return nil
		}
		if entry.IsDir() {
			entries = append(entries, legacyManifestEntry{Path: relative, Dir: true})
			return nil
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("legacy data contains unsupported file type: %s", relative)
		}
		digest, err := legacyFileSHA256(path)
		if err != nil {
			return err
		}
		entries = append(entries, legacyManifestEntry{
			Path: relative, Size: info.Size(), SHA256: digest,
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("inspect legacy data: %w", err)
	}
	sort.Slice(entries, func(left, right int) bool { return entries[left].Path < entries[right].Path })
	return entries, nil
}

func copyLegacyManifest(
	sourceRoot string,
	targetRoot string,
	manifest []legacyManifestEntry,
	copyFile func(string, string) error,
) error {
	for _, entry := range manifest {
		relative := filepath.FromSlash(entry.Path)
		sourcePath := filepath.Join(sourceRoot, relative)
		targetPath := filepath.Join(targetRoot, relative)
		if !pathWithin(sourceRoot, sourcePath) || !pathWithin(targetRoot, targetPath) {
			return fmt.Errorf("legacy copy path escaped root: %s", entry.Path)
		}
		if entry.Dir {
			if err := os.MkdirAll(targetPath, 0o700); err != nil {
				return fmt.Errorf("create staged directory %s: %w", entry.Path, err)
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(targetPath), 0o700); err != nil {
			return fmt.Errorf("create staged file parent %s: %w", entry.Path, err)
		}
		if err := copyFile(sourcePath, targetPath); err != nil {
			return fmt.Errorf("copy legacy file %s: %w", entry.Path, err)
		}
	}
	return nil
}

func copyLegacyFile(sourcePath, targetPath string) error {
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

func regularFileExists(path string) (bool, error) {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if info.Mode()&os.ModeSymlink != 0 || isReparsePoint(info) {
		return false, fmt.Errorf("database path is a link or reparse point: %s", path)
	}
	if !info.Mode().IsRegular() {
		return false, fmt.Errorf("not a regular file: %s", path)
	}
	return true, nil
}

func directoryHasEntries(path string) (bool, error) {
	entries, err := os.ReadDir(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return len(entries) > 0, nil
}

func removeEmptyTarget(path string) error {
	entries, err := os.ReadDir(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("inspect migration target: %w", err)
	}
	if len(entries) != 0 {
		return errors.New("migration target became non-empty before activation")
	}
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("remove empty migration target: %w", err)
	}
	return nil
}

func legacyConflict(safePath, legacyPath, reason string) LegacyMigrationResult {
	return LegacyMigrationResult{
		Status:  LegacyConflict,
		DataDir: safePath,
		Conflict: &LegacyConflictInfo{
			SafePath: safePath, LegacyPath: legacyPath, Reason: reason,
		},
	}
}

func randomLegacySuffix() string {
	var bytes [12]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes[:])
}

func isSQLiteArtifact(relative string) bool {
	path := strings.ToLower(filepath.ToSlash(filepath.Clean(filepath.FromSlash(relative))))
	return path == "railkeeper.db" || path == "railkeeper.db-wal" || path == "railkeeper.db-shm"
}

func validateDataRoot(path, label string) error {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("inspect %s data directory: %w", label, err)
	}
	if info.Mode()&os.ModeSymlink != 0 || isReparsePoint(info) {
		return fmt.Errorf("%s data directory is a link or reparse point: %s", label, path)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s data directory is not a directory: %s", label, path)
	}
	return nil
}

func legacyFileSHA256(path string) ([sha256.Size]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return [sha256.Size]byte{}, err
	}
	defer func() { _ = file.Close() }()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return [sha256.Size]byte{}, err
	}
	var digest [sha256.Size]byte
	copy(digest[:], hash.Sum(nil))
	return digest, nil
}

func regularFileCount(manifest []legacyManifestEntry) int {
	count := 0
	for _, entry := range manifest {
		if !entry.Dir {
			count++
		}
	}
	return count
}

func pathWithin(root, candidate string) bool {
	relative, err := filepath.Rel(root, candidate)
	if err != nil {
		return false
	}
	return relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) &&
		!filepath.IsAbs(relative)
}

func samePath(left, right string) bool {
	return strings.EqualFold(filepath.Clean(left), filepath.Clean(right))
}
