package application

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"railkeeper/backend/internal/domain"
)

type accessoryRepositorySpy struct {
	AccessoryRepository
	createdProduct        CreateAccessoryProductInput
	updatedProduct        UpdateAccessoryProductInput
	createdLocation       CreateStorageLocationInput
	updatedLocation       UpdateStorageLocationInput
	stockAdjustment       StockAdjustmentInput
	stockTransfer         TransferAccessoryStockInput
	createdAsset          CreateAccessoryAssetInput
	updatedAsset          UpdateAccessoryAssetInput
	individualization     IndividualizeAccessoryInput
	purchase              CreateAccessoryPurchaseInput
	activeSubtypeKeys     map[string]bool
	activeArticleTypeKeys map[domain.AccessoryArticleType]bool
	customAttributes      []domain.AccessoryAttributeDefinition
	currentProduct        *AccessoryProduct
}

func stringPointer(value string) *string { return &value }

func TestAccessoryServiceDerivesCompatibilityArticleDefaults(t *testing.T) {
	repository := &accessoryRepositorySpy{}
	service := NewAccessoryService(repository)

	if _, err := service.CreateProduct(t.Context(), CreateAccessoryProductInput{
		Manufacturer: " Tillig ", Name: " Gleis ", Category: " Gleismaterial ",
		TrackingMode: domain.AccessoryTrackingModeQuantity,
	}, "editor-1"); err != nil {
		t.Fatal(err)
	}
	product := repository.createdProduct
	if product.ArticleType != domain.AccessoryArticleOther || product.Subtype != "other:other" ||
		product.InventoryStrategy != domain.AccessoryInventoryQuantity || product.PackageQuantity != 1 ||
		product.StockUnit != "piece" {
		t.Fatalf("unexpected compatibility defaults: %#v", product)
	}
}

func (spy *accessoryRepositorySpy) AccessorySubtypeActive(_ context.Context, key string) (bool, error) {
	if spy.activeSubtypeKeys != nil {
		return spy.activeSubtypeKeys[key], nil
	}
	return key == "track:straight" || key == "other:other", nil
}

func (spy *accessoryRepositorySpy) AccessoryArticleTypeActive(
	_ context.Context,
	key domain.AccessoryArticleType,
) (bool, error) {
	if spy.activeArticleTypeKeys != nil {
		return spy.activeArticleTypeKeys[key], nil
	}
	return key == domain.AccessoryArticleTrack || key == domain.AccessoryArticleOther, nil
}

func (spy *accessoryRepositorySpy) AccessoryCustomAttributeDefinitions(
	context.Context,
) ([]domain.AccessoryAttributeDefinition, error) {
	return append([]domain.AccessoryAttributeDefinition(nil), spy.customAttributes...), nil
}

func (spy *accessoryRepositorySpy) GetProduct(_ context.Context, _ string) (*AccessoryProduct, error) {
	if spy.currentProduct != nil {
		copy := *spy.currentProduct
		return &copy, nil
	}
	return &AccessoryProduct{
		ArticleType: domain.AccessoryArticleOther,
		Subtype:     "other:other",
	}, nil
}

func TestAccessoryServiceEnforcesAuthoritativeArticleTypeLifecycle(t *testing.T) {
	valid := func(articleType domain.AccessoryArticleType, subtype string) CreateAccessoryProductInput {
		return CreateAccessoryProductInput{
			Manufacturer: "Tillig", Name: "Article", Category: subtype,
			TrackingMode: domain.AccessoryTrackingModeQuantity, ArticleType: articleType, Subtype: subtype,
			PackageQuantity: 1, StockUnit: "piece", InventoryStrategy: domain.AccessoryInventoryQuantity,
		}
	}
	repository := &accessoryRepositorySpy{
		activeArticleTypeKeys: map[domain.AccessoryArticleType]bool{
			domain.AccessoryArticleTrack:  true,
			domain.AccessoryArticleSignal: false,
		},
		activeSubtypeKeys: map[string]bool{
			"track:straight":        true,
			"signal:legacy":         false,
			"signal:inactive_other": false,
		},
	}
	service := NewAccessoryService(repository)

	if _, err := service.CreateProduct(t.Context(), valid(domain.AccessoryArticleSignal, "signal:legacy"), "editor"); !errors.Is(err, ErrAccessoryValidation) {
		t.Fatalf("create accepted inactive article type: %v", err)
	}
	if _, err := service.CreateProduct(t.Context(), valid(domain.AccessoryArticleTrack, "track:straight"), "editor"); err != nil {
		t.Fatalf("create rejected active configured type and subtype: %v", err)
	}

	repository.currentProduct = &AccessoryProduct{
		ID: "historical", ArticleType: domain.AccessoryArticleSignal, Subtype: "signal:legacy",
	}
	if _, err := service.UpdateProduct(t.Context(), "historical", UpdateAccessoryProductInput{
		CreateAccessoryProductInput: valid(domain.AccessoryArticleSignal, "signal:legacy"),
	}, "editor"); err != nil {
		t.Fatalf("update rejected unchanged historical inactive type and subtype: %v", err)
	}
	if _, err := service.UpdateProduct(t.Context(), "historical", UpdateAccessoryProductInput{
		CreateAccessoryProductInput: valid(domain.AccessoryArticleSignal, "signal:inactive_other"),
	}, "editor"); !errors.Is(err, ErrAccessoryValidation) {
		t.Fatalf("update accepted switch to inactive subtype: %v", err)
	}

	repository.currentProduct = &AccessoryProduct{
		ID: "active", ArticleType: domain.AccessoryArticleTrack, Subtype: "track:straight",
	}
	if _, err := service.UpdateProduct(t.Context(), "active", UpdateAccessoryProductInput{
		CreateAccessoryProductInput: valid(domain.AccessoryArticleSignal, "signal:legacy"),
	}, "editor"); !errors.Is(err, ErrAccessoryValidation) {
		t.Fatalf("update accepted switch to inactive article type: %v", err)
	}
}

func TestAccessoryServiceValidatesAndNormalizesArticleCore(t *testing.T) {
	repository := &accessoryRepositorySpy{}
	service := NewAccessoryService(repository)
	text := " TT Modellgleis "
	length := 166.0
	valid := CreateAccessoryProductInput{
		Manufacturer: " Tillig ", ArticleNumber: " 83101 ", Name: " Straight track ", Category: " Track ",
		TrackingMode: domain.AccessoryTrackingModeQuantity,
		ArticleType:  domain.AccessoryArticleTrack, Subtype: " track:straight ", Gauges: []string{" TT ", "", "TT"},
		PackageQuantity: 2, StockUnit: " piece ", MinimumStock: 0,
		InventoryStrategy:  domain.AccessoryInventoryQuantityLaterIndividual,
		AlternativeNumbers: []string{" 83101-A ", "83101-A"}, Keywords: []string{" track ", "Track", ""},
		Attributes: []domain.AccessoryAttributeValue{
			{Key: " trackSystem ", Kind: domain.AccessoryAttributeText, TextValue: &text},
			{Key: "lengthMm", Kind: domain.AccessoryAttributeNumber, NumberValue: &length, Unit: stringPointer(" mm ")},
		},
	}
	if _, err := service.CreateProduct(t.Context(), valid, "editor-1"); err != nil {
		t.Fatal(err)
	}
	product := repository.createdProduct
	if product.Subtype != "track:straight" || !slices.Equal(product.Gauges, []string{"TT"}) ||
		!slices.Equal(product.AlternativeNumbers, []string{"83101-A"}) ||
		!slices.Equal(product.Keywords, []string{"track"}) || product.Attributes[0].Key != "trackSystem" ||
		*product.Attributes[0].TextValue != "TT Modellgleis" || *product.Attributes[1].Unit != "mm" {
		t.Fatalf("unexpected normalized article input: %#v", product)
	}

	invalid := []CreateAccessoryProductInput{
		{Manufacturer: "Tillig", Name: "Missing new subtype", Category: "Track", TrackingMode: domain.AccessoryTrackingModeQuantity,
			ArticleType: domain.AccessoryArticleTrack, InventoryStrategy: domain.AccessoryInventoryQuantity},
		{Manufacturer: "Tillig", Name: "Wrong subtype", Category: "Track", TrackingMode: domain.AccessoryTrackingModeQuantity,
			ArticleType: domain.AccessoryArticleTrack, Subtype: "signal:main", InventoryStrategy: domain.AccessoryInventoryQuantity, PackageQuantity: 1, StockUnit: "piece"},
		{Manufacturer: "Tillig", Name: "Missing package", Category: "Track", TrackingMode: domain.AccessoryTrackingModeQuantity,
			ArticleType: domain.AccessoryArticleTrack, Subtype: "track:straight", InventoryStrategy: domain.AccessoryInventoryQuantity, StockUnit: "piece"},
		{Manufacturer: "Tillig", Name: "Negative minimum", Category: "Track", TrackingMode: domain.AccessoryTrackingModeQuantity,
			ArticleType: domain.AccessoryArticleTrack, Subtype: "track:straight", InventoryStrategy: domain.AccessoryInventoryQuantity, PackageQuantity: 1, StockUnit: "piece", MinimumStock: -1},
		{Manufacturer: "Tillig", Name: "Mismatched attribute", Category: "Track", TrackingMode: domain.AccessoryTrackingModeQuantity,
			ArticleType: domain.AccessoryArticleTrack, Subtype: "track:straight", InventoryStrategy: domain.AccessoryInventoryQuantity, PackageQuantity: 1, StockUnit: "piece",
			Attributes: []domain.AccessoryAttributeValue{{Key: "custom", Kind: domain.AccessoryAttributeText, TextValue: &text}}},
	}
	for _, input := range invalid {
		if _, err := service.CreateProduct(t.Context(), input, "editor-1"); !errors.Is(err, ErrAccessoryValidation) {
			t.Fatalf("expected article validation error for %#v, got %v", input, err)
		}
	}
}

func TestAccessoryServiceEnforcesControlledCustomAttributeDefinitions(t *testing.T) {
	text := "Club"
	number := 12.5
	boolean := false
	date := "2026-08-08"
	repository := &accessoryRepositorySpy{customAttributes: []domain.AccessoryAttributeDefinition{
		{Key: "text", Kind: domain.AccessoryAttributeText, Active: true},
		{Key: "number", Kind: domain.AccessoryAttributeNumber, Active: true, Unit: "mm", Minimum: floatPointer(10), Maximum: floatPointer(20)},
		{Key: "boolean", Kind: domain.AccessoryAttributeBoolean, Active: true},
		{Key: "date", Kind: domain.AccessoryAttributeDate, Active: true},
		{Key: "single", Kind: domain.AccessoryAttributeSingleSelect, Active: true, Options: []string{"DCC", "MM"}},
		{Key: "multi", Kind: domain.AccessoryAttributeMultiSelect, Active: true, Options: []string{"DCC", "MM"}},
	}}
	service := NewAccessoryService(repository)
	input := validOtherAccessoryProductInput([]domain.AccessoryAttributeValue{
		{Key: "text", Kind: domain.AccessoryAttributeText, TextValue: &text},
		{Key: "number", Kind: domain.AccessoryAttributeNumber, NumberValue: &number, Unit: stringPointer("mm")},
		{Key: "boolean", Kind: domain.AccessoryAttributeBoolean, BooleanValue: &boolean},
		{Key: "date", Kind: domain.AccessoryAttributeDate, DateValue: &date},
		{Key: "single", Kind: domain.AccessoryAttributeSingleSelect, OptionValues: []string{"DCC"}},
		{Key: "multi", Kind: domain.AccessoryAttributeMultiSelect, OptionValues: []string{"DCC", "MM"}},
	})
	if _, err := service.CreateProduct(t.Context(), input, "editor"); err != nil {
		t.Fatalf("configured custom attributes rejected: %v", err)
	}

	wrongKind := validOtherAccessoryProductInput([]domain.AccessoryAttributeValue{{
		Key: "text", Kind: domain.AccessoryAttributeNumber, NumberValue: &number,
	}})
	if _, err := service.CreateProduct(t.Context(), wrongKind, "editor"); !errors.Is(err, ErrAccessoryValidation) || !strings.Contains(err.Error(), "requires") {
		t.Fatalf("incompatible custom kind error = %v", err)
	}
	undefined := validOtherAccessoryProductInput([]domain.AccessoryAttributeValue{{
		Key: "undefined", Kind: domain.AccessoryAttributeText, TextValue: &text,
	}})
	if _, err := service.CreateProduct(t.Context(), undefined, "editor"); !errors.Is(err, ErrAccessoryValidation) || !strings.Contains(err.Error(), "undefined") {
		t.Fatalf("undefined custom field error = %v", err)
	}
}

func TestAccessoryServicePreservesOnlyExactHistoricalInactiveCustomAttributes(t *testing.T) {
	oldValue := "Legacy"
	changedValue := "Changed"
	inactive := domain.AccessoryAttributeDefinition{
		Key: "retired", Kind: domain.AccessoryAttributeText, Active: false,
	}
	currentAttribute := domain.AccessoryAttributeValue{
		Key: "retired", Kind: domain.AccessoryAttributeText, TextValue: &oldValue,
	}
	repository := &accessoryRepositorySpy{
		customAttributes: []domain.AccessoryAttributeDefinition{inactive},
		currentProduct: &AccessoryProduct{
			ID: "historical", ArticleType: domain.AccessoryArticleOther, Subtype: "other:other",
			Attributes: []domain.AccessoryAttributeValue{currentAttribute},
		},
	}
	service := NewAccessoryService(repository)

	unchanged := UpdateAccessoryProductInput{CreateAccessoryProductInput: validOtherAccessoryProductInput(
		[]domain.AccessoryAttributeValue{currentAttribute})}
	if _, err := service.UpdateProduct(t.Context(), "historical", unchanged, "editor"); err != nil {
		t.Fatalf("unchanged historical inactive custom value rejected: %v", err)
	}

	for name, attributes := range map[string][]domain.AccessoryAttributeValue{
		"changed": {{Key: "retired", Kind: domain.AccessoryAttributeText, TextValue: &changedValue}},
		"removed": {},
		"new use": {{Key: "retired", Kind: domain.AccessoryAttributeText, TextValue: &oldValue},
			{Key: "retired-new", Kind: domain.AccessoryAttributeText, TextValue: &oldValue}},
	} {
		t.Run(name, func(t *testing.T) {
			input := UpdateAccessoryProductInput{CreateAccessoryProductInput: validOtherAccessoryProductInput(attributes)}
			if _, err := service.UpdateProduct(t.Context(), "historical", input, "editor"); !errors.Is(err, ErrAccessoryValidation) {
				t.Fatalf("inactive custom attribute mutation error = %v", err)
			}
		})
	}
}

func validOtherAccessoryProductInput(attributes []domain.AccessoryAttributeValue) CreateAccessoryProductInput {
	return CreateAccessoryProductInput{
		Manufacturer: "Club", Name: "Other article", Category: "other", ArticleType: domain.AccessoryArticleOther,
		Subtype: "other:other", PackageQuantity: 1, StockUnit: "piece",
		InventoryStrategy: domain.AccessoryInventoryQuantity, Attributes: attributes,
	}
}

func floatPointer(value float64) *float64 { return &value }

func (spy *accessoryRepositorySpy) CreateProduct(
	ctx context.Context,
	input CreateAccessoryProductInput,
	_ string,
	validate AccessoryProductMutationValidator,
) (*AccessoryProduct, error) {
	state, err := spy.productMutationState(ctx, input, nil)
	if err != nil {
		return nil, err
	}
	if validate != nil {
		if err := validate(state); err != nil {
			return nil, err
		}
	}
	spy.createdProduct = input
	return &AccessoryProduct{Name: input.Name}, nil
}

func (spy *accessoryRepositorySpy) UpdateProduct(
	ctx context.Context,
	_ string,
	input UpdateAccessoryProductInput,
	_ string,
	validate AccessoryProductMutationValidator,
) (*AccessoryProduct, error) {
	current, err := spy.GetProduct(ctx, "")
	if err != nil {
		return nil, err
	}
	state, err := spy.productMutationState(ctx, input.CreateAccessoryProductInput, current)
	if err != nil {
		return nil, err
	}
	if validate != nil {
		if err := validate(state); err != nil {
			return nil, err
		}
	}
	spy.updatedProduct = input
	return &AccessoryProduct{Name: input.Name}, nil
}

func (spy *accessoryRepositorySpy) productMutationState(
	ctx context.Context,
	input CreateAccessoryProductInput,
	current *AccessoryProduct,
) (AccessoryProductMutationState, error) {
	articleTypeActive, err := spy.AccessoryArticleTypeActive(ctx, input.ArticleType)
	if err != nil {
		return AccessoryProductMutationState{}, err
	}
	subtypeActive, err := spy.AccessorySubtypeActive(ctx, input.Subtype)
	if err != nil {
		return AccessoryProductMutationState{}, err
	}
	definitions, err := spy.AccessoryCustomAttributeDefinitions(ctx)
	if err != nil {
		return AccessoryProductMutationState{}, err
	}
	return AccessoryProductMutationState{
		Current: current, ArticleTypeActive: articleTypeActive, SubtypeActive: subtypeActive,
		CustomAttributeDefinitions: definitions,
	}, nil
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

func TestAccessoryServiceValidatesAssetDates(t *testing.T) {
	service := NewAccessoryService(&accessoryRepositorySpy{})
	for _, input := range []CreateAccessoryAssetInput{
		{PurchaseDate: "08/01/2026"},
		{PurchaseDate: "2026-02-30"},
		{WarrantyUntil: "2028/08/01"},
	} {
		if _, err := service.CreateAsset(t.Context(), "product-1", input, "editor-1"); !errors.Is(err, ErrAccessoryValidation) {
			t.Fatalf("expected asset date validation error for %#v, got %v", input, err)
		}
		if _, err := service.UpdateAsset(t.Context(), "asset-1", UpdateAccessoryAssetInput{
			CreateAccessoryAssetInput: input,
		}, "editor-1"); !errors.Is(err, ErrAccessoryValidation) {
			t.Fatalf("expected asset update date validation error for %#v, got %v", input, err)
		}
	}
}
