package decoderlibrary

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestDH12AHasOwnLocatedFactsAndLogicLevelOutputs(t *testing.T) {
	profiles, inventory := Bundled().Profiles, BundledInventory()
	match := Resolve(profiles, "", []Reading{{CVNumber: 8, Value: 97}, {CVNumber: 261, Value: 120}}, &inventory)
	if match.ProfileID != "dh-dh12a" || match.Status != "matched" || match.MatchedBy != "cv" {
		t.Fatal(match)
	}
	var profile Profile
	for _, p := range profiles {
		if p.ID == match.ProfileID {
			profile = p
		}
	}
	if len(profile.Definitions) != 120 || profile.FirmwareMin != "3.15" || !profile.Rules.ModelDefinitionsOnly || profile.Rules.MotorCurveCV != 48 {
		t.Fatal("DH12A scope changed")
	}
	body, _ := json.Marshal(profile)
	var data map[string]any
	_ = json.Unmarshal(body, &data)
	parsed, err := Parse(data)
	if err != nil || !reflect.DeepEqual(parsed, profile) {
		t.Fatalf("metadata round-trip/limit: %v", err)
	}
	if strings.Contains(string(body), "DH16A:") || strings.Contains(string(body), "Generation2") || strings.Contains(string(body), "(amplified)") {
		t.Fatal("other model hardware copied")
	}
	for _, cv := range []int{33, 35, 47, 64, 119, 125, 137, 149, 153, 156, 461, 463, 465, 467} {
		d := FindDefinition(profiles, match, Reading{CVNumber: cv})
		if d == nil || d.Reference == nil || d.Reference.SourceIndex != 0 || !strings.Contains(d.Help.DE, "DH12A") || !strings.Contains(d.Help.DE, "unverstärkt") {
			t.Fatalf("unqualified hardware fact CV%d", cv)
		}
	}
	d := FindDefinition(profiles, match, Reading{CVNumber: 137})
	if d.FactoryDefault == nil || *d.FactoryDefault != 0 || d.Reference.Location != "S.18/75/105, CV137" || !strings.Contains(d.Help.DE, "AUX3/4 haben eigene Anschlüsse") {
		t.Fatal("DH12A-specific pin use missing")
	}
	for _, cv := range []int{35, 153, 156} {
		for _, bit := range FindDefinition(profiles, match, Reading{CVNumber: cv}).Bits {
			if (bit.Bit == 4 || bit.Bit == 5) && !strings.Contains(bit.Name.DE, "Logikpegel") {
				t.Fatalf("amplified AUX3/4 borrowed in CV%d", cv)
			}
		}
	}
	if d := FindDefinition(profiles, match, Reading{CVNumber: 261}); !strings.Contains(d.Help.DE, "120") || d.Access != "read-only" {
		t.Fatal("type code not documented")
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

func TestDH12AInventoryUpgradeKeepsHistoricalVariantAndOtherModels(t *testing.T) {
	raw, err := os.ReadFile("testdata/inventory-2026-10-20.5.json")
	if err != nil || !PreviousBundled(InventoryMasterDataType, InventoryKey, string(raw)) {
		t.Fatalf("old inventory not recognized: %v", err)
	}
	var previous Inventory
	_ = json.Unmarshal(raw, &previous)
	current := BundledInventory()
	for i, m := range current.Models {
		old := previous.Models[i]
		if m.ID == "dh-dh-dh12a" {
			if m.ProfileID != "dh-dh12a" || m.Notes == nil || m.Identification == nil || !reflect.DeepEqual(m.Identification.Values, []int{120}) || m.DocumentStatus != "historical" {
				t.Fatal("historical model evidence missing")
			}
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
			t.Fatalf("unrelated catalogue data changed: %s", m.ID)
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
