package application

import (
	"context"
	"strings"

	"railkeeper/backend/internal/domain"
)

type AccessoryStockLevel struct {
	LocationID   string `json:"locationId"`
	LocationName string `json:"locationName"`
	Quantity     int    `json:"quantity"`
	UpdatedAt    string `json:"updatedAt"`
}

type AccessoryStockSummary struct {
	ProductID     string                       `json:"productId"`
	TrackingMode  domain.AccessoryTrackingMode `json:"trackingMode"`
	TotalQuantity int                          `json:"totalQuantity"`
	Locations     []AccessoryStockLevel        `json:"locations"`
}

type StockAdjustmentInput struct {
	LocationID string `json:"locationId"`
	Delta      int    `json:"delta"`
}

type AccessoryAsset struct {
	ID                string                    `json:"id"`
	ProductID         string                    `json:"productId"`
	InventoryNumber   string                    `json:"inventoryNumber,omitempty"`
	SerialNumber      string                    `json:"serialNumber,omitempty"`
	Condition         domain.AccessoryCondition `json:"condition"`
	Lifecycle         domain.AccessoryLifecycle `json:"lifecycle"`
	StorageLocationID string                    `json:"storageLocationId,omitempty"`
	PurchaseDate      string                    `json:"purchaseDate,omitempty"`
	PurchasePrice     string                    `json:"purchasePrice,omitempty"`
	WarrantyUntil     string                    `json:"warrantyUntil,omitempty"`
	Notes             string                    `json:"notes,omitempty"`
	CreatedAt         string                    `json:"createdAt"`
	UpdatedAt         string                    `json:"updatedAt"`
}

type CreateAccessoryAssetInput struct {
	InventoryNumber   string                    `json:"inventoryNumber"`
	SerialNumber      string                    `json:"serialNumber"`
	Condition         domain.AccessoryCondition `json:"condition"`
	Lifecycle         domain.AccessoryLifecycle `json:"lifecycle"`
	StorageLocationID string                    `json:"storageLocationId"`
	PurchaseDate      string                    `json:"purchaseDate"`
	PurchasePrice     string                    `json:"purchasePrice"`
	WarrantyUntil     string                    `json:"warrantyUntil"`
	Notes             string                    `json:"notes"`
}

type UpdateAccessoryAssetInput struct {
	CreateAccessoryAssetInput
}

func (s *AccessoryService) AdjustStock(
	ctx context.Context,
	productID string,
	input StockAdjustmentInput,
	actor string,
) (*AccessoryStockSummary, error) {
	productID = strings.TrimSpace(productID)
	input.LocationID = strings.TrimSpace(input.LocationID)
	if productID == "" || input.LocationID == "" || input.Delta == 0 {
		return nil, ErrAccessoryValidation
	}
	return s.repository.AdjustStock(ctx, productID, input, actor)
}

func (s *AccessoryService) GetStock(ctx context.Context, productID string) (*AccessoryStockSummary, error) {
	return s.repository.GetStock(ctx, strings.TrimSpace(productID))
}

func (s *AccessoryService) ListAssets(ctx context.Context, productID string) ([]AccessoryAsset, error) {
	return s.repository.ListAssets(ctx, strings.TrimSpace(productID))
}

func (s *AccessoryService) CreateAsset(
	ctx context.Context,
	productID string,
	input CreateAccessoryAssetInput,
	actor string,
) (*AccessoryAsset, error) {
	productID = strings.TrimSpace(productID)
	input = cleanAccessoryAssetInput(input)
	if productID == "" || !validAccessoryAssetInput(input) {
		return nil, ErrAccessoryValidation
	}
	return s.repository.CreateAsset(ctx, productID, input, actor)
}

func (s *AccessoryService) UpdateAsset(
	ctx context.Context,
	id string,
	input UpdateAccessoryAssetInput,
	actor string,
) (*AccessoryAsset, error) {
	id = strings.TrimSpace(id)
	input.CreateAccessoryAssetInput = cleanAccessoryAssetInput(input.CreateAccessoryAssetInput)
	if id == "" || !validAccessoryAssetInput(input.CreateAccessoryAssetInput) {
		return nil, ErrAccessoryValidation
	}
	return s.repository.UpdateAsset(ctx, id, input, actor)
}

func cleanAccessoryAssetInput(input CreateAccessoryAssetInput) CreateAccessoryAssetInput {
	input.InventoryNumber = strings.TrimSpace(input.InventoryNumber)
	input.SerialNumber = strings.TrimSpace(input.SerialNumber)
	input.StorageLocationID = strings.TrimSpace(input.StorageLocationID)
	input.PurchaseDate = strings.TrimSpace(input.PurchaseDate)
	input.PurchasePrice = strings.TrimSpace(input.PurchasePrice)
	input.WarrantyUntil = strings.TrimSpace(input.WarrantyUntil)
	input.Notes = strings.TrimSpace(input.Notes)
	if input.Condition == "" {
		input.Condition = domain.AccessoryConditionUnknown
	}
	if input.Lifecycle == "" {
		input.Lifecycle = domain.AccessoryLifecycleStored
	}
	return input
}

func validAccessoryAssetInput(input CreateAccessoryAssetInput) bool {
	return input.Condition.Valid() && input.Lifecycle.Valid()
}
