package application_test

import (
	"context"
	"encoding/json"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestFunctionFlagsUnknownIndependentAndLegacyClients(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()
	s := application.NewVehicleService(db)
	v, err := s.Create(ctx, application.CreateVehicleInput{Manufacturer: "ESU", Name: "Flags", Gauge: "TT", Category: "Lokomotive", Gattung: "Diesellok"}, "actor")
	if err != nil {
		t.Fatal(err)
	}
	legacy, err := s.UpsertFunction(ctx, v.ID, "F1", application.VehicleFunctionInput{Name: "Legacy", DirectionDependent: boolPtr(true)})
	if err != nil {
		t.Fatal(err)
	}
	if legacy.DirectionDependent != nil || legacy.Inverted != nil || legacy.LegacyDirectionDependent == nil || !*legacy.LegacyDirectionDependent {
		t.Fatalf("guessed legacy flags: %#v", legacy)
	}
	for _, direction := range []*bool{nil, boolPtr(false), boolPtr(true)} {
		for _, inverted := range []*bool{nil, boolPtr(false), boolPtr(true)} {
			f, err := s.UpsertFunction(ctx, v.ID, "F1", application.VehicleFunctionInput{Name: "Independent", FlagsVersion: 2, DirectionDependent: direction, Inverted: inverted})
			if err != nil {
				t.Fatal(err)
			}
			same := func(a, b *bool) bool { return a == nil && b == nil || a != nil && b != nil && *a == *b }
			if !same(f.DirectionDependent, direction) || !same(f.Inverted, inverted) || f.LegacyDirectionDependent == nil {
				t.Fatalf("flags did not round trip: %#v", f)
			}
			list, err := s.ListFunctions(ctx, v.ID)
			if err != nil || !same(list[0].Inverted, inverted) {
				t.Fatalf("list: %#v %v", list, err)
			}
		}
	}
	// A stale browser must not reinterpret its mislabelled checkbox or erase reviewed flags.
	f, err := s.UpsertFunction(ctx, v.ID, "F1", application.VehicleFunctionInput{Name: "Old client", DirectionDependent: boolPtr(false)})
	if err != nil {
		t.Fatal(err)
	}
	if f.Inverted == nil || !*f.Inverted || f.DirectionDependent == nil || !*f.DirectionDependent || f.LegacyDirectionDependent == nil || !*f.LegacyDirectionDependent {
		t.Fatalf("old client replaced confirmed meanings: %#v", f)
	}
	f, err = s.UpsertFunction(ctx, v.ID, "F1", application.VehicleFunctionInput{Name: "Unknown", FlagsVersion: 2})
	if err != nil {
		t.Fatal(err)
	}
	encoded, _ := json.Marshal(f)
	var fields map[string]any
	_ = json.Unmarshal(encoded, &fields)
	if value, exists := fields["inverted"]; !exists || value != nil {
		t.Fatalf("unknown must be explicit null: %s", encoded)
	}
	if f.DirectionDependent != nil || f.Inverted != nil {
		t.Fatal("unknown became false")
	}
	if _, err = s.UpsertFunction(ctx, v.ID, "F2", application.VehicleFunctionInput{FlagsVersion: 3}); err == nil {
		t.Fatal("accepted unknown semantics version")
	}
}

func TestFunctionFlagBackupCurrentAndLegacyRestore(t *testing.T) {
	dir := t.TempDir()
	db := backupTestDB(t, dir)
	ctx := context.Background()
	s := application.NewVehicleService(db)
	v, err := s.Create(ctx, application.CreateVehicleInput{Manufacturer: "ESU", Name: "Backup", Gauge: "TT", Category: "Lokomotive", Gattung: "Diesellok"}, "actor")
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.UpsertFunction(ctx, v.ID, "F1", application.VehicleFunctionInput{FlagsVersion: 2, DirectionDependent: boolPtr(false), Inverted: boolPtr(true)})
	if err != nil {
		t.Fatal(err)
	}
	backup := application.NewBackupService(db, dir)
	doc, err := backup.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Version != 24 {
		t.Fatal(doc.Version)
	}
	if _, err = backup.Import(ctx, doc); err != nil {
		t.Fatal(err)
	}
	f, _ := s.GetFunction(ctx, v.ID, "F1")
	if f.Inverted == nil || !*f.Inverted || f.DirectionDependent == nil || *f.DirectionDependent {
		t.Fatalf("current restore changed flags %#v", f)
	}
	doc.Version = 23
	for _, row := range doc.Tables["vehicle_functions"] {
		row["direction_dependent"] = int64(1)
		delete(row, "inverted")
		delete(row, "legacy_direction_dependent")
	}
	if _, err = backup.Import(ctx, doc); err != nil {
		t.Fatal(err)
	}
	f, _ = s.GetFunction(ctx, v.ID, "F1")
	if f.Inverted != nil || f.DirectionDependent != nil || f.LegacyDirectionDependent == nil || !*f.LegacyDirectionDependent {
		t.Fatalf("legacy restore guessed flags %#v", f)
	}
	doc.Version = 24
	doc.Tables["vehicle_functions"][0]["inverted"] = map[string]any{"bad": true}
	validation, err := backup.Validate(ctx, doc)
	if err != nil {
		t.Fatal(err)
	}
	if validation.Compatible {
		t.Fatal("invalid flag backup accepted")
	}
}
