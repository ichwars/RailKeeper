package decoderlibrary

import (
	"encoding/json"
	"testing"
)

func pointer(value int) *int { return &value }
func readings(label string) []Reading {
	return []Reading{{CVNumber: 8, Value: 151, DecoderProfile: label},
		{CVNumber: 1, Value: 3, DecoderProfile: label},
		{CVNumber: 27, Value: 24, DecoderProfile: label},
		{CVNumber: 257, Value: 20, IndexCV31: pointer(16), IndexCV32: pointer(3), DecoderProfile: label}}
}

func TestResolveSeparatesEvidenceAndFirmwareScope(t *testing.T) {
	for _, test := range []struct {
		name, label, status string
		change              func([]Reading) []Reading
		known               int
	}{
		{"exact", "LokPilot 5 micro DCC", "matched", nil, 4},
		{"family", "LokPilot 5", "family", nil, 4},
		{"manufacturer only", "Unknown", "manufacturer", nil, 2},
		{"name only", "LokPilot 5 micro DCC", "suggested", func(r []Reading) []Reading { return r[1:] }, 1},
		{"conflicting manufacturer", "LokPilot 5 micro DCC", "conflict", func(r []Reading) []Reading { r[0].Value = 99; return r }, 2},
		{"unsupported", "LokPilot 5 micro DCC", "unsupported", func(r []Reading) []Reading { r[0].Protocol = "MM"; return r }, 0},
		{"older firmware", "LokPilot 5 micro DCC", "matched", func(r []Reading) []Reading {
			return append(r, Reading{CVNumber: 7, Value: 255, Description: "LokProgrammer Decoder: LokPilot 5 micro DCC (5.2.100)", DecoderProfile: r[0].DecoderProfile})
		}, 3},
		{"conflicting readings", "LokPilot 5 micro DCC", "conflict", func(r []Reading) []Reading {
			return append(r, Reading{CVNumber: 8, Value: 99, DecoderProfile: r[0].DecoderProfile})
		}, 3},
	} {
		t.Run(test.name, func(t *testing.T) {
			rows := readings(test.label)
			if test.change != nil {
				rows = test.change(rows)
			}
			result := Resolve(Bundled().Profiles, test.label, rows)
			if result.Status != test.status || result.KnownValues != test.known {
				t.Fatalf("unexpected resolution: %+v", result)
			}
			if test.name == "family" && result.ProfileName != "LokPilot 5" {
				t.Fatal("family was reported as a model")
			}
			if test.name == "older firmware" && (result.FirmwareCompatible == nil || *result.FirmwareCompatible) {
				t.Fatal("older firmware was accepted")
			}
		})
	}
}

func TestDefinitionKeepsProfileAndPageIdentities(t *testing.T) {
	catalog := Bundled().Profiles
	rows := readings("LokPilot 5 micro DCC")
	rows = append(rows, Reading{CVNumber: 8, Value: 99, DecoderProfile: "Other"})
	match := Resolve(catalog, "LokPilot 5 micro DCC", rows)
	if match.TotalValues != 4 || match.Status != "matched" {
		t.Fatalf("profiles mixed: %+v", match)
	}
	for _, test := range []struct {
		cv, index, page int
		indexed, known  bool
	}{
		{1, 0, 0, false, true}, {1, 0, 0, true, false}, {257, 16, 3, true, true},
		{257, 16, 8, true, true}, {257, 16, 13, true, true}, {257, 0, 0, true, false},
		{291, 16, 0, true, true}, {291, 0, 0, false, false},
	} {
		row := Reading{CVNumber: test.cv}
		if test.indexed {
			row.IndexCV31 = pointer(test.index)
			row.IndexCV32 = pointer(test.page)
		}
		if (FindDefinition(catalog, match, row) != nil) != test.known {
			t.Fatalf("incorrect lookup: %+v", test)
		}
	}
	catalog[1].Active = false
	if m := Resolve(catalog, "LokPilot 5 micro DCC", rows); m.ProfileID != "dcc-basics" || m.KnownValues != 2 {
		t.Fatalf("inactive model used: %+v", m)
	}
	catalog[0].Active = false
	if m := Resolve(catalog, "LokPilot 5 micro DCC", rows); m.Status != "unavailable" || m.KnownValues != 0 {
		t.Fatalf("inactive catalog used: %+v", m)
	}
}

func TestAmbiguousModelDoesNotUseFirstProfile(t *testing.T) {
	profiles := Bundled().Profiles
	copy := profiles[1]
	copy.ID = "another-model"
	profiles = append(profiles, copy)
	match := Resolve(profiles, "LokPilot 5 micro DCC", readings("LokPilot 5 micro DCC"))
	if match.Status != "conflict" || match.ProfileID != "dcc-basics" {
		t.Fatalf("ambiguous model guessed: %+v", match)
	}
}

func TestMetadataValidationRejectsUnsafeOrAmbiguousDefinitions(t *testing.T) {
	for _, change := range []func(*Profile){
		func(p *Profile) { p.Sources[0].URL = "javascript:alert(1)" },
		func(p *Profile) { p.Definitions[0].IndexCV31 = pointer(16) },
		func(p *Profile) { p.Definitions = append(p.Definitions, p.Definitions[0]) },
		func(p *Profile) { p.Mapping.Rows = 999 },
		func(p *Profile) { p.Definitions[0].Maximum = 256 },
		func(p *Profile) { p.Definitions[0].CVStride = 16 },
		func(p *Profile) { p.Definitions[0].FirmwareMin = "current" },
	} {
		profile := Bundled().Profiles[1]
		change(&profile)
		body, _ := json.Marshal(profile)
		var metadata map[string]any
		_ = json.Unmarshal(body, &metadata)
		if _, err := Parse(metadata); err == nil {
			t.Fatal("invalid metadata accepted")
		}
	}
}

func TestOptionalAliasesAreSerializedAsArrays(t *testing.T) {
	profile := Bundled().Profiles[1]
	body, _ := json.Marshal(profile)
	var metadata map[string]any
	_ = json.Unmarshal(body, &metadata)
	delete(metadata, "aliases")
	delete(metadata, "familyAliases")
	parsed, err := Parse(metadata)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Aliases == nil || parsed.FamilyAliases == nil {
		t.Fatal("API arrays would be null")
	}
}
