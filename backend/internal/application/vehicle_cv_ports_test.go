package application_test

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"

	"railkeeper/backend/internal/application"
)

func TestCVDraftHistoryFailureRestoresDeletedRowsAndPermitsSameRevisionRetry(t *testing.T) {
	db := backupTestDB(t, t.TempDir())
	db.SetMaxOpenConns(1)
	s := application.NewVehicleService(db)
	v, err := s.Create(t.Context(), application.CreateVehicleInput{
		Manufacturer: "Test", Name: "CV transaction", Gauge: "TT", Category: "Lokomotive", Gattung: "Diesellok",
	}, "test")
	if err != nil {
		t.Fatal(err)
	}
	file, err := s.CreateCVFile(t.Context(), v.ID, application.VehicleCVFileInput{
		FileName: "original.txt", OriginalName: "original.txt", StoragePath: "cv/original.txt", DecoderProfile: "A",
	})
	if err != nil {
		t.Fatal(err)
	}
	line, zero := 73, 0
	first, err := s.CreateCVValue(t.Context(), v.ID, application.VehicleCVValueInput{
		CVNumber: 257, Value: 1, DecoderProfile: "A", SourceFileID: file.ID,
		SourceLine: &line, SourceExportedAt: "2026-10-17T17:41:59",
	})
	if err != nil {
		t.Fatal(err)
	}
	indexed, err := s.CreateCVValue(t.Context(), v.ID, application.VehicleCVValueInput{
		CVNumber: 257, Value: 2, DecoderProfile: "A", IndexCV31: &zero, IndexCV32: &zero,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TRIGGER fail_cv_history BEFORE INSERT ON vehicle_cv_value_history
BEGIN SELECT RAISE(FAIL,'late CV history failure'); END`); err != nil {
		t.Fatal(err)
	}
	before := databaseRowFingerprints(t, db)
	input := application.VehicleCVChangesInput{Changes: []application.VehicleCVChange{
		{ID: indexed.ID, ExpectedRevision: indexed.Revision, Delete: true},
		{ID: first.ID, ExpectedRevision: first.Revision, Value: &application.VehicleCVValueInput{
			CVNumber: 257, Value: 4, DecoderProfile: "B", SourceFileID: file.ID,
		}},
	}}
	count, err := s.ApplyCVChanges(t.Context(), v.ID, input)
	if count != 0 || err == nil || !strings.Contains(err.Error(), "late CV history failure") {
		t.Fatal("late history failure was acknowledged or did not execute", count, err)
	}
	if after := databaseRowFingerprints(t, db); !reflect.DeepEqual(before, after) {
		t.Fatal("failed batch changed values, history, source references, workspaces or other tables")
	}
	if _, err := db.Exec(`DROP TRIGGER fail_cv_history`); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	if count, err := s.ApplyCVChanges(ctx, v.ID, input); err != nil || count != 2 {
		t.Fatal("same-revision retry failed", count, err)
	}
	values, err := s.ListCVValues(ctx, v.ID)
	if err != nil || len(values) != 1 {
		t.Fatal("unexpected values after retry", len(values), err)
	}
	value := values[0]
	if value.ID != first.ID || value.CreatedAt != first.CreatedAt || value.DecoderProfile != "B" ||
		value.IndexCV31 != nil || value.IndexCV32 != nil || value.SourceFileID != file.ID ||
		value.SourceLine == nil || *value.SourceLine != line || value.SourceExportedAt != first.SourceExportedAt ||
		value.Value != 4 || len(value.History) != 1 || value.History[0].OldValue != 1 || value.History[0].NewValue != 4 {
		t.Fatal("retry lost identity, provenance, index state or history", value)
	}
	source, err := s.GetCVFile(ctx, v.ID, file.ID)
	if err != nil || source.SourceReferenceCount != 1 {
		t.Fatal("source reference lost", err)
	}
}
