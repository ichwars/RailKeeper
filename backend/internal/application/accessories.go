package application

import (
	"context"
	"errors"
	"fmt"
	"reflect"
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
	ID                 string                            `json:"id"`
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

func cleanAccessoryProductInput(input CreateAccessoryProductInput) CreateAccessoryProductInput {
	legacyInput := input.ArticleType == "" && input.Subtype == "" && input.InventoryStrategy == ""
	input.Manufacturer = strings.TrimSpace(input.Manufacturer)
	input.ArticleNumber = strings.TrimSpace(input.ArticleNumber)
	input.Name = strings.TrimSpace(input.Name)
	input.Category = strings.TrimSpace(input.Category)
	input.Description = strings.TrimSpace(input.Description)
	input.EAN = strings.TrimSpace(input.EAN)
	input.ManufacturerStatus = strings.TrimSpace(input.ManufacturerStatus)
	input.Subtype = strings.TrimSpace(input.Subtype)
	input.Scale = strings.TrimSpace(input.Scale)
	input.StockUnit = strings.TrimSpace(input.StockUnit)
	input.ManufacturerURL = strings.TrimSpace(input.ManufacturerURL)
	input.ProductURL = strings.TrimSpace(input.ProductURL)
	input.CompatibilityNotes = strings.TrimSpace(input.CompatibilityNotes)
	input.InternalNotes = strings.TrimSpace(input.InternalNotes)
	input.Gauges = cleanStringArray(input.Gauges)
	input.AlternativeNumbers = cleanStringArray(input.AlternativeNumbers)
	input.Keywords = cleanStringArray(input.Keywords)
	input.Attributes = cleanAccessoryAttributes(input.Attributes, input.ArticleType == domain.AccessoryArticleOther)
	if legacyInput {
		input.ArticleType = domain.AccessoryArticleOther
		input.Subtype = "other:other"
		input.InventoryStrategy = domain.InventoryStrategyFromTrackingMode(input.TrackingMode)
		input.PackageQuantity = 1
		input.StockUnit = "piece"
		return input
	}
	if input.InventoryStrategy.Valid() {
		input.TrackingMode = input.InventoryStrategy.TrackingMode()
	}
	if input.Category == "" {
		input.Category = input.Subtype
	}
	input.Subtype = normalizeAccessorySubtype(input.ArticleType, input.Subtype)
	return input
}

func validateAccessoryProductStructure(input CreateAccessoryProductInput) error {
	if !validAccessoryProductInput(input) {
		return ErrAccessoryValidation
	}
	if input.ArticleType != domain.AccessoryArticleOther {
		if err := domain.ValidateAccessoryAttributeValues(input.ArticleType, input.Attributes); err != nil {
			return ErrAccessoryValidation
		}
	}
	return nil
}

func validateAccessoryProductInput(
	input CreateAccessoryProductInput,
	state AccessoryProductMutationState,
) error {
	current := state.Current
	currentArticleType := domain.AccessoryArticleType("")
	currentSubtype := ""
	if current != nil {
		currentArticleType = domain.AccessoryArticleType(strings.TrimSpace(string(current.ArticleType)))
		currentSubtype = normalizeAccessorySubtype(currentArticleType, strings.TrimSpace(current.Subtype))
	}
	typeUnchanged := current != nil && currentArticleType == input.ArticleType
	if !typeUnchanged && !state.ArticleTypeActive {
		return ErrAccessoryValidation
	}
	if input.ArticleType == domain.AccessoryArticleOther {
		if err := validateControlledAccessoryAttributes(
			input.Attributes, current, state.CustomAttributeDefinitions,
		); err != nil {
			return err
		}
	}
	subtypeUnchanged := typeUnchanged && currentSubtype == input.Subtype
	if subtypeUnchanged {
		return nil
	}
	if !state.SubtypeActive {
		return ErrAccessoryValidation
	}
	return nil
}

func validateControlledAccessoryAttributes(
	values []domain.AccessoryAttributeValue,
	current *AccessoryProduct,
	definitions []domain.AccessoryAttributeDefinition,
) error {
	if err := domain.ValidateAccessoryAttributeValues(domain.AccessoryArticleOther, values); err != nil {
		return fmt.Errorf("%w: %v", ErrAccessoryValidation, err)
	}
	active := make([]domain.AccessoryAttributeDefinition, 0, len(definitions))
	inactiveByKey := make(map[string]domain.AccessoryAttributeDefinition)
	for _, definition := range definitions {
		if definition.Active {
			active = append(active, definition)
		} else {
			inactiveByKey[definition.Key] = definition
		}
	}

	valuesToValidate := values
	if current != nil && current.ArticleType == domain.AccessoryArticleOther {
		currentByKey := accessoryAttributesByKey(current.Attributes)
		valuesByKey := accessoryAttributesByKey(values)
		for key := range inactiveByKey {
			currentValue, existed := currentByKey[key]
			value, preserved := valuesByKey[key]
			if existed && (!preserved || !reflect.DeepEqual(value, currentValue)) {
				return fmt.Errorf("%w: inactive custom attribute %q must remain unchanged", ErrAccessoryValidation, key)
			}
			if preserved && (!existed || !reflect.DeepEqual(value, currentValue)) {
				return fmt.Errorf("%w: inactive custom attribute %q cannot be added or changed", ErrAccessoryValidation, key)
			}
		}
		valuesToValidate = make([]domain.AccessoryAttributeValue, 0, len(values))
		for _, value := range values {
			if _, inactive := inactiveByKey[value.Key]; !inactive {
				valuesToValidate = append(valuesToValidate, value)
			}
		}
	}
	if err := domain.ValidateControlledAccessoryAttributeValues(valuesToValidate, active); err != nil {
		return fmt.Errorf("%w: %v", ErrAccessoryValidation, err)
	}
	return nil
}

func accessoryAttributesByKey(values []domain.AccessoryAttributeValue) map[string]domain.AccessoryAttributeValue {
	out := make(map[string]domain.AccessoryAttributeValue, len(values))
	for _, value := range values {
		out[value.Key] = value
	}
	return out
}

func validAccessoryProductInput(input CreateAccessoryProductInput) bool {
	return input.Manufacturer != "" && input.Name != "" && input.Category != "" && input.TrackingMode.Valid() &&
		input.ArticleType.Valid() && accessorySubtypeMatchesType(input.ArticleType, input.Subtype) &&
		input.InventoryStrategy.Valid() && input.PackageQuantity > 0 && input.StockUnit != "" &&
		input.MinimumStock >= 0
}

func normalizeAccessorySubtype(articleType domain.AccessoryArticleType, subtype string) string {
	if subtype == "" || strings.Contains(subtype, ":") {
		return subtype
	}
	return string(articleType) + ":" + subtype
}

func accessorySubtypeMatchesType(articleType domain.AccessoryArticleType, subtype string) bool {
	if subtype == "" {
		return false
	}
	prefix := string(articleType) + ":"
	return strings.HasPrefix(subtype, prefix) && len(subtype) > len(prefix)
}

func cleanStringArray(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	cleaned := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || containsFold(cleaned, value) {
			continue
		}
		cleaned = append(cleaned, value)
	}
	return cleaned
}

func cleanAccessoryAttributes(
	values []domain.AccessoryAttributeValue,
	preserveOptionDuplicates bool,
) []domain.AccessoryAttributeValue {
	for index := range values {
		value := &values[index]
		value.Key = strings.TrimSpace(value.Key)
		if value.TextValue != nil {
			trimmed := strings.TrimSpace(*value.TextValue)
			value.TextValue = &trimmed
		}
		if value.DateValue != nil {
			trimmed := strings.TrimSpace(*value.DateValue)
			value.DateValue = &trimmed
		}
		if value.Unit != nil {
			trimmed := strings.TrimSpace(*value.Unit)
			value.Unit = &trimmed
		}
		if preserveOptionDuplicates {
			for optionIndex := range value.OptionValues {
				value.OptionValues[optionIndex] = strings.TrimSpace(value.OptionValues[optionIndex])
			}
		} else {
			value.OptionValues = cleanStringArray(value.OptionValues)
		}
	}
	return values
}

func containsFold(values []string, candidate string) bool {
	for _, value := range values {
		if strings.EqualFold(value, candidate) {
			return true
		}
	}
	return false
}

func cleanStorageLocationInput(input CreateStorageLocationInput) CreateStorageLocationInput {
	input.ParentID = strings.TrimSpace(input.ParentID)
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	return input
}
