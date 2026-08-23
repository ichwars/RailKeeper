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

func TestVehicleExternalMappingRebindReplacesStaleProviderIdentity(t *testing.T) {
	db := testDB(t)
	service := application.NewVehicleService(db)
	vehicle, err := service.Create(t.Context(), application.CreateVehicleInput{
		Manufacturer: "Roco", Name: "BR 18 201 Roco S", Gauge: "H0",
		Category: "Lokomotive", Gattung: "Dampflok",
	}, "admin-1")
	if err != nil {
		t.Fatal(err)
	}
	old, err := service.UpsertExternalMapping(t.Context(), vehicle.ID, application.VehicleExternalMapInput{
		Provider: "ecos", ExternalID: "1020", ExternalName: vehicle.Name,
		ExternalAddress: "4405", ExternalProtocol: "DCC28",
	}, "admin-1")
	if err != nil {
		t.Fatal(err)
	}
	rebound, err := service.RebindExternalMapping(t.Context(), vehicle.ID, old.ExternalID,
		application.VehicleExternalMapInput{
			Provider: "ecos", ExternalID: "1002", ExternalName: vehicle.Name,
			ExternalAddress: "4405", ExternalProtocol: "DCC", SyncStatus: "synced",
		}, "admin-1")
	if err != nil || rebound.ExternalID != "1002" || rebound.VehicleID != vehicle.ID {
		t.Fatalf("rebound=%#v err=%v", rebound, err)
	}
	var oldCount, newCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM vehicle_external_mappings
WHERE vehicle_id=? AND provider='ecos' AND external_id='1020'`, vehicle.ID).Scan(&oldCount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM vehicle_external_mappings
WHERE vehicle_id=? AND provider='ecos' AND external_id='1002'`, vehicle.ID).Scan(&newCount); err != nil {
		t.Fatal(err)
	}
	if oldCount != 0 || newCount != 1 {
		t.Fatalf("oldCount=%d newCount=%d", oldCount, newCount)
	}
}

func TestVehicleExternalMappingRebindConsolidatesExistingMappingForSameVehicle(t *testing.T) {
	db := testDB(t)
	service := application.NewVehicleService(db)
	vehicle, err := service.Create(t.Context(), application.CreateVehicleInput{
		Manufacturer: "Roco", Name: "BR 18", Gauge: "H0",
		Category: "Lokomotive", Gattung: "Dampflok",
	}, "admin-1")
	if err != nil {
		t.Fatal(err)
	}
	for _, externalID := range []string{"1020", "1002"} {
		if _, err := service.UpsertExternalMapping(t.Context(), vehicle.ID,
			application.VehicleExternalMapInput{Provider: "ecos", ExternalID: externalID,
				ExternalName: "Alt", ExternalAddress: "3", ExternalProtocol: "DCC28"}, "admin-1"); err != nil {
			t.Fatal(err)
		}
	}

	rebound, err := service.RebindExternalMapping(t.Context(), vehicle.ID, "1020",
		application.VehicleExternalMapInput{Provider: "ecos", ExternalID: "1002",
			ExternalName: "BR 18 aktuell", ExternalAddress: "18", ExternalProtocol: "DCC",
			SyncStatus: "synced"}, "admin-1")
	if err != nil {
		t.Fatal(err)
	}
	if rebound.ExternalName != "BR 18 aktuell" || rebound.ExternalAddress != "18" ||
		rebound.ExternalProtocol != "DCC" || rebound.SyncStatus != "synced" {
		t.Fatalf("rebound=%#v", rebound)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM vehicle_external_mappings
WHERE vehicle_id=? AND provider='ecos'`, vehicle.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("mapping count=%d", count)
	}
}
