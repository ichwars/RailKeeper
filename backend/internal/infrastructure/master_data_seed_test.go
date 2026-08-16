package infrastructure_test

import (
	"database/sql"
	"path/filepath"
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/infrastructure"
)

func TestSeedMasterDataLoadsGeneratedSeed(t *testing.T) {
	db, err := infrastructure.OpenSQLite(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	migrationsDir := filepath.Join("..", "..", "migrations")
	if err := infrastructure.Migrate(db, migrationsDir); err != nil {
		t.Fatal(err)
	}

	seedsDir := filepath.Join("..", "..", "seeds")
	if err := infrastructure.SeedMasterData(db, seedsDir); err != nil {
		t.Fatal(err)
	}

	var manufacturers int
	if err := db.QueryRow(`SELECT COUNT(*) FROM master_data_entries WHERE type='manufacturer'`).Scan(&manufacturers); err != nil {
		t.Fatal(err)
	}
	if manufacturers < 500 {
		t.Fatalf("expected generated manufacturer seed, got %d entries", manufacturers)
	}

	var relations int
	if err := db.QueryRow(`SELECT COUNT(*) FROM master_data_relations WHERE parent_type='vehicle_category' AND child_type='vehicle_gattung'`).Scan(&relations); err != nil {
		t.Fatal(err)
	}
	if relations == 0 {
		t.Fatal("expected category to gattung relations")
	}

	all, err := application.NewMasterDataService(db).ListAll(t.Context(), false)
	if err != nil {
		t.Fatal(err)
	}
	if len(all["manufacturer"]) < 500 || len(all["vehicle_category"]) != 9 || len(all["epoch"]) != 6 {
		t.Fatalf("unexpected bundled master data counts: %#v", map[string]int{
			"manufacturer":     len(all["manufacturer"]),
			"vehicle_category": len(all["vehicle_category"]),
			"epoch":            len(all["epoch"]),
		})
	}
}

func TestSeedMasterDataPreservesBundledEditsAndCustomOrigin(t *testing.T) {
	db, err := infrastructure.OpenSQLite(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := infrastructure.Migrate(db, filepath.Join("..", "..", "migrations")); err != nil {
		t.Fatal(err)
	}
	seedsDir := filepath.Join("..", "..", "seeds")
	if err := infrastructure.SeedMasterData(db, seedsDir); err != nil {
		t.Fatal(err)
	}

	if _, err := db.Exec(`
UPDATE master_data_entries
SET label='Eigener Name', active=0, sort_order=42,
    source_url='https://example.test/custom', metadata_json='{"note":"custom"}'
WHERE type='railway_company' AND key='badstb'`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`
INSERT INTO master_data_entries(
  id, type, key, label, active, sort_order, source_url, metadata_json,
  created_at, updated_at, origin
) VALUES(
  'manufacturer:club', 'manufacturer', 'club', 'Club', 1, 7,
  '', '{}', 'now', 'now', 'custom'
)`); err != nil {
		t.Fatal(err)
	}

	if err := infrastructure.SeedMasterData(db, seedsDir); err != nil {
		t.Fatal(err)
	}

	assertSeedEntryState(t, db, "railway_company", "badstb", seedEntryState{
		label:        "Eigener Name",
		active:       0,
		sortOrder:    42,
		sourceURL:    "https://example.test/custom",
		metadataJSON: `{"note":"custom"}`,
		origin:       "bundled",
	})
	assertSeedEntryState(t, db, "manufacturer", "club", seedEntryState{
		label:        "Club",
		active:       1,
		sortOrder:    7,
		sourceURL:    "",
		metadataJSON: `{}`,
		origin:       "custom",
	})
}

type seedEntryState struct {
	label        string
	active       int
	sortOrder    int
	sourceURL    string
	metadataJSON string
	origin       string
}

func assertSeedEntryState(t *testing.T, db *sql.DB, typeName, key string, want seedEntryState) {
	t.Helper()
	var got seedEntryState
	if err := db.QueryRow(`
SELECT label, active, sort_order, source_url, metadata_json, origin
FROM master_data_entries WHERE type=? AND key=?`, typeName, key).Scan(
		&got.label,
		&got.active,
		&got.sortOrder,
		&got.sourceURL,
		&got.metadataJSON,
		&got.origin,
	); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("master data %s/%s: got %#v, want %#v", typeName, key, got, want)
	}
}
