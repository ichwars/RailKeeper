package infrastructure

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateSQLiteSnapshotIncludesCommittedWALState(t *testing.T) {
	sourceDir := t.TempDir()
	db, err := OpenSQLite(sourceDir)
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	defer func() { _ = db.Close() }()

	if _, err = db.Exec(`PRAGMA wal_autocheckpoint=0`); err != nil {
		t.Fatalf("disable auto checkpoint: %v", err)
	}
	if _, err = db.Exec(`CREATE TABLE snapshot_items(id INTEGER PRIMARY KEY, value TEXT NOT NULL)`); err != nil {
		t.Fatalf("create snapshot table: %v", err)
	}
	if _, err = db.Exec(`INSERT INTO snapshot_items(value) VALUES ('from-wal')`); err != nil {
		t.Fatalf("insert snapshot row: %v", err)
	}
	if _, err = os.Stat(filepath.Join(sourceDir, "railkeeper.db-wal")); err != nil {
		t.Fatalf("expected WAL file before snapshot: %v", err)
	}

	target := filepath.Join(t.TempDir(), "snapshot.db")
	if err = CreateSQLiteSnapshot(context.Background(), db, target); err != nil {
		t.Fatalf("CreateSQLiteSnapshot() error = %v", err)
	}
	if err = ValidateSQLiteSnapshot(context.Background(), target); err != nil {
		t.Fatalf("ValidateSQLiteSnapshot() error = %v", err)
	}

	snapshot := openSnapshotTestDB(t, target)
	defer func() { _ = snapshot.Close() }()
	var value string
	if err = snapshot.QueryRow(`SELECT value FROM snapshot_items WHERE id=1`).Scan(&value); err != nil {
		t.Fatalf("read snapshot row: %v", err)
	}
	if value != "from-wal" {
		t.Fatalf("snapshot value = %q, want from-wal", value)
	}
}

func TestCreateSQLiteSnapshotFromPath(t *testing.T) {
	sourceDir := t.TempDir()
	db, err := OpenSQLite(sourceDir)
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	if _, err = db.Exec(`CREATE TABLE source_path_items(value TEXT NOT NULL)`); err != nil {
		t.Fatalf("create source-path table: %v", err)
	}
	if _, err = db.Exec(`INSERT INTO source_path_items(value) VALUES ('copied')`); err != nil {
		t.Fatalf("insert source-path row: %v", err)
	}
	if err = db.Close(); err != nil {
		t.Fatalf("close source DB: %v", err)
	}

	target := filepath.Join(t.TempDir(), "snapshot.db")
	if err = CreateSQLiteSnapshotFromPath(
		context.Background(), filepath.Join(sourceDir, "railkeeper.db"), target,
	); err != nil {
		t.Fatalf("CreateSQLiteSnapshotFromPath() error = %v", err)
	}

	snapshot := openSnapshotTestDB(t, target)
	defer func() { _ = snapshot.Close() }()
	var value string
	if err = snapshot.QueryRow(`SELECT value FROM source_path_items`).Scan(&value); err != nil {
		t.Fatalf("read source-path snapshot: %v", err)
	}
	if value != "copied" {
		t.Fatalf("snapshot value = %q, want copied", value)
	}
}

func TestCreateSQLiteSnapshotRefusesExistingTarget(t *testing.T) {
	sourceDir := t.TempDir()
	db, err := OpenSQLite(sourceDir)
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	defer func() { _ = db.Close() }()

	target := filepath.Join(t.TempDir(), "existing.db")
	if err = os.WriteFile(target, []byte("keep-me"), 0o600); err != nil {
		t.Fatalf("write existing target: %v", err)
	}
	if err = CreateSQLiteSnapshot(context.Background(), db, target); err == nil {
		t.Fatal("expected existing target to be rejected")
	}
	content, readErr := os.ReadFile(target)
	if readErr != nil {
		t.Fatalf("read existing target: %v", readErr)
	}
	if string(content) != "keep-me" {
		t.Fatalf("existing target changed to %q", content)
	}
}

func TestValidateSQLiteSnapshotRejectsCorruptFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "corrupt.db")
	if err := os.WriteFile(path, []byte("not a sqlite database"), 0o600); err != nil {
		t.Fatalf("write corrupt DB: %v", err)
	}
	if err := ValidateSQLiteSnapshot(context.Background(), path); err == nil {
		t.Fatal("expected corrupt snapshot to be rejected")
	}
}

func TestSQLiteSnapshotsEquivalentComparesLogicalContent(t *testing.T) {
	left := createEquivalentSnapshotDB(t, "same")
	right := createEquivalentSnapshotDB(t, "same")

	equivalent, err := SQLiteSnapshotsEquivalent(context.Background(), left, right, t.TempDir())
	if err != nil {
		t.Fatalf("SQLiteSnapshotsEquivalent() error = %v", err)
	}
	if !equivalent {
		t.Fatal("expected logically identical databases to be equivalent")
	}

	rightDB := openSnapshotTestDB(t, right)
	if _, err = rightDB.Exec(`INSERT INTO equivalent_items(value) VALUES ('different')`); err != nil {
		_ = rightDB.Close()
		t.Fatalf("insert differing row: %v", err)
	}
	if err = rightDB.Close(); err != nil {
		t.Fatalf("close differing DB: %v", err)
	}
	equivalent, err = SQLiteSnapshotsEquivalent(context.Background(), left, right, t.TempDir())
	if err != nil {
		t.Fatalf("SQLiteSnapshotsEquivalent() difference error = %v", err)
	}
	if equivalent {
		t.Fatal("expected different databases not to be equivalent")
	}
}

func createEquivalentSnapshotDB(t *testing.T, value string) string {
	t.Helper()
	dir := t.TempDir()
	db, err := OpenSQLite(dir)
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	if _, err = db.Exec(`CREATE TABLE equivalent_items(id INTEGER PRIMARY KEY, value TEXT NOT NULL)`); err != nil {
		_ = db.Close()
		t.Fatalf("create equivalent table: %v", err)
	}
	if _, err = db.Exec(`INSERT INTO equivalent_items(value) VALUES (?)`, value); err != nil {
		_ = db.Close()
		t.Fatalf("insert equivalent row: %v", err)
	}
	if err = db.Close(); err != nil {
		t.Fatalf("close equivalent DB: %v", err)
	}
	return filepath.Join(dir, "railkeeper.db")
}

func openSnapshotTestDB(t *testing.T, path string) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", path+"?_foreign_keys=on")
	if err != nil {
		t.Fatalf("open snapshot DB: %v", err)
	}
	if err = db.PingContext(context.Background()); err != nil {
		_ = db.Close()
		t.Fatalf("ping snapshot DB: %v", err)
	}
	return db
}
