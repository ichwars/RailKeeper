package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestInspectMigrationPlanDoesNotCreateSchemaTable(t *testing.T) {
	dataDir := t.TempDir()
	db, err := OpenSQLite(dataDir)
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	defer func() { _ = db.Close() }()
	migrationsDir := safetyMigrations(t, true)

	plan, err := InspectMigrationPlan(context.Background(), db, migrationsDir)
	if err != nil {
		t.Fatalf("InspectMigrationPlan() error = %v", err)
	}
	if len(plan.Applied) != 0 || strings.Join(plan.Pending, ",") != "0001_initial,0002_marker" {
		t.Fatalf("unexpected migration plan: %#v", plan)
	}
	if plan.From != "unversioned" || plan.To != "0002_marker" {
		t.Fatalf("unexpected migration range: %#v", plan)
	}
	if safetyTableExists(t, db, "schema_migrations") {
		t.Fatal("inspection created schema_migrations")
	}
}

func TestMigrateSafelySkipsBackupForNewDatabase(t *testing.T) {
	dataDir := t.TempDir()
	db, err := OpenSQLite(dataDir)
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	defer func() { _ = db.Close() }()
	called := false

	result, err := MigrateSafely(context.Background(), db, dataDir, safetyMigrations(t, true),
		MigrationSafetyOptions{
			DatabaseExisted: false,
			Snapshot: func(context.Context, *sql.DB, string) error {
				called = true
				return nil
			},
		})
	if err != nil {
		t.Fatalf("MigrateSafely() error = %v", err)
	}
	if called || result.BackupPath != "" {
		t.Fatalf("new database created a backup: called=%v path=%q", called, result.BackupPath)
	}
	if !safetyTableExists(t, db, "migration_marker") {
		t.Fatal("new database migrations were not applied")
	}
}

func TestMigrateSafelySkipsBackupWhenExistingDatabaseIsCurrent(t *testing.T) {
	dataDir := t.TempDir()
	db, err := OpenSQLite(dataDir)
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	defer func() { _ = db.Close() }()
	migrationsDir := safetyMigrations(t, true)
	if err = Migrate(db, migrationsDir); err != nil {
		t.Fatalf("initial migrate: %v", err)
	}
	called := false

	result, err := MigrateSafely(context.Background(), db, dataDir, migrationsDir,
		MigrationSafetyOptions{
			DatabaseExisted: true,
			Snapshot: func(context.Context, *sql.DB, string) error {
				called = true
				return nil
			},
		})
	if err != nil {
		t.Fatalf("MigrateSafely() error = %v", err)
	}
	if called || len(result.Plan.Pending) != 0 || result.BackupPath != "" {
		t.Fatalf("current database created a backup: result=%#v called=%v", result, called)
	}
}

func TestMigrateSafelyBacksUpExistingDatabaseBeforePendingMigration(t *testing.T) {
	dataDir := t.TempDir()
	db, err := OpenSQLite(dataDir)
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	defer func() { _ = db.Close() }()
	initial := safetyMigrations(t, false)
	if err = Migrate(db, initial); err != nil {
		t.Fatalf("initial migrate: %v", err)
	}
	full := safetyMigrations(t, true)
	now := time.Date(2026, 8, 16, 12, 34, 56, 0, time.UTC)

	result, err := MigrateSafely(context.Background(), db, dataDir, full,
		MigrationSafetyOptions{
			DatabaseExisted: true,
			Now:             func() time.Time { return now },
			Snapshot:        CreateSQLiteSnapshot,
		})
	if err != nil {
		t.Fatalf("MigrateSafely() error = %v", err)
	}
	wantName := "railkeeper-pre-migration-0001_initial-to-0002_marker-20260816T123456Z.db"
	if filepath.Base(result.BackupPath) != wantName {
		t.Fatalf("backup filename = %q, want %q", filepath.Base(result.BackupPath), wantName)
	}
	if err = ValidateSQLiteSnapshot(context.Background(), result.BackupPath); err != nil {
		t.Fatalf("backup validation: %v", err)
	}
	backup := openSnapshotTestDB(t, result.BackupPath)
	defer func() { _ = backup.Close() }()
	if safetyTableExists(t, backup, "migration_marker") {
		t.Fatal("backup unexpectedly contains pending migration")
	}
	if !safetyTableExists(t, db, "migration_marker") {
		t.Fatal("pending migration was not applied after backup")
	}
}

func TestMigrateSafelyStopsBeforeMigrationWhenSnapshotFails(t *testing.T) {
	dataDir, db, full := pendingSafetyMigrationDB(t)
	defer func() { _ = db.Close() }()

	_, err := MigrateSafely(context.Background(), db, dataDir, full,
		MigrationSafetyOptions{
			DatabaseExisted: true,
			Snapshot: func(context.Context, *sql.DB, string) error {
				return errors.New("disk full")
			},
		})
	if err == nil || !strings.Contains(err.Error(), "disk full") {
		t.Fatalf("expected snapshot failure, got %v", err)
	}
	assertPendingMigrationUnchanged(t, db)
}

func TestMigrateSafelyStopsBeforeMigrationWhenSnapshotIsCorrupt(t *testing.T) {
	dataDir, db, full := pendingSafetyMigrationDB(t)
	defer func() { _ = db.Close() }()

	_, err := MigrateSafely(context.Background(), db, dataDir, full,
		MigrationSafetyOptions{
			DatabaseExisted: true,
			Snapshot: func(_ context.Context, _ *sql.DB, target string) error {
				return os.WriteFile(target, []byte("corrupt"), 0o600)
			},
		})
	if err == nil || !strings.Contains(err.Error(), "validate") {
		t.Fatalf("expected validation failure, got %v", err)
	}
	assertPendingMigrationUnchanged(t, db)
}

func pendingSafetyMigrationDB(t *testing.T) (string, *sql.DB, string) {
	t.Helper()
	dataDir := t.TempDir()
	db, err := OpenSQLite(dataDir)
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	if err = Migrate(db, safetyMigrations(t, false)); err != nil {
		_ = db.Close()
		t.Fatalf("initial migrate: %v", err)
	}
	return dataDir, db, safetyMigrations(t, true)
}

func assertPendingMigrationUnchanged(t *testing.T, db *sql.DB) {
	t.Helper()
	if safetyTableExists(t, db, "migration_marker") {
		t.Fatal("pending migration changed the schema after backup failure")
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&count); err != nil {
		t.Fatalf("count applied migrations: %v", err)
	}
	if count != 1 {
		t.Fatalf("applied migration count = %d, want 1", count)
	}
}

func safetyMigrations(t *testing.T, includeMarker bool) string {
	t.Helper()
	dir := t.TempDir()
	writeSafetyMigration(t, dir, "0001_initial.sql", `
CREATE TABLE initial_data(id INTEGER PRIMARY KEY, value TEXT NOT NULL);
INSERT INTO initial_data(value) VALUES ('preserve-me');
`)
	if includeMarker {
		writeSafetyMigration(t, dir, "0002_marker.sql", `
CREATE TABLE migration_marker(id INTEGER PRIMARY KEY);
`)
	}
	return dir
}

func writeSafetyMigration(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
		t.Fatalf("write migration %s: %v", name, err)
	}
}

func safetyTableExists(t *testing.T, db *sql.DB, name string) bool {
	t.Helper()
	var count int
	if err := db.QueryRow(
		`SELECT COUNT(*) FROM sqlite_schema WHERE type='table' AND name=?`, name,
	).Scan(&count); err != nil {
		t.Fatalf("inspect table %s: %v", name, err)
	}
	return count == 1
}
