package decoderlibrary

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestSharedESUDefinitionCases(t *testing.T) {
	body, err := os.ReadFile("definition_cases_test.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Name     string  `json:"name"`
		Row      Reading `json:"row"`
		Firmware string  `json:"firmware"`
		Meaning  *string `json:"meaning"`
	}
	if err := json.Unmarshal(body, &cases); err != nil {
		t.Fatal(err)
	}
	for _, test := range cases {
		t.Run(test.Name, func(t *testing.T) {
			label := "LokPilot 5 micro DCC"
			rows := []Reading{{CVNumber: 8, Value: 151, DecoderProfile: label}}
			if test.Firmware != "" {
				rows = append(rows, Reading{CVNumber: 7, Value: 255, DecoderProfile: label, Description: "LokProgrammer Decoder: LokPilot 5 micro DCC (" + test.Firmware + ")"})
			}
			profiles := Bundled().Profiles
			match := Resolve(profiles, label, rows)
			definition := FindDefinition(profiles, match, test.Row)
			if test.Meaning == nil {
				if definition != nil {
					t.Fatalf("undocumented address or firmware interpreted: %+v", definition)
				}
			} else if definition == nil || !strings.Contains(definition.Name.DE, *test.Meaning) {
				t.Fatalf("wrong model/page/firmware interpretation: %+v", definition)
			}
		})
	}
}

func TestExpandedESUMetadataStaysValidForMasterDataAndRestore(t *testing.T) {
	profile := Bundled().Profiles[1]
	body, err := json.Marshal(profile)
	if err != nil {
		t.Fatal(err)
	}
	var metadata map[string]any
	if err := json.Unmarshal(body, &metadata); err != nil {
		t.Fatal(err)
	}
	if _, err := Parse(metadata); err != nil {
		t.Fatalf("expanded metadata is not importable: %v", err)
	}
	previous, err := os.ReadFile("testdata/esu-lokpilot-5-micro-dcc-2026-10-18.4.json")
	if err != nil {
		t.Fatal(err)
	}
	if !PreviousBundled(MasterDataType, profile.ID, string(previous)) {
		t.Fatal("previous live ESU revision cannot be upgraded")
	}
}
