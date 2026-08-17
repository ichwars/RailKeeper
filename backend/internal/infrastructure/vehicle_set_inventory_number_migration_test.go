package infrastructure_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"railkeeper/backend/internal/infrastructure"
)

func TestVehicleSetInventoryNumberMigrationBackfillsDeterministically(t *testing.T) {
	db := migratedVehicleSetInventoryNumberTestDB(t)
	seedVehicleSetBeforeInventoryNumberMigration(t, db, "set-b", "2026-08-02T00:00:00Z")
	seedVehicleSetBeforeInventoryNumberMigration(t, db, "set-a", "2026-08-01T00:00:00Z")

	applyMigrationFile(t, db, "0060_vehicle_set_inventory_number.sql")

	assertText(t, db, `SELECT inventory_number FROM vehicle_sets WHERE id='set-a'`, "RK-SET-000001")
	assertText(t, db, `SELECT inventory_number FROM vehicle_sets WHERE id='set-b'`, "RK-SET-000002")
	assertText(t, db, `SELECT CAST(next_number AS TEXT) FROM inventory_number_schemes WHERE category='Set'`, "3")
	expectConstraintFailure(t, db, `UPDATE vehicle_sets SET inventory_number='' WHERE id='set-a'`)
	expectConstraintFailure(t, db, `UPDATE vehicle_sets SET inventory_number='RK-SET-000001' WHERE id='set-b'`)
	assertForeignKeyCheck(t, db)
}

func TestVehicleSetInventoryNumberMigrationUsesExistingScheme(t *testing.T) {
	db := migratedVehicleSetInventoryNumberTestDB(t)
	if _, err := db.Exec(`
INSERT INTO inventory_number_schemes(
  id, category, prefix, next_number, padding, active, created_at, updated_at
) VALUES('set-scheme', 'Set', 'CLUB-SET', 7, 3, 0, '2026-08-01', '2026-08-01')`); err != nil {
		t.Fatal(err)
	}
	seedVehicleSetBeforeInventoryNumberMigration(t, db, "set-custom", "2026-08-01T00:00:00Z")

	applyMigrationFile(t, db, "0060_vehicle_set_inventory_number.sql")

	assertText(t, db, `SELECT inventory_number FROM vehicle_sets WHERE id='set-custom'`, "CLUB-SET-007")
	assertText(t, db, `SELECT CAST(next_number AS TEXT) FROM inventory_number_schemes WHERE category='Set'`, "8")
	assertText(t, db, `SELECT CAST(active AS TEXT) FROM inventory_number_schemes WHERE category='Set'`, "1")
}

func migratedVehicleSetInventoryNumberTestDB(t *testing.T) *sql.DB {
	t.Helper()
	root := t.TempDir()
	migrationsDir := filepath.Join(root, "migrations")
	if err := os.Mkdir(migrationsDir, 0700); err != nil {
		t.Fatal(err)
	}
	copyMigrationsThrough(t, filepath.Join("..", "..", "migrations"), migrationsDir,
		"0059_vehicle_sets.sql")

	db, err := infrastructure.OpenSQLite(filepath.Join(root, "data"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := infrastructure.Migrate(db, migrationsDir); err != nil {
		t.Fatal(err)
	}
	return db
}

func seedVehicleSetBeforeInventoryNumberMigration(t *testing.T, db *sql.DB, id, createdAt string) {
	t.Helper()
	if _, err := db.Exec(`
INSERT INTO vehicle_sets(
  id, name, manufacturer, gauge, category, gattung, created_at, updated_at
) VALUES(?, ?, 'Roco', 'H0', 'Wagen', 'Reisezugwagen', ?, ?)`, id, id, createdAt, createdAt); err != nil {
		t.Fatal(err)
	}
}
