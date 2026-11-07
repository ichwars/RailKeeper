package application_test

import (
	"reflect"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestDecoderConfigurationWorkingWritesPreserveAdoptionAndExplicitHardware(t *testing.T) {
	db, backup, root, vehicle := legacyDecoderCaptureFixture(t)
	if err := application.BackfillLegacyDecoderCV(t.Context(), db, root); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE decoder_cv_workspaces SET installation_id='installation-a', updated_at='2000-01-01T00:00:00Z'`); err != nil {
		t.Fatal(err)
	}
	before := decoderSnapshots(t, backup)
	s := application.NewVehicleService(db)
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
	moved, err := s.UpdateCVValue(t.Context(), vehicle, first.ID, application.VehicleCVValueInput{
		CVNumber: 1, Value: 77, DecoderProfile: "New working group", ExpectedRevision: &first.Revision,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.ImportCVValues(t.Context(), vehicle, application.VehicleCVImportInput{
		Values: []application.VehicleCVValueInput{{CVNumber: 2, Value: 12, DecoderProfile: "New working group"}},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.DeleteCVValue(t.Context(), vehicle, moved.ID, moved.Revision); err != nil {
		t.Fatal(err)
	}
	if err := application.BackfillLegacyDecoderCV(t.Context(), db, root); err != nil {
		t.Fatal(err)
	}
	after := decoderSnapshots(t, backup)
	if len(after.Tables["decoder_cv_workspaces"]) != 2 {
		t.Fatal("manual writes lost or duplicated working groups")
	}
	for _, row := range after.Tables["decoder_cv_workspaces"] {
		if row["profile_text"] == " Legacy " {
			old := before.Tables["decoder_cv_workspaces"][0]
			if row["id"] != old["id"] || row["origin"] != "legacy" || row["installation_id"] != "installation-a" ||
				row["created_at"] != old["created_at"] || row["updated_at"] == old["updated_at"] {
				t.Fatal("writing changed raw group identity, adoption origin or confirmed hardware")
			}
		} else if row["profile_text"] != "New working group" || row["origin"] != "manual" || row["installation_id"] != nil {
			t.Fatal("new working group inferred hardware or invented legacy provenance")
		}
	}
	for _, table := range []string{"decoder_cv_snapshots", "decoder_cv_snapshot_sources", "decoder_instances", "decoder_installations"} {
		if !reflect.DeepEqual(before.Tables[table], after.Tables[table]) {
			t.Fatalf("ordinary CV writes changed frozen evidence or hardware in %s", table)
		}
	}
	if _, err := backup.Import(t.Context(), after); err != nil {
		t.Fatal(err)
	}
	if restored := decoderSnapshots(t, backup); !reflect.DeepEqual(catalogBackupTableSnapshot(t, after), catalogBackupTableSnapshot(t, restored)) {
		t.Fatal("restore changed updated workspaces or old evidence")
	}
}

func TestDecoderConfigurationEveryCVWriterRollsBackWithWorkspaceFailure(t *testing.T) {
	db, backup, _, vehicle := legacyDecoderCaptureFixture(t)
	s := application.NewVehicleService(db)
	value, err := s.CreateCVValue(t.Context(), vehicle, application.VehicleCVValueInput{CVNumber: 10, Value: 3, DecoderProfile: "Manual"})
	if err != nil {
		t.Fatal(err)
	}
	fileInput := application.VehicleCVFileInput{FileName: "source.csv", OriginalName: "source.csv",
		StoragePath: "uploads/source.csv", DecoderProfile: "Manual"}
	file, err := s.CreateCVFile(t.Context(), vehicle, fileInput)
	if err != nil {
		t.Fatal(err)
	}
	for _, statement := range []string{
		`CREATE TRIGGER fail_workspace_insert BEFORE INSERT ON decoder_cv_workspaces
WHEN NOT EXISTS(SELECT 1 FROM decoder_cv_workspaces WHERE vehicle_id=NEW.vehicle_id AND profile_text=NEW.profile_text)
BEGIN SELECT RAISE(ABORT,'workspace failure'); END`,
		`CREATE TRIGGER fail_workspace_update BEFORE UPDATE ON decoder_cv_workspaces BEGIN SELECT RAISE(ABORT,'workspace failure'); END`,
	} {
		if _, err := db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	before := decoderSnapshots(t, backup)
	for _, scenario := range []struct {
		name  string
		write func() error
	}{
		{"legacy create", func() error {
			_, err := s.CreateCVValue(t.Context(), vehicle, application.VehicleCVValueInput{CVNumber: 11, Value: 4, DecoderProfile: "New"})
			return err
		}},
		{"legacy update and history", func() error {
			_, err := s.UpdateCVValue(t.Context(), vehicle, value.ID, application.VehicleCVValueInput{CVNumber: 10, Value: 5, DecoderProfile: "Manual"})
			return err
		}},
		{"legacy delete", func() error { _, err := s.DeleteCVValue(t.Context(), vehicle, value.ID); return err }},
		{"manual draft", func() error {
			_, err := s.ApplyCVChanges(t.Context(), vehicle, application.VehicleCVChangesInput{Changes: []application.VehicleCVChange{{ID: value.ID, ExpectedRevision: value.Revision, Delete: true}}})
			return err
		}},
		{"file and station import", func() error {
			_, err := s.ImportCVValues(t.Context(), vehicle, application.VehicleCVImportInput{Values: []application.VehicleCVValueInput{{CVNumber: 10, Value: 6, DecoderProfile: "Manual"}, {CVNumber: 12, Value: 7, DecoderProfile: "New"}}})
			return err
		}},
		{"source create", func() error { _, err := s.CreateCVFile(t.Context(), vehicle, fileInput); return err }},
		{"source delete", func() error { _, err := s.DeleteCVFile(t.Context(), vehicle, file.ID); return err }},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			if err := scenario.write(); err == nil {
				t.Fatal("workspace failure did not reject writer")
			}
			after := decoderSnapshots(t, backup)
			if !reflect.DeepEqual(catalogBackupTableSnapshot(t, before), catalogBackupTableSnapshot(t, after)) || !reflect.DeepEqual(before.Files, after.Files) {
				t.Fatal("CV, history, source or group survived a failed transaction")
			}
		})
	}
}
