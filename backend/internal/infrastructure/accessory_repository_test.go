package infrastructure_test

import (
	"database/sql"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/domain"
	"railkeeper/backend/internal/infrastructure"
)

func TestAccessoryServicePersistsCatalogAndPreventsLocationCycles(t *testing.T) {
	service, _ := testAccessoryService(t)
	ctx := t.Context()

	product, err := service.CreateProduct(ctx, application.CreateAccessoryProductInput{
		Manufacturer: "Tillig", ArticleNumber: "83125", Name: "Weiche rechts", Category: "Gleismaterial",
		TrackingMode: domain.AccessoryTrackingModeQuantity,
	}, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"Unnumbered A", "Unnumbered B"} {
		if _, err := service.CreateProduct(ctx, application.CreateAccessoryProductInput{
			Manufacturer: "Tillig", Name: name, Category: "Zubehör",
			TrackingMode: domain.AccessoryTrackingModeQuantity,
		}, "editor-1"); err != nil {
			t.Fatalf("blank article numbers must not conflict: %v", err)
		}
	}
	products, err := service.ListProducts(ctx, "83125")
	if err != nil {
		t.Fatal(err)
	}
	if len(products) != 1 || products[0].ID != product.ID {
		t.Fatalf("unexpected product search: %#v", products)
	}
	storedProduct, err := service.GetProduct(ctx, product.ID)
	if err != nil {
		t.Fatal(err)
	}
	if storedProduct.Name != "Weiche rechts" {
		t.Fatalf("unexpected stored product: %#v", storedProduct)
	}
	updatedProduct, err := service.UpdateProduct(ctx, product.ID, application.UpdateAccessoryProductInput{
		CreateAccessoryProductInput: application.CreateAccessoryProductInput{
			Manufacturer: "Tillig", ArticleNumber: "83125", Name: "Weiche rechts EW1", Category: "Gleismaterial",
			TrackingMode: domain.AccessoryTrackingModeQuantity,
		},
	}, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	if updatedProduct.Name != "Weiche rechts EW1" {
		t.Fatalf("unexpected updated product: %#v", updatedProduct)
	}

	root, err := service.CreateLocation(ctx, application.CreateStorageLocationInput{Name: "Werkstatt"}, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	drawer, err := service.CreateLocation(ctx, application.CreateStorageLocationInput{
		ParentID: root.ID, Name: "Schrank A",
	}, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	box, err := service.CreateLocation(ctx, application.CreateStorageLocationInput{
		ParentID: drawer.ID, Name: "Schublade 1",
	}, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.UpdateLocation(ctx, root.ID, application.UpdateStorageLocationInput{
		CreateStorageLocationInput: application.CreateStorageLocationInput{ParentID: box.ID, Name: root.Name},
	}, "editor-1"); !errors.Is(err, application.ErrAccessoryValidation) {
		t.Fatalf("expected descendant-parent cycle rejection, got %v", err)
	}
	if _, err := service.CreateLocation(ctx, application.CreateStorageLocationInput{
		ParentID: "missing", Name: "Orphan",
	}, "editor-1"); !errors.Is(err, application.ErrAccessoryNotFound) {
		t.Fatalf("expected missing parent error, got %v", err)
	}
	updatedDrawer, err := service.UpdateLocation(ctx, drawer.ID, application.UpdateStorageLocationInput{
		CreateStorageLocationInput: application.CreateStorageLocationInput{Name: "Schrank B"},
	}, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	if updatedDrawer.ParentID != "" || updatedDrawer.Name != "Schrank B" {
		t.Fatalf("unexpected updated location: %#v", updatedDrawer)
	}
	locations, err := service.ListLocations(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(locations) != 3 {
		t.Fatalf("unexpected locations: %#v", locations)
	}
}

func TestAccessoryServiceSeparatesQuantityAndIndividualInventory(t *testing.T) {
	service, _ := testAccessoryService(t)
	ctx := t.Context()
	location, err := service.CreateLocation(ctx, application.CreateStorageLocationInput{Name: "Lager"}, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	quantityProduct, err := service.CreateProduct(ctx, application.CreateAccessoryProductInput{
		Manufacturer: "Tillig", ArticleNumber: "83501", Name: "Schienenverbinder", Category: "Gleismaterial",
		TrackingMode: domain.AccessoryTrackingModeQuantity,
	}, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	individualProduct, err := service.CreateProduct(ctx, application.CreateAccessoryProductInput{
		Manufacturer: "Lenz", ArticleNumber: "LS150", Name: "Schaltdecoder", Category: "Decoder",
		TrackingMode: domain.AccessoryTrackingModeIndividual,
	}, "editor-1")
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.AdjustStock(ctx, quantityProduct.ID, application.StockAdjustmentInput{
		LocationID: location.ID, Delta: 5,
	}, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	stock, err := service.AdjustStock(ctx, quantityProduct.ID, application.StockAdjustmentInput{
		LocationID: location.ID, Delta: -2,
	}, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	if stock.TotalQuantity != 3 || len(stock.Locations) != 1 || stock.Locations[0].Quantity != 3 {
		t.Fatalf("unexpected stock summary: %#v", stock)
	}
	if _, err := service.UpdateProduct(ctx, quantityProduct.ID, application.UpdateAccessoryProductInput{
		CreateAccessoryProductInput: application.CreateAccessoryProductInput{
			Manufacturer: quantityProduct.Manufacturer, ArticleNumber: quantityProduct.ArticleNumber,
			Name: quantityProduct.Name, Category: quantityProduct.Category,
			TrackingMode: domain.AccessoryTrackingModeIndividual,
		},
	}, "editor-1"); !errors.Is(err, application.ErrAccessoryConflict) {
		t.Fatalf("expected tracking mode conflict for stocked product, got %v", err)
	}
	if _, err := service.AdjustStock(ctx, quantityProduct.ID, application.StockAdjustmentInput{
		LocationID: location.ID, Delta: -4,
	}, "editor-1"); !errors.Is(err, application.ErrAccessoryInsufficientStock) {
		t.Fatalf("expected insufficient stock error, got %v", err)
	}
	if _, err := service.AdjustStock(ctx, individualProduct.ID, application.StockAdjustmentInput{
		LocationID: location.ID, Delta: 1,
	}, "editor-1"); !errors.Is(err, application.ErrAccessoryTrackingMode) {
		t.Fatalf("expected individual stock tracking rejection, got %v", err)
	}
	if _, err := service.CreateAsset(ctx, quantityProduct.ID, application.CreateAccessoryAssetInput{
		InventoryNumber: "Z-0001",
	}, "editor-1"); !errors.Is(err, application.ErrAccessoryTrackingMode) {
		t.Fatalf("expected quantity asset rejection, got %v", err)
	}

	asset, err := service.CreateAsset(ctx, individualProduct.ID, application.CreateAccessoryAssetInput{
		InventoryNumber: "Z-0001", SerialNumber: "ABC", StorageLocationID: location.ID,
		Condition: domain.AccessoryConditionReady, Lifecycle: domain.AccessoryLifecycleStored,
		PurchasePrice: "49.95",
	}, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.CreateAsset(ctx, individualProduct.ID, application.CreateAccessoryAssetInput{
		InventoryNumber: "z-0001",
	}, "editor-1"); !errors.Is(err, application.ErrAccessoryConflict) {
		t.Fatalf("expected inventory number conflict, got %v", err)
	}
	if _, err := service.UpdateProduct(ctx, individualProduct.ID, application.UpdateAccessoryProductInput{
		CreateAccessoryProductInput: application.CreateAccessoryProductInput{
			Manufacturer: individualProduct.Manufacturer, ArticleNumber: individualProduct.ArticleNumber,
			Name: individualProduct.Name, Category: individualProduct.Category,
			TrackingMode: domain.AccessoryTrackingModeQuantity,
		},
	}, "editor-1"); !errors.Is(err, application.ErrAccessoryConflict) {
		t.Fatalf("expected tracking mode conflict for product with assets, got %v", err)
	}
	asset, err = service.UpdateAsset(ctx, asset.ID, application.UpdateAccessoryAssetInput{
		CreateAccessoryAssetInput: application.CreateAccessoryAssetInput{
			InventoryNumber: asset.InventoryNumber, SerialNumber: asset.SerialNumber,
			Condition: domain.AccessoryConditionMaintenanceDue, Lifecycle: domain.AccessoryLifecycleMaintenance,
			StorageLocationID: location.ID, PurchasePrice: asset.PurchasePrice,
		},
	}, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	if asset.Condition != domain.AccessoryConditionMaintenanceDue ||
		asset.Lifecycle != domain.AccessoryLifecycleMaintenance {
		t.Fatalf("unexpected updated asset: %#v", asset)
	}
	assets, err := service.ListAssets(ctx, individualProduct.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(assets) != 1 || assets[0].ID != asset.ID {
		t.Fatalf("unexpected assets: %#v", assets)
	}
}

func TestAccessoryServiceWritesMinimalAuditDetails(t *testing.T) {
	service, db := testAccessoryService(t)
	ctx := t.Context()
	location, err := service.CreateLocation(ctx, application.CreateStorageLocationInput{Name: "Audit Lager"}, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	product, err := service.CreateProduct(ctx, application.CreateAccessoryProductInput{
		Manufacturer: "Tillig", Name: "Audit product", Category: "Track",
		TrackingMode: domain.AccessoryTrackingModeIndividual,
	}, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	asset, err := service.CreateAsset(ctx, product.ID, application.CreateAccessoryAssetInput{
		InventoryNumber: "AUDIT-1", StorageLocationID: location.ID, Notes: "private workshop note",
	}, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.UpdateAsset(ctx, asset.ID, application.UpdateAccessoryAssetInput{
		CreateAccessoryAssetInput: application.CreateAccessoryAssetInput{
			InventoryNumber: asset.InventoryNumber, StorageLocationID: location.ID,
			Condition: domain.AccessoryConditionDefective, Lifecycle: domain.AccessoryLifecycleMaintenance,
			Notes: "another private note",
		},
	}, "editor-1"); err != nil {
		t.Fatal(err)
	}

	rows, err := db.QueryContext(ctx, `SELECT action, COALESCE(details_json, '{}') FROM audit_logs ORDER BY rowid`)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	count := 0
	for rows.Next() {
		var action, details string
		if err := rows.Scan(&action, &details); err != nil {
			t.Fatal(err)
		}
		if strings.Contains(details, "private") {
			t.Fatalf("audit action %s leaked free-text notes: %s", action, details)
		}
		count++
	}
	if count != 4 {
		t.Fatalf("expected 4 audit entries, got %d", count)
	}
}

func testAccessoryService(t *testing.T) (*application.AccessoryService, *sql.DB) {
	t.Helper()
	db, err := infrastructure.OpenSQLite(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := infrastructure.Migrate(db, filepath.Join("..", "..", "migrations")); err != nil {
		t.Fatal(err)
	}
	return application.NewAccessoryService(infrastructure.NewAccessoryRepository(db)), db
}
