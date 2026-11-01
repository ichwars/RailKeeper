package decoderlibrary

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

// Historical fixture checks normalize only the later, separately reviewed model changes.
func retainLaterDHModels(current, previous Model) Model {
	switch current.ID {
	case "dh-dh-dh21b", "dh-dh-dh22a", "dh-dh-dh22b", "dh-dh-dh24a", "dh-dh-dhsp10a":
		current.ProfileID, current.Notes, current.Identification = previous.ProfileID, previous.Notes, previous.Identification
	case "esu-lokpilot-5-lokpilot-5-nano-dcc", "esu-lokpilot-legacy-lokpilot-micro-v4-0",
		"esu-lokpilot-legacy-lokpilot-micro-v4-0-dcc", "esu-lokpilot-legacy-lokpilot-standard":
		current.ProfileID = previous.ProfileID
	}
	return current
}

func TestDHExtendedModelsKeepHardwareAndCVBoundaries(t *testing.T) {
	profiles, inventory := Bundled().Profiles, BundledInventory()
	for _, tc := range []struct {
		model       string
		code, count int
		extended    bool
	}{{"DH21B", 201, 126, true}, {"DH22A", 202, 120, false}, {"DH22B", 203, 126, true},
		{"DH24A", 204, 127, true}, {"DHSP10A", 110, 120, false}} {
		t.Run(tc.model, func(t *testing.T) {
			match := Resolve(profiles, "", []Reading{{CVNumber: 8, Value: 97}, {CVNumber: 261, Value: tc.code}}, &inventory)
			if match.ProfileID != "dh-"+strings.ToLower(tc.model) || match.Status != "matched" {
				t.Fatal(match)
			}
			var profile Profile
			for _, p := range profiles {
				if p.ID == match.ProfileID {
					profile = p
				}
			}
			if len(profile.Definitions) != tc.count || profile.FirmwareMin != "3.15" || !profile.Rules.ModelDefinitionsOnly {
				t.Fatal("incorrect model scope")
			}
			body, _ := json.Marshal(profile)
			var data map[string]any
			_ = json.Unmarshal(body, &data)
			parsed, err := Parse(data)
			if err != nil || !reflect.DeepEqual(profile, parsed) {
				t.Fatalf("metadata size or round trip: %v", err)
			}
			for _, d := range profile.Definitions {
				if strings.Contains(d.Help.DE, "DH16A") || d.Reference == nil || d.Reference.SourceIndex >= len(profile.Sources) {
					t.Fatalf("borrowed model or missing source CV%d", d.CV)
				}
			}
			susi := FindDefinition(profiles, match, Reading{CVNumber: 137})
			if !strings.Contains(susi.Help.DE, tc.model) || susi.Bits[0].On != susi.Bits[0].Off {
				t.Fatal("SUSI state guessed")
			}
			if tc.extended {
				if !strings.Contains(susi.Help.DE, "AUX11/12") {
					t.Fatal("SUSI shared pin mismatch")
				}
				for _, bit := range susi.Bits {
					if bit.Bit == 5 {
						t.Fatal("irrelevant bit5 interpreted")
					}
				}
				for _, cv := range []int{98, 99, 102, 103, 158, 159} {
					if FindDefinition(profiles, match, Reading{CVNumber: cv}) == nil {
						t.Fatalf("missing CV%d", cv)
					}
				}
				if FindDefinition(profiles, match, Reading{CVNumber: 159}).Reference.Location != "S.79, CV159; CV145 S.77" {
					t.Fatal("wrong source page")
				}
			} else {
				for _, cv := range []int{98, 99, 102, 103, 158, 159} {
					if FindDefinition(profiles, match, Reading{CVNumber: cv}) != nil {
						t.Fatal("B hardware borrowed")
					}
				}
			}
			if (FindDefinition(profiles, match, Reading{CVNumber: 101}) != nil) != (tc.model == "DH24A") {
				t.Fatal("AUX10 invented")
			}
			for _, cv := range []int{6, 67, 94, 131, 144} {
				if FindDefinition(profiles, match, Reading{CVNumber: cv}) != nil {
					t.Fatalf("unverified CV%d borrowed", cv)
				}
			}
			for _, cv := range []int{137, 261, 401, 465} {
				if FindDefinition(profiles, match, Reading{CVNumber: cv, IndexCV31: pointer(1), IndexCV32: pointer(1)}) != nil {
					t.Fatal("bank contamination")
				}
			}
		})
	}
}

func TestDHExtendedUpgradeChangesOnlyFiveModels(t *testing.T) {
	raw, err := os.ReadFile("testdata/inventory-2026-10-21.1.json")
	if err != nil || !PreviousBundled(InventoryMasterDataType, InventoryKey, string(raw)) {
		t.Fatalf("previous inventory not recognized: %v", err)
	}
	var old Inventory
	_ = json.Unmarshal(raw, &old)
	current := BundledInventory()
	changed := 0
	for i, model := range current.Models {
		if strings.HasPrefix(model.ID, "dh-") && !reflect.DeepEqual(model, old.Models[i]) {
			changed++
		}
		if !reflect.DeepEqual(retainLaterDHModels(model, old.Models[i]), old.Models[i]) {
			t.Fatalf("unrelated data changed: %s", model.ID)
		}
	}
	if changed != 5 || !reflect.DeepEqual(current.Families, old.Families) || !reflect.DeepEqual(current.Manufacturers, old.Manufacturers) {
		t.Fatal("unexpected catalogue changes")
	}
}
