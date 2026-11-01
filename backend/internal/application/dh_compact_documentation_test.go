package application_test

import (
	"encoding/json"
	"os"
	"testing"

	"railkeeper/backend/internal/decoderlibrary"
	"railkeeper/backend/internal/infrastructure"
)

func TestDHCompactSeedUpgradesOnlyReleasedMetadata(t *testing.T) {
	for _, mode := range []string{"bundled", "edited", "custom"} {
		t.Run(mode, func(t *testing.T) {
			db := testDB(t)
			before := map[string]string{}
			for _, item := range []struct{ file, kind, key string }{
				{"dh18a-2026-10-20.2.json", decoderlibrary.MasterDataType, "dh-dh18a"},
				{"inventory-2026-10-20.3.json", decoderlibrary.InventoryMasterDataType, decoderlibrary.InventoryKey},
			} {
				body, err := os.ReadFile("../decoderlibrary/testdata/" + item.file)
				if err != nil {
					t.Fatal(err)
				}
				var value map[string]any
				_ = json.Unmarshal(body, &value)
				if mode == "edited" {
					value["revision"] = "local-change"
				}
				raw, _ := json.Marshal(value)
				before[item.kind] = string(raw)
				origin := "bundled"
				if mode == "custom" {
					origin = "custom"
				}
				_, err = db.Exec(`INSERT INTO master_data_entries
    (id,type,key,label,active,sort_order,source_url,metadata_json,created_at,updated_at,origin)
    VALUES(?,?,?,'Eigener Titel',0,77,'https://example.test/private',?,'old','old',?)`, item.kind+":"+item.key, item.kind, item.key, string(raw), origin)
				if err != nil {
					t.Fatal(err)
				}
			}
			if err := infrastructure.SeedMasterData(db, "../../seeds"); err != nil {
				t.Fatal(err)
			}
			for kind, original := range before {
				var raw, label, source, updated string
				var active, order int
				err := db.QueryRow(`SELECT metadata_json,label,source_url,active,sort_order,updated_at FROM master_data_entries WHERE type=? AND label='Eigener Titel'`, kind).Scan(&raw, &label, &source, &active, &order, &updated)
				if err != nil {
					t.Fatal(err)
				}
				if active != 0 || order != 77 || source != "https://example.test/private" {
					t.Fatal("operator settings changed")
				}
				var value map[string]any
				_ = json.Unmarshal([]byte(raw), &value)
				if mode == "bundled" {
					expected := "2026-10-20.3"
					if kind == decoderlibrary.InventoryMasterDataType {
						expected = decoderlibrary.BundledInventory().Revision
					}
					if value["revision"] != expected {
						t.Fatal("DH compact evidence not upgraded")
					}
				} else if raw != original || updated != "old" {
					t.Fatal("local metadata overwritten")
				}
			}
			var count int
			if err := db.QueryRow("SELECT count(*) FROM vehicle_cv_values").Scan(&count); err != nil || count != 0 {
				t.Fatal("documentation seeded vehicle readings")
			}
		})
	}
}
