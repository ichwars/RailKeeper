package application_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/infrastructure"
)

func TestOverviewValuationSeparatesCurrentListValueAndHistoricalPurchaseCost(t *testing.T) {
	db, err := infrastructure.OpenSQLite(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := infrastructure.Migrate(db, filepath.Join("..", "..", "migrations")); err != nil {
		t.Fatal(err)
	}

	execValuationFixture(t, db)
	valuation, err := application.NewOverviewValuationService(db).Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if valuation.VehicleListValue != "129.90" {
		t.Fatalf("vehicle list value = %q", valuation.VehicleListValue)
	}
	if valuation.VehiclePurchaseValue != "100.50" {
		t.Fatalf("vehicle purchase value = %q", valuation.VehiclePurchaseValue)
	}
	if valuation.AccessoryListValue != "120.00" {
		t.Fatalf("accessory list value = %q", valuation.AccessoryListValue)
	}
	if valuation.AccessoryPurchaseCost != "89.00" {
		t.Fatalf("accessory purchase cost = %q", valuation.AccessoryPurchaseCost)
	}
	if valuation.ExcludedForeignCurrencyPurchases != 2 {
		t.Fatalf("excluded foreign purchases = %d", valuation.ExcludedForeignCurrencyPurchases)
	}
}

func execValuationFixture(t *testing.T, db *sql.DB) {
	t.Helper()
	exec := func(query string, args ...any) {
		t.Helper()
		if _, err := db.Exec(query, args...); err != nil {
			t.Fatalf("execute valuation fixture: %v\n%s", err, query)
		}
	}

	for _, vehicle := range []struct {
		id, inventoryNumber, listPrice, purchasePrice string
	}{
		{"vehicle-1", "RK-VAL-000001", "100.00 €", "EUR 80.00"},
		{"vehicle-2", "RK-VAL-000002", "29,90", "20.50"},
		{"vehicle-3", "RK-VAL-000003", "invalid", ""},
	} {
		exec(`INSERT INTO vehicles(
  id, inventory_number, manufacturer, name, gauge, category, list_price, purchase_price, created_at, updated_at
) VALUES(?, ?, 'Test', 'Fahrzeug', 'TT', 'Lok', ?, ?, 'now', 'now')`,
			vehicle.id, vehicle.inventoryNumber, vehicle.listPrice, vehicle.purchasePrice)
	}

	exec(`INSERT INTO storage_locations(id, name, created_at, updated_at)
VALUES('location-1', 'Lager', 'now', 'now')`)
	for _, product := range []struct {
		id, inventoryNumber, strategy, listPrice string
		archived                                 int
	}{
		{"quantity", "RK-ART-VAL-000001", "quantity", "10.00", 1},
		{"individual", "RK-ART-VAL-000002", "individual", "20.00", 0},
		{"hybrid", "RK-ART-VAL-000003", "quantity_later_individual", "5.00", 0},
	} {
		trackingMode := "quantity"
		if product.strategy == "individual" {
			trackingMode = "individual"
		}
		exec(`INSERT INTO accessory_products(
  id, inventory_number, manufacturer, name, category, tracking_mode, article_type, subtype,
  gauges_json, package_quantity, stock_unit, minimum_stock, inventory_strategy, archived,
  list_price, created_at, updated_at
) VALUES(?, ?, 'Test', 'Artikel', 'other', ?, 'other', 'other:other', '[]', 1, 'piece', 0, ?, ?, ?, 'now', 'now')`,
			product.id, product.inventoryNumber, trackingMode,
			product.strategy, product.archived, product.listPrice)
	}
	exec(`INSERT INTO accessory_stock(product_id, location_id, quantity, updated_at) VALUES
  ('quantity', 'location-1', 3, 'now'),
  ('hybrid', 'location-1', 2, 'now')`)

	for _, purchase := range []struct {
		id, productID, unitPrice, currency string
		quantity                           int
	}{
		{"purchase-blank", "quantity", "3.50", "", 2},
		{"purchase-eur", "quantity", "4.00", "EUR", 3},
		{"purchase-linked", "individual", "50.00", "EUR", 1},
		{"purchase-hybrid-eur", "hybrid", "4.00", "EUR", 1},
		{"purchase-usd", "hybrid", "10.00", "USD", 4},
		{"purchase-chf", "hybrid", "20.00", "CHF", 1},
	} {
		exec(`INSERT INTO accessory_purchases(
  id, product_id, quantity, purchased_at, unit_price, currency, created_at, updated_at
) VALUES(?, ?, ?, '2026-08-16', ?, ?, 'now', 'now')`,
			purchase.id, purchase.productID, purchase.quantity, purchase.unitPrice, purchase.currency)
	}

	for _, asset := range []struct {
		id, productID, lifecycle, purchasePrice, purchaseID string
	}{
		{"asset-stored", "individual", "stored", "7.00", ""},
		{"asset-installed", "individual", "installed", "8.00", "purchase-linked"},
		{"asset-retired", "individual", "retired", "9.00", ""},
		{"asset-hybrid", "hybrid", "stored", "4.00", ""},
	} {
		exec(`INSERT INTO accessory_assets(
  id, product_id, condition_state, lifecycle_state, purchase_price, purchase_id, created_at, updated_at
) VALUES(?, ?, 'ready', ?, ?, NULLIF(?, ''), 'now', 'now')`,
			asset.id, asset.productID, asset.lifecycle, asset.purchasePrice, asset.purchaseID)
	}

	for _, installation := range []struct {
		id, productID, assetID string
		quantity               int
	}{
		{"installation-quantity", "quantity", "", 2},
		{"installation-individual", "individual", "asset-installed", 1},
		{"installation-hybrid", "hybrid", "", 3},
	} {
		exec(`INSERT INTO accessory_installations(
  id, product_id, asset_id, source_location_id, quantity, vehicle_id, condition_state,
  installed_by, installed_at
) VALUES(?, ?, NULLIF(?, ''), 'location-1', ?, 'vehicle-1', 'ready', 'tester', 'now')`,
			installation.id, installation.productID, installation.assetID, installation.quantity)
	}
}
