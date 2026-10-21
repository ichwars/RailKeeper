package application_test

import (
	"encoding/json"
	"errors"
	"maps"
	"os"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/decoderlibrary"
	"railkeeper/backend/internal/infrastructure"
)

func TestDecoderLibrarySeedLifecycleAndValidation(t *testing.T) {
	db := testDB(t)
	service := application.NewMasterDataService(db)
	ctx := t.Context()
	if err := infrastructure.SeedMasterData(db, "../../seeds"); err != nil {
		t.Fatal(err)
	}
	catalog, err := service.DecoderLibrary(ctx, false)
	if err != nil || len(catalog.Profiles) != len(decoderlibrary.Bundled().Profiles) || catalog.Inventory == nil || len(catalog.Inventory.Models) != len(decoderlibrary.BundledInventory().Models) {
		t.Fatalf("seed: %+v %v", catalog, err)
	}
	if _, err = db.Exec(`UPDATE master_data_entries SET active=0 WHERE type='decoder_profile' AND key='esu-lokpilot-5-micro-dcc'`); err != nil {
		t.Fatal(err)
	}
	if err = infrastructure.SeedMasterData(db, "../../seeds"); err != nil {
		t.Fatal(err)
	}
	active, err := service.DecoderLibrary(ctx, true)
	if err != nil || len(active.Profiles) != len(decoderlibrary.Bundled().Profiles)-1 || active.Profiles[0].ID != "dcc-basics" {
		t.Fatalf("activation lost on reseed: %+v %v", active, err)
	}
	profile := decoderlibrary.Bundled().Profiles[1]
	profile.ID = "local-decoder"
	body, _ := json.Marshal(profile)
	var metadata map[string]any
	_ = json.Unmarshal(body, &metadata)
	if _, err = service.Create(ctx, "decoder_profile", application.MasterDataInput{Key: profile.ID, Label: profile.Name, Metadata: metadata}); err != nil {
		t.Fatal(err)
	}
	metadata["protocol"] = "MM"
	if _, err = service.Update(ctx, "decoder_profile", profile.ID, application.MasterDataInput{Label: profile.Name, Metadata: metadata}); !errors.Is(err, application.ErrMasterDataValidation) {
		t.Fatalf("invalid update: %v", err)
	}
	doc, err := service.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for i := range doc.Entries["decoder_profile"] {
		doc.Entries["decoder_profile"][i].Metadata = maps.Clone(doc.Entries["decoder_profile"][i].Metadata)
		doc.Entries["decoder_profile"][i].Metadata["protocol"] = "MM"
	}
	if _, err = service.Import(ctx, doc); !errors.Is(err, application.ErrMasterDataValidation) {
		t.Fatalf("invalid import: %v", err)
	}
	saved, err := service.DecoderLibrary(ctx, false)
	if err != nil || len(saved.Profiles) != len(decoderlibrary.Bundled().Profiles)+1 {
		t.Fatalf("invalid import mutated catalog: %+v %v", saved, err)
	}
	validDocument, err := service.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	target := application.NewMasterDataService(testDB(t))
	if _, err := target.Import(ctx, validDocument); err != nil {
		t.Fatal(err)
	}
	transferred, err := target.DecoderLibrary(ctx, false)
	if err != nil {
		t.Fatal(err)
	}
	want, _ := json.Marshal(saved)
	got, _ := json.Marshal(transferred)
	if string(want) != string(got) {
		t.Fatal("master-data transfer changed decoder definitions or activation")
	}
}

func TestDecoderLibraryBackupPreservesProfilesAndRejectsInvalidMetadata(t *testing.T) {
	dir := t.TempDir()
	db := backupTestDB(t, dir)
	ctx := t.Context()
	if err := infrastructure.SeedMasterData(db, "../../seeds"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE master_data_entries SET active=0 WHERE type='decoder_profile' AND key='esu-lokpilot-5-micro-dcc'`); err != nil {
		t.Fatal(err)
	}
	service := application.NewBackupService(db, dir)
	doc, err := service.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = service.Import(ctx, doc); err != nil {
		t.Fatal(err)
	}
	catalog, err := application.NewMasterDataService(db).DecoderLibrary(ctx, false)
	if err != nil || len(catalog.Profiles) != len(decoderlibrary.Bundled().Profiles) {
		t.Fatalf("restored catalog: %+v %v", catalog, err)
	}
	for _, profile := range catalog.Profiles {
		if profile.ID == "esu-lokpilot-5-micro-dcc" && profile.Active {
			t.Fatal("restored activation changed")
		}
	}
	for _, row := range doc.Tables["master_data_entries"] {
		if row["type"] != "decoder_catalog" {
			continue
		}
		original := row["metadata_json"]
		row["metadata_json"] = `{"revision":"broken-inventory"}`
		validation, err := service.Validate(ctx, doc)
		if err != nil || validation.Compatible {
			t.Fatalf("invalid inventory backup accepted: %+v %v", validation, err)
		}
		row["metadata_json"] = original
	}
	var damaged map[string]any
	for _, row := range doc.Tables["master_data_entries"] {
		if row["type"] == "decoder_profile" {
			damaged = row
			break
		}
	}
	original := damaged["metadata_json"]
	damaged["metadata_json"] = `{"id":"broken"}`
	validation, err := service.Validate(ctx, doc)
	if err != nil || validation.Compatible || !strings.Contains(strings.Join(validation.Errors, " "), "Decoderprofil") {
		t.Fatalf("bad metadata accepted: %+v %v", validation, err)
	}
	if _, err = service.Import(ctx, doc); !errors.Is(err, application.ErrBackupInvalid) {
		t.Fatalf("bad restore: %v", err)
	}
	damaged["metadata_json"] = original
	// A backup from before the library gets definitions without requiring an application restart.
	rows := doc.Tables["master_data_entries"][:0]
	for _, row := range doc.Tables["master_data_entries"] {
		if row["type"] != "decoder_profile" {
			if row["type"] != "decoder_catalog" {
				rows = append(rows, row)
			}
		}
	}
	doc.Tables["master_data_entries"] = rows
	if _, err = service.Import(ctx, doc); err != nil {
		t.Fatal(err)
	}
	catalog, err = application.NewMasterDataService(db).DecoderLibrary(ctx, true)
	if err != nil || len(catalog.Profiles) != len(decoderlibrary.Bundled().Profiles) || catalog.Inventory == nil || len(catalog.Inventory.Models) != len(decoderlibrary.BundledInventory().Models) {
		t.Fatalf("legacy backup missing catalog: %+v %v", catalog, err)
	}
}

func TestInventoryEditsSurviveSeedAndInvalidImportIsAtomic(t *testing.T) {
	db := testDB(t)
	if err := infrastructure.SeedMasterData(db, "../../seeds"); err != nil {
		t.Fatal(err)
	}
	service := application.NewMasterDataService(db)
	inventory := decoderlibrary.BundledInventory()
	inventory.Revision = "local-reviewed-revision"
	body, _ := json.Marshal(inventory)
	var metadata map[string]any
	_ = json.Unmarshal(body, &metadata)
	if _, err := service.Update(t.Context(), "decoder_catalog", "library", application.MasterDataInput{
		Label: "Decoderbibliothek", Metadata: metadata,
	}); err != nil {
		t.Fatal(err)
	}
	if err := infrastructure.SeedMasterData(db, "../../seeds"); err != nil {
		t.Fatal(err)
	}
	catalog, err := service.DecoderLibrary(t.Context(), false)
	if err != nil || catalog.Inventory == nil || catalog.Inventory.Revision != "local-reviewed-revision" {
		t.Fatalf("local inventory overwritten: %+v %v", catalog, err)
	}
	doc, err := service.Export(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	doc.Entries["decoder_catalog"][0].Metadata["models"] = []any{}
	if _, err := service.Import(t.Context(), doc); !errors.Is(err, application.ErrMasterDataValidation) {
		t.Fatalf("invalid inventory accepted: %v", err)
	}
	catalog, err = service.DecoderLibrary(t.Context(), false)
	if err != nil || catalog.Inventory == nil || len(catalog.Inventory.Models) != len(decoderlibrary.BundledInventory().Models) {
		t.Fatalf("failed import mutated catalog: %+v %v", catalog, err)
	}
}

func TestBundledDecoderUpdatesPreserveOperatorChanges(t *testing.T) {
	for _, fixture := range []struct{ key, file, revision string }{
		{"kuehn-n45", "kuehn-n45-2026-10-20.1.json", "2026-10-20.2"},
		{"lenz-standard-plus-v2", "lenz-standard-plus-v2-2026-10-20.1.json", "2026-10-20.2"},
		{"zimo-mn180n18", "mn180n18-2026-10-20.1.json", "2026-10-20.3"},
		{"zimo-mn180n18", "mn180n18-2026-10-20.2.json", "2026-10-20.3"},
		{"dh-dh18a", "dh18a-2026-10-18.2.json", "2026-10-20.3"},
		{"dh-dh18a", "dh18a-2026-10-20.1.json", "2026-10-20.3"},
		{"esu-lokpilot-5-micro-dcc", "esu-lokpilot-5-micro-dcc-2026-10-18.3.json", "2026-10-20.2"},
		{"esu-lokpilot-5-micro-dcc", "esu-lokpilot-5-micro-dcc-2026-10-18.4.json", "2026-10-20.2"},
		{"esu-lokpilot-5-micro-dcc", "esu-lokpilot-5-micro-dcc-2026-10-20.1.json", "2026-10-20.2"},
	} {
		body, err := os.ReadFile("../decoderlibrary/testdata/" + fixture.file)
		if err != nil {
			t.Fatal(err)
		}
		tests := []struct {
			name                         string
			edited, custom, legacyActive bool
		}{
			{name: "unchanged bundled"}, {name: "edited metadata", edited: true}, {name: "custom origin", custom: true},
		}
		if fixture.file == "esu-lokpilot-5-micro-dcc-2026-10-20.1.json" {
			for _, state := range []string{"bundled", "edited", "custom"} {
				tests = append(tests, struct {
					name                         string
					edited, custom, legacyActive bool
				}{"legacy active metadata/" + state, state == "edited", state == "custom", true})
			}
		}
		for _, test := range tests {
			t.Run(fixture.key+"/"+test.name, func(t *testing.T) {
				db := testDB(t)
				metadata := string(body)
				if test.edited || test.legacyActive {
					var document map[string]any
					_ = json.Unmarshal(body, &document)
					if test.legacyActive {
						document["active"] = true
					}
					if test.edited {
						document["name"] = "Mein Decoderprofil"
					}
					changed, _ := json.Marshal(document)
					metadata = string(changed)
				}
				origin := "bundled"
				if test.custom {
					origin = "custom"
				}
				_, err = db.Exec(`INSERT INTO master_data_entries
    (id,type,key,label,active,sort_order,source_url,metadata_json,created_at,updated_at,origin)
    VALUES(?,'decoder_profile',?,'Meine Beschriftung',0,99,'https://example.test/local',?,'old','old',?)`, "decoder_profile:"+fixture.key, fixture.key, metadata, origin)
				if err != nil {
					t.Fatal(err)
				}
				if err := infrastructure.SeedMasterData(db, "../../seeds"); err != nil {
					t.Fatal(err)
				}
				var saved, label, source, updated string
				var active, order int
				if err := db.QueryRow(`SELECT metadata_json,label,source_url,updated_at,active,sort_order FROM master_data_entries WHERE type='decoder_profile' AND key=?`, fixture.key).Scan(&saved, &label, &source, &updated, &active, &order); err != nil {
					t.Fatal(err)
				}
				if label != "Meine Beschriftung" || source != "https://example.test/local" || active != 0 || order != 99 {
					t.Fatal("operator settings changed")
				}
				if test.edited || test.custom {
					if saved != metadata || updated != "old" {
						t.Fatal("local profile overwritten")
					}
				} else {
					var profile decoderlibrary.Profile
					_ = json.Unmarshal([]byte(saved), &profile)
					if profile.Revision != fixture.revision || len(profile.Definitions) <= 5 {
						t.Fatal("pristine old documentation not upgraded")
					}
				}
				if err := infrastructure.SeedMasterData(db, "../../seeds"); err != nil {
					t.Fatal(err)
				}
				var repeated string
				if err := db.QueryRow(`SELECT metadata_json FROM master_data_entries WHERE type='decoder_profile' AND key=?`, fixture.key).Scan(&repeated); err != nil || repeated != saved {
					t.Fatal("seed not idempotent")
				}
			})
		}
	}
}

func TestInventoryUpgradePreservesOperatorSettingsAndLocalEdits(t *testing.T) {
	for _, fixture := range []string{"inventory-2026-10-18.3.json", "inventory-2026-10-20.1.json", "inventory-2026-10-20.2.json"} {
		body, err := os.ReadFile("../decoderlibrary/testdata/" + fixture)
		if err != nil {
			t.Fatal(err)
		}
		for _, state := range []string{"bundled", "edited", "custom"} {
			t.Run(fixture+"/"+state, func(t *testing.T) {
				db := testDB(t)
				metadata, origin := string(body), "bundled"
				if state == "edited" {
					var data map[string]any
					_ = json.Unmarshal(body, &data)
					data["revision"] = "my-catalog"
					changed, _ := json.Marshal(data)
					metadata = string(changed)
				}
				if state == "custom" {
					origin = "custom"
				}
				_, err := db.Exec(`INSERT INTO master_data_entries
			(id,type,key,label,active,sort_order,source_url,metadata_json,created_at,updated_at,origin)
			VALUES('decoder_catalog:library','decoder_catalog','library','Meine Bibliothek',0,27,
			'https://example.test/local',?,'old','old',?)`, metadata, origin)
				if err != nil {
					t.Fatal(err)
				}
				if err := infrastructure.SeedMasterData(db, "../../seeds"); err != nil {
					t.Fatal(err)
				}
				var saved, label, source string
				var active, order int
				if err := db.QueryRow(`SELECT metadata_json,label,source_url,active,sort_order
			FROM master_data_entries WHERE type='decoder_catalog' AND key='library'`).Scan(
					&saved, &label, &source, &active, &order); err != nil {
					t.Fatal(err)
				}
				if label != "Meine Bibliothek" || source != "https://example.test/local" || active != 0 || order != 27 {
					t.Fatal("operator settings changed during catalog upgrade")
				}
				if state != "bundled" {
					if saved != metadata {
						t.Fatal("local inventory overwritten")
					}
					return
				}
				var data map[string]any
				_ = json.Unmarshal([]byte(saved), &data)
				inventory, err := decoderlibrary.ParseInventory(data)
				if err != nil || inventory.Revision != decoderlibrary.BundledInventory().Revision || len(inventory.Manufacturers) != 8 || len(inventory.Models[0].Variants) == 0 {
					t.Fatalf("documented variant upgrade missing: %v", err)
				}
			})
		}
	}
}
