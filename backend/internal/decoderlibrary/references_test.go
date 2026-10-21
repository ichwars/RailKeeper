package decoderlibrary

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestManufacturerReferencesAccountForEveryCatalogModel(t *testing.T) {
	models := map[string]bool{}
	for _, m := range BundledInventory().Models {
		models[m.ID] = true
	}
	covered, documented := map[string]bool{}, map[string]bool{}
	rows := 0
	for _, reference := range BundledReferences() {
		for _, id := range reference.ModelIDs {
			if !models[id] {
				t.Fatalf("reference to unknown model %s", id)
			}
			covered[id] = true
			if len(reference.Rows) > 0 {
				documented[id] = true
			}
		}
		rows += len(reference.Rows)
		if reference.Status == "table" {
			t.Fatal("bounded reference collection must not claim full table coverage")
		}
		body, err := json.Marshal(reference)
		if err != nil {
			t.Fatal(err)
		}
		var metadata map[string]any
		if err = json.Unmarshal(body, &metadata); err != nil {
			t.Fatal(err)
		}
		parsed, err := ParseReference(metadata)
		if err != nil || !reflect.DeepEqual(reference, parsed) {
			t.Fatalf("lossy reference %s: %v", reference.ID, err)
		}
	}
	if len(models) != 599 || len(covered) != len(models) || len(documented) != 580 || rows != 30410 {
		t.Fatalf("coverage changed: models=%d accounted=%d documented=%d rows=%d", len(models), len(covered), len(documented), rows)
	}
}

func TestReferencesKeepSourceAndIndexUncertaintySeparateFromRules(t *testing.T) {
	hasPIKOBank, hasUnknown, hasTemplate := false, false, false
	for _, r := range BundledReferences() {
		for _, row := range r.Rows {
			if row.SourceIndex >= len(r.Sources) {
				t.Fatal("missing source")
			}
			if row.IndexScope == "unspecified" {
				hasUnknown = true
			}
			if strings.Contains(r.Sources[0].URL, "id=28620") && row.CV == 267 {
				hasPIKOBank = row.IndexScope == "indexed" && row.IndexCV31 != nil && *row.IndexCV31 == 18 && row.IndexCV32 != nil && *row.IndexCV32 == 0
			}
			if strings.HasPrefix(r.ID, "cv-ref-lenz-template-") && row.CV == 8 {
				hasTemplate = true
				if row.FactoryValue != "" {
					t.Fatal("editor placeholder became a hardware factory value")
				}
			}
			if strings.Contains(r.Sources[0].URL, "/DH06A.pdf") && row.CV == 13 {
				t.Fatal("SX parameter 013 became DCC CV13")
			}
		}
		if r.ID == "cv-ref-kuehn-f060" {
			found := false
			for _, row := range r.Rows {
				if row.CV == 57 && row.Page == 9 {
					found = true
				}
			}
			if !found {
				t.Fatal("F060 detailed CV57 evidence lost")
			}
		}
	}
	if !hasPIKOBank || !hasUnknown || !hasTemplate {
		t.Fatal("reference evidence boundaries missing")
	}
	match := Resolve(Bundled().Profiles, "Kühn N025", []Reading{{CVNumber: 8, Value: 157}})
	if FindDefinition(Bundled().Profiles, match, Reading{CVNumber: 54}) != nil {
		t.Fatal("documentation reference enabled an unreviewed automatic rule")
	}
}

func TestReferenceMetadataRejectsBrokenSourcesBanksAndCoverage(t *testing.T) {
	original := BundledReferences()[0]
	for _, change := range []func(*CVReference){
		func(r *CVReference) { r.ModelIDs = append(r.ModelIDs, r.ModelIDs[0]) },
		func(r *CVReference) { r.Status = "unavailable" },
		func(r *CVReference) { r.Sources[0].URL = "file:///data/private" },
		func(r *CVReference) { r.Rows[0].SourceIndex = len(r.Sources) },
		func(r *CVReference) {
			r.Rows[0].IndexScope = "indexed"
			r.Rows[0].IndexCV31 = pointer(18)
			r.Rows[0].IndexCV32 = nil
		},
		func(r *CVReference) { r.Rows[0].IndexScope = "base"; r.Rows[0].IndexCV31 = pointer(18) },
		func(r *CVReference) { r.Rows[0].CV = 1025 },
		func(r *CVReference) { r.Rows[0].Applicability = "all-models-guessed" },
		func(r *CVReference) { r.Rows = append(r.Rows, r.Rows[0]) },
	} {
		body, _ := json.Marshal(original)
		var r CVReference
		_ = json.Unmarshal(body, &r)
		change(&r)
		if r.Validate() == nil {
			t.Fatal("invalid reference accepted")
		}
	}
}
