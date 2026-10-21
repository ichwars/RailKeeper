package decoderlibrary

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestDH21AFactsPreserveVariantAndSourceConflicts(t *testing.T) {
	profiles, inventory := Bundled().Profiles, BundledInventory()
	match := Resolve(profiles, "", []Reading{{CVNumber: 8, Value: 97}, {CVNumber: 261, Value: 200}}, &inventory)
	if match.ProfileID != "dh-dh21a" || match.Status != "matched" || match.MatchedBy != "cv" {
		t.Fatal(match)
	}
	var profile Profile
	for _, p := range profiles {
		if p.ID == match.ProfileID {
			profile = p
		}
	}
	if len(profile.Definitions) != 121 || profile.FirmwareMin != "3.15" || !profile.Rules.ModelDefinitionsOnly || profile.Rules.MotorCurveCV != 48 || len(profile.Sources) != 3 {
		t.Fatal("DH21A scope changed")
	}
	body, _ := json.Marshal(profile)
	var data map[string]any
	_ = json.Unmarshal(body, &data)
	parsed, err := Parse(data)
	if err != nil || !reflect.DeepEqual(parsed, profile) {
		t.Fatalf("metadata round-trip/limit: %v", err)
	}
	for _, source := range profile.Sources {
		if source.RetrievedAt != "2026-10-21" || len(source.SHA256) != 64 {
			t.Fatal("fresh source evidence missing")
		}
	}
	for _, cv := range []int{33, 35, 47, 64, 119, 125, 127, 129, 149, 151, 153, 156, 461, 463, 465, 467} {
		d := FindDefinition(profiles, match, Reading{CVNumber: cv})
		if d == nil || d.Reference == nil || !strings.Contains(d.Help.DE, "DH21A") || !strings.Contains(d.Help.DE, "widersprüchlich") {
			t.Fatalf("output conflict hidden at CV%d", cv)
		}
	}
	d := FindDefinition(profiles, match, Reading{CVNumber: 137})
	if d.FactoryDefault == nil || *d.FactoryDefault != 0 || d.Reference.Location != "S.24–27/75/105, CV137" {
		t.Fatal("DH21A-specific reference missing")
	}
	for _, required := range []string{"-4 in beiden Generationen verstärkt", "-5 dagegen Logikpegel", "S.26", "S.27", "Pegel bleibt ungeklärt", "CV261=200 bestimmt keine Generation"} {
		if !strings.Contains(d.Help.DE, required) {
			t.Fatal("variant or conflict evidence missing: " + required)
		}
	}
	if d.Bits[0].Off != d.Bits[0].On {
		t.Fatal("SUSI state guessed from bit0 alone")
	}
	for _, cv := range []int{35, 153, 156} {
		for _, bit := range FindDefinition(profiles, match, Reading{CVNumber: cv}).Bits {
			if (bit.Bit == 4 || bit.Bit == 5) && !strings.Contains(bit.Name.DE, "varianteabhängig") {
				t.Fatalf("AUX3/4 level guessed in CV%d", cv)
			}
		}
	}
	partial := FindDefinition(profiles, match, Reading{CVNumber: 144})
	if partial == nil || len(partial.Bits) != 5 || partial.Maximum != 63 || !strings.Contains(partial.Help.DE, "Bit0 bleibt ungeklärt") || !strings.Contains(partial.Help.DE, "nur für Generation2") {
		t.Fatal("partial CV144 evidence lost")
	}
	for i, bit := range partial.Bits {
		if bit.Bit != i+1 || (bit.Bit == 3 || bit.Bit == 4) && !strings.Contains(bit.On.DE, "Hardware prüfen") {
			t.Fatal("unknown bit0 or GPIO hardware inferred")
		}
	}
	if d := FindDefinition(profiles, match, Reading{CVNumber: 12}); !strings.Contains(d.Help.DE, "AC-Analogfähigkeit") || !strings.Contains(d.Help.DE, "weder den aktuellen Betrieb") {
		t.Fatal("capability became active analog operation")
	}
	for _, cv := range []int{6, 67, 94, 98, 99, 101, 102, 103, 131, 132, 133, 158, 159} {
		if FindDefinition(profiles, match, Reading{CVNumber: cv}) != nil {
			t.Fatalf("DH21B/foundation register borrowed CV%d", cv)
		}
	}
	for _, cv := range []int{144, 261, 401, 461} {
		if FindDefinition(profiles, match, Reading{CVNumber: cv, IndexCV31: pointer(1), IndexCV32: pointer(1)}) != nil {
			t.Fatal("unindexed register became indexed")
		}
	}
}

func TestDH21AInventoryUpgradePreservesBothGenerationsAndOtherModels(t *testing.T) {
	raw, err := os.ReadFile("testdata/inventory-2026-10-20.7.json")
	if err != nil || !PreviousBundled(InventoryMasterDataType, InventoryKey, string(raw)) {
		t.Fatalf("old inventory not recognized: %v", err)
	}
	var previous Inventory
	_ = json.Unmarshal(raw, &previous)
	current := BundledInventory()
	for i, m := range current.Models {
		old := previous.Models[i]
		if m.ID == "dh-dh-dh21a" {
			if m.ProfileID != "dh-dh21a" || m.Notes == nil || m.Identification == nil || !reflect.DeepEqual(m.Identification.Values, []int{200}) || m.DocumentStatus != "historical" || len(m.Variants) != 5 {
				t.Fatal("historical model evidence missing")
			}
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
