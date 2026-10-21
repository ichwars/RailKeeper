package decoderlibrary

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

func zimoProfile(t *testing.T) Profile {
	t.Helper()
	for _, p := range Bundled().Profiles {
		if p.ID == "zimo-mn180n18" {
			return p
		}
	}
	t.Fatal("MN180N18 profile missing")
	return Profile{}
}

func TestZIMOReviewedFactsAndGuardedUpgrade(t *testing.T) {
	p := zimoProfile(t)
	if len(p.Definitions) != 68 || p.FirmwareMin != "5.29" {
		t.Fatal("reviewed scope changed")
	}
	for _, d := range p.Definitions {
		if d.Reference == nil || d.Reference.SourceIndex != 1 || d.Reference.Location == "" ||
			p.Sources[1].SHA256 != "2e40762b6f4dcd5d761fe1507d70c8b997657b2fbe98d480ce63993f01f5ec43" {
			t.Fatalf("missing reviewed evidence: CV%d", d.CV)
		}
	}
	profiles := Bundled().Profiles
	label := "MN180N18"
	rows := []Reading{{CVNumber: 8, Value: 145, DecoderProfile: label}, {CVNumber: 250, Value: 127, DecoderProfile: label}}
	match := Resolve(profiles, label, rows)
	for _, test := range []struct{ cv, minimum, maximum, factory int }{
		{2, 1, 255, 1}, {5, 0, 255, 1}, {6, 1, 255, 1}, {9, 0, 199, 55}, {16, 0, 255, 0},
		{33, 0, 255, 1}, {37, 0, 255, 2}, {46, 0, 255, 128}, {56, 0, 99, 55}, {57, 0, 255, 0},
		{134, 1, 14, 6}, {140, 0, 13, 0}, {141, 0, 255, 20}, {147, 0, 255, 100}, {153, 0, 255, 100},
		{193, 0, 255, 0}, {194, 0, 255, 0}, {830, 0, 255, 0},
	} {
		d := FindDefinition(profiles, match, Reading{CVNumber: test.cv})
		if d == nil || d.Minimum != test.minimum || d.Maximum != test.maximum || d.FactoryDefault == nil || *d.FactoryDefault != test.factory {
			t.Fatalf("wrong model fact CV%d: %+v", test.cv, d)
		}
	}
	body, _ := json.Marshal(p)
	var metadata map[string]any
	_ = json.Unmarshal(body, &metadata)
	parsed, err := Parse(metadata)
	if err != nil || !reflect.DeepEqual(p, parsed) {
		t.Fatalf("lossy profile storage: %v", err)
	}
	previous, err := os.ReadFile("testdata/mn180n18-2026-10-20.1.json")
	if err != nil || !PreviousBundled(MasterDataType, p.ID, string(previous)) {
		t.Fatalf("previous profile cannot upgrade: %v", err)
	}
}

func TestZIMOBoundsHardwareFirmwareAndUnprovenKeyFlags(t *testing.T) {
	profiles := Bundled().Profiles
	p := zimoProfile(t)
	inventory := BundledInventory()
	generic := Resolve(profiles, "Zentrale", []Reading{{CVNumber: 8, Value: 145, DecoderProfile: "Zentrale"}, {CVNumber: 250, Value: 127, DecoderProfile: "Zentrale"}}, &inventory)
	if generic.ProfileID != "dcc-basics" || generic.Identification == nil || generic.Identification.Scope != "family" {
		t.Fatalf("base type established an interface: %+v", generic)
	}
	match := Resolve(profiles, "MN180N18", []Reading{{CVNumber: 8, Value: 145, DecoderProfile: "MN180N18"}})
	for _, cv := range []int{58, 123, 151, 260, 267, 300, 309, 394, 430} {
		if d := FindDefinition(profiles, match, Reading{CVNumber: cv}); d != nil {
			t.Fatalf("future/sound/unreviewed function: CV%d", cv)
		}
	}
	for _, cv := range []int{250, 251, 830, 839, 842} {
		if d := FindDefinition(profiles, match, Reading{CVNumber: cv, IndexCV31: pointer(1), IndexCV32: pointer(1)}); d != nil {
			t.Fatalf("index alias: CV%d", cv)
		}
	}
	for _, cv := range []int{33, 37, 46, 56, 140} {
		d := FindDefinition(profiles, match, Reading{CVNumber: cv})
		if d == nil || len(d.Bits) > 0 {
			t.Fatalf("conditional field became independent bits: CV%d", cv)
		}
	}
	for cv, excluded := range map[int]int{12: 6, 22: 2, 27: 2, 28: 6} {
		d := FindDefinition(profiles, match, Reading{CVNumber: cv})
		for _, b := range d.Bits {
			if b.Bit == excluded {
				t.Fatalf("unsupported or ambiguous bit promoted: CV%d bit%d", cv, excluded)
			}
		}
	}
	for cv, bad := range map[int]int{9: 10, 19: 128, 56: 20, 57: 99, 140: 4} {
		d := FindDefinition(profiles, match, Reading{CVNumber: cv})
		for _, v := range d.AllowedValues {
			if v == bad {
				t.Fatalf("undocumented parameter accepted: CV%d=%d", cv, bad)
			}
		}
	}
	if p.Rules.ConstantBrake != nil || p.Mapping != nil {
		t.Fatal("documentation acquired unreviewed summary/mapping rules")
	}
	for _, cv := range []int{7, 65, 100, 250, 251, 839, 842} {
		d := FindDefinition(profiles, match, Reading{CVNumber: cv})
		if d == nil || d.Access != "read-only" || d.FactoryDefault != nil {
			t.Fatalf("read-only ID/measurement gained a factory value: CV%d", cv)
		}
	}
	old := match
	old.Firmware = "4.229"
	compatible := false
	old.FirmwareCompatible = &compatible
	if FindDefinition(profiles, old, Reading{CVNumber: 147}) != nil {
		t.Fatal("outside reviewed firmware scope")
	}
	if d := FindDefinition(profiles, match, Reading{CVNumber: 6}); !strings.Contains(d.Help.DE, "Wert 0 ist nicht dokumentiert") {
		t.Fatal("invalid CV6 special case")
	}
}

func TestZIMOSummaryRuleRequiresReviewedHardwareAndFirmware(t *testing.T) {
	p := zimoProfile(t)
	if !p.Rules.ZIMOMN || p.Rules.DCAnalog == nil || p.Rules.DCAnalog.CV != 12 || p.Rules.DCAnalog.Mask != 1 {
		t.Fatal("reviewed ZIMO summary rule missing")
	}
	previous, err := os.ReadFile("testdata/mn180n18-2026-10-20.2.json")
	if err != nil || !PreviousBundled(MasterDataType, p.ID, string(previous)) {
		t.Fatalf("previous ZIMO metadata cannot upgrade safely: %v", err)
	}
	for _, change := range []func(*Profile){
		func(p *Profile) { p.Model = "MN170" },
		func(p *Profile) { p.Family = "MX" },
		func(p *Profile) { p.ManufacturerID = pointer(151) },
		func(p *Profile) { p.FirmwareMin = "4.229" },
		func(p *Profile) { p.FirmwareMin = "" },
		func(p *Profile) { p.Rules.MotorCurveCV = 48 },
		func(p *Profile) { p.Rules.ConstantBrake = &BrakeRule{ForwardCV: 141, ReverseCV: 141} },
	} {
		p := zimoProfile(t)
		change(&p)
		if p.Validate() == nil {
			t.Fatal("unsupported ZIMO summary scope accepted")
		}
	}
}
