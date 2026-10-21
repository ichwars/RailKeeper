package application_test

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/decoderlibrary"
	"railkeeper/backend/internal/infrastructure"
)

func TestReviewedGapMetadataUpgradesWithoutReplacingOperatorChanges(t *testing.T) {
	for _, fixture := range []string{
		"reference-gaps-2026-10-20.1.json", "reference-gaps-2026-10-21.3.json", "roco-gap-2026-10-21.3.json",
		"reference-gaps-before-complete-2026-10-21.json",
	} {
		t.Run(fixture, func(t *testing.T) { testReviewedGapMetadataUpgrade(t, fixture) })
	}
}

func testReviewedGapMetadataUpgrade(t *testing.T, fixture string) {
	t.Helper()
	body, err := os.ReadFile("../decoderlibrary/testdata/" + fixture)
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
					old.Note.DE = "Eigene Quellenprüfung"
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
			for _, old := range previous {
				var label, source, stored string
				var active, order int
				err := db.QueryRow(`SELECT label,active,sort_order,source_url,metadata_json FROM master_data_entries
 WHERE type='decoder_reference' AND key=?`, old.ID).Scan(&label, &active, &order, &source, &stored)
				if err != nil || label != "Meine Quelle" || active != 0 || order != 777 || source != "https://example.test/local" {
					t.Fatalf("operator settings changed for %s: %v", old.ID, err)
				}
				if tc.edited || tc.origin == "custom" {
					if stored != before[old.ID] {
						t.Fatalf("local metadata overwritten: %s", old.ID)
					}
				} else {
					var upgraded decoderlibrary.CVReference
					if err := json.Unmarshal([]byte(stored), &upgraded); err != nil || !reflect.DeepEqual(upgraded, current[old.ID]) {
						t.Fatalf("released gap did not upgrade: %s, %v", old.ID, err)
					}
				}
			}
			refs, err := application.NewMasterDataService(db).DecoderReferences(t.Context(), "piko-sound-products-36520", true)
			if err != nil || len(refs) != 1 || refs[0].ID != "cv-ref-piko-36520-br491" || len(refs[0].Rows) != 254 {
				t.Fatalf("new exact-project reference did not persist: %v", err)
			}
			service := application.NewMasterDataService(db)
			for _, model := range []string{"lenz-legacy-le030", "lenz-legacy-le0511a", "lenz-legacy-le0511d"} {
				refs, err := service.DecoderReferences(t.Context(), model, true)
				if err != nil || len(refs) != 1 || refs[0].Status != "partial" || len(refs[0].Rows) == 0 {
					t.Fatalf("original Lenz manual was not seeded for %s: %v", model, err)
				}
			}
			if tc.name == "released" {
				doc, err := service.Export(t.Context())
				if err != nil {
					t.Fatal(err)
				}
				target := application.NewMasterDataService(testDB(t))
				if _, err := target.Import(t.Context(), doc); err != nil {
					t.Fatal(err)
				}
				for _, model := range []string{"lenz-legacy-le030", "lenz-legacy-le0511a", "lenz-legacy-le0511d"} {
					before, err := service.DecoderReferences(t.Context(), model, true)
					if err != nil {
						t.Fatal(err)
					}
					after, err := target.DecoderReferences(t.Context(), model, true)
					if err != nil || !reflect.DeepEqual(before, after) {
						t.Fatalf("Lenz source/range facts changed during transfer: %s, %v", model, err)
					}
				}
			}
		})
	}
}

func TestKuehnReferenceUpgradePreservesOperatorSettingsAndLocalEdits(t *testing.T) {
	body, err := os.ReadFile("../decoderlibrary/testdata/kuehn-n45-reference-2026-10-20.1.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, edited := range []bool{false, true} {
		t.Run(map[bool]string{false: "bundled", true: "edited"}[edited], func(t *testing.T) {
			db := testDB(t)
			var previous map[string]any
			_ = json.Unmarshal(body, &previous)
			if edited {
				previous["revision"] = "my-reference"
			}
			metadata, _ := json.Marshal(previous)
			_, err := db.Exec(`INSERT INTO master_data_entries
 (id,type,key,label,active,sort_order,source_url,metadata_json,created_at,updated_at,origin)
 VALUES('reference-test','decoder_reference','cv-ref-kuehn-n45','Meine Quelle',0,777,
 'https://example.test/local',?,'before','before','bundled')`, string(metadata))
			if err != nil {
				t.Fatal(err)
			}
			if err := infrastructure.SeedMasterData(db, "../../seeds"); err != nil {
				t.Fatal(err)
			}
			var label, source, stored string
			var active, order int
			err = db.QueryRow(`SELECT label,active,sort_order,source_url,metadata_json FROM master_data_entries
 WHERE type='decoder_reference' AND key='cv-ref-kuehn-n45'`).Scan(&label, &active, &order, &source, &stored)
			if err != nil || label != "Meine Quelle" || active != 0 || order != 777 || source != "https://example.test/local" {
				t.Fatalf("operator reference settings overwritten: %v", err)
			}
			if edited {
				if stored != string(metadata) {
					t.Fatal("edited N45 reference overwritten")
				}
			} else {
				var upgraded decoderlibrary.CVReference
				if err := json.Unmarshal([]byte(stored), &upgraded); err != nil || upgraded.Revision != "2026-10-20.2" || len(upgraded.Rows) != 59 {
					t.Fatalf("known N45 reference did not upgrade: %v", err)
				}
			}
		})
	}
}

func TestDecoderReferencesPersistTransferAndProtectLocalMetadata(t *testing.T) {
	db := testDB(t)
	ctx := t.Context()
	if err := infrastructure.SeedMasterData(db, "../../seeds"); err != nil {
		t.Fatal(err)
	}
	service := application.NewMasterDataService(db)
	refs, err := service.DecoderReferences(ctx, "kuehn-n-n025", true)
	if err != nil || len(refs) != 1 || len(refs[0].Rows) == 0 {
		t.Fatalf("reference missing: %v", err)
	}
	original := refs[0]
	edited := original
	edited.Title.DE = "Lokaler Nachweis"
	body, _ := json.Marshal(edited)
	var metadata map[string]any
	_ = json.Unmarshal(body, &metadata)
	if _, err = service.Update(ctx, decoderlibrary.ReferenceMasterDataType, edited.ID, application.MasterDataInput{Label: "Eigene Quelle", Metadata: metadata}); err != nil {
		t.Fatal(err)
	}
	if err = infrastructure.SeedMasterData(db, "../../seeds"); err != nil {
		t.Fatal(err)
	}
	refs, err = service.DecoderReferences(ctx, "kuehn-n-n025", true)
	if err != nil || !reflect.DeepEqual(refs[0], edited) {
		t.Fatal("reseed changed local reference")
	}
	doc, err := service.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	target := application.NewMasterDataService(testDB(t))
	if _, err = target.Import(ctx, doc); err != nil {
		t.Fatal(err)
	}
	transferred, err := target.DecoderReferences(ctx, "kuehn-n-n025", true)
	if err != nil || !reflect.DeepEqual(refs, transferred) {
		t.Fatal("master-data transfer changed reference facts")
	}
	metadata["status"] = "unavailable"
	if _, err = service.Update(ctx, decoderlibrary.ReferenceMasterDataType, edited.ID, application.MasterDataInput{Label: "Invalid", Metadata: metadata}); !errors.Is(err, application.ErrMasterDataValidation) {
		t.Fatalf("invalid reference accepted: %v", err)
	}
	if _, err = service.SetActive(ctx, decoderlibrary.ReferenceMasterDataType, edited.ID, false); err != nil {
		t.Fatal(err)
	}
	refs, err = service.DecoderReferences(ctx, "kuehn-n-n025", true)
	if err != nil || len(refs) != 0 {
		t.Fatal("inactive reference returned")
	}
}

func TestBackupContainsReferencesAndRejectsInvalidBanks(t *testing.T) {
	dir := t.TempDir()
	db := backupTestDB(t, dir)
	ctx := t.Context()
	if err := infrastructure.SeedMasterData(db, "../../seeds"); err != nil {
		t.Fatal(err)
	}
	backup := application.NewBackupService(db, dir)
	doc, err := backup.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var changed map[string]any
	count := 0
	for _, row := range doc.Tables["master_data_entries"] {
		if row["type"] == decoderlibrary.ReferenceMasterDataType {
			count++
			if changed == nil {
				changed = row
			}
		}
	}
	if count != len(decoderlibrary.BundledReferences()) {
		t.Fatal("backup omitted manufacturer references")
	}
	if _, err = backup.Import(ctx, doc); err != nil {
		t.Fatal(err)
	}
	before, err := application.NewMasterDataService(db).DecoderReferences(ctx, "", false)
	if err != nil {
		t.Fatal(err)
	}
	var metadata map[string]any
	_ = json.Unmarshal([]byte(changed["metadata_json"].(string)), &metadata)
	metadata["sources"] = []any{}
	body, _ := json.Marshal(metadata)
	changed["metadata_json"] = string(body)
	validation, err := backup.Validate(ctx, doc)
	if err != nil || validation.Compatible {
		t.Fatal("invalid reference backup accepted")
	}
	after, err := application.NewMasterDataService(db).DecoderReferences(ctx, "", false)
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatal("failed backup validation changed reference data")
	}
}
