package decoderlibrary

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

func lenzProfile(t *testing.T) Profile {
	t.Helper()
	for _, p := range Bundled().Profiles {
		if p.ID == "lenz-standard-plus-v2" {
			return p
		}
	}
	t.Fatal("Lenz profile missing")
	return Profile{}
}

func TestLenzReviewedFactsAndUpgradeRemainLossless(t *testing.T) {
	p := lenzProfile(t)
	if len(p.Definitions) != 77 || !p.Rules.LenzStandardV2 || p.Revision != "2026-10-20.2" {
		t.Fatal("reviewed Lenz scope changed")
	}
	for _, d := range p.Definitions {
		if d.Reference == nil || d.Reference.Location == "" || p.Sources[d.Reference.SourceIndex].SHA256 == "" {
			t.Fatalf("CV%d lost located evidence", d.CV)
		}
	}
	rows := []Reading{{CVNumber: 8, Value: 99, DecoderProfile: "STANDARD+ V2"}}
	profiles := Bundled().Profiles
	match := Resolve(profiles, "STANDARD+ V2", rows)
	for _, test := range []struct{ cv, maximum, factory int }{
		{19, 127, 0}, {28, 3, 3}, {29, 63, 14}, {33, 31, 1}, {34, 31, 2}, {35, 31, 4},
		{36, 31, 8}, {39, 31, 16}, {47, 31, 4}, {49, 28, 0}, {50, 255, 0}, {51, 255, 0},
		{58, 255, 4}, {59, 255, 8}, {144, 31, 0}, {145, 16, 0}, {146, 16, 0}, {149, 255, 30},
	} {
		d := FindDefinition(profiles, match, Reading{CVNumber: test.cv})
		if d == nil || d.Maximum != test.maximum || d.FactoryDefault == nil || *d.FactoryDefault != test.factory {
			t.Fatalf("wrong Lenz fact CV%d: %+v", test.cv, d)
		}
	}
	body, _ := json.Marshal(p)
	var metadata map[string]any
	_ = json.Unmarshal(body, &metadata)
	parsed, err := Parse(metadata)
	if err != nil || !reflect.DeepEqual(p, parsed) {
		t.Fatalf("Lenz metadata exceeds storage or loses facts: %v", err)
	}
	previous, err := os.ReadFile("testdata/lenz-standard-plus-v2-2026-10-20.1.json")
	if err != nil || !PreviousBundled(MasterDataType, p.ID, string(previous)) {
		t.Fatalf("released Lenz metadata cannot upgrade: %v", err)
	}
}

func TestLenzDoesNotBorrowHardwareOrInferKeyFlags(t *testing.T) {
	profiles := Bundled().Profiles
	match := Resolve(profiles, "STANDARD+ V2", []Reading{{CVNumber: 8, Value: 99, DecoderProfile: "STANDARD+ V2"}})
	for _, cv := range []int{33, 39, 145, 146} {
		d := FindDefinition(profiles, match, Reading{CVNumber: cv})
		if d == nil || len(d.Bits) != 4 || d.Bits[3].Bit != 3 || !strings.Contains(d.Help.DE, "ungeklärt") && !strings.Contains(d.Help.DE, "nicht eindeutig") {
			t.Fatalf("unconfirmed output E acquired meaning: CV%d", cv)
		}
	}
	for _, cv := range []int{118, 250, 254, 255} {
		if FindDefinition(profiles, match, Reading{CVNumber: cv}) != nil {
			t.Fatalf("unconfirmed/other hardware CV%d interpreted", cv)
		}
	}
	for _, cv := range []int{49, 51, 129, 145, 149} {
		if FindDefinition(profiles, match, Reading{CVNumber: cv, IndexCV31: pointer(0), IndexCV32: pointer(0)}) != nil {
			t.Fatalf("indexed value replaced Lenz base CV%d", cv)
		}
	}
	if d := FindDefinition(profiles, match, Reading{CVNumber: 54}); d.Unit != nil || !strings.Contains(d.Help.DE, "keine exakte") {
		t.Fatal("inconsistent dwell scale became an exact conversion")
	}
	if d := FindDefinition(profiles, match, Reading{CVNumber: 60}); len(d.Bits) != 0 || !strings.Contains(d.Help.DE, "Dezimale") {
		t.Fatal("decimal effects became bit flags")
	}
}

func TestLenzSummaryRuleRequiresReviewedModelScope(t *testing.T) {
	for _, change := range []func(*Profile){
		func(p *Profile) { p.ManufacturerID = pointer(151) },
		func(p *Profile) { p.Model = "GOLD+" },
		func(p *Profile) { p.Family = "STANDARD" },
		func(p *Profile) { p.Rules.ZIMOMN = true },
		func(p *Profile) { p.Rules.ABC.Mask = 1 },
		func(p *Profile) { p.Rules.ABC = nil },
		func(p *Profile) { p.Rules.DCAnalog = &BitRule{CV: 50, Mask: 2} },
		func(p *Profile) { p.Rules.ConstantBrake = &BrakeRule{ForwardCV: 52, ReverseCV: 52} },
	} {
		p := lenzProfile(t)
		change(&p)
		if p.Validate() == nil {
			t.Fatal("Lenz rule escaped its reviewed scope")
		}
	}
}
