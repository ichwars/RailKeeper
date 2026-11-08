package masterdata

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestCacheSnapshotsOwnNestedMetadata(t *testing.T) {
	source := map[string][]MasterDataEntry{"decoder": {{
		Key: "recorded", Active: true,
		Metadata: map[string]any{
			"nested":  map[string]any{"values": []any{map[string]any{"label": "recorded"}}},
			"options": []string{"original"},
			"number":  json.Number("123.4500"),
			"empty":   []any{},
			"missing": []any(nil),
		},
		Capabilities: &MasterDataCapabilities{CanDelete: true},
	}}}
	caches := MasterDataCaches(source)
	copy := CloneMasterDataMap(caches[false])
	copy["decoder"][0].Metadata["nested"].(map[string]any)["values"].([]any)[0].(map[string]any)["label"] = "changed"
	copy["decoder"][0].Metadata["options"].([]string)[0] = "changed"
	copy["decoder"][0].Capabilities.CanDelete = false
	for _, state := range []map[string][]MasterDataEntry{source, caches[false], caches[true]} {
		entry := state["decoder"][0]
		if entry.Metadata["nested"].(map[string]any)["values"].([]any)[0].(map[string]any)["label"] != "recorded" ||
			entry.Metadata["options"].([]string)[0] != "original" || !entry.Capabilities.CanDelete {
			t.Fatal("returned metadata or capabilities mutated another cache owner")
		}
		if entry.Metadata["number"] != json.Number("123.4500") ||
			!reflect.DeepEqual(entry.Metadata["empty"], []any{}) || !reflect.DeepEqual(entry.Metadata["missing"], []any(nil)) {
			t.Fatal("cloning changed recorded scalar types or missing/empty values")
		}
	}
	// A snapshot loader may retain its returned map; both cached views own their values.
	source["decoder"][0].Metadata["options"].([]string)[0] = "loader changed"
	for _, state := range []map[string][]MasterDataEntry{caches[false], caches[true]} {
		if state["decoder"][0].Metadata["options"].([]string)[0] != "original" {
			t.Fatal("snapshot loader mutated a cached view")
		}
	}
}
