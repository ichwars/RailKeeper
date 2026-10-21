package decoderlibrary

import (
	"encoding/json"
	"os"
	"reflect"
	"slices"
	"testing"
)

func TestInventoryReferencesReviewedProfilesAndDistinctManufacturers(t *testing.T) {
	inventory := BundledInventory()
	profiles := Bundled().Profiles
	if len(inventory.Manufacturers) != 8 || len(inventory.Families) != 40 || len(inventory.Models) != 599 {
		t.Fatal("incomplete researched inventory")
	}
	for _, model := range inventory.Models {
		profileID := model.ProfileID
		if profileID == "" {
			profileID = model.DocumentationProfileID
		}
		if profileID == "" {
			continue
		}
		var profile *Profile
		for i := range profiles {
			if profiles[i].ID == profileID {
				profile = &profiles[i]
			}
		}
		if profile == nil || profile.Model != model.Name {
			t.Fatalf("model links to unrelated profile: %+v", model)
		}
	}
}

func TestInventoryOrderVariantsHaveUnambiguousConnectionsAndSources(t *testing.T) {
	inventory := BundledInventory()
	for _, test := range []struct{ maker, article, connection string }{
		{"esu", "59827", "NEM 651 (direkt)"},
		{"esu", "59857", "NEM 651 (direkt, gewinkelt)"},
		{"esu", "59222", "PluX22"},
		{"doehler-haass", "DH10C-2", "Kabel (NEM 651 Fleischmann)"},
		{"doehler-haass", "SD21A-5", "21-polig (NEM 660/RCN-121)"},
		{"zimo", "MX645P22", "PluX22"},
		{"lenz", "10433-01", "NEM 652"},
		{"kuehn", "82350", "Next18"},
		{"uhlenbrock", "34760", "PluX22"},
		{"roco", "10735", "NEM 651 (direkt)"},
		{"piko", "56514", "PIKO 5- und 7-polig / 5- and 7-pin"},
	} {
		found := false
		for _, model := range inventory.Models {
			for _, variant := range model.Variants {
				if variant.ArticleNumber != test.article {
					continue
				}
				family := inventory.Families[slices.IndexFunc(inventory.Families, func(f Family) bool { return f.ID == model.FamilyID })]
				if family.ManufacturerID == test.maker && variant.Interface == test.connection && validSource(variant.Source) {
					found = true
				}
			}
		}
		if !found {
			t.Fatalf("missing documented order variant: %+v", test)
		}
	}
	for _, maker := range inventory.Manufacturers {
		if maker.Coverage == nil || !validText(*maker.Coverage) {
			t.Fatalf("manufacturer lacks honest coverage statement: %s", maker.ID)
		}
	}
}

func TestLegacyInventoryAndBundledUpgradeDigestRemainCompatible(t *testing.T) {
	body, err := os.ReadFile("testdata/inventory-2026-10-18.3.json")
	if err != nil {
		t.Fatal(err)
	}
	var metadata map[string]any
	if err := json.Unmarshal(body, &metadata); err != nil {
		t.Fatal(err)
	}
	old, err := ParseInventory(metadata)
	if err != nil || len(old.Models) != 55 || len(old.Manufacturers) != 4 ||
		!PreviousBundled(InventoryMasterDataType, InventoryKey, string(body)) {
		t.Fatalf("legacy inventory or upgrade digest lost: %v", err)
	}
	metadata["revision"] = "local-customization"
	changed, _ := json.Marshal(metadata)
	if PreviousBundled(InventoryMasterDataType, InventoryKey, string(changed)) {
		t.Fatal("local catalog change treated as released metadata")
	}
}

func TestInventoryValidationRejectsBrokenHierarchyAndUnsafeSources(t *testing.T) {
	for _, change := range []func(*Inventory){
		func(i *Inventory) { i.Models[0].FamilyID = "missing" },
		func(i *Inventory) { i.Families[0].ManufacturerID = "missing" },
		func(i *Inventory) { i.Models = append(i.Models, i.Models[0]) },
		func(i *Inventory) { i.Manufacturers[0].CV8 = 256 },
		func(i *Inventory) { i.Families[0].Sources[0].URL = "javascript:alert(1)" },
		func(i *Inventory) { i.ReviewedAt = "yesterday" },
		func(i *Inventory) { i.Models[0].Variants[0].Interface = "Undocumented connector" },
		func(i *Inventory) { i.Models[0].Variants[0].ArticleNumber = "Not an order code" },
		func(i *Inventory) { i.Models[0].Variants[0].Source.URL = "javascript:alert(1)" },
		func(i *Inventory) { i.Models[0].Variants = append(i.Models[0].Variants, i.Models[0].Variants[0]) },
		func(i *Inventory) { i.Models[0].DocumentStatus = "definitely-in-stock" },
		func(i *Inventory) { i.Models[0].BrandOnly = true; i.Models[0].ProfileID = "esu-lokpilot-5-micro-dcc" },
		func(i *Inventory) { i.Models[0].DocumentationProfileID = "roco-10882" },
		func(i *Inventory) {
			i.Models[0].Identification = &IdentityRule{CV: 8, Values: []int{151}, Scope: "model", Name: "Unsafe manufacturer-only match"}
		},
		func(i *Inventory) {
			i.Models[0].Identification = &IdentityRule{CV: 250, Values: []int{256}, Scope: "model", Name: "Outside byte range"}
		},
		func(i *Inventory) {
			i.Models[0].Identification = &IdentityRule{CV: 250, Values: []int{127, 127}, Scope: "model", Name: "Duplicate values"}
		},
		func(i *Inventory) {
			i.Models[0].Identification = &IdentityRule{CV: 250, Values: []int{127}, Scope: "interface", Name: "Unsupported scope"}
		},
	} {
		inventory := BundledInventory()
		change(&inventory)
		body, _ := json.Marshal(inventory)
		var metadata map[string]any
		_ = json.Unmarshal(body, &metadata)
		if _, err := ParseInventory(metadata); err == nil {
			t.Fatal("invalid inventory accepted")
		}
	}
}

func TestManufacturerRulesAreNotTransferredAcrossModels(t *testing.T) {
	for _, test := range []struct {
		label        string
		manufacturer int
		cv           int
		name         string
	}{
		{"DH18A", 97, 261, "Decoderkennzeichen"},
		{"MN180N18", 145, 250, "Decodertyp"},
		{"STANDARD+ V2", 99, 51, "Brems-Konfiguration"},
	} {
		rows := []Reading{{CVNumber: 8, Value: test.manufacturer, DecoderProfile: test.label}}
		match := Resolve(Bundled().Profiles, test.label, rows)
		definition := FindDefinition(Bundled().Profiles, match, Reading{CVNumber: test.cv})
		if match.Status != "matched" || definition == nil || definition.Name.DE != test.name {
			t.Fatalf("missing reviewed rules: %+v", match)
		}
	}
	// Catalogued products without reviewed rules cannot borrow another model's interpretation.
	rows := []Reading{{CVNumber: 8, Value: 97, DecoderProfile: "PD06A"}}
	match := Resolve(Bundled().Profiles, "PD06A", rows)
	if match.Status != "manufacturer" || FindDefinition(Bundled().Profiles, match, Reading{CVNumber: 27}) != nil {
		t.Fatal("ABC rules guessed for a PD06A")
	}
}

func TestExpandedInventoryPreservesReleasedModelsAndInterpretation(t *testing.T) {
	body, err := os.ReadFile("testdata/inventory-2026-10-20.1.json")
	if err != nil {
		t.Fatal(err)
	}
	var old Inventory
	if err := json.Unmarshal(body, &old); err != nil {
		t.Fatal(err)
	}
	if !PreviousBundled(InventoryMasterDataType, InventoryKey, string(body)) {
		t.Fatal("released eight-manufacturer inventory cannot be upgraded")
	}
	current := BundledInventory()
	for _, previous := range old.Models {
		index := slices.IndexFunc(current.Models, func(m Model) bool { return m.ID == previous.ID })
		if index < 0 {
			t.Fatalf("released model removed: %s", previous.ID)
		}
		model := retainLaterDHModels(current.Models[index], previous)
		if (previous.ProfileID != "" && model.ProfileID != previous.ProfileID) ||
			!reflect.DeepEqual(model.Identification, previous.Identification) {
			t.Fatalf("released model or reviewed interpretation changed: %s", previous.ID)
		}
	}
	metadata, _ := json.Marshal(current)
	var values map[string]any
	_ = json.Unmarshal(metadata, &values)
	if _, err := ParseInventory(values); err != nil {
		t.Fatalf("expanded inventory exceeds the import/backup contract: %v", err)
	}
}

func TestArchivedConnectionsDoNotGuessUndocumentedOrderPairs(t *testing.T) {
	inventory := BundledInventory()
	for _, test := range []struct{ name, article, connection string }{
		{"LokPilot V4.0 M4 MKL", "64618", "21MTC MKL"},
		{"LokPilot V4.0 DCC", "54617", "PluX22"},
		{"MS501F", "MS501F", "NEM 651 (Kabel / wires)"},
		{"MX695KN", "MX695KN", "Schraubklemmen / screw terminals"},
		{"DH21A", "DH21A-5", "21-polig (NEM 660/RCN-121)"},
		{"DHP250", "DHP252", "NEM 652 (Kabel / wires)"},
		{"LE1025E", "10133", "JST 9-polig / JST 9-pin"},
		{"IntelliSound 4 33230", "33230", "Next18S"},
		{"SmartDecoder XP G", "36500", "PIKO G Trägerplatine (Steckverbindung) / carrier board (plug-in)"},
	} {
		index := slices.IndexFunc(inventory.Models, func(m Model) bool { return m.Name == test.name })
		if index < 0 || !slices.ContainsFunc(inventory.Models[index].Variants, func(v Variant) bool {
			return v.ArticleNumber == test.article && v.Interface == test.connection && validSource(v.Source)
		}) {
			t.Fatalf("missing manufacturer-confirmed connection: %+v", test)
		}
	}
	for _, model := range inventory.Models {
		if model.Name == "F060" && (len(model.ArticleNumbers) != 0 || len(model.Interfaces) != 1) {
			t.Fatal("F060 manual has no order number; wire connection must remain documented")
		}
		if model.Name == "LokPilot V2.0" && (len(model.ArticleNumbers) != 3 || len(model.Variants) != 0) {
			t.Fatal("unconfirmed V2.0 article-to-connector pairs were inferred")
		}
		if model.Name == "MS501R" {
			t.Fatal("MS501R is not in the cited MS501/-F/-N table")
		}
		if slices.Contains(model.ArticleNumbers, "36521") && len(model.Interfaces) != 0 {
			t.Fatal("G sound article inherited another decoder's connection without evidence")
		}
	}
}
