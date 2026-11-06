package application_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/infrastructure"
	"railkeeper/backend/migrations"
)

func TestVehicleCatalogCanonicalProjectionKeepsLegacyWritesScopedAndRevisionsImmutable(t *testing.T) {
	db := testDB(t)
	service := application.NewVehicleService(db)
	input := application.CreateVehicleInput{Manufacturer: "Piko", Name: "Original", Gauge: "H0", Category: "Lokomotive", Gattung: "Diesellok", PurchasePrice: "140.50", Condition: "Used"}
	first, err := service.Create(t.Context(), input, "actor")
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.Create(t.Context(), input, "actor")
	if err != nil {
		t.Fatal(err)
	}
	var revision, beforeFields, beforeClock string
	if err := db.QueryRow(`SELECT l.revision_id,r.fields_json,l.updated_at FROM vehicle_article_links l JOIN catalog_article_revisions r ON r.id=l.revision_id WHERE l.vehicle_id=?`, first.ID).Scan(&revision, &beforeFields, &beforeClock); err != nil {
		t.Fatal(err)
	}
	input.Name = "Own label"
	if _, err := service.Update(t.Context(), first.ID, input, "actor"); err != nil {
		t.Fatal(err)
	}
	var raw, afterRevision, afterFields, authority string
	if err := db.QueryRow(`SELECT l.revision_id,r.fields_json,l.overrides_json,l.authority FROM vehicle_article_links l JOIN catalog_article_revisions r ON r.id=l.revision_id WHERE l.vehicle_id=?`, first.ID).Scan(&afterRevision, &afterFields, &raw, &authority); err != nil {
		t.Fatal(err)
	}
	if revision != afterRevision || beforeFields != afterFields || authority != "catalog" {
		t.Fatal("legacy save replaced its pinned revision")
	}
	var overrides application.VehicleCatalogFields
	if err := json.Unmarshal([]byte(raw), &overrides); err != nil || !reflect.DeepEqual(overrides, application.VehicleCatalogFields{"name": "Own label"}) {
		t.Fatalf("overrides=%s / %v", raw, err)
	}
	stored, err := service.Get(t.Context(), first.ID)
	if err != nil || stored.PurchasePrice != "140.50" || stored.Condition != "Used" {
		t.Fatal("catalog change modified ownership")
	}
	other, err := service.Get(t.Context(), second.ID)
	if err != nil || other.Name != "Original" {
		t.Fatal("catalog update changed another owned vehicle")
	}
	var consistent bool
	if err := db.QueryRow(`SELECT consistent FROM vehicle_catalog_consistency WHERE entity_id=?`, first.ID).Scan(&consistent); err != nil || !consistent {
		t.Fatal("canonical and compatibility data differ")
	}
	var clock string
	if err := db.QueryRow(`SELECT updated_at FROM vehicle_article_links WHERE vehicle_id=?`, first.ID).Scan(&clock); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE vehicles SET article_source_url=NULL, purchase_price='150.00' WHERE id=?`, first.ID); err != nil {
		t.Fatal(err)
	}
	var unchangedClock string
	if err := db.QueryRow(`SELECT updated_at FROM vehicle_article_links WHERE vehicle_id=?`, first.ID).Scan(&unchangedClock); err != nil || clock != unchangedClock {
		t.Fatal("ownership or NULL/empty equivalence rewrote catalog metadata")
	}
}

func TestVehicleCatalogAuthorityMigrationPreservesPendingDataAndItsOriginalMarker(t *testing.T) {
	root := t.TempDir()
	folder := filepath.Join(root, "migrations")
	if err := os.Mkdir(folder, 0700); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join("..", "..", "migrations")
	entries, err := os.ReadDir(source)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.Name() > "0074_vehicle_catalog_legacy.sql" || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		body, err := os.ReadFile(filepath.Join(source, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(folder, entry.Name()), body, 0600); err != nil {
			t.Fatal(err)
		}
	}
	db, err := infrastructure.OpenSQLite(filepath.Join(root, "data"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := infrastructure.Migrate(db, folder); err != nil {
		t.Fatal(err)
	}
	service := application.NewVehicleService(db)
	input := application.CreateVehicleInput{Manufacturer: "Piko", Name: "Saved legacy", Gauge: "H0", Category: "Lokomotive", Gattung: "Diesellok"}
	vehicle, err := service.Create(t.Context(), input, "actor")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.BackfillLegacyCatalog(t.Context(), "actor"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE vehicle_article_links SET authority='catalog', overrides_json='{"name":"Pending proposal"}' WHERE vehicle_id=?`, vehicle.ID); err != nil {
		t.Fatal(err)
	}
	var originalRevision, originalOverrides, originalClock string
	if err := db.QueryRow(`SELECT revision_id,overrides_json,updated_at FROM vehicle_article_links WHERE vehicle_id=?`, vehicle.ID).Scan(&originalRevision, &originalOverrides, &originalClock); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(migrations.VehicleCatalogAuthoritySQL); err != nil {
		t.Fatal(err)
	}
	var authority, prior, revision, overrides, clock string
	if err := db.QueryRow(`SELECT authority,source_authority,revision_id,overrides_json,updated_at FROM vehicle_article_links WHERE vehicle_id=?`, vehicle.ID).Scan(&authority, &prior, &revision, &overrides, &clock); err != nil {
		t.Fatal(err)
	}
	if authority != "legacy" || prior != "catalog" || revision != originalRevision || overrides != originalOverrides || clock != originalClock {
		t.Fatal("migration discarded pending catalog data or its provenance")
	}
	stored, err := service.Get(t.Context(), vehicle.ID)
	if err != nil || stored.Name != "Saved legacy" {
		t.Fatal("migration silently applied a pending catalog value")
	}
	report, err := service.CatalogBackfillReport(t.Context())
	if err != nil || report.DifferentRepresentations != 1 {
		t.Fatal("unresolved catalog data is not visible")
	}
	input.Name = "Another owned label"
	if _, err := service.Update(t.Context(), vehicle.ID, input, "actor"); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT authority,source_authority,revision_id,overrides_json,updated_at FROM vehicle_article_links WHERE vehicle_id=?`, vehicle.ID).Scan(&authority, &prior, &revision, &overrides, &clock); err != nil {
		t.Fatal(err)
	}
	if authority != "legacy" || prior != "catalog" || revision != originalRevision || overrides != originalOverrides || clock != originalClock {
		t.Fatal("legacy edit replaced an unresolved catalog proposal")
	}
}

func TestVehicleCatalogNewSetSharesRevisionWithoutPropagatingLaterSetEdits(t *testing.T) {
	db := testDB(t)
	service := application.NewVehicleService(db)
	input := application.VehicleSetInput{Manufacturer: "Roco", Name: "Set", Gauge: "H0", Category: "Wagen", Gattung: "Reisezugwagen", PurchasePrice: "180"}
	set, err := service.CreateSet(t.Context(), application.CreateVehicleSetInput{
		Set: input, Members: []application.CreateVehicleInput{{Name: "A"}, {Name: "B"}},
	}, "actor")
	if err != nil {
		t.Fatal(err)
	}
	var revision string
	if err := db.QueryRow(`SELECT revision_id FROM vehicle_set_article_links WHERE vehicle_set_id=?`, set.ID).Scan(&revision); err != nil {
		t.Fatal(err)
	}
	for _, member := range set.Members {
		var pin, raw string
		if err := db.QueryRow(`SELECT revision_id,overrides_json FROM vehicle_article_links WHERE vehicle_id=?`, member.ID).Scan(&pin, &raw); err != nil {
			t.Fatal(err)
		}
		var overrides application.VehicleCatalogFields
		if err := json.Unmarshal([]byte(raw), &overrides); err != nil || pin != revision || overrides["name"] != member.Name {
			t.Fatal("new set member lost its shared base or individual designation")
		}
	}
	input.Name = "Corrected set only"
	if _, err := service.UpdateSet(t.Context(), set.ID, input, "actor"); err != nil {
		t.Fatal(err)
	}
	for _, member := range set.Members {
		stored, err := service.Get(t.Context(), member.ID)
		if err != nil || stored.Name != member.Name {
			t.Fatal("set edit silently changed an owned member")
		}
	}
}

func TestVehicleCatalogRestoreFailureRollsBackDataAndProjectionDefinitions(t *testing.T) {
	db := testDB(t)
	service := application.NewVehicleService(db)
	vehicle, err := service.Create(t.Context(), application.CreateVehicleInput{Manufacturer: "Piko", Name: "Kept", Gauge: "H0", Category: "Lokomotive", Gattung: "Diesellok"}, "actor")
	if err != nil {
		t.Fatal(err)
	}
	backup := application.NewBackupService(db, t.TempDir())
	before, err := backup.Export(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TRIGGER reject_restore_catalog BEFORE INSERT ON vehicles BEGIN SELECT RAISE(ABORT,'injected restore failure'); END`); err != nil {
		t.Fatal(err)
	}
	if _, err := backup.Import(t.Context(), before); err == nil {
		t.Fatal("injected failure was accepted")
	}
	after, err := backup.Export(t.Context())
	if err != nil || !reflect.DeepEqual(catalogBackupTableSnapshot(t, before), catalogBackupTableSnapshot(t, after)) {
		t.Fatal("restore failure changed data")
	}
	if _, err := db.Exec(`DROP TRIGGER reject_restore_catalog`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE vehicles SET name='After failure' WHERE id=?`, vehicle.ID); err != nil {
		t.Fatal(err)
	}
	var consistent bool
	if err := db.QueryRow(`SELECT consistent FROM vehicle_catalog_consistency WHERE entity_id=?`, vehicle.ID).Scan(&consistent); err != nil || !consistent {
		t.Fatal("failed restore left projection disabled")
	}
	var triggers int
	if err := db.QueryRow(`SELECT COUNT(*) FROM sqlite_schema WHERE type='trigger' AND name IN ('catalog_vehicle_create','catalog_vehicle_update','catalog_vehicle_link','catalog_set_create','catalog_set_update','catalog_set_link')`).Scan(&triggers); err != nil || triggers != 6 {
		t.Fatal("projection definitions were not restored atomically")
	}
}
