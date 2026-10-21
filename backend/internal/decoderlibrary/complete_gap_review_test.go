package decoderlibrary

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestCompleteReviewAccountsForAll26ModelsAndProtectsLocalMetadata(t *testing.T) {
	body, err := os.ReadFile("testdata/reference-gaps-before-complete-2026-10-21.json")
	if err != nil {
		t.Fatal(err)
	}
	var previous []CVReference
	if err := json.Unmarshal(body, &previous); err != nil || len(previous) != 7 {
		t.Fatalf("invalid complete review baseline: %v", err)
	}
	models := map[string]bool{}
	withRows, withoutRows := 0, 0
	for _, old := range previous {
		current := reviewedReference(t, old.ID)
		if current.Revision != "2026-10-21.7" || !reflect.DeepEqual(old.ModelIDs, current.ModelIDs) {
			t.Fatalf("model omitted or assigned to a different family during complete review: %s", old.ID)
		}
		for _, id := range old.ModelIDs {
			if models[id] {
				t.Fatalf("review counts a model twice: %s", id)
			}
			models[id] = true
			if len(current.Rows) > 0 {
				withRows++
			} else {
				withoutRows++
			}
		}
		raw, _ := json.Marshal(old)
		if !PreviousBundled(ReferenceMasterDataType, old.ID, string(raw)) {
			t.Fatalf("last released gap cannot upgrade: %s", old.ID)
		}
		old.Note.DE = "Meine eigene Modell-/Firmwareprüfung"
		raw, _ = json.Marshal(old)
		if PreviousBundled(ReferenceMasterDataType, old.ID, string(raw)) {
			t.Fatal("operator edits matched a bundled upgrade fingerprint")
		}
	}
	if len(models) != 26 || withRows != 11 || withoutRows != 15 {
		t.Fatal("complete review no longer accounts for all 26 requested models")
	}
}

func TestESUFirmwareEvidenceDoesNotBorrowTheEntireManualOrSoundCVs(t *testing.T) {
	r := reviewedReference(t, "cv-gap-esu-lokpilot-5")
	if len(r.ModelIDs) != 5 || len(r.Rows) != 5 || r.Status != "partial" ||
		!strings.Contains(r.Note.DE, "vollständige variantenspezifische CV-Tabelle bleibt offen") {
		t.Fatal("shared firmware evidence became complete variant coverage")
	}
	versions := map[int]string{11: "5.14.183", 114: "5.4.139", 115: "5.4.139", 131: "5.4.139", 199: "5.8.156"}
	for _, row := range r.Rows {
		version, ok := versions[row.CV]
		if !ok || !strings.Contains(row.Label, version) || row.SourceIndex != 0 || row.Page != 0 ||
			row.Applicability != "manual" || row.IndexScope != "base" || row.FactoryValue != "" || row.ValueRange != "" {
			t.Fatalf("CV%d lost its firmware/source boundary or borrowed an unsupported fact", row.CV)
		}
	}
	if r.Sources[0].URL != "https://www.esu.eu/download/software/lokprogrammer/" || r.Sources[0].SHA256 != "" {
		t.Fatal("web firmware notes became a downloaded PDF")
	}
}

func TestPIKOSharedFirmwareEvidenceKeepsChangesConflictsAndProjectLimits(t *testing.T) {
	r := reviewedReference(t, "cv-gap-piko-sound-products")
	if len(r.ModelIDs) != 6 || len(r.Rows) != 28 || r.Status != "partial" || len(r.Sources) != 10 ||
		!strings.Contains(r.Note.DE, "Projektspezifische Tabellen bleiben") ||
		!strings.Contains(r.Note.DE, "Historie führt Bit4 erst mit 1.20") {
		t.Fatal("XP family coverage lost project or contradictory firmware restrictions")
	}
	for _, row := range r.Rows {
		if row.SourceIndex > 1 || row.Page != 0 || row.Applicability != "manual" || row.IndexScope != "base" ||
			row.FactoryValue != "" || row.ValueRange != "" || row.IndexCV31 != nil || row.IndexCV32 != nil {
			t.Fatalf("CV%d: project, default or bank facts inferred from firmware notes", row.CV)
		}
		if row.CV == 191 || row.CV == 224 || row.CV == 236 || row.CV == 251 || row.CV > 254 {
			t.Fatal("unbounded ff ranges or XP-G-only/sound registers transferred to H0 projects")
		}
		switch row.CV {
		case 15:
			if row.CVEnd != 16 || !strings.Contains(row.Label, "1.21") {
				t.Fatal("lock comparison lost its pair or firmware change")
			}
		case 17:
			if row.CVEnd != 18 {
				t.Fatal("combined long address was split into guessed bytes")
			}
		case 48:
			if !strings.Contains(row.Label, "1,0 V ab FW 1.16; 0,5 V ab 1.18") {
				t.Fatal("changed ABC voltage threshold became a universal conversion")
			}
		case 152:
			if row.CVEnd != 153 || !strings.Contains(row.Label, "ABV2") {
				t.Fatal("combined ABV2 evidence became an unsupported per-byte assignment")
			}
		case 254:
			if !strings.Contains(row.Label, "1.20") {
				t.Fatal("new ramp option lost its minimum documented firmware")
			}
		}
	}
}
