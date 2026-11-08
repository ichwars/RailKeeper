package application

import (
	"reflect"
	"testing"
)

func TestManufacturerMetadataFixturePreservesSearchInputAndFiltersUnusableDomains(t *testing.T) {
	for _, metadata := range []string{
		`{"aliases":["Recorded Alias","Recorded Alias",""],"searchDomains":["reference.example","forum.reference.example","en.wikipedia.org","www.catalog.example"],"website":"https://www.reference.example/manuals"}`,
		`{"aliases":"Recorded Alias, Recorded Alias, ","searchDomains":"reference.example, forum.reference.example, en.wikipedia.org, www.catalog.example","website":"https://www.reference.example/manuals"}`,
	} {
		db := masterDataCacheTestDB(t)
		entry := masterDataCacheEntry("Canonical Maker", true)
		entry.ID, entry.Type, entry.Key = "manufacturer:recorded", "manufacturer", "canonical-maker"
		insertMasterDataCacheEntry(t, db, entry)
		if _, err := db.Exec(`UPDATE master_data_entries SET metadata_json=? WHERE id=?`, metadata, entry.ID); err != nil {
			t.Fatal(err)
		}
		service := NewArticleSearchService(NewMasterDataService(db))
		input := ArticleSearchInput{Manufacturer: "Recorded Alias", ArticleNumber: "000123", Gauge: "TT", Name: "Recorded Name",
			ManufacturerAliases: []string{"existing alias"}, PreferredDomains: []string{"existing.example"}, Fields: map[string]string{"ean": "0000123456789"}}
		got := service.withManufacturerMetadata(t.Context(), input)
		if got.Manufacturer != input.Manufacturer || got.ArticleNumber != input.ArticleNumber || got.Gauge != input.Gauge || got.Name != input.Name ||
			!reflect.DeepEqual(got.Fields, input.Fields) || !reflect.DeepEqual(got.ManufacturerAliases, []string{"existing alias", "Recorded Alias"}) ||
			!reflect.DeepEqual(got.PreferredDomains, []string{"existing.example", "reference.example", "catalog.example"}) {
			t.Fatal("recorded alias or search preferences changed", got)
		}
		if _, err := db.Exec(`UPDATE master_data_entries SET active=0 WHERE id=?`, entry.ID); err != nil {
			t.Fatal(err)
		}
		if got := service.withManufacturerMetadata(t.Context(), input); !reflect.DeepEqual(got, input) {
			t.Fatal("inactive manufacturer influenced suggestions", got)
		}
		if err := db.Close(); err != nil {
			t.Fatal(err)
		}
		if got := service.withManufacturerMetadata(t.Context(), input); !reflect.DeepEqual(got, input) {
			t.Fatal("metadata outage discarded recorded search input", got)
		}
	}
}
