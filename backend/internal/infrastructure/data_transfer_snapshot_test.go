package infrastructure_test

import (
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/infrastructure"
)

func TestDataTransferSnapshotReadsSelectedAreasInStableOrder(t *testing.T) {
	db := testDB(t)
	for _, statement := range []string{
		`INSERT INTO vehicles(id, inventory_number, manufacturer, name, gauge, created_at, updated_at)
         VALUES('v-2', 'RK-010', 'Roco', 'Zweite', 'H0', '2026-08-20T10:00:00Z', '2026-08-20T10:00:00Z')`,
		`INSERT INTO vehicles(id, inventory_number, manufacturer, name, gauge, created_at, updated_at)
         VALUES('v-1', 'RK-002', 'Märklin', 'Erste', 'H0', '2026-08-20T10:00:00Z', '2026-08-20T10:00:00Z')`,
		`INSERT INTO exhibition_lists(id, designation, list_date, locked, created_at, updated_at)
         VALUES('list-2', 'Zweite', '2026-08-21', 0, '2026-08-20T10:00:00Z', '2026-08-20T10:00:00Z')`,
		`INSERT INTO exhibition_lists(id, designation, list_date, locked, created_at, updated_at)
         VALUES('list-1', 'Erste', '2026-08-20', 0, '2026-08-20T10:00:00Z', '2026-08-20T10:00:00Z')`,
	} {
		if _, err := db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	repository := infrastructure.NewDataTransferRepository(db)
	snapshot, err := repository.Snapshot(t.Context(), []application.TransferArea{
		application.TransferVehicles, application.TransferExhibitionLists,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Vehicles) != 2 || snapshot.Vehicles[0].InventoryNumber != "RK-002" {
		t.Fatalf("vehicles are not stable: %#v", snapshot.Vehicles)
	}
	if snapshot.Accessories != nil {
		t.Fatalf("unselected accessories were queried: %#v", snapshot.Accessories)
	}
	if len(snapshot.ExhibitionLists) != 2 || snapshot.ExhibitionLists[0].ID != "list-1" {
		t.Fatalf("exhibition lists are not stable: %#v", snapshot.ExhibitionLists)
	}
}

func TestDataTransferSnapshotAccessoryIncludesCurrentStateOnly(t *testing.T) {
	db := testDB(t)
	for _, statement := range []string{
		`INSERT INTO storage_locations(id, name, created_at, updated_at)
         VALUES('location-1', 'Schrank', '2026-08-20T10:00:00Z', '2026-08-20T10:00:00Z')`,
		`INSERT INTO accessory_products(id, inventory_number, manufacturer, article_number, name, category,
           tracking_mode, inventory_strategy, created_at, updated_at)
         VALUES('product-1', 'RK-ART-1', 'Viessmann', '4011', 'Signal', 'signal', 'individual', 'individual',
           '2026-08-20T10:00:00Z', '2026-08-20T10:00:00Z')`,
		`INSERT INTO accessory_stock(product_id, location_id, quantity, updated_at)
         VALUES('product-1', 'location-1', 2, '2026-08-20T10:00:00Z')`,
		`INSERT INTO accessory_assets(id, product_id, inventory_number, serial_number, condition_state,
           lifecycle_state, storage_location_id, created_at, updated_at)
         VALUES('asset-1', 'product-1', 'RK-ASSET-1', 'S1', 'ready', 'stored', 'location-1',
           '2026-08-20T10:00:00Z', '2026-08-20T10:00:00Z')`,
	} {
		if _, err := db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	repository := infrastructure.NewDataTransferRepository(db)
	snapshot, err := repository.Snapshot(t.Context(), []application.TransferArea{application.TransferAccessories})
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Accessories) != 1 || len(snapshot.Accessories[0].Stock) != 1 ||
		len(snapshot.Accessories[0].Assets) != 1 {
		t.Fatalf("unexpected accessory snapshot: %#v", snapshot.Accessories)
	}
	if snapshot.Accessories[0].Stock[0].LocationName != "Schrank" ||
		snapshot.Accessories[0].Assets[0].InventoryNumber != "RK-ASSET-1" {
		t.Fatalf("unexpected accessory current state: %#v", snapshot.Accessories[0])
	}
}
