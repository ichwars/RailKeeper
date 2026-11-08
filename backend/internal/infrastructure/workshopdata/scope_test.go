package workshopdata_test

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"testing"

	"railkeeper/backend/internal/domain"
	"railkeeper/backend/internal/infrastructure/workshopdata"

	_ "modernc.org/sqlite"
)

func workshopStore(t *testing.T) (*sql.DB, *workshopdata.Store) {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })
	for _, schema := range []string{
		`CREATE TABLE vehicle_maintenance(id TEXT PRIMARY KEY, vehicle_id TEXT, kind TEXT, status TEXT,
condition_rating TEXT, due_date TEXT, completed_at TEXT, cost TEXT, notes TEXT, created_at TEXT, updated_at TEXT)`,
		`CREATE TABLE vehicle_spare_parts(id TEXT PRIMARY KEY, vehicle_id TEXT, article_number TEXT,
description TEXT, price TEXT, url TEXT, created_at TEXT, updated_at TEXT)`,
	} {
		if _, err := db.Exec(schema); err != nil {
			t.Fatal(err)
		}
	}
	return db, workshopdata.New(db)
}

func TestMutationsCannotReadChangeOrDeleteAnotherVehiclesRecords(t *testing.T) {
	_, store := workshopStore(t)
	ctx := t.Context()
	maintenance := domain.VehicleMaintenance{ID: "maintenance", VehicleID: "owner", Kind: "Wartung",
		Status: "geplant", Cost: "12,50", CreatedAt: "recorded", UpdatedAt: "recorded"}
	part := domain.VehicleSparePart{ID: "part", VehicleID: "owner", ArticleNumber: "ET-12 34",
		Price: "original price", URL: "https://example.invalid/original", CreatedAt: "recorded", UpdatedAt: "recorded"}
	if err := store.CreateMaintenance(ctx, maintenance); err != nil {
		t.Fatal(err)
	}
	if err := store.CreateSparePart(ctx, part); err != nil {
		t.Fatal(err)
	}
	for _, attempt := range []func() error{
		func() error { _, err := store.GetMaintenance(ctx, "other", maintenance.ID); return err },
		func() error {
			return store.UpdateMaintenance(ctx, "other", maintenance.ID, domain.VehicleMaintenanceInput{Notes: "replacement"}, "later")
		},
		func() error { return store.DeleteMaintenance(ctx, "other", maintenance.ID) },
		func() error { _, err := store.GetSparePart(ctx, "other", part.ID); return err },
		func() error {
			return store.UpdateSparePart(ctx, "other", part.ID, domain.VehicleSparePartInput{Price: "replacement"}, "later")
		},
		func() error { return store.DeleteSparePart(ctx, "other", part.ID) },
	} {
		if err := attempt(); !errors.Is(err, domain.ErrVehicleNotFound) {
			t.Fatalf("cross-vehicle request: %v", err)
		}
	}
	gotMaintenance, err := store.GetMaintenance(ctx, "owner", maintenance.ID)
	if err != nil || !reflect.DeepEqual(gotMaintenance, &maintenance) {
		t.Fatal("maintenance changed", gotMaintenance, err)
	}
	gotPart, err := store.GetSparePart(ctx, "owner", part.ID)
	if err != nil || !reflect.DeepEqual(gotPart, &part) {
		t.Fatal("spare part changed", gotPart, err)
	}
}

func TestStorageOutageNeverAcknowledgesAnEmptyListOrSuccessfulMutation(t *testing.T) {
	db, store := workshopStore(t)
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	for _, attempt := range []func() error{
		func() error {
			rows, err := store.ListMaintenance(ctx, "owner")
			if rows != nil {
				t.Error("outage returned maintenance rows")
			}
			return err
		},
		func() error { _, err := store.GetMaintenance(ctx, "owner", "maintenance"); return err },
		func() error {
			return store.CreateMaintenance(ctx, domain.VehicleMaintenance{ID: "maintenance", VehicleID: "owner"})
		},
		func() error {
			return store.UpdateMaintenance(ctx, "owner", "maintenance", domain.VehicleMaintenanceInput{}, "later")
		},
		func() error { return store.DeleteMaintenance(ctx, "owner", "maintenance") },
		func() error {
			rows, err := store.ListSpareParts(ctx, "owner")
			if rows != nil {
				t.Error("outage returned spare-part rows")
			}
			return err
		},
		func() error { _, err := store.GetSparePart(ctx, "owner", "part"); return err },
		func() error {
			return store.CreateSparePart(ctx, domain.VehicleSparePart{ID: "part", VehicleID: "owner"})
		},
		func() error {
			return store.UpdateSparePart(ctx, "owner", "part", domain.VehicleSparePartInput{}, "later")
		},
		func() error { return store.DeleteSparePart(ctx, "owner", "part") },
	} {
		if err := attempt(); err == nil || errors.Is(err, domain.ErrVehicleNotFound) {
			t.Fatalf("outage hidden: %v", err)
		}
	}
}
