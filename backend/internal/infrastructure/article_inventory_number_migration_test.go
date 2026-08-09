package infrastructure_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"railkeeper/backend/internal/infrastructure"
)

func TestArticleInventoryNumberMigrationBackfillsDeterministically(t *testing.T) {
	db := migratedArticleInventoryNumberTestDB(t)
	seedArticleBeforeInventoryNumberMigration(t, db, "article-b", "2026-01-02")
	seedArticleBeforeInventoryNumberMigration(t, db, "article-a", "2026-01-01")

	applyMigrationFile(t, db, "0043_article_inventory_numbers.sql")

	assertText(t, db, `SELECT inventory_number FROM accessory_products WHERE id='article-a'`, "RK-ART-000001")
	assertText(t, db, `SELECT inventory_number FROM accessory_products WHERE id='article-b'`, "RK-ART-000002")
	assertText(t, db, `SELECT CAST(next_number AS TEXT) FROM inventory_number_schemes WHERE category='Artikel'`, "3")
	expectConstraintFailure(t, db, `UPDATE accessory_products SET inventory_number='' WHERE id='article-a'`)
	expectConstraintFailure(t, db, `UPDATE accessory_products SET inventory_number='RK-ART-000001' WHERE id='article-b'`)
	assertForeignKeyCheck(t, db)
}

func TestArticleInventoryNumberMigrationUsesExistingScheme(t *testing.T) {
	db := migratedArticleInventoryNumberTestDB(t)
	if _, err := db.Exec(`
INSERT INTO inventory_number_schemes(
  id, category, prefix, next_number, padding, active, created_at, updated_at
) VALUES('article-scheme', 'Artikel', 'ART', 7, 3, 1, '2026-01-01', '2026-01-01')`); err != nil {
		t.Fatal(err)
	}
	seedArticleBeforeInventoryNumberMigration(t, db, "article-custom", "2026-01-01")

	applyMigrationFile(t, db, "0043_article_inventory_numbers.sql")

	assertText(t, db, `SELECT inventory_number FROM accessory_products WHERE id='article-custom'`, "ART-007")
	assertText(t, db, `SELECT CAST(next_number AS TEXT) FROM inventory_number_schemes WHERE category='Artikel'`, "8")
}

func migratedArticleInventoryNumberTestDB(t *testing.T) *sql.DB {
	t.Helper()
	root := t.TempDir()
	migrationsDir := filepath.Join(root, "migrations")
	if err := os.Mkdir(migrationsDir, 0700); err != nil {
		t.Fatal(err)
	}
	copyMigrationsThrough(t, filepath.Join("..", "..", "migrations"), migrationsDir,
		"0042_normalize_accessory_subtype_master_data.sql")

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

func seedArticleBeforeInventoryNumberMigration(t *testing.T, db *sql.DB, id, createdAt string) {
	t.Helper()
	if _, err := db.Exec(`
INSERT INTO accessory_products(
  id, manufacturer, article_number, name, category, tracking_mode, article_type, subtype,
  gauges_json, package_quantity, stock_unit, minimum_stock, inventory_strategy, created_at, updated_at
) VALUES(?, 'Tillig', ?, ?, 'track:straight', 'quantity', 'track', 'track:straight',
  '[]', 1, 'piece', 0, 'quantity', ?, ?)`, id, id, id, createdAt, createdAt); err != nil {
		t.Fatal(err)
	}
}
