package startup

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"railkeeper/backend/internal/infrastructure"
)

const legacyMigrationReceiptName = ".railkeeper-legacy-migration.json"

type LegacyMigrationStatus string

const (
	LegacyReady    LegacyMigrationStatus = "ready"
	LegacyMigrated LegacyMigrationStatus = "migrated"
	LegacyConflict LegacyMigrationStatus = "conflict"
)

type MigrationReceipt struct {
	SourcePath    string `json:"sourcePath"`
	TargetPath    string `json:"targetPath"`
	MigratedAt    string `json:"migratedAt"`
	Version       string `json:"version"`
	FilesVerified int    `json:"filesVerified"`
	Acknowledged  bool   `json:"acknowledged"`
}

type LegacyConflictInfo struct {
	SafePath   string
	LegacyPath string
	Reason     string
}

type LegacyMigrationOptions struct {
	Version      string
	Now          func() time.Time
	RandomSuffix func() string
	CopyFile     func(sourcePath, targetPath string) error
	Promote      func(stagingPath, targetPath string) error
}

type LegacyMigrationResult struct {
	Status   LegacyMigrationStatus
	DataDir  string
	Receipt  *MigrationReceipt
	Conflict *LegacyConflictInfo
}

type legacyManifestEntry struct {
	Path   string
	Dir    bool
	Size   int64
	SHA256 [sha256.Size]byte
}

func ResolveLegacyData(
	ctx context.Context,
	safeDataDir string,
	legacyDataDir string,
	options LegacyMigrationOptions,
) (LegacyMigrationResult, error) {
	if strings.TrimSpace(safeDataDir) == "" || strings.TrimSpace(legacyDataDir) == "" {
		return LegacyMigrationResult{}, errors.New("safe and legacy data directories are required")
	}
	safeDataDir = filepath.Clean(safeDataDir)
	legacyDataDir = filepath.Clean(legacyDataDir)
	if samePath(safeDataDir, legacyDataDir) {
		return LegacyMigrationResult{}, errors.New("safe and legacy data directories must differ")
	}
	if err := validateDataRoot(safeDataDir, "safe"); err != nil {
		return LegacyMigrationResult{}, err
	}
	if err := validateDataRoot(legacyDataDir, "legacy"); err != nil {
		return LegacyMigrationResult{}, err
	}

	safeDB := filepath.Join(safeDataDir, "railkeeper.db")
	legacyDB := filepath.Join(legacyDataDir, "railkeeper.db")
	safeExists, err := regularFileExists(safeDB)
	if err != nil {
		return LegacyMigrationResult{}, fmt.Errorf("inspect safe database: %w", err)
	}
	legacyExists, err := regularFileExists(legacyDB)
	if err != nil {
		return LegacyMigrationResult{}, fmt.Errorf("inspect legacy database: %w", err)
	}

	if !legacyExists {
		receipt, err := readMigrationReceipt(safeDataDir)
		if err != nil {
			return LegacyMigrationResult{}, err
		}
		return LegacyMigrationResult{Status: LegacyReady, DataDir: safeDataDir, Receipt: receipt}, nil
	}
	if safeExists {
		equivalent, compareErr := infrastructure.SQLiteSnapshotsEquivalent(
			ctx, safeDB, legacyDB, os.TempDir(),
		)
		if compareErr != nil {
			reason := "database equivalence could not be proven: " + compareErr.Error()
			return legacyConflict(safeDataDir, legacyDataDir, reason), nil //nolint:nilerr // Comparison failures require explicit user resolution.
		}
		if !equivalent {
			return legacyConflict(safeDataDir, legacyDataDir, "safe and legacy databases differ"), nil
		}
		receipt, err := readMigrationReceipt(safeDataDir)
		if err != nil {
			return LegacyMigrationResult{}, err
		}
		return LegacyMigrationResult{Status: LegacyReady, DataDir: safeDataDir, Receipt: receipt}, nil
	}
	nonEmpty, err := directoryHasEntries(safeDataDir)
	if err != nil {
		return LegacyMigrationResult{}, fmt.Errorf("inspect safe data directory: %w", err)
	}
	if nonEmpty {
		return legacyConflict(safeDataDir, legacyDataDir,
			"safe data directory contains files but no database"), nil
	}

	return migrateLegacyData(ctx, safeDataDir, legacyDataDir, legacyDB, options)
}

func migrateLegacyData(
	ctx context.Context,
	safeDataDir string,
	legacyDataDir string,
	legacyDB string,
	options LegacyMigrationOptions,
) (LegacyMigrationResult, error) {
	parent := filepath.Dir(safeDataDir)
	if err := os.MkdirAll(parent, 0o700); err != nil {
		return LegacyMigrationResult{}, fmt.Errorf("create safe data parent: %w", err)
	}
	suffix := randomLegacySuffix
	if options.RandomSuffix != nil {
		suffix = options.RandomSuffix
	}
	staging := filepath.Join(parent, ".railkeeper-migration-"+suffix())
	if !pathWithin(parent, staging) {
		return LegacyMigrationResult{}, errors.New("legacy staging path escaped safe parent")
	}
	if err := os.Mkdir(staging, 0o700); err != nil {
		return LegacyMigrationResult{}, fmt.Errorf("create legacy migration staging directory: %w", err)
	}
	promoted := false
	defer func() {
		if !promoted {
			_ = os.RemoveAll(staging)
		}
	}()

	beforeFull, err := buildLegacyManifest(legacyDataDir, true)
	if err != nil {
		return LegacyMigrationResult{}, err
	}
	sourceContent, err := buildLegacyManifest(legacyDataDir, false)
	if err != nil {
		return LegacyMigrationResult{}, err
	}
	if err := infrastructure.CreateSQLiteSnapshotFromPath(
		ctx, legacyDB, filepath.Join(staging, "railkeeper.db"),
	); err != nil {
		return LegacyMigrationResult{}, fmt.Errorf("snapshot legacy database: %w", err)
	}
	copyFile := copyLegacyFile
	if options.CopyFile != nil {
		copyFile = options.CopyFile
	}
	if err := copyLegacyManifest(legacyDataDir, staging, sourceContent, copyFile); err != nil {
		return LegacyMigrationResult{}, err
	}
	afterFull, err := buildLegacyManifest(legacyDataDir, true)
	if err != nil {
		return LegacyMigrationResult{}, err
	}
	if !slices.Equal(beforeFull, afterFull) {
		return LegacyMigrationResult{}, errors.New("legacy data changed during migration")
	}
	targetContent, err := buildLegacyManifest(staging, false)
	if err != nil {
		return LegacyMigrationResult{}, err
	}
	if !slices.Equal(sourceContent, targetContent) {
		return LegacyMigrationResult{}, errors.New("staged legacy file manifest does not match source")
	}
	if err := infrastructure.ValidateSQLiteSnapshot(
		ctx, filepath.Join(staging, "railkeeper.db"),
	); err != nil {
		return LegacyMigrationResult{}, fmt.Errorf("validate staged legacy database: %w", err)
	}

	now := time.Now
	if options.Now != nil {
		now = options.Now
	}
	receipt := &MigrationReceipt{
		SourcePath:    legacyDataDir,
		TargetPath:    safeDataDir,
		MigratedAt:    now().UTC().Format(time.RFC3339),
		Version:       options.Version,
		FilesVerified: regularFileCount(sourceContent) + 1,
	}
	if err := writeMigrationReceipt(staging, receipt); err != nil {
		return LegacyMigrationResult{}, err
	}

	if err := removeEmptyTarget(safeDataDir); err != nil {
		return LegacyMigrationResult{}, err
	}
	promote := os.Rename
	if options.Promote != nil {
		promote = options.Promote
	}
	if err := promote(staging, safeDataDir); err != nil {
		return LegacyMigrationResult{}, fmt.Errorf("activate migrated data directory: %w", err)
	}
	promoted = true
	return LegacyMigrationResult{
		Status:  LegacyMigrated,
		DataDir: safeDataDir,
		Receipt: receipt,
	}, nil
}

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

func writeMigrationReceipt(root string, receipt *MigrationReceipt) error {
	data, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		return fmt.Errorf("encode legacy migration receipt: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(filepath.Join(root, legacyMigrationReceiptName), data, 0o600); err != nil {
		return fmt.Errorf("write legacy migration receipt: %w", err)
	}
	return nil
}

func readMigrationReceipt(root string) (*MigrationReceipt, error) {
	path := filepath.Join(root, legacyMigrationReceiptName)
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read legacy migration receipt: %w", err)
	}
	var receipt MigrationReceipt
	if err := json.Unmarshal(data, &receipt); err != nil {
		return nil, fmt.Errorf("decode legacy migration receipt: %w", err)
	}
	if filepath.Clean(receipt.TargetPath) != filepath.Clean(root) || receipt.SourcePath == "" {
		return nil, errors.New("legacy migration receipt paths are invalid")
	}
	return &receipt, nil
}

func AcknowledgeMigrationReceipt(ctx context.Context, root string) (*MigrationReceipt, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	root = filepath.Clean(root)
	if err := validateDataRoot(root, "safe"); err != nil {
		return nil, err
	}
	receipt, err := readMigrationReceipt(root)
	if err != nil {
		return nil, err
	}
	if receipt == nil {
		return nil, errors.New("legacy migration receipt does not exist")
	}
	if receipt.Acknowledged {
		return receipt, nil
	}

	updated := *receipt
	updated.Acknowledged = true
	data, err := json.MarshalIndent(&updated, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode acknowledged migration receipt: %w", err)
	}
	data = append(data, '\n')
	targetPath := filepath.Join(root, legacyMigrationReceiptName)
	temporaryPath := filepath.Join(root, legacyMigrationReceiptName+".tmp-"+randomLegacySuffix())
	if !pathWithin(root, temporaryPath) {
		return nil, errors.New("migration receipt temporary path escaped data directory")
	}
	temporary, err := os.OpenFile(temporaryPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return nil, fmt.Errorf("create migration receipt update: %w", err)
	}
	removeTemporary := true
	defer func() {
		if removeTemporary {
			_ = os.Remove(temporaryPath)
		}
	}()
	if _, err = temporary.Write(data); err != nil {
		_ = temporary.Close()
		return nil, fmt.Errorf("write migration receipt update: %w", err)
	}
	if err = temporary.Sync(); err != nil {
		_ = temporary.Close()
		return nil, fmt.Errorf("sync migration receipt update: %w", err)
	}
	if err = temporary.Close(); err != nil {
		return nil, fmt.Errorf("close migration receipt update: %w", err)
	}
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	if err = replaceFileAtomically(temporaryPath, targetPath); err != nil {
		return nil, fmt.Errorf("activate migration receipt update: %w", err)
	}
	removeTemporary = false
	return &updated, nil
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
