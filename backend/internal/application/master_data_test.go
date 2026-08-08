package application_test

import (
	"context"
	"errors"
	"reflect"
	"slices"
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

func TestMasterDataExportUsesAccessorySubtypeAndPreservesItOnImport(t *testing.T) {
	ctx := context.Background()
	source := application.NewMasterDataService(testDB(t))
	document, err := source.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(document.Entries["accessory_subtype"]) != 54 {
		t.Fatalf("expected 54 accessory subtypes, got %d", len(document.Entries["accessory_subtype"]))
	}
	if len(document.Entries["article_subtype"]) != 0 {
		t.Fatalf("obsolete article_subtype escaped export: %#v", document.Entries["article_subtype"])
	}

	target := application.NewMasterDataService(testDB(t))
	if _, err := target.Import(ctx, document); err != nil {
		t.Fatal(err)
	}
	entries, err := target.List(ctx, "accessory_subtype", false)
	if err != nil {
		t.Fatal(err)
	}
	if got := masterDataKeys(entries); !slices.Contains(got, "track:straight") {
		t.Fatalf("canonical track subtype missing after import: %#v", got)
	}
}

func TestMasterDataImportNormalizesLegacyAccessorySubtypeTypesBeforeMutation(t *testing.T) {
	ctx := context.Background()
	service := application.NewMasterDataService(testDB(t))

	_, err := service.Import(ctx, &application.MasterDataDocument{
		Format: "railkeeper-master-data", Version: 1,
		Entries: map[string][]application.MasterDataEntry{
			"article_subtype": {{
				Key: "track:legacy_profile", Label: "Legacy profile", Active: true,
			}},
			"vehicle_category": {{
				Type: "article_subtype", Key: "signal:club_signal", Label: "Club signal", Active: true,
			}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	canonical, err := service.List(ctx, "accessory_subtype", false)
	if err != nil {
		t.Fatal(err)
	}
	if got := masterDataKeys(canonical); !slices.Contains(got, "track:legacy_profile") ||
		!slices.Contains(got, "signal:club_signal") {
		t.Fatalf("legacy subtypes were not normalized into the canonical bucket: %#v", got)
	}
	legacy, err := service.List(ctx, "article_subtype", false)
	if err != nil {
		t.Fatal(err)
	}
	if len(legacy) != 0 {
		t.Fatalf("legacy subtype bucket survived import: %#v", legacy)
	}
}

func TestMasterDataImportPreservesAccessorySubtypesWhenCanonicalTypeIsAbsent(t *testing.T) {
	ctx := context.Background()
	service := application.NewMasterDataService(testDB(t))
	before, err := service.List(ctx, "accessory_subtype", false)
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.Import(ctx, &application.MasterDataDocument{
		Format: "railkeeper-master-data", Version: 1,
		Entries: map[string][]application.MasterDataEntry{
			"vehicle_category": {{Key: "lok", Label: "Lok", Active: true}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	after, err := service.List(ctx, "accessory_subtype", false)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(masterDataKeys(after), masterDataKeys(before)) {
		t.Fatalf("legacy import removed canonical accessory subtypes: before=%#v after=%#v",
			masterDataKeys(before), masterDataKeys(after))
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

func TestMasterDataServiceValidatesControlledCustomFieldMetadata(t *testing.T) {
	service := application.NewMasterDataService(testDB(t))
	active := true
	valid := []application.MasterDataInput{
		{Key: "text", Label: "Text", Active: &active, Metadata: map[string]any{"kind": "text"}},
		{Key: "number", Label: "Number", Active: &active, Metadata: map[string]any{
			"kind": "number", "unit": " mm ", "min": 10.0, "max": 20.0,
		}},
		{Key: "boolean", Label: "Boolean", Active: &active, Metadata: map[string]any{"kind": "boolean"}},
		{Key: "date", Label: "Date", Active: &active, Metadata: map[string]any{"kind": "date"}},
		{Key: "single", Label: "Single", Active: &active, Metadata: map[string]any{
			"kind": "single_select", "options": []any{" DCC ", "MM"},
		}},
		{Key: "multi", Label: "Multi", Active: &active, Metadata: map[string]any{
			"kind": "multi_select", "options": []string{"DCC", "MM"},
		}},
	}
	for _, input := range valid {
		created, err := service.Create(t.Context(), "accessory_custom_field", input)
		if err != nil {
			t.Fatalf("valid %s custom field rejected: %v", input.Key, err)
		}
		if input.Key == "number" && (created.Metadata["unit"] != "mm" || created.Metadata["min"] != 10.0) {
			t.Fatalf("number metadata was not normalized: %#v", created.Metadata)
		}
	}

	number, err := service.Get(t.Context(), "accessory_custom_field", "number")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Update(t.Context(), "accessory_custom_field", "number", application.MasterDataInput{
		Label: "Renamed number", Active: &active, Metadata: number.Metadata,
	}); err != nil {
		t.Fatalf("label rename with unchanged metadata rejected: %v", err)
	}

	invalid := []application.MasterDataInput{
		{Key: "missing-kind", Label: "Missing", Metadata: map[string]any{}},
		{Key: "unknown-kind", Label: "Unknown", Metadata: map[string]any{"kind": "json"}},
		{Key: "bad-unit", Label: "Unit", Metadata: map[string]any{"kind": "number", "unit": 12}},
		{Key: "bad-bounds", Label: "Bounds", Metadata: map[string]any{"kind": "number", "min": 20, "max": 10}},
		{Key: "missing-options", Label: "Options", Metadata: map[string]any{"kind": "single_select"}},
		{Key: "duplicate-options", Label: "Options", Metadata: map[string]any{
			"kind": "multi_select", "options": []string{"DCC", "DCC"},
		}},
		{Key: "options-on-text", Label: "Options", Metadata: map[string]any{
			"kind": "text", "options": []string{"DCC"},
		}},
	}
	for _, input := range invalid {
		if _, err := service.Create(t.Context(), "accessory_custom_field", input); !errors.Is(err, application.ErrMasterDataValidation) {
			t.Fatalf("invalid custom field %s error = %v", input.Key, err)
		}
	}
}

func TestMasterDataImportRejectsMalformedControlledCustomFieldsWithoutMutation(t *testing.T) {
	db := testDB(t)
	service := application.NewMasterDataService(db)
	active := true
	if _, err := service.Create(t.Context(), "epoch", application.MasterDataInput{
		Key: "sentinel", Label: "Sentinel", Active: &active,
	}); err != nil {
		t.Fatal(err)
	}
	_, err := service.Import(t.Context(), &application.MasterDataDocument{
		Format: "railkeeper-master-data", Version: 1,
		Entries: map[string][]application.MasterDataEntry{
			"accessory_custom_field": {{
				Type: "accessory_custom_field", Key: "bad", Label: "Bad", Active: true,
				Metadata: map[string]any{"kind": "single_select", "options": []string{}},
			}},
		},
	})
	if !errors.Is(err, application.ErrMasterDataValidation) {
		t.Fatalf("malformed custom field import error = %v", err)
	}
	if _, err := service.Get(t.Context(), "epoch", "sentinel"); err != nil {
		t.Fatalf("malformed import changed existing master data: %v", err)
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
