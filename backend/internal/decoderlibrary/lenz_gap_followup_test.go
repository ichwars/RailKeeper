package decoderlibrary

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestLenzGapFollowupKeepsOriginalCVBoundsAndDocumentEditions(t *testing.T) {
	for _, tc := range []struct {
		id, edition          string
		models, rows, lastCV int
	}{
		{"cv-ref-lenz-le030-1997", "Juni 1997", 1, 15, 81},
		{"cv-ref-lenz-le0511ad-2003", "3. Auflage 07/03", 2, 16, 94},
	} {
		r := reviewedReference(t, tc.id)
		if r.Revision != "2026-10-21.5" || r.Status != "partial" || len(r.ModelIDs) != tc.models ||
			len(r.Rows) != tc.rows || len(r.Sources) != 1 || !strings.Contains(r.Sources[0].Version, tc.edition) {
			t.Fatalf("wrong model, edition or coverage: %s", tc.id)
		}
		if r.Sources[0].SHA256 != "" || r.Sources[0].RetrievedAt != "2026-10-21" {
			t.Fatal("indexed text was misrepresented as a downloaded PDF")
		}
		curve := false
		for _, row := range r.Rows {
			if row.IndexScope != "base" || row.IndexCV31 != nil || row.IndexCV32 != nil ||
				row.SourceIndex != 0 || row.Page < 1 || row.Applicability != "manual" {
				t.Fatalf("source location or qualification lost: %s CV%d", tc.id, row.CV)
			}
			if row.CV == 6 || row.CV == 7 && row.FactoryValue != "" {
				t.Fatal("register 6 or a fixed firmware default became a CV fact")
			}
			if row.CV == 67 {
				curve = row.CVEnd == tc.lastCV && row.FactoryValue == "0"
			}
		}
		if !curve {
			t.Fatalf("different speed-table boundaries were merged: %s", tc.id)
		}
	}
	for _, row := range reviewedReference(t, "cv-ref-lenz-le0511ad-2003").Rows {
		if row.CV == 5 || row.CV == 9 || row.CV == 60 {
			t.Fatal("LE030 or another generation leaked into LE0511A/D")
		}
		if (row.CV == 50 || row.CV == 51 || row.CV == 57) && row.FactoryValue != "" {
			t.Fatal("partial bit columns became an invented byte default")
		}
		if row.CV == 19 && row.FactoryValue != "" {
			t.Fatal("inconsistent consist range/default was silently resolved")
		}
	}
}

func TestLatestGapMetadataCanUpgradeWithoutMatchingLocalEdits(t *testing.T) {
	body, err := os.ReadFile("testdata/reference-gaps-2026-10-21.3.json")
	if err != nil {
		t.Fatal(err)
	}
	var previous []CVReference
	if err := json.Unmarshal(body, &previous); err != nil || len(previous) != 3 {
		t.Fatalf("invalid prior gap fixture: %v", err)
	}
	for _, old := range previous {
		raw, _ := json.Marshal(old)
		if !PreviousBundled(ReferenceMasterDataType, old.ID, string(raw)) {
			t.Fatalf("last released metadata cannot upgrade: %s", old.ID)
		}
		old.Note.DE = "Meine eigene Recherche"
		raw, _ = json.Marshal(old)
		if PreviousBundled(ReferenceMasterDataType, old.ID, string(raw)) {
			t.Fatal("local changes matched an upgrade fingerprint")
		}
	}
	for _, ref := range BundledReferences() {
		for _, model := range ref.ModelIDs {
			if model != "lenz-legacy-le030" && model != "lenz-legacy-le0511a" && model != "lenz-legacy-le0511d" {
				continue
			}
			if ref.Status != "partial" || len(ref.Rows) == 0 || strings.HasPrefix(ref.ID, "cv-gap-") {
				t.Fatalf("closed model still appears as a documentation gap: %s", model)
			}
		}
	}
}
