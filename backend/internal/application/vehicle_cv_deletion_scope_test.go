package application_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestDecoderConfigurationDeletionScopeRejectsChangedBindingAndRollsBackWholeDraft(t *testing.T) {
	db, backup, original := decoderConfigurationFixture(t)
	s := application.NewVehicleService(db)
	vehicle := original.Tables["vehicles"][0]["id"].(string)
	current, err := s.DecoderConfigurationContext(t.Context(), vehicle)
	if err != nil {
		t.Fatal(err)
	}
	group := decoderWorkspace(t, current, " Legacy ")
	values, err := s.ListCVValues(t.Context(), vehicle)
	if err != nil {
		t.Fatal(err)
	}
	var first application.VehicleCVValue
	for _, value := range values {
		if value.CVNumber == 1 {
			first = value
		}
	}
	blob := original.Tables["decoder_cv_snapshot_sources"][0]["blob_id"].(string)
	file, err := s.CreateCVFile(t.Context(), vehicle, application.VehicleCVFileInput{FileName: "delete-scope.csv", OriginalName: "delete-scope.csv",
		DecoderProfile: group.Profile, WorkspaceID: group.ID, ExpectedWorkspaceRevision: group.BindingRevision, BlobID: blob})
	if err != nil {
		t.Fatal(err)
	}
	other, err := s.CreateCVValue(t.Context(), vehicle, application.VehicleCVValueInput{CVNumber: 1, Value: 21, DecoderProfile: "Other"})
	if err != nil {
		t.Fatal(err)
	}
	observed, err := s.ListCVValues(t.Context(), vehicle)
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range observed {
		if value.ID == other.ID {
			*other = value
		}
	}
	current, err = s.DecoderConfigurationContext(t.Context(), vehicle)
	if err != nil {
		t.Fatal(err)
	}
	installation := "installation-a"
	assigned, err := s.AssignDecoderCVWorkspace(t.Context(), vehicle, group.ID, application.DecoderCVAssignmentInput{
		ExpectedRevision: current.Revision, InstallationID: &installation, SnapshotID: strings.Repeat("f", 32)}, "", "")
	if err != nil {
		t.Fatal(err)
	}
	untouched := decoderSnapshots(t, backup)
	unchanged := func(t *testing.T) {
		t.Helper()
		after := decoderSnapshots(t, backup)
		if !reflect.DeepEqual(untouched.Tables, after.Tables) || !reflect.DeepEqual(untouched.Files, after.Files) {
			t.Fatal("rejected deletion changed rows, history, working groups, snapshots or source bytes")
		}
	}
	stale := application.VehicleCVDeletionInput{ExpectedRevision: &first.Revision, WorkspaceID: group.ID, ExpectedWorkspaceRevision: group.BindingRevision}
	if _, err := s.DeleteCVValueCompared(t.Context(), vehicle, first.ID, stale); !errors.Is(err, application.ErrVehicleCVConflict) {
		t.Fatalf("direct CV deletion accepted a previous hardware binding: %v", err)
	}
	unchanged(t)
	stale.ExpectedRevision = &file.Revision
	if _, err := s.DeleteCVFileCompared(t.Context(), vehicle, file.ID, stale); !errors.Is(err, application.ErrVehicleCVFileConflict) {
		t.Fatalf("source deletion accepted a previous hardware binding: %v", err)
	}
	unchanged(t)
	changed := application.VehicleCVValueInput{CVNumber: other.CVNumber, Value: 99, DecoderProfile: other.DecoderProfile}
	if _, err := s.ApplyCVChanges(t.Context(), vehicle, application.VehicleCVChangesInput{Changes: []application.VehicleCVChange{
		{ID: other.ID, ExpectedRevision: other.Revision, Value: &changed, WorkspaceID: group.ID, ExpectedWorkspaceRevision: group.BindingRevision},
	}}); !errors.Is(err, application.ErrVehicleValidation) {
		t.Fatalf("value change accepted misplaced deletion comparisons: %v", err)
	}
	unchanged(t)
	if _, err := s.ApplyCVChanges(t.Context(), vehicle, application.VehicleCVChangesInput{Changes: []application.VehicleCVChange{
		{ID: other.ID, ExpectedRevision: other.Revision, Value: &changed},
		{ID: first.ID, ExpectedRevision: first.Revision, Delete: true, WorkspaceID: group.ID, ExpectedWorkspaceRevision: group.BindingRevision},
	}}); !errors.Is(err, application.ErrVehicleCVConflict) {
		t.Fatalf("mixed draft accepted a previous hardware binding: %v", err)
	}
	unchanged(t)
	bound := decoderWorkspace(t, assigned, group.Profile)
	wrongGroup := decoderWorkspace(t, assigned, "Other")
	for _, input := range []application.VehicleCVDeletionInput{
		{ExpectedRevision: &first.Revision, WorkspaceID: group.ID},
		{ExpectedRevision: &first.Revision, ExpectedWorkspaceRevision: bound.BindingRevision},
		{WorkspaceID: group.ID, ExpectedWorkspaceRevision: bound.BindingRevision},
		{ExpectedRevision: &first.Revision, WorkspaceID: wrongGroup.ID, ExpectedWorkspaceRevision: wrongGroup.BindingRevision},
		{ExpectedRevision: &first.Revision, WorkspaceID: "foreign-group", ExpectedWorkspaceRevision: bound.BindingRevision},
	} {
		if _, err := s.DeleteCVValueCompared(t.Context(), vehicle, first.ID, input); !errors.Is(err, application.ErrVehicleValidation) {
			t.Fatalf("invalid deletion scope accepted: %v", err)
		}
		unchanged(t)
	}
	removed, err := s.RemoveDecoder(t.Context(), vehicle, installation, application.DecoderRemovalInput{ExpectedRevision: assigned.Revision}, "")
	if err != nil {
		t.Fatal(err)
	}
	untouched = decoderSnapshots(t, backup)
	if _, err := s.DeleteCVValueCompared(t.Context(), vehicle, first.ID, application.VehicleCVDeletionInput{
		ExpectedRevision: &first.Revision, WorkspaceID: group.ID, ExpectedWorkspaceRevision: bound.BindingRevision,
	}); !errors.Is(err, application.ErrVehicleCVConflict) {
		t.Fatalf("deletion accepted an installation removed after confirmation: %v", err)
	}
	unchanged(t)
	fresh := decoderWorkspace(t, removed, group.Profile)
	archived, err := s.DecoderCVSnapshot(t.Context(), strings.Repeat("f", 32))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.DeleteCVFileCompared(t.Context(), vehicle, file.ID, application.VehicleCVDeletionInput{
		ExpectedRevision: &file.Revision, WorkspaceID: group.ID, ExpectedWorkspaceRevision: fresh.BindingRevision,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ApplyCVChanges(t.Context(), vehicle, application.VehicleCVChangesInput{Changes: []application.VehicleCVChange{
		{ID: first.ID, ExpectedRevision: first.Revision, Delete: true, WorkspaceID: group.ID, ExpectedWorkspaceRevision: fresh.BindingRevision},
	}}); err != nil {
		t.Fatal(err)
	}
	if after, err := s.DecoderCVSnapshot(t.Context(), archived.ID); err != nil || !reflect.DeepEqual(archived, after) {
		t.Fatal("confirmed live deletion changed recorded CVs, histories or original source metadata")
	}
	remaining, err := s.ListCVValues(t.Context(), vehicle)
	if err != nil {
		t.Fatal(err)
	}
	expected := []application.VehicleCVValue{}
	for _, value := range observed {
		if value.ID != first.ID {
			expected = append(expected, value)
		}
	}
	if !reflect.DeepEqual(remaining, expected) {
		t.Fatal("confirmed deletion changed unselected values, index pages, provenance or histories")
	}
}
