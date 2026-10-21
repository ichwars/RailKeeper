package decoderlibrary

import (
	"encoding/json"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func uhlenbrockProfile(t *testing.T) Profile {
	t.Helper()
	for _, p := range Bundled().Profiles {
		if p.ID == "uhlenbrock-76420" {
			return p
		}
	}
	t.Fatal("76420 profile missing")
	return Profile{}
}

func TestUhlenbrockGermanSourceResolvesEnglishConflicts(t *testing.T) {
	p := uhlenbrockProfile(t)
	if len(p.Definitions) != 43 || p.Revision != "2026-10-20.2" || !p.Rules.Uhlenbrock76420 {
		t.Fatal("wrong 76420 scope")
	}
	profiles := Bundled().Profiles
	match := Resolve(profiles, "Uhlenbrock 76420", []Reading{{CVNumber: 8, Value: 85, DecoderProfile: "Uhlenbrock 76420"}})
	for _, test := range []struct{ cv, lo, hi, factory int }{{17, 192, 231, 199}, {18, 0, 255, 208}, {55, 0, 63, 10}, {56, 0, 63, 32}, {98, 0, 3, 0}, {99, 0, 255, 0}} {
		d := FindDefinition(profiles, match, Reading{CVNumber: test.cv})
		if d == nil || d.Minimum != test.lo || d.Maximum != test.hi || d.FactoryDefault == nil || *d.FactoryDefault != test.factory {
			t.Fatalf("incorrect German fact CV%d: %+v", test.cv, d)
		}
	}
	for _, d := range p.Definitions {
		if d.Reference == nil || d.Reference.SourceIndex != 1 || !strings.Contains(d.Reference.Location, "CV") {
			t.Fatalf("missing precise source CV%d", d.CV)
		}
		if p.Sources[1].SHA256 != "afd8ec41bf923ed7bf9a34b9900cdeb289ca44cf456ab48b730bd6cd985c46c3" {
			t.Fatal("unverified PDF")
		}
		if FindDefinition(profiles, match, Reading{CVNumber: d.CV, IndexCV31: pointer(0), IndexCV32: pointer(0)}) != nil {
			t.Fatal("indexed value interpreted as base CV")
		}
	}
	for _, cv := range []int{21, 22, 27, 28, 95, 97, 100, 254, 255} {
		if FindDefinition(profiles, match, Reading{CVNumber: cv}) != nil {
			t.Fatalf("foreign CV%d borrowed", cv)
		}
	}
	d := FindDefinition(profiles, match, Reading{CVNumber: 29})
	if slices.ContainsFunc(d.Bits, func(b Bit) bool { return b.Bit == 3 }) {
		t.Fatal("unused bit3 interpreted as RailCom")
	}
	if d := FindDefinition(profiles, match, Reading{CVNumber: 60}); !reflect.DeepEqual(d.AllowedValues, []int{0, 9}) {
		t.Fatal("undocumented monitoring value accepted")
	}
	for _, cv := range []int{47, 48, 53, 54, 55, 56, 57, 58} {
		if FindDefinition(profiles, match, Reading{CVNumber: cv}).Unit != nil {
			t.Fatal("unproven physical unit")
		}
	}
	body, _ := json.Marshal(p)
	var metadata map[string]any
	_ = json.Unmarshal(body, &metadata)
	parsed, err := Parse(metadata)
	if err != nil || !reflect.DeepEqual(parsed, p) {
		t.Fatalf("metadata round-trip: %v", err)
	}
}

func TestUhlenbrockExactReleasedMetadataUpgrades(t *testing.T) {
	for _, test := range []struct{ file, kind, key string }{
		{"uhlenbrock-76420-2026-10-20.1.json", MasterDataType, "uhlenbrock-76420"},
		{"uhlenbrock-76420-reference-2026-10-20.1.json", ReferenceMasterDataType, "cv-ref-8d54ef2095152cad"},
	} {
		body, err := os.ReadFile("testdata/" + test.file)
		if err != nil || !PreviousBundled(test.kind, test.key, string(body)) {
			t.Fatalf("shipped metadata cannot upgrade: %s %v", test.file, err)
		}
		var old map[string]any
		_ = json.Unmarshal(body, &old)
		old["revision"] = "local"
		edited, _ := json.Marshal(old)
		if PreviousBundled(test.kind, test.key, string(edited)) {
			t.Fatal("local change trusted")
		}
	}
	for _, r := range BundledReferences() {
		if r.ID == "cv-ref-8d54ef2095152cad" {
			if len(r.Rows) != 34 || r.Status != "partial" {
				t.Fatal("German table incomplete")
			}
			for _, cv := range []int{55, 56, 98} {
				if !slices.ContainsFunc(r.Rows, func(row CVReferenceRow) bool { return row.CV == cv && row.SourceIndex == 1 && row.Page == 6 }) {
					t.Fatal("new raw evidence missing")
				}
			}
		}
	}
}

func TestUhlenbrockRuleCannotEscapeReviewedModel(t *testing.T) {
	for _, change := range []func(*Profile){
		func(p *Profile) { p.ManufacturerID = pointer(151) }, func(p *Profile) { p.Family = "IntelliSound" },
		func(p *Profile) { p.Model = "IntelliDrive 76425" }, func(p *Profile) { p.Kind = "standard" },
		func(p *Profile) { p.Rules.KuehnN45 = true }, func(p *Profile) { p.Rules.LenzStandardV2 = true }, func(p *Profile) { p.Rules.ZIMOMN = true },
		func(p *Profile) { p.Rules.ABC = &BitRule{CV: 27, Mask: 3} }, func(p *Profile) { p.Rules.DCAnalog = &BitRule{CV: 51, Mask: 2} },
		func(p *Profile) { p.Rules.ConstantBrake = &BrakeRule{ForwardCV: 254, ReverseCV: 255} }, func(p *Profile) { p.Rules.MotorCurveCV = 56 },
	} {
		p := uhlenbrockProfile(t)
		change(&p)
		if p.Validate() == nil {
			t.Fatal("76420 rule escaped scope")
		}
	}
	for _, label := range []string{"76425", "76320", "IntelliDrive", "76400", "Zentrale"} {
		match := Resolve(Bundled().Profiles, label, []Reading{{CVNumber: 8, Value: 85, DecoderProfile: label}})
		if match.ProfileID == "uhlenbrock-76420" {
			t.Fatalf("76420 rules borrowed for %s", label)
		}
	}
}
