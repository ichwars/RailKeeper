package decoderlibrary

import (
	"encoding/json"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func kuehnProfile(t *testing.T) Profile {
	t.Helper()
	for _, p := range Bundled().Profiles {
		if p.ID == "kuehn-n45" {
			return p
		}
	}
	t.Fatal("N45 profile missing")
	return Profile{}
}

func TestKuehnLocatedFactsAndUpgrade(t *testing.T) {
	p := kuehnProfile(t)
	if len(p.Definitions) != 59 || !p.Rules.KuehnN45 || p.Revision != "2026-10-20.2" {
		t.Fatal("reviewed N45 scope changed")
	}
	for _, d := range p.Definitions {
		if d.Reference == nil || !strings.Contains(d.Reference.Location, "CV") {
			t.Fatalf("missing register source CV%d", d.CV)
		}
	}
	profiles := Bundled().Profiles
	m := Resolve(profiles, "Kühn N45", []Reading{{CVNumber: 8, Value: 157, DecoderProfile: "Kühn N45"}})
	for _, test := range []struct{ cv, maximum, factory int }{
		{5, 255, 0}, {6, 255, 0}, {27, 63, 0}, {28, 7, 0}, {29, 63, 14},
		{51, 255, 0}, {52, 255, 0}, {55, 99, 0}, {60, 99, 0}, {61, 99, 0},
		{62, 99, 0}, {63, 98, 48}, {112, 63, 0}, {115, 9, 4},
	} {
		d := FindDefinition(profiles, m, Reading{CVNumber: test.cv})
		if d == nil || d.Maximum != test.maximum || d.FactoryDefault == nil || *d.FactoryDefault != test.factory {
			t.Fatalf("wrong N45 fact CV%d: %+v", test.cv, d)
		}
	}
	if d := FindDefinition(profiles, m, Reading{CVNumber: 63}); !slices.Contains(d.AllowedValues, 48) || slices.Contains(d.AllowedValues, 49) {
		t.Fatal("invalid decimal holding digit acquired meaning")
	}
	for _, test := range []struct{ file, kind, key string }{
		{"kuehn-n45-2026-10-20.1.json", MasterDataType, p.ID},
		{"kuehn-n45-reference-2026-10-20.1.json", ReferenceMasterDataType, "cv-ref-kuehn-n45"},
	} {
		body, err := os.ReadFile("testdata/" + test.file)
		if err != nil || !PreviousBundled(test.kind, test.key, string(body)) {
			t.Fatalf("shipped N45 metadata cannot upgrade: %s %v", test.file, err)
		}
		var old map[string]any
		_ = json.Unmarshal(body, &old)
		old["revision"] = "local-edit"
		edited, _ := json.Marshal(old)
		if PreviousBundled(test.kind, test.key, string(edited)) {
			t.Fatal("operator metadata edit trusted as bundled")
		}
	}
	body, _ := json.Marshal(p)
	var metadata map[string]any
	_ = json.Unmarshal(body, &metadata)
	parsed, err := Parse(metadata)
	if err != nil || !reflect.DeepEqual(parsed, p) {
		t.Fatalf("N45 metadata does not round-trip: %v", err)
	}
}

func TestKuehnShiftedMappingsAndSourceConflicts(t *testing.T) {
	profiles := Bundled().Profiles
	m := Resolve(profiles, "N45", []Reading{{CVNumber: 8, Value: 157, DecoderProfile: "N45"}})
	for _, test := range []struct {
		cv, bit int
		output  string
	}{
		{35, 2, "C"}, {37, 3, "D"}, {38, 7, "C"}, {39, 1, "E"},
		{40, 2, "F"}, {42, 5, "A"}, {43, 2, "A"}, {46, 7, "F"},
	} {
		d := FindDefinition(profiles, m, Reading{CVNumber: test.cv})
		if d == nil || !slices.ContainsFunc(d.Bits, func(b Bit) bool { return b.Bit == test.bit && b.Name.DE == "Ausgang "+test.output }) {
			t.Fatalf("wrong shifted output CV%d bit%d", test.cv, test.bit)
		}
	}
	for _, cv := range []int{55, 60, 61, 62, 63} {
		d := FindDefinition(profiles, m, Reading{CVNumber: cv})
		if len(d.Bits) != 0 || !strings.Contains(d.Help.DE, "Dezimale") {
			t.Fatalf("decimal parameter interpreted as bits: CV%d", cv)
		}
		if FindDefinition(profiles, m, Reading{CVNumber: cv, IndexCV31: pointer(0), IndexCV32: pointer(0)}) != nil {
			t.Fatal("indexed value became base N45 CV")
		}
	}
	for _, cv := range []int{64, 114, 116, 117, 125, 254, 255} {
		if FindDefinition(profiles, m, Reading{CVNumber: cv}) != nil {
			t.Fatalf("unconfirmed/foreign register interpreted CV%d", cv)
		}
	}
	for _, cv := range []int{55, 123} {
		if FindDefinition(profiles, m, Reading{CVNumber: cv}).Unit != nil {
			t.Fatalf("unproven complete timing scale assigned CV%d", cv)
		}
	}
	for _, label := range []string{"N025", "T62", "T65", "T125", "T145", "F060", "N45-18"} {
		other := Resolve(profiles, label, []Reading{{CVNumber: 8, Value: 157, DecoderProfile: label}})
		if other.ProfileID == pIDKuehn || FindDefinition(profiles, other, Reading{CVNumber: 55}) != nil {
			t.Fatalf("N45 rule applied to %s", label)
		}
	}
}

const pIDKuehn = "kuehn-n45"

func TestKuehnSummaryRequiresReviewedScope(t *testing.T) {
	for _, change := range []func(*Profile){
		func(p *Profile) { p.ManufacturerID = pointer(151) },
		func(p *Profile) { p.Family = "T" }, func(p *Profile) { p.Model = "N45-18" },
		func(p *Profile) { p.Rules.ZIMOMN = true }, func(p *Profile) { p.Rules.LenzStandardV2 = true },
		func(p *Profile) { p.Rules.ABC = nil }, func(p *Profile) { p.Rules.ABC.Mask = 1 },
		func(p *Profile) { p.Rules.DCAnalog = &BitRule{CV: 56, Mask: 4} },
		func(p *Profile) { p.Rules.ConstantBrake = &BrakeRule{ForwardCV: 123, ReverseCV: 124} },
	} {
		p := kuehnProfile(t)
		change(&p)
		if p.Validate() == nil {
			t.Fatal("N45 rule escaped reviewed scope")
		}
	}
}
