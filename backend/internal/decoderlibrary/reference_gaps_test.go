package decoderlibrary

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func reviewedReference(t *testing.T, id string) CVReference {
	t.Helper()
	for _, r := range BundledReferences() {
		if r.ID == id {
			return r
		}
	}
	t.Fatalf("reference missing: %s", id)
	return CVReference{}
}

func TestGapReviewKeepsOverlappingCVBanksDistinct(t *testing.T) {
	r := reviewedReference(t, "cv-ref-piko-36520-br491")
	banks := map[int]CVReferenceRow{}
	for _, row := range r.Rows {
		if row.CV != 257 {
			continue
		}
		if row.IndexScope != "indexed" || row.IndexCV31 == nil || row.IndexCV32 == nil || *row.IndexCV32 != 0 {
			t.Fatal("PIKO effect/sound bank became a base CV")
		}
		banks[*row.IndexCV31] = row
	}
	if len(banks) != 2 || banks[18].Label != "Ausgang F0v Effekt A" || banks[18].Page != 4 || banks[16].Label != "Gesamtlautstärke" || banks[16].Page != 5 {
		t.Fatal("BR491 effect and sound evidence merged or lost its source page")
	}
	for _, row := range r.Rows {
		if row.FactoryValue != "" || row.ValueRange != "" {
			t.Fatal("contradictory project range/default columns must remain unimported")
		}
	}
	r = reviewedReference(t, "cv-gap-uhlenbrock-intellidrive")
	seen := map[int]bool{}
	for _, row := range r.Rows {
		if row.CV == 257 {
			if row.IndexScope != "indexed" || row.IndexCV31 == nil || *row.IndexCV31 != 8 || row.IndexCV32 == nil || row.CVEnd != 512 || row.Page != 6 {
				t.Fatal("74xxx mapping/PWM bank lost explicit index context")
			}
			seen[*row.IndexCV32] = true
		}
	}
	if len(seen) != 4 || !seen[0] || !seen[1] || !seen[3] || !seen[4] {
		t.Fatal("74xxx banks differ from manufacturer table")
	}
}

func TestGapReviewDoesNotBorrowCVMappingsAcrossGenerations(t *testing.T) {
	for _, tc := range []struct{ id, text string }{
		{"cv-ref-piko-56123-archive", "Regelfrequenz"},
		{"cv-ref-lenz-lf100xf-2001", "Blinkfrequenz"},
	} {
		r := reviewedReference(t, tc.id)
		found := false
		for _, row := range r.Rows {
			if row.CV == 56 {
				found = strings.Contains(row.Label, tc.text)
			}
		}
		if !found {
			t.Fatalf("different CV56 meanings merged: %s", tc.id)
		}
	}
	for _, r := range BundledReferences() {
		for _, model := range r.ModelIDs {
			if model == "lenz-legacy-le103x" || model == "lenz-legacy-le110-v1" || model == "lenz-silver-silver-mini-v2" {
				if len(r.Rows) != 0 {
					t.Fatalf("similar name or unrelated product download became evidence: %s", model)
				}
			}
		}
	}
	r := reviewedReference(t, "cv-ref-lenz-le077xf-archive")
	for _, row := range r.Rows {
		if row.CV == 5 || row.CV == 6 {
			t.Fatal("unsupported CV5/6 borrowed from a different Lenz generation")
		}
	}
	r = reviewedReference(t, "cv-ref-lenz-silver-mini-identification")
	for _, row := range r.Rows {
		if row.CV == 43 || row.CV == 62 || row.CV == 63 || row.CV == 64 {
			t.Fatal("unsupported mapping or outputs borrowed from a different SILVER generation")
		}
	}
}

func TestGapReviewKeepsSUSIAddressesAndNonCVHardwareExplicit(t *testing.T) {
	r := reviewedReference(t, "cv-gap-uhlenbrock-sound-legacy")
	seen := map[int]bool{}
	for _, row := range r.Rows {
		if row.CV == 902 || row.CV == 942 || row.CV == 982 {
			if row.IndexScope != "base" || row.IndexCV31 != nil || row.Page != 8 {
				t.Fatal("SUSI address range was incorrectly converted to CV31/32 indexing")
			}
			seen[row.CV] = true
		}
	}
	if len(seen) != 3 || !strings.Contains(r.Note.DE, "separat bekannt") {
		t.Fatal("alternative SUSI address ranges lost their restriction")
	}
	for _, id := range []string{"cv-gap-lenz-non-cv-modules", "cv-gap-lenz-motorola"} {
		r := reviewedReference(t, id)
		if r.Status != "unavailable" || len(r.Rows) != 0 || !strings.Contains(r.Note.DE, "DCC") {
			t.Fatal("non-DCC hardware received invented DCC CVs")
		}
	}
}

func TestAllPreviousGapMetadataRemainEligibleForSafeUpgrade(t *testing.T) {
	body, err := os.ReadFile("testdata/reference-gaps-2026-10-20.1.json")
	if err != nil {
		t.Fatal(err)
	}
	var previous []CVReference
	if err = json.Unmarshal(body, &previous); err != nil {
		t.Fatal(err)
	}
	models := 0
	for _, old := range previous {
		models += len(old.ModelIDs)
		raw, _ := json.Marshal(old)
		if !PreviousBundled(ReferenceMasterDataType, old.ID, string(raw)) {
			t.Fatalf("released gap cannot upgrade: %s", old.ID)
		}
		current := reviewedReference(t, old.ID)
		revision := "2026-10-21.3"
		if old.ID == "cv-gap-lenz-legacy" || old.ID == "cv-gap-piko-sound-products" || old.ID == "cv-gap-uhlenbrock-intellisound-8" {
			revision = "2026-10-21.5"
		}
		if old.ID == "cv-gap-roco-motor" {
			revision = "2026-10-21.6"
		}
		switch old.ID {
		case "cv-gap-esu-lokpilot-5", "cv-gap-piko-sound-products", "cv-gap-lenz-legacy",
			"cv-gap-lenz-silver", "cv-gap-kuehn-t", "cv-gap-uhlenbrock-intellisound-8", "cv-gap-roco-motor":
			revision = "2026-10-21.7"
		}
		if current.Revision != revision {
			t.Fatalf("gap was not reviewed: %s", old.ID)
		}
		old.Note.DE = "Meine lokale Dokumentation"
		raw, _ = json.Marshal(old)
		if PreviousBundled(ReferenceMasterDataType, old.ID, string(raw)) {
			t.Fatal("local reference edit became eligible for replacement")
		}
	}
	if models != 67 {
		t.Fatal("review no longer accounts for the 67 original gap models")
	}
}
