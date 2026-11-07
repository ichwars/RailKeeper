package application_test

import (
	"errors"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestVehicleMappingCreationRollsBackRepeatedSourceAndKeepsLegacyCreation(t *testing.T) {
	for _, provider := range []string{"ecos", "cs3"} {
		t.Run(provider, func(t *testing.T) {
			db := testDB(t)
			service := application.NewVehicleService(db)
			input := application.CreateVehicleInput{Manufacturer: "Piko", Name: "Confirmed source", Gauge: "H0",
				Category: "Lokomotive", Gattung: "Diesellok"}
			source := application.VehicleExternalMapInput{Provider: provider, ExternalID: "77", ExternalName: "Source"}
			first, err := service.CreateWithExternalMapping(t.Context(), input, source, "tester")
			if err != nil || len(first.ExternalMappings) != 1 || first.ExternalMappings[0].VehicleID != first.ID {
				t.Fatalf("first mapped creation: %#v / %v", first, err)
			}
			var counterBefore, counterAfter int
			if err := db.QueryRow(`SELECT SUM(next_number) FROM inventory_number_schemes`).Scan(&counterBefore); err != nil {
				t.Fatal(err)
			}
			for range 2 {
				input.Name = "Retry must not create"
				if _, err := service.CreateWithExternalMapping(t.Context(), input, source, "tester"); !errors.Is(err, application.ErrVehicleExternalMappingConflict) {
					t.Fatalf("repeated creation: %v", err)
				}
			}
			if err := db.QueryRow(`SELECT SUM(next_number) FROM inventory_number_schemes`).Scan(&counterAfter); err != nil {
				t.Fatal(err)
			}
			if counterAfter != counterBefore {
				t.Fatalf("rejected creation consumed inventory numbers: %d -> %d", counterBefore, counterAfter)
			}
			for table, predicate := range map[string]string{
				"vehicles": "1=1", "vehicle_external_mappings": "1=1", "audit_logs": "action='VehicleCreated'",
			} {
				var count int
				if err := db.QueryRow("SELECT COUNT(*) FROM " + table + " WHERE " + predicate).Scan(&count); err != nil || count != 1 {
					t.Fatalf("%s after rejected replays: %d / %v", table, count, err)
				}
			}
			if _, err := service.Create(t.Context(), input, "tester"); err != nil {
				t.Fatalf("ordinary distinct vehicle creation must remain available: %v", err)
			}
		})
	}
}

func TestVehicleMappingWriteKeepsOwnerAcquiredDuringInsert(t *testing.T) {
	db := testDB(t)
	service := application.NewVehicleService(db)
	input := application.CreateVehicleInput{Manufacturer: "Piko", Name: "Owner", Gauge: "H0",
		Category: "Lokomotive", Gattung: "Diesellok"}
	owner, err := service.Create(t.Context(), input, "tester")
	if err != nil {
		t.Fatal(err)
	}
	contender, err := service.Create(t.Context(), input, "tester")
	if err != nil {
		t.Fatal(err)
	}
	// A competing owner appears after any preliminary read but before the UPSERT conflict branch.
	_, err = db.Exec(`CREATE TRIGGER competing_mapping BEFORE INSERT ON vehicle_external_mappings
WHEN NEW.external_id='raced' AND NOT EXISTS(SELECT 1 FROM vehicle_external_mappings WHERE provider='ecos' AND external_id='raced')
BEGIN
 INSERT INTO vehicle_external_mappings(id,vehicle_id,provider,external_id,external_name,sync_status,created_at,updated_at)
 SELECT 'competing-owner',id,'ecos','raced','Keep owner','linked','before','before' FROM vehicles WHERE id<>NEW.vehicle_id LIMIT 1;
END`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = service.UpsertExternalMapping(t.Context(), contender.ID,
		application.VehicleExternalMapInput{Provider: "ecos", ExternalID: "raced", ExternalName: "Must not replace"}, "tester")
	if !errors.Is(err, application.ErrVehicleExternalMappingConflict) {
		t.Fatalf("racing owner conflict: %v", err)
	}
	var id, vehicle, name string
	if err := db.QueryRow(`SELECT id,vehicle_id,external_name FROM vehicle_external_mappings WHERE external_id='raced'`).
		Scan(&id, &vehicle, &name); err != nil {
		t.Fatal(err)
	}
	if id != "competing-owner" || vehicle != owner.ID || name != "Keep owner" {
		t.Fatalf("competing owner overwritten: %s / %s / %s", id, vehicle, name)
	}
}
