package vehicledata_test

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"strings"
	"testing"

	"railkeeper/backend/internal/infrastructure/vehicledata"
	vehiclemodel "railkeeper/backend/internal/inventory/vehicle"
)

func functionFixture(t *testing.T) (*sql.DB, vehiclemodel.FunctionRepository) {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })
	if _, err := db.Exec(`CREATE TABLE vehicle_functions(
 id TEXT PRIMARY KEY,vehicle_id TEXT,function_key TEXT,name TEXT,symbol_key TEXT,function_type TEXT,mode TEXT,
 direction_dependent INTEGER,inverted INTEGER,legacy_direction_dependent INTEGER,notes TEXT,sort_order INTEGER,
 created_at TEXT,updated_at TEXT,UNIQUE(vehicle_id,function_key))`); err != nil {
		t.Fatal(err)
	}
	return db, vehicledata.NewFunctionStore(db)
}
func functionInput(name string) vehiclemodel.VehicleFunctionInput {
	return vehiclemodel.VehicleFunctionInput{Name: name, FunctionType: "standard", Mode: "dauer", FlagsVersion: 2}
}
func TestFunctionUpsertPreservesLegacyIdentityAndConfirmedFlags(t *testing.T) {
	db, repository := functionFixture(t)
	yes, no := true, false
	original := functionInput("recorded")
	original.DirectionDependent, original.Inverted, original.LegacyDirectionDependent = &yes, &no, &no
	if err := repository.StoreFunction(t.Context(), vehiclemodel.PlanFunctionWrite("owner", "F1", original, "recorded")); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE vehicle_functions SET id='historic-function-id'`); err != nil {
		t.Fatal(err)
	}
	parentCalls := 0
	service := vehiclemodel.NewFunctionService(repository, func(_ context.Context, id string) error {
		parentCalls++
		if id != "owner" {
			t.Fatalf("wrong parent %q", id)
		}
		return nil
	})
	unconfirmed := functionInput("  renamed  ")
	unconfirmed.FlagsVersion = 0
	unconfirmed.DirectionDependent, unconfirmed.Inverted = &no, &yes
	item, err := service.UpsertFunction(t.Context(), " owner ", "f01", unconfirmed)
	if err != nil {
		t.Fatal(err)
	}
	if item.ID != "historic-function-id" || item.CreatedAt != "recorded" || item.Name != "renamed" ||
		item.FlagsVersion != 2 || item.DirectionDependent == nil || !*item.DirectionDependent ||
		item.Inverted == nil || *item.Inverted || item.LegacyDirectionDependent == nil || *item.LegacyDirectionDependent {
		t.Fatalf("unconfirmed legacy request changed identity or flags: %+v", item)
	}
	if unconfirmed.DirectionDependent != &no || unconfirmed.Inverted != &yes || unconfirmed.LegacyDirectionDependent != nil {
		t.Fatal("write policy modified caller input")
	}
	confirmed := functionInput("confirmed")
	confirmed.DirectionDependent, confirmed.Inverted, confirmed.LegacyDirectionDependent = &no, &yes, &yes
	updated, err := service.UpsertFunction(t.Context(), "owner", "F1", confirmed)
	if err != nil {
		t.Fatal(err)
	}
	if updated.ID != item.ID || updated.CreatedAt != item.CreatedAt || updated.DirectionDependent == nil ||
		*updated.DirectionDependent || updated.Inverted == nil || !*updated.Inverted ||
		updated.LegacyDirectionDependent == nil || *updated.LegacyDirectionDependent || updated.Revision == item.Revision {
		t.Fatalf("explicit independent flags or immutable legacy value lost: %+v", updated)
	}
	// Get and Delete deliberately do not read the parent aggregate.
	independent := vehiclemodel.NewFunctionService(repository, nil)
	deleted, err := independent.DeleteFunction(t.Context(), " owner ", " F01 ")
	if err != nil || !reflect.DeepEqual(deleted, updated) || parentCalls != 2 {
		t.Fatal(deleted, err, parentCalls)
	}
	if _, err := independent.GetFunction(t.Context(), "owner", "F1"); !errors.Is(err, vehiclemodel.ErrVehicleNotFound) {
		t.Fatal(err)
	}
}
func TestFunctionLegacyInsertionKeepsUnknownFlagsAndDeterministicID(t *testing.T) {
	_, repository := functionFixture(t)
	no := false
	input := functionInput("legacy")
	input.FlagsVersion = 1
	input.DirectionDependent = &no
	if err := repository.StoreFunction(t.Context(), vehiclemodel.PlanFunctionWrite("owner", "F0", input, "original time")); err != nil {
		t.Fatal(err)
	}
	item, exists, err := repository.FindFunction(t.Context(), "owner", "F0")
	if err != nil || !exists {
		t.Fatal(err)
	}
	if item.ID != "owner:F0" || item.CreatedAt != "original time" || item.DirectionDependent != nil || item.Inverted != nil ||
		item.LegacyDirectionDependent == nil || *item.LegacyDirectionDependent || item.FlagsVersion != 2 || item.Revision == "" {
		t.Fatalf("unknown was converted into an asserted modern flag: %+v", item)
	}
}
func TestFunctionBulkFailureUsesCallerRollbackForEveryStoredField(t *testing.T) {
	for _, failure := range []string{"late conflict", "late SQL failure"} {
		t.Run(failure, func(t *testing.T) {
			db, repository := functionFixture(t)
			for _, key := range []string{"F1", "F2"} {
				input := functionInput("original " + key)
				yes := true
				input.Inverted = &yes
				if err := repository.StoreFunction(t.Context(), vehiclemodel.PlanFunctionWrite("owner", key, input, "original "+key)); err != nil {
					t.Fatal(err)
				}
			}
			before, err := repository.Functions(t.Context(), "owner")
			if err != nil {
				t.Fatal(err)
			}
			if failure == "late SQL failure" {
				if _, err := db.Exec(`CREATE TRIGGER fail_second_function BEFORE UPDATE ON vehicle_functions
WHEN NEW.function_key='F2' BEGIN SELECT RAISE(ABORT,'synthetic late function failure'); END`); err != nil {
					t.Fatal(err)
				}
			}
			tx, err := db.BeginTx(t.Context(), nil)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = tx.Rollback() }()
			changes := make([]vehiclemodel.VehicleFunctionChange, 0, 2)
			for _, item := range before {
				input := functionInput("must roll back " + item.FunctionKey)
				revision := item.Revision
				if failure == "late conflict" && item.FunctionKey == "F2" {
					revision = strings.Repeat("0", 64)
				}
				changes = append(changes, vehiclemodel.VehicleFunctionChange{FunctionKey: item.FunctionKey, ExpectedRevision: revision, Value: &input})
			}
			mutation := vehicledata.FunctionMutationFor(tx)
			err = vehiclemodel.ApplyFunctionChanges(t.Context(), mutation, "owner", changes, "must roll back time")
			if err == nil || (failure == "late conflict" && !errors.Is(err, vehiclemodel.ErrVehicleFunctionConflict)) {
				t.Fatal("late failure acknowledged", err)
			}
			partial, exists, readErr := mutation.FindFunction(t.Context(), "owner", "F1")
			if readErr != nil || !exists || partial.Name != "must roll back F1" {
				t.Fatal("first change never executed", partial, readErr)
			}
			if err := tx.Rollback(); err != nil {
				t.Fatal("function port committed caller transaction", err)
			}
			after, err := repository.Functions(t.Context(), "owner")
			if err != nil || !reflect.DeepEqual(before, after) {
				t.Fatal("rollback lost a stored field", before, after, err)
			}
		})
	}
}
