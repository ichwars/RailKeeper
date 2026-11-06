package application_test

import (
	"encoding/json"
	"errors"
	"reflect"
	"sort"
	"testing"

	"railkeeper/backend/internal/application"
)

func catalogBackupTableSnapshot(t *testing.T, doc *application.BackupDocument) map[string][]string {
	t.Helper()
	tables := map[string][]string{}
	for table, rows := range doc.Tables {
		tables[table] = []string{}
		for _, row := range rows {
			raw, err := json.Marshal(row)
			if err != nil {
				t.Fatal(err)
			}
			tables[table] = append(tables[table], string(raw))
		}
		sort.Strings(tables[table])
	}
	return tables
}

func catalogBackupFixture(t *testing.T) (*application.BackupService, *application.BackupDocument) {
	t.Helper()
	dir := t.TempDir()
	db := backupTestDB(t, dir)
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
	vehicle, err := service.Create(t.Context(), application.CreateVehicleInput{Manufacturer: "Piko", Name: "Owned label", Gauge: "H0", Category: "Lokomotive", Gattung: "Diesellok"}, "actor")
	if err != nil {
		t.Fatal(err)
	}
	set, err := service.CreateSet(t.Context(), application.CreateVehicleSetInput{
		Set:     application.VehicleSetInput{Manufacturer: "Roco", Name: "Set", Gauge: "H0", Category: "Wagen", Gattung: "Reisezugwagen"},
		Members: []application.CreateVehicleInput{{Name: "A"}, {Name: "B"}},
	}, "actor")
	if err != nil {
		t.Fatal(err)
	}
	fields := application.VehicleCatalogFields{"name": "Catalog label", "manufacturer": "Piko", "articleNumber": "123", "articleSourceUrl": "", "gauge": "H0", "epoch": "", "railwayCompany": "", "category": "Lokomotive", "gattung": "Diesellok", "description": "", "ean": "", "productionPeriod": "", "listPrice": ""}
	raw, err := json.Marshal(fields)
	if err != nil {
		t.Fatal(err)
	}
	now := "2026-11-06T20:00:00Z"
	for _, query := range []struct {
		sql  string
		args []any
	}{
		{`INSERT INTO catalog_articles(id, origin, source_entity_id, created_at) VALUES('article', 'manual', NULL, ?)`, []any{now}},
		{`INSERT INTO catalog_article_revisions(id, article_id, revision_number, fields_json, note, created_at) VALUES('revision-1', 'article', 1, ?, 'Original', ?), ('revision-2', 'article', 2, ?, 'Explicit correction', ?)`, []any{string(raw), now, string(raw), now}},
		{`INSERT INTO vehicle_article_links(vehicle_id, revision_id, overrides_json, authority, updated_at) VALUES(?, 'revision-1', '{"name":"Owned label","manufacturer":""}', 'legacy', ?)`, []any{vehicle.ID, now}},
		{`INSERT INTO vehicle_set_article_links(vehicle_set_id, revision_id, overrides_json, authority, updated_at) VALUES(?, 'revision-2', '{"gauge":"TT"}', 'legacy', ?)`, []any{set.ID, now}},
	} {
		if _, err := db.ExecContext(t.Context(), query.sql, query.args...); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec(`UPDATE catalog_article_revisions SET note='Mutated' WHERE id='revision-1'`); err == nil {
		t.Fatal("revision was mutable")
	}
	backup := application.NewBackupService(db, dir)
	doc, err := backup.Export(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	return backup, doc
}

func TestVehicleCatalogBackupRoundTripPreservesRevisionsLinksAndIndividualOverrides(t *testing.T) {
	service, doc := catalogBackupFixture(t)
	if doc.Version != 25 {
		t.Fatalf("version=%d", doc.Version)
	}
	validation, err := service.Validate(t.Context(), doc)
	if err != nil || !validation.Compatible {
		t.Fatalf("validation: %#v / %v", validation, err)
	}
	if _, err := service.Import(t.Context(), doc); err != nil {
		t.Fatal(err)
	}
	after, err := service.Export(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(catalogBackupTableSnapshot(t, after), catalogBackupTableSnapshot(t, doc)) || !reflect.DeepEqual(after.Files, doc.Files) {
		beforeTables, afterTables := catalogBackupTableSnapshot(t, doc), catalogBackupTableSnapshot(t, after)
		for table, rows := range beforeTables {
			if !reflect.DeepEqual(rows, afterTables[table]) {
				t.Logf("changed table: %s", table)
			}
		}
		t.Fatal("catalog or existing inventory changed across restore")
	}
	if len(after.Tables["catalog_article_revisions"]) != 2 {
		t.Fatal("historical revision missing")
	}
}

func TestVehicleCatalogBackupPreflightRejectsIncompleteRelationshipsBeforeAnyRestore(t *testing.T) {
	service, original := catalogBackupFixture(t)
	for name, mutate := range map[string]func(*application.BackupDocument){
		"missing table":  func(d *application.BackupDocument) { delete(d.Tables, "vehicle_article_links") },
		"unknown column": func(d *application.BackupDocument) { d.Tables["catalog_articles"][0]["future_value"] = "must not drop" },
		"orphan revision": func(d *application.BackupDocument) {
			d.Tables["catalog_article_revisions"][0]["article_id"] = "missing"
		},
		"duplicate revision":  func(d *application.BackupDocument) { d.Tables["catalog_article_revisions"][1]["revision_number"] = 1 },
		"fractional revision": func(d *application.BackupDocument) { d.Tables["catalog_article_revisions"][0]["revision_number"] = 1.5 },
		"orphan vehicle":      func(d *application.BackupDocument) { d.Tables["vehicle_article_links"][0]["vehicle_id"] = "missing" },
		"orphan set": func(d *application.BackupDocument) {
			d.Tables["vehicle_set_article_links"][0]["vehicle_set_id"] = "missing"
		},
		"orphan link":             func(d *application.BackupDocument) { d.Tables["vehicle_article_links"][0]["revision_id"] = "missing" },
		"missing snapshot fields": func(d *application.BackupDocument) { d.Tables["catalog_article_revisions"][0]["fields_json"] = "{}" },
		"null snapshot":           func(d *application.BackupDocument) { d.Tables["catalog_article_revisions"][0]["fields_json"] = "null" },
		"unknown ownership override": func(d *application.BackupDocument) {
			d.Tables["vehicle_article_links"][0]["overrides_json"] = `{"purchasePrice":"20"}`
		},
		"null override": func(d *application.BackupDocument) {
			d.Tables["vehicle_article_links"][0]["overrides_json"] = `{"name":null}`
		},
		"nonstring override": func(d *application.BackupDocument) {
			d.Tables["vehicle_article_links"][0]["overrides_json"] = `{"name":12}`
		},
		"unknown authority":     func(d *application.BackupDocument) { d.Tables["vehicle_article_links"][0]["authority"] = "both" },
		"unknown origin":        func(d *application.BackupDocument) { d.Tables["catalog_articles"][0]["origin"] = "guessed" },
		"missing legacy source": func(d *application.BackupDocument) { d.Tables["catalog_articles"][0]["origin"] = "legacy_vehicle" },
		"invalid timestamp": func(d *application.BackupDocument) {
			d.Tables["catalog_article_revisions"][0]["created_at"] = "unknown"
		},
		"downgraded format": func(d *application.BackupDocument) { d.Version = 24 },
	} {
		t.Run(name, func(t *testing.T) {
			raw, err := json.Marshal(original)
			if err != nil {
				t.Fatal(err)
			}
			var doc application.BackupDocument
			if err := json.Unmarshal(raw, &doc); err != nil {
				t.Fatal(err)
			}
			mutate(&doc)
			validation, err := service.Validate(t.Context(), &doc)
			if err != nil || validation.Compatible {
				t.Fatalf("accepted invalid catalog: %#v / %v", validation, err)
			}
			if _, err := service.Import(t.Context(), &doc); !errors.Is(err, application.ErrBackupInvalid) {
				t.Fatalf("restore error=%v", err)
			}
			after, err := service.Export(t.Context())
			if err != nil || !reflect.DeepEqual(after.Tables, original.Tables) || !reflect.DeepEqual(after.Files, original.Files) {
				t.Fatal("rejected catalog restore changed existing data")
			}
		})
	}
}

func TestVehicleCatalogBackupAcceptsVersion24WithoutCatalogAndRetainsLegacyInventory(t *testing.T) {
	service, doc := catalogBackupFixture(t)
	doc.Version = 24
	for _, table := range []string{"catalog_articles", "catalog_article_revisions", "vehicle_article_links", "vehicle_set_article_links"} {
		delete(doc.Tables, table)
	}
	validation, err := service.Validate(t.Context(), doc)
	if err != nil || !validation.Compatible {
		t.Fatalf("legacy validation: %#v / %v", validation, err)
	}
	if _, err := service.Import(t.Context(), doc); err != nil {
		t.Fatal(err)
	}
	after, err := service.Export(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	for _, table := range []string{"vehicles", "vehicle_sets", "vehicle_set_members"} {
		if !reflect.DeepEqual(after.Tables[table], doc.Tables[table]) {
			t.Fatalf("legacy %s changed", table)
		}
	}
	for _, table := range []string{"catalog_articles", "catalog_article_revisions", "vehicle_article_links", "vehicle_set_article_links"} {
		if len(after.Tables[table]) != 0 {
			t.Fatalf("legacy restore invented %s", table)
		}
	}
}
