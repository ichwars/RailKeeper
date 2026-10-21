package decoderlibrary

import (
	"encoding/json"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestDHCompactProfilesKeepExactModelAndGenerationBoundaries(t *testing.T) {
	profiles, inventory := Bundled().Profiles, BundledInventory()
	for _, test := range []struct {
		model string
		code  int
	}{{"DH05C", 52}, {"DH10C", 102}, {"DH18A", 180}} {
		t.Run(test.model, func(t *testing.T) {
			label := ""
			rows := []Reading{{CVNumber: 8, Value: 97}, {CVNumber: 261, Value: test.code}}
			match := Resolve(profiles, label, rows, &inventory)
			if match.Status != "matched" || match.ProfileID != "dh-"+strings.ToLower(test.model) || match.MatchedBy != "cv" {
				t.Fatal(match)
			}
			var profile Profile
			for _, p := range profiles {
				if p.ID == match.ProfileID {
					profile = p
				}
			}
			if len(profile.Definitions) != 120 || !profile.Rules.ModelDefinitionsOnly || profile.FirmwareMin != "3.15" {
				t.Fatal("scope lost")
			}
			body, _ := json.Marshal(profile)
			var data map[string]any
			_ = json.Unmarshal(body, &data)
			parsed, err := Parse(data)
			if err != nil || !reflect.DeepEqual(profile, parsed) {
				t.Fatalf("lossy/oversized metadata: %v", err)
			}
			for _, cv := range []int{17, 18, 19, 21, 22} {
				d := FindDefinition(profiles, match, Reading{CVNumber: cv})
				if d == nil || d.Reference.Location != "S.67, CV"+strconv.Itoa(cv) {
					t.Fatalf("unlocated address register CV%d", cv)
				}
			}
			if FindDefinition(profiles, match, Reading{CVNumber: 27}).Reference.Location != "S.68, CV27" {
				t.Fatal("wrong brake source page")
			}
			for _, cv := range []int{6, 67, 94, 98, 99, 101, 102, 103, 131, 132, 133, 144, 158, 159} {
				if FindDefinition(profiles, match, Reading{CVNumber: cv}) != nil {
					t.Fatalf("other hardware or DCC fallback borrowed: %d", cv)
				}
			}
			for _, cv := range []int{261, 401, 461, 467} {
				if FindDefinition(profiles, match, Reading{CVNumber: cv, IndexCV31: pointer(1), IndexCV32: pointer(1)}) != nil {
					t.Fatal("unindexed register became banked")
				}
			}
			susi := FindDefinition(profiles, match, Reading{CVNumber: 137})
			expected := 0
			if test.model == "DH18A" {
				expected = 1
			}
			if susi.FactoryDefault == nil || *susi.FactoryDefault != expected {
				t.Fatal("DH18A-specific default copied")
			}
			if test.model != "DH18A" {
				if !strings.Contains(susi.Help.DE, "Generation1") || !strings.Contains(susi.Help.DE, "Generation2") {
					t.Fatal("hardware uncertainty lost")
				}
				for _, cv := range []int{127, 129, 151, 465} {
					if !strings.Contains(strings.ToLower(FindDefinition(profiles, match, Reading{CVNumber: cv}).Help.DE), "generation") {
						t.Fatal("generation-dependent output claimed unconditionally")
					}
				}
			}
		})
	}
}

func TestDHCompactUpgradeTrustsOnlyExactReleasedMetadata(t *testing.T) {
	for _, test := range []struct{ file, kind, key string }{{"inventory-2026-10-20.3.json", InventoryMasterDataType, InventoryKey}, {"dh18a-2026-10-20.2.json", MasterDataType, "dh-dh18a"}} {
		raw, err := os.ReadFile("testdata/" + test.file)
		if err != nil || !PreviousBundled(test.kind, test.key, string(raw)) {
			t.Fatalf("released metadata unrecognized: %v", err)
		}
		var edited map[string]any
		_ = json.Unmarshal(raw, &edited)
		edited["revision"] = "local"
		changed, _ := json.Marshal(edited)
		if PreviousBundled(test.kind, test.key, string(changed)) {
			t.Fatal("operator edit trusted as bundle")
		}
	}
	raw, _ := os.ReadFile("testdata/inventory-2026-10-20.3.json")
	var old Inventory
	_ = json.Unmarshal(raw, &old)
	current := BundledInventory()
	for i, m := range current.Models {
		prior := old.Models[i]
		if m.ID == "dh-dh-dh05c" || m.ID == "dh-dh-dh10c" || m.ID == "dh-dh-dh16a" {
			m.ProfileID = prior.ProfileID
			m.Notes = prior.Notes
			if m.ID == "dh-dh-dh05c" {
				if len(m.Variants) != len(prior.Variants)+1 || m.Variants[len(prior.Variants)].ArticleNumber != "DH05C-1" || !strings.Contains(m.Variants[len(prior.Variants)].Source.Scope, "S.14") {
					t.Fatal("historical connection not sourced")
				}
				if !reflect.DeepEqual(m.Variants[:len(prior.Variants)], prior.Variants) {
					t.Fatal("existing connections changed")
				}
				m.Interfaces = prior.Interfaces
				m.ArticleNumbers = prior.ArticleNumbers
				m.Variants = prior.Variants
			}
		}
		if m.ID == "dh-dh-dh12a" {
			m.ProfileID, m.Notes, m.Identification = prior.ProfileID, prior.Notes, prior.Identification
		}
		if m.ID == "dh-dh-dh14b" {
			m.ProfileID, m.Notes, m.Identification = prior.ProfileID, prior.Notes, prior.Identification
		}
		if m.ID == "dh-dh-dh21a" {
			m.ProfileID, m.Notes, m.Identification = prior.ProfileID, prior.Notes, prior.Identification
		}
		m = retainLaterDHModels(m, prior)
		if !reflect.DeepEqual(m, prior) {
			t.Fatalf("unrelated catalogue model changed: %s", m.ID)
		}
	}
	if !reflect.DeepEqual(old.Families, current.Families) || !reflect.DeepEqual(old.Manufacturers, current.Manufacturers) {
		t.Fatal("families/manufacturers changed")
	}
}
