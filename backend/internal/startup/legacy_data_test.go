package startup

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"railkeeper/backend/internal/infrastructure"

	_ "modernc.org/sqlite"
)

func TestResolveLegacyDataUsesSafeDirectoryForNewInstallation(t *testing.T) {
	safeDir := filepath.Join(t.TempDir(), "RailKeeper", "data")
	legacyDir := filepath.Join(t.TempDir(), "data")

	result, err := ResolveLegacyData(context.Background(), safeDir, legacyDir, LegacyMigrationOptions{})
	if err != nil {
		t.Fatalf("ResolveLegacyData() error = %v", err)
	}
	if result.Status != LegacyReady || result.DataDir != safeDir || result.Conflict != nil {
		t.Fatalf("unexpected new-install result: %#v", result)
	}
	if _, err = os.Stat(safeDir); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("new-install inspection created data directory: %v", err)
	}
}

func TestResolveLegacyDataUsesExistingSafeDatabase(t *testing.T) {
	safeDir := t.TempDir()
	createLegacyTestDB(t, safeDir, "safe")
	legacyDir := filepath.Join(t.TempDir(), "data")

	result, err := ResolveLegacyData(context.Background(), safeDir, legacyDir, LegacyMigrationOptions{})
	if err != nil {
		t.Fatalf("ResolveLegacyData() error = %v", err)
	}
	if result.Status != LegacyReady || result.DataDir != safeDir {
		t.Fatalf("unexpected safe-data result: %#v", result)
	}
}

func TestResolveLegacyDataMigratesCompleteLegacyDirectoryAndPreservesMasterDataState(t *testing.T) {
	root := t.TempDir()
	legacyDir := filepath.Join(root, "program", "data")
	safeDir := filepath.Join(root, "local-app-data", "RailKeeper", "data")
	createMasterDataLegacyDB(t, legacyDir)
	writeLegacyFile(t, legacyDir, "uploads/vehicles/rk-1/manual.pdf", "manual")
	writeLegacyFile(t, legacyDir, "backups/export.json", "backup")
	writeLegacyFile(t, legacyDir, "future/unknown.bin", "future")
	writeLegacyFile(t, legacyDir, "future/railkeeper.db", "not-the-main-database")
	before := legacyTreeDigest(t, legacyDir)
	now := time.Date(2026, 8, 16, 14, 30, 0, 0, time.UTC)

	result, err := ResolveLegacyData(context.Background(), safeDir, legacyDir, LegacyMigrationOptions{
		Version:      "0.1.18",
		Now:          func() time.Time { return now },
		RandomSuffix: func() string { return "fixed" },
	})
	if err != nil {
		t.Fatalf("ResolveLegacyData() error = %v", err)
	}
	if result.Status != LegacyMigrated || result.DataDir != safeDir || result.Receipt == nil {
		t.Fatalf("unexpected migrated result: %#v", result)
	}
	if result.Receipt.SourcePath != legacyDir || result.Receipt.TargetPath != safeDir ||
		result.Receipt.MigratedAt != "2026-08-16T14:30:00Z" || result.Receipt.Acknowledged {
		t.Fatalf("unexpected migration receipt: %#v", result.Receipt)
	}
	after := legacyTreeDigest(t, legacyDir)
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("legacy source changed:\nbefore=%v\nafter=%v", before, after)
	}
	for path, want := range map[string]string{
		"uploads/vehicles/rk-1/manual.pdf": "manual",
		"backups/export.json":              "backup",
		"future/unknown.bin":               "future",
		"future/railkeeper.db":             "not-the-main-database",
	} {
		content, readErr := os.ReadFile(filepath.Join(safeDir, filepath.FromSlash(path)))
		if readErr != nil || string(content) != want {
			t.Fatalf("migrated file %s = %q, %v", path, content, readErr)
		}
	}
	if err = infrastructure.ValidateSQLiteSnapshot(
		context.Background(), filepath.Join(safeDir, "railkeeper.db"),
	); err != nil {
		t.Fatalf("migrated DB validation: %v", err)
	}
	targetDB := openLegacyTestDB(t, filepath.Join(safeDir, "railkeeper.db"))
	defer func() { _ = targetDB.Close() }()
	var origin string
	var active int
	if err = targetDB.QueryRow(
		`SELECT origin, active FROM master_data_entries WHERE type='manufacturer' AND key='tillig-modellbahnen'`,
	).Scan(&origin, &active); err != nil {
		t.Fatalf("read migrated Tillig state: %v", err)
	}
	if origin != "bundled" || active != 0 {
		t.Fatalf("migrated Tillig state = origin:%q active:%d", origin, active)
	}
}

func TestResolveLegacyDataReturnsConflictForDifferentDatabases(t *testing.T) {
	root := t.TempDir()
	safeDir := filepath.Join(root, "safe")
	legacyDir := filepath.Join(root, "legacy")
	createLegacyTestDB(t, safeDir, "safe")
	createLegacyTestDB(t, legacyDir, "legacy")
	beforeSafe := legacyTreeDigest(t, safeDir)
	beforeLegacy := legacyTreeDigest(t, legacyDir)

	result, err := ResolveLegacyData(context.Background(), safeDir, legacyDir, LegacyMigrationOptions{})
	if err != nil {
		t.Fatalf("ResolveLegacyData() error = %v", err)
	}
	if result.Status != LegacyConflict || result.Conflict == nil ||
		result.Conflict.SafePath != safeDir || result.Conflict.LegacyPath != legacyDir {
		t.Fatalf("unexpected conflict result: %#v", result)
	}
	if !reflect.DeepEqual(beforeSafe, legacyTreeDigest(t, safeDir)) ||
		!reflect.DeepEqual(beforeLegacy, legacyTreeDigest(t, legacyDir)) {
		t.Fatal("database conflict changed source data")
	}
}

func TestResolveLegacyDataAcceptsEquivalentDatabases(t *testing.T) {
	root := t.TempDir()
	safeDir := filepath.Join(root, "safe")
	legacyDir := filepath.Join(root, "legacy")
	createLegacyTestDB(t, legacyDir, "same")
	if err := os.MkdirAll(safeDir, 0o700); err != nil {
		t.Fatalf("create safe dir: %v", err)
	}
	if err := infrastructure.CreateSQLiteSnapshotFromPath(
		context.Background(), filepath.Join(legacyDir, "railkeeper.db"),
		filepath.Join(safeDir, "railkeeper.db"),
	); err != nil {
		t.Fatalf("create equivalent safe DB: %v", err)
	}

	result, err := ResolveLegacyData(context.Background(), safeDir, legacyDir, LegacyMigrationOptions{})
	if err != nil {
		t.Fatalf("ResolveLegacyData() error = %v", err)
	}
	if result.Status != LegacyReady || result.Conflict != nil || result.DataDir != safeDir {
		t.Fatalf("unexpected equivalent result: %#v", result)
	}
}

func TestResolveLegacyDataConflictsWhenEquivalentDatabasesHaveDifferentFiles(t *testing.T) {
	root := t.TempDir()
	safeDir := filepath.Join(root, "safe")
	legacyDir := filepath.Join(root, "legacy")
	createLegacyTestDB(t, legacyDir, "same")
	if err := os.MkdirAll(safeDir, 0o700); err != nil {
		t.Fatalf("create safe dir: %v", err)
	}
	if err := infrastructure.CreateSQLiteSnapshotFromPath(
		context.Background(), filepath.Join(legacyDir, "railkeeper.db"),
		filepath.Join(safeDir, "railkeeper.db"),
	); err != nil {
		t.Fatalf("create equivalent safe DB: %v", err)
	}
	writeLegacyFile(t, legacyDir, "uploads/vehicles/manual.pdf", "legacy-only")

	result, err := ResolveLegacyData(context.Background(), safeDir, legacyDir, LegacyMigrationOptions{})
	if err != nil {
		t.Fatalf("ResolveLegacyData() error = %v", err)
	}
	if result.Status != LegacyConflict || result.Conflict == nil {
		t.Fatalf("expected non-database data conflict, got %#v", result)
	}
	if _, statErr := os.Stat(filepath.Join(safeDir, "uploads", "vehicles", "manual.pdf")); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("conflict copied legacy-only file: %v", statErr)
	}
}

func TestResolveLegacyDataConflictsWithNonEmptySafeDirectoryWithoutDatabase(t *testing.T) {
	root := t.TempDir()
	safeDir := filepath.Join(root, "safe")
	legacyDir := filepath.Join(root, "legacy")
	writeLegacyFile(t, safeDir, "unknown-user-file.txt", "keep")
	createLegacyTestDB(t, legacyDir, "legacy")

	result, err := ResolveLegacyData(context.Background(), safeDir, legacyDir, LegacyMigrationOptions{})
	if err != nil {
		t.Fatalf("ResolveLegacyData() error = %v", err)
	}
	if result.Status != LegacyConflict || result.Conflict == nil {
		t.Fatalf("expected non-empty target conflict, got %#v", result)
	}
	content, readErr := os.ReadFile(filepath.Join(safeDir, "unknown-user-file.txt"))
	if readErr != nil || string(content) != "keep" {
		t.Fatalf("safe user file changed: %q, %v", content, readErr)
	}
}

func TestResolveLegacyDataCleansStagingAfterCopyFailure(t *testing.T) {
	root := t.TempDir()
	safeDir := filepath.Join(root, "local", "data")
	legacyDir := filepath.Join(root, "legacy")
	createLegacyTestDB(t, legacyDir, "legacy")
	writeLegacyFile(t, legacyDir, "uploads/manual.pdf", "manual")
	before := legacyTreeDigest(t, legacyDir)

	_, err := ResolveLegacyData(context.Background(), safeDir, legacyDir, LegacyMigrationOptions{
		RandomSuffix: func() string { return "copy-failure" },
		CopyFile: func(string, string) error {
			return errors.New("copy interrupted")
		},
	})
	if err == nil {
		t.Fatal("expected copy failure")
	}
	if _, statErr := os.Stat(safeDir); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("copy failure left active target: %v", statErr)
	}
	staging := filepath.Join(filepath.Dir(safeDir), ".railkeeper-migration-copy-failure")
	if _, statErr := os.Stat(staging); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("copy failure left staging directory: %v", statErr)
	}
	if !reflect.DeepEqual(before, legacyTreeDigest(t, legacyDir)) {
		t.Fatal("copy failure changed legacy source")
	}
}

func TestResolveLegacyDataCleansStagingAfterPromotionFailure(t *testing.T) {
	root := t.TempDir()
	safeDir := filepath.Join(root, "local", "data")
	legacyDir := filepath.Join(root, "legacy")
	createLegacyTestDB(t, legacyDir, "legacy")

	_, err := ResolveLegacyData(context.Background(), safeDir, legacyDir, LegacyMigrationOptions{
		RandomSuffix: func() string { return "promote-failure" },
		Promote: func(string, string) error {
			return errors.New("promotion interrupted")
		},
	})
	if err == nil {
		t.Fatal("expected promotion failure")
	}
	if _, statErr := os.Stat(safeDir); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("promotion failure left active target: %v", statErr)
	}
	staging := filepath.Join(filepath.Dir(safeDir), ".railkeeper-migration-promote-failure")
	if _, statErr := os.Stat(staging); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("promotion failure left staging directory: %v", statErr)
	}
}

func TestResolveLegacyDataCleansVerifiedAbandonedStagingBeforeMigration(t *testing.T) {
	root := t.TempDir()
	safeDir := filepath.Join(root, "local", "data")
	legacyDir := filepath.Join(root, "legacy")
	createLegacyTestDB(t, legacyDir, "legacy")
	abandoned := filepath.Join(filepath.Dir(safeDir), ".railkeeper-migration-abandoned")
	writeLegacyFile(t, abandoned, "uploads/private.pdf", "private")

	result, err := ResolveLegacyData(context.Background(), safeDir, legacyDir, LegacyMigrationOptions{
		RandomSuffix: func() string { return "new" },
	})
	if err != nil {
		t.Fatalf("ResolveLegacyData() error = %v", err)
	}
	if result.Status != LegacyMigrated {
		t.Fatalf("migration status = %q", result.Status)
	}
	if _, statErr := os.Stat(abandoned); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("abandoned staging remains: %v", statErr)
	}
}

func TestResolveLegacyDataRejectsLinksInLegacyDirectory(t *testing.T) {
	root := t.TempDir()
	safeDir := filepath.Join(root, "safe")
	legacyDir := filepath.Join(root, "legacy")
	createLegacyTestDB(t, legacyDir, "legacy")
	outside := filepath.Join(root, "outside.txt")
	if err := os.WriteFile(outside, []byte("outside"), 0o600); err != nil {
		t.Fatalf("write outside file: %v", err)
	}
	if err := os.Symlink(outside, filepath.Join(legacyDir, "outside-link")); err != nil {
		t.Skipf("symlink creation unavailable: %v", err)
	}

	_, err := ResolveLegacyData(context.Background(), safeDir, legacyDir, LegacyMigrationOptions{})
	if err == nil {
		t.Fatal("expected legacy link to be rejected")
	}
	if _, statErr := os.Stat(safeDir); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("link rejection left active target: %v", statErr)
	}
}

func TestResolveLegacyDataRejectsLinkedLegacyRoot(t *testing.T) {
	root := t.TempDir()
	safeDir := filepath.Join(root, "safe")
	legacyActual := filepath.Join(root, "legacy-actual")
	legacyLink := filepath.Join(root, "legacy-link")
	createLegacyTestDB(t, legacyActual, "legacy")
	if err := os.Symlink(legacyActual, legacyLink); err != nil {
		t.Skipf("directory symlink creation unavailable: %v", err)
	}

	_, err := ResolveLegacyData(context.Background(), safeDir, legacyLink, LegacyMigrationOptions{})
	if err == nil {
		t.Fatal("expected linked legacy root to be rejected")
	}
	if _, statErr := os.Stat(safeDir); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("linked-root rejection left active target: %v", statErr)
	}
}

func createLegacyTestDB(t *testing.T, dir, value string) {
	t.Helper()
	db, err := infrastructure.OpenSQLite(dir)
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	if _, err = db.Exec(`CREATE TABLE legacy_identity(value TEXT NOT NULL)`); err != nil {
		_ = db.Close()
		t.Fatalf("create legacy identity: %v", err)
	}
	if _, err = db.Exec(`INSERT INTO legacy_identity(value) VALUES (?)`, value); err != nil {
		_ = db.Close()
		t.Fatalf("insert legacy identity: %v", err)
	}
	if err = db.Close(); err != nil {
		t.Fatalf("close legacy DB: %v", err)
	}
}

func createMasterDataLegacyDB(t *testing.T, dir string) {
	t.Helper()
	db, err := infrastructure.OpenSQLite(dir)
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	if err = infrastructure.Migrate(db, filepath.Join("..", "..", "migrations")); err != nil {
		_ = db.Close()
		t.Fatalf("migrate legacy DB: %v", err)
	}
	if err = infrastructure.SeedMasterData(db, filepath.Join("..", "..", "seeds")); err != nil {
		_ = db.Close()
		t.Fatalf("seed legacy DB: %v", err)
	}
	if _, err = db.Exec(
		`UPDATE master_data_entries SET active=0 WHERE type='manufacturer' AND key='tillig-modellbahnen'`,
	); err != nil {
		_ = db.Close()
		t.Fatalf("deactivate Tillig: %v", err)
	}
	if err = db.Close(); err != nil {
		t.Fatalf("close master-data DB: %v", err)
	}
}

func writeLegacyFile(t *testing.T, root, relative, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("create legacy directory: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write legacy file: %v", err)
	}
}

func legacyTreeDigest(t *testing.T, root string) map[string]string {
	t.Helper()
	digest := map[string]string{}
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
		if entry.IsDir() {
			digest[filepath.ToSlash(relative)+"/"] = "directory"
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			target, err := os.Readlink(path)
			if err != nil {
				return err
			}
			digest[filepath.ToSlash(relative)] = "link:" + target
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		digest[filepath.ToSlash(relative)] = fmt.Sprintf("%x", sha256.Sum256(content))
		return nil
	})
	if err != nil {
		t.Fatalf("digest tree %s: %v", root, err)
	}
	return digest
}

func openLegacyTestDB(t *testing.T, path string) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open legacy test DB: %v", err)
	}
	if err = db.Ping(); err != nil {
		_ = db.Close()
		t.Fatalf("ping legacy test DB: %v", err)
	}
	return db
}
