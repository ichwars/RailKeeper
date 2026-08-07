package infrastructure_test

import (
	"database/sql"
	"errors"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/domain"
	"railkeeper/backend/internal/infrastructure"
)

func TestAccessoryArticlePersistsFullProductAndAttributes(t *testing.T) {
	service, _ := testAccessoryService(t)
	text := "TT Modellgleis"
	length := 166.0
	unit := "mm"
	input := application.CreateAccessoryProductInput{
		Manufacturer: "Tillig", ArticleNumber: "83101", Name: "Straight track", Category: "Track",
		Description: "Description", EAN: "4012500831012", ManufacturerStatus: "available",
		ArticleType: domain.AccessoryArticleTrack, Subtype: "track:straight", Gauges: []string{"TT"}, Scale: "1:120",
		PackageQuantity: 2, StockUnit: "piece", MinimumStock: 3,
		InventoryStrategy: domain.AccessoryInventoryQuantityLaterIndividual,
		ManufacturerURL:   "https://example.test/tillig", ProductURL: "https://example.test/product",
		AlternativeNumbers: []string{"83101-A"}, Keywords: []string{"track"},
		CompatibilityNotes: "compatible", InternalNotes: "internal", Archived: true,
		Attributes: []domain.AccessoryAttributeValue{
			{Key: "trackSystem", Kind: domain.AccessoryAttributeText, TextValue: &text},
			{Key: "lengthMm", Kind: domain.AccessoryAttributeNumber, NumberValue: &length, Unit: &unit},
		},
	}
	created, err := service.CreateProduct(t.Context(), input, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := service.GetProduct(t.Context(), created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(loaded.Gauges, input.Gauges) ||
		!reflect.DeepEqual(loaded.AlternativeNumbers, input.AlternativeNumbers) ||
		!reflect.DeepEqual(loaded.Keywords, input.Keywords) || !reflect.DeepEqual(loaded.Attributes, input.Attributes) ||
		loaded.EAN != input.EAN || loaded.ManufacturerStatus != input.ManufacturerStatus ||
		loaded.ArticleType != input.ArticleType || loaded.Subtype != input.Subtype || loaded.Scale != input.Scale ||
		loaded.PackageQuantity != input.PackageQuantity || loaded.StockUnit != input.StockUnit ||
		loaded.MinimumStock != input.MinimumStock || loaded.InventoryStrategy != input.InventoryStrategy ||
		loaded.ManufacturerURL != input.ManufacturerURL || loaded.ProductURL != input.ProductURL ||
		loaded.CompatibilityNotes != input.CompatibilityNotes || loaded.InternalNotes != input.InternalNotes ||
		!loaded.Archived {
		t.Fatalf("full product did not round trip:\ninput=%#v\nloaded=%#v", input, loaded)
	}
	if loaded.Attributes[1].Unit == nil || *loaded.Attributes[1].Unit != "mm" {
		t.Fatalf("create/get lost numeric attribute unit: %#v", loaded.Attributes[1])
	}

	input.Name = "Updated track"
	input.Archived = false
	input.Attributes = input.Attributes[1:]
	updatedUnit := "cm"
	input.Attributes[0].Unit = &updatedUnit
	updated, err := service.UpdateProduct(t.Context(), created.ID,
		application.UpdateAccessoryProductInput{CreateAccessoryProductInput: input}, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != input.Name || updated.Archived || !reflect.DeepEqual(updated.Attributes, input.Attributes) {
		t.Fatalf("full update did not round trip: %#v", updated)
	}
	if updated.Attributes[0].Unit == nil || *updated.Attributes[0].Unit != "cm" {
		t.Fatalf("update/get lost numeric attribute unit: %#v", updated.Attributes[0])
	}
}

func TestAccessoryArticleDuplicateLookupExcludesCurrentAndAllowsVariants(t *testing.T) {
	service, _ := testAccessoryService(t)
	create := func(name, subtype string) *application.AccessoryProduct {
		product, err := service.CreateProduct(t.Context(), application.CreateAccessoryProductInput{
			Manufacturer: "Tillig", ArticleNumber: "83125", Name: name, Category: "Track",
			ArticleType: domain.AccessoryArticleTrack, Subtype: subtype, PackageQuantity: 1,
			StockUnit: "piece", InventoryStrategy: domain.AccessoryInventoryQuantity,
		}, "editor-1")
		if err != nil {
			t.Fatal(err)
		}
		return product
	}
	first := create("Left turnout", "track:turnout")
	second := create("Right turnout", "track:turnout")
	result, err := service.CheckDuplicateProducts(t.Context(), application.AccessoryDuplicateCheckInput{
		Manufacturer: " tillig ", ArticleNumber: " 83125 ", ExcludeID: first.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Candidates) != 1 || result.Candidates[0].ID != second.ID {
		t.Fatalf("unexpected candidates: %#v", result.Candidates)
	}
	first.Name = "Updated variant"
	if _, err := service.UpdateProduct(t.Context(), first.ID, application.UpdateAccessoryProductInput{
		CreateAccessoryProductInput: application.CreateAccessoryProductInput{
			Manufacturer: first.Manufacturer, ArticleNumber: first.ArticleNumber, Name: first.Name, Category: first.Category,
			ArticleType: first.ArticleType, Subtype: first.Subtype, PackageQuantity: first.PackageQuantity,
			StockUnit: first.StockUnit, InventoryStrategy: first.InventoryStrategy,
		},
	}, "editor-1"); err != nil {
		t.Fatalf("duplicate variant blocked update: %v", err)
	}
}

func TestAccessoryArticleCatalogueSearchFiltersSortsAndAggregates(t *testing.T) {
	service, db := testAccessoryService(t)
	ctx := t.Context()
	alpha, err := service.CreateLocation(ctx, application.CreateStorageLocationInput{Name: "Alpha shelf"}, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	beta, err := service.CreateLocation(ctx, application.CreateStorageLocationInput{Name: "Beta shelf"}, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	create := func(input application.CreateAccessoryProductInput) *application.AccessoryProduct {
		product, err := service.CreateProduct(ctx, input, "editor-1")
		if err != nil {
			t.Fatal(err)
		}
		return product
	}
	trackSystem := "TT Modellgleis"
	track := create(application.CreateAccessoryProductInput{
		Manufacturer: "Z-Tillig", ArticleNumber: "83125", EAN: "4012500831258", Name: "Right turnout",
		Category: "Track", ArticleType: domain.AccessoryArticleTrack, Subtype: "track:turnout", Gauges: []string{"TT"},
		PackageQuantity: 1, StockUnit: "piece", InventoryStrategy: domain.AccessoryInventoryQuantity,
		Attributes: []domain.AccessoryAttributeValue{
			{Key: "trackSystem", Kind: domain.AccessoryAttributeText, TextValue: &trackSystem},
		},
	})
	signal := create(application.CreateAccessoryProductInput{
		Manufacturer: "Viessmann", ArticleNumber: "4011", EAN: "4026602040110", Name: "Block signal",
		Category: "Signal", ArticleType: domain.AccessoryArticleSignal, Subtype: "signal:block", Gauges: []string{"H0"},
		PackageQuantity: 1, StockUnit: "piece", InventoryStrategy: domain.AccessoryInventoryIndividual,
	})
	archived := create(application.CreateAccessoryProductInput{
		Manufacturer: "Noch", ArticleNumber: "60870", Name: "Old grass", Category: "Landscape",
		ArticleType: domain.AccessoryArticleLandscapeConsumable, Subtype: "landscape_consumable:grass",
		PackageQuantity: 1, StockUnit: "gram", InventoryStrategy: domain.AccessoryInventoryQuantity, Archived: true,
	})

	mustExec := func(query string, args ...any) {
		t.Helper()
		if _, err := db.ExecContext(ctx, query, args...); err != nil {
			t.Fatalf("fixture SQL failed: %v\n%s", err, query)
		}
	}
	mustExec(`INSERT INTO accessory_stock(product_id, location_id, quantity, updated_at) VALUES
		(?, ?, 8, '2026-01-01'), (?, ?, 2, '2026-01-01')`, track.ID, alpha.ID, track.ID, beta.ID)
	mustExec(`INSERT INTO accessory_assets(
		id, product_id, inventory_number, condition_state, lifecycle_state, storage_location_id, created_at, updated_at
	) VALUES
		('signal-stored', ?, 'SIG-1', 'maintenance_due', 'stored', ?, '2026-01-01', '2026-01-01'),
		('signal-reserved', ?, 'SIG-2', 'ready', 'reserved', ?, '2026-01-01', '2026-01-01'),
		('signal-installed', ?, 'SIG-3', 'defective', 'installed', NULL, '2026-01-01', '2026-01-01')`,
		signal.ID, beta.ID, signal.ID, beta.ID, signal.ID)
	mustExec(`INSERT INTO layouts(id, name, kind, gauge, scale, created_at, updated_at)
		VALUES('catalog-layout', 'Catalog layout', 'private', 'TT', '1:120', '2026-01-01', '2026-01-01')`)
	mustExec(`INSERT INTO accessory_reservations(
		id, product_id, asset_id, location_id, quantity, layout_id, status, created_by, created_at, updated_at
	) VALUES
		('track-reservation', ?, NULL, ?, 3, 'catalog-layout', 'active', 'editor-1', '2026-01-01', '2026-01-01'),
		('signal-reservation', ?, 'signal-reserved', ?, 1, 'catalog-layout', 'active', 'editor-1', '2026-01-01', '2026-01-01')`,
		track.ID, alpha.ID, signal.ID, beta.ID)
	mustExec(`INSERT INTO accessory_installations(
		id, product_id, asset_id, source_location_id, quantity, layout_id, condition_state, installed_by, installed_at
	) VALUES
		('track-installation', ?, NULL, ?, 4, 'catalog-layout', 'ready', 'editor-1', '2026-01-01'),
		('signal-installation', ?, 'signal-installed', ?, 1, 'catalog-layout', 'defective', 'editor-1', '2026-01-01')`,
		track.ID, beta.ID, signal.ID, beta.ID)
	mustExec(`UPDATE accessory_products SET updated_at=CASE id WHEN ? THEN '2026-01-01' ELSE '2026-02-01' END
		WHERE id IN (?, ?)`, track.ID, track.ID, signal.ID)

	assertIDs := func(query application.AccessoryArticleListQuery, want ...string) *application.AccessoryArticleListResult {
		t.Helper()
		result, err := service.ListArticles(ctx, query)
		if err != nil {
			t.Fatal(err)
		}
		got := make([]string, len(result.Items))
		for index := range result.Items {
			got[index] = result.Items[index].ID
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("ListArticles(%#v) IDs = %#v, want %#v", query, got, want)
		}
		return result
	}

	for _, search := range []string{"Tillig", "83125", "4012500831258", "turnout"} {
		assertIDs(application.AccessoryArticleListQuery{Query: search}, track.ID)
	}
	assertIDs(application.AccessoryArticleListQuery{Manufacturer: "Viessmann"}, signal.ID)
	assertIDs(application.AccessoryArticleListQuery{ArticleTypes: []domain.AccessoryArticleType{domain.AccessoryArticleTrack}}, track.ID)
	assertIDs(application.AccessoryArticleListQuery{Gauges: []string{"H0"}}, signal.ID)
	assertIDs(application.AccessoryArticleListQuery{LocationID: alpha.ID}, track.ID)
	assertIDs(application.AccessoryArticleListQuery{Statuses: []application.AccessoryArticleStatus{
		application.AccessoryArticleAvailable,
	}}, signal.ID, track.ID)
	assertIDs(application.AccessoryArticleListQuery{Statuses: []application.AccessoryArticleStatus{
		application.AccessoryArticleReserved,
	}}, signal.ID, track.ID)
	assertIDs(application.AccessoryArticleListQuery{Statuses: []application.AccessoryArticleStatus{
		application.AccessoryArticleInstalled,
	}}, signal.ID, track.ID)
	assertIDs(application.AccessoryArticleListQuery{Statuses: []application.AccessoryArticleStatus{
		application.AccessoryArticleReserved, application.AccessoryArticleInstalled,
	}}, signal.ID, track.ID)
	assertIDs(application.AccessoryArticleListQuery{Statuses: []application.AccessoryArticleStatus{
		application.AccessoryArticleMaintenanceDue,
	}}, signal.ID)
	assertIDs(application.AccessoryArticleListQuery{Statuses: []application.AccessoryArticleStatus{
		application.AccessoryArticleDefective,
	}}, signal.ID)
	assertIDs(application.AccessoryArticleListQuery{Statuses: []application.AccessoryArticleStatus{
		application.AccessoryArticleArchived,
	}}, archived.ID)

	for sortKey, ascending := range map[string][]string{
		"article": {signal.ID, track.ID}, "type": {track.ID, signal.ID},
		"gauge": {signal.ID, track.ID}, "stock": {signal.ID, track.ID},
		"storage": {track.ID, signal.ID}, "updatedAt": {track.ID, signal.ID},
	} {
		descending := []string{ascending[1], ascending[0]}
		asc := assertIDs(application.AccessoryArticleListQuery{Sort: sortKey, Direction: "asc"}, ascending...)
		desc := assertIDs(application.AccessoryArticleListQuery{Sort: sortKey, Direction: "desc"}, descending...)
		if len(asc.Items) != len(desc.Items) {
			t.Fatalf("sort %s changed result size", sortKey)
		}
	}

	result := assertIDs(application.AccessoryArticleListQuery{}, signal.ID, track.ID)
	trackItem := result.Items[1]
	if trackItem.Owned != 14 || trackItem.Available != 7 || trackItem.Reserved != 3 ||
		trackItem.Installed != 4 || !trackItem.HasUsageHistory ||
		!reflect.DeepEqual(trackItem.LocationNames, []string{"Alpha shelf", "Beta shelf"}) ||
		len(trackItem.Attributes) != 1 {
		t.Fatalf("unexpected mixed quantity aggregation: %#v", trackItem)
	}
	if result.Metrics.ArticleCount != 2 || result.Metrics.ArticleTypeCount != 2 ||
		result.Metrics.Available != 8 || result.Metrics.Reserved != 4 || result.Metrics.Installed != 5 ||
		result.Metrics.LocationCount != 2 || result.Metrics.CareHintCount != 0 {
		t.Fatalf("unexpected global metrics: %#v", result.Metrics)
	}
	filtered := assertIDs(application.AccessoryArticleListQuery{Manufacturer: "Z-Tillig"}, track.ID)
	if filtered.Metrics != result.Metrics {
		t.Fatalf("metrics changed with table filter: %#v != %#v", filtered.Metrics, result.Metrics)
	}
	if !reflect.DeepEqual(result.FilterOptions.Manufacturers, []string{"Viessmann", "Z-Tillig"}) ||
		len(result.FilterOptions.StorageLocations) != 2 {
		t.Fatalf("unexpected filter options: %#v", result.FilterOptions)
	}
}

func TestAccessoryArticleCareHintsUseApprovedMissingFields(t *testing.T) {
	service, db := testAccessoryService(t)
	_, err := db.ExecContext(t.Context(), `INSERT INTO accessory_products(
		id, manufacturer, article_number, name, category, tracking_mode, article_type, subtype,
		gauges_json, package_quantity, stock_unit, inventory_strategy, created_at, updated_at
	) VALUES('care-hints', '', '', 'Incomplete track', 'Track', 'quantity', 'track', 'track:straight',
		'[]', 1, '', 'quantity', '2026-01-01', '2026-01-01')`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.ExecContext(t.Context(), `INSERT INTO accessory_products(
		id, manufacturer, article_number, name, category, tracking_mode, article_type, subtype,
		gauges_json, package_quantity, stock_unit, inventory_strategy, created_at, updated_at
	) VALUES('missing-type', 'Maker', '1', 'Missing type', 'Other', 'quantity', '', 'other',
		'[]', 1, 'piece', 'quantity', '2026-01-01', '2026-01-01')`)
	if err != nil {
		t.Fatal(err)
	}
	result, err := service.ListArticles(t.Context(), application.AccessoryArticleListQuery{})
	if err != nil {
		t.Fatal(err)
	}
	counts := map[string]int{}
	for _, item := range result.Items {
		counts[item.ID] = item.CareHintCount
	}
	if counts["care-hints"] != 4 || counts["missing-type"] != 1 || result.Metrics.CareHintCount != 5 {
		t.Fatalf("unexpected care hints: %#v", result)
	}
	if !reflect.DeepEqual(result.FilterOptions.Manufacturers, []string{"Maker"}) ||
		!reflect.DeepEqual(result.FilterOptions.ArticleTypes, []domain.AccessoryArticleType{domain.AccessoryArticleTrack}) {
		t.Fatalf("missing values leaked into filter options: %#v", result.FilterOptions)
	}
}

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
