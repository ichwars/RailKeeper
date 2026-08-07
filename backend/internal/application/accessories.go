package application

import (
	"context"
	"errors"
	"strings"

	"railkeeper/backend/internal/domain"
)

var (
	ErrAccessoryValidation        = errors.New("accessory validation failed")
	ErrAccessoryNotFound          = errors.New("accessory resource not found")
	ErrAccessoryConflict          = errors.New("accessory resource conflict")
	ErrAccessoryInsufficientStock = errors.New("insufficient accessory stock")
	ErrAccessoryTrackingMode      = errors.New("invalid accessory tracking mode")
)

type AccessoryProduct struct {
	ID            string                       `json:"id"`
	Manufacturer  string                       `json:"manufacturer"`
	ArticleNumber string                       `json:"articleNumber,omitempty"`
	Name          string                       `json:"name"`
	Category      string                       `json:"category"`
	TrackingMode  domain.AccessoryTrackingMode `json:"trackingMode"`
	Description   string                       `json:"description,omitempty"`
	CreatedAt     string                       `json:"createdAt"`
	UpdatedAt     string                       `json:"updatedAt"`
}

type CreateAccessoryProductInput struct {
	Manufacturer  string                       `json:"manufacturer"`
	ArticleNumber string                       `json:"articleNumber"`
	Name          string                       `json:"name"`
	Category      string                       `json:"category"`
	TrackingMode  domain.AccessoryTrackingMode `json:"trackingMode"`
	Description   string                       `json:"description"`
}

type UpdateAccessoryProductInput struct {
	CreateAccessoryProductInput
}

type StorageLocation struct {
	ID          string `json:"id"`
	ParentID    string `json:"parentId,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Archived    bool   `json:"archived"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type CreateStorageLocationInput struct {
	ParentID    string `json:"parentId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Archived    bool   `json:"archived"`
}

type UpdateStorageLocationInput struct {
	CreateStorageLocationInput
}

type AccessoryRepository interface {
	ListProducts(context.Context, string) ([]AccessoryProduct, error)
	GetProduct(context.Context, string) (*AccessoryProduct, error)
	CreateProduct(context.Context, CreateAccessoryProductInput, string) (*AccessoryProduct, error)
	UpdateProduct(context.Context, string, UpdateAccessoryProductInput, string) (*AccessoryProduct, error)
	ListLocations(context.Context) ([]StorageLocation, error)
	CreateLocation(context.Context, CreateStorageLocationInput, string) (*StorageLocation, error)
	UpdateLocation(context.Context, string, UpdateStorageLocationInput, string) (*StorageLocation, error)
	AdjustStock(context.Context, string, StockAdjustmentInput, string) (*AccessoryStockSummary, error)
	GetStock(context.Context, string) (*AccessoryStockSummary, error)
	ListAssets(context.Context, string) ([]AccessoryAsset, error)
	CreateAsset(context.Context, string, CreateAccessoryAssetInput, string) (*AccessoryAsset, error)
	UpdateAsset(context.Context, string, UpdateAccessoryAssetInput, string) (*AccessoryAsset, error)
}

type AccessoryService struct {
	repository AccessoryRepository
}

func NewAccessoryService(repository AccessoryRepository) *AccessoryService {
	return &AccessoryService{repository: repository}
}

func (s *AccessoryService) ListProducts(ctx context.Context, query string) ([]AccessoryProduct, error) {
	return s.repository.ListProducts(ctx, strings.TrimSpace(query))
}

func (s *AccessoryService) GetProduct(ctx context.Context, id string) (*AccessoryProduct, error) {
	return s.repository.GetProduct(ctx, strings.TrimSpace(id))
}

func (s *AccessoryService) CreateProduct(
	ctx context.Context,
	input CreateAccessoryProductInput,
	actor string,
) (*AccessoryProduct, error) {
	input = cleanAccessoryProductInput(input)
	if !validAccessoryProductInput(input) {
		return nil, ErrAccessoryValidation
	}
	return s.repository.CreateProduct(ctx, input, actor)
}

func (s *AccessoryService) UpdateProduct(
	ctx context.Context,
	id string,
	input UpdateAccessoryProductInput,
	actor string,
) (*AccessoryProduct, error) {
	input.CreateAccessoryProductInput = cleanAccessoryProductInput(input.CreateAccessoryProductInput)
	id = strings.TrimSpace(id)
	if id == "" || !validAccessoryProductInput(input.CreateAccessoryProductInput) {
		return nil, ErrAccessoryValidation
	}
	return s.repository.UpdateProduct(ctx, id, input, actor)
}

func (s *AccessoryService) ListLocations(ctx context.Context) ([]StorageLocation, error) {
	return s.repository.ListLocations(ctx)
}

func (s *AccessoryService) CreateLocation(
	ctx context.Context,
	input CreateStorageLocationInput,
	actor string,
) (*StorageLocation, error) {
	input = cleanStorageLocationInput(input)
	if input.Name == "" {
		return nil, ErrAccessoryValidation
	}
	return s.repository.CreateLocation(ctx, input, actor)
}

func (s *AccessoryService) UpdateLocation(
	ctx context.Context,
	id string,
	input UpdateStorageLocationInput,
	actor string,
) (*StorageLocation, error) {
	id = strings.TrimSpace(id)
	input.CreateStorageLocationInput = cleanStorageLocationInput(input.CreateStorageLocationInput)
	if id == "" || input.Name == "" || input.ParentID == id {
		return nil, ErrAccessoryValidation
	}
	return s.repository.UpdateLocation(ctx, id, input, actor)
}

func cleanAccessoryProductInput(input CreateAccessoryProductInput) CreateAccessoryProductInput {
	input.Manufacturer = strings.TrimSpace(input.Manufacturer)
	input.ArticleNumber = strings.TrimSpace(input.ArticleNumber)
	input.Name = strings.TrimSpace(input.Name)
	input.Category = strings.TrimSpace(input.Category)
	input.Description = strings.TrimSpace(input.Description)
	return input
}

func validAccessoryProductInput(input CreateAccessoryProductInput) bool {
	return input.Manufacturer != "" && input.Name != "" && input.Category != "" && input.TrackingMode.Valid()
}

func cleanStorageLocationInput(input CreateStorageLocationInput) CreateStorageLocationInput {
	input.ParentID = strings.TrimSpace(input.ParentID)
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	return input
}
