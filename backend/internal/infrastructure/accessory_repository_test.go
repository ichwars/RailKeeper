package infrastructure_test

import (
	"database/sql"
	"errors"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
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
		ListPrice:   "129.90",
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
		loaded.ListPrice != input.ListPrice ||
		!loaded.Archived {
		t.Fatalf("full product did not round trip:\ninput=%#v\nloaded=%#v", input, loaded)
	}
	if loaded.Attributes[1].Unit == nil || *loaded.Attributes[1].Unit != "mm" {
		t.Fatalf("create/get lost numeric attribute unit: %#v", loaded.Attributes[1])
	}

	input.Name = "Updated track"
	input.ListPrice = "139,90"
	input.Archived = false
	input.Attributes = input.Attributes[1:]
	updatedUnit := "cm"
	input.Attributes[0].Unit = &updatedUnit
	updated, err := service.UpdateProduct(t.Context(), created.ID,
		application.UpdateAccessoryProductInput{CreateAccessoryProductInput: input}, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != input.Name || updated.ListPrice != input.ListPrice || updated.Archived ||
		!reflect.DeepEqual(updated.Attributes, input.Attributes) {
		t.Fatalf("full update did not round trip: %#v", updated)
	}
	if updated.Attributes[0].Unit == nil || *updated.Attributes[0].Unit != "cm" {
		t.Fatalf("update/get lost numeric attribute unit: %#v", updated.Attributes[0])
	}
}

func TestAccessoryArticleAssignsInventoryNumbersTransactionally(t *testing.T) {
	service, db := testAccessoryService(t)
	ctx := t.Context()
	validInput := application.CreateAccessoryProductInput{
		Manufacturer: "Tillig", ArticleNumber: "83101", Name: "Straight track",
		Category: "Track", ArticleType: domain.AccessoryArticleTrack, Subtype: "track:straight",
		PackageQuantity: 1, StockUnit: "piece", InventoryStrategy: domain.AccessoryInventoryQuantity,
	}

	first, err := service.CreateProduct(ctx, validInput, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	if first.InventoryNumber != "RK-ART-000001" {
		t.Fatalf("first inventory number: got %q", first.InventoryNumber)
	}

	invalidInput := validInput
	invalidInput.ArticleNumber = "invalid"
	invalidInput.Subtype = "track:not-active"
	if _, err := service.CreateProduct(ctx, invalidInput, "editor-1"); !errors.Is(err, application.ErrAccessoryValidation) {
		t.Fatalf("inactive subtype was not rejected: %v", err)
	}

	validInput.ArticleNumber = "83102"
	second, err := service.CreateProduct(ctx, validInput, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	if second.InventoryNumber != "RK-ART-000002" {
		t.Fatalf("rolled-back create consumed a number: got %q", second.InventoryNumber)
	}

	if _, err := db.ExecContext(ctx, `UPDATE inventory_number_schemes SET active=0 WHERE category='Artikel'`); err != nil {
		t.Fatal(err)
	}
	validInput.ArticleNumber = "83103"
	if _, err := service.CreateProduct(ctx, validInput, "editor-1"); !errors.Is(err, application.ErrInventoryNumberNotFound) {
		t.Fatalf("inactive article scheme did not block create: %v", err)
	}
}

func TestAccessoryArticleUpdatesMigrationStyleUnknownSubtypeOnlyWhenUnchanged(t *testing.T) {
	service, db := testAccessoryService(t)
	ctx := t.Context()
	if _, err := db.ExecContext(ctx, `
INSERT INTO accessory_products(
  id, inventory_number, manufacturer, article_number, name, category, tracking_mode, article_type, subtype,
  gauges_json, package_quantity, stock_unit, minimum_stock, inventory_strategy, created_at, updated_at
) VALUES(
  'legacy-migrated', 'RK-ART-LEGACY', 'Faller', '180001', 'Legacy accessory', 'legacy-category', 'quantity',
  'other', 'legacy-category', '[]', 1, 'piece', 0, 'quantity', '2026-01-01', '2026-01-01'
)`); err != nil {
		t.Fatal(err)
	}

	input := application.CreateAccessoryProductInput{
		Manufacturer: "Faller", ArticleNumber: "180001", Name: "Updated legacy accessory",
		Category: "legacy-category", TrackingMode: domain.AccessoryTrackingModeQuantity,
		ArticleType: domain.AccessoryArticleOther, Subtype: "legacy-category", PackageQuantity: 1,
		StockUnit: "piece", MinimumStock: 0, InventoryStrategy: domain.AccessoryInventoryQuantity,
	}
	updated, err := service.UpdateProduct(ctx, "legacy-migrated",
		application.UpdateAccessoryProductInput{CreateAccessoryProductInput: input}, "editor-1")
	if err != nil {
		t.Fatalf("unchanged migration-style subtype blocked unrelated update: %v", err)
	}
	if updated.Name != input.Name || updated.Subtype != "other:legacy-category" {
		t.Fatalf("legacy subtype was not canonically persisted: %#v", updated)
	}

	input.Subtype = "different-unknown"
	if _, err := service.UpdateProduct(ctx, "legacy-migrated",
		application.UpdateAccessoryProductInput{CreateAccessoryProductInput: input}, "editor-1"); !errors.Is(err, application.ErrAccessoryValidation) {
		t.Fatalf("changed unknown subtype was accepted: %v", err)
	}
	input.ArticleType = domain.AccessoryArticleTrack
	input.Subtype = "legacy-category"
	if _, err := service.UpdateProduct(ctx, "legacy-migrated",
		application.UpdateAccessoryProductInput{CreateAccessoryProductInput: input}, "editor-1"); !errors.Is(err, application.ErrAccessoryValidation) {
		t.Fatalf("unknown subtype was accepted across article types: %v", err)
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
		ListPrice: "29.90",
		Category:  "Track", ArticleType: domain.AccessoryArticleTrack, Subtype: "track:turnout", Gauges: []string{"TT"},
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
	mustExec(`INSERT INTO file_blobs(id, original_size, compressed_size, compression, sha256, data, created_at)
		VALUES('catalog-image-blob', 1, 1, 'none', 'catalog-image', X'00', '2026-01-01')`)
	mustExec(`INSERT INTO accessory_documents(
		id, product_id, file_blob_id, file_name, original_name, category, mime_type, size_bytes,
		is_primary, created_by, created_at, updated_at
	) VALUES('track-primary-image', ?, 'catalog-image-blob', 'track.png', 'track.png', 'image',
		'image/png', 1, 1, 'editor-1', '2026-01-01', '2026-01-01')`, track.ID)
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

	for _, search := range []string{"Tillig", "83125", "4012500831258", "turnout", track.InventoryNumber} {
		assertIDs(application.AccessoryArticleListQuery{Query: search}, track.ID)
	}
	assertIDs(application.AccessoryArticleListQuery{Manufacturer: "Viessmann"}, signal.ID)
	assertIDs(application.AccessoryArticleListQuery{ArticleTypes: []domain.AccessoryArticleType{domain.AccessoryArticleTrack}}, track.ID)
	assertIDs(application.AccessoryArticleListQuery{Gauges: []string{"H0"}}, signal.ID)
	assertIDs(application.AccessoryArticleListQuery{LocationID: alpha.ID}, track.ID)
	assertIDs(application.AccessoryArticleListQuery{Statuses: []application.AccessoryArticleStatus{
		application.AccessoryArticleAvailable,
	}}, track.ID, signal.ID)
	assertIDs(application.AccessoryArticleListQuery{Statuses: []application.AccessoryArticleStatus{
		application.AccessoryArticleReserved,
	}}, track.ID, signal.ID)
	assertIDs(application.AccessoryArticleListQuery{Statuses: []application.AccessoryArticleStatus{
		application.AccessoryArticleInstalled,
	}}, track.ID, signal.ID)
	assertIDs(application.AccessoryArticleListQuery{Statuses: []application.AccessoryArticleStatus{
		application.AccessoryArticleReserved, application.AccessoryArticleInstalled,
	}}, track.ID, signal.ID)
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
		"article": {signal.ID, track.ID}, "inventoryNumber": {track.ID, signal.ID},
		"image":        {signal.ID, track.ID},
		"manufacturer": {signal.ID, track.ID}, "articleNumber": {signal.ID, track.ID},
		"name": {signal.ID, track.ID}, "type": {track.ID, signal.ID},
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

	result := assertIDs(application.AccessoryArticleListQuery{}, track.ID, signal.ID)
	trackItem := result.Items[0]
	if trackItem.Owned != 14 || trackItem.Available != 7 || trackItem.Reserved != 3 ||
		trackItem.Installed != 4 || !trackItem.HasUsageHistory ||
		trackItem.ListPrice != "29.90" ||
		trackItem.InventoryNumber != track.InventoryNumber ||
		trackItem.PrimaryImageURL != "/api/v1/accessory-products/"+track.ID+
			"/documents/track-primary-image/download" ||
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

func TestAccessoryArticleHybridAggregationMatchesAllocationSummary(t *testing.T) {
	fixture := newAllocationFixture(t)
	ctx := t.Context()
	baseline, err := fixture.accessories.ListArticles(ctx, application.AccessoryArticleListQuery{})
	if err != nil {
		t.Fatal(err)
	}
	hybrid := createAccessoryTestProduct(t, fixture.accessories, "Hybrid overview",
		domain.AccessoryInventoryQuantityLaterIndividual)
	if _, err := fixture.accessories.AdjustStock(ctx, hybrid.ID, application.StockAdjustmentInput{
		LocationID: fixture.location.ID, Delta: 4,
	}, "editor-1"); err != nil {
		t.Fatal(err)
	}
	for _, inventoryNumber := range []string{"HYB-OV-1", "HYB-OV-2"} {
		if _, err := fixture.accessories.Individualize(ctx, hybrid.ID, application.IndividualizeAccessoryInput{
			LocationID: fixture.location.ID,
			Asset:      application.CreateAccessoryAssetInput{InventoryNumber: inventoryNumber},
		}, "editor-1"); err != nil {
			t.Fatal(err)
		}
	}
	assets, err := fixture.accessories.ListAssets(ctx, hybrid.ID)
	if err != nil {
		t.Fatal(err)
	}
	assetReservation, err := fixture.allocations.CreateReservation(ctx,
		application.CreateAccessoryReservationInput{
			ProductID: hybrid.ID, AssetID: assets[0].ID, LocationID: fixture.location.ID, Quantity: 1,
			AllocationTargetInput: application.AllocationTargetInput{LayoutID: fixture.layout.ID},
		}, "planner-1")
	if err != nil {
		t.Fatal(err)
	}
	assertHybridArticleTotals(t, fixture.accessories, fixture.allocations, hybrid.ID, baseline.Metrics,
		application.AccessoryAllocationSummary{Owned: 4, Stored: 4, Reserved: 1, Available: 3})
	quantityReservation, err := fixture.allocations.CreateReservation(ctx,
		application.CreateAccessoryReservationInput{
			ProductID: hybrid.ID, LocationID: fixture.location.ID, Quantity: 2,
			AllocationTargetInput: application.AllocationTargetInput{LayoutUnitID: fixture.unit.ID},
		}, "planner-1")
	if err != nil {
		t.Fatalf("asset reservation reduced remaining hybrid quantity: %v", err)
	}
	assertHybridArticleTotals(t, fixture.accessories, fixture.allocations, hybrid.ID, baseline.Metrics,
		application.AccessoryAllocationSummary{Owned: 4, Stored: 4, Reserved: 3, Available: 1})
	if _, err := fixture.allocations.CancelReservation(ctx, quantityReservation.ID, "planner-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.allocations.Install(ctx, application.CreateAccessoryInstallationInput{
		ReservationID: assetReservation.ID, ProductID: hybrid.ID, AssetID: assets[0].ID,
		SourceLocationID: fixture.location.ID, Quantity: 1,
		AllocationTargetInput: application.AllocationTargetInput{LayoutID: fixture.layout.ID},
		Condition:             domain.AccessoryConditionReady,
	}, "editor-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.allocations.CreateReservation(ctx, application.CreateAccessoryReservationInput{
		ProductID: hybrid.ID, LocationID: fixture.location.ID, Quantity: 1,
		AllocationTargetInput: application.AllocationTargetInput{LayoutUnitID: fixture.unit.ID},
	}, "planner-1"); err != nil {
		t.Fatal(err)
	}
	assertHybridArticleTotals(t, fixture.accessories, fixture.allocations, hybrid.ID, baseline.Metrics,
		application.AccessoryAllocationSummary{Owned: 4, Stored: 3, Reserved: 1, Installed: 1, Available: 2})

	for _, status := range []application.AccessoryArticleStatus{
		application.AccessoryArticleAvailable,
		application.AccessoryArticleReserved,
		application.AccessoryArticleInstalled,
	} {
		result, err := fixture.accessories.ListArticles(ctx, application.AccessoryArticleListQuery{
			Statuses: []application.AccessoryArticleStatus{status},
		})
		if err != nil {
			t.Fatal(err)
		}
		if !accessoryArticleResultContains(result, hybrid.ID) {
			t.Fatalf("hybrid article missing from %q filter: %#v", status, result.Items)
		}
	}
}

func assertHybridArticleTotals(
	t *testing.T,
	service *application.AccessoryService,
	allocations *application.AccessoryAllocationService,
	productID string,
	baseline application.AccessoryOverviewMetrics,
	want application.AccessoryAllocationSummary,
) {
	t.Helper()
	allocationSummary, err := allocations.GetAllocationSummary(t.Context(), productID)
	if err != nil {
		t.Fatal(err)
	}
	want.ProductID = productID
	if *allocationSummary != want {
		t.Fatalf("unexpected hybrid allocation summary: got %#v, want %#v", *allocationSummary, want)
	}
	result, err := service.ListArticles(t.Context(), application.AccessoryArticleListQuery{})
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range result.Items {
		if item.ID != productID {
			continue
		}
		if item.Owned != allocationSummary.Owned || item.Available != allocationSummary.Available ||
			item.Reserved != allocationSummary.Reserved || item.Installed != allocationSummary.Installed {
			t.Fatalf("hybrid article totals %#v do not match allocation totals %#v", item, allocationSummary)
		}
		if result.Metrics.Available != baseline.Available+want.Available ||
			result.Metrics.Reserved != baseline.Reserved+want.Reserved ||
			result.Metrics.Installed != baseline.Installed+want.Installed {
			t.Fatalf("hybrid totals missing from overview metrics: baseline=%#v got=%#v want=%#v",
				baseline, result.Metrics, want)
		}
		return
	}
	t.Fatalf("hybrid article %q not found in %#v", productID, result.Items)
}

func accessoryArticleResultContains(result *application.AccessoryArticleListResult, productID string) bool {
	for _, item := range result.Items {
		if item.ID == productID {
			return true
		}
	}
	return false
}

func TestAccessoryArticleCareHintsUseApprovedMissingFields(t *testing.T) {
	service, db := testAccessoryService(t)
	_, err := db.ExecContext(t.Context(), `INSERT INTO accessory_products(
		id, inventory_number, manufacturer, article_number, name, category, tracking_mode, article_type, subtype,
		gauges_json, package_quantity, stock_unit, inventory_strategy, created_at, updated_at
	) VALUES('care-hints', 'RK-ART-CARE', '', '', 'Incomplete track', 'Track', 'quantity', 'track', 'track:straight',
		'[]', 1, '', 'quantity', '2026-01-01', '2026-01-01')`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.ExecContext(t.Context(), `INSERT INTO accessory_products(
		id, inventory_number, manufacturer, article_number, name, category, tracking_mode, article_type, subtype,
		gauges_json, package_quantity, stock_unit, inventory_strategy, created_at, updated_at
	) VALUES('missing-type', 'RK-ART-TYPE', 'Maker', '1', 'Missing type', 'Other', 'quantity', '', 'other',
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

func TestAccessoryInventoryStrategyChangesPreserveExistingInventoryKinds(t *testing.T) {
	fixture := newAllocationFixture(t)
	ctx := t.Context()
	hybrid := createAccessoryTestProduct(t, fixture.accessories, "Hybrid transition",
		domain.AccessoryInventoryQuantityLaterIndividual)
	if _, err := fixture.accessories.AdjustStock(ctx, hybrid.ID, application.StockAdjustmentInput{
		LocationID: fixture.location.ID, Delta: 2,
	}, "editor-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.accessories.Individualize(ctx, hybrid.ID, application.IndividualizeAccessoryInput{
		LocationID: fixture.location.ID,
		Asset:      application.CreateAccessoryAssetInput{InventoryNumber: "HYB-TRANSITION-1"},
	}, "editor-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.accessories.UpdateProduct(ctx, hybrid.ID,
		accessoryProductUpdateWithStrategy(hybrid, domain.AccessoryInventoryQuantity),
		"editor-1"); !errors.Is(err, application.ErrAccessoryConflict) {
		t.Fatalf("expected hybrid-to-quantity conflict with individualized asset, got %v", err)
	}
	unchanged, err := fixture.accessories.GetProduct(ctx, hybrid.ID)
	if err != nil {
		t.Fatal(err)
	}
	if unchanged.InventoryStrategy != domain.AccessoryInventoryQuantityLaterIndividual {
		t.Fatalf("failed strategy update changed product: %#v", unchanged)
	}

	if _, err := fixture.accessories.UpdateProduct(ctx, fixture.individualProduct.ID,
		accessoryProductUpdateWithStrategy(fixture.individualProduct,
			domain.AccessoryInventoryQuantityLaterIndividual),
		"editor-1"); err != nil {
		t.Fatalf("individual-to-hybrid transition should retain asset support: %v", err)
	}
	emptyHybrid := createAccessoryTestProduct(t, fixture.accessories, "Empty hybrid transition",
		domain.AccessoryInventoryQuantityLaterIndividual)
	updated, err := fixture.accessories.UpdateProduct(ctx, emptyHybrid.ID,
		accessoryProductUpdateWithStrategy(emptyHybrid, domain.AccessoryInventoryQuantity),
		"editor-1")
	if err != nil {
		t.Fatalf("empty hybrid-to-quantity transition should remain safe: %v", err)
	}
	if updated.InventoryStrategy != domain.AccessoryInventoryQuantity {
		t.Fatalf("unexpected safe strategy transition result: %#v", updated)
	}
}

func accessoryProductUpdateWithStrategy(
	product *application.AccessoryProduct,
	strategy domain.AccessoryInventoryStrategy,
) application.UpdateAccessoryProductInput {
	return application.UpdateAccessoryProductInput{CreateAccessoryProductInput: application.CreateAccessoryProductInput{
		Manufacturer: product.Manufacturer, ArticleNumber: product.ArticleNumber, Name: product.Name,
		Category: product.Category, Description: product.Description, EAN: product.EAN,
		ManufacturerStatus: product.ManufacturerStatus, ArticleType: product.ArticleType, Subtype: product.Subtype,
		Gauges: product.Gauges, Scale: product.Scale, PackageQuantity: product.PackageQuantity,
		StockUnit: product.StockUnit, MinimumStock: product.MinimumStock, InventoryStrategy: strategy,
		ManufacturerURL: product.ManufacturerURL, ProductURL: product.ProductURL,
		AlternativeNumbers: product.AlternativeNumbers, Keywords: product.Keywords,
		CompatibilityNotes: product.CompatibilityNotes, InternalNotes: product.InternalNotes,
		Archived: product.Archived, Attributes: product.Attributes,
	}}
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

func TestAccessoryStockMovementJournalsAdjustmentsAndRollsBackInsufficientStock(t *testing.T) {
	service, db := testAccessoryService(t)
	ctx := t.Context()
	location := createAccessoryTestLocation(t, service, "Stock shelf")
	product := createAccessoryTestProduct(t, service, "Stock movement", domain.AccessoryInventoryQuantity)

	if _, err := service.AdjustStock(ctx, product.ID, application.StockAdjustmentInput{
		LocationID: location.ID, Delta: 4,
	}, "editor-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.AdjustStock(ctx, product.ID, application.StockAdjustmentInput{
		LocationID: location.ID, Delta: -5,
	}, "editor-1"); !errors.Is(err, application.ErrAccessoryInsufficientStock) {
		t.Fatalf("expected insufficient stock, got %v", err)
	}

	movements, err := service.ListStockMovements(ctx, product.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(movements) != 1 || movements[0].MovementType != "adjustment" ||
		movements[0].Quantity != 4 || movements[0].LocationID != location.ID ||
		movements[0].Actor != "editor-1" {
		t.Fatalf("unexpected adjustment journal: %#v", movements)
	}
	assertAccessoryTestStock(t, service, product.ID, map[string]int{location.ID: 4})
	assertAccessoryAuditCount(t, db, "AccessoryStockAdjusted", 1)
}

func TestAccessoryTransferStockWritesPairedMovementsAndRollsBack(t *testing.T) {
	service, db := testAccessoryService(t)
	ctx := t.Context()
	source := createAccessoryTestLocation(t, service, "Source shelf")
	destination := createAccessoryTestLocation(t, service, "Destination shelf")
	product := createAccessoryTestProduct(t, service, "Transferred stock", domain.AccessoryInventoryQuantity)
	if _, err := service.AdjustStock(ctx, product.ID, application.StockAdjustmentInput{
		LocationID: source.ID, Delta: 5,
	}, "editor-1"); err != nil {
		t.Fatal(err)
	}

	if _, err := service.TransferStock(ctx, product.ID, application.TransferAccessoryStockInput{
		FromLocationID: source.ID, ToLocationID: destination.ID, Quantity: 3, Note: "layout preparation",
	}, "editor-2"); err != nil {
		t.Fatal(err)
	}
	assertAccessoryTestStock(t, service, product.ID, map[string]int{source.ID: 2, destination.ID: 3})
	movements, err := service.ListStockMovements(ctx, product.ID)
	if err != nil {
		t.Fatal(err)
	}
	out := findAccessoryMovement(t, movements, "transfer_out")
	in := findAccessoryMovement(t, movements, "transfer_in")
	if out.Quantity != -3 || out.LocationID != source.ID || in.Quantity != 3 ||
		in.LocationID != destination.ID || out.SourceType != "transfer" ||
		out.SourceID == "" || out.SourceID != in.SourceID || out.Note != "layout preparation" ||
		in.Note != "layout preparation" || out.Actor != "editor-2" || in.Actor != "editor-2" {
		t.Fatalf("unexpected transfer journal: out=%#v in=%#v", out, in)
	}
	assertAccessoryAuditCount(t, db, "AccessoryStockTransferred", 1)

	before := len(movements)
	if _, err := service.TransferStock(ctx, product.ID, application.TransferAccessoryStockInput{
		FromLocationID: source.ID, ToLocationID: destination.ID, Quantity: 3,
	}, "editor-2"); !errors.Is(err, application.ErrAccessoryInsufficientStock) {
		t.Fatalf("expected failed over-transfer, got %v", err)
	}
	after, err := service.ListStockMovements(ctx, product.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(after) != before {
		t.Fatalf("failed transfer wrote movements: before=%d after=%d", before, len(after))
	}
	assertAccessoryTestStock(t, service, product.ID, map[string]int{source.ID: 2, destination.ID: 3})
	assertAccessoryAuditCount(t, db, "AccessoryStockTransferred", 1)

	if _, err := service.UpdateLocation(ctx, destination.ID, application.UpdateStorageLocationInput{
		CreateStorageLocationInput: application.CreateStorageLocationInput{Name: destination.Name, Archived: true},
	}, "editor-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.TransferStock(ctx, product.ID, application.TransferAccessoryStockInput{
		FromLocationID: source.ID, ToLocationID: destination.ID, Quantity: 1,
	}, "editor-2"); !errors.Is(err, application.ErrAccessoryConflict) {
		t.Fatalf("expected archived destination rejection, got %v", err)
	}
}

func TestAccessoryPurchaseRecordsUnbookedAndBooksQuantityStock(t *testing.T) {
	service, db := testAccessoryService(t)
	ctx := t.Context()
	location := createAccessoryTestLocation(t, service, "Purchase shelf")
	product := createAccessoryTestProduct(t, service, "Purchased stock", domain.AccessoryInventoryQuantity)

	unbooked, err := service.CreatePurchase(ctx, product.ID, application.CreateAccessoryPurchaseInput{
		PurchasedAt: "2026-08-01", Supplier: "Dealer", Quantity: 2, UnitPrice: "3.50",
		Currency: "EUR", InvoiceNumber: "INV-1", WarrantyUntil: "2028-08-01", Notes: "ordered",
	}, "buyer-1")
	if err != nil {
		t.Fatal(err)
	}
	if unbooked.BookToStock || unbooked.StorageLocationID != "" || unbooked.InvoiceNumber != "INV-1" {
		t.Fatalf("unexpected unbooked purchase: %#v", unbooked)
	}
	assertAccessoryTestStock(t, service, product.ID, map[string]int{})
	movements, err := service.ListStockMovements(ctx, product.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(movements) != 0 {
		t.Fatalf("unbooked purchase changed physical stock: %#v", movements)
	}

	booked, err := service.CreatePurchase(ctx, product.ID, application.CreateAccessoryPurchaseInput{
		PurchasedAt: "2026-08-02", Supplier: "Dealer", Quantity: 3, UnitPrice: "4.25",
		Currency: "EUR", InvoiceNumber: "INV-2", StorageLocationID: location.ID,
		BookToStock: true, Notes: "received",
	}, "buyer-1")
	if err != nil {
		t.Fatal(err)
	}
	assertAccessoryTestStock(t, service, product.ID, map[string]int{location.ID: 3})
	movements, err = service.ListStockMovements(ctx, product.ID)
	if err != nil {
		t.Fatal(err)
	}
	purchaseMovement := findAccessoryMovement(t, movements, "purchase")
	if purchaseMovement.Quantity != 3 || purchaseMovement.LocationID != location.ID ||
		purchaseMovement.SourceType != "purchase" || purchaseMovement.SourceID != booked.ID ||
		purchaseMovement.Actor != "buyer-1" || purchaseMovement.Note != "received" {
		t.Fatalf("unexpected purchase movement: %#v", purchaseMovement)
	}
	purchases, err := service.ListPurchases(ctx, product.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(purchases) != 2 || purchases[0].ID != booked.ID || purchases[1].ID != unbooked.ID {
		t.Fatalf("unexpected purchase history: %#v", purchases)
	}
	assertAccessoryAuditCount(t, db, "AccessoryPurchaseCreated", 2)
}

func TestAccessoryPurchaseBooksIndividualAssetsWithoutQuantityStock(t *testing.T) {
	service, db := testAccessoryService(t)
	ctx := t.Context()
	location := createAccessoryTestLocation(t, service, "Decoder shelf")
	product := createAccessoryTestProduct(t, service, "Purchased decoder", domain.AccessoryInventoryIndividual)

	purchase, err := service.CreatePurchase(ctx, product.ID, application.CreateAccessoryPurchaseInput{
		PurchasedAt: "2026-08-03", Supplier: "Decoder dealer", Quantity: 2, UnitPrice: "49.95",
		Currency: "EUR", InvoiceNumber: "D-42", WarrantyUntil: "2028-08-03",
		StorageLocationID: location.ID, BookToStock: true,
	}, "buyer-1")
	if err != nil {
		t.Fatal(err)
	}
	assets, err := service.ListAssets(ctx, product.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(assets) != 2 {
		t.Fatalf("expected two purchased assets, got %#v", assets)
	}
	for _, asset := range assets {
		if asset.PurchaseID != purchase.ID || asset.PurchaseDate != "2026-08-03" ||
			asset.PurchasePrice != "49.95" || asset.WarrantyUntil != "2028-08-03" ||
			asset.StorageLocationID != location.ID || asset.Lifecycle != domain.AccessoryLifecycleStored {
			t.Fatalf("unexpected purchased asset: %#v", asset)
		}
	}
	stock, err := service.GetStock(ctx, product.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stock.TotalQuantity != 0 || len(stock.Locations) != 0 {
		t.Fatalf("individual purchase also created quantity stock: %#v", stock)
	}
	movements, err := service.ListStockMovements(ctx, product.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(movements) != 0 {
		t.Fatalf("individual purchase wrote quantity movements: %#v", movements)
	}
	assertAccessoryAuditCount(t, db, "AccessoryPurchaseCreated", 1)
}

func TestAccessoryPurchaseBooksHybridQuantityStock(t *testing.T) {
	service, _ := testAccessoryService(t)
	ctx := t.Context()
	location := createAccessoryTestLocation(t, service, "Hybrid purchase shelf")
	product := createAccessoryTestProduct(t, service, "Purchased hybrid",
		domain.AccessoryInventoryQuantityLaterIndividual)

	purchase, err := service.CreatePurchase(ctx, product.ID, application.CreateAccessoryPurchaseInput{
		PurchasedAt: "2026-08-04", Quantity: 2, StorageLocationID: location.ID, BookToStock: true,
	}, "buyer-1")
	if err != nil {
		t.Fatal(err)
	}
	assertAccessoryTestStock(t, service, product.ID, map[string]int{location.ID: 2})
	movements, err := service.ListStockMovements(ctx, product.ID)
	if err != nil {
		t.Fatal(err)
	}
	movement := findAccessoryMovement(t, movements, "purchase")
	if movement.Quantity != 2 || movement.SourceID != purchase.ID {
		t.Fatalf("unexpected hybrid purchase movement: %#v", movement)
	}
	assets, err := service.ListAssets(ctx, product.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(assets) != 0 {
		t.Fatalf("hybrid purchase individualized assets eagerly: %#v", assets)
	}
}

func TestAccessoryIndividualizationConsumesHybridStockAndRollsBackConflicts(t *testing.T) {
	service, db := testAccessoryService(t)
	ctx := t.Context()
	location := createAccessoryTestLocation(t, service, "Hybrid shelf")
	product := createAccessoryTestProduct(t, service, "Hybrid decoder",
		domain.AccessoryInventoryQuantityLaterIndividual)
	if _, err := service.AdjustStock(ctx, product.ID, application.StockAdjustmentInput{
		LocationID: location.ID, Delta: 2,
	}, "editor-1"); err != nil {
		t.Fatal(err)
	}

	asset, err := service.Individualize(ctx, product.ID, application.IndividualizeAccessoryInput{
		LocationID: location.ID, Asset: application.CreateAccessoryAssetInput{
			InventoryNumber: "HYB-1", SerialNumber: "SER-1", PurchaseDate: "2026-08-01",
			PurchasePrice: "25.00", WarrantyUntil: "2028-08-01",
		},
	}, "editor-2")
	if err != nil {
		t.Fatal(err)
	}
	if asset.StorageLocationID != location.ID || asset.Lifecycle != domain.AccessoryLifecycleStored {
		t.Fatalf("unexpected individualized asset: %#v", asset)
	}
	assertAccessoryTestStock(t, service, product.ID, map[string]int{location.ID: 1})
	assets, err := service.ListAssets(ctx, product.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(assets) != 1 || assets[0].ID != asset.ID {
		t.Fatalf("hybrid asset listing failed: %#v", assets)
	}
	if _, err := service.CreateAsset(ctx, product.ID, application.CreateAccessoryAssetInput{
		InventoryNumber: "HYB-DIRECT", StorageLocationID: location.ID,
	}, "editor-2"); !errors.Is(err, application.ErrAccessoryTrackingMode) {
		t.Fatalf("direct hybrid asset creation must stay disabled, got %v", err)
	}
	movements, err := service.ListStockMovements(ctx, product.ID)
	if err != nil {
		t.Fatal(err)
	}
	individualization := findAccessoryMovement(t, movements, "individualization")
	if individualization.Quantity != -1 || individualization.LocationID != location.ID ||
		individualization.SourceType != "asset" || individualization.SourceID != asset.ID ||
		individualization.Actor != "editor-2" {
		t.Fatalf("unexpected individualization movement: %#v", individualization)
	}

	beforeMovements := len(movements)
	if _, err := service.Individualize(ctx, product.ID, application.IndividualizeAccessoryInput{
		LocationID: location.ID, Asset: application.CreateAccessoryAssetInput{InventoryNumber: "HYB-1"},
	}, "editor-2"); !errors.Is(err, application.ErrAccessoryConflict) {
		t.Fatalf("expected duplicate inventory conflict, got %v", err)
	}
	assertAccessoryTestStock(t, service, product.ID, map[string]int{location.ID: 1})
	movements, err = service.ListStockMovements(ctx, product.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(movements) != beforeMovements {
		t.Fatalf("failed individualization left a movement: %#v", movements)
	}
	assertAccessoryAuditCount(t, db, "AccessoryAssetIndividualized", 1)
}

func TestAccessoryConcurrentIndividualizationNeverProducesNegativeStock(t *testing.T) {
	service, _ := testAccessoryService(t)
	ctx := t.Context()
	location := createAccessoryTestLocation(t, service, "Concurrent shelf")
	product := createAccessoryTestProduct(t, service, "Concurrent hybrid",
		domain.AccessoryInventoryQuantityLaterIndividual)
	if _, err := service.AdjustStock(ctx, product.ID, application.StockAdjustmentInput{
		LocationID: location.ID, Delta: 1,
	}, "editor-1"); err != nil {
		t.Fatal(err)
	}

	start := make(chan struct{})
	errorsByCall := make(chan error, 2)
	var wait sync.WaitGroup
	for _, inventoryNumber := range []string{"CON-1", "CON-2"} {
		wait.Add(1)
		go func(number string) {
			defer wait.Done()
			<-start
			_, err := service.Individualize(ctx, product.ID, application.IndividualizeAccessoryInput{
				LocationID: location.ID,
				Asset:      application.CreateAccessoryAssetInput{InventoryNumber: number},
			}, "editor-1")
			errorsByCall <- err
		}(inventoryNumber)
	}
	close(start)
	wait.Wait()
	close(errorsByCall)
	successes, insufficient := 0, 0
	for err := range errorsByCall {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, application.ErrAccessoryInsufficientStock):
			insufficient++
		default:
			t.Fatalf("unexpected concurrent decrement error: %v", err)
		}
	}
	if successes != 1 || insufficient != 1 {
		t.Fatalf("unexpected concurrent results: successes=%d insufficient=%d", successes, insufficient)
	}
	assertAccessoryTestStock(t, service, product.ID, map[string]int{location.ID: 0})
	assets, err := service.ListAssets(ctx, product.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(assets) != 1 {
		t.Fatalf("concurrent decrement created %d assets", len(assets))
	}
}

func createAccessoryTestLocation(
	t *testing.T,
	service *application.AccessoryService,
	name string,
) *application.StorageLocation {
	t.Helper()
	location, err := service.CreateLocation(t.Context(), application.CreateStorageLocationInput{Name: name}, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	return location
}

func createAccessoryTestProduct(
	t *testing.T,
	service *application.AccessoryService,
	name string,
	strategy domain.AccessoryInventoryStrategy,
) *application.AccessoryProduct {
	t.Helper()
	product, err := service.CreateProduct(t.Context(), application.CreateAccessoryProductInput{
		Manufacturer: "Test", Name: name, Category: "Other", ArticleType: domain.AccessoryArticleOther,
		Subtype: "other", PackageQuantity: 1, StockUnit: "piece", InventoryStrategy: strategy,
	}, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	return product
}

func assertAccessoryTestStock(
	t *testing.T,
	service *application.AccessoryService,
	productID string,
	want map[string]int,
) {
	t.Helper()
	stock, err := service.GetStock(t.Context(), productID)
	if err != nil {
		t.Fatal(err)
	}
	got := make(map[string]int, len(stock.Locations))
	for _, level := range stock.Locations {
		got[level.LocationID] = level.Quantity
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected stock: got %#v, want %#v", got, want)
	}
}

func findAccessoryMovement(
	t *testing.T,
	movements []application.AccessoryStockMovement,
	movementType string,
) application.AccessoryStockMovement {
	t.Helper()
	for _, movement := range movements {
		if movement.MovementType == movementType {
			return movement
		}
	}
	t.Fatalf("movement %q not found in %#v", movementType, movements)
	return application.AccessoryStockMovement{}
}

func assertAccessoryAuditCount(t *testing.T, db *sql.DB, action string, want int) {
	t.Helper()
	var got int
	if err := db.QueryRowContext(t.Context(), `SELECT COUNT(*) FROM audit_logs WHERE action=?`, action).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("audit count for %s: got %d, want %d", action, got, want)
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
