package application_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestCVFileDeletionGuardIncludesNewSourceReferences(t *testing.T) {
	db := testDB(t)
	s := application.NewVehicleService(db)
	v, err := s.Create(t.Context(), application.CreateVehicleInput{Manufacturer: "Reference", Name: "File guards",
		Gauge: "TT", Category: "Lokomotive", Gattung: "Diesellok"}, "actor")
	if err != nil {
		t.Fatal(err)
	}
	file, err := s.CreateCVFile(t.Context(), v.ID, application.VehicleCVFileInput{FileName: "decoder.csv", OriginalName: "decoder.csv",
		DecoderProfile: "Legacy", StoragePath: "uploads/decoder.csv"})
	if err != nil {
		t.Fatal(err)
	}
	initial, err := s.GetCVFile(t.Context(), v.ID, file.ID)
	if err != nil || initial.Revision != file.Revision {
		t.Fatalf("inconsistent initial revision: %v", err)
	}
	line, zero := 2, 0
	value, err := s.CreateCVValue(t.Context(), v.ID, application.VehicleCVValueInput{CVNumber: 257, Value: 7,
		IndexCV31: &zero, IndexCV32: &zero, SourceFileID: file.ID, SourceLine: &line, SourceExportedAt: "2026-10-20T08:30:00"})
	if err != nil {
		t.Fatal(err)
	}
	current, err := s.GetCVFile(t.Context(), v.ID, file.ID)
	if err != nil {
		t.Fatal(err)
	}
	if current.Revision == initial.Revision || current.UpdatedAt != initial.UpdatedAt || current.SourceReferenceCount != 1 || initial.SourceReferenceCount != 0 {
		t.Fatal("source links must affect the revision without changing the file clock")
	}
	listed, err := s.ListCVFiles(t.Context(), v.ID)
	if err != nil || len(listed) != 1 || listed[0].Revision != current.Revision {
		t.Fatalf("list/read revision mismatch: %v", err)
	}
	before, err := s.ListCVValues(t.Context(), v.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.DeleteCVFile(t.Context(), v.ID, file.ID, initial.Revision); !errors.Is(err, application.ErrVehicleCVFileConflict) {
		t.Fatalf("expected source-reference conflict: %v", err)
	}
	if _, err := s.DeleteCVFile(t.Context(), v.ID, file.ID, strings.Repeat("Z", 64)); !errors.Is(err, application.ErrVehicleValidation) {
		t.Fatalf("expected invalid guard: %v", err)
	}
	if _, err := s.GetCVFile(t.Context(), v.ID, file.ID); err != nil {
		t.Fatal("conflict deleted the file")
	}
	if _, err := s.DeleteCVFile(t.Context(), "other-vehicle", file.ID, current.Revision); !errors.Is(err, application.ErrVehicleNotFound) {
		t.Fatalf("foreign file accepted: %v", err)
	}
	if _, err := s.DeleteCVFile(t.Context(), v.ID, file.ID, current.Revision); err != nil {
		t.Fatal(err)
	}
	after, err := s.ListCVValues(t.Context(), v.ID)
	if err != nil {
		t.Fatal(err)
	}
	// File deletion preserves the legacy historical source ID, lines, times, identities and history.
	if !reflect.DeepEqual(before, after) || after[0].ID != value.ID {
		t.Fatal("file deletion changed stored CV data/provenance")
	}
	// Keep historical provenance usable after its original file was explicitly removed.
	expected := after[0].Revision
	updated, err := s.UpdateCVValue(t.Context(), v.ID, value.ID, application.VehicleCVValueInput{
		CVNumber: 257, Value: 8, IndexCV31: &zero, IndexCV32: &zero, SourceFileID: file.ID, ExpectedRevision: &expected,
	})
	if err != nil || updated.SourceFileID != file.ID || updated.SourceLine == nil || *updated.SourceLine != line ||
		updated.SourceExportedAt != "2026-10-20T08:30:00" || len(updated.History) != 1 {
		t.Fatalf("historical source lost or blocked editing: %v", err)
	}
	if _, err := s.CreateCVValue(t.Context(), v.ID, application.VehicleCVValueInput{CVNumber: 3, Value: 1, SourceFileID: file.ID}); !errors.Is(err, application.ErrVehicleValidation) {
		t.Fatalf("a missing file cannot be assigned to a new CV: %v", err)
	}
	bad := application.VehicleCVValueInput{CVNumber: 257, Value: 9, IndexCV31: &zero, IndexCV32: &zero, SourceFileID: "other-missing-file"}
	if _, err := s.UpdateCVValue(t.Context(), v.ID, value.ID, bad); !errors.Is(err, application.ErrVehicleValidation) {
		t.Fatalf("a different missing source was accepted: %v", err)
	}

}
