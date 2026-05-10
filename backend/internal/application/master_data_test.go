package application_test

import (
	"context"
	"testing"

	"railkeeper2/backend/internal/application"
)

func TestMasterDataExportImportReplacesEntriesAndRelations(t *testing.T) {
	ctx := context.Background()
	sourceDB := testDB(t)
	source := application.NewMasterDataService(sourceDB)

	active := true
	inactive := false
	if _, err := source.Create(ctx, "vehicle_category", application.MasterDataInput{
		Key:    "lokomotive",
		Label:  "Lokomotive",
		Active: &active,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := source.Create(ctx, "vehicle_gattung", application.MasterDataInput{
		Key:    "diesellok",
		Label:  "Diesellok",
		Active: &inactive,
		Metadata: map[string]any{
			"note": "import-test",
		},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := sourceDB.ExecContext(ctx, `
INSERT INTO master_data_relations(id, parent_type, parent_key, child_type, child_key, sort_order, created_at)
VALUES('rel-1', 'vehicle_category', 'lokomotive', 'vehicle_gattung', 'diesellok', 7, '2026-05-10T00:00:00Z')
`); err != nil {
		t.Fatal(err)
	}

	doc, err := source.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Format == "" || len(doc.Entries["vehicle_category"]) != 1 || len(doc.Relations) != 1 {
		t.Fatalf("unexpected master data export: %#v", doc)
	}
	expectedEntries := 0
	for _, entries := range doc.Entries {
		expectedEntries += len(entries)
	}

	targetDB := testDB(t)
	target := application.NewMasterDataService(targetDB)
	if _, err := target.Create(ctx, "epoch", application.MasterDataInput{
		Key:    "old",
		Label:  "Alt",
		Active: &active,
	}); err != nil {
		t.Fatal(err)
	}
	result, err := target.Import(ctx, doc)
	if err != nil {
		t.Fatal(err)
	}
	if result.ImportedEntries != expectedEntries || result.ImportedRelations != len(doc.Relations) {
		t.Fatalf("unexpected import result: %#v", result)
	}

	all, err := target.ListAll(ctx, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range all["epoch"] {
		if entry.Key == "old" {
			t.Fatalf("expected old target entry to be replaced, got %#v", all["epoch"])
		}
	}
	if len(all["vehicle_category"]) != 1 || all["vehicle_category"][0].Label != "Lokomotive" {
		t.Fatalf("expected imported category, got %#v", all["vehicle_category"])
	}
	if len(all["vehicle_gattung"]) != 1 || all["vehicle_gattung"][0].Metadata["note"] != "import-test" {
		t.Fatalf("expected imported gattung metadata, got %#v", all["vehicle_gattung"])
	}
	relations, err := target.Relations(ctx, "vehicle_category", "vehicle_gattung")
	if err != nil {
		t.Fatal(err)
	}
	if len(relations) != 1 || relations[0].SortOrder != 7 {
		t.Fatalf("expected imported relation, got %#v", relations)
	}
}

func TestMasterDataImportRejectsInvalidDocument(t *testing.T) {
	service := application.NewMasterDataService(testDB(t))
	if _, err := service.Import(context.Background(), &application.MasterDataDocument{
		Format: "wrong",
	}); err != application.ErrMasterDataValidation {
		t.Fatalf("expected validation error, got %v", err)
	}
}
