package application_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestCVDraftPreservesIndexedIdentityProvenanceHistoryAndBackup(t *testing.T) {
	dir := t.TempDir()
	db := backupTestDB(t, dir)
	s := application.NewVehicleService(db)
	v, err := s.Create(t.Context(), application.CreateVehicleInput{Manufacturer: "Test", Name: "CV draft", Gauge: "TT", Category: "Lokomotive", Gattung: "Diesellok"}, "test")
	if err != nil {
		t.Fatal(err)
	}
	zero, one, line := 0, 1, 90
	for _, value := range []application.VehicleCVValueInput{
		{CVNumber: 257, Value: 3, DecoderProfile: "A", SourceLine: &line, SourceExportedAt: "2026-10-17T17:41:59"},
		{CVNumber: 257, Value: 4, DecoderProfile: "A", IndexCV31: &zero, IndexCV32: &zero},
		{CVNumber: 257, Value: 5, DecoderProfile: "A", IndexCV31: &one, IndexCV32: &zero},
	} {
		if _, err := s.CreateCVValue(t.Context(), v.ID, value); err != nil {
			t.Fatal(err)
		}
	}
	before, err := s.ListCVValues(t.Context(), v.ID)
	if err != nil {
		t.Fatal(err)
	}
	first, indexed := before[0], before[1]
	input := application.VehicleCVChangesInput{Changes: []application.VehicleCVChange{
		{ID: first.ID, ExpectedRevision: first.Revision, Value: &application.VehicleCVValueInput{CVNumber: 257, Value: 7, DecoderProfile: "A"}},
		{ID: indexed.ID, ExpectedRevision: indexed.Revision, Value: &application.VehicleCVValueInput{CVNumber: 257, Value: 8, DecoderProfile: "A", IndexCV31: &zero, IndexCV32: &zero}},
		{ExpectedRevision: "", Value: &application.VehicleCVValueInput{CVNumber: 1, Value: 0, DecoderProfile: "B"}},
	}}
	if count, err := s.ApplyCVChanges(t.Context(), v.ID, input); err != nil || count != 3 {
		t.Fatalf("save: %d %v", count, err)
	}
	after, err := s.ListCVValues(t.Context(), v.ID)
	if err != nil || len(after) != 4 {
		t.Fatalf("values: %+v %v", after, err)
	}
	if after[0].ID != first.ID || after[0].Value != 7 || after[0].IndexCV31 != nil ||
		after[0].SourceLine == nil || *after[0].SourceLine != line || after[0].SourceExportedAt != first.SourceExportedAt ||
		after[0].Revision == first.Revision || len(after[0].History) != 1 ||
		after[1].ID != indexed.ID || after[1].Value != 8 || after[1].IndexCV31 == nil || *after[1].IndexCV31 != 0 ||
		!reflect.DeepEqual(before[2], after[2]) {
		t.Fatalf("identity/source/history changed unexpectedly: %+v", after)
	}
	// Repeating a request with create-only identities and original revisions is rejected without duplicates.
	if _, err := s.ApplyCVChanges(t.Context(), v.ID, input); !errors.Is(err, application.ErrVehicleCVConflict) {
		t.Fatalf("retry: %v", err)
	}
	unchanged, _ := s.ListCVValues(t.Context(), v.ID)
	if !reflect.DeepEqual(after, unchanged) {
		t.Fatal("repeated draft changed rows/history")
	}
	backup := application.NewBackupService(db, dir)
	document, err := backup.Export(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := backup.Import(t.Context(), document); err != nil {
		t.Fatal(err)
	}
	restored, err := s.ListCVValues(t.Context(), v.ID)
	if err != nil || !reflect.DeepEqual(after, restored) {
		t.Fatalf("backup changed CVs, revisions or history: %v", err)
	}
}

func TestCVDraftRollsBackStaleRowsIdentityMovesForeignSourcesAndSQLFailure(t *testing.T) {
	db := testDB(t)
	s := application.NewVehicleService(db)
	v, err := s.Create(t.Context(), application.CreateVehicleInput{Manufacturer: "Test", Name: "CV rollback", Gauge: "TT", Category: "Lokomotive", Gattung: "Diesellok"}, "test")
	if err != nil {
		t.Fatal(err)
	}
	a, err := s.CreateCVValue(t.Context(), v.ID, application.VehicleCVValueInput{CVNumber: 1, Value: 3, DecoderProfile: "A"})
	if err != nil {
		t.Fatal(err)
	}
	b, err := s.CreateCVValue(t.Context(), v.ID, application.VehicleCVValueInput{CVNumber: 2, Value: 4, DecoderProfile: "A"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateCVValue(t.Context(), v.ID, application.VehicleCVValueInput{CVNumber: 2, Value: 6, DecoderProfile: "B"}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TRIGGER reject_cv_test BEFORE UPDATE ON vehicle_cv_values
WHEN NEW.value=99 BEGIN SELECT RAISE(ABORT, 'injected CV write failure'); END`); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name   string
		second application.VehicleCVChange
		want   error
	}{
		{"stale", application.VehicleCVChange{ID: b.ID, ExpectedRevision: strings.Repeat("0", 64), Value: &application.VehicleCVValueInput{CVNumber: 2, Value: 9, DecoderProfile: "A"}}, application.ErrVehicleCVConflict},
		{"occupied profile identity", application.VehicleCVChange{ID: b.ID, ExpectedRevision: b.Revision, Value: &application.VehicleCVValueInput{CVNumber: 2, Value: 9, DecoderProfile: "B"}}, application.ErrVehicleCVConflict},
		{"foreign source", application.VehicleCVChange{ID: b.ID, ExpectedRevision: b.Revision, Value: &application.VehicleCVValueInput{CVNumber: 2, Value: 9, DecoderProfile: "A", SourceFileID: "foreign-file"}}, application.ErrVehicleValidation},
		{"SQL failure", application.VehicleCVChange{ID: b.ID, ExpectedRevision: b.Revision, Value: &application.VehicleCVValueInput{CVNumber: 2, Value: 99, DecoderProfile: "A"}}, nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			before, _ := s.ListCVValues(t.Context(), v.ID)
			_, err := s.ApplyCVChanges(t.Context(), v.ID, application.VehicleCVChangesInput{Changes: []application.VehicleCVChange{
				{ID: a.ID, ExpectedRevision: a.Revision, Value: &application.VehicleCVValueInput{CVNumber: 1, Value: 7, DecoderProfile: "A"}}, test.second,
			}})
			if err == nil || (test.want != nil && !errors.Is(err, test.want)) {
				t.Fatalf("expected rejection: %v", err)
			}
			after, _ := s.ListCVValues(t.Context(), v.ID)
			if !reflect.DeepEqual(before, after) {
				t.Fatal("partial CV or history write survived rollback")
			}
		})
	}
}

func TestCVDraftLegacyWritesShareRevisionAndIdentityGuards(t *testing.T) {
	s := application.NewVehicleService(testDB(t))
	v, err := s.Create(t.Context(), application.CreateVehicleInput{Manufacturer: "Test", Name: "CV guards", Gauge: "TT", Category: "Lokomotive", Gattung: "Diesellok"}, "test")
	if err != nil {
		t.Fatal(err)
	}
	input := application.VehicleCVValueInput{CVNumber: 1, Value: 3}
	value, err := s.CreateCVValue(t.Context(), v.ID, input)
	if err != nil {
		t.Fatal(err)
	}
	old := value.Revision
	input.ExpectedRevision, input.Value = &old, 4
	value, err = s.UpdateCVValue(t.Context(), v.ID, value.ID, input)
	if err != nil || len(value.Revision) != 64 {
		t.Fatalf("guarded legacy update: %+v %v", value, err)
	}
	if _, err := s.UpdateCVValue(t.Context(), v.ID, value.ID, input); !errors.Is(err, application.ErrVehicleCVConflict) {
		t.Fatalf("stale update: %v", err)
	}
	if _, err := s.ImportCVValues(t.Context(), v.ID, application.VehicleCVImportInput{Values: []application.VehicleCVValueInput{input}}); !errors.Is(err, application.ErrVehicleCVConflict) {
		t.Fatalf("stale import: %v", err)
	}
	if _, err := s.DeleteCVValue(t.Context(), v.ID, value.ID, old); !errors.Is(err, application.ErrVehicleCVConflict) {
		t.Fatalf("stale deletion: %v", err)
	}
	input.ExpectedRevision = nil
	if _, err := s.UpdateCVValue(t.Context(), v.ID, value.ID, input); err != nil {
		t.Fatalf("old caller without revision: %v", err)
	}
	if _, err := s.CreateCVValue(t.Context(), v.ID, input); !errors.Is(err, application.ErrVehicleCVConflict) {
		t.Fatalf("duplicate identity: %v", err)
	}
	if _, err := s.DeleteCVValue(t.Context(), v.ID, value.ID); err != nil {
		t.Fatalf("old delete caller: %v", err)
	}
}

func TestCVDraftRevisionDetectsMetadataChangesWithTheSameClock(t *testing.T) {
	db := testDB(t)
	s := application.NewVehicleService(db)
	v, err := s.Create(t.Context(), application.CreateVehicleInput{Manufacturer: "Test", Name: "CV metadata", Gauge: "TT", Category: "Lokomotive", Gattung: "Diesellok"}, "test")
	if err != nil {
		t.Fatal(err)
	}
	value, err := s.CreateCVValue(t.Context(), v.ID, application.VehicleCVValueInput{CVNumber: 1, Value: 3})
	if err != nil {
		t.Fatal(err)
	}
	// An older writer can change metadata within the same second, without a different edit clock.
	if _, err := db.Exec(`UPDATE vehicle_cv_values SET description='Changed elsewhere' WHERE id=?`, value.ID); err != nil {
		t.Fatal(err)
	}
	current, err := s.GetCVValue(t.Context(), v.ID, value.ID)
	if err != nil || current.UpdatedAt != value.UpdatedAt || current.Revision == value.Revision {
		t.Fatalf("metadata revision: %+v %v", current, err)
	}
	_, err = s.ApplyCVChanges(t.Context(), v.ID, application.VehicleCVChangesInput{Changes: []application.VehicleCVChange{
		{ID: value.ID, ExpectedRevision: value.Revision, Value: &application.VehicleCVValueInput{CVNumber: 1, Value: 5}},
	}})
	if !errors.Is(err, application.ErrVehicleCVConflict) {
		t.Fatalf("same-clock overwrite accepted: %v", err)
	}
}
