package infrastructure_test

import (
	"errors"
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/domain"
)

func TestAccessoryStockRejectsBrokenStorageLocationChains(t *testing.T) {
	service, db := testAccessoryService(t)
	ctx := t.Context()
	root, err := service.CreateLocation(ctx, application.CreateStorageLocationInput{Name: "Root"}, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	child, err := service.CreateLocation(ctx, application.CreateStorageLocationInput{
		ParentID: root.ID, Name: "Child",
	}, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	product, err := service.CreateProduct(ctx, application.CreateAccessoryProductInput{
		Manufacturer: "Tillig", Name: "Test", Category: "Track",
		TrackingMode: domain.AccessoryTrackingModeQuantity,
	}, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.AdjustStock(ctx, product.ID, application.StockAdjustmentInput{
		LocationID: child.ID, Delta: 1,
	}, "editor-1"); err != nil {
		t.Fatalf("valid root/child chain was rejected: %v", err)
	}

	if _, err := db.ExecContext(ctx, `PRAGMA foreign_keys=OFF`); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if _, err := db.ExecContext(ctx, `PRAGMA foreign_keys=ON`); err != nil {
			t.Errorf("restore foreign keys: %v", err)
		}
	}()
	if _, err := db.ExecContext(ctx, `
INSERT INTO storage_locations(id, parent_id, name, archived, created_at, updated_at) VALUES
  ('broken-missing', 'missing-parent', 'Broken missing', 0, 'now', 'now'),
  ('broken-cycle-a', 'broken-cycle-b', 'Broken cycle A', 0, 'now', 'now'),
  ('broken-cycle-b', 'broken-cycle-a', 'Broken cycle B', 0, 'now', 'now')`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `PRAGMA foreign_keys=ON`); err != nil {
		t.Fatal(err)
	}
	var foreignKeys int
	if err := db.QueryRowContext(ctx, `PRAGMA foreign_keys`).Scan(&foreignKeys); err != nil {
		t.Fatal(err)
	}
	if foreignKeys != 1 {
		t.Fatal("foreign key enforcement was not restored")
	}

	for _, locationID := range []string{"broken-missing", "broken-cycle-a"} {
		_, err := service.AdjustStock(ctx, product.ID, application.StockAdjustmentInput{
			LocationID: locationID, Delta: 1,
		}, "editor-1")
		if !errors.Is(err, application.ErrAccessoryConflict) {
			t.Fatalf("%s: expected broken-chain conflict, got %v", locationID, err)
		}
	}
}
