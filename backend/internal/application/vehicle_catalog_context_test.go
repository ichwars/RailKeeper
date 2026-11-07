package application_test

import (
	"errors"
	"reflect"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestVehicleCatalogContextShowsPendingAndPinnedValuesWithoutChangingInventory(t *testing.T) {
	db := testDB(t)
	service := application.NewVehicleService(db)
	vehicle, err := service.Create(t.Context(), application.CreateVehicleInput{Manufacturer: "Piko", Name: "Owned label", Gauge: "H0", Category: "Lokomotive", Gattung: "Diesellok"}, "actor")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE vehicle_article_links SET authority='legacy',source_authority='catalog',overrides_json='{"name":"Pending proposal"}' WHERE vehicle_id=?`, vehicle.ID); err != nil {
		t.Fatal(err)
	}
	backup := application.NewBackupService(db, t.TempDir())
	before, err := backup.Export(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	state, err := service.CatalogContext(t.Context(), "vehicle", vehicle.ID)
	if err != nil || state.Pinned == nil || state.CurrentFields["name"] != "Owned label" || state.Pinned.Fields["name"] != "Owned label" || state.Overrides["name"] != "Pending proposal" || state.Authority != "legacy" || state.SourceAuthority == nil || *state.SourceAuthority != "catalog" || len(state.Revisions) != 1 || len(state.Revision) != 64 {
		t.Fatalf("context: %#v / %v", state, err)
	}
	stored, err := service.Get(t.Context(), vehicle.ID)
	if err != nil || stored.Name != "Owned label" {
		t.Fatal("reading catalog applied a pending proposal")
	}
	after, err := backup.Export(t.Context())
	if err != nil || !reflect.DeepEqual(catalogBackupTableSnapshot(t, before), catalogBackupTableSnapshot(t, after)) {
		t.Fatal("reading catalog changed saved data")
	}
	if _, err := db.Exec(`DELETE FROM vehicle_article_links WHERE vehicle_id=?`, vehicle.ID); err != nil {
		t.Fatal(err)
	}
	state, err = service.CatalogContext(t.Context(), "vehicle", vehicle.ID)
	if err != nil || state.Authority != "unlinked" || state.Pinned != nil || len(state.Revisions) != 0 || len(state.Overrides) != 0 {
		t.Fatal("reading unlinked inventory created a catalog assignment")
	}
	if _, err := service.CatalogContext(t.Context(), "vehicle", "missing"); !errors.Is(err, application.ErrVehicleCatalogNotFound) {
		t.Fatal(err)
	}
	if _, err := service.CatalogContext(t.Context(), "unexpected-table", vehicle.ID); !errors.Is(err, application.ErrVehicleCatalogValidation) {
		t.Fatal(err)
	}
}

func TestVehicleCatalogContextGuardsAssignmentChangesEvenWhenVehicleFieldsStayIdentical(t *testing.T) {
	db := testDB(t)
	service := application.NewVehicleService(db)
	input := application.CreateVehicleInput{Manufacturer: "Piko", Name: "Same", Gauge: "H0", Category: "Lokomotive", Gattung: "Diesellok"}
	vehicle, err := service.Create(t.Context(), input, "actor")
	if err != nil {
		t.Fatal(err)
	}
	before, err := service.CatalogContext(t.Context(), "vehicle", vehicle.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE vehicles SET purchase_price='150' WHERE id=?`, vehicle.ID); err != nil {
		t.Fatal(err)
	}
	unchanged, err := service.CatalogContext(t.Context(), "vehicle", vehicle.ID)
	if err != nil || unchanged.Revision != before.Revision {
		t.Fatal("ownership-only edit invalidated the independent catalog comparison")
	}
	if _, err := db.Exec(`INSERT INTO catalog_article_revisions(id,article_id,revision_number,fields_json,note,created_at)
SELECT 'identical-revision',article_id,2,fields_json,'Same values, distinct revision',created_at FROM catalog_article_revisions WHERE id=?`, before.Pinned.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE vehicle_article_links SET revision_id='identical-revision' WHERE vehicle_id=?`, vehicle.ID); err != nil {
		t.Fatal(err)
	}
	after, err := service.CatalogContext(t.Context(), "vehicle", vehicle.ID)
	if err != nil || after.Revision == before.Revision || after.Pinned.ID != "identical-revision" || !reflect.DeepEqual(after.CurrentFields, before.CurrentFields) || len(after.Revisions) != 2 || after.Revisions[0].RevisionNumber != 2 {
		t.Fatal("comparison failed to distinguish catalog identities with equal vehicle values")
	}
}
