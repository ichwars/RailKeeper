package infrastructure_test

import (
	"os"
	"path/filepath"
	"testing"

	"railkeeper/backend/internal/infrastructure"
)

func TestMasterDataOriginMigrationPreservesExistingRowsAsCustom(t *testing.T) {
	root := t.TempDir()
	migrationsDir := filepath.Join(root, "migrations")
	if err := os.Mkdir(migrationsDir, 0o700); err != nil {
		t.Fatal(err)
	}
	copyMigrationsThrough(t, filepath.Join("..", "..", "migrations"), migrationsDir,
		"0057_accessory_list_price.sql")
	db, err := infrastructure.OpenSQLite(filepath.Join(root, "data"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := infrastructure.Migrate(db, migrationsDir); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`
INSERT INTO master_data_entries(
  id, type, key, label, active, sort_order, metadata_json, created_at, updated_at
) VALUES('manufacturer:club', 'manufacturer', 'club', 'Club', 0, 0, '{}', 'now', 'now')`); err != nil {
		t.Fatal(err)
	}

	applyMigrationFile(t, db, "0058_master_data_origin.sql")

	var origin string
	var active int
	if err := db.QueryRow(`
SELECT origin, active FROM master_data_entries
WHERE type='manufacturer' AND key='club'`).Scan(&origin, &active); err != nil {
		t.Fatal(err)
	}
	if origin != "custom" || active != 0 {
		t.Fatalf("origin=%q active=%d", origin, active)
	}
}
