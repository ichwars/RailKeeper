package application_test

import (
	"errors"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestVehicleExternalMappingKeepsAuthoritativeOwner(t *testing.T) {
	db := testDB(t)
	service := application.NewVehicleService(db)
	create := func(name string) *application.Vehicle {
		vehicle, err := service.Create(t.Context(), application.CreateVehicleInput{
			Manufacturer: "Piko",
			Name:         name,
			Gauge:        "H0",
			Category:     "Lokomotive",
			Gattung:      "Diesellok",
		}, "admin-1")
		if err != nil {
			t.Fatal(err)
		}
		return vehicle
	}
	first, second := create("First"), create("Second")
	input := application.VehicleExternalMapInput{
		Provider:         "ecos",
		ExternalID:       "77",
		ExternalName:     "BR 106",
		ExternalAddress:  "3",
		ExternalProtocol: "DCC",
		SyncStatus:       "linked",
	}

	created, err := service.UpsertExternalMapping(t.Context(), first.ID, input, "admin-1")
	if err != nil || created.VehicleID != first.ID {
		t.Fatalf("created=%#v err=%v", created, err)
	}
	repeated, err := service.UpsertExternalMapping(t.Context(), first.ID, input, "admin-1")
	if err != nil || repeated.VehicleID != first.ID {
		t.Fatalf("repeated=%#v err=%v", repeated, err)
	}
	if _, err := service.UpsertExternalMapping(t.Context(), second.ID, input, "admin-1"); !errors.Is(err, application.ErrVehicleExternalMappingConflict) {
		t.Fatalf("conflict error=%v", err)
	}
}
