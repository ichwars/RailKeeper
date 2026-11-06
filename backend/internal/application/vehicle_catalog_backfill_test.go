package application_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestVehicleCatalogBackfillKeepsIdenticalOwnedVehiclesSeparateAndSetOverridesExact(t *testing.T) {
	db := testDB(t)
	service := application.NewVehicleService(db)
	input := application.CreateVehicleInput{Manufacturer: "Piko", ArticleNumber: "123", Name: "Same model", Gauge: "H0", Category: "Lokomotive", Gattung: "Diesellok"}
	first, err := service.Create(t.Context(), input, "actor")
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.Create(t.Context(), input, "actor")
	if err != nil {
		t.Fatal(err)
	}
	set, err := service.CreateSet(t.Context(), application.CreateVehicleSetInput{Set: application.VehicleSetInput{
		Manufacturer: "Roco", ArticleNumber: "123", Name: "Set", Gauge: "H0", Category: "Wagen", Gattung: "Reisezugwagen", Description: "Shared catalog text", PurchasePrice: "120.00",
	}, Members: []application.CreateVehicleInput{{Name: "Member A"}, {Name: "Member B", Gauge: "TT"}}}, "actor")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE vehicles SET description='', article_source_url=NULL WHERE id=?`, set.Members[1].ID); err != nil {
		t.Fatal(err)
	}
	before, err := service.List(t.Context(), "")
	if err != nil {
		t.Fatal(err)
	}
	report, err := service.BackfillLegacyCatalog(t.Context(), "actor")
	if err != nil {
		t.Fatal(err)
	}
	if report.Vehicles != 4 || report.LinkedVehicles != 4 || report.Sets != 1 || report.LinkedSets != 1 || report.DifferentRepresentations != 0 {
		t.Fatalf("report=%#v", report)
	}
	articleIDs := map[string]string{}
	for _, assignment := range report.Assignments {
		articleIDs[assignment.TargetID] = assignment.ArticleID
		if assignment.Authority != "legacy" {
			t.Fatal("backfill changed authority")
		}
	}
	if articleIDs[first.ID] == articleIDs[second.ID] || articleIDs[first.ID] == articleIDs[set.ID] {
		t.Fatal("backfill guessed article equivalence")
	}
	if articleIDs[set.Members[0].ID] != articleIDs[set.ID] || articleIDs[set.Members[1].ID] != articleIDs[set.ID] {
		t.Fatal("known set catalog relationship missing")
	}
	var raw string
	if err := db.QueryRow(`SELECT overrides_json FROM vehicle_article_links WHERE vehicle_id=?`, set.Members[1].ID).Scan(&raw); err != nil {
		t.Fatal(err)
	}
	var overrides application.VehicleCatalogFields
	if err := json.Unmarshal([]byte(raw), &overrides); err != nil {
		t.Fatal(err)
	}
	if overrides["name"] != "Member B" || overrides["gauge"] != "TT" {
		t.Fatalf("member overrides=%#v", overrides)
	}
	if value, present := overrides["description"]; !present || value != "" {
		t.Fatal("explicit empty member difference disappeared")
	}
	var isNull bool
	if err := db.QueryRow(`SELECT article_source_url IS NULL FROM vehicles WHERE id=?`, set.Members[1].ID).Scan(&isNull); err != nil || !isNull {
		t.Fatal("legacy NULL was normalized in place")
	}
	after, err := service.List(t.Context(), "")
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatal("backfill changed legacy vehicle responses")
	}
	backup := application.NewBackupService(db, t.TempDir())
	snapshot, err := backup.Export(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.BackfillLegacyCatalog(t.Context(), "actor"); err != nil {
		t.Fatal(err)
	}
	repeated, err := backup.Export(t.Context())
	if err != nil || !reflect.DeepEqual(catalogBackupTableSnapshot(t, snapshot), catalogBackupTableSnapshot(t, repeated)) {
		t.Fatal("repeating backfill changed stored data")
	}
	var audits int
	if err := db.QueryRow(`SELECT COUNT(*) FROM audit_logs WHERE action='VehicleCatalogLegacyBackfilled'`).Scan(&audits); err != nil || audits != 1 {
		t.Fatalf("duplicate backfill audit: %d / %v", audits, err)
	}
	if _, err := db.Exec(`UPDATE vehicles SET name='Later legacy correction' WHERE id=?`, first.ID); err != nil {
		t.Fatal(err)
	}
	different, err := service.CatalogBackfillReport(t.Context())
	if err != nil || different.DifferentRepresentations != 1 {
		t.Fatalf("legacy drift not reported: %#v / %v", different, err)
	}
}

func TestVehicleCatalogBackfillRollsBackAllNewRelationshipsOnFailure(t *testing.T) {
	db := testDB(t)
	service := application.NewVehicleService(db)
	input := application.CreateVehicleInput{Manufacturer: "Piko", Name: "Preserved", Gauge: "H0", Category: "Lokomotive", Gattung: "Diesellok"}
	vehicle, err := service.Create(t.Context(), input, "actor")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TRIGGER reject_catalog_link BEFORE INSERT ON vehicle_article_links BEGIN SELECT RAISE(ABORT, 'injected storage failure'); END`); err != nil {
		t.Fatal(err)
	}
	if _, err := service.BackfillLegacyCatalog(t.Context(), "actor"); err == nil {
		t.Fatal("injected failure was accepted")
	}
	for _, table := range []string{"catalog_articles", "catalog_article_revisions", "vehicle_article_links", "vehicle_set_article_links"} {
		var count int
		if err := db.QueryRow(`SELECT COUNT(*) FROM ` + table).Scan(&count); err != nil || count != 0 {
			t.Fatalf("partial backfill in %s: %d / %v", table, count, err)
		}
	}
	stored, err := service.Get(t.Context(), vehicle.ID)
	if err != nil || stored.Name != vehicle.Name || stored.InventoryNumber != vehicle.InventoryNumber {
		t.Fatal("failed backfill changed inventory")
	}
}
