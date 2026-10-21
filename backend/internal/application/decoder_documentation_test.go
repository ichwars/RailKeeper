package application_test

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/decoderlibrary"
	"railkeeper/backend/internal/infrastructure"
)

func TestCVTableSeedUpgradesReleasedDefinitionsAndPreservesLocalChanges(t *testing.T) {
	body, err := os.ReadFile("../decoderlibrary/testdata/catalog-before-cv-tables.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"bundled", "edited", "custom"} {
		t.Run(mode, func(t *testing.T) {
			db := testDB(t)
			var previous decoderlibrary.Catalog
			if err := json.Unmarshal(body, &previous); err != nil {
				t.Fatal(err)
			}
			original := map[string]string{}
			for _, p := range previous.Profiles {
				if mode == "edited" {
					p.Definitions[0].Help.DE += " lokale Ergänzung"
				}
				metadata, _ := json.Marshal(p)
				origin := "bundled"
				if mode == "custom" {
					origin = "custom"
				}
				_, err := db.Exec(`INSERT INTO master_data_entries
 (id,type,key,label,active,sort_order,source_url,metadata_json,created_at,updated_at,origin)
 VALUES(?,'decoder_profile',?,'Eigene Bezeichnung',0,77,'https://example.test/local',?,'old','old',?)`,
					"decoder_profile:"+p.ID, p.ID, string(metadata), origin)
				if err != nil {
					t.Fatal(err)
				}
				original[p.ID] = string(metadata)
			}
			if err := infrastructure.SeedMasterData(db, "../../seeds"); err != nil {
				t.Fatal(err)
			}
			for _, p := range decoderlibrary.Bundled().Profiles {
				var metadata, label, source string
				var active, order int
				if err := db.QueryRow(`SELECT metadata_json,label,source_url,active,sort_order FROM master_data_entries
 WHERE type='decoder_profile' AND key=?`, p.ID).Scan(&metadata, &label, &source, &active, &order); err != nil {
					t.Fatal(err)
				}
				old, existed := original[p.ID]
				if existed && (label != "Eigene Bezeichnung" || source != "https://example.test/local" || active != 0 || order != 77) {
					t.Fatal("operator settings changed during CV documentation upgrade")
				}
				if existed && mode != "bundled" {
					if metadata != old {
						t.Fatal("local documentation overwritten")
					}
				} else {
					var actual decoderlibrary.Profile
					if err := json.Unmarshal([]byte(metadata), &actual); err != nil || !reflect.DeepEqual(actual, p) {
						t.Fatalf("new documentation not persisted losslessly: %s %v", p.ID, err)
					}
				}
			}
			catalog, err := application.NewMasterDataService(db).DecoderLibrary(t.Context(), false)
			if err != nil || len(catalog.Profiles) != len(decoderlibrary.Bundled().Profiles) {
				t.Fatalf("catalog missing: %v", err)
			}
			var count int
			if err := db.QueryRow("SELECT count(*) FROM vehicle_cv_values").Scan(&count); err != nil || count != 0 {
				t.Fatal("factory documentation created vehicle CV readings")
			}
		})
	}
}
