package application_test

import (
	"errors"
	"railkeeper/backend/internal/application"
	"testing"
)

func TestCVProvenanceSurvivesEditsImportsAndBackup(t *testing.T) {
	dir := t.TempDir()
	db := backupTestDB(t, dir)
	s := application.NewVehicleService(db)
	v, err := s.Create(t.Context(), application.CreateVehicleInput{Manufacturer: "Test", Name: "CV", Gauge: "TT", Category: "Lokomotive", Gattung: "Diesellok"}, "test")
	if err != nil {
		t.Fatal(err)
	}
	line, high, low := 140, 16, 3
	input := application.VehicleCVValueInput{CVNumber: 257, Value: 3, DecoderProfile: "Test", IndexCV31: &high, IndexCV32: &low, SourceLine: &line, SourceExportedAt: "2026-10-17T17:41:59"}
	cv, err := s.CreateCVValue(t.Context(), v.ID, input)
	if err != nil {
		t.Fatal(err)
	}
	input.SourceLine = nil
	input.SourceExportedAt = ""
	input.Value = 4
	cv, err = s.UpdateCVValue(t.Context(), v.ID, cv.ID, input)
	if err != nil {
		t.Fatal(err)
	}
	if cv.SourceLine == nil || *cv.SourceLine != line || cv.SourceExportedAt != "2026-10-17T17:41:59" || len(cv.History) != 1 {
		t.Fatalf("edit lost provenance: %+v", cv)
	}
	if _, err := s.ImportCVValues(t.Context(), v.ID, application.VehicleCVImportInput{Values: []application.VehicleCVValueInput{input}}); err != nil {
		t.Fatal(err)
	}
	values, err := s.ListCVValues(t.Context(), v.ID)
	if err != nil || len(values) != 1 || values[0].SourceLine == nil || *values[0].SourceLine != line || values[0].SourceExportedAt != "2026-10-17T17:41:59" {
		t.Fatalf("import lost provenance: %+v %v", values, err)
	}
	backup := application.NewBackupService(db, dir)
	doc, err := backup.Export(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if doc.Version < 22 {
		t.Fatal("metadata needs version 22")
	}
	if _, err := backup.Import(t.Context(), doc); err != nil {
		t.Fatal(err)
	}
	cv, err = s.GetCVValue(t.Context(), v.ID, cv.ID)
	if err != nil || cv.SourceLine == nil || *cv.SourceLine != line || cv.SourceExportedAt != "2026-10-17T17:41:59" {
		t.Fatalf("restore lost provenance: %+v %v", cv, err)
	}
	doc.Tables["vehicle_cv_values"][0]["source_exported_at"] = "2026-02-30T17:41:59"
	if _, err := backup.Import(t.Context(), doc); !errors.Is(err, application.ErrBackupInvalid) {
		t.Fatalf("invalid date accepted: %v", err)
	}
	delete(doc.Tables["vehicle_cv_values"][0], "source_line")
	delete(doc.Tables["vehicle_cv_values"][0], "source_exported_at")
	doc.Version = 21
	if _, err := backup.Import(t.Context(), doc); err != nil {
		t.Fatal(err)
	}
	cv, err = s.GetCVValue(t.Context(), v.ID, cv.ID)
	if err != nil || cv.SourceLine != nil || cv.SourceExportedAt != "" {
		t.Fatalf("legacy restore guessed metadata: %+v %v", cv, err)
	}
}

func TestCVProvenanceValidationRejectsInvalidMetadata(t *testing.T) {
	s := application.NewVehicleService(testDB(t))
	zero, huge := 0, 10000001
	for _, input := range []application.VehicleCVValueInput{
		{CVNumber: 1, Value: 3, SourceLine: &zero}, {CVNumber: 1, Value: 3, SourceLine: &huge},
		{CVNumber: 1, Value: 3, SourceExportedAt: "2026-02-30T17:41:59"},
		{CVNumber: 1, Value: 3, SourceExportedAt: "2026-10-17T17:41:59Z"},
	} {
		if _, err := s.ImportCVValues(t.Context(), "missing", application.VehicleCVImportInput{Values: []application.VehicleCVValueInput{{CVNumber: 2, Value: 4}, input}}); !errors.Is(err, application.ErrVehicleValidation) {
			t.Fatalf("accepted invalid source: %+v %v", input, err)
		}
	}
}
