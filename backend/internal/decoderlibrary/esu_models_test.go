package decoderlibrary

import (
	"encoding/json"
	"testing"
)

func TestESUInventoryUpgradeRecognizesOnlyReleasedUneditedMetadata(t *testing.T) {
	previous := BundledInventory()
	previous.Revision, previous.ReviewedAt = "2026-10-21.2", "2026-10-21"
	changed := 0
	for i := range previous.Models {
		switch previous.Models[i].ID {
		case "esu-lokpilot-5-lokpilot-5-nano-dcc", "esu-lokpilot-legacy-lokpilot-micro-v4-0",
			"esu-lokpilot-legacy-lokpilot-micro-v4-0-dcc", "esu-lokpilot-legacy-lokpilot-standard":
			if previous.Models[i].ProfileID == "" {
				t.Fatal("reviewed ESU model has no profile")
			}
			previous.Models[i].ProfileID = ""
			changed++
		}
	}
	metadata, _ := json.Marshal(previous)
	if changed != 4 || !PreviousBundled(InventoryMasterDataType, InventoryKey, string(metadata)) {
		t.Fatal("previous released ESU inventory metadata is not recognized")
	}
	previous.Models[0].Name = "Meine lokale Änderung"
	metadata, _ = json.Marshal(previous)
	if PreviousBundled(InventoryMasterDataType, InventoryKey, string(metadata)) {
		t.Fatal("local inventory edits are upgradeable")
	}
}

func TestESULayoutsKeepVariantAndPageBoundaries(t *testing.T) {
	profiles, inventory := Bundled().Profiles, BundledInventory()
	for _, tc := range []struct {
		name     string
		cv, page int
		known    bool
	}{
		{"LokPilot 5 nano DCC", 257, 3, true},
		{"LokPilot 5 nano DCC", 369, 7, true},
		{"LokPilot 5 nano DCC", 385, 7, false},
		{"LokPilot 5 nano DCC", 257, 2, false},
		{"LokPilot V4.0 micro DCC", 266, 2, true},
		{"LokPilot V4.0 micro DCC", 266, 8, false},
	} {
		match := ResolveWithContext(profiles, "ECoS", nil, &inventory,
			ResolutionContext{ECoSProfile: tc.name})
		if match.Status != "matched" || match.ManufacturerID != nil {
			t.Fatal("source model resolution invented identity", match)
		}
		index := 16
		known := FindDefinition(profiles, match, Reading{CVNumber: tc.cv, Value: 0, IndexCV31: &index, IndexCV32: &tc.page}) != nil
		if known != tc.known {
			t.Fatalf("incorrect indexed definition boundary: %+v", tc)
		}
	}
}
