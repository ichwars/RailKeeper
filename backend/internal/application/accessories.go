package application

import (
	"context"
	"errors"
	"fmt"
	"railkeeper/backend/internal/domain"
	"strings"
)

var (
	ErrAccessoryValidation        = errors.New("accessory validation failed")
	ErrAccessoryNotFound          = errors.New("accessory resource not found")
	ErrAccessoryConflict          = errors.New("accessory resource conflict")
	ErrAccessoryDeleteBlocked     = errors.New("accessory deletion blocked")
	ErrAccessoryInsufficientStock = errors.New("insufficient accessory stock")
	ErrAccessoryTrackingMode      = errors.New("invalid accessory tracking mode")
)

type AccessoryProduct struct {
	ID                 string                            `json:"id"`
	InventoryNumber    string                            `json:"inventoryNumber"`
	Manufacturer       string                            `json:"manufacturer"`
	ArticleNumber      string                            `json:"articleNumber,omitempty"`
	Name               string                            `json:"name"`
	Category           string                            `json:"category"`
	TrackingMode       domain.AccessoryTrackingMode      `json:"trackingMode"`
	Description        string                            `json:"description,omitempty"`
	EAN                string                            `json:"ean,omitempty"`
	ManufacturerStatus string                            `json:"manufacturerStatus,omitempty"`
	ArticleType        domain.AccessoryArticleType       `json:"articleType"`
	Subtype            string                            `json:"subtype"`
	Gauges             []string                          `json:"gauges"`
	Scale              string                            `json:"scale,omitempty"`
	ListPrice          string                            `json:"listPrice,omitempty"`
	PackageQuantity    int                               `json:"packageQuantity"`
	StockUnit          string                            `json:"stockUnit"`
	MinimumStock       int                               `json:"minimumStock"`
	InventoryStrategy  domain.AccessoryInventoryStrategy `json:"inventoryStrategy"`
	ManufacturerURL    string                            `json:"manufacturerUrl,omitempty"`
	ProductURL         string                            `json:"productUrl,omitempty"`
	AlternativeNumbers []string                          `json:"alternativeNumbers"`
	Keywords           []string                          `json:"keywords"`
	CompatibilityNotes string                            `json:"compatibilityNotes,omitempty"`
	InternalNotes      string                            `json:"internalNotes,omitempty"`
	Archived           bool                              `json:"archived"`
	Attributes         []domain.AccessoryAttributeValue  `json:"attributes"`
	PrimaryImageURL    string                            `json:"primaryImageUrl,omitempty"`
	CreatedAt          string                            `json:"createdAt"`
	UpdatedAt          string                            `json:"updatedAt"`
}

type CreateAccessoryProductInput struct {
	Manufacturer       string                            `json:"manufacturer"`
	ArticleNumber      string                            `json:"articleNumber"`
	Name               string                            `json:"name"`
	Category           string                            `json:"category"`
	TrackingMode       domain.AccessoryTrackingMode      `json:"trackingMode"`
	Description        string                            `json:"description"`
	EAN                string                            `json:"ean"`
	ManufacturerStatus string                            `json:"manufacturerStatus"`
	ArticleType        domain.AccessoryArticleType       `json:"articleType"`
	Subtype            string                            `json:"subtype"`
	Gauges             []string                          `json:"gauges"`
	Scale              string                            `json:"scale"`
	ListPrice          string                            `json:"listPrice"`
	PackageQuantity    int                               `json:"packageQuantity"`
	StockUnit          string                            `json:"stockUnit"`
	MinimumStock       int                               `json:"minimumStock"`
	InventoryStrategy  domain.AccessoryInventoryStrategy `json:"inventoryStrategy"`
	ManufacturerURL    string                            `json:"manufacturerUrl"`
	ProductURL         string                            `json:"productUrl"`
	AlternativeNumbers []string                          `json:"alternativeNumbers"`
	Keywords           []string                          `json:"keywords"`
	CompatibilityNotes string                            `json:"compatibilityNotes"`
	InternalNotes      string                            `json:"internalNotes"`
	Archived           bool                              `json:"archived"`
	Attributes         []domain.AccessoryAttributeValue  `json:"attributes"`
}

type UpdateAccessoryProductInput struct {
	CreateAccessoryProductInput
}

type AccessoryDuplicateCheckInput struct {
	Manufacturer  string `json:"manufacturer"`
	ArticleNumber string `json:"articleNumber"`
	ExcludeID     string `json:"excludeId"`
}

type AccessoryDuplicateCandidate struct {
	ID            string                      `json:"id"`
	Manufacturer  string                      `json:"manufacturer"`
	ArticleNumber string                      `json:"articleNumber"`
	Name          string                      `json:"name"`
	ArticleType   domain.AccessoryArticleType `json:"articleType"`
	Subtype       string                      `json:"subtype"`
}

type AccessoryDuplicateCheckResult struct {
	Candidates []AccessoryDuplicateCandidate `json:"candidates"`
}

type AccessoryProductMutationState struct {
	Current                    *AccessoryProduct
	ArticleTypeActive          bool
	SubtypeActive              bool
	CustomAttributeDefinitions []domain.AccessoryAttributeDefinition
}

type AccessoryProductMutationValidator func(AccessoryProductMutationState) error

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

type AccessoryCatalogRepository interface {
	ListArticles(context.Context, AccessoryArticleListQuery) (*AccessoryArticleListResult, error)
	GetProduct(context.Context, string) (*AccessoryProduct, error)
	FindDuplicateCandidates(context.Context, string, string, string) ([]AccessoryDuplicateCandidate, error)
	CreateProduct(
		context.Context,
		CreateAccessoryProductInput,
		string,
		AccessoryProductMutationValidator,
	) (*AccessoryProduct, error)
	UpdateProduct(
		context.Context,
		string,
		UpdateAccessoryProductInput,
		string,
		AccessoryProductMutationValidator,
	) (*AccessoryProduct, error)
	SetProductArchived(context.Context, string, bool, string) (*AccessoryProduct, error)
	DeleteProduct(context.Context, string, string) ([]string, error)
}

type AccessoryLocationRepository interface {
	ListLocations(context.Context) ([]StorageLocation, error)
	CreateLocation(context.Context, CreateStorageLocationInput, string) (*StorageLocation, error)
	UpdateLocation(context.Context, string, UpdateStorageLocationInput, string) (*StorageLocation, error)
}

type AccessoryInventoryRepository interface {
	AdjustStock(context.Context, string, StockAdjustmentInput, string) (*AccessoryStockSummary, error)
	GetStock(context.Context, string) (*AccessoryStockSummary, error)
	ListStockMovements(context.Context, string) ([]AccessoryStockMovement, error)
	TransferStock(context.Context, string, TransferAccessoryStockInput, string) (*AccessoryStockSummary, error)
	ListAssets(context.Context, string) ([]AccessoryAsset, error)
	CreateAsset(context.Context, string, CreateAccessoryAssetInput, string) (*AccessoryAsset, error)
	UpdateAsset(context.Context, string, UpdateAccessoryAssetInput, string) (*AccessoryAsset, error)
	Individualize(context.Context, string, IndividualizeAccessoryInput, string) (*AccessoryAsset, error)
	ListPurchases(context.Context, string) ([]AccessoryPurchase, error)
	CreatePurchase(context.Context, string, CreateAccessoryPurchaseInput, string) (*AccessoryPurchase, error)
}

type AccessoryRepository interface {
	AccessoryCatalogRepository
	AccessoryLocationRepository
	AccessoryInventoryRepository
	// ListProducts remains until the legacy accessory route moves to ListArticles.
	ListProducts(context.Context, string) ([]AccessoryProduct, error)
}

type AccessoryService struct {
	repository AccessoryRepository
	blobs      FileBlobReferenceCleaner
}

func NewAccessoryService(
	repository AccessoryRepository,
	blobCleaners ...FileBlobReferenceCleaner,
) *AccessoryService {
	var blobs FileBlobReferenceCleaner
	if len(blobCleaners) > 0 {
		blobs = blobCleaners[0]
	}
	return &AccessoryService{repository: repository, blobs: blobs}
}

func (s *AccessoryService) ListProducts(ctx context.Context, query string) ([]AccessoryProduct, error) {
	return s.repository.ListProducts(ctx, strings.TrimSpace(query))
}

func (s *AccessoryService) GetProduct(ctx context.Context, id string) (*AccessoryProduct, error) {
	return s.repository.GetProduct(ctx, strings.TrimSpace(id))
}

func (s *AccessoryService) DeleteProduct(ctx context.Context, id, actor string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return ErrAccessoryValidation
	}
	blobIDs, err := s.repository.DeleteProduct(ctx, id, actor)
	if err != nil {
		return err
	}
	if s.blobs == nil {
		return nil
	}
	for _, blobID := range blobIDs {
		if err := s.blobs.DeleteIfUnreferenced(ctx, blobID); err != nil {
			return fmt.Errorf("delete accessory product blob %q: %w", blobID, err)
		}
	}
	return nil
}

func (s *AccessoryService) CreateProduct(
	ctx context.Context,
	input CreateAccessoryProductInput,
	actor string,
) (*AccessoryProduct, error) {
	input = cleanAccessoryProductInput(input)
	if err := validateAccessoryProductStructure(input); err != nil {
		return nil, err
	}
	return s.repository.CreateProduct(ctx, input, actor, func(state AccessoryProductMutationState) error {
		return validateAccessoryProductInput(input, state)
	})
}

func (s *AccessoryService) UpdateProduct(
	ctx context.Context,
	id string,
	input UpdateAccessoryProductInput,
	actor string,
) (*AccessoryProduct, error) {
	input.CreateAccessoryProductInput = cleanAccessoryProductInput(input.CreateAccessoryProductInput)
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrAccessoryValidation
	}
	if err := validateAccessoryProductStructure(input.CreateAccessoryProductInput); err != nil {
		return nil, err
	}
	return s.repository.UpdateProduct(ctx, id, input, actor, func(state AccessoryProductMutationState) error {
		return validateAccessoryProductInput(input.CreateAccessoryProductInput, state)
	})
}

func (s *AccessoryService) SetProductArchived(
	ctx context.Context,
	id string,
	archived bool,
	actor string,
) (*AccessoryProduct, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrAccessoryValidation
	}
	return s.repository.SetProductArchived(ctx, id, archived, actor)
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
