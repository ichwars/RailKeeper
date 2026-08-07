package infrastructure_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"railkeeper/backend/internal/infrastructure"
)

func TestArticleManagementMigrationPreservesAccessoryFoundation(t *testing.T) {
	db := migratedTestDB(t, 40)
	seedLegacyAccessoryGraph(t, db)
	applyMigrationFile(t, db, "0041_article_management_redesign.sql")

	assertRowCount(t, db, "accessory_products", 2)
	assertText(t, db, `SELECT inventory_strategy FROM accessory_products WHERE id='quantity-product'`, "quantity")
	assertText(t, db, `SELECT inventory_strategy FROM accessory_products WHERE id='individual-product'`, "individual")
	assertText(t, db, `SELECT product_id || ':' || location_id FROM accessory_stock`, "quantity-product:location-1")
	assertText(t, db, `SELECT product_id FROM accessory_assets WHERE id='asset-1'`, "individual-product")
	assertText(t, db, `SELECT product_id || ':' || location_id || ':' || layout_id
FROM accessory_reservations WHERE id='reservation-1'`, "quantity-product:location-1:layout-1")
	assertText(t, db, `SELECT product_id || ':' || asset_id || ':' || source_location_id || ':' || layout_id
FROM accessory_installations WHERE id='installation-1'`, "individual-product:asset-1:location-1:layout-1")
	assertForeignKeyCheck(t, db)
}

func migratedTestDB(t *testing.T, lastMigration int) *sql.DB {
	t.Helper()
	root := t.TempDir()
	migrationsDir := filepath.Join(root, "migrations")
	if err := os.Mkdir(migrationsDir, 0700); err != nil {
		t.Fatal(err)
	}
	copyMigrationsThrough(t, filepath.Join("..", "..", "migrations"), migrationsDir,
		"0040_accessory_allocation_targets.sql")

	db, err := infrastructure.OpenSQLite(filepath.Join(root, "data"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := infrastructure.Migrate(db, migrationsDir); err != nil {
		t.Fatal(err)
	}
	if lastMigration != 40 {
		t.Fatalf("unsupported migration version %d", lastMigration)
	}
	return db
}

func seedLegacyAccessoryGraph(t *testing.T, db *sql.DB) {
	t.Helper()
	statements := []string{
		`INSERT INTO storage_locations(id, name, created_at, updated_at)
         VALUES('location-1', 'Lager', 'now', 'now')`,
		`INSERT INTO accessory_products(
           id, manufacturer, article_number, name, category, tracking_mode, created_at, updated_at
         ) VALUES('quantity-product', 'Tillig', '83101', 'Gerades Gleis', 'track', 'quantity', 'now', 'now')`,
		`INSERT INTO accessory_products(
           id, manufacturer, article_number, name, category, tracking_mode, created_at, updated_at
         ) VALUES('individual-product', 'ESU', '51800', 'Decoder', 'decoder', 'individual', 'now', 'now')`,
		`INSERT INTO accessory_stock(product_id, location_id, quantity, updated_at)
         VALUES('quantity-product', 'location-1', 4, 'now')`,
		`INSERT INTO accessory_assets(
           id, product_id, storage_location_id, created_at, updated_at
         ) VALUES('asset-1', 'individual-product', 'location-1', 'now', 'now')`,
		`INSERT INTO layouts(id, name, kind, gauge, scale, created_at, updated_at)
         VALUES('layout-1', 'Heimanlage', 'private', 'TT', '1:120', 'now', 'now')`,
		`INSERT INTO accessory_reservations(
           id, product_id, location_id, quantity, layout_id, status, created_by, created_at, updated_at
         ) VALUES('reservation-1', 'quantity-product', 'location-1', 1, 'layout-1', 'active', 'planner-1', 'now', 'now')`,
		`INSERT INTO accessory_installations(
           id, product_id, asset_id, source_location_id, quantity, layout_id, condition_state, installed_by, installed_at
         ) VALUES('installation-1', 'individual-product', 'asset-1', 'location-1', 1, 'layout-1', 'ready', 'planner-1', 'now')`,
	}
	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			t.Fatalf("seed legacy accessory graph: %v", err)
		}
	}
}

func applyMigrationFile(t *testing.T, db *sql.DB, file string) {
	t.Helper()
	body, err := os.ReadFile(filepath.Join("..", "..", "migrations", file))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(string(body)); err != nil {
		t.Fatal(err)
	}
}

func assertRowCount(t *testing.T, db *sql.DB, table string, want int) {
	t.Helper()
	var got int
	if err := db.QueryRow(`SELECT COUNT(*) FROM ` + table).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("row count for %s: got %d, want %d", table, got, want)
	}
}

func assertText(t *testing.T, db *sql.DB, query, want string) {
	t.Helper()
	var got string
	if err := db.QueryRow(query).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("query result: got %q, want %q", got, want)
	}
}

func assertForeignKeyCheck(t *testing.T, db *sql.DB) {
	t.Helper()
	rows, err := db.Query(`PRAGMA foreign_key_check`)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	if rows.Next() {
		t.Fatal("foreign key check found violations")
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
}
