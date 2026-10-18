package application_test

import (
	"errors"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestCVImportIndexIdentityRollbackHistoryAndBackup(t *testing.T) {
	dataDir := t.TempDir()
	db := backupTestDB(t, dataDir)
	s := application.NewVehicleService(db)
	v, err := s.Create(t.Context(), application.CreateVehicleInput{
		Manufacturer: "Test", Name: "Decoderwagen", Gauge: "TT", Category: "Personenwagen", Gattung: "Personenwagen",
	}, "test")
	if err != nil {
		t.Fatal(err)
	}
	high, zero, three := 16, 0, 3
	values := []application.VehicleCVValueInput{
		{CVNumber: 257, Value: 1, DecoderProfile: "Test"},
		{CVNumber: 257, Value: 2, DecoderProfile: "Test", IndexCV31: &high, IndexCV32: &zero},
		{CVNumber: 257, Value: 3, DecoderProfile: "Test", IndexCV31: &high, IndexCV32: &three},
	}
	if _, err := s.ImportCVValues(t.Context(), v.ID, application.VehicleCVImportInput{Values: values}); err != nil {
		t.Fatal(err)
	}
	changed := values[1]
	changed.Value = 44
	sourceLine := 400
	changed.SourceLine = &sourceLine
	changed.SourceExportedAt = "2026-10-17T17:41:59"
	foreign := values[2]
	foreign.SourceFileID = "another-vehicle-file"
	if _, err := s.ImportCVValues(t.Context(), v.ID, application.VehicleCVImportInput{Values: []application.VehicleCVValueInput{changed, foreign}}); !errors.Is(err, application.ErrVehicleValidation) {
		t.Fatalf("expected rejected source: %v", err)
	}
	stored, err := s.ListCVValues(t.Context(), v.ID)
	if err != nil || len(stored) != 3 || stored[1].Value != 2 || len(stored[1].History) != 0 || stored[1].SourceLine != nil {
		t.Fatalf("partial import after rollback: %+v %v", stored, err)
	}
	if _, err := s.ImportCVValues(t.Context(), v.ID, application.VehicleCVImportInput{Values: []application.VehicleCVValueInput{changed}}); err != nil {
		t.Fatal(err)
	}
	stored, err = s.ListCVValues(t.Context(), v.ID)
	if err != nil || stored[1].Value != 44 || len(stored[1].History) != 1 || stored[2].Value != 3 || stored[1].SourceLine == nil || *stored[1].SourceLine != sourceLine {
		t.Fatalf("wrong page or history: %+v %v", stored, err)
	}
	backup := application.NewBackupService(db, dataDir)
	doc, err := backup.Export(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if doc.Version != 22 {
		t.Fatalf("indexed data must reject old readers: %d", doc.Version)
	}
	if _, err := backup.Import(t.Context(), doc); err != nil {
		t.Fatal(err)
	}
	stored, err = s.ListCVValues(t.Context(), v.ID)
	if err != nil || len(stored) != 3 || stored[1].IndexCV32 == nil || *stored[1].IndexCV32 != 0 || len(stored[1].History) != 1 || stored[1].SourceExportedAt != changed.SourceExportedAt {
		t.Fatalf("lost indexed data on restore: %+v %v", stored, err)
	}
	// Legacy backups have no index columns and remain restorable as unindexed values.
	doc.Version = 20
	doc.Tables["vehicle_cv_values"] = doc.Tables["vehicle_cv_values"][:1]
	doc.Tables["vehicle_cv_value_history"] = nil
	delete(doc.Tables["vehicle_cv_values"][0], "index_cv31")
	delete(doc.Tables["vehicle_cv_values"][0], "index_cv32")
	if _, err := backup.Import(t.Context(), doc); err != nil {
		t.Fatal(err)
	}
	stored, err = s.ListCVValues(t.Context(), v.ID)
	if err != nil || len(stored) != 1 || stored[0].IndexCV31 != nil {
		t.Fatalf("legacy restore: %+v %v", stored, err)
	}
}

func TestCVImportRejectsInvalidIndexPairsAndDuplicates(t *testing.T) {
	s := application.NewVehicleService(testDB(t))
	high, low, invalid := 16, 0, 256
	for _, values := range [][]application.VehicleCVValueInput{
		{{CVNumber: 257, Value: 0, IndexCV31: &high}},
		{{CVNumber: 257, Value: 0, IndexCV31: &invalid, IndexCV32: &low}},
		{{CVNumber: 1, Value: 0, IndexCV31: &high, IndexCV32: &low}},
		{{CVNumber: 513, Value: 0, IndexCV31: &high, IndexCV32: &low}},
		{{CVNumber: 1, Value: 1}, {CVNumber: 1, Value: 2}},
	} {
		if _, err := s.ImportCVValues(t.Context(), "test", application.VehicleCVImportInput{Values: values}); !errors.Is(err, application.ErrVehicleValidation) {
			t.Fatalf("accepted invalid index or duplicate: %+v %v", values, err)
		}
	}
}
