package application_test

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestVehicleFunctionChangesCommitWithVehicleAndRollbackOnConflict(t *testing.T) {
	db := testDB(t)
	s := application.NewVehicleService(db)
	ctx := context.Background()
	input := application.CreateVehicleInput{Manufacturer: "ESU", Name: "Before", Gauge: "TT", Category: "Lokomotive", Gattung: "Diesellok"}
	value := application.VehicleFunctionInput{Name: "Light", Mode: "dauer", FlagsVersion: 2, Inverted: boolPtr(false), Notes: "Own note\nRailComPlus · CV31=1 / CV32=1 · CV257=0; CV258=3"}
	input.FunctionChanges = []application.VehicleFunctionChange{{FunctionKey: "F0", Value: &value}}
	v, err := s.Create(ctx, input, "actor")
	if err != nil {
		t.Fatal(err)
	}
	if len(v.Functions) != 1 || len(v.Functions[0].Revision) != 64 {
		t.Fatalf("missing created revision: %#v", v.Functions)
	}
	f0 := v.Functions[0]
	value.Name = "Manual light"
	input.Name = "After"
	input.FunctionChanges = []application.VehicleFunctionChange{
		{FunctionKey: "F0", ExpectedRevision: f0.Revision, Value: &value},
		{FunctionKey: "F3", Value: &application.VehicleFunctionInput{Name: "Shunting", Mode: "moment", FlagsVersion: 2}},
	}
	updated, err := s.Update(ctx, v.ID, input, "actor")
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "After" || len(updated.Functions) != 2 || updated.Functions[0].Notes != f0.Notes ||
		updated.Functions[0].Inverted == nil || *updated.Functions[0].Inverted || updated.Functions[0].DirectionDependent != nil {
		t.Fatalf("bad combined update: %#v", updated)
	}
	if updated.Functions[0].Revision == f0.Revision {
		t.Fatal("changed function kept its revision")
	}
	beforeConflict, err := s.Get(ctx, v.ID)
	if err != nil {
		t.Fatal(err)
	}
	input.Name = "Must not persist"
	// First change would succeed, second has a stale revision; the whole transaction must roll back.
	input.FunctionChanges = []application.VehicleFunctionChange{
		{FunctionKey: "F3", ExpectedRevision: updated.Functions[1].Revision, Delete: true},
		{FunctionKey: "F0", ExpectedRevision: f0.Revision, Value: &value},
	}
	if _, err = s.Update(ctx, v.ID, input, "actor"); !errors.Is(err, application.ErrVehicleFunctionConflict) {
		t.Fatalf("want conflict: %v", err)
	}
	after, err := s.Get(ctx, v.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(beforeConflict, after) {
		t.Fatal("conflict changed vehicle or functions")
	}
	input.FunctionChanges = []application.VehicleFunctionChange{{FunctionKey: "F0", ExpectedRevision: updated.Functions[0].Revision, Delete: true}}
	input.Name = "After deletion"
	deleted, err := s.Update(ctx, v.ID, input, "actor")
	if err != nil {
		t.Fatal(err)
	}
	if len(deleted.Functions) != 1 || deleted.Functions[0].FunctionKey != "F3" {
		t.Fatal("delete was not committed")
	}
	input.FunctionChanges = nil
	kept, err := s.Update(ctx, v.ID, input, "actor")
	if err != nil || !reflect.DeepEqual(kept.Functions, deleted.Functions) {
		t.Fatalf("omitted changes touched functions: %v", err)
	}
}

func TestVehicleFunctionChangesRollbackDatabaseFailure(t *testing.T) {
	db := testDB(t)
	s := application.NewVehicleService(db)
	input := application.CreateVehicleInput{Manufacturer: "ESU", Name: "Before", Gauge: "TT", Category: "Lokomotive", Gattung: "Diesellok"}
	v, err := s.Create(t.Context(), input, "actor")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`CREATE TRIGGER fail_function BEFORE INSERT ON vehicle_functions WHEN NEW.function_key='F3' BEGIN SELECT RAISE(ABORT, 'test failure'); END`); err != nil {
		t.Fatal(err)
	}
	var auditsBefore int
	if err = db.QueryRow("SELECT COUNT(*) FROM audit_logs").Scan(&auditsBefore); err != nil {
		t.Fatal(err)
	}
	input.Name = "Must roll back"
	value := application.VehicleFunctionInput{Name: "Function", FlagsVersion: 2}
	input.FunctionChanges = []application.VehicleFunctionChange{{FunctionKey: "F1", Value: &value}, {FunctionKey: "F3", Value: &value}}
	if _, err = s.Update(t.Context(), v.ID, input, "actor"); err == nil {
		t.Fatal("expected database failure")
	}
	after, err := s.Get(t.Context(), v.ID)
	if err != nil {
		t.Fatal(err)
	}
	var auditsAfter int
	if err = db.QueryRow("SELECT COUNT(*) FROM audit_logs").Scan(&auditsAfter); err != nil {
		t.Fatal(err)
	}
	if after.Name != "Before" || len(after.Functions) != 0 || auditsAfter != auditsBefore {
		t.Fatal("partial update survived failure")
	}
	if _, err = s.Create(t.Context(), input, "actor"); err == nil {
		t.Fatal("expected create rollback")
	}
	var vehicles int
	if err = db.QueryRow("SELECT COUNT(*) FROM vehicles").Scan(&vehicles); err != nil || vehicles != 1 {
		t.Fatalf("partial creation survived: %d, %v", vehicles, err)
	}
}

func TestVehicleFunctionChangesRejectInvalidAndUnexpectedExistingKeys(t *testing.T) {
	db := testDB(t)
	s := application.NewVehicleService(db)
	input := application.CreateVehicleInput{Manufacturer: "ESU", Name: "Before", Gauge: "TT", Category: "Lokomotive", Gattung: "Diesellok"}
	v, err := s.Create(t.Context(), input, "actor")
	if err != nil {
		t.Fatal(err)
	}
	value := application.VehicleFunctionInput{Name: "Function", FlagsVersion: 2}
	for _, changes := range [][]application.VehicleFunctionChange{
		{{FunctionKey: "F32", Value: &value}}, {{FunctionKey: "F1"}},
		{{FunctionKey: "F1", Delete: true}}, {{FunctionKey: "F1", Delete: true, Value: &value}},
		{{FunctionKey: "F1", Value: &value}, {FunctionKey: "f01", Value: &value}},
		{{FunctionKey: "F1", Value: &value, ExpectedRevision: "invalid"}},
		{{FunctionKey: "F1", Value: &application.VehicleFunctionInput{Name: "Old flags"}}},
		{{FunctionKey: "F1", Value: &application.VehicleFunctionInput{Name: "Long", FlagsVersion: 2, Notes: strings.Repeat("ä", 501)}}},
	} {
		input.FunctionChanges = changes
		if _, err = s.Update(t.Context(), v.ID, input, "actor"); !errors.Is(err, application.ErrVehicleFunctionValidation) {
			t.Fatalf("invalid changes accepted: %v", err)
		}
	}
	if _, err = s.UpsertFunction(t.Context(), v.ID, "F1", value); err != nil {
		t.Fatal(err)
	}
	input.FunctionChanges = []application.VehicleFunctionChange{{FunctionKey: "F1", Value: &value}}
	if _, err = s.Update(t.Context(), v.ID, input, "actor"); !errors.Is(err, application.ErrVehicleFunctionConflict) {
		t.Fatal("new draft overwrote existing key")
	}
}
