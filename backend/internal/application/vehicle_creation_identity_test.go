package application_test

import (
	"errors"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestVehicleCreationIdentityKeepsOneResultAndAllowsDistinctIdenticalVehicles(t *testing.T) {
	db := testDB(t)
	service := application.NewVehicleService(db)
	input := application.CreateVehicleInput{Manufacturer: "Piko", Name: "Same article", Gauge: "H0", Category: "Lokomotive", Gattung: "Diesellok"}
	id := strings.Repeat("a", 32)
	created, err := service.CreateWithOptions(t.Context(), input, application.VehicleCreationOptions{CreationID: id}, "actor")
	if err != nil || created.ID != id {
		t.Fatalf("first creation: %#v / %v", created, err)
	}
	var before, after int
	if err := db.QueryRow(`SELECT SUM(next_number) FROM inventory_number_schemes`).Scan(&before); err != nil {
		t.Fatal(err)
	}
	input.Name = "Retry must not replace"
	for range 2 {
		if _, err := service.CreateWithOptions(t.Context(), input, application.VehicleCreationOptions{CreationID: id}, "actor"); !errors.Is(err, application.ErrVehicleCreationConflict) {
			t.Fatalf("replay: %v", err)
		}
	}
	if err := db.QueryRow(`SELECT SUM(next_number) FROM inventory_number_schemes`).Scan(&after); err != nil {
		t.Fatal(err)
	}
	if before != after {
		t.Fatal("replay consumed inventory numbers")
	}
	stored, err := service.Get(t.Context(), id)
	if err != nil || stored.Name != "Same article" {
		t.Fatalf("replay changed stored vehicle: %#v / %v", stored, err)
	}
	input.Name = "Same article"
	second, err := service.CreateWithOptions(t.Context(), input, application.VehicleCreationOptions{CreationID: strings.Repeat("b", 32)}, "actor")
	if err != nil || second.ID == created.ID {
		t.Fatalf("intentional distinct creation: %#v / %v", second, err)
	}
	for _, invalid := range []string{"short", strings.Repeat("A", 32), strings.Repeat("z", 32)} {
		if _, err := service.CreateWithOptions(t.Context(), input, application.VehicleCreationOptions{CreationID: invalid}, "actor"); !errors.Is(err, application.ErrVehicleCreationIdentity) {
			t.Fatalf("invalid identity %q: %v", invalid, err)
		}
	}
	var vehicles, audits int
	if err := db.QueryRow(`SELECT COUNT(*) FROM vehicles`).Scan(&vehicles); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM audit_logs WHERE action='VehicleCreated'`).Scan(&audits); err != nil {
		t.Fatal(err)
	}
	if vehicles != 2 || audits != 2 {
		t.Fatalf("unexpected extra creations: %d / %d", vehicles, audits)
	}
}

func TestVehicleCreationIdentityPinsSetAndMembersAndRejectsCrossKindReplay(t *testing.T) {
	db := testDB(t)
	service := application.NewVehicleService(db)
	id := strings.Repeat("c", 32)
	input := application.CreateVehicleSetInput{CreationID: id, Set: application.VehicleSetInput{Manufacturer: "Roco", Name: "Same set", Gauge: "H0", Category: "Wagen", Gattung: "Reisezugwagen"},
		Members: []application.CreateVehicleInput{{Name: "Member A"}, {Name: "Member B"}}}
	created, err := service.CreateSet(t.Context(), input, "actor")
	if err != nil || created.ID != id || len(created.Members) != 2 {
		t.Fatalf("first set: %#v / %v", created, err)
	}
	if _, err := service.CreateSet(t.Context(), input, "actor"); !errors.Is(err, application.ErrVehicleCreationConflict) {
		t.Fatalf("set replay: %v", err)
	}
	_, err = service.CreateWithOptions(t.Context(), application.CreateVehicleInput{Manufacturer: "Roco", Name: "Another kind", Gauge: "H0", Category: "Wagen", Gattung: "Reisezugwagen"}, application.VehicleCreationOptions{CreationID: id}, "actor")
	if !errors.Is(err, application.ErrVehicleCreationConflict) {
		t.Fatalf("cross-kind replay: %v", err)
	}
	var sets, members, vehicles int
	if err := db.QueryRow(`SELECT COUNT(*) FROM vehicle_sets`).Scan(&sets); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM vehicle_set_members`).Scan(&members); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM vehicles`).Scan(&vehicles); err != nil {
		t.Fatal(err)
	}
	if sets != 1 || members != 2 || vehicles != 2 {
		t.Fatalf("replay created extra rows: sets=%d members=%d vehicles=%d", sets, members, vehicles)
	}
	loaded, err := service.GetSet(t.Context(), id)
	if err != nil || loaded.Members[0].ID != created.Members[0].ID || loaded.Members[1].ID != created.Members[1].ID {
		t.Fatalf("set member identities changed: %#v / %v", loaded, err)
	}
}
