package infrastructure_test

import (
	"os"
	"path/filepath"
	"testing"

	"railkeeper/backend/internal/infrastructure"
)

func TestAccessoryListPriceMigrationPreservesExistingRows(t *testing.T) {
	root := t.TempDir()
	migrationsDir := filepath.Join(root, "migrations")
	if err := os.Mkdir(migrationsDir, 0700); err != nil {
		t.Fatal(err)
	}
	copyMigrationsThrough(t, filepath.Join("..", "..", "migrations"), migrationsDir,
		"0056_vehicle_operational_fields.sql")

	db, err := infrastructure.OpenSQLite(filepath.Join(root, "data"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := infrastructure.Migrate(db, migrationsDir); err != nil {
		t.Fatal(err)
	}

	if _, err := db.Exec(`
INSERT INTO accessory_products(
  id, inventory_number, manufacturer, name, category, tracking_mode, article_type, subtype,
  gauges_json, package_quantity, stock_unit, minimum_stock, inventory_strategy, created_at, updated_at
) VALUES(
  'product-1', 'RK-ART-000001', 'Tillig', 'Gleis', 'Track', 'quantity', 'track', 'track:straight',
  '[]', 1, 'piece', 0, 'quantity', 'now', 'now'
)`); err != nil {
		t.Fatal(err)
	}

	applyMigrationFile(t, db, "0057_accessory_list_price.sql")

	var listPrice string
	if err := db.QueryRow(`SELECT list_price FROM accessory_products WHERE id='product-1'`).Scan(&listPrice); err != nil {
		t.Fatal(err)
	}
	if listPrice != "" {
		t.Fatalf("unexpected migrated list price %q", listPrice)
	}
}
