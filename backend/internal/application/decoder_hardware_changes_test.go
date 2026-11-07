package application_test

import (
	"errors"
	"reflect"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestDecoderConfigurationHardwareCommandsKeepEpisodesDatesAndOldEvidence(t *testing.T) {
	db, backup, original := decoderConfigurationFixture(t)
	s := application.NewVehicleService(db)
	vehicle := original.Tables["vehicles"][0]["id"].(string)
	current, err := s.DecoderConfigurationContext(t.Context(), vehicle)
	if err != nil {
		t.Fatal(err)
	}
	if len(current.Instances) != 2 || len(current.Installations) != 2 || len(current.Workspaces) != 2 || len(current.Snapshots) != 2 {
		t.Fatalf("incomplete context: %#v", current)
	}
	if read := decoderSnapshots(t, backup); !reflect.DeepEqual(catalogBackupTableSnapshot(t, original), catalogBackupTableSnapshot(t, read)) {
		t.Fatal("reading decoder context changed inventory or evidence")
	}
	instance := application.DecoderInstanceInput{ID: "11111111111111111111111111111111", ExpectedRevision: current.Revision,
		Label: "Additional physical decoder", Manufacturer: "Test", Model: "Same model"}
	current, err = s.CreateDecoderInstance(t.Context(), vehicle, instance, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(current.Instances) != 3 || len(current.Installations) != 2 {
		t.Fatal("matching model was merged or installed implicitly")
	}
	if _, err := s.CreateDecoderInstance(t.Context(), vehicle, instance, ""); !errors.Is(err, application.ErrDecoderConflict) {
		t.Fatalf("lost-response replay: %v", err)
	}
	installation := application.DecoderInstallationInput{ID: "22222222222222222222222222222222", ExpectedRevision: current.Revision,
		DecoderID: instance.ID, Notes: "Known installation evidence"}
	current, err = s.InstallDecoder(t.Context(), vehicle, installation, "")
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range current.Installations {
		if item.ID == installation.ID && (item.InstalledAt != nil || item.RemovedAt != nil || item.State != "installed") {
			t.Fatal("recording time invented installation dates")
		}
	}
	installation.ID = "33333333333333333333333333333333"
	installation.ExpectedRevision = current.Revision
	if _, err := s.InstallDecoder(t.Context(), vehicle, installation, ""); !errors.Is(err, application.ErrDecoderConflict) {
		t.Fatalf("double installation: %v", err)
	}
	installation.ID = "22222222222222222222222222222222"
	date := "2026-10-20"
	current, err = s.CorrectDecoderInstallationDates(t.Context(), vehicle, installation.ID, application.DecoderInstallationDatesInput{
		ExpectedRevision: current.Revision, InstalledAt: &date}, "")
	if err != nil {
		t.Fatal(err)
	}
	invalid := "2026-02-30"
	if _, err := s.RemoveDecoder(t.Context(), vehicle, installation.ID, application.DecoderRemovalInput{ExpectedRevision: current.Revision, RemovedAt: &invalid}, ""); !errors.Is(err, application.ErrDecoderValidation) {
		t.Fatalf("invalid date: %v", err)
	}
	tooEarly := "2026-10-19"
	if _, err := s.RemoveDecoder(t.Context(), vehicle, installation.ID, application.DecoderRemovalInput{ExpectedRevision: current.Revision, RemovedAt: &tooEarly}, ""); !errors.Is(err, application.ErrDecoderValidation) {
		t.Fatalf("negative period: %v", err)
	}
	current, err = s.RemoveDecoder(t.Context(), vehicle, installation.ID, application.DecoderRemovalInput{ExpectedRevision: current.Revision, Note: "Removal date unknown"}, "")
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range current.Installations {
		if item.ID == installation.ID && (item.State != "removed" || item.RemovedAt != nil || item.Notes != "Known installation evidence\nRemoval date unknown") {
			t.Fatal("removal lost notes or invented a date")
		}
	}
	current, err = s.UpdateDecoderInstance(t.Context(), vehicle, application.DecoderInstanceInput{ID: "decoder-b", ExpectedRevision: current.Revision,
		Label: "Corrected description", Model: "Corrected model"}, "")
	if err != nil {
		t.Fatal(err)
	}
	other, err := s.Create(t.Context(), application.CreateVehicleInput{Name: "New vehicle", Manufacturer: "Test", Gauge: "H0", Category: "Lokomotive", Gattung: "Diesellok"}, "")
	if err != nil {
		t.Fatal(err)
	}
	otherContext, err := s.DecoderConfigurationContext(t.Context(), other.ID)
	if err != nil {
		t.Fatal(err)
	}
	installation.ID = "44444444444444444444444444444444"
	installation.ExpectedRevision = otherContext.Revision
	if _, err := s.InstallDecoder(t.Context(), other.ID, installation, ""); err != nil {
		t.Fatal(err)
	}
	after := decoderSnapshots(t, backup)
	for _, table := range []string{"vehicle_cv_values", "vehicle_cv_value_history", "decoder_cv_workspaces", "decoder_cv_snapshots", "decoder_cv_snapshot_sources"} {
		if !reflect.DeepEqual(original.Tables[table], after.Tables[table]) {
			t.Fatalf("hardware commands rewrote existing %s", table)
		}
	}
	if _, err := backup.Import(t.Context(), after); err != nil {
		t.Fatal(err)
	}
	if restored := decoderSnapshots(t, backup); !reflect.DeepEqual(catalogBackupTableSnapshot(t, after), catalogBackupTableSnapshot(t, restored)) {
		t.Fatal("decoder commands did not survive restore")
	}
	if err := s.Delete(t.Context(), vehicle, ""); err != nil {
		t.Fatal(err)
	}
	deleted := decoderSnapshots(t, backup)
	for _, table := range []string{"decoder_instances", "decoder_cv_snapshots", "decoder_cv_snapshot_sources"} {
		if !reflect.DeepEqual(after.Tables[table], deleted.Tables[table]) {
			t.Fatalf("vehicle deletion lost historical %s", table)
		}
	}
}

func TestDecoderConfigurationContextDetectsSameClockCVHistoryAndPrivateSourceChanges(t *testing.T) {
	db, backup, original := decoderConfigurationFixture(t)
	s := application.NewVehicleService(db)
	vehicle := original.Tables["vehicles"][0]["id"].(string)
	file, err := s.CreateCVFile(t.Context(), vehicle, application.VehicleCVFileInput{FileName: "working.csv", OriginalName: "working.csv", StoragePath: "uploads/private-path.csv", DecoderProfile: " Legacy "})
	if err != nil {
		t.Fatal(err)
	}
	for _, statement := range []string{
		`UPDATE vehicle_cv_values SET description='changed without clock' WHERE cv_number=1`,
		`UPDATE vehicle_cv_value_history SET old_value=old_value+1`,
		`UPDATE vehicle_cv_files SET storage_path='uploads/replaced-source.csv'`,
	} {
		before, err := s.DecoderConfigurationContext(t.Context(), vehicle)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(statement); err != nil {
			t.Fatal(err)
		}
		after, err := s.DecoderConfigurationContext(t.Context(), vehicle)
		if err != nil || before.Revision == after.Revision {
			t.Fatalf("stale evidence was not detected: %s / %v", statement, err)
		}
		unchanged := decoderSnapshots(t, backup)
		if _, err := s.CreateDecoderInstance(t.Context(), vehicle, application.DecoderInstanceInput{ID: "55555555555555555555555555555555", ExpectedRevision: before.Revision, Label: "Stale"}, ""); !errors.Is(err, application.ErrDecoderConflict) {
			t.Fatalf("stale confirmation: %v", err)
		}
		if rejected := decoderSnapshots(t, backup); !reflect.DeepEqual(catalogBackupTableSnapshot(t, unchanged), catalogBackupTableSnapshot(t, rejected)) {
			t.Fatal("stale command changed data")
		}
	}
	if _, err := s.DeleteCVFile(t.Context(), vehicle, file.ID); err != nil {
		t.Fatal(err)
	}
}

func TestDecoderConfigurationHardwareAndAuditCommitTogether(t *testing.T) {
	db, backup, original := decoderConfigurationFixture(t)
	s := application.NewVehicleService(db)
	vehicle := original.Tables["vehicles"][0]["id"].(string)
	before, err := s.DecoderConfigurationContext(t.Context(), vehicle)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TRIGGER fail_decoder_audit BEFORE INSERT ON audit_logs WHEN NEW.action LIKE 'Decoder%' BEGIN SELECT RAISE(ABORT,'decoder audit failure'); END`); err != nil {
		t.Fatal(err)
	}
	for _, scenario := range []struct {
		name  string
		write func() error
	}{
		{"physical instance", func() error {
			_, err := s.CreateDecoderInstance(t.Context(), vehicle, application.DecoderInstanceInput{ID: "66666666666666666666666666666666", ExpectedRevision: before.Revision, Label: "Atomic"}, "")
			return err
		}},
		{"removal", func() error {
			_, err := s.RemoveDecoder(t.Context(), vehicle, "installation-a", application.DecoderRemovalInput{ExpectedRevision: before.Revision}, "")
			return err
		}},
		{"date correction", func() error {
			date := "2026-10-20"
			_, err := s.CorrectDecoderInstallationDates(t.Context(), vehicle, "installation-a", application.DecoderInstallationDatesInput{ExpectedRevision: before.Revision, InstalledAt: &date}, "")
			return err
		}},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			if err := scenario.write(); err == nil {
				t.Fatal("audit failure accepted")
			}
			if after := decoderSnapshots(t, backup); !reflect.DeepEqual(catalogBackupTableSnapshot(t, original), catalogBackupTableSnapshot(t, after)) {
				t.Fatal("failed audit left partial hardware change")
			}
		})
	}
}
