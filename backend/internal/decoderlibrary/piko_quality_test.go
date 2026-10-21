package decoderlibrary

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

func pikoProfile(t *testing.T) Profile {
	t.Helper()
	for _, p := range Bundled().Profiles {
		if p.ID == "piko-56500" {
			return p
		}
	}
	t.Fatal("PIKO56500 profile missing")
	return Profile{}
}

func TestPIKOExactLocatedDefinitionsAndBankIdentity(t *testing.T) {
	p := pikoProfile(t)
	if len(p.Definitions) != 111 || !p.Rules.PIKO56500 || p.Revision != "2026-10-20.2" {
		t.Fatal("wrong PIKO scope")
	}
	profiles := Bundled().Profiles
	match := Resolve(profiles, "PIKO 56500", []Reading{{CVNumber: 8, Value: 162, DecoderProfile: "PIKO 56500"}})
	if match.Status != "matched" {
		t.Fatal(match)
	}
	for cv := 257; cv <= 346; cv++ {
		d := FindDefinition(profiles, match, Reading{CVNumber: cv, IndexCV31: pointer(18), IndexCV32: pointer(0)})
		if d == nil || d.Reference == nil || !strings.Contains(d.Reference.Location, "10–11") {
			t.Fatalf("missing output CV%d", cv)
		}
		if d.Maximum != 255 && d.Maximum != 64 {
			t.Fatal("wrong bank bounds")
		}
		for _, row := range []Reading{{CVNumber: cv}, {CVNumber: cv, IndexCV31: pointer(1), IndexCV32: pointer(1)}, {CVNumber: cv, IndexCV31: pointer(18), IndexCV32: pointer(1)}} {
			if FindDefinition(profiles, match, row) != nil {
				t.Fatal("wrong bank interpreted")
			}
		}
	}
	for _, cv := range []int{101, 250, 252, 160, 162, 163} {
		if FindDefinition(profiles, match, Reading{CVNumber: cv}) != nil {
			t.Fatalf("undocumented/out-of-scope CV%d inherited", cv)
		}
	}
	for _, cv := range []int{102, 103, 104} {
		d := FindDefinition(profiles, match, Reading{CVNumber: cv})
		if d == nil || d.FactoryDefault != nil || d.Unit != nil || d.Reference.SourceIndex != 1 || !strings.Contains(d.Help.DE, "widersch") && !strings.Contains(d.Help.DE, "unterschiedliche") {
			t.Fatalf("analog conflict concealed CV%d", cv)
		}
	}
	body, _ := json.Marshal(p)
	var metadata map[string]any
	_ = json.Unmarshal(body, &metadata)
	parsed, err := Parse(metadata)
	if err != nil || !reflect.DeepEqual(parsed, p) {
		t.Fatalf("lossy PIKO metadata: %v", err)
	}
	for _, r := range BundledReferences() {
		if r.ID != "cv-ref-fab060dbad1c373b" {
			continue
		}
		if len(r.Rows) != 204 || len(r.Sources) != 2 || r.ModelIDs[0] != "piko-xp-5-1-smartdecoder-xp-5-1-plux22" {
			t.Fatal("reference scope")
		}
		for _, row := range r.Rows {
			if row.CV == 101 {
				t.Fatal("duplicate printed CV101 guessed")
			}
		}
	}
}

func TestPIKOReleasedMetadataAndAlgorithmBoundaries(t *testing.T) {
	for _, item := range []struct{ file, kind, key string }{
		{"piko-56500-2026-10-20.1.json", MasterDataType, "piko-56500"},
		{"piko-56500-reference-2026-10-20.1.json", ReferenceMasterDataType, "cv-ref-fab060dbad1c373b"},
	} {
		body, err := os.ReadFile("testdata/" + item.file)
		if err != nil || !PreviousBundled(item.kind, item.key, string(body)) {
			t.Fatalf("released metadata upgrade: %v", err)
		}
		var edited map[string]any
		_ = json.Unmarshal(body, &edited)
		edited["revision"] = "local"
		raw, _ := json.Marshal(edited)
		if PreviousBundled(item.kind, item.key, string(raw)) {
			t.Fatal("local edits trusted")
		}
	}
	for _, change := range []func(*Profile){
		func(p *Profile) { p.ManufacturerID = pointer(151) }, func(p *Profile) { p.Family = "SmartDecoder 4.1" }, func(p *Profile) { p.Model = "Next18" },
		func(p *Profile) { p.Rules.Roco10882 = true }, func(p *Profile) { p.Rules.Uhlenbrock76420 = true }, func(p *Profile) { p.Rules.ZIMOMN = true },
		func(p *Profile) { p.Rules.ABC = nil }, func(p *Profile) { p.Rules.DCAnalog.Mask = 16 }, func(p *Profile) { p.Rules.ConstantBrake = &BrakeRule{ForwardCV: 141, ReverseCV: 141} },
	} {
		p := pikoProfile(t)
		change(&p)
		if p.Validate() == nil {
			t.Fatal("PIKO rule escaped exact source scope")
		}
	}
}
