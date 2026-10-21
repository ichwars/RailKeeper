package decoderlibrary

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestDH14BFactsKeepSharedPinsAndLogicalOutputsSeparate(t *testing.T) {
	profiles, inventory := Bundled().Profiles, BundledInventory()
	match := Resolve(profiles, "", []Reading{{CVNumber: 8, Value: 97}, {CVNumber: 261, Value: 141}}, &inventory)
	if match.ProfileID != "dh-dh14b" || match.Status != "matched" || match.MatchedBy != "cv" {
		t.Fatal(match)
	}
	var profile Profile
	for _, p := range profiles {
		if p.ID == match.ProfileID {
			profile = p
		}
	}
	if len(profile.Definitions) != 120 || profile.FirmwareMin != "3.15" || !profile.Rules.ModelDefinitionsOnly || profile.Rules.MotorCurveCV != 48 {
		t.Fatal("DH14B scope changed")
	}
	body, _ := json.Marshal(profile)
	var data map[string]any
	_ = json.Unmarshal(body, &data)
	parsed, err := Parse(data)
	if err != nil || !reflect.DeepEqual(parsed, profile) {
		t.Fatalf("metadata round-trip/limit: %v", err)
	}
	for _, cv := range []int{33, 35, 47, 64, 119, 125, 127, 129, 149, 151, 153, 156, 461, 463, 465, 467} {
		d := FindDefinition(profiles, match, Reading{CVNumber: cv})
		if d == nil || d.Reference == nil || d.Reference.SourceIndex != 0 || !strings.Contains(d.Help.DE, "DH14B") || !strings.Contains(d.Help.DE, "dieselben Pins") || !strings.Contains(d.Help.DE, "keine zusätzlichen Anschlüsse") {
			t.Fatalf("logical register became extra hardware at CV%d", cv)
		}
	}
	d := FindDefinition(profiles, match, Reading{CVNumber: 137})
	if d.FactoryDefault == nil || *d.FactoryDefault != 0 || d.Reference.Location != "S.19/75/105, CV137" || !strings.Contains(d.Help.DE, "Separate AUX3/4-Anschlüsse sind nicht dokumentiert") {
		t.Fatal("other hardware pin use copied")
	}
	for _, text := range []string{"Bits0/4/5 gemeinsam", "49 alternativ AUX5/6", "33 deaktiviert SUSI", "Weitere Bits bleiben getrennt"} {
		if !strings.Contains(d.Help.DE, text) {
			t.Fatal("combined CV137 selection missing: " + text)
		}
	}
	if d.Bits[0].Off != d.Bits[0].On || !strings.Contains(d.Bits[0].Off.DE, "gemeinsam") {
		t.Fatal("SUSI state guessed from bit0 alone")
	}
	for _, cv := range []int{35, 153, 156} {
		for _, bit := range FindDefinition(profiles, match, Reading{CVNumber: cv}).Bits {
			if (bit.Bit == 4 || bit.Bit == 5) && !strings.Contains(bit.Name.DE, "ZCLK/ZDAT") ||
				(bit.Bit == 6 || bit.Bit == 7) && !strings.Contains(bit.Name.DE, "alternativ") {
				t.Fatalf("separate AUX3–6 connections borrowed in CV%d", cv)
			}
		}
	}
	if d := FindDefinition(profiles, match, Reading{CVNumber: 261}); !strings.Contains(d.Help.DE, "141") || d.Access != "read-only" {
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

func TestDH14BInventoryUpgradePreservesHistoricalVariantAndOtherModels(t *testing.T) {
	raw, err := os.ReadFile("testdata/inventory-2026-10-20.6.json")
	if err != nil || !PreviousBundled(InventoryMasterDataType, InventoryKey, string(raw)) {
		t.Fatalf("old inventory not recognized: %v", err)
	}
	var previous Inventory
	_ = json.Unmarshal(raw, &previous)
	current := BundledInventory()
	for i, m := range current.Models {
		old := previous.Models[i]
		if m.ID == "dh-dh-dh14b" {
			if m.ProfileID != "dh-dh14b" || m.Notes == nil || m.Identification == nil || !reflect.DeepEqual(m.Identification.Values, []int{141}) || m.DocumentStatus != "historical" {
				t.Fatal("historical model evidence missing")
			}
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
