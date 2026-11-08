package catalog

import (
	"encoding/json"
	"maps"
	"reflect"
	"testing"
)

func recordedFields() VehicleCatalogFields {
	fields := VehicleCatalogFields{}
	for _, name := range Names() {
		fields[name] = "recorded " + name
	}
	return fields
}

func TestProjectionChangesOnlySelectedValuesAndPreservesIndependentOverrides(t *testing.T) {
	current := recordedFields()
	chosen := maps.Clone(current)
	chosen["name"], chosen["description"] = "catalog name", "catalog description"
	before := &VehicleCatalogContext{Authority: "legacy", CurrentFields: current}
	projection, err := PlanProjection(before, &VehicleCatalogRevision{Fields: chosen}, []string{"name"})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(projection.Changes, []FieldChange{{Name: "name", Value: "catalog name"}}) {
		t.Fatal("unselected fields would change", projection.Changes)
	}
	if !reflect.DeepEqual(projection.Overrides, VehicleCatalogFields{"description": "recorded description"}) {
		t.Fatal("individual override lost", projection.Overrides)
	}
	if projection.SourceAuthority == nil || *projection.SourceAuthority != "legacy" || before.SourceAuthority != nil ||
		current["name"] != "recorded name" || chosen["description"] != "catalog description" {
		t.Fatal("planning changed recorded inputs or source authority")
	}
	chosen["name"] = "  "
	if _, err := PlanProjection(before, &VehicleCatalogRevision{Fields: chosen}, []string{"name"}); err != ErrVehicleCatalogValidation {
		t.Fatal("required field accepted", err)
	}
}

func TestBackfillReportKeepsUnlinkedRecordsAndRejectsPartialInvalidReports(t *testing.T) {
	fields := recordedFields()
	raw, err := json.Marshal(fields)
	if err != nil {
		t.Fatal(err)
	}
	rows := []RawAssignment{
		{Assignment: VehicleCatalogAssignment{TargetType: "vehicle", TargetID: "unlinked"}, LegacyJSON: string(raw)},
		{Assignment: VehicleCatalogAssignment{TargetType: "set", TargetID: "set", RevisionID: "revision"},
			LegacyJSON: string(raw), FieldsJSON: string(raw), OverridesJSON: `{"name":"individual","description":"individual description"}`},
	}
	report, err := BuildReport(rows)
	if err != nil || report.Vehicles != 1 || report.Sets != 1 || report.LinkedVehicles != 0 || report.LinkedSets != 1 || report.DifferentRepresentations != 1 {
		t.Fatal(report, err)
	}
	if !reflect.DeepEqual(report.Assignments[1].OverrideFields, []string{"description", "name"}) ||
		!reflect.DeepEqual(report.Assignments[1].DifferentFields, []string{"description", "name"}) ||
		report.Assignments[0].OverrideFields == nil {
		t.Fatal("report lost stable fields", report)
	}
	for _, invalid := range []RawAssignment{
		{LegacyJSON: `{}`},
		{Assignment: rows[1].Assignment, LegacyJSON: string(raw), FieldsJSON: `null`},
		{Assignment: rows[1].Assignment, LegacyJSON: string(raw), FieldsJSON: string(raw), OverridesJSON: `{"name":null}`},
	} {
		if partial, err := BuildReport(append(rows[:1:1], invalid)); err == nil || partial != nil {
			t.Fatal("invalid report acknowledged", partial, err)
		}
	}
}
