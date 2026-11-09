package startup

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"railkeeper/backend/internal/infrastructure"
	"slices"
	"strings"
	"time"
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
		filesEquivalent, reason := nonDatabaseDataEquivalent(safeDataDir, legacyDataDir)
		if !filesEquivalent {
			return legacyConflict(safeDataDir, legacyDataDir, reason), nil
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
	if err := cleanupAbandonedLegacyStaging(parent, safeDataDir); err != nil {
		return LegacyMigrationResult{}, err
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

func nonDatabaseDataEquivalent(safeDataDir, legacyDataDir string) (bool, string) {
	safeContent, err := buildLegacyManifest(safeDataDir, false)
	if err != nil {
		return false, "safe non-database data could not be verified: " + err.Error()
	}
	legacyContent, err := buildLegacyManifest(legacyDataDir, false)
	if err != nil {
		return false, "legacy non-database data could not be verified: " + err.Error()
	}
	if !slices.Equal(safeContent, legacyContent) {
		return false, "safe and legacy non-database data differ"
	}
	return true, ""
}

func cleanupAbandonedLegacyStaging(parent, safeDataDir string) error {
	entries, err := os.ReadDir(parent)
	if err != nil {
		return fmt.Errorf("inspect legacy migration staging parent: %w", err)
	}
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), ".railkeeper-migration-") {
			continue
		}
		stagingPath := filepath.Join(parent, entry.Name())
		if !pathWithin(parent, stagingPath) || samePath(stagingPath, safeDataDir) {
			return fmt.Errorf("invalid abandoned legacy staging path: %s", stagingPath)
		}
		info, infoErr := entry.Info()
		if infoErr != nil {
			return fmt.Errorf("inspect abandoned legacy staging path: %w", infoErr)
		}
		if !entry.IsDir() || info.Mode()&os.ModeSymlink != 0 || isReparsePoint(info) {
			return fmt.Errorf("abandoned legacy staging path is not a verified directory: %s", stagingPath)
		}
		if _, manifestErr := buildLegacyManifest(stagingPath, true); manifestErr != nil {
			return fmt.Errorf("verify abandoned legacy staging directory: %w", manifestErr)
		}
		if removeErr := os.RemoveAll(stagingPath); removeErr != nil {
			return fmt.Errorf("remove abandoned legacy staging directory: %w", removeErr)
		}
	}
	return nil
}
