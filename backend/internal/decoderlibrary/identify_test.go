package decoderlibrary

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

func TestSharedIdentityCases(t *testing.T) {
	body, err := os.ReadFile("identity_cases_test.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Name       string    `json:"name"`
		Label      string    `json:"label"`
		Rows       []Reading `json:"rows"`
		Deactivate []string  `json:"deactivate"`
		Expected   Match     `json:"expected"`
	}
	if err := json.Unmarshal(body, &cases); err != nil {
		t.Fatal(err)
	}
	for _, test := range cases {
		t.Run(test.Name, func(t *testing.T) {
			profiles := Bundled().Profiles
			inventory := BundledInventory()
			for i := range profiles {
				if slices.Contains(test.Deactivate, profiles[i].ID) {
					profiles[i].Active = false
				}
			}
			for i := range test.Rows {
				if test.Rows[i].DecoderProfile == "" {
					test.Rows[i].DecoderProfile = test.Label
				}
			}
			before, _ := json.Marshal(test.Rows)
			got := Resolve(profiles, test.Label, test.Rows, &inventory)
			if got.Status != test.Expected.Status || got.ProfileID != test.Expected.ProfileID || got.MatchedBy != test.Expected.MatchedBy {
				t.Fatalf("resolution mismatch: %+v", got)
			}
			if test.Expected.Identification == nil {
				if got.Identification != nil {
					t.Fatalf("unexpected identity: %+v", got)
				}
			} else if got.Identification == nil || got.Identification.Name != test.Expected.Identification.Name || got.Identification.Scope != test.Expected.Identification.Scope {
				t.Fatalf("identity mismatch: %+v", got)
			}
			if got.ProfileID == "dcc-basics" && FindDefinition(profiles, got, Reading{CVNumber: 27}) != nil {
				t.Fatal("manufacturer rules guessed")
			}
			after, _ := json.Marshal(test.Rows)
			if string(before) != string(after) {
				t.Fatal("identity resolution mutated readings")
			}
		})
	}
}

func TestPreviousBundledDigestRequiresPristineMetadata(t *testing.T) {
	body, err := os.ReadFile("testdata/dh18a-2026-10-18.2.json")
	if err != nil {
		t.Fatal(err)
	}
	if !PreviousBundled(MasterDataType, "dh-dh18a", string(body)) {
		t.Fatal("known previous revision not recognized")
	}
	var document map[string]any
	_ = json.Unmarshal(body, &document)
	document["name"] = "Locally edited"
	changed, _ := json.Marshal(document)
	if PreviousBundled(MasterDataType, "dh-dh18a", string(changed)) {
		t.Fatal("local edit mistaken for pristine bundle")
	}
}
