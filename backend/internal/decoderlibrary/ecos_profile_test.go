package decoderlibrary

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

func TestSharedECoSProfileCases(t *testing.T) {
	data, err := os.ReadFile("ecos_profile_cases_test.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Name, Label, ModelLabel, ECoSProfile, Duplicate string
		Rows                                            []Reading
		Deactivate                                      []string
		InventoryMissing                                bool
		Probe                                           Reading
		Definition                                      bool
		Expected                                        struct {
			Status, ProfileID, MatchedBy string
			SourceStatus                 *string
			ManufacturerID               json.RawMessage
		}
	}
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, test := range cases {
		t.Run(test.Name, func(t *testing.T) {
			profiles, inventory := Bundled().Profiles, BundledInventory()
			for index := range profiles {
				profiles[index].Active = !slices.Contains(test.Deactivate, profiles[index].ID)
				if profiles[index].ID == test.Duplicate {
					copy := profiles[index]
					copy.ID += "-ambiguous"
					profiles = append(profiles, copy)
				}
			}
			for index := range test.Rows {
				test.Rows[index].DecoderProfile = test.Label
			}
			before, _ := json.Marshal(test.Rows)
			sourceInventory := &inventory
			if test.InventoryMissing {
				sourceInventory = nil
			}
			match := ResolveWithContext(profiles, test.Label, test.Rows, sourceInventory,
				ResolutionContext{ModelLabel: test.ModelLabel, ECoSProfile: test.ECoSProfile})
			if match.Status != test.Expected.Status || match.ProfileID != test.Expected.ProfileID || match.MatchedBy != test.Expected.MatchedBy {
				t.Fatalf("unexpected resolution: %+v", match)
			}
			if test.Expected.SourceStatus == nil {
				if match.ECoSProfile != nil {
					t.Fatal("ECoS source escaped its CV group")
				}
			} else if match.ECoSProfile == nil || match.ECoSProfile.Status != *test.Expected.SourceStatus {
				t.Fatalf("incorrect source context: %+v", match.ECoSProfile)
			}
			if string(test.Expected.ManufacturerID) != `"conflicting"` {
				actual, _ := json.Marshal(match.ManufacturerID)
				if string(actual) != string(test.Expected.ManufacturerID) {
					t.Fatalf("manufacturer reading invented or replaced: %s", actual)
				}
			}
			test.Probe.DecoderProfile = test.Label
			if (FindDefinition(profiles, match, test.Probe) != nil) != test.Definition {
				t.Fatalf("incorrect register association: %+v", match)
			}
			after, _ := json.Marshal(test.Rows)
			if string(before) != string(after) {
				t.Fatal("profile matching changed raw CV values")
			}
		})
	}
}
