package application

import (
	"fmt"
	"strings"

	"railkeeper/backend/internal/domain"
)

// ValidateTransferVehicle applies the same aggregate validation as the regular vehicle create flow.
func ValidateTransferVehicle(vehicle TransferVehicle) error {
	if strings.TrimSpace(vehicle.InventoryNumber) == "" {
		return fmt.Errorf("validate transferred vehicle inventory number: %w", ErrVehicleValidation)
	}
	_, err := validateCreateVehicleInput(CreateVehicleInput{
		InventoryNumber: vehicle.InventoryNumber, Manufacturer: vehicle.Manufacturer,
		ArticleNumber: vehicle.ArticleNumber, ArticleSourceURL: vehicle.ArticleSourceURL, Name: vehicle.Name,
		Gauge: vehicle.Gauge, Epoch: vehicle.Epoch, RailwayCompany: vehicle.RailwayCompany,
		Category: vehicle.Category, Gattung: vehicle.Gattung, Description: vehicle.Description,
		Series: vehicle.Series, VehicleNumber: vehicle.VehicleNumber, MaximumSpeedKmh: vehicle.MaximumSpeedKmh,
		HomeBase: vehicle.HomeBase, Digital: vehicle.Digital, DigitalDecoderNumber: vehicle.DigitalDecoderNumber,
		DTDecoder: vehicle.DTDecoder, DTDecoderNumber: vehicle.DTDecoderNumber, DecoderType: vehicle.DecoderType,
		ExhibitionReady: vehicle.ExhibitionReady, Exhibition: vehicle.Exhibition, ABCBrakes: vehicle.ABCBrakes,
		EAN: vehicle.EAN, ProductionPeriod: vehicle.ProductionPeriod, ListPrice: vehicle.ListPrice,
		AcquisitionType: vehicle.AcquisitionType, AcquiredFrom: vehicle.AcquiredFrom,
		PurchasePrice: vehicle.PurchasePrice, PurchaseDate: vehicle.PurchaseDate,
		StorageLocation: vehicle.StorageLocation, StorageDetails: vehicle.StorageDetails,
		Condition: vehicle.Condition, ConditionDetails: vehicle.ConditionDetails, Packaging: vehicle.Packaging,
	})
	if err != nil {
		return fmt.Errorf("validate transferred vehicle: %w", err)
	}
	return nil
}

// ValidateTransferAccessory applies the regular product and asset validation before an import can mutate storage.
func ValidateTransferAccessory(accessory TransferAccessory) error {
	if strings.TrimSpace(accessory.InventoryNumber) == "" {
		return fmt.Errorf("validate transferred accessory inventory number: %w", ErrAccessoryValidation)
	}
	if strings.TrimSpace(accessory.Subtype) == "" &&
		domain.AccessoryArticleType(accessory.ArticleType) == domain.AccessoryArticleOther {
		accessory.Subtype = "other:other"
	}
	input := cleanAccessoryProductInput(CreateAccessoryProductInput{
		Manufacturer: accessory.Manufacturer, ArticleNumber: accessory.ArticleNumber, Name: accessory.Name,
		Category: accessory.Category, TrackingMode: domain.AccessoryTrackingMode(accessory.TrackingMode),
		Description: accessory.Description, EAN: accessory.EAN, ManufacturerStatus: accessory.ManufacturerStatus,
		ArticleType: domain.AccessoryArticleType(accessory.ArticleType), Subtype: accessory.Subtype,
		Gauges: accessory.Gauges, Scale: accessory.Scale, ListPrice: accessory.ListPrice,
		PackageQuantity: accessory.PackageQuantity, StockUnit: accessory.StockUnit,
		MinimumStock:      accessory.MinimumStock,
		InventoryStrategy: domain.AccessoryInventoryStrategy(accessory.InventoryStrategy),
		ManufacturerURL:   accessory.ManufacturerURL, ProductURL: accessory.ProductURL,
		AlternativeNumbers: accessory.AlternativeNumbers, Keywords: accessory.Keywords,
		CompatibilityNotes: accessory.CompatibilityNotes, InternalNotes: accessory.InternalNotes,
		Archived: accessory.Archived,
	})
	if err := validateAccessoryProductStructure(input); err != nil {
		return fmt.Errorf("validate transferred accessory product: %w", err)
	}
	if status := strings.TrimSpace(accessory.ManufacturerStatus); status != "" &&
		status != "announced" && status != "available" && status != "discontinued" && status != "unknown" {
		return fmt.Errorf("validate transferred accessory manufacturer status: %w", ErrAccessoryValidation)
	}
	for _, level := range accessory.Stock {
		if level.Quantity < 0 || (strings.TrimSpace(level.LocationID) == "" && strings.TrimSpace(level.LocationName) == "") {
			return fmt.Errorf("validate transferred accessory stock: %w", ErrAccessoryValidation)
		}
	}
	for _, asset := range accessory.Assets {
		input := cleanAccessoryAssetInput(CreateAccessoryAssetInput{
			InventoryNumber: asset.InventoryNumber, SerialNumber: asset.SerialNumber,
			Condition: domain.AccessoryCondition(asset.Condition), Lifecycle: domain.AccessoryLifecycle(asset.Lifecycle),
			StorageLocationID: asset.StorageLocationID, PurchaseDate: asset.PurchaseDate,
			PurchasePrice: asset.PurchasePrice, WarrantyUntil: asset.WarrantyUntil, Notes: asset.Notes,
		})
		persistedActiveAllocation := strings.TrimSpace(asset.ID) != "" &&
			(input.Lifecycle == domain.AccessoryLifecycleReserved ||
				input.Lifecycle == domain.AccessoryLifecycleInstalled)
		validAsset := validAccessoryAssetInput(input)
		if persistedActiveAllocation {
			validAsset = input.Condition.Valid() && input.Lifecycle.Valid() &&
				validOptionalAccessoryDate(input.PurchaseDate) && validOptionalAccessoryDate(input.WarrantyUntil)
		}
		if !validAsset {
			return fmt.Errorf("validate transferred accessory asset: %w", ErrAccessoryValidation)
		}
	}
	return nil
}

func TransferAccessoryReferenceKeys(accessory TransferAccessory) (domain.AccessoryArticleType, string) {
	articleType := domain.AccessoryArticleType(strings.TrimSpace(accessory.ArticleType))
	subtype := normalizeAccessorySubtype(articleType, strings.TrimSpace(accessory.Subtype))
	if subtype == "" && articleType == domain.AccessoryArticleOther {
		subtype = "other:other"
	}
	return articleType, subtype
}

// ValidateTransferAccessoryReferenceState mirrors the active article-type/subtype rules of regular mutations.
func ValidateTransferAccessoryReferenceState(
	accessory TransferAccessory,
	state AccessoryProductMutationState,
) error {
	articleType, subtype := TransferAccessoryReferenceKeys(accessory)
	currentType := domain.AccessoryArticleType("")
	currentSubtype := ""
	if state.Current != nil {
		currentType = domain.AccessoryArticleType(strings.TrimSpace(string(state.Current.ArticleType)))
		currentSubtype = normalizeAccessorySubtype(currentType, strings.TrimSpace(state.Current.Subtype))
	}
	typeUnchanged := state.Current != nil && currentType == articleType
	if !typeUnchanged && !state.ArticleTypeActive {
		return fmt.Errorf("validate transferred accessory article type: %w", ErrAccessoryValidation)
	}
	subtypeUnchanged := typeUnchanged && currentSubtype == subtype
	if !subtypeUnchanged && !state.SubtypeActive {
		return fmt.Errorf("validate transferred accessory subtype: %w", ErrAccessoryValidation)
	}
	return nil
}
