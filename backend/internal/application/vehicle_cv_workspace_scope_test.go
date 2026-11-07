package application_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
)

func decoderWorkspace(t *testing.T, current *application.DecoderConfigurationContext, profile string) application.DecoderCVWorkspace {
	t.Helper()
	for _, item := range current.Workspaces {
		if item.Profile == profile {
			return item
		}
	}
	t.Fatalf("missing exact group %q", profile)
	return application.DecoderCVWorkspace{}
}

func scopedCV(value application.VehicleCVValue, workspace application.DecoderCVWorkspace) application.VehicleCVValueInput {
	return application.VehicleCVValueInput{CVNumber: value.CVNumber, Value: value.Value, IndexCV31: value.IndexCV31, IndexCV32: value.IndexCV32,
		DecoderProfile: strings.TrimSpace(workspace.Profile), WorkspaceID: workspace.ID, ExpectedWorkspaceRevision: workspace.BindingRevision,
		ExpectedRevision: &value.Revision, Description: value.Description, Category: value.Category, Protocol: value.Protocol,
		SourceFileID: value.SourceFileID, SourceLine: value.SourceLine, SourceExportedAt: value.SourceExportedAt}
}

func TestDecoderConfigurationCVScopeKeepsDistinctRawProfilesAcrossAllWritersAndRestore(t *testing.T) {
	db, backup, original := decoderConfigurationFixture(t)
	s := application.NewVehicleService(db)
	vehicle := original.Tables["vehicles"][0]["id"].(string)
	normal, err := s.CreateCVValue(t.Context(), vehicle, application.VehicleCVValueInput{CVNumber: 1, Value: 20, DecoderProfile: " Legacy "})
	if err != nil {
		t.Fatal(err)
	}
	if normal.DecoderProfile != "Legacy" {
		t.Fatal("legacy caller cleaning changed")
	}
	current, err := s.DecoderConfigurationContext(t.Context(), vehicle)
	if err != nil {
		t.Fatal(err)
	}
	rawGroup, normalGroup := decoderWorkspace(t, current, " Legacy "), decoderWorkspace(t, current, "Legacy")
	values, err := s.ListCVValues(t.Context(), vehicle)
	if err != nil {
		t.Fatal(err)
	}
	var raw application.VehicleCVValue
	unchosen := map[string]application.VehicleCVValue{}
	for _, item := range values {
		if item.DecoderProfile == rawGroup.Profile && item.CVNumber == 1 {
			raw = item
		} else if item.CVNumber == 257 {
			unchosen[item.ID] = item
		}
	}
	zero, one, empty := 0, 1, ""
	newInput := application.VehicleCVValueInput{CVNumber: 257, Value: 30, IndexCV31: &one, IndexCV32: &zero,
		DecoderProfile: "Legacy", WorkspaceID: rawGroup.ID, ExpectedWorkspaceRevision: rawGroup.BindingRevision, ExpectedRevision: &empty}
	created, err := s.CreateCVValue(t.Context(), vehicle, newInput)
	if err != nil || created.DecoderProfile != rawGroup.Profile {
		t.Fatalf("scoped create merged exact labels: %v", err)
	}
	a, b, c := scopedCV(raw, rawGroup), scopedCV(*created, rawGroup), scopedCV(*normal, normalGroup)
	a.Value, b.Value, c.Value = 7, 31, 21
	a.ExpectedRevision, b.ExpectedRevision, c.ExpectedRevision = nil, nil, nil
	if _, err := s.ApplyCVChanges(t.Context(), vehicle, application.VehicleCVChangesInput{Changes: []application.VehicleCVChange{
		{ID: raw.ID, ExpectedRevision: raw.Revision, Value: &a}, {ID: created.ID, ExpectedRevision: created.Revision, Value: &b}, {ID: normal.ID, ExpectedRevision: normal.Revision, Value: &c},
	}}); err != nil {
		t.Fatal(err)
	}
	imports := application.VehicleCVImportInput{Values: []application.VehicleCVValueInput{
		{CVNumber: 2, Value: 8, DecoderProfile: "Legacy", WorkspaceID: rawGroup.ID, ExpectedWorkspaceRevision: rawGroup.BindingRevision, ExpectedRevision: &empty},
		{CVNumber: 2, Value: 9, DecoderProfile: "Legacy", WorkspaceID: normalGroup.ID, ExpectedWorkspaceRevision: normalGroup.BindingRevision, ExpectedRevision: &empty},
	}}
	requestProfiles := []string{imports.Values[0].DecoderProfile, imports.Values[1].DecoderProfile}
	if _, err := s.ImportCVValues(t.Context(), vehicle, imports); err != nil {
		t.Fatal(err)
	}
	if imports.Values[0].DecoderProfile != requestProfiles[0] || imports.Values[1].DecoderProfile != requestProfiles[1] {
		t.Fatal("import mutated its confirmation payload")
	}
	file, err := s.CreateCVFile(t.Context(), vehicle, application.VehicleCVFileInput{FileName: "scoped.csv", OriginalName: "scoped.csv", StoragePath: "uploads/scoped.csv",
		DecoderProfile: "Legacy", WorkspaceID: rawGroup.ID, ExpectedWorkspaceRevision: rawGroup.BindingRevision})
	if err != nil || file.DecoderProfile != rawGroup.Profile {
		t.Fatalf("file lost exact working group: %v", err)
	}
	afterValues, err := s.ListCVValues(t.Context(), vehicle)
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range afterValues {
		if old, ok := unchosen[value.ID]; ok && !reflect.DeepEqual(value, old) {
			t.Fatal("scoped edits changed unchosen unknown/indexed CV evidence")
		}
		if value.ID == raw.ID && (value.DecoderProfile != rawGroup.Profile || value.Value != 7 || len(value.History) != 2 || value.SourceFileID != raw.SourceFileID) {
			t.Fatal("raw identity, history or source changed")
		}
		if value.ID == normal.ID && (value.DecoderProfile != normalGroup.Profile || value.Value != 21) {
			t.Fatal("normal identity was overwritten")
		}
	}
	after := decoderSnapshots(t, backup)
	for _, table := range []string{"decoder_cv_snapshots", "decoder_cv_snapshot_sources", "decoder_instances", "decoder_installations"} {
		if !reflect.DeepEqual(original.Tables[table], after.Tables[table]) {
			t.Fatalf("ordinary scoped writes changed %s", table)
		}
	}
	if _, err := backup.Import(t.Context(), after); err != nil {
		t.Fatal(err)
	}
	if restored := decoderSnapshots(t, backup); !reflect.DeepEqual(catalogBackupTableSnapshot(t, after), catalogBackupTableSnapshot(t, restored)) {
		t.Fatal("scoped data changed through restore")
	}
}

func TestDecoderConfigurationCVScopeRejectsStaleHardwareBindingWithoutPartialWrites(t *testing.T) {
	db, backup, original := decoderConfigurationFixture(t)
	s := application.NewVehicleService(db)
	vehicle := original.Tables["vehicles"][0]["id"].(string)
	before, err := s.DecoderConfigurationContext(t.Context(), vehicle)
	if err != nil {
		t.Fatal(err)
	}
	workspace := decoderWorkspace(t, before, " Legacy ")
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
	input := scopedCV(first, workspace)
	input.Value = 42
	installation := "installation-a"
	afterBinding, err := s.AssignDecoderCVWorkspace(t.Context(), vehicle, workspace.ID, application.DecoderCVAssignmentInput{
		ExpectedRevision: before.Revision, InstallationID: &installation, SnapshotID: strings.Repeat("a", 32)}, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if decoderWorkspace(t, afterBinding, workspace.Profile).BindingRevision == workspace.BindingRevision {
		t.Fatal("hardware change did not change binding comparison")
	}
	saved := decoderSnapshots(t, backup)
	for _, scenario := range []struct {
		name  string
		write func() error
	}{
		{"single", func() error { _, err := s.UpdateCVValue(t.Context(), vehicle, first.ID, input); return err }},
		{"import", func() error {
			_, err := s.ImportCVValues(t.Context(), vehicle, application.VehicleCVImportInput{Values: []application.VehicleCVValueInput{input}})
			return err
		}},
		{"draft", func() error {
			nested := input
			nested.ExpectedRevision = nil
			_, err := s.ApplyCVChanges(t.Context(), vehicle, application.VehicleCVChangesInput{Changes: []application.VehicleCVChange{{ID: first.ID, ExpectedRevision: first.Revision, Value: &nested}}})
			return err
		}},
		{"file", func() error {
			_, err := s.CreateCVFile(t.Context(), vehicle, application.VehicleCVFileInput{FileName: "stale.csv", OriginalName: "stale.csv", StoragePath: "uploads/stale.csv", DecoderProfile: "Legacy", WorkspaceID: workspace.ID, ExpectedWorkspaceRevision: workspace.BindingRevision})
			return err
		}},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			if err := scenario.write(); !errors.Is(err, application.ErrVehicleCVConflict) {
				t.Fatalf("stale binding accepted: %v", err)
			}
			if after := decoderSnapshots(t, backup); !reflect.DeepEqual(catalogBackupTableSnapshot(t, saved), catalogBackupTableSnapshot(t, after)) {
				t.Fatal("stale group caused a partial CV/history/source write")
			}
		})
	}
	assigned := decoderWorkspace(t, afterBinding, workspace.Profile)
	removed, err := s.RemoveDecoder(t.Context(), vehicle, installation, application.DecoderRemovalInput{ExpectedRevision: afterBinding.Revision}, "")
	if err != nil {
		t.Fatal(err)
	}
	if decoderWorkspace(t, removed, workspace.Profile).BindingRevision == assigned.BindingRevision {
		t.Fatal("removal did not invalidate the previous installed-state comparison")
	}
	input.ExpectedWorkspaceRevision = assigned.BindingRevision
	if _, err := s.UpdateCVValue(t.Context(), vehicle, first.ID, input); !errors.Is(err, application.ErrVehicleCVConflict) {
		t.Fatalf("removed installation accepted an old CV confirmation: %v", err)
	}
}

func TestDecoderConfigurationCVScopeRejectsMixedAliasesAndForeignWorkspace(t *testing.T) {
	db, backup, original := decoderConfigurationFixture(t)
	s := application.NewVehicleService(db)
	vehicle := original.Tables["vehicles"][0]["id"].(string)
	normal, err := s.CreateCVValue(t.Context(), vehicle, application.VehicleCVValueInput{CVNumber: 1, Value: 20, DecoderProfile: "Legacy"})
	if err != nil {
		t.Fatal(err)
	}
	current, err := s.DecoderConfigurationContext(t.Context(), vehicle)
	if err != nil {
		t.Fatal(err)
	}
	workspace := decoderWorkspace(t, current, "Legacy")
	empty := ""
	scoped := application.VehicleCVValueInput{CVNumber: 3, Value: 10, DecoderProfile: "Legacy", WorkspaceID: workspace.ID, ExpectedWorkspaceRevision: workspace.BindingRevision, ExpectedRevision: &empty}
	unscoped := application.VehicleCVValueInput{CVNumber: 3, Value: 11, DecoderProfile: "Legacy", ExpectedRevision: &empty}
	before := decoderSnapshots(t, backup)
	missingComparison := scoped
	missingComparison.ExpectedRevision = nil
	if _, err := s.CreateCVValue(t.Context(), vehicle, missingComparison); !errors.Is(err, application.ErrVehicleValidation) {
		t.Fatalf("scoped write without a row comparison accepted: %v", err)
	}
	if _, err := s.ImportCVValues(t.Context(), vehicle, application.VehicleCVImportInput{Values: []application.VehicleCVValueInput{scoped, unscoped}}); !errors.Is(err, application.ErrVehicleValidation) {
		t.Fatalf("duplicate import aliases accepted: %v", err)
	}
	scoped.ExpectedRevision, unscoped.ExpectedRevision = nil, nil
	if _, err := s.ApplyCVChanges(t.Context(), vehicle, application.VehicleCVChangesInput{Changes: []application.VehicleCVChange{{Value: &scoped}, {Value: &unscoped}}}); !errors.Is(err, application.ErrVehicleValidation) {
		t.Fatalf("duplicate draft aliases accepted: %v", err)
	}
	foreign, err := s.Create(t.Context(), application.CreateVehicleInput{Name: "Foreign", Manufacturer: "Test", Gauge: "H0", Category: "Lokomotive", Gattung: "Diesellok"}, "")
	if err != nil {
		t.Fatal(err)
	}
	wrong := scopedCV(*normal, workspace)
	if _, err := s.CreateCVValue(t.Context(), foreign.ID, wrong); !errors.Is(err, application.ErrVehicleValidation) {
		t.Fatalf("foreign scope accepted: %v", err)
	}
	after := decoderSnapshots(t, backup)
	for _, table := range []string{"vehicle_cv_values", "vehicle_cv_value_history", "vehicle_cv_files", "decoder_cv_workspaces", "decoder_cv_snapshots", "decoder_cv_snapshot_sources"} {
		if !reflect.DeepEqual(before.Tables[table], after.Tables[table]) {
			t.Fatalf("rejected aliases changed %s", table)
		}
	}
}
