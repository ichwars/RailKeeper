package application_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestVehicleConflictRollsBackImagesInventoryHistoryAndAudit(t *testing.T) {
	db := testDB(t)
	service := application.NewVehicleService(db)
	input := application.CreateVehicleInput{
		InventoryNumber: "KEEP-1", Manufacturer: "Reference", Name: "Before",
		Gauge: "TT", Category: "Lokomotive", Gattung: "Diesellok",
		Images: []application.VehicleImageInput{{URL: "https://example.test/original.png", Title: "Original", IsPrimary: true}},
		FunctionChanges: []application.VehicleFunctionChange{{FunctionKey: "F0", Value: &application.VehicleFunctionInput{
			Name: "Original light", FlagsVersion: 2,
		}}},
	}
	created, err := service.Create(t.Context(), input, "actor")
	if err != nil {
		t.Fatal(err)
	}
	before, err := service.Get(t.Context(), created.ID)
	if err != nil {
		t.Fatal(err)
	}
	counts := func() [2]int {
		t.Helper()
		var result [2]int
		for index, table := range []string{"audit_logs", "inventory_number_history"} {
			if err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&result[index]); err != nil {
				t.Fatal(err)
			}
		}
		return result
	}
	beforeCounts := counts()
	input.Name = "Must roll back"
	input.InventoryNumber = "KEEP-2"
	input.Images = []application.VehicleImageInput{{URL: "https://example.test/replacement.png", Title: "Replacement"}}
	input.FunctionChanges = []application.VehicleFunctionChange{{
		FunctionKey: "F0", ExpectedRevision: strings.Repeat("0", 64), Delete: true,
	}}
	if _, err := service.Update(t.Context(), created.ID, input, "actor"); !errors.Is(err, application.ErrVehicleFunctionConflict) {
		t.Fatalf("expected a late function conflict, got %v", err)
	}
	after, err := service.Get(t.Context(), created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, after) || beforeCounts != counts() {
		t.Fatal("conflict left a partial vehicle, image, history or audit update")
	}
}

func TestVehicleImagesDistinguishOmissionFromExplicitClearing(t *testing.T) {
	for _, clear := range []bool{false, true} {
		t.Run(map[bool]string{false: "omitted", true: "explicit-empty"}[clear], func(t *testing.T) {
			service := application.NewVehicleService(testDB(t))
			input := application.CreateVehicleInput{
				Manufacturer: "Reference", Name: "Before", Gauge: "TT", Category: "Lokomotive", Gattung: "Diesellok",
				Images: []application.VehicleImageInput{{URL: "https://example.test/original.png", SourceURL: "https://example.test/source", IsPrimary: true}},
			}
			created, err := service.Create(t.Context(), input, "actor")
			if err != nil {
				t.Fatal(err)
			}
			input.Name = "After"
			input.Images = nil
			if clear {
				input.Images = []application.VehicleImageInput{}
			}
			updated, err := service.Update(t.Context(), created.ID, input, "actor")
			if err != nil {
				t.Fatal(err)
			}
			if clear && len(updated.Images) != 0 {
				t.Fatal("explicit clearing kept images")
			}
			if !clear && !reflect.DeepEqual(created.Images, updated.Images) {
				t.Fatal("an omitted image field changed persisted images or provenance")
			}
		})
	}
}
