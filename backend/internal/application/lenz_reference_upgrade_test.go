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

func TestLenzManualReferencesUpgradeAndTransferWithoutReplacingLocalChanges(t *testing.T) {
	body, err := os.ReadFile("../decoderlibrary/testdata/lenz-legacy-five-2026-10-21.3.json")
	if err != nil {
		t.Fatal(err)
	}
	var previous []decoderlibrary.CVReference
	if err = json.Unmarshal(body, &previous); err != nil {
		t.Fatal(err)
	}
	current := map[string]decoderlibrary.CVReference{}
	for _, ref := range decoderlibrary.BundledReferences() {
		current[ref.ID] = ref
	}
	for _, tc := range []struct {
		name, origin string
		edited       bool
	}{{"released", "bundled", false}, {"edited", "bundled", true}, {"custom", "custom", false}} {
		t.Run(tc.name, func(t *testing.T) {
			db := testDB(t)
			before := map[string]string{}
			for _, old := range previous {
				if tc.edited {
					old.Note.DE = "Meine Quellenprüfung"
				}
				metadata, _ := json.Marshal(old)
				before[old.ID] = string(metadata)
				_, err := db.Exec(`INSERT INTO master_data_entries
 (id,type,key,label,active,sort_order,source_url,metadata_json,created_at,updated_at,origin)
 VALUES(?, 'decoder_reference', ?, 'Meine Quelle', 0, 777,
 'https://example.test/local', ?, 'before', 'before', ?)`, "legacy-"+old.ID, old.ID, string(metadata), tc.origin)
				if err != nil {
					t.Fatal(err)
				}
			}
			for range 2 {
				if err := infrastructure.SeedMasterData(db, "../../seeds"); err != nil {
					t.Fatal(err)
				}
			}
			service := application.NewMasterDataService(db)
			for _, old := range previous {
				var label, source, stored, origin string
				var active, order int
				err := db.QueryRow(`SELECT label,active,sort_order,source_url,metadata_json,origin
 FROM master_data_entries WHERE type='decoder_reference' AND key=?`, old.ID).
					Scan(&label, &active, &order, &source, &stored, &origin)
				if err != nil || label != "Meine Quelle" || active != 0 || order != 777 ||
					source != "https://example.test/local" || origin != tc.origin {
					t.Fatalf("operator settings changed for %s: %v", old.ID, err)
				}
				if tc.edited || tc.origin == "custom" {
					if stored != before[old.ID] {
						t.Fatalf("local metadata overwritten: %s", old.ID)
					}
				} else {
					var upgraded decoderlibrary.CVReference
					if err := json.Unmarshal([]byte(stored), &upgraded); err != nil || !reflect.DeepEqual(upgraded, current[old.ID]) {
						t.Fatalf("released reference did not upgrade: %s, %v", old.ID, err)
					}
				}
				refs, err := service.DecoderReferences(t.Context(), old.ModelIDs[0], false)
				if err != nil || len(refs) != 1 || refs[0].ID != old.ID || refs[0].Active {
					t.Fatalf("duplicate reference or inactive state lost for %s: %v", old.ID, err)
				}
			}
			if tc.name != "released" {
				return
			}
			doc, err := service.Export(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			target := application.NewMasterDataService(testDB(t))
			if _, err = target.Import(t.Context(), doc); err != nil {
				t.Fatal(err)
			}
			for _, old := range previous {
				transferred, err := target.DecoderReferences(t.Context(), old.ModelIDs[0], false)
				expected := current[old.ID]
				expected.Active = false
				if err != nil || len(transferred) != 1 || !reflect.DeepEqual(transferred[0], expected) {
					t.Fatalf("source facts changed during transfer: %s, %v", old.ID, err)
				}
			}
		})
	}
}
