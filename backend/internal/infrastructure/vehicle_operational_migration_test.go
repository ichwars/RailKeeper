package infrastructure_test

import (
	"os"
	"path/filepath"
	"testing"

	"railkeeper/backend/internal/infrastructure"
)

func TestVehicleOperationalMigrationPreservesExistingRows(t *testing.T) {
	root := t.TempDir()
	migrationsDir := filepath.Join(root, "migrations")
	if err := os.Mkdir(migrationsDir, 0700); err != nil {
		t.Fatal(err)
	}
	copyMigrationsThrough(t, filepath.Join("..", "..", "migrations"), migrationsDir,
		"0055_track_library_exchange.sql")

	db, err := infrastructure.OpenSQLite(filepath.Join(root, "data"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := infrastructure.Migrate(db, migrationsDir); err != nil {
		t.Fatal(err)
	}

	if _, err := db.Exec(`
INSERT INTO vehicles(id, inventory_number, manufacturer, name, gauge, category, created_at, updated_at)
VALUES('vehicle-1', 'RK-LOK-000001', 'Piko', 'BR 118', 'H0', 'Lokomotive', 'now', 'now')`); err != nil {
		t.Fatal(err)
	}

	applyMigrationFile(t, db, "0056_vehicle_operational_fields.sql")

	var maximumSpeed *int
	var homeBase string
	if err := db.QueryRow(`
SELECT maximum_speed_kmh, home_base FROM vehicles WHERE id='vehicle-1'`).Scan(&maximumSpeed, &homeBase); err != nil {
		t.Fatal(err)
	}
	if maximumSpeed != nil || homeBase != "" {
		t.Fatalf("unexpected migrated defaults: speed=%v homeBase=%q", maximumSpeed, homeBase)
	}
	if _, err := db.Exec(`UPDATE vehicles SET maximum_speed_kmh=0 WHERE id='vehicle-1'`); err == nil {
		t.Fatal("expected migration check constraint to reject zero speed")
	}
}
