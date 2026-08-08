package application_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestMasterDataImportPreservesArticleTypesForLegacyDocuments(t *testing.T) {
	ctx := context.Background()
	service := application.NewMasterDataService(testDB(t))

	result, err := service.Import(ctx, &application.MasterDataDocument{
		Format: "railkeeper-master-data", Version: 1,
		Entries: map[string][]application.MasterDataEntry{
			"vehicle_category": {{Key: "lok", Label: "Lok", Active: true}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.ImportedTypes != 1 || result.ImportedEntries != 1 {
		t.Fatalf("preserved article types were reported as imported: %#v", result)
	}
	articleTypes, err := service.List(ctx, "article_type", false)
	if err != nil {
		t.Fatal(err)
	}
	if got := masterDataKeys(articleTypes); !reflect.DeepEqual(got, authoritativeArticleTypeKeys()) {
		t.Fatalf("legacy import changed authoritative article types: %#v", got)
	}
}

func TestMasterDataImportRejectsMalformedArticleTypesBeforeMutation(t *testing.T) {
	for _, test := range []struct {
		name   string
		mutate func([]application.MasterDataEntry) []application.MasterDataEntry
	}{
		{name: "partial", mutate: func(entries []application.MasterDataEntry) []application.MasterDataEntry {
			return entries[:len(entries)-1]
		}},
		{name: "extra", mutate: func(entries []application.MasterDataEntry) []application.MasterDataEntry {
			return append(entries, application.MasterDataEntry{Key: "custom", Label: "Custom", Active: true})
		}},
		{name: "duplicate", mutate: func(entries []application.MasterDataEntry) []application.MasterDataEntry {
			return append(entries, entries[0])
		}},
		{name: "renamed", mutate: func(entries []application.MasterDataEntry) []application.MasterDataEntry {
			entries[0].Key = "renamed"
			return entries
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			db := testDB(t)
			service := application.NewMasterDataService(db)
			active := true
			if _, err := service.Create(ctx, "epoch", application.MasterDataInput{
				Key: "sentinel", Label: "Sentinel", Active: &active,
			}); err != nil {
				t.Fatal(err)
			}
			if _, err := db.ExecContext(ctx, `
CREATE TRIGGER reject_master_data_mutation
BEFORE DELETE ON master_data_entries
BEGIN
  SELECT RAISE(FAIL, 'mutation reached');
END`); err != nil {
				t.Fatal(err)
			}

			entries := test.mutate(currentArticleTypes(t, service))
			_, err := service.Import(ctx, &application.MasterDataDocument{
				Format: "railkeeper-master-data", Version: 1,
				Entries: map[string][]application.MasterDataEntry{"article_type": entries},
			})
			if !errors.Is(err, application.ErrMasterDataValidation) {
				t.Fatalf("expected protected article-type validation, got %v", err)
			}
			if _, err := service.Get(ctx, "epoch", "sentinel"); err != nil {
				t.Fatalf("preflight failure mutated sentinel data: %v", err)
			}
		})
	}
}

func TestMasterDataImportRejectsEffectiveArticleTypesAcrossBucketsBeforeMutation(t *testing.T) {
	t.Run("foreign bucket extra", func(t *testing.T) {
		assertMasterDataImportRejectedBeforeMutation(t, func(
			_ []application.MasterDataEntry,
		) map[string][]application.MasterDataEntry {
			return map[string][]application.MasterDataEntry{
				"vehicle_category": {{
					Type: "article_type", Key: "custom", Label: "Custom", Active: true,
				}},
			}
		})
	})

	t.Run("duplicate across buckets", func(t *testing.T) {
		assertMasterDataImportRejectedBeforeMutation(t, func(
			articleTypes []application.MasterDataEntry,
		) map[string][]application.MasterDataEntry {
			return map[string][]application.MasterDataEntry{
				"article_type":     articleTypes,
				"vehicle_category": {articleTypes[0]},
			}
		})
	})
}

func TestMasterDataImportAcceptsArticleTypesSplitAcrossBuckets(t *testing.T) {
	ctx := context.Background()
	service := application.NewMasterDataService(testDB(t))
	articleTypes := currentArticleTypes(t, service)

	_, err := service.Import(ctx, &application.MasterDataDocument{
		Format: "railkeeper-master-data", Version: 1,
		Entries: map[string][]application.MasterDataEntry{
			"vehicle_category": articleTypes[:4],
			"epoch":            articleTypes[4:],
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	got, err := service.List(ctx, "article_type", false)
	if err != nil {
		t.Fatal(err)
	}
	if keys := masterDataKeys(got); !reflect.DeepEqual(keys, authoritativeArticleTypeKeys()) {
		t.Fatalf("split import changed authoritative article types: %#v", keys)
	}
}

func TestMasterDataImportKeepsNonArticleEntryFromArticleTypeBucket(t *testing.T) {
	ctx := context.Background()
	service := application.NewMasterDataService(testDB(t))
	articleTypes := currentArticleTypes(t, service)
	articleTypes = append(articleTypes, application.MasterDataEntry{
		Type: "vehicle_category", Key: "foreign-category", Label: "Foreign category", Active: true,
	})

	_, err := service.Import(ctx, &application.MasterDataDocument{
		Format: "railkeeper-master-data", Version: 1,
		Entries: map[string][]application.MasterDataEntry{"article_type": articleTypes},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Get(ctx, "vehicle_category", "foreign-category"); err != nil {
		t.Fatalf("non-article entry from article-type bucket was not imported: %v", err)
	}
	got, err := service.List(ctx, "article_type", false)
	if err != nil {
		t.Fatal(err)
	}
	if keys := masterDataKeys(got); !reflect.DeepEqual(keys, authoritativeArticleTypeKeys()) {
		t.Fatalf("non-article entry changed authoritative article types: %#v", keys)
	}
}

func assertMasterDataImportRejectedBeforeMutation(
	t *testing.T,
	entries func([]application.MasterDataEntry) map[string][]application.MasterDataEntry,
) {
	t.Helper()
	ctx := context.Background()
	db := testDB(t)
	service := application.NewMasterDataService(db)
	active := true
	if _, err := service.Create(ctx, "epoch", application.MasterDataInput{
		Key: "sentinel", Label: "Sentinel", Active: &active,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `
CREATE TRIGGER reject_master_data_mutation
BEFORE DELETE ON master_data_entries
BEGIN
  SELECT RAISE(FAIL, 'mutation reached');
END`); err != nil {
		t.Fatal(err)
	}

	_, err := service.Import(ctx, &application.MasterDataDocument{
		Format: "railkeeper-master-data", Version: 1,
		Entries: entries(currentArticleTypes(t, service)),
	})
	if !errors.Is(err, application.ErrMasterDataValidation) {
		t.Fatalf("expected protected article-type validation before mutation, got %v", err)
	}
	if _, err := service.Get(ctx, "epoch", "sentinel"); err != nil {
		t.Fatalf("preflight failure mutated sentinel data: %v", err)
	}
}

func TestMasterDataImportChangesOnlyMutableArticleTypeFields(t *testing.T) {
	ctx := context.Background()
	service := application.NewMasterDataService(testDB(t))
	entries := currentArticleTypes(t, service)
	before, err := service.Get(ctx, "article_type", "track")
	if err != nil {
		t.Fatal(err)
	}
	for index := range entries {
		if entries[index].Key != "track" {
			continue
		}
		entries[index].ID = "malicious-id"
		entries[index].Label = "Gleismaterial"
		entries[index].Active = false
		entries[index].SortOrder = 999
		entries[index].SourceURL = "https://malicious.invalid"
		entries[index].Metadata = map[string]any{"note": "mutable"}
		entries[index].CreatedAt = "2000-01-01T00:00:00Z"
	}
	if _, err := service.Import(ctx, &application.MasterDataDocument{
		Format: "railkeeper-master-data", Version: 1,
		Entries: map[string][]application.MasterDataEntry{"article_type": entries},
	}); err != nil {
		t.Fatal(err)
	}
	after, err := service.Get(ctx, "article_type", "track")
	if err != nil {
		t.Fatal(err)
	}
	if after.ID != before.ID || after.SourceURL != before.SourceURL || after.CreatedAt != before.CreatedAt {
		t.Fatalf("import changed stable article-type fields: before=%#v after=%#v", before, after)
	}
	if after.Label != "Gleismaterial" || after.Active || after.SortOrder != 999 ||
		after.Metadata["note"] != "mutable" {
		t.Fatalf("import did not update mutable article-type fields: %#v", after)
	}
}

func currentArticleTypes(t *testing.T, service *application.MasterDataService) []application.MasterDataEntry {
	t.Helper()
	entries, err := service.List(context.Background(), "article_type", false)
	if err != nil {
		t.Fatal(err)
	}
	return entries
}

func masterDataKeys(entries []application.MasterDataEntry) []string {
	keys := make([]string, 0, len(entries))
	for _, entry := range entries {
		keys = append(keys, entry.Key)
	}
	return keys
}

func authoritativeArticleTypeKeys() []string {
	return []string{"track", "signal", "decoder", "electrical_control", "building_equipment",
		"landscape_consumable", "lighting", "other"}
}

func TestMasterDataServiceProtectsStandardArticleTypeKeys(t *testing.T) {
	ctx := context.Background()
	service := application.NewMasterDataService(testDB(t))
	active := true
	inactive := false

	if _, err := service.Create(ctx, "article_type", application.MasterDataInput{
		Key: "custom", Label: "Custom", Active: &active,
	}); !errors.Is(err, application.ErrMasterDataValidation) {
		t.Fatalf("expected article type creation to be rejected, got %v", err)
	}
	if _, err := service.Update(ctx, "article_type", "track", application.MasterDataInput{
		Key: "renamed", Label: "Gleismaterial", Active: &active,
	}); !errors.Is(err, application.ErrMasterDataValidation) {
		t.Fatalf("expected article type key change to be rejected, got %v", err)
	}
	updated, err := service.Update(ctx, "article_type", "track", application.MasterDataInput{
		Label: "Gleismaterial", Active: &inactive,
	})
	if err != nil {
		t.Fatalf("expected article type label and active state to remain editable: %v", err)
	}
	if updated.Key != "track" || updated.Label != "Gleismaterial" || updated.Active {
		t.Fatalf("unexpected standard article type update: %#v", updated)
	}
	if err := service.Delete(ctx, "article_type", "track"); !errors.Is(err, application.ErrMasterDataValidation) {
		t.Fatalf("expected article type deletion to be rejected, got %v", err)
	}
}

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

func TestESUFunctionSymbolsSeededWithImages(t *testing.T) {
	service := application.NewMasterDataService(testDB(t))
	entries, err := service.List(context.Background(), "symbols", true)
	if err != nil {
		t.Fatal(err)
	}

	esuCount := 0
	var fahrgeraeusch *application.MasterDataEntry
	for i := range entries {
		if entries[i].Metadata["category"] == "ESU ECoS" {
			esuCount++
		}
		if entries[i].Key == "esu-f006-fahrgeraeusch" {
			fahrgeraeusch = &entries[i]
		}
	}

	if esuCount != 86 {
		t.Fatalf("expected 86 ESU function symbols, got %d", esuCount)
	}
	if fahrgeraeusch == nil {
		t.Fatal("expected Fahrgeraeusch symbol")
	}
	if fahrgeraeusch.Label != "Fahrgeräusch" {
		t.Fatalf("unexpected symbol label %q", fahrgeraeusch.Label)
	}
	if fahrgeraeusch.Metadata["description"] == "" {
		t.Fatalf("expected symbol description metadata: %#v", fahrgeraeusch.Metadata)
	}
	for _, key := range []string{"imageData", "activeImageData", "inactiveImageData"} {
		value, ok := fahrgeraeusch.Metadata[key].(string)
		if !ok || len(value) < 100 || value[:26] != "data:image/svg+xml;base64," {
			t.Fatalf("expected %s SVG data URL, got %#v", key, fahrgeraeusch.Metadata[key])
		}
	}
}
