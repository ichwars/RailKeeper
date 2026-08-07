package infrastructure_test

import (
	"database/sql"
	"path/filepath"
	"slices"
	"testing"

	"railkeeper/backend/internal/infrastructure"
)

func TestLayoutAccessoryMigrationCreatesFoundationTables(t *testing.T) {
	db, err := infrastructure.OpenSQLite(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	migrationsDir := filepath.Join("..", "..", "migrations")
	if err := infrastructure.Migrate(db, migrationsDir); err != nil {
		t.Fatal(err)
	}
	if err := infrastructure.Migrate(db, migrationsDir); err != nil {
		t.Fatalf("migration must be idempotent: %v", err)
	}

	rows, err := db.Query(`
SELECT name
FROM sqlite_master
WHERE type='table' AND name IN (
  'storage_locations', 'accessory_products', 'accessory_stock', 'accessory_assets',
  'layouts', 'layout_units', 'plan_variants', 'plan_revisions', 'layout_configurations',
  'layout_configuration_units', 'accessory_reservations', 'accessory_installations'
)
ORDER BY name`)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			t.Fatal(err)
		}
		tables = append(tables, table)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}

	want := []string{
		"accessory_assets", "accessory_installations", "accessory_products",
		"accessory_reservations", "accessory_stock", "layout_configuration_units",
		"layout_configurations", "layout_units", "layouts", "plan_revisions", "plan_variants",
		"storage_locations",
	}
	if !slices.Equal(tables, want) {
		t.Fatalf("unexpected foundation tables: got %v, want %v", tables, want)
	}
}

func TestLayoutAccessoryMigrationEnforcesDomainConstraints(t *testing.T) {
	db, err := infrastructure.OpenSQLite(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := infrastructure.Migrate(db, filepath.Join("..", "..", "migrations")); err != nil {
		t.Fatal(err)
	}

	expectConstraintFailure(t, db, `
INSERT INTO accessory_products(id, manufacturer, name, category, tracking_mode, created_at, updated_at)
VALUES('invalid-product', 'Tillig', 'Gleis', 'track', 'bulk', 'now', 'now')`)

	statements := []string{
		`INSERT INTO storage_locations(id, name, created_at, updated_at) VALUES('location-1', 'Lager', 'now', 'now')`,
		`INSERT INTO accessory_products(id, manufacturer, article_number, name, category, tracking_mode, created_at, updated_at)
         VALUES('product-1', 'Tillig', '83101', 'Gerades Gleis', 'track', 'quantity', 'now', 'now')`,
		`INSERT INTO layouts(id, name, kind, gauge, scale, created_at, updated_at)
         VALUES('layout-1', 'Clubanlage', 'club', 'TT', '1:120', 'now', 'now')`,
		`INSERT INTO layout_units(id, layout_id, name, kind, created_at, updated_at)
         VALUES('unit-1', 'layout-1', 'Bahnhof', 'module', 'now', 'now')`,
		`INSERT INTO plan_variants(id, layout_unit_id, name, created_at, updated_at)
         VALUES('variant-1', 'unit-1', 'Standard', 'now', 'now')`,
		`INSERT INTO plan_revisions(id, variant_id, revision_number, status, created_by, created_at, updated_at)
         VALUES('revision-1', 'variant-1', 1, 'published', 'user-1', 'now', 'now')`,
	}
	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			t.Fatalf("seed schema constraint test: %v", err)
		}
	}

	expectConstraintFailure(t, db, `
INSERT INTO accessory_stock(product_id, location_id, quantity, updated_at)
VALUES('product-1', 'location-1', -1, 'now')`)
	expectConstraintFailure(t, db, `
INSERT INTO plan_revisions(id, variant_id, revision_number, status, created_by, created_at, updated_at)
VALUES('revision-2', 'variant-1', 2, 'published', 'user-1', 'now', 'now')`)
	expectConstraintFailure(t, db, `
INSERT INTO accessory_reservations(
  id, product_id, location_id, quantity, status, created_by, created_at, updated_at
) VALUES('reservation-1', 'product-1', 'location-1', 1, 'active', 'user-1', 'now', 'now')`)
	expectConstraintFailure(t, db, `
INSERT INTO accessory_installations(
  id, product_id, source_location_id, quantity, layout_id, layout_unit_id,
  condition_state, installed_by, installed_at
) VALUES(
  'installation-1', 'product-1', 'location-1', 1, 'layout-1', 'unit-1',
  'ready', 'user-1', 'now'
)`)

	var plannerCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM roles WHERE name='Planner'`).Scan(&plannerCount); err != nil {
		t.Fatal(err)
	}
	if plannerCount != 1 {
		t.Fatalf("expected one Planner role, got %d", plannerCount)
	}
}

func expectConstraintFailure(t *testing.T, db *sql.DB, statement string) {
	t.Helper()
	if _, err := db.Exec(statement); err == nil {
		t.Fatalf("expected database constraint failure for %s", statement)
	}
}
