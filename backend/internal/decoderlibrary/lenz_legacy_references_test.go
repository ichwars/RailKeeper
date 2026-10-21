package decoderlibrary

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestLenzLegacyManualsKeepExactGenerationsAndSourceLocations(t *testing.T) {
	for _, tc := range []struct {
		key, edition string
		rows         int
	}{
		{"gold-maxi", "10440", 37},
		{"silver-mini", "06/07", 40},
		{"silver-direct", "Februar 2007", 49},
		{"silver-21", "Juli 2007", 50},
		{"lf101xf", "Februar 2007", 41},
	} {
		r := reviewedReference(t, "cv-ref-lenz-"+tc.key+"-identification")
		if r.Revision != "2026-10-21.4" || r.Status != "partial" || len(r.Rows) != tc.rows ||
			len(r.ModelIDs) != 1 || r.ModelIDs[0] != "lenz-legacy-"+tc.key ||
			!strings.Contains(r.Sources[1].Version, tc.edition) {
			t.Fatalf("wrong model, edition or coverage: %s", tc.key)
		}
		for _, row := range r.Rows {
			if row.IndexScope != "base" || row.IndexCV31 != nil || row.IndexCV32 != nil || row.Applicability != "manual" {
				t.Fatalf("manual qualification or unindexed context lost: %s CV%d", tc.key, row.CV)
			}
			if row.CV == 7 || row.CV == 8 {
				if row.SourceIndex != 0 || row.Page != 0 || row.FactoryValue != "" {
					t.Fatal("FAQ identification was replaced with a guessed firmware default")
				}
				continue
			}
			if row.SourceIndex < 1 || row.Page < 1 || r.Sources[row.SourceIndex].RetrievedAt != "2026-10-21" {
				t.Fatalf("original manual location missing: %s CV%d", tc.key, row.CV)
			}
		}
		for _, source := range r.Sources[1:] {
			if strings.Contains(source.URL, "manualslib") || strings.Contains(source.URL, "yumpu") {
				if source.SHA256 != "" {
					t.Fatal("web reading falsely advertised as a downloaded PDF")
				}
			} else if len(source.SHA256) != 64 {
				t.Fatal("downloaded original lost its fingerprint")
			}
		}
	}
}

func TestLenzLegacyReferenceMeaningDoesNotLeakAcrossDecoderTypes(t *testing.T) {
	find := func(key string, cv int) CVReferenceRow {
		t.Helper()
		for _, row := range reviewedReference(t, "cv-ref-lenz-"+key+"-identification").Rows {
			if row.CV == cv {
				return row
			}
		}
		t.Fatalf("missing %s CV%d", key, cv)
		return CVReferenceRow{}
	}
	lf := find("lf101xf", 50)
	direct := find("silver-direct", 50)
	if lf.Label != "Dimming value for output A" || lf.FactoryValue != "255" || lf.Page != 10 ||
		!strings.Contains(direct.Label, "Motor configuration") || direct.FactoryValue != "0" || direct.Page != 9 {
		t.Fatal("LF101XF lighting was confused with SILVER motor regulation")
	}
	if find("lf101xf", 56).Label != "Function mapping: dimming output C" ||
		find("lf101xf", 64).Label != "Flashing frequency at outputs C and D" {
		t.Fatal("LF100XF CV56 flashing rule leaked into LF101XF")
	}
	for _, key := range []string{"silver-mini", "gold-maxi"} {
		for _, cv := range []int{52, 54} {
			if find(key, cv).FactoryValue != "" {
				t.Fatal("conflicting braking-distance/time defaults guessed")
			}
		}
	}
	if find("silver-mini", 55).Label != "Dimmhelligkeit Ausgang A" ||
		find("silver-direct", 55).Label != "Brightness at outputs A and C" {
		t.Fatal("different output counts were merged")
	}
	if find("silver-mini", 67).CVEnd != 94 || find("silver-21", 115).ValueRange != "6-30" ||
		find("silver-21", 115).FactoryValue != "" {
		t.Fatal("documented range or ambiguous default changed")
	}
	for _, cv := range []int{126, 127} {
		if !strings.Contains(find("gold-maxi", cv).Label, "SUSI") {
			t.Fatal("SUSI selector/transfer lost its meaning")
		}
	}
}

func TestLenzIdentificationMetadataUpgradeOnlyMatchesReleasedDocuments(t *testing.T) {
	body, err := os.ReadFile("testdata/lenz-legacy-five-2026-10-21.3.json")
	if err != nil {
		t.Fatal(err)
	}
	var previous []CVReference
	if err := json.Unmarshal(body, &previous); err != nil || len(previous) != 5 {
		t.Fatalf("invalid old metadata fixture: %v", err)
	}
	for _, old := range previous {
		raw, _ := json.Marshal(old)
		if len(old.Rows) != 2 || !PreviousBundled(ReferenceMasterDataType, old.ID, string(raw)) {
			t.Fatalf("released identification record cannot be safely upgraded: %s", old.ID)
		}
		old.Note.DE = "Eigene Quellenprüfung"
		raw, _ = json.Marshal(old)
		if PreviousBundled(ReferenceMasterDataType, old.ID, string(raw)) {
			t.Fatal("local edits matched a released metadata fingerprint")
		}
	}
}
