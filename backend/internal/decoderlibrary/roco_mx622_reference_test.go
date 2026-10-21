package decoderlibrary

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestRocoMX622ReferenceRequiresExplicitOEMEvidenceAndKeepsFirmwareLimits(t *testing.T) {
	r := reviewedReference(t, "cv-ref-roco-mx622-oem-2022")
	want := []string{
		"roco-motor-685404-6-poliger-decoder-mit-stecker-nem-651",
		"roco-motor-685504-6-poliger-decoder-mit-stecker-nem-651",
		"roco-motor-686101-6-poliger-decoder-gerade-stifte-nem-651",
		"roco-motor-686201-8-poliger-decoder-nem-652",
	}
	if r.Revision != "2026-10-21.6" || r.Status != "partial" || len(r.Rows) != 40 ||
		!reflect.DeepEqual(r.ModelIDs, want) || len(r.Sources) != 2 {
		t.Fatal("MX622 evidence changed scope or became full interpretation")
	}
	if r.Sources[0].SHA256 != "02122c7f769a17c9a2102db43cbc69fff2ec561a38131679a3913dc250f571a1" ||
		r.Sources[1].SHA256 != "46aae51d2d08ef167ead121ad066f0d67683fac492f926a81d885a200af91927" ||
		!strings.Contains(r.Sources[0].Scope, "S.17") || !strings.Contains(r.Sources[1].Version, "19.08.2024") ||
		!strings.Contains(r.Note.DE, "Keine OEM-Werkswerte") || !strings.Contains(r.Note.EN, "firmware") {
		t.Fatal("catalogue identity, CV source or firmware qualification lost")
	}
	addresses := 0
	for _, row := range r.Rows {
		if row.SourceIndex != 1 || row.Page != 69 || row.Applicability != "manual" ||
			row.IndexScope != "base" || row.IndexCV31 != nil || row.IndexCV32 != nil || row.FactoryValue != "" {
			t.Fatalf("CV%d: catalogue, OEM default or index page used as CV evidence", row.CV)
		}
		end := row.CVEnd
		if end == 0 {
			end = row.CV
		}
		addresses += end - row.CV + 1
		if row.CV == 17 && (row.CVEnd != 18 || row.ValueRange != "") {
			t.Fatal("combined long address became a per-byte value range")
		}
		if row.CV == 67 && (row.CVEnd != 94 || row.ValueRange != "0-255") {
			t.Fatal("speed-table boundary lost")
		}
		if row.CV == 6 && row.ValueRange != "" || row.CV == 61 && row.ValueRange != "" {
			t.Fatal("conditional speed or incomplete mapping options became an exhaustive range")
		}
		if row.CV == 11 || row.CV == 12 || row.CV == 20 || row.CV == 97 || row.CV >= 100 {
			t.Fatal("unsupported or unchecked firmware/sound registers borrowed")
		}
	}
	if addresses != 83 {
		t.Fatal("documented address coverage changed")
	}
	gap := reviewedReference(t, "cv-gap-roco-motor")
	if gap.Revision != "2026-10-21.7" || len(gap.ModelIDs) != 1 ||
		gap.ModelIDs[0] != "roco-motor-10734-h0e-n-dcc-decoder-mit-geraden-kontaktstiften" || len(gap.Rows) != 0 {
		t.Fatal("10734 was inferred from 10735 or the MX622 OEM articles")
	}
	for _, ref := range BundledReferences() {
		if ref.ID == r.ID {
			continue
		}
		for _, model := range ref.ModelIDs {
			for _, id := range want {
				if model == id {
					t.Fatalf("OEM model retained an unrelated reference or gap: %s", model)
				}
			}
		}
	}
}

func TestPreviousRocoGapCanUpgradeWithoutMatchingLocalEdits(t *testing.T) {
	body, err := os.ReadFile("testdata/roco-gap-2026-10-21.3.json")
	if err != nil {
		t.Fatal(err)
	}
	var previous []CVReference
	if err := json.Unmarshal(body, &previous); err != nil || len(previous) != 1 {
		t.Fatalf("invalid released Roco fixture: %v", err)
	}
	old := previous[0]
	raw, _ := json.Marshal(old)
	if !PreviousBundled(ReferenceMasterDataType, old.ID, string(raw)) {
		t.Fatal("released Roco gap cannot upgrade")
	}
	old.Sources[0].Scope = "Meine eigene Modellzuordnung"
	raw, _ = json.Marshal(old)
	if PreviousBundled(ReferenceMasterDataType, old.ID, string(raw)) {
		t.Fatal("local source edit matched a released fingerprint")
	}
}
