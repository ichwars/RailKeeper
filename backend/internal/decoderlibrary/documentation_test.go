package decoderlibrary

import (
	"encoding/json"
	"os"
	"reflect"
	"slices"
	"testing"
)

func TestManufacturerTablesPreserveReviewedFactsAndSourceLocations(t *testing.T) {
	profiles := Bundled().Profiles
	for _, test := range []struct {
		id, label                            string
		maker, cv, minimum, maximum, factory int
	}{
		{"esu-lokpilot-5-micro-dcc", "LokPilot 5 micro DCC", 151, 9, 10, 50, 40},
		{"dh-dh18a", "DH18A", 97, 56, 0, 7, 3},
		{"zimo-mn180n18", "MN180N18", 145, 18, 0, 255, 128},
		{"lenz-standard-plus-v2", "STANDARD+ V2", 99, 147, 0, 255, 30},
		{"kuehn-n45", "Kühn N45", 157, 54, 2, 63, 16},
		{"uhlenbrock-76420", "Uhlenbrock 76420", 85, 50, 0, 63, 32},
		{"roco-10882", "Roco 10882", 145, 2, 1, 255, 4},
		{"piko-56500", "PIKO 56500", 162, 141, 0, 255, 30},
	} {
		t.Run(test.id, func(t *testing.T) {
			match := Resolve(profiles, test.label, []Reading{{CVNumber: 8, Value: test.maker, DecoderProfile: test.label}})
			definition := FindDefinition(profiles, match, Reading{CVNumber: test.cv})
			if match.ProfileID != test.id || definition == nil || definition.Minimum != test.minimum ||
				definition.Maximum != test.maximum || definition.FactoryDefault == nil || *definition.FactoryDefault != test.factory {
				t.Fatalf("manufacturer table fact changed: %+v %+v", match, definition)
			}
			profile := profiles[slices.IndexFunc(profiles, func(p Profile) bool { return p.ID == test.id })]
			if definition.Reference == nil || definition.Reference.Location == "" || profile.Sources[definition.Reference.SourceIndex].RetrievedAt != "2026-10-20" {
				t.Fatal("reviewed table row lost its source location")
			}
			body, _ := json.Marshal(profile)
			var metadata map[string]any
			_ = json.Unmarshal(body, &metadata)
			parsed, err := Parse(metadata)
			if err != nil || !reflect.DeepEqual(profile, parsed) {
				t.Fatalf("documented metadata is not lossless in storage/backup: %v", err)
			}
		})
	}
}

func TestContradictoryTablesAndOEMBrandDoNotEnableGuessedRules(t *testing.T) {
	profiles := Bundled().Profiles
	for _, test := range []struct {
		label     string
		maker, cv int
	}{
		{"Uhlenbrock 76420", 85, 27}, {"Uhlenbrock 76420", 85, 28}, {"Uhlenbrock 76420", 85, 95},
		{"PIKO 56500", 162, 101}, {"Roco 10882", 161, 155}, {"Kühn N025", 157, 54},
	} {
		match := Resolve(profiles, test.label, []Reading{{CVNumber: 8, Value: test.maker, DecoderProfile: test.label}})
		if FindDefinition(profiles, match, Reading{CVNumber: test.cv}) != nil {
			t.Fatalf("unverified table or OEM rule used: %+v", test)
		}
	}
	label := "PIKO 56500"
	match := Resolve(profiles, label, []Reading{{CVNumber: 8, Value: 162, DecoderProfile: label}})
	if d := FindDefinition(profiles, match, Reading{CVNumber: 267, IndexCV31: pointer(18), IndexCV32: pointer(0)}); d == nil || d.Reference == nil {
		t.Fatal("documented PIKO effect bank missing")
	}
	if FindDefinition(profiles, match, Reading{CVNumber: 267, IndexCV31: pointer(1), IndexCV32: pointer(1)}) != nil {
		t.Fatal("PIKO effect address confused with RailComPlus bank")
	}
	for _, model := range BundledInventory().Models {
		if model.DocumentationProfileID != "" && (!model.BrandOnly || model.ProfileID != "" || model.Identification != nil) {
			t.Fatal("documentation-only link enabled identity matching")
		}
	}
}

func TestDocumentedMetadataRejectsBrokenReferencesAndDefaults(t *testing.T) {
	for _, change := range []func(*Profile){
		func(p *Profile) { p.Definitions[0].Reference.SourceIndex = len(p.Sources) },
		func(p *Profile) { p.Definitions[0].Reference.Location = "" },
		func(p *Profile) { p.Definitions[0].FactoryDefault = pointer(256) },
		func(p *Profile) { p.Definitions[0].Access = "execute" },
		func(p *Profile) { p.Definitions[0].Unit = &Text{DE: "s"} },
		func(p *Profile) { p.Definitions[0].AllowedValues = []int{1, 1} },
		func(p *Profile) { p.Definitions[0].AllowedValues = []int{0} },
		func(p *Profile) { p.Definitions[0].AllowedValues = []int{1, 2} },
		func(p *Profile) { p.Sources[0].SHA256 = "unverified" },
		func(p *Profile) { p.Sources[0].RetrievedAt = "yesterday" },
	} {
		p := Bundled().Profiles[5]
		change(&p)
		if err := p.Validate(); err == nil {
			t.Fatal("invalid documentation accepted")
		}
	}
}

func TestPreviousShippedCatalogCanUpgradeWithoutTrustingEdits(t *testing.T) {
	body, err := os.ReadFile("testdata/catalog-before-cv-tables.json")
	if err != nil {
		t.Fatal(err)
	}
	var previous Catalog
	if err := json.Unmarshal(body, &previous); err != nil {
		t.Fatal(err)
	}
	for _, p := range previous.Profiles {
		metadata, _ := json.Marshal(p)
		if !PreviousBundled(MasterDataType, p.ID, string(metadata)) {
			t.Fatalf("previous deployed profile cannot upgrade: %s", p.ID)
		}
		p.Definitions[0].Help.DE += " lokal geändert"
		metadata, _ = json.Marshal(p)
		if PreviousBundled(MasterDataType, p.ID, string(metadata)) {
			t.Fatal("local edit trusted")
		}
	}
	body, err = os.ReadFile("testdata/inventory-2026-10-20.2.json")
	if err != nil || !PreviousBundled(InventoryMasterDataType, InventoryKey, string(body)) {
		t.Fatalf("previous deployed inventory cannot upgrade: %v", err)
	}
}
