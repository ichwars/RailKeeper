package application_test

import (
	"encoding/json"
	"errors"
	"maps"
	"reflect"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestVehicleCatalogRevisionChangesOnlyConfirmedFieldsAndPreservesOtherOwnedVehicles(t *testing.T) {
	db := testDB(t)
	tx, err := db.BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := application.SeedDecoderLibrary(t.Context(), tx, "2026-11-06T20:00:00Z"); err != nil {
		_ = tx.Rollback()
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	service := application.NewVehicleService(db)
	input := application.CreateVehicleInput{Manufacturer: "Piko", Name: "Original", Gauge: "H0", Category: "Lokomotive", Gattung: "Diesellok", PurchasePrice: "42", Condition: "Used"}
	first, err := service.Create(t.Context(), input, "actor")
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.Create(t.Context(), input, "actor")
	if err != nil {
		t.Fatal(err)
	}
	before, err := service.CatalogContext(t.Context(), "vehicle", first.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE vehicle_article_links SET revision_id=?,overrides_json='{}' WHERE vehicle_id=?`, before.Pinned.ID, second.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE vehicles SET article_source_url=NULL WHERE id=?`, first.ID); err != nil {
		t.Fatal(err)
	}
	fields := maps.Clone(before.CurrentFields)
	fields["manufacturer"], fields["name"] = "Corrected manufacturer", "Catalog correction"
	command := application.VehicleCatalogRevisionInput{ID: strings.Repeat("a", 32), ExpectedRevision: before.Revision,
		Fields: fields, Note: "Explicit correction", SelectedFields: []string{"manufacturer"}}
	changed, err := service.CreateCatalogRevision(t.Context(), "vehicle", first.ID, command, "actor")
	if err != nil || changed.Pinned.ID != command.ID || changed.CurrentFields["manufacturer"] != fields["manufacturer"] || changed.CurrentFields["name"] != "Original" || changed.Overrides["name"] != "Original" || changed.Pinned.RevisionNumber != 2 {
		t.Fatalf("confirmed selection: %#v / %v", changed, err)
	}
	stored, err := service.Get(t.Context(), first.ID)
	if err != nil || stored.PurchasePrice != "42" || stored.Condition != "Used" || stored.ID != first.ID || stored.InventoryNumber != first.InventoryNumber {
		t.Fatal("catalog correction changed ownership or identity")
	}
	var rawNull bool
	if err := db.QueryRow(`SELECT article_source_url IS NULL FROM vehicles WHERE id=?`, first.ID).Scan(&rawNull); err != nil || !rawNull {
		t.Fatal("unselected NULL was rewritten")
	}
	other, err := service.CatalogContext(t.Context(), "vehicle", second.ID)
	if err != nil || other.Pinned.ID != before.Pinned.ID || !maps.Equal(other.CurrentFields, before.CurrentFields) {
		t.Fatal("publishing a revision silently changed another owned vehicle")
	}
	backup := application.NewBackupService(db, t.TempDir())
	snapshot, err := backup.Export(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.CreateCatalogRevision(t.Context(), "vehicle", first.ID, command, "actor"); !errors.Is(err, application.ErrVehicleCatalogConflict) {
		t.Fatalf("repeated confirmation: %v", err)
	}
	afterReplay, err := backup.Export(t.Context())
	if err != nil || !reflect.DeepEqual(catalogBackupTableSnapshot(t, snapshot), catalogBackupTableSnapshot(t, afterReplay)) {
		t.Fatal("repeated confirmation created another revision or changed data")
	}
	adopted, err := service.AdoptCatalogRevision(t.Context(), "vehicle", second.ID, application.VehicleCatalogAdoptionInput{
		ExpectedRevision: other.Revision, RevisionID: command.ID, SelectedFields: []string{"name"},
	}, "actor")
	if err != nil || adopted.CurrentFields["name"] != "Catalog correction" || adopted.CurrentFields["manufacturer"] != "Piko" || adopted.Overrides["manufacturer"] != "Piko" {
		t.Fatalf("independent adoption: %#v / %v", adopted, err)
	}
	independent, err := service.CreateCatalogRevision(t.Context(), "vehicle", first.ID, application.VehicleCatalogRevisionInput{
		ID: strings.Repeat("b", 32), ExpectedRevision: changed.Revision, Fields: changed.CurrentFields,
		Independent: true, SelectedFields: []string{},
	}, "actor")
	if err != nil || independent.Pinned.ArticleID == before.Pinned.ArticleID || independent.Pinned.RevisionNumber != 1 || !maps.Equal(independent.CurrentFields, changed.CurrentFields) {
		t.Fatalf("explicit independent article: %#v / %v", independent, err)
	}
	if _, err := service.AdoptCatalogRevision(t.Context(), "vehicle", second.ID, application.VehicleCatalogAdoptionInput{
		ExpectedRevision: adopted.Revision, RevisionID: independent.Pinned.ID, SelectedFields: []string{"name"},
	}, "actor"); !errors.Is(err, application.ErrVehicleCatalogNotFound) {
		t.Fatal("a foreign article revision was silently linked")
	}
	snapshot, err = backup.Export(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := backup.Import(t.Context(), snapshot); err != nil {
		t.Fatal(err)
	}
	restored, err := backup.Export(t.Context())
	if err != nil || !reflect.DeepEqual(catalogBackupTableSnapshot(t, snapshot), catalogBackupTableSnapshot(t, restored)) {
		if err == nil {
			originalTables, restoredTables := catalogBackupTableSnapshot(t, snapshot), catalogBackupTableSnapshot(t, restored)
			for table, rows := range originalTables {
				if !reflect.DeepEqual(rows, restoredTables[table]) {
					t.Logf("changed table: %s", table)
				}
			}
		}
		t.Fatal("controlled revisions or overrides did not survive restore")
	}
}

func TestVehicleCatalogAdoptionRetainsEveryUnresolvedProposalBeforeActivation(t *testing.T) {
	db := testDB(t)
	service := application.NewVehicleService(db)
	vehicle, err := service.Create(t.Context(), application.CreateVehicleInput{Manufacturer: "Piko", Name: "Owned", Gauge: "H0", Category: "Lokomotive", Gattung: "Diesellok"}, "actor")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE vehicle_article_links SET authority='legacy',source_authority='catalog',overrides_json='{"name":"Unselected pending name","description":"Unselected pending description"}' WHERE vehicle_id=?`, vehicle.ID); err != nil {
		t.Fatal(err)
	}
	before, err := service.CatalogContext(t.Context(), "vehicle", vehicle.ID)
	if err != nil {
		t.Fatal(err)
	}
	after, err := service.AdoptCatalogRevision(t.Context(), "vehicle", vehicle.ID, application.VehicleCatalogAdoptionInput{
		ExpectedRevision: before.Revision, RevisionID: before.Pinned.ID, SelectedFields: []string{},
	}, "actor")
	if err != nil || after.Authority != "catalog" || *after.SourceAuthority != "catalog" || !maps.Equal(after.CurrentFields, before.CurrentFields) || len(after.Revisions) != 2 {
		t.Fatalf("pending proposal activation: %#v / %v", after, err)
	}
	archived := after.Revisions[0]
	if archived.Fields["name"] != "Unselected pending name" || archived.Fields["description"] != "Unselected pending description" || !strings.Contains(archived.Note, "source time unknown") {
		t.Fatal("activation discarded unselected pending values or invented source time")
	}
	var auditBefore, auditAfter int
	if err := db.QueryRow(`SELECT COUNT(*) FROM audit_logs WHERE action='VehicleCatalogRevisionAdopted'`).Scan(&auditBefore); err != nil {
		t.Fatal(err)
	}
	noop, err := service.AdoptCatalogRevision(t.Context(), "vehicle", vehicle.ID, application.VehicleCatalogAdoptionInput{
		ExpectedRevision: after.Revision, RevisionID: after.Pinned.ID, SelectedFields: []string{},
	}, "actor")
	if err != nil || noop.Revision != after.Revision {
		t.Fatal("unchanged confirmation rewrote catalog data")
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM audit_logs WHERE action='VehicleCatalogRevisionAdopted'`).Scan(&auditAfter); err != nil || auditBefore != auditAfter {
		t.Fatal("unchanged confirmation created another audit event")
	}
}

func TestVehicleCatalogRevisionFailureAndStaleComparisonLeaveAllDataUnchanged(t *testing.T) {
	for _, failure := range []string{"projection", "audit", "stale"} {
		t.Run(failure, func(t *testing.T) {
			db := testDB(t)
			service := application.NewVehicleService(db)
			vehicle, err := service.Create(t.Context(), application.CreateVehicleInput{Manufacturer: "Piko", Name: "Owned", Gauge: "H0", Category: "Lokomotive", Gattung: "Diesellok"}, "actor")
			if err != nil {
				t.Fatal(err)
			}
			before, err := service.CatalogContext(t.Context(), "vehicle", vehicle.ID)
			if err != nil {
				t.Fatal(err)
			}
			switch failure {
			case "projection":
				_, err = db.Exec(`CREATE TRIGGER reject_catalog_projection BEFORE UPDATE ON vehicles BEGIN SELECT RAISE(ABORT,'injected projection failure'); END`)
			case "audit":
				_, err = db.Exec(`CREATE TRIGGER reject_catalog_audit BEFORE INSERT ON audit_logs WHEN NEW.action='VehicleCatalogRevisionCreated' BEGIN SELECT RAISE(ABORT,'injected audit failure'); END`)
			case "stale":
				_, err = db.Exec(`UPDATE vehicles SET name='Newer owned value' WHERE id=?`, vehicle.ID)
			}
			if err != nil {
				t.Fatal(err)
			}
			backup := application.NewBackupService(db, t.TempDir())
			snapshot, err := backup.Export(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			fields := maps.Clone(before.CurrentFields)
			fields["name"] = "Must not persist"
			_, err = service.CreateCatalogRevision(t.Context(), "vehicle", vehicle.ID, application.VehicleCatalogRevisionInput{
				ID: strings.Repeat("c", 32), ExpectedRevision: before.Revision, Fields: fields, Independent: true, SelectedFields: []string{"name"},
			}, "actor")
			if err == nil || (failure == "stale" && !errors.Is(err, application.ErrVehicleCatalogConflict)) {
				t.Fatalf("failure was accepted: %v", err)
			}
			after, err := backup.Export(t.Context())
			if err != nil || !reflect.DeepEqual(catalogBackupTableSnapshot(t, snapshot), catalogBackupTableSnapshot(t, after)) {
				t.Fatal("rejected confirmation retained a partial article, revision, link or vehicle update")
			}
		})
	}
}

func TestVehicleCatalogFieldsRejectNullUnknownAndNonStringValues(t *testing.T) {
	for _, raw := range []string{`null`, `{"name":null}`, `{"description":null}`, `{"purchasePrice":"1"}`, `{"name":2}`} {
		var fields application.VehicleCatalogFields
		if err := json.Unmarshal([]byte(raw), &fields); err == nil {
			t.Fatalf("accepted invalid catalog JSON: %s", raw)
		}
	}
}

func TestVehicleCatalogRevisionRejectsIncompleteOrAmbiguousConfirmation(t *testing.T) {
	db := testDB(t)
	service := application.NewVehicleService(db)
	vehicle, err := service.Create(t.Context(), application.CreateVehicleInput{Manufacturer: "Piko", Name: "Owned", Gauge: "H0", Category: "Lokomotive", Gattung: "Diesellok"}, "actor")
	if err != nil {
		t.Fatal(err)
	}
	state, err := service.CatalogContext(t.Context(), "vehicle", vehicle.ID)
	if err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func(*application.VehicleCatalogRevisionInput){
		"missing selection":   func(input *application.VehicleCatalogRevisionInput) { input.SelectedFields = nil },
		"duplicate selection": func(input *application.VehicleCatalogRevisionInput) { input.SelectedFields = []string{"name", "name"} },
		"ownership selection": func(input *application.VehicleCatalogRevisionInput) { input.SelectedFields = []string{"purchasePrice"} },
		"missing model field": func(input *application.VehicleCatalogRevisionInput) { delete(input.Fields, "description") },
		"unknown model field": func(input *application.VehicleCatalogRevisionInput) {
			delete(input.Fields, "description")
			input.Fields["purchasePrice"] = "1"
		},
		"blank required field": func(input *application.VehicleCatalogRevisionInput) { input.Fields["manufacturer"] = " " },
		"missing comparison":   func(input *application.VehicleCatalogRevisionInput) { input.ExpectedRevision = "" },
		"invalid identity":     func(input *application.VehicleCatalogRevisionInput) { input.ID = "not-a-stable-identity" },
		"overlong note":        func(input *application.VehicleCatalogRevisionInput) { input.Note = strings.Repeat("ö", 2001) },
	} {
		t.Run(name, func(t *testing.T) {
			input := application.VehicleCatalogRevisionInput{ID: strings.Repeat("d", 32), ExpectedRevision: state.Revision, Fields: maps.Clone(state.CurrentFields), SelectedFields: []string{}}
			mutate(&input)
			if _, err := service.CreateCatalogRevision(t.Context(), "vehicle", vehicle.ID, input, "actor"); !errors.Is(err, application.ErrVehicleCatalogValidation) {
				t.Fatalf("invalid confirmation accepted: %v", err)
			}
			after, err := service.CatalogContext(t.Context(), "vehicle", vehicle.ID)
			if err != nil || after.Revision != state.Revision {
				t.Fatal("invalid confirmation changed catalog data")
			}
		})
	}
}

func TestVehicleCatalogSetRevisionKeepsMembersAndAcquisitionUnchanged(t *testing.T) {
	db := testDB(t)
	service := application.NewVehicleService(db)
	set, err := service.CreateSet(t.Context(), application.CreateVehicleSetInput{
		Set:     application.VehicleSetInput{Manufacturer: "Roco", Name: "Set", Gauge: "H0", Category: "Wagen", Gattung: "Reisezugwagen", PurchasePrice: "180"},
		Members: []application.CreateVehicleInput{{Name: "A"}, {Name: "B"}},
	}, "actor")
	if err != nil {
		t.Fatal(err)
	}
	state, err := service.CatalogContext(t.Context(), "set", set.ID)
	if err != nil {
		t.Fatal(err)
	}
	fields := maps.Clone(state.CurrentFields)
	fields["name"] = "Set correction"
	if _, err := service.CreateCatalogRevision(t.Context(), "set", set.ID, application.VehicleCatalogRevisionInput{
		ID: strings.Repeat("e", 32), ExpectedRevision: state.Revision, Fields: fields, SelectedFields: []string{"name"},
	}, "actor"); err != nil {
		t.Fatal(err)
	}
	after, err := service.GetSet(t.Context(), set.ID)
	if err != nil || after.Name != "Set correction" || after.PurchasePrice != "180" || after.InventoryNumber != set.InventoryNumber {
		t.Fatal("set correction changed acquisition or identity")
	}
	for _, member := range after.Members {
		context, err := service.CatalogContext(t.Context(), "vehicle", member.ID)
		if err != nil || context.Pinned.ID != state.Pinned.ID || context.CurrentFields["name"] != member.Name {
			t.Fatal("set revision changed a member without confirmation")
		}
	}
}
