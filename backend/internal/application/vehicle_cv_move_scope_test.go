package application_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestDecoderConfigurationCVMoveComparesSourceAndTargetBindings(t *testing.T) {
	db, backup, original := decoderConfigurationFixture(t)
	s := application.NewVehicleService(db)
	vehicle := original.Tables["vehicles"][0]["id"].(string)
	current, err := s.DecoderConfigurationContext(t.Context(), vehicle)
	if err != nil {
		t.Fatal(err)
	}
	source, target := decoderWorkspace(t, current, " Legacy "), decoderWorkspace(t, current, "Other")
	before, err := s.ListCVValues(t.Context(), vehicle)
	if err != nil {
		t.Fatal(err)
	}
	var first application.VehicleCVValue
	for _, row := range before {
		if row.CVNumber == 1 {
			first = row
		}
	}
	move := scopedCV(first, target)
	move.SourceWorkspaceID, move.ExpectedSourceWorkspaceRevision = source.ID, source.BindingRevision
	installation := "installation-a"
	current, err = s.AssignDecoderCVWorkspace(t.Context(), vehicle, source.ID, application.DecoderCVAssignmentInput{
		ExpectedRevision: current.Revision, InstallationID: &installation, SnapshotID: strings.Repeat("d", 32)}, "", "")
	if err != nil {
		t.Fatal(err)
	}
	saved := decoderSnapshots(t, backup)
	unchanged := func() {
		t.Helper()
		after := decoderSnapshots(t, backup)
		if !reflect.DeepEqual(saved.Tables, after.Tables) || !reflect.DeepEqual(saved.Files, after.Files) {
			t.Fatal("rejected profile move changed saved values, history, groups or snapshots")
		}
	}
	if _, err := s.UpdateCVValue(t.Context(), vehicle, first.ID, move); !errors.Is(err, application.ErrVehicleCVConflict) {
		t.Fatalf("profile move ignored the changed original hardware binding: %v", err)
	}
	unchanged()
	nested := move
	nested.ExpectedRevision, nested.SourceWorkspaceID, nested.ExpectedSourceWorkspaceRevision = nil, "", ""
	if _, err := s.ApplyCVChanges(t.Context(), vehicle, application.VehicleCVChangesInput{Changes: []application.VehicleCVChange{
		{ID: first.ID, ExpectedRevision: first.Revision, Value: &nested, WorkspaceID: source.ID, ExpectedWorkspaceRevision: source.BindingRevision},
	}}); !errors.Is(err, application.ErrVehicleCVConflict) {
		t.Fatalf("batch profile move ignored its original hardware binding: %v", err)
	}
	unchanged()
	missing := move
	missing.SourceWorkspaceID, missing.ExpectedSourceWorkspaceRevision = "", ""
	if _, err := s.UpdateCVValue(t.Context(), vehicle, first.ID, missing); !errors.Is(err, application.ErrVehicleValidation) {
		t.Fatalf("scoped profile move accepted a missing original-group comparison: %v", err)
	}
	unchanged()
	source = decoderWorkspace(t, current, source.Profile)
	move.ExpectedSourceWorkspaceRevision = source.BindingRevision
	current, err = s.RemoveDecoder(t.Context(), vehicle, "installation-b", application.DecoderRemovalInput{ExpectedRevision: current.Revision}, "")
	if err != nil {
		t.Fatal(err)
	}
	saved = decoderSnapshots(t, backup)
	if _, err := s.UpdateCVValue(t.Context(), vehicle, first.ID, move); !errors.Is(err, application.ErrVehicleCVConflict) {
		t.Fatalf("profile move ignored the changed destination hardware binding: %v", err)
	}
	unchanged()
	move.ExpectedWorkspaceRevision = decoderWorkspace(t, current, target.Profile).BindingRevision
	nested = move
	nested.ExpectedRevision, nested.SourceWorkspaceID, nested.ExpectedSourceWorkspaceRevision = nil, "", ""
	if _, err := s.ApplyCVChanges(t.Context(), vehicle, application.VehicleCVChangesInput{Changes: []application.VehicleCVChange{
		{ID: first.ID, ExpectedRevision: first.Revision, Value: &nested, WorkspaceID: source.ID, ExpectedWorkspaceRevision: source.BindingRevision},
	}}); err != nil {
		t.Fatal(err)
	}
	after, err := s.ListCVValues(t.Context(), vehicle)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, row := range after {
		if row.ID == first.ID {
			found = true
			if row.DecoderProfile != target.Profile || row.Value != first.Value || !reflect.DeepEqual(row.History, first.History) ||
				row.SourceFileID != first.SourceFileID || row.SourceLine == nil || *row.SourceLine != *first.SourceLine || row.SourceExportedAt != first.SourceExportedAt {
				t.Fatal("confirmed profile move lost the original identity, history or source evidence")
			}
		} else {
			for _, original := range before {
				if original.ID == row.ID && !reflect.DeepEqual(original, row) {
					t.Fatal("profile move changed unselected CVs or evidence")
				}
			}
		}
	}
	if !found || len(before) != len(after) {
		t.Fatal("confirmed profile move lost or duplicated a CV identity")
	}
	final := decoderSnapshots(t, backup)
	for _, table := range []string{"decoder_cv_snapshots", "decoder_cv_snapshot_sources", "decoder_instances", "decoder_installations"} {
		if !reflect.DeepEqual(saved.Tables[table], final.Tables[table]) {
			t.Fatalf("profile move changed recorded or physical data in %s", table)
		}
	}
}
