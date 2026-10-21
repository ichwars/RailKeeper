package application

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestLargeMasterDataPreservesActivationOrderAndMetadata(t *testing.T) {
	db := masterDataCacheTestDB(t)
	const count = 160
	for i := 0; i < count; i++ {
		metadata, _ := json.Marshal(map[string]any{"reference": strings.Repeat("CV documentation ", 4096)})
		_, err := db.Exec(`INSERT INTO master_data_entries
 (id,type,key,label,active,sort_order,source_url,metadata_json,created_at,updated_at,origin)
 VALUES(?,?,?,?,?,?,?,?,?,?,'bundled')`, fmt.Sprint(i), "large", fmt.Sprint(i),
			fmt.Sprintf("Modell %03d", i), i%2 == 0, i%3, "", string(metadata), "now", "now")
		if err != nil {
			t.Fatal(err)
		}
	}
	service := NewMasterDataService(db)
	if err := service.WarmCache(t.Context()); err != nil {
		t.Fatal(err)
	}
	listed, err := service.List(t.Context(), "large", false)
	if err != nil {
		t.Fatal(err)
	}
	all, err := service.ListAll(t.Context(), false)
	if err != nil || len(listed) != count || len(all["large"]) != count {
		t.Fatalf("large metadata lost: %v", err)
	}
	for _, values := range [][]MasterDataEntry{listed, all["large"]} {
		for i, entry := range values {
			if entry.Metadata["reference"] != strings.Repeat("CV documentation ", 4096) {
				t.Fatal("metadata changed during sorting")
			}
			if i == 0 {
				continue
			}
			previous := values[i-1]
			if (!previous.Active && entry.Active) || (previous.Active == entry.Active &&
				(previous.SortOrder > entry.SortOrder || (previous.SortOrder == entry.SortOrder && previous.Label > entry.Label))) {
				t.Fatal("activation, sort order or label ordering changed")
			}
		}
	}
	active, err := service.List(t.Context(), "large", true)
	if err != nil || len(active) != count/2 {
		t.Fatalf("active entries changed: %v", err)
	}
}
