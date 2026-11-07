package application_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/infrastructure/valuationdata"
	"railkeeper/backend/internal/valuation"
)

func TestValuationBackupRestoreAndLegacyCompatibility(t *testing.T) {
	dir := t.TempDir()
	db := backupTestDB(t, dir)
	user, err := application.NewAuthService(db).CreateUser(t.Context(), "", application.CreateUserInput{Username: "source", Password: "synthetic-only-password", Roles: []string{"Editor"}})
	if err != nil {
		t.Fatal(err)
	}
	vehicle, err := application.NewVehicleService(db).Create(t.Context(), application.CreateVehicleInput{Name: "History", Manufacturer: "Synthetic", Gauge: "H0", Category: "Lokomotive", Gattung: "Diesellok"}, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	values := valuation.New(valuationdata.NewStore(db), func() time.Time { return time.Date(2026, 11, 7, 12, 0, 0, 0, time.UTC) })
	input := valuation.Input{ID: fmt.Sprintf("%032x", 1), Amount: "0.123456", Currency: "CHF", ValuedOn: "2026-01-01", Source: "Manual original", Notes: "Original provenance"}
	if _, err := values.Record(t.Context(), vehicle.ID, input, user.ID); err != nil {
		t.Fatal(err)
	}
	input.ID, input.ReplacesID, input.Amount = fmt.Sprintf("%032x", 2), input.ID, "123.00"
	if _, err := values.Record(t.Context(), vehicle.ID, input, user.ID); err != nil {
		t.Fatal(err)
	}
	backup := application.NewBackupService(db, dir)
	doc, err := backup.Export(t.Context())
	if err != nil || doc.Version != 28 || len(doc.Tables["vehicle_valuations"]) != 2 {
		t.Fatalf("backup missing valuations: %v", err)
	}
	validation, err := backup.Validate(t.Context(), doc)
	if err != nil || !validation.Compatible {
		t.Fatalf("valid history rejected: %+v %v", validation, err)
	}
	targetDir := t.TempDir()
	targetDB := backupTestDB(t, targetDir)
	targetAuth := application.NewAuthService(targetDB)
	targetUser, err := targetAuth.CreateUser(t.Context(), "", application.CreateUserInput{Username: "target", Password: "other-synthetic-password", Roles: []string{"Admin"}})
	if err != nil {
		t.Fatal(err)
	}
	target := application.NewBackupService(targetDB, targetDir)
	if _, err := target.Import(t.Context(), doc); err != nil {
		t.Fatal(err)
	}
	roundTrip, err := target.Export(t.Context())
	if err != nil || !reflect.DeepEqual(roundTrip.Tables["vehicle_valuations"], doc.Tables["vehicle_valuations"]) {
		t.Fatal("history or exact amounts changed during restore")
	}
	if _, err := targetAuth.Login(t.Context(), application.LoginInput{Username: "target", Password: "other-synthetic-password"}); err != nil {
		t.Fatalf("local target authentication replaced (%s): %v", targetUser.ID, err)
	}
	clone := func() *application.BackupDocument {
		raw, err := json.Marshal(doc)
		if err != nil {
			t.Fatal(err)
		}
		var copy application.BackupDocument
		if err := json.Unmarshal(raw, &copy); err != nil {
			t.Fatal(err)
		}
		return &copy
	}
	for _, mutate := range []func(*application.BackupDocument){
		func(d *application.BackupDocument) { delete(d.Tables, "vehicle_valuations") },
		func(d *application.BackupDocument) { d.Tables["vehicle_valuations"][0]["amount"] = "1e9" },
		func(d *application.BackupDocument) { d.Tables["vehicle_valuations"][0]["currency"] = "eur" },
		func(d *application.BackupDocument) { d.Tables["vehicle_valuations"][0]["vehicle_id"] = "missing" },
		func(d *application.BackupDocument) { d.Tables["vehicle_valuations"][0]["unknown_column"] = "ignored?" },
		func(d *application.BackupDocument) {
			d.Tables["vehicle_valuations"][0]["replaces_id"] = d.Tables["vehicle_valuations"][1]["id"]
		},
		func(d *application.BackupDocument) { d.Version = 27 },
	} {
		bad := clone()
		mutate(bad)
		result, err := target.Validate(t.Context(), bad)
		if err != nil || result.Compatible {
			t.Fatalf("bad history accepted: %+v %v", result, err)
		}
		if _, err := target.Import(t.Context(), bad); err == nil {
			t.Fatal("invalid history restored")
		}
		after, err := target.Export(t.Context())
		if err != nil || !reflect.DeepEqual(after.Tables["vehicle_valuations"], doc.Tables["vehicle_valuations"]) {
			t.Fatal("rejected restore changed history")
		}
	}
	legacy := clone()
	setLegacyBackupVersion(legacy, 27)
	if _, err := target.Import(t.Context(), legacy); err != nil {
		t.Fatalf("format27 compatibility: %v", err)
	}
	var count int
	if err := targetDB.QueryRow(`SELECT COUNT(*) FROM vehicle_valuations`).Scan(&count); err != nil || count != 0 {
		t.Fatal("legacy restore fabricated history")
	}
}
