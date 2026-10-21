package decoderlibrary

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

func dhProfile(t *testing.T) Profile {
	t.Helper()
	for _, p := range Bundled().Profiles {
		if p.ID == "dh-dh18a" {
			return p
		}
	}
	t.Fatal("DH18A profile missing")
	return Profile{}
}

func TestDHReviewedFactsRemainBoundedAndLossless(t *testing.T) {
	p := dhProfile(t)
	if len(p.Definitions) != 120 || p.Rules.MotorCurveCV != 48 || p.Rules.ConstantBrake == nil ||
		p.Rules.ConstantBrake.ForwardCV != 154 || p.Rules.ConstantBrake.EnableCV != 154 ||
		p.Rules.ConstantBrake.ReverseCV != 155 {
		t.Fatal("DH characteristic or brake source changed")
	}
	for _, d := range p.Definitions {
		if d.Reference == nil || d.Reference.Location == "" || p.Sources[d.Reference.SourceIndex].RetrievedAt != "2026-10-20" {
			t.Fatalf("missing located DH evidence for CV%d", d.CV)
		}
	}
	profiles := Bundled().Profiles
	match := Resolve(profiles, "DH18A", []Reading{{CVNumber: 8, Value: 97, DecoderProfile: "DH18A"}})
	for _, test := range []struct{ cv, maximum, factory int }{
		{14, 63, 3}, {28, 7, 3}, {60, 15, 0}, {115, 255, 0}, {119, 250, 0},
		{137, 127, 1}, {145, 161, 0}, {148, 161, 0}, {401, 29, 1}, {412, 29, 12}, {467, 31, 31},
	} {
		d := FindDefinition(profiles, match, Reading{CVNumber: test.cv})
		if d == nil || d.Maximum != test.maximum || d.FactoryDefault == nil || *d.FactoryDefault != test.factory {
			t.Fatalf("wrong DH model-specific fact CV%d: %+v", test.cv, d)
		}
	}
	body, _ := json.Marshal(p)
	var metadata map[string]any
	_ = json.Unmarshal(body, &metadata)
	parsed, err := Parse(metadata)
	if err != nil || !reflect.DeepEqual(p, parsed) {
		t.Fatalf("DH metadata exceeds storage or loses facts: %v", err)
	}
	previous, err := os.ReadFile("testdata/dh18a-2026-10-20.1.json")
	if err != nil || !PreviousBundled(MasterDataType, p.ID, string(previous)) {
		t.Fatalf("released DH profile cannot upgrade: %v", err)
	}
}

func TestDHDoesNotInferKeyFlagsOrBorrowOtherHardware(t *testing.T) {
	profiles := Bundled().Profiles
	match := Resolve(profiles, "DH18A", []Reading{{CVNumber: 8, Value: 97, DecoderProfile: "DH18A"}})
	for _, cv := range []int{115, 119, 137, 145, 401} {
		d := FindDefinition(profiles, match, Reading{CVNumber: cv})
		if d == nil || !strings.Contains(d.Help.DE, "keine") {
			t.Fatalf("assignment/condition boundary missing for CV%d", cv)
		}
	}
	for _, cv := range []int{6, 67, 80, 94, 98, 99, 101, 102, 103, 131, 132, 133, 144, 158, 159} {
		if d := FindDefinition(profiles, match, Reading{CVNumber: cv}); d != nil {
			t.Fatalf("other hardware or contradictory DH evidence interpreted: CV%d", cv)
		}
	}
	for _, cv := range []int{260, 261, 401, 467} {
		if d := FindDefinition(profiles, match, Reading{CVNumber: cv, IndexCV31: pointer(1), IndexCV32: pointer(1)}); d != nil {
			t.Fatalf("unindexed DH register confused with ESU index: CV%d", cv)
		}
	}
	for _, cv := range []int{145, 148} {
		if d := FindDefinition(profiles, match, Reading{CVNumber: cv}); d == nil || len(d.Bits) > 0 {
			t.Fatal("additive output conditions became independent bits")
		}
	}
	d := FindDefinition(profiles, match, Reading{CVNumber: 63})
	if d == nil || d.Unit != nil || !strings.Contains(d.Help.DE, "0,5 s") {
		t.Fatal("mode-dependent dwell time became a fixed unit")
	}
	d = FindDefinition(profiles, match, Reading{CVNumber: 27})
	for _, value := range d.AllowedValues {
		if value&12 != 0 {
			t.Fatal("unused brake bits acquired functionality")
		}
	}
	d = FindDefinition(profiles, match, Reading{CVNumber: 29})
	if d.FactoryDefault != nil {
		t.Fatal("ambiguous DH CV29 factory default guessed")
	}
	for _, bit := range d.Bits {
		if bit.Bit == 4 {
			t.Fatal("DCC curve-selection bit applied to DH CV48 characteristic")
		}
	}
}

func TestMotorCharacteristicRequiresAnOwnBaseMotorDefinition(t *testing.T) {
	for _, cv := range []int{-1, 29, 257, 500} {
		p := dhProfile(t)
		p.Rules.MotorCurveCV = cv
		if p.Validate() == nil {
			t.Fatalf("unreviewed characteristic register accepted: %d", cv)
		}
	}
}
