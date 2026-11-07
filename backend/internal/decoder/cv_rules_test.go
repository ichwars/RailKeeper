package decoder

import (
	"encoding/json"
	"errors"
	"testing"
)

func comparisonRecord() CVRecord {
	line := 17
	return CVRecord{ID: "row", VehicleID: "vehicle", CVNumber: 257, Value: 42, DecoderProfile: "  Profil ä  ", SourceFileID: "source", SourceLine: &line,
		SourceExportedAt: "2026-01-02T03:04:05", CreatedAt: "2026-01-02T03:04:05Z", UpdatedAt: "2026-02-03T04:05:06Z"}
}

func TestExistingCVComparisonTokensRemainCompatible(t *testing.T) {
	zero, one := 0, 1
	for _, test := range []struct {
		index31, index32 *int
		expected         string
	}{
		{nil, nil, "c1b90cfb588877e1c7c1e8589902b2dda3cc4dcb95c61c85e9e30636d2a95e2e"},
		{&zero, &zero, "a798f3af9e1dca4e2454d3b1606b75400b9519c9146c3a37a27296071c3bafa8"},
		{&zero, &one, "3d8e87fa0ae0abe3036f81216a44d1793142b42f1a4784e0f7c1141de367bc4b"},
	} {
		record := comparisonRecord()
		record.IndexCV31, record.IndexCV32 = test.index31, test.index32
		if CVRecordRevision(record) != test.expected {
			t.Fatal("a prepared comparison changed after package extraction")
		}
		if CVRecordValue(record).DecoderProfile != "  Profil ä  " {
			t.Fatal("stored profile text normalized")
		}
	}
}

func TestLegacyAndComparedCVWritesRemainDistinct(t *testing.T) {
	record := comparisonRecord()
	current := CVRecordRevision(record)
	empty, invalid := "", "not-a-revision"
	for _, test := range []struct {
		stored   *CVRecord
		expected *string
		err      error
	}{
		{&record, nil, nil}, {nil, nil, nil}, {nil, &empty, nil}, {&record, &empty, ErrCVConflict},
		{nil, &current, ErrCVConflict}, {&record, &current, nil}, {&record, &invalid, ErrCVValidation},
	} {
		if err := CheckCVRevision(test.stored, test.expected); !errors.Is(err, test.err) {
			t.Fatalf("comparison decision: %v, want %v", err, test.err)
		}
	}
}

func TestWorkspaceIdentityRetainsExactStoredProfilesAndDetectsReassignment(t *testing.T) {
	binding := CVBinding{"workspace", "vehicle", "  Profil ä  ", "installation", "decoder", "installed"}
	expected := "4af4d28e6eb7096bf29617f5450906edd3cc2989ee9bc9f028daab63d629cdd1"
	if CVBindingRevision(binding) != expected {
		t.Fatal("existing workspace comparison changed")
	}
	profile, err := ResolveCVProfile("vehicle", "workspace", expected, "Profil ä", &binding)
	if err != nil || profile != binding.Profile {
		t.Fatal("workspace selection lost the original profile spelling")
	}
	changed := binding
	changed.State = "removed"
	if _, err := ResolveCVProfile("vehicle", "workspace", expected, "Profil ä", &changed); !errors.Is(err, ErrCVConflict) {
		t.Fatal("stale hardware assignment accepted")
	}
	if _, err := ResolveCVProfile("another-vehicle", "workspace", expected, "Profil ä", &binding); !errors.Is(err, ErrCVValidation) {
		t.Fatal("foreign workspace accepted")
	}
	if _, err := ResolveCVProfile("vehicle", "", expected, "Profil ä", nil); !errors.Is(err, ErrCVValidation) {
		t.Fatal("comparison without an identity accepted")
	}
	input := VehicleCVDeletionInput{WorkspaceID: "workspace", ExpectedWorkspaceRevision: expected, ExpectedRevision: &expected}
	if err := CheckCVDeletionScope("vehicle", binding.Profile, input, &binding); err != nil {
		t.Fatal(err)
	}
	if err := CheckCVDeletionScope("vehicle", "Profil ä", input, &binding); !errors.Is(err, ErrCVValidation) {
		t.Fatal("deletion retargeted a differently spelled profile")
	}
	if err := CheckCVDeletionScope("vehicle", "legacy", VehicleCVDeletionInput{}, nil); err != nil {
		t.Fatal("legacy deletion contract removed")
	}
}

func TestCVProvenanceKeepsUnknownTimeZoneAndIndexPages(t *testing.T) {
	zero := 0
	valid := VehicleCVValueInput{CVNumber: 257, Value: 255, IndexCV31: &zero, IndexCV32: &zero, SourceExportedAt: "2026-01-02T03:04:05"}
	if !ValidCVValueInput(valid) {
		t.Fatal("valid indexed CV rejected")
	}
	for _, invalid := range []VehicleCVValueInput{
		{CVNumber: 256, Value: 1, IndexCV31: &zero, IndexCV32: &zero},
		{CVNumber: 257, Value: 1, IndexCV31: &zero},
		{CVNumber: 1, Value: 256},
		{CVNumber: 1, Value: 1, SourceExportedAt: "2026-01-02T03:04:05Z"},
	} {
		if ValidCVValueInput(invalid) {
			t.Fatal("invalid page, byte or invented time zone accepted")
		}
	}
	if !ValidCVExportTimestamp("") {
		t.Fatal("unknown export date was invented")
	}
}

func TestWorkingSnapshotRetainsIdentityHistoryAndSourceBoundaries(t *testing.T) {
	record := CVRecordValue(comparisonRecord())
	record.History = []VehicleCVValueHistory{{ID: "history", CVValueID: record.ID, VehicleID: record.VehicleID}}
	encode := func(values []VehicleCVValue) string {
		data, err := json.Marshal(DecoderCVSnapshotPayload{Values: values})
		if err != nil {
			t.Fatal(err)
		}
		return string(data)
	}
	valid := encode([]VehicleCVValue{record})
	if err := CheckSnapshotPayload(valid, record.VehicleID, record.DecoderProfile, map[string]bool{"source": true}); err != nil {
		t.Fatal(err)
	}
	for _, input := range []struct {
		raw, vehicle, profile string
		sources               map[string]bool
	}{
		{valid, "foreign", record.DecoderProfile, map[string]bool{"source": true}},
		{valid, record.VehicleID, "Profil ä", map[string]bool{"source": true}},
		{valid, record.VehicleID, record.DecoderProfile, map[string]bool{}},
		{encode([]VehicleCVValue{record, record}), record.VehicleID, record.DecoderProfile, map[string]bool{"source": true}},
		{valid + " {}", record.VehicleID, record.DecoderProfile, map[string]bool{"source": true}},
	} {
		if err := CheckSnapshotPayload(input.raw, input.vehicle, input.profile, input.sources); err == nil {
			t.Fatal("incompatible immutable snapshot accepted")
		}
	}
}
