package application

import (
	"context"
	"errors"
	"testing"

	"railkeeper/backend/internal/domain"
)

type accessoryRepositorySpy struct {
	AccessoryRepository
	createdProduct  CreateAccessoryProductInput
	updatedProduct  UpdateAccessoryProductInput
	createdLocation CreateStorageLocationInput
	updatedLocation UpdateStorageLocationInput
	stockAdjustment StockAdjustmentInput
	createdAsset    CreateAccessoryAssetInput
	updatedAsset    UpdateAccessoryAssetInput
}

func (spy *accessoryRepositorySpy) CreateProduct(
	_ context.Context,
	input CreateAccessoryProductInput,
	_ string,
) (*AccessoryProduct, error) {
	spy.createdProduct = input
	return &AccessoryProduct{Name: input.Name}, nil
}

func (spy *accessoryRepositorySpy) UpdateProduct(
	_ context.Context,
	_ string,
	input UpdateAccessoryProductInput,
	_ string,
) (*AccessoryProduct, error) {
	spy.updatedProduct = input
	return &AccessoryProduct{Name: input.Name}, nil
}

func (spy *accessoryRepositorySpy) CreateLocation(
	_ context.Context,
	input CreateStorageLocationInput,
	_ string,
) (*StorageLocation, error) {
	spy.createdLocation = input
	return &StorageLocation{Name: input.Name}, nil
}

func (spy *accessoryRepositorySpy) UpdateLocation(
	_ context.Context,
	_ string,
	input UpdateStorageLocationInput,
	_ string,
) (*StorageLocation, error) {
	spy.updatedLocation = input
	return &StorageLocation{Name: input.Name}, nil
}

func (spy *accessoryRepositorySpy) AdjustStock(
	_ context.Context,
	_ string,
	input StockAdjustmentInput,
	_ string,
) (*AccessoryStockSummary, error) {
	spy.stockAdjustment = input
	return &AccessoryStockSummary{}, nil
}

func (spy *accessoryRepositorySpy) CreateAsset(
	_ context.Context,
	_ string,
	input CreateAccessoryAssetInput,
	_ string,
) (*AccessoryAsset, error) {
	spy.createdAsset = input
	return &AccessoryAsset{Condition: input.Condition, Lifecycle: input.Lifecycle}, nil
}

func (spy *accessoryRepositorySpy) UpdateAsset(
	_ context.Context,
	_ string,
	input UpdateAccessoryAssetInput,
	_ string,
) (*AccessoryAsset, error) {
	spy.updatedAsset = input
	return &AccessoryAsset{Condition: input.Condition, Lifecycle: input.Lifecycle}, nil
}

func TestAccessoryServiceNormalizesProductAndLocationInputs(t *testing.T) {
	repository := &accessoryRepositorySpy{}
	service := NewAccessoryService(repository)

	if _, err := service.CreateProduct(t.Context(), CreateAccessoryProductInput{
		Manufacturer: " Tillig ", ArticleNumber: " 83125 ", Name: " Weiche ", Category: " Gleismaterial ",
		TrackingMode: domain.AccessoryTrackingModeQuantity, Description: " Rechts ",
	}, "editor-1"); err != nil {
		t.Fatal(err)
	}
	product := repository.createdProduct
	if product.Manufacturer != "Tillig" || product.ArticleNumber != "83125" || product.Name != "Weiche" ||
		product.Category != "Gleismaterial" || product.Description != "Rechts" {
		t.Fatalf("unexpected normalized product: %#v", product)
	}
	if _, err := service.UpdateProduct(t.Context(), " product-1 ", UpdateAccessoryProductInput{
		CreateAccessoryProductInput: CreateAccessoryProductInput{
			Manufacturer: " Tillig ", Name: " Decoder ", Category: " Elektronik ",
			TrackingMode: domain.AccessoryTrackingModeIndividual,
		},
	}, "editor-1"); err != nil {
		t.Fatal(err)
	}
	if repository.updatedProduct.Name != "Decoder" || repository.updatedProduct.Category != "Elektronik" {
		t.Fatalf("unexpected normalized product update: %#v", repository.updatedProduct)
	}

	if _, err := service.CreateLocation(t.Context(), CreateStorageLocationInput{
		ParentID: " parent-1 ", Name: " Schublade 1 ", Description: " Gleismaterial ",
	}, "editor-1"); err != nil {
		t.Fatal(err)
	}
	location := repository.createdLocation
	if location.ParentID != "parent-1" || location.Name != "Schublade 1" || location.Description != "Gleismaterial" {
		t.Fatalf("unexpected normalized location: %#v", location)
	}
	if _, err := service.UpdateLocation(t.Context(), " location-1 ", UpdateStorageLocationInput{
		CreateStorageLocationInput: CreateStorageLocationInput{ParentID: " parent-2 ", Name: " Fach 2 "},
	}, "editor-1"); err != nil {
		t.Fatal(err)
	}
	if repository.updatedLocation.ParentID != "parent-2" || repository.updatedLocation.Name != "Fach 2" {
		t.Fatalf("unexpected normalized location update: %#v", repository.updatedLocation)
	}
	if _, err := service.AdjustStock(t.Context(), " product-1 ", StockAdjustmentInput{
		LocationID: " location-1 ", Delta: 3,
	}, "editor-1"); err != nil {
		t.Fatal(err)
	}
	if repository.stockAdjustment.LocationID != "location-1" || repository.stockAdjustment.Delta != 3 {
		t.Fatalf("unexpected normalized stock adjustment: %#v", repository.stockAdjustment)
	}
}

func TestAccessoryServiceRejectsInvalidInputs(t *testing.T) {
	service := NewAccessoryService(&accessoryRepositorySpy{})

	products := []CreateAccessoryProductInput{
		{Name: "Missing manufacturer", Category: "Track", TrackingMode: domain.AccessoryTrackingModeQuantity},
		{Manufacturer: "Tillig", Name: "Missing category", TrackingMode: domain.AccessoryTrackingModeQuantity},
		{Manufacturer: "Tillig", Name: "Invalid mode", Category: "Track", TrackingMode: "shared"},
	}
	for _, input := range products {
		if _, err := service.CreateProduct(t.Context(), input, "editor-1"); !errors.Is(err, ErrAccessoryValidation) {
			t.Fatalf("expected product validation error for %#v, got %v", input, err)
		}
	}
	if _, err := service.UpdateLocation(t.Context(), "location-1", UpdateStorageLocationInput{
		CreateStorageLocationInput: CreateStorageLocationInput{Name: "Self", ParentID: "location-1"},
	}, "editor-1"); !errors.Is(err, ErrAccessoryValidation) {
		t.Fatalf("expected self-parent validation error, got %v", err)
	}
	if _, err := service.CreateLocation(t.Context(), CreateStorageLocationInput{}, "editor-1"); !errors.Is(err, ErrAccessoryValidation) {
		t.Fatalf("expected location validation error, got %v", err)
	}
	if _, err := service.AdjustStock(t.Context(), "product-1", StockAdjustmentInput{
		LocationID: "location-1", Delta: 0,
	}, "editor-1"); !errors.Is(err, ErrAccessoryValidation) {
		t.Fatalf("expected zero adjustment validation error, got %v", err)
	}
}

func TestAccessoryServiceDefaultsAndValidatesAssetState(t *testing.T) {
	repository := &accessoryRepositorySpy{}
	service := NewAccessoryService(repository)

	asset, err := service.CreateAsset(t.Context(), "product-1", CreateAccessoryAssetInput{
		InventoryNumber: " Z-0001 ", SerialNumber: " ABC ", StorageLocationID: " location-1 ",
		PurchasePrice: " 19.99 ", Notes: " Test ",
	}, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	if asset.Condition != domain.AccessoryConditionUnknown || asset.Lifecycle != domain.AccessoryLifecycleStored {
		t.Fatalf("unexpected default asset state: %#v", asset)
	}
	input := repository.createdAsset
	if input.InventoryNumber != "Z-0001" || input.SerialNumber != "ABC" || input.StorageLocationID != "location-1" ||
		input.PurchasePrice != "19.99" || input.Notes != "Test" {
		t.Fatalf("unexpected normalized asset: %#v", input)
	}
	if _, err := service.UpdateAsset(t.Context(), "asset-1", UpdateAccessoryAssetInput{
		CreateAccessoryAssetInput: CreateAccessoryAssetInput{Condition: "broken", Lifecycle: domain.AccessoryLifecycleStored},
	}, "editor-1"); !errors.Is(err, ErrAccessoryValidation) {
		t.Fatalf("expected asset state validation error, got %v", err)
	}
	for _, lifecycle := range []domain.AccessoryLifecycle{
		domain.AccessoryLifecycleReserved,
		domain.AccessoryLifecycleInstalled,
	} {
		if _, err := service.CreateAsset(t.Context(), "product-1", CreateAccessoryAssetInput{
			Condition: domain.AccessoryConditionReady, Lifecycle: lifecycle,
		}, "editor-1"); !errors.Is(err, ErrAccessoryValidation) {
			t.Fatalf("expected allocation-owned lifecycle %q to be rejected, got %v", lifecycle, err)
		}
	}
	if _, err := service.UpdateAsset(t.Context(), " asset-1 ", UpdateAccessoryAssetInput{
		CreateAccessoryAssetInput: CreateAccessoryAssetInput{
			InventoryNumber: " Z-0002 ", Condition: domain.AccessoryConditionReady,
			Lifecycle: domain.AccessoryLifecycleMaintenance,
		},
	}, "editor-1"); err != nil {
		t.Fatal(err)
	}
	if repository.updatedAsset.InventoryNumber != "Z-0002" ||
		repository.updatedAsset.Lifecycle != domain.AccessoryLifecycleMaintenance {
		t.Fatalf("unexpected normalized asset update: %#v", repository.updatedAsset)
	}
}
