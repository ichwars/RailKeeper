package infrastructure_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"railkeeper/backend/internal/infrastructure"
)

func TestAccessoryAllocationTargetMigrationPreservesLegacyReservations(t *testing.T) {
	root := t.TempDir()
	partialMigrations := filepath.Join(root, "migrations-through-0039")
	if err := os.Mkdir(partialMigrations, 0700); err != nil {
		t.Fatal(err)
	}
	copyMigrationsThrough(t, filepath.Join("..", "..", "migrations"), partialMigrations,
		"0039_layout_accessory_foundation.sql")
	db, err := infrastructure.OpenSQLite(filepath.Join(root, "data"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := infrastructure.Migrate(db, partialMigrations); err != nil {
		t.Fatal(err)
	}
	statements := []string{
		`INSERT INTO storage_locations(id, name, created_at, updated_at)
         VALUES('location-1', 'Lager', 'now', 'now')`,
		`INSERT INTO accessory_products(
           id, manufacturer, name, category, tracking_mode, created_at, updated_at
         ) VALUES('product-1', 'Tillig', 'Gleis', 'Gleismaterial', 'quantity', 'now', 'now')`,
		`INSERT INTO layouts(id, name, kind, gauge, scale, created_at, updated_at)
         VALUES('layout-1', 'Heimanlage', 'private', 'TT', '1:120', 'now', 'now')`,
		`INSERT INTO accessory_reservations(
           id, product_id, location_id, quantity, layout_id, status, note, created_by, created_at, updated_at
         ) VALUES(
           'reservation-1', 'product-1', 'location-1', 1, 'layout-1', 'active',
           'legacy note', 'planner-1', 'now', 'now'
         )`,
	}
	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			t.Fatalf("seed legacy reservation: %v", err)
		}
	}
	if err := infrastructure.Migrate(db, filepath.Join("..", "..", "migrations")); err != nil {
		t.Fatal(err)
	}
	var productID, layoutID, vehicleID, note string
	if err := db.QueryRow(`
SELECT product_id, layout_id, COALESCE(vehicle_id, ''), note
FROM accessory_reservations WHERE id='reservation-1'`).
		Scan(&productID, &layoutID, &vehicleID, &note); err != nil {
		t.Fatal(err)
	}
	if productID != "product-1" || layoutID != "layout-1" || vehicleID != "" || note != "legacy note" {
		t.Fatalf("legacy reservation changed during migration: %q %q %q %q",
			productID, layoutID, vehicleID, note)
	}
}

func copyMigrationsThrough(t *testing.T, sourceDir, targetDir, lastFile string) {
	t.Helper()
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") || entry.Name() > lastFile {
			continue
		}
		body, err := os.ReadFile(filepath.Join(sourceDir, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(targetDir, entry.Name()), body, 0600); err != nil {
			t.Fatal(err)
		}
	}
}
