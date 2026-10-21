package decoderlibrary

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestDH16AHasOwnLocatedFactsAndBothGenerationHardware(t *testing.T) {
	profiles, inventory := Bundled().Profiles, BundledInventory()
	rows := []Reading{{CVNumber: 8, Value: 97}, {CVNumber: 261, Value: 160}}
	match := Resolve(profiles, "", rows, &inventory)
	if match.ProfileID != "dh-dh16a" || match.Status != "matched" || match.MatchedBy != "cv" {
		t.Fatal(match)
	}
	var profile Profile
	for _, p := range profiles {
		if p.ID == match.ProfileID {
			profile = p
		}
	}
	if len(profile.Definitions) != 120 || profile.FirmwareMin != "3.15" || !profile.Rules.ModelDefinitionsOnly || profile.Rules.MotorCurveCV != 48 {
		t.Fatal("DH16A scope changed")
	}
	body, _ := json.Marshal(profile)
	var data map[string]any
	_ = json.Unmarshal(body, &data)
	parsed, err := Parse(data)
	if err != nil || !reflect.DeepEqual(parsed, profile) {
		t.Fatalf("metadata round-trip/limit: %v", err)
	}
	for _, cv := range []int{33, 35, 47, 64, 137, 153, 156, 461, 463, 465, 467} {
		d := FindDefinition(profiles, match, Reading{CVNumber: cv})
		if d == nil || d.Reference == nil || d.Reference.SourceIndex != 0 || !strings.Contains(d.Help.DE, "DH16A") || !strings.Contains(d.Help.DE, "SUSI") && !strings.Contains(d.Help.DE, "CV137") {
			t.Fatalf("unqualified hardware fact CV%d", cv)
		}
	}
	d := FindDefinition(profiles, match, Reading{CVNumber: 137})
	if d.FactoryDefault == nil || *d.FactoryDefault != 0 || d.Reference.Location != "S.20–21/75/105, CV137" || !strings.Contains(d.Help.DE, "in beiden Generationen verstärkt") {
		t.Fatal("DH10C/DH18A hardware copied")
	}
	if d := FindDefinition(profiles, match, Reading{CVNumber: 261}); !strings.Contains(d.Help.DE, "160") || !strings.Contains(d.Help.DE, "keine Bestimmung der Hardwaregeneration") {
		t.Fatal("type code became generation detector")
	}
	for _, cv := range []int{127, 128, 129, 130, 151, 152} {
		d := FindDefinition(profiles, match, Reading{CVNumber: cv})
		if strings.Contains(d.Help.DE, "Generation2") || !strings.Contains(d.Help.DE, "mit SUSI") {
			t.Fatal("AUX5/6 falsely restricted to generation2")
		}
	}
	for _, cv := range []int{6, 67, 94, 98, 99, 101, 102, 103, 131, 132, 133, 144, 158, 159} {
		if FindDefinition(profiles, match, Reading{CVNumber: cv}) != nil {
			t.Fatalf("other hardware/foundation borrowed CV%d", cv)
		}
	}
	for _, cv := range []int{261, 401, 461} {
		if FindDefinition(profiles, match, Reading{CVNumber: cv, IndexCV31: pointer(1), IndexCV32: pointer(1)}) != nil {
			t.Fatal("unindexed register became indexed")
		}
	}
}

func TestDH16AInventoryUpgradeChangesOnlyItsProfileLinkAndNotes(t *testing.T) {
	raw, err := os.ReadFile("testdata/inventory-2026-10-20.4.json")
	if err != nil || !PreviousBundled(InventoryMasterDataType, InventoryKey, string(raw)) {
		t.Fatalf("old inventory not recognized: %v", err)
	}
	var previous Inventory
	_ = json.Unmarshal(raw, &previous)
	current := BundledInventory()
	for i, m := range current.Models {
		old := previous.Models[i]
		if m.ID == "dh-dh-dh16a" {
			if m.ProfileID != "dh-dh16a" || m.Notes == nil {
				t.Fatal("new documentation link missing")
			}
			m.ProfileID = old.ProfileID
			m.Notes = old.Notes
		}
		if m.ID == "dh-dh-dh12a" {
			m.ProfileID, m.Notes, m.Identification = old.ProfileID, old.Notes, old.Identification
		}
		if m.ID == "dh-dh-dh14b" {
			m.ProfileID, m.Notes, m.Identification = old.ProfileID, old.Notes, old.Identification
		}
		if m.ID == "dh-dh-dh21a" {
			m.ProfileID, m.Notes, m.Identification = old.ProfileID, old.Notes, old.Identification
		}
		m = retainLaterDHModels(m, old)
		if !reflect.DeepEqual(m, old) {
			t.Fatalf("unrelated model changed: %s", m.ID)
		}
	}
	if !reflect.DeepEqual(previous.Families, current.Families) || !reflect.DeepEqual(previous.Manufacturers, current.Manufacturers) {
		t.Fatal("directory identities changed")
	}
	var edited map[string]any
	_ = json.Unmarshal(raw, &edited)
	edited["revision"] = "local"
	changed, _ := json.Marshal(edited)
	if PreviousBundled(InventoryMasterDataType, InventoryKey, string(changed)) {
		t.Fatal("local edit trusted")
	}
}
