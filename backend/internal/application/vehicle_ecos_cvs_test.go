package application_test

import (
	"reflect"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestECoSCVImportAllValuesPreservesManualProvenanceAndIndexedPages(t *testing.T) {
	db := backupTestDB(t, t.TempDir())
	s := application.NewVehicleService(db)
	v, err := s.Create(t.Context(), application.CreateVehicleInput{
		Manufacturer: "Test", Name: "Test", Gauge: "H0", Category: "Lokomotive", Gattung: "Diesellok",
	}, "admin")
	if err != nil {
		t.Fatal(err)
	}
	page31, page32, line := 16, 0, 42
	manual := []application.VehicleCVValueInput{
		{CVNumber: 1, Value: 99, DecoderProfile: "LokProgrammer", Description: "manual", SourceLine: &line, SourceExportedAt: "2026-10-17T17:41:59"},
		{CVNumber: 257, Value: 66, DecoderProfile: "LokProgrammer", IndexCV31: &page31, IndexCV32: &page32},
	}
	if _, err := s.ImportCVValues(t.Context(), v.ID, application.VehicleCVImportInput{Values: manual}); err != nil {
		t.Fatal(err)
	}
	before, err := s.ListCVValues(t.Context(), v.ID)
	if err != nil {
		t.Fatal(err)
	}
	source := application.ECoSStoredCVSource{Host: "ecos.local", ObjectID: "1002", ReadAt: "2026-10-22T11:00:00Z", Protocol: "DCC"}
	for number := 1; number <= 1024; number++ {
		source.Values = append(source.Values, application.ECoSCVValue{Number: number, Value: number % 256})
	}
	if _, err := s.ImportECoSStoredCVs(t.Context(), v.ID, source); err == nil {
		t.Fatal("unmapped import accepted")
	}
	if _, err := s.UpsertExternalMapping(t.Context(), v.ID, application.VehicleExternalMapInput{Provider: "ecos", ExternalID: "1002"}, "admin"); err != nil {
		t.Fatal(err)
	}
	result, err := s.ImportECoSStoredCVs(t.Context(), v.ID, source)
	if err != nil || result.Imported != 1023 || result.Preserved != 1 {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	stored, err := s.ListCVValues(t.Context(), v.ID)
	if err != nil || len(stored) != 1025 {
		t.Fatalf("count=%d err=%v", len(stored), err)
	}
	for _, original := range before {
		found := false
		for _, current := range stored {
			if current.ID == original.ID {
				found = true
				if !reflect.DeepEqual(current, original) {
					t.Fatal("original data/provenance changed")
				}
			}
		}
		if !found {
			t.Fatal("original CV missing")
		}
	}
	for _, value := range stored {
		if value.DecoderProfile == "ECoS" && (value.IndexCV31 != nil || value.IndexCV32 != nil ||
			!strings.Contains(value.Description, "ecos.local · #1002") || !strings.Contains(value.Description, source.ReadAt)) {
			t.Fatalf("incorrect provenance/index: %+v", value)
		}
	}
	result, err = s.ImportECoSStoredCVs(t.Context(), v.ID, source)
	if err != nil || result.Imported != 0 || result.Preserved != 1024 {
		t.Fatalf("non-idempotent import: %+v %v", result, err)
	}
	after, err := s.ListCVValues(t.Context(), v.ID)
	if err != nil || !reflect.DeepEqual(after, stored) {
		t.Fatal("repeated import changed stored data")
	}
	// Invalid data at the end must not leave earlier rows inserted or changed.
	source.Values[1023].Value = 256
	if _, err := s.ImportECoSStoredCVs(t.Context(), v.ID, source); err == nil {
		t.Fatal("invalid byte accepted")
	}
	final, _ := s.ListCVValues(t.Context(), v.ID)
	if !reflect.DeepEqual(final, stored) {
		t.Fatal("partial failed import")
	}
}
