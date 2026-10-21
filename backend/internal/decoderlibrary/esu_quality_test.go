package decoderlibrary

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestESUReservationsDoNotCountAsInterpretedFunctions(t *testing.T) {
	profiles := Bundled().Profiles
	label := "LokPilot 5 micro DCC"
	rows := []Reading{{CVNumber: 8, Value: 151, DecoderProfile: label}}
	for _, cv := range []int{57, 58, 141, 142, 143, 250} {
		rows = append(rows, Reading{CVNumber: cv, Value: 125, DecoderProfile: label})
	}
	before, _ := json.Marshal(rows)
	match := Resolve(profiles, label, rows)
	if match.KnownValues != 1 || match.TotalValues != 7 {
		t.Fatalf("reservation inflated interpretation coverage: %+v", match)
	}
	for _, row := range rows[1:] {
		d := FindDefinition(profiles, match, row)
		if d == nil || !d.Reserved || d.Reference == nil || d.FactoryDefault != nil || len(d.Bits) != 0 {
			t.Fatalf("lost reservation boundary for CV%d: %+v", row.CVNumber, d)
		}
		indexed := row
		indexed.IndexCV31, indexed.IndexCV32 = pointer(1), pointer(1)
		if FindDefinition(profiles, match, indexed) != nil {
			t.Fatal("base reservation leaked into indexed data")
		}
	}
	after, _ := json.Marshal(rows)
	if string(before) != string(after) {
		t.Fatal("documentation changed readings")
	}
}

func TestESUReviewedMetadataFitsStorageAndKeepsPreciseEvidence(t *testing.T) {
	for _, p := range Bundled().Profiles {
		if p.ID != "esu-lokpilot-5-micro-dcc" {
			continue
		}
		for _, d := range p.Definitions {
			if d.Reference == nil || p.Sources[d.Reference.SourceIndex].Version == "" {
				t.Fatalf("ESU CV%d has no located, versioned evidence", d.CV)
			}
		}
		body, _ := json.Marshal(p)
		var metadata map[string]any
		_ = json.Unmarshal(body, &metadata)
		parsed, err := Parse(metadata)
		if err != nil || !reflect.DeepEqual(p, parsed) {
			t.Fatalf("ESU metadata exceeds storage or loses facts: %v (%d bytes)", err, len(body))
		}
	}
	body, err := os.ReadFile("testdata/esu-lokpilot-5-micro-dcc-2026-10-20.1.json")
	if err != nil || !PreviousBundled(MasterDataType, "esu-lokpilot-5-micro-dcc", string(body)) {
		t.Fatalf("previous live ESU profile cannot upgrade: %v", err)
	}
}

func TestESUUnitsAndMappingBitsKeepPhysicalMeaningSeparate(t *testing.T) {
	profiles := Bundled().Profiles
	match := Resolve(profiles, "LokPilot 5 micro DCC", readings("LokPilot 5 micro DCC"))
	for _, test := range []struct {
		cv, factory int
		unit        string
	}{
		{102, 12, "16 ms"}, {112, 20, "0,065536 s"}, {113, 32, "0,032768 s"},
		{116, 50, "0,1 ms"}, {132, 80, "0,065 s"}, {149, 255, "0 ="}, {247, 0, "0,016 s"},
	} {
		d := FindDefinition(profiles, match, Reading{CVNumber: test.cv})
		if d == nil || d.Unit == nil || !strings.Contains(d.Unit.DE, test.unit) ||
			d.FactoryDefault == nil || *d.FactoryDefault != test.factory {
			t.Fatalf("wrong documented scale/default for CV%d: %+v", test.cv, d)
		}
	}
	for _, cv := range []int{3, 4, 11, 114, 115, 131, 199, 245, 254, 255} {
		d := FindDefinition(profiles, match, Reading{CVNumber: cv})
		if d != nil && d.Unit != nil {
			t.Fatalf("unverified time/distance scale inferred for CV%d", cv)
		}
	}
	for _, page := range []int{8, 9, 10, 11, 12} {
		d := FindDefinition(profiles, match, Reading{CVNumber: 260, IndexCV31: pointer(16), IndexCV32: pointer(page)})
		if d == nil || len(d.Bits) != 8 || d.Bits[1].Name.DE != "Rangiergang" ||
			d.Bits[1].On.DE != "Dieser Mappingzeile zugeordnet" {
			t.Fatalf("mapping assignment confused with live state: %+v", d)
		}
	}
	d := FindDefinition(profiles, match, Reading{CVNumber: 124})
	if d == nil || d.Bits[0].Bit != 0 || !strings.Contains(d.Help.DE, "keine Invertierung") {
		t.Fatal("direction memory confused with function inversion")
	}
}

func TestReservedDefinitionsRejectGuessedFunctionalMetadata(t *testing.T) {
	for _, change := range []func(*Definition){
		func(d *Definition) { d.Reference = nil },
		func(d *Definition) { d.FactoryDefault = pointer(0) },
		func(d *Definition) { d.Unit = &Text{DE: "s", EN: "s"} },
		func(d *Definition) { d.Access = "read-write" },
		func(d *Definition) { d.AllowedValues = []int{0} },
		func(d *Definition) { d.Bits = []Bit{{Bit: 0}} },
	} {
		d := Definition{CV: 57, Minimum: 0, Maximum: 255, Reserved: true, Topic: "communication",
			Name: Text{DE: "Reserviert", EN: "Reserved"}, Help: Text{DE: "Keine Funktion belegt", EN: "No function documented"},
			Reference: &Reference{SourceIndex: 0, Location: "S.84"}}
		change(&d)
		if d.validate() == nil {
			t.Fatal("reserved CV accepted functional metadata without evidence")
		}
	}
}
