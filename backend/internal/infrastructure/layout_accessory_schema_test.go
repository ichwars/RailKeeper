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

func TestLayoutAccessorySchemaCreatesArticleManagementTables(t *testing.T) {
	db, err := infrastructure.OpenSQLite(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := infrastructure.Migrate(db, filepath.Join("..", "..", "migrations")); err != nil {
		t.Fatal(err)
	}

	for _, table := range []string{
		"accessory_product_attributes", "accessory_stock_movements", "accessory_purchases", "accessory_documents",
		"accessory_installation_condition_history",
	} {
		var count int
		if err := db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 1 {
			t.Fatalf("missing article management table %q", table)
		}
	}

	for _, index := range []string{
		"ix_accessory_products_article_lookup", "ix_accessory_products_article_type",
		"ix_accessory_products_archived", "ix_accessory_products_ean",
		"ux_accessory_documents_primary_image",
	} {
		var count int
		if err := db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name=?`, index).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 1 {
			t.Fatalf("missing article management index %q", index)
		}
	}
	var legacyArticleIndex int
	if err := db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name='ux_accessory_products_article'`).
		Scan(&legacyArticleIndex); err != nil {
		t.Fatal(err)
	}
	if legacyArticleIndex != 0 {
		t.Fatal("legacy unique article index still exists")
	}

	if _, err := db.Exec(`INSERT INTO accessory_products(
  id, manufacturer, name, category, tracking_mode, created_at, updated_at
) VALUES('article-2', 'Tillig', 'Gleis', 'track', 'quantity', 'now', 'now')`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO storage_locations(id, name, created_at, updated_at)
VALUES('location-2', 'Werkstatt', 'now', 'now')`); err != nil {
		t.Fatal(err)
	}
	assertSchemaColumn(t, db, "accessory_product_attributes", "unit")
	for _, column := range []string{
		"file_name", "original_name", "description", "category", "mime_type", "size_bytes",
		"is_primary", "created_by", "created_at", "updated_at",
	} {
		assertSchemaColumn(t, db, "accessory_documents", column)
	}
	if _, err := db.Exec(`INSERT INTO accessory_product_attributes(
  id, product_id, attribute_key, value_type, number_value, unit, created_at, updated_at
) VALUES('attribute-with-unit', 'article-2', 'lengthMm', 'number', 120, 'mm', 'now', 'now')`); err != nil {
		t.Fatalf("number attribute unit must be nullable text: %v", err)
	}

	statements := []string{
		`INSERT INTO accessory_products(
           id, manufacturer, name, category, tracking_mode, inventory_strategy, created_at, updated_at
         ) VALUES('article-1', 'Tillig', 'Gleis', 'track', 'quantity', 'unsupported', 'now', 'now')`,
		`INSERT INTO accessory_product_attributes(
           id, product_id, attribute_key, value_type, text_value, number_value, created_at, updated_at
         ) VALUES('attribute-1', 'article-2', 'lengthMm', 'text', '120', 120, 'now', 'now')`,
		`INSERT INTO accessory_product_attributes(
           id, product_id, attribute_key, value_type, text_value, unit, created_at, updated_at
         ) VALUES('attribute-unit-text', 'article-2', 'trackSystem', 'text', 'TT', 'mm', 'now', 'now')`,
		`INSERT INTO accessory_stock_movements(
           id, product_id, location_id, movement_type, quantity, created_at
         ) VALUES('movement-1', 'article-2', 'location-2', 'unsupported', 1, 'now')`,
	}
	for _, statement := range statements {
		expectConstraintFailure(t, db, statement)
	}

	assertMasterDataCount(t, db, "article_type", 8)
	assertMasterDataCount(t, db, "article_subtype", 54)
	assertMasterDataCount(t, db, "stock_unit", 5)
	assertMasterDataCount(t, db, "controlled_field_kind", 6)

	assertSchemaColumn(t, db, "accessory_assets", "purchase_id")
	for _, column := range []string{"currency", "invoice_number", "warranty_until", "booked_to_stock"} {
		assertSchemaColumn(t, db, "accessory_purchases", column)
	}
	assertSchemaMissingColumn(t, db, "accessory_purchases", "order_number")
	expectConstraintFailure(t, db, `INSERT INTO accessory_purchases(
  id, product_id, quantity, purchased_at, booked_to_stock, created_at, updated_at
) VALUES('invalid-booking-state', 'article-2', 1, '2026-08-01', 2, 'now', 'now')`)
	for _, table := range []string{"accessory_reservations", "accessory_installations"} {
		for _, column := range []string{"placement", "digital_address", "decoder_output", "connection", "wiring_notes"} {
			assertSchemaColumn(t, db, table, column)
		}
	}
}

func assertSchemaMissingColumn(t *testing.T, db *sql.DB, table, unwanted string) {
	t.Helper()
	rows, err := db.Query(`SELECT name FROM pragma_table_info(?)`, table)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var got string
		if err := rows.Scan(&got); err != nil {
			t.Fatal(err)
		}
		if got == unwanted {
			t.Fatalf("unexpected column %q on table %q", unwanted, table)
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
}

func assertMasterDataCount(t *testing.T, db *sql.DB, entryType string, want int) {
	t.Helper()
	var got int
	if err := db.QueryRow(`SELECT COUNT(*) FROM master_data_entries WHERE type=?`, entryType).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("master data count for %s: got %d, want %d", entryType, got, want)
	}
}

func assertSchemaColumn(t *testing.T, db *sql.DB, table, want string) {
	t.Helper()
	rows, err := db.Query(`SELECT name FROM pragma_table_info(?)`, table)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var got string
		if err := rows.Scan(&got); err != nil {
			t.Fatal(err)
		}
		if got == want {
			return
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	t.Fatalf("missing column %s.%s", table, want)
}

func expectConstraintFailure(t *testing.T, db *sql.DB, statement string) {
	t.Helper()
	if _, err := db.Exec(statement); err == nil {
		t.Fatalf("expected database constraint failure for %s", statement)
	}
}
