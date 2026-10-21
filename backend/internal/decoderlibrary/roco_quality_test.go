package decoderlibrary

import (
	"encoding/json"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func rocoProfile(t *testing.T) Profile {
	t.Helper()
	for _, p := range Bundled().Profiles {
		if p.ID == "roco-10882" {
			return p
		}
	}
	t.Fatal("Roco10882 profile missing")
	return Profile{}
}

func TestRocoLocatedFactsAndModelBoundaries(t *testing.T) {
	p := rocoProfile(t)
	if len(p.Definitions) != 35 || p.Revision != "2026-10-20.2" || !p.Rules.Roco10882 {
		t.Fatal("wrong Roco scope")
	}
	profiles := Bundled().Profiles
	match := Resolve(profiles, "Roco 10882", []Reading{{CVNumber: 8, Value: 145, DecoderProfile: "Roco 10882"}})
	for _, test := range []struct{ cv, lo, hi, factory int }{{1, 1, 127, 3}, {2, 1, 255, 4}, {3, 1, 255, 6}, {4, 1, 255, 2}, {5, 1, 255, 252}, {6, 1, 252, 85}, {17, 192, 231, 192}, {18, 0, 255, 128}, {28, 0, 3, 3}, {29, 0, 255, 14}, {60, 0, 255, 0}, {112, 0, 255, 0}, {124, 0, 255, 0}} {
		d := FindDefinition(profiles, match, Reading{CVNumber: test.cv})
		if d == nil || d.Minimum != test.lo || d.Maximum != test.hi || d.FactoryDefault == nil || *d.FactoryDefault != test.factory {
			t.Fatalf("wrong source fact CV%d: %+v", test.cv, d)
		}
	}
	for _, d := range p.Definitions {
		if d.Reference == nil || !strings.Contains(d.Reference.Location, "CV") {
			t.Fatalf("missing located source CV%d", d.CV)
		}
		if FindDefinition(profiles, match, Reading{CVNumber: d.CV, IndexCV31: pointer(0), IndexCV32: pointer(0)}) != nil {
			t.Fatal("indexed CV interpreted as base")
		}
	}
	for _, cv := range []int{19, 21, 22, 27, 66, 95, 152, 250, 254, 255} {
		if FindDefinition(profiles, match, Reading{CVNumber: cv}) != nil {
			t.Fatalf("undocumented/other model CV%d borrowed", cv)
		}
	}
	for _, cv := range []int{37, 38, 39, 40, 41, 42, 43, 44, 45, 46} {
		if FindDefinition(profiles, match, Reading{CVNumber: cv}).FactoryDefault != nil {
			t.Fatal("shared/conflicting mapping default applied to10882")
		}
	}
	if d := FindDefinition(profiles, match, Reading{CVNumber: 114}); len(d.Bits) != 4 || d.Bits[3].Bit != 3 {
		t.Fatal("10883 outputs borrowed")
	}
	if d := FindDefinition(profiles, match, Reading{CVNumber: 29}); !strings.Contains(d.Help.DE, "fälschlich Bit3") {
		t.Fatal("source conflict lost")
	}
	for _, cv := range []int{3, 4, 60} {
		if FindDefinition(profiles, match, Reading{CVNumber: cv}).Unit != nil {
			t.Fatal("unproven unit")
		}
	}
	body, _ := json.Marshal(p)
	var metadata map[string]any
	_ = json.Unmarshal(body, &metadata)
	parsed, err := Parse(metadata)
	if err != nil || !reflect.DeepEqual(parsed, p) {
		t.Fatalf("Roco metadata round-trip: %v", err)
	}
	for _, r := range BundledReferences() {
		if r.ID == "cv-ref-44b6dd44ba5cb681" {
			if len(r.Rows) != 22 {
				t.Fatal("missing reviewed reference rows")
			}
			if slices.ContainsFunc(r.Rows, func(row CVReferenceRow) bool { return row.CV == 152 }) {
				t.Fatal("10883-only reference retained")
			}
			if !slices.ContainsFunc(r.Rows, func(row CVReferenceRow) bool {
				return row.CV == 2 && row.ValueRange == "1-255" && row.FactoryValue == "4"
			}) {
				t.Fatal("range copied into factory column")
			}
		}
	}
}

func TestRocoExactReleasedMetadataAndScope(t *testing.T) {
	for _, item := range []struct{ file, kind, key string }{
		{"roco-10882-2026-10-20.1.json", MasterDataType, "roco-10882"},
		{"roco-10882-reference-2026-10-20.1.json", ReferenceMasterDataType, "cv-ref-44b6dd44ba5cb681"},
	} {
		body, err := os.ReadFile("testdata/" + item.file)
		if err != nil || !PreviousBundled(item.kind, item.key, string(body)) {
			t.Fatalf("old Roco metadata cannot upgrade: %v", err)
		}
		var old map[string]any
		_ = json.Unmarshal(body, &old)
		old["revision"] = "local"
		edited, _ := json.Marshal(old)
		if PreviousBundled(item.kind, item.key, string(edited)) {
			t.Fatal("local edit trusted")
		}
	}
	for _, change := range []func(*Profile){
		func(p *Profile) { p.ManufacturerID = pointer(161) }, func(p *Profile) { p.Family = "ZIMO MX" }, func(p *Profile) { p.Model = "10883" },
		func(p *Profile) { p.Kind = "standard" }, func(p *Profile) { p.Rules.ZIMOMN = true }, func(p *Profile) { p.Rules.Uhlenbrock76420 = true },
		func(p *Profile) { p.Rules.LenzStandardV2 = true }, func(p *Profile) { p.Rules.KuehnN45 = true },
		func(p *Profile) { p.Rules.ABC = &BitRule{CV: 27, Mask: 3} }, func(p *Profile) { p.Rules.DCAnalog = &BitRule{CV: 12, Mask: 1} },
		func(p *Profile) { p.Rules.ConstantBrake = &BrakeRule{ForwardCV: 140, ReverseCV: 140} }, func(p *Profile) { p.Rules.MotorCurveCV = 48 },
	} {
		p := rocoProfile(t)
		change(&p)
		if p.Validate() == nil {
			t.Fatal("Roco rule escaped source scope")
		}
	}
	for _, label := range []string{"Roco 10883", "Roco 10884", "Roco 10895", "MX630", "MN180N18", "Zentrale"} {
		m := Resolve(Bundled().Profiles, label, []Reading{{CVNumber: 8, Value: 145, DecoderProfile: label}})
		if m.ProfileID == "roco-10882" {
			t.Fatalf("10882 rules borrowed for %s", label)
		}
	}
}
