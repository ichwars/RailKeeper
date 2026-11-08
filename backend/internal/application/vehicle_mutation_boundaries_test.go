package application_test

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/infrastructure/vehicledata"
	vehiclemodel "railkeeper/backend/internal/inventory/vehicle"
)

type vehicleMutationEnvironment struct {
	db                  *sql.DB
	t                   *testing.T
	completionError     error
	cleanupCalls        int
	cleanupVehicleCount int
}

func (e *vehicleMutationEnvironment) Existing(ctx context.Context, id string) (*vehiclemodel.Vehicle, error) {
	return application.NewVehicleService(e.db).Get(ctx, id)
}
func (e *vehicleMutationEnvironment) LocalizeImages(_ context.Context, _ string, images []vehiclemodel.VehicleImageInput) ([]vehiclemodel.VehicleImageInput, error) {
	return images, nil
}
func (e *vehicleMutationEnvironment) CleanupImages(ctx context.Context, _ []vehiclemodel.VehicleImageInput) {
	e.cleanupCalls++
	// A single DB connection makes an unreleased transaction observable here.
	checkCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
	defer cancel()
	if err := e.db.QueryRowContext(checkCtx, `SELECT COUNT(*) FROM vehicles`).Scan(&e.cleanupVehicleCount); err != nil {
		e.t.Fatal("cleanup ran before transaction release", err)
	}
}
func (e *vehicleMutationEnvironment) CompleteCreated(context.Context, *vehiclemodel.Vehicle) error {
	return e.completionError
}
func (e *vehicleMutationEnvironment) CompleteUpdated(context.Context, *vehiclemodel.Vehicle) error {
	return e.completionError
}
func mutationCounts(t *testing.T, db *sql.DB) map[string]int {
	t.Helper()
	counts := map[string]int{}
	for _, table := range []string{"vehicles", "vehicle_images", "vehicle_functions", "vehicle_external_mappings", "audit_logs", "inventory_number_history"} {
		var count int
		if err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count); err != nil {
			t.Fatal(err)
		}
		counts[table] = count
	}
	var next int
	if err := db.QueryRow(`SELECT SUM(next_number) FROM inventory_number_schemes`).Scan(&next); err != nil {
		t.Fatal(err)
	}
	counts["next-number-sum"] = next
	return counts
}
func mutationInput() vehiclemodel.CreateVehicleInput {
	return vehiclemodel.CreateVehicleInput{Manufacturer: "Reference", Name: "preserved", Gauge: "TT", Category: "Lokomotive", Gattung: "Diesellok",
		Images:          []vehiclemodel.VehicleImageInput{{URL: "https://example.test/original.png", SourceURL: "https://example.test/source", IsPrimary: true}},
		FunctionChanges: []vehiclemodel.VehicleFunctionChange{{FunctionKey: "F0", Value: &vehiclemodel.VehicleFunctionInput{Name: "light", FlagsVersion: 2}}},
	}
}
func TestVehicleSourceReadbackFailureRollsBackEntireCreationBeforeCleanup(t *testing.T) {
	db := testDB(t)
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`CREATE TRIGGER fail_created_source_readback AFTER INSERT ON vehicle_external_mappings
BEGIN DELETE FROM vehicles WHERE id=NEW.vehicle_id; END`); err != nil {
		t.Fatal(err)
	}
	before := mutationCounts(t, db)
	environment := &vehicleMutationEnvironment{db: db, t: t}
	service := vehiclemodel.NewCoreService(vehicledata.NewCoreStore(db), environment)
	result, err := service.CreateWithOptions(t.Context(), mutationInput(), vehiclemodel.CreationOptions{
		CreationID: strings.Repeat("d", 32), ExternalMapping: &vehiclemodel.VehicleExternalMapInput{Provider: "ecos", ExternalID: "source-identity"},
	}, "actor")
	if result != nil || !errors.Is(err, application.ErrVehicleNotFound) {
		t.Fatal("failed readback acknowledged", result, err)
	}
	if environment.cleanupCalls != 1 || environment.cleanupVehicleCount != before["vehicles"] {
		t.Fatal("cleanup saw unrolled state", environment)
	}
	if after := mutationCounts(t, db); !reflect.DeepEqual(before, after) {
		t.Fatal("failed source readback consumed numbers or partial data", before, after)
	}
}
func TestVehicleLostPostCommitResponseKeepsOneCreationAndDoesNotConsumeAnotherNumber(t *testing.T) {
	db := testDB(t)
	db.SetMaxOpenConns(1)
	failure := errors.New("synthetic response readback failure")
	environment := &vehicleMutationEnvironment{db: db, t: t, completionError: failure}
	service := vehiclemodel.NewCoreService(vehicledata.NewCoreStore(db), environment)
	options := vehiclemodel.CreationOptions{CreationID: strings.Repeat("e", 32), ExternalMapping: &vehiclemodel.VehicleExternalMapInput{Provider: "ecos", ExternalID: "source-identity"}}
	result, err := service.CreateWithOptions(t.Context(), mutationInput(), options, "actor")
	if result != nil || !errors.Is(err, failure) {
		t.Fatal("expected a lost postcommit response", result, err)
	}
	committed := mutationCounts(t, db)
	if committed["vehicles"] != 1 || committed["vehicle_images"] != 1 || committed["vehicle_functions"] != 1 ||
		committed["vehicle_external_mappings"] != 1 || committed["audit_logs"] != 1 || environment.cleanupVehicleCount != 1 {
		t.Fatal("postcommit response failure lost committed aggregate", committed, environment)
	}
	for range 2 {
		if result, err := service.CreateWithOptions(t.Context(), mutationInput(), options, "actor"); result != nil || !errors.Is(err, application.ErrVehicleCreationConflict) {
			t.Fatal("replay created another vehicle", result, err)
		}
	}
	if after := mutationCounts(t, db); !reflect.DeepEqual(committed, after) || environment.cleanupCalls != 1 {
		t.Fatal("retry mutated committed data or number counter", committed, after, environment.cleanupCalls)
	}
	stored, err := application.NewVehicleService(db).Get(t.Context(), options.CreationID)
	if err != nil || stored.Name != "preserved" || len(stored.Functions) != 1 || len(stored.Images) != 1 || len(stored.ExternalMappings) != 1 {
		t.Fatal("committed result cannot be recovered", stored, err)
	}
}
