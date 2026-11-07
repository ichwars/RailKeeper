package application_test

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestDecoderConfigurationCheckpointsAssignmentAndArchivePreserveCompleteEvidence(t *testing.T) {
	db, backup, original := decoderConfigurationFixture(t)
	s := application.NewVehicleService(db)
	vehicle := original.Tables["vehicles"][0]["id"].(string)
	current, err := s.DecoderConfigurationContext(t.Context(), vehicle)
	if err != nil {
		t.Fatal(err)
	}
	checkpoint := application.DecoderCVCheckpointInput{ID: strings.Repeat("a", 32), ExpectedRevision: current.Revision, Note: "Explicit partial working state"}
	current, err = s.CaptureDecoderCVCheckpoint(t.Context(), vehicle, "workspace-a", checkpoint, "", "")
	if err != nil {
		t.Fatal(err)
	}
	recorded, err := s.DecoderCVSnapshot(t.Context(), checkpoint.ID)
	if err != nil {
		t.Fatal(err)
	}
	values, err := s.ListCVValues(t.Context(), vehicle)
	if err != nil {
		t.Fatal(err)
	}
	for index := range values {
		values[index].Revision = ""
	}
	// Compare complete public values after their established JSON normalization.
	expected, _ := json.Marshal(application.DecoderCVSnapshotPayload{Values: values})
	var expectedPayload application.DecoderCVSnapshotPayload
	if err := json.Unmarshal(expected, &expectedPayload); err != nil {
		t.Fatal(err)
	}
	byID := func(payload application.DecoderCVSnapshotPayload) map[string]application.VehicleCVValue {
		result := map[string]application.VehicleCVValue{}
		for _, value := range payload.Values {
			result[value.ID] = value
		}
		return result
	}
	if !reflect.DeepEqual(byID(expectedPayload), byID(recorded.Payload)) || recorded.Hardware != nil || recorded.Scope != "partial_working_state" || recorded.Profile != " Legacy " {
		t.Fatal("checkpoint changed CV identities, unknown values, index pages, provenance or histories")
	}
	if len(recorded.Sources) != 1 || recorded.Sources[0].ContentState != "stored" {
		t.Fatal("archived original was not retained")
	}
	sourceID := recorded.Sources[0].SourceFileID
	file, bytes, err := s.DecoderCVSnapshotSourceContent(t.Context(), checkpoint.ID, sourceID)
	if err != nil || string(bytes) != "CV;Value\n1;3\n" || file.ID != sourceID {
		t.Fatalf("original bytes changed: %v", err)
	}
	if _, err := s.CaptureDecoderCVCheckpoint(t.Context(), vehicle, "workspace-a", checkpoint, "", ""); !errors.Is(err, application.ErrDecoderConflict) {
		t.Fatalf("replayed checkpoint: %v", err)
	}
	checkpoint.ExpectedRevision = current.Revision
	if _, err := s.CaptureDecoderCVCheckpoint(t.Context(), vehicle, "workspace-a", checkpoint, "", ""); !errors.Is(err, application.ErrDecoderConflict) {
		t.Fatalf("reused checkpoint identity: %v", err)
	}
	current, err = s.CreateDecoderCVWorkspace(t.Context(), vehicle, application.DecoderCVWorkspaceInput{ID: strings.Repeat("b", 32), ExpectedRevision: current.Revision, Profile: " Fresh working group "}, "")
	if err != nil {
		t.Fatal(err)
	}
	for _, workspace := range current.Workspaces {
		if workspace.ID == strings.Repeat("b", 32) && (workspace.Profile != "Fresh working group" || workspace.ValueCount != 0 || workspace.InstallationID != nil) {
			t.Fatal("empty group invented values or hardware")
		}
	}
	installation := "installation-a"
	current, err = s.AssignDecoderCVWorkspace(t.Context(), vehicle, "workspace-a", application.DecoderCVAssignmentInput{ExpectedRevision: current.Revision, InstallationID: &installation, SnapshotID: strings.Repeat("c", 32)}, "", "")
	if err != nil {
		t.Fatal(err)
	}
	previous, err := s.DecoderCVSnapshot(t.Context(), strings.Repeat("c", 32))
	if err != nil || previous.Hardware != nil {
		t.Fatalf("assignment did not retain prior unassigned state: %v", err)
	}
	checkpoint.ID, checkpoint.ExpectedRevision = strings.Repeat("d", 32), current.Revision
	current, err = s.CaptureDecoderCVCheckpoint(t.Context(), vehicle, "workspace-a", checkpoint, "", "")
	if err != nil {
		t.Fatal(err)
	}
	assigned, err := s.DecoderCVSnapshot(t.Context(), checkpoint.ID)
	if err != nil || assigned.Hardware == nil || assigned.Hardware.DecoderID != "decoder-a" || assigned.Hardware.InstallationState != "installed" || assigned.Hardware.InstalledAt != nil {
		t.Fatalf("confirmed hardware not frozen accurately: %v", err)
	}
	current, err = s.UpdateDecoderInstance(t.Context(), vehicle, application.DecoderInstanceInput{ID: "decoder-a", ExpectedRevision: current.Revision, Label: "Corrected hardware", Model: "Corrected model"}, "")
	if err != nil {
		t.Fatal(err)
	}
	current, err = s.RemoveDecoder(t.Context(), vehicle, installation, application.DecoderRemovalInput{ExpectedRevision: current.Revision}, "")
	if err != nil {
		t.Fatal(err)
	}
	checkpoint.ID, checkpoint.ExpectedRevision = strings.Repeat("e", 32), current.Revision
	if _, err := s.CaptureDecoderCVCheckpoint(t.Context(), vehicle, "workspace-a", checkpoint, "", ""); err != nil {
		t.Fatal(err)
	}
	removed, err := s.DecoderCVSnapshot(t.Context(), checkpoint.ID)
	if err != nil || removed.Hardware == nil || removed.Hardware.InstallationState != "removed" || removed.Hardware.RemovedAt != nil || removed.Hardware.Model != "Corrected model" {
		t.Fatalf("removed state meaning changed: %v", err)
	}
	stillAssigned, err := s.DecoderCVSnapshot(t.Context(), assigned.ID)
	if err != nil || !reflect.DeepEqual(stillAssigned, assigned) {
		t.Fatal("hardware correction or removal rewrote earlier evidence")
	}
	after := decoderSnapshots(t, backup)
	if _, err := backup.Import(t.Context(), after); err != nil {
		t.Fatal(err)
	}
	if restored := decoderSnapshots(t, backup); !reflect.DeepEqual(catalogBackupTableSnapshot(t, after), catalogBackupTableSnapshot(t, restored)) {
		t.Fatal("workspace assignment or checkpoints changed through restore")
	}
	if err := s.Delete(t.Context(), vehicle, ""); err != nil {
		t.Fatal(err)
	}
	if historical, err := s.DecoderCVSnapshot(t.Context(), assigned.ID); err != nil || !reflect.DeepEqual(historical, assigned) {
		t.Fatal("historical snapshot requires a live vehicle")
	}
	for _, filter := range []string{"", vehicle} {
		items, err := s.ListDecoderCVSnapshots(t.Context(), filter)
		if err != nil || len(items) != 6 {
			t.Fatalf("archive lost snapshots for deleted inventory: %d / %v", len(items), err)
		}
	}
	if _, data, err := s.DecoderCVSnapshotSourceContent(t.Context(), assigned.ID, sourceID); err != nil || string(data) != "CV;Value\n1;3\n" {
		t.Fatalf("vehicle deletion lost original bytes: %v", err)
	}
}

func TestDecoderConfigurationCheckpointAndAssignmentRollBackCapturedBlobAndAudit(t *testing.T) {
	db, backup, root, vehicle := legacyDecoderCaptureFixture(t)
	s := application.NewVehicleService(db)
	if err := os.MkdirAll(filepath.Join(root, "uploads"), 0700); err != nil {
		t.Fatal(err)
	}
	content := []byte("New filesystem original, rolled back with checkpoint")
	if err := os.WriteFile(filepath.Join(root, "uploads", "captured.csv"), content, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateCVFile(t.Context(), vehicle, application.VehicleCVFileInput{FileName: "captured.csv", OriginalName: "captured.csv", DecoderProfile: "New data", StoragePath: "uploads/captured.csv"}); err != nil {
		t.Fatal(err)
	}
	current, err := s.DecoderConfigurationContext(t.Context(), vehicle)
	if err != nil {
		t.Fatal(err)
	}
	workspace := ""
	for _, item := range current.Workspaces {
		if item.Profile == "New data" {
			workspace = item.ID
		}
	}
	if workspace == "" {
		t.Fatal("source writer did not create workspace")
	}
	if _, err := db.Exec(`CREATE TRIGGER fail_checkpoint_audit BEFORE INSERT ON audit_logs WHEN NEW.action LIKE 'Decoder%' BEGIN SELECT RAISE(ABORT,'checkpoint audit failure'); END`); err != nil {
		t.Fatal(err)
	}
	before := decoderSnapshots(t, backup)
	installation := "installation-a"
	for _, scenario := range []struct {
		name  string
		write func() error
	}{
		{"capture", func() error {
			_, err := s.CaptureDecoderCVCheckpoint(t.Context(), vehicle, workspace, application.DecoderCVCheckpointInput{ID: strings.Repeat("f", 32), ExpectedRevision: current.Revision}, root, "")
			return err
		}},
		{"assignment", func() error {
			_, err := s.AssignDecoderCVWorkspace(t.Context(), vehicle, workspace, application.DecoderCVAssignmentInput{ExpectedRevision: current.Revision, InstallationID: &installation, SnapshotID: strings.Repeat("f", 32)}, root, "")
			return err
		}},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			if err := scenario.write(); err == nil {
				t.Fatal("audit failure was ignored")
			}
			if after := decoderSnapshots(t, backup); !reflect.DeepEqual(catalogBackupTableSnapshot(t, before), catalogBackupTableSnapshot(t, after)) || !reflect.DeepEqual(before.Files, after.Files) {
				t.Fatal("failed command left captured blob, snapshot, assignment or audit")
			}
		})
	}
}

func TestDecoderConfigurationArchiveReportsMissingOriginalWithoutInventingContent(t *testing.T) {
	db, _, root, vehicle := legacyDecoderCaptureFixture(t)
	if err := application.BackfillLegacyDecoderCV(t.Context(), db, root); err != nil {
		t.Fatal(err)
	}
	s := application.NewVehicleService(db)
	items, err := s.ListDecoderCVSnapshots(t.Context(), vehicle)
	if err != nil || len(items) != 1 {
		t.Fatalf("missing legacy archive: %v", err)
	}
	item, err := s.DecoderCVSnapshot(t.Context(), items[0].ID)
	if err != nil || len(item.Sources) != 1 || item.Sources[0].ContentState != "unavailable" {
		t.Fatalf("missing source was not explicit: %v", err)
	}
	if _, _, err := s.DecoderCVSnapshotSourceContent(t.Context(), item.ID, item.Sources[0].SourceFileID); !errors.Is(err, application.ErrDecoderSourceUnavailable) {
		t.Fatalf("missing source fabricated: %v", err)
	}
	if _, err := s.DecoderCVSnapshot(t.Context(), "missing"); !errors.Is(err, application.ErrDecoderNotFound) {
		t.Fatalf("missing snapshot: %v", err)
	}
}

func TestDecoderConfigurationWorkspaceRejectsForeignAssignmentsAndKeepsPriorAssociationWhenCleared(t *testing.T) {
	db, backup, original := decoderConfigurationFixture(t)
	s := application.NewVehicleService(db)
	vehicle := original.Tables["vehicles"][0]["id"].(string)
	other, err := s.Create(t.Context(), application.CreateVehicleInput{Name: "Other vehicle", Manufacturer: "Test", Gauge: "H0", Category: "Lokomotive", Gattung: "Diesellok"}, "")
	if err != nil {
		t.Fatal(err)
	}
	primaryContext, err := s.DecoderConfigurationContext(t.Context(), vehicle)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateDecoderCVWorkspace(t.Context(), vehicle, application.DecoderCVWorkspaceInput{ID: strings.Repeat("6", 32), ExpectedRevision: primaryContext.Revision, Profile: "Other owner's group"}, ""); err != nil {
		t.Fatal(err)
	}
	otherContext, err := s.DecoderConfigurationContext(t.Context(), other.ID)
	if err != nil {
		t.Fatal(err)
	}
	workspaceID := strings.Repeat("1", 32)
	otherContext, err = s.CreateDecoderCVWorkspace(t.Context(), other.ID, application.DecoderCVWorkspaceInput{ID: workspaceID, ExpectedRevision: otherContext.Revision, Profile: "Other working group"}, "")
	if err != nil {
		t.Fatal(err)
	}
	before := decoderSnapshots(t, backup)
	installation := "installation-a"
	for _, scenario := range []struct {
		name  string
		want  error
		write func() error
	}{
		{"foreign installation", application.ErrDecoderValidation, func() error {
			_, err := s.AssignDecoderCVWorkspace(t.Context(), other.ID, workspaceID, application.DecoderCVAssignmentInput{ExpectedRevision: otherContext.Revision, InstallationID: &installation, SnapshotID: strings.Repeat("2", 32)}, "", "")
			return err
		}},
		{"duplicate label", application.ErrDecoderConflict, func() error {
			_, err := s.CreateDecoderCVWorkspace(t.Context(), other.ID, application.DecoderCVWorkspaceInput{ID: strings.Repeat("3", 32), ExpectedRevision: otherContext.Revision, Profile: "Other working group"}, "")
			return err
		}},
		{"identity used by another vehicle", application.ErrDecoderConflict, func() error {
			_, err := s.CreateDecoderCVWorkspace(t.Context(), other.ID, application.DecoderCVWorkspaceInput{ID: strings.Repeat("6", 32), ExpectedRevision: otherContext.Revision, Profile: "Distinct"}, "")
			return err
		}},
		{"missing workspace", application.ErrDecoderNotFound, func() error {
			_, err := s.CaptureDecoderCVCheckpoint(t.Context(), other.ID, "missing", application.DecoderCVCheckpointInput{ID: strings.Repeat("4", 32), ExpectedRevision: otherContext.Revision}, "", "")
			return err
		}},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			if err := scenario.write(); !errors.Is(err, scenario.want) {
				t.Fatalf("expected %v: %v", scenario.want, err)
			}
			if after := decoderSnapshots(t, backup); !reflect.DeepEqual(catalogBackupTableSnapshot(t, before), catalogBackupTableSnapshot(t, after)) {
				t.Fatal("rejected workspace command changed data or evidence")
			}
		})
	}
	current, err := s.DecoderConfigurationContext(t.Context(), vehicle)
	if err != nil {
		t.Fatal(err)
	}
	current, err = s.AssignDecoderCVWorkspace(t.Context(), vehicle, "workspace-b", application.DecoderCVAssignmentInput{ExpectedRevision: current.Revision, SnapshotID: strings.Repeat("5", 32)}, "", "")
	if err != nil {
		t.Fatal(err)
	}
	for _, workspace := range current.Workspaces {
		if workspace.ID == "workspace-b" && workspace.InstallationID != nil {
			t.Fatal("explicit clear did not clear current association")
		}
	}
	previous, err := s.DecoderCVSnapshot(t.Context(), strings.Repeat("5", 32))
	if err != nil || previous.Hardware == nil || previous.Hardware.DecoderID != "decoder-b" {
		t.Fatal("clearing association lost previous hardware evidence")
	}
}
