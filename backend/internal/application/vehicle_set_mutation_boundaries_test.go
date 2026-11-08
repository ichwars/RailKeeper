package application_test

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/infrastructure/vehicledata"
	vehiclemodel "railkeeper/backend/internal/inventory/vehicle"
)

type setMutationEnvironment struct {
	*vehicleMutationEnvironment
	responseError error
}

func (e *setMutationEnvironment) LoadMembers(context.Context, string) ([]vehiclemodel.Vehicle, error) {
	return nil, errors.New("unexpected full member reload during set creation")
}
func (e *setMutationEnvironment) AttachMainImage(context.Context, *vehiclemodel.VehicleSet) error {
	return e.responseError
}
func setMutationInput() vehiclemodel.CreateVehicleSetInput {
	first, second := mutationInput(), mutationInput()
	first.Name, first.Manufacturer = "Individual A", "Individual maker"
	second.Name, second.Manufacturer = "Individual B", ""
	return vehiclemodel.CreateVehicleSetInput{CreationID: strings.Repeat("f", 32),
		Set:     vehiclemodel.VehicleSetInput{Manufacturer: "Reference", Name: "original set", Gauge: "TT", Category: "Wagen", Gattung: "Reisezugwagen"},
		Members: []vehiclemodel.CreateVehicleInput{first, second},
	}
}
func TestVehicleSetFinalAuditFailureRollsBackMembersNumbersAndFilesBeforeCleanup(t *testing.T) {
	db := testDB(t)
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`CREATE TRIGGER fail_final_set_audit BEFORE INSERT ON audit_logs
WHEN NEW.action='VehicleSetCreated' BEGIN SELECT RAISE(ABORT,'synthetic final set audit failure'); END`); err != nil {
		t.Fatal(err)
	}
	before := mutationCounts(t, db)
	environment := &setMutationEnvironment{vehicleMutationEnvironment: &vehicleMutationEnvironment{db: db, t: t}}
	service := vehiclemodel.NewSetService(vehicledata.NewSetStore(db), environment)
	result, err := service.CreateSet(t.Context(), setMutationInput(), "actor")
	if result != nil || err == nil || !strings.Contains(err.Error(), "synthetic final set audit failure") {
		t.Fatal("failed final set audit acknowledged", result, err)
	}
	if environment.cleanupCalls != 2 || environment.cleanupVehicleCount != before["vehicles"] {
		t.Fatal("member cleanup saw an unreleased or partial transaction", environment)
	}
	if after := mutationCounts(t, db); !reflect.DeepEqual(before, after) {
		t.Fatal("late set failure left vehicle/media/functions/audit or consumed numbers", before, after)
	}
	for _, table := range []string{"vehicle_sets", "vehicle_set_members"} {
		var count int
		if err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count); err != nil || count != 0 {
			t.Fatal("late set failure left membership", table, count, err)
		}
	}
}
func TestVehicleSetLostResponseRetainsOrderedMembersAndOneCreation(t *testing.T) {
	db := testDB(t)
	db.SetMaxOpenConns(1)
	failure := errors.New("synthetic postcommit set-image read failure")
	environment := &setMutationEnvironment{vehicleMutationEnvironment: &vehicleMutationEnvironment{db: db, t: t}, responseError: failure}
	service := vehiclemodel.NewSetService(vehicledata.NewSetStore(db), environment)
	input := setMutationInput()
	result, err := service.CreateSet(t.Context(), input, "actor")
	if result != nil || !errors.Is(err, failure) {
		t.Fatal("expected failed response", result, err)
	}
	committed := mutationCounts(t, db)
	if committed["vehicles"] != 2 || committed["vehicle_images"] != 2 || committed["vehicle_functions"] != 2 || committed["audit_logs"] != 3 ||
		environment.cleanupCalls != 2 || environment.cleanupVehicleCount != 2 {
		t.Fatal("failed response lost committed members", committed, environment)
	}
	for range 2 {
		if result, err := service.CreateSet(t.Context(), input, "actor"); result != nil || !errors.Is(err, application.ErrVehicleCreationConflict) {
			t.Fatal("replayed set created new members", result, err)
		}
	}
	if after := mutationCounts(t, db); !reflect.DeepEqual(committed, after) || environment.cleanupCalls != 2 {
		t.Fatal("set retry consumed numbers or mutated data", committed, after, environment.cleanupCalls)
	}
	recovered, err := application.NewVehicleService(db).GetSet(t.Context(), input.CreationID)
	if err != nil {
		t.Fatal(err)
	}
	if len(recovered.Members) != 2 || recovered.Members[0].Name != "Individual A" || recovered.Members[0].Manufacturer != "Individual maker" ||
		recovered.Members[1].Name != "Individual B" || recovered.Members[1].Manufacturer != "Reference" ||
		recovered.Members[0].VehicleSetPosition != 1 || recovered.Members[1].VehicleSetPosition != 2 {
		t.Fatal("committed member identity/order/defaults/individual values lost", recovered)
	}
}

func TestVehicleSetUpdateKeepsCommittedStateWhenAggregateReloadFails(t *testing.T) {
	db := testDB(t)
	service := application.NewVehicleService(db)
	input := setMutationInput()
	created, err := service.CreateSet(t.Context(), input, "actor")
	if err != nil {
		t.Fatal(err)
	}
	// The canonical set write succeeds before the separately owned CV enrichment fails.
	if _, err := db.Exec(`DROP TABLE vehicle_cv_values`); err != nil {
		t.Fatal(err)
	}
	input.Set.Name = "committed despite lost response"
	result, err := service.UpdateSet(t.Context(), created.ID, input.Set, "actor")
	if err == nil || result != nil {
		t.Fatal("failed aggregate reload acknowledged", result, err)
	}
	var name, memberName string
	var audits int
	if err := db.QueryRow(`SELECT name FROM vehicle_sets WHERE id=?`, created.ID).Scan(&name); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT name FROM vehicles WHERE id=?`, created.Members[0].ID).Scan(&memberName); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM audit_logs WHERE action='VehicleSetUpdated' AND target_id=?`, created.ID).Scan(&audits); err != nil {
		t.Fatal(err)
	}
	if name != input.Set.Name || memberName != "Individual A" || audits != 1 {
		t.Fatal("postcommit reload error lost canonical or individual state", name, memberName, audits)
	}
}
