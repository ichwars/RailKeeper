package application_test

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestFunctionSymbolCodeUpgradePreservesLocalSettingsAndMetadata(t *testing.T) {
	db := testDB(t)
	service := application.NewMasterDataService(db)
	ctx := context.Background()
	keys := []string{"esu-f009-rangiergang", "esu-f010-abv", "esu-f006-fahrgeraeusch", "esu-f004-innenraumbeleuchtung"}
	before := map[string]map[string]any{}
	for i, key := range keys {
		entry, err := service.Get(ctx, "symbols", key)
		if err != nil {
			t.Fatal(err)
		}
		entry.Metadata["ecosCode"] = []int{9, 10, 6, 4}[i]
		if i == 1 {
			entry.Metadata["description"] = "Meine eigene Bedeutung"
		}
		body, _ := json.Marshal(entry.Metadata)
		before[key] = entry.Metadata
		if _, err := db.Exec(`UPDATE master_data_entries SET metadata_json=?,label='Eigenes Etikett',
 active=0,sort_order=999,source_url='https://example.org/local' WHERE type='symbols' AND key=?`, string(body), key); err != nil {
			t.Fatal(err)
		}
	}
	// Even pristine metadata with a custom origin is not a bundled upgrade candidate.
	if _, err := db.Exec(`UPDATE master_data_entries SET origin='custom' WHERE type='symbols' AND key=?`, keys[3]); err != nil {
		t.Fatal(err)
	}
	upgrade := func() {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = tx.Rollback() }()
		if err := application.SeedFunctionSymbolCodes(ctx, tx, "2026-10-27T12:00:00Z"); err != nil {
			t.Fatal(err)
		}
		if err := tx.Commit(); err != nil {
			t.Fatal(err)
		}
	}
	upgrade()
	for i, key := range keys {
		entry, err := service.Get(ctx, "symbols", key)
		if err != nil {
			t.Fatal(err)
		}
		expected := before[key]
		if i == 0 {
			expected["ecosCode"] = float64(10)
		}
		if i == 2 {
			delete(expected, "ecosCode")
		}
		// Round-trip numeric metadata through JSON, as the application does.
		raw, _ := json.Marshal(expected)
		_ = json.Unmarshal(raw, &expected)
		if !reflect.DeepEqual(entry.Metadata, expected) || entry.Label != "Eigenes Etikett" || entry.Active ||
			entry.SortOrder != 999 || entry.SourceURL != "https://example.org/local" {
			t.Fatalf("local symbol settings changed or code not corrected: %s", key)
		}
	}
	first, _ := service.List(ctx, "symbols", true)
	upgrade()
	second, _ := service.List(ctx, "symbols", true)
	if !reflect.DeepEqual(first, second) {
		t.Fatal("symbol upgrade is not idempotent")
	}
}
