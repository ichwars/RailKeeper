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

type TransferAccessoryStockInput struct {
	FromLocationID string `json:"fromLocationId"`
	ToLocationID   string `json:"toLocationId"`
	Quantity       int    `json:"quantity"`
	Note           string `json:"note"`
}

type IndividualizeAccessoryInput struct {
	LocationID string                    `json:"locationId"`
	Asset      CreateAccessoryAssetInput `json:"asset"`
}

type AccessoryStockMovement struct {
	ID           string `json:"id"`
	ProductID    string `json:"productId"`
	LocationID   string `json:"locationId"`
	MovementType string `json:"movementType"`
	Quantity     int    `json:"quantity"`
	SourceType   string `json:"sourceType,omitempty"`
	SourceID     string `json:"sourceId,omitempty"`
	Actor        string `json:"actor,omitempty"`
	Note         string `json:"note,omitempty"`
	CreatedAt    string `json:"createdAt"`
}

type AccessoryAsset struct {
	ID                string                    `json:"id"`
	ProductID         string                    `json:"productId"`
	PurchaseID        string                    `json:"purchaseId,omitempty"`
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
	productID = strings.TrimSpace(productID)
	if productID == "" {
		return nil, ErrAccessoryValidation
	}
	return s.repository.GetStock(ctx, productID)
}

func (s *AccessoryService) ListStockMovements(
	ctx context.Context,
	productID string,
) ([]AccessoryStockMovement, error) {
	productID = strings.TrimSpace(productID)
	if productID == "" {
		return nil, ErrAccessoryValidation
	}
	return s.repository.ListStockMovements(ctx, productID)
}

func (s *AccessoryService) TransferStock(
	ctx context.Context,
	productID string,
	input TransferAccessoryStockInput,
	actor string,
) (*AccessoryStockSummary, error) {
	productID = strings.TrimSpace(productID)
	input.FromLocationID = strings.TrimSpace(input.FromLocationID)
	input.ToLocationID = strings.TrimSpace(input.ToLocationID)
	input.Note = strings.TrimSpace(input.Note)
	if productID == "" || input.FromLocationID == "" || input.ToLocationID == "" ||
		input.FromLocationID == input.ToLocationID || input.Quantity <= 0 {
		return nil, ErrAccessoryValidation
	}
	return s.repository.TransferStock(ctx, productID, input, actor)
}

func (s *AccessoryService) ListAssets(ctx context.Context, productID string) ([]AccessoryAsset, error) {
	productID = strings.TrimSpace(productID)
	if productID == "" {
		return nil, ErrAccessoryValidation
	}
	return s.repository.ListAssets(ctx, productID)
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

func (s *AccessoryService) Individualize(
	ctx context.Context,
	productID string,
	input IndividualizeAccessoryInput,
	actor string,
) (*AccessoryAsset, error) {
	productID = strings.TrimSpace(productID)
	input.LocationID = strings.TrimSpace(input.LocationID)
	input.Asset = cleanAccessoryAssetInput(input.Asset)
	input.Asset.StorageLocationID = input.LocationID
	if productID == "" || input.LocationID == "" || !validAccessoryAssetInput(input.Asset) ||
		input.Asset.Lifecycle != domain.AccessoryLifecycleStored ||
		!validOptionalAccessoryDate(input.Asset.PurchaseDate) ||
		!validOptionalAccessoryDate(input.Asset.WarrantyUntil) {
		return nil, ErrAccessoryValidation
	}
	return s.repository.Individualize(ctx, productID, input, actor)
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
	return input.Condition.Valid() && input.Lifecycle.Valid() &&
		input.Lifecycle != domain.AccessoryLifecycleReserved &&
		input.Lifecycle != domain.AccessoryLifecycleInstalled &&
		validOptionalAccessoryDate(input.PurchaseDate) &&
		validOptionalAccessoryDate(input.WarrantyUntil)
}
