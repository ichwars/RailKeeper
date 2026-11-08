package vehicle

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

var allowedFunctionTypes = map[string]struct{}{
	"kupplung":       {},
	"licht":          {},
	"rauch":          {},
	"sonderfunktion": {},
	"sound":          {},
	"standard":       {},
}
var allowedFunctionModes = map[string]struct{}{
	"dauer":  {},
	"moment": {},
}

func CleanVehicleInput(input CreateVehicleInput) CreateVehicleInput {
	input.InventoryNumber = strings.TrimSpace(input.InventoryNumber)
	input.Manufacturer = strings.TrimSpace(input.Manufacturer)
	input.ArticleNumber = strings.TrimSpace(input.ArticleNumber)
	input.ArticleSourceURL = strings.TrimSpace(input.ArticleSourceURL)
	input.Name = strings.TrimSpace(input.Name)
	input.Gauge = strings.TrimSpace(input.Gauge)
	input.Epoch = strings.TrimSpace(input.Epoch)
	input.RailwayCompany = strings.TrimSpace(input.RailwayCompany)
	input.Category = strings.TrimSpace(input.Category)
	input.Gattung = strings.TrimSpace(input.Gattung)
	input.Description = strings.TrimSpace(input.Description)
	input.Series = strings.TrimSpace(input.Series)
	input.VehicleNumber = strings.TrimSpace(input.VehicleNumber)
	input.HomeBase = strings.TrimSpace(input.HomeBase)
	input.DigitalDecoderNumber = strings.TrimSpace(input.DigitalDecoderNumber)
	input.DTDecoderNumber = strings.TrimSpace(input.DTDecoderNumber)
	input.DecoderType = strings.TrimSpace(input.DecoderType)
	input.EAN = strings.TrimSpace(input.EAN)
	input.ProductionPeriod = strings.TrimSpace(input.ProductionPeriod)
	input.ListPrice = strings.TrimSpace(input.ListPrice)
	input.AcquisitionType = strings.TrimSpace(input.AcquisitionType)
	input.AcquiredFrom = strings.TrimSpace(input.AcquiredFrom)
	input.PurchasePrice = strings.TrimSpace(input.PurchasePrice)
	input.PurchaseDate = strings.TrimSpace(input.PurchaseDate)
	input.StorageLocation = strings.TrimSpace(input.StorageLocation)
	input.StorageDetails = strings.TrimSpace(input.StorageDetails)
	input.Condition = strings.TrimSpace(input.Condition)
	input.ConditionDetails = strings.TrimSpace(input.ConditionDetails)
	input.Packaging = strings.TrimSpace(input.Packaging)
	input.LengthMM = strings.TrimSpace(input.LengthMM)
	input.WeightG = strings.TrimSpace(input.WeightG)
	input.Color = strings.TrimSpace(input.Color)
	input.Lettering = strings.TrimSpace(input.Lettering)
	input.Load = strings.TrimSpace(input.Load)
	input.Interior = strings.TrimSpace(input.Interior)
	input.Axles = strings.TrimSpace(input.Axles)
	input.AxleCount = strings.TrimSpace(input.AxleCount)
	input.TractionTireCount = strings.TrimSpace(input.TractionTireCount)
	input.Wheelset = strings.TrimSpace(input.Wheelset)
	input.CouplingFront = strings.TrimSpace(input.CouplingFront)
	input.CouplingRear = strings.TrimSpace(input.CouplingRear)
	input.PowerPickup = strings.TrimSpace(input.PowerPickup)
	input.Adapter = strings.TrimSpace(input.Adapter)
	input.DriveDescription = strings.TrimSpace(input.DriveDescription)
	input.HeadlightsDescription = strings.TrimSpace(input.HeadlightsDescription)
	input.LightingDescription = strings.TrimSpace(input.LightingDescription)
	input.SoundGeneratorDescription = strings.TrimSpace(input.SoundGeneratorDescription)
	input.SmokeGeneratorDescription = strings.TrimSpace(input.SmokeGeneratorDescription)
	input.AdditionalInfo = strings.TrimSpace(input.AdditionalInfo)
	input.Images = CleanVehicleImageInputs(input.Images)
	if input.CouplingSame {
		input.CouplingRear = input.CouplingFront
	}
	return input
}
func IsValidVehicleOperationalInput(input CreateVehicleInput) bool {
	if utf8.RuneCountInString(input.HomeBase) > 200 {
		return false
	}
	return input.MaximumSpeedKmh == nil ||
		(*input.MaximumSpeedKmh >= 1 && *input.MaximumSpeedKmh <= 1000)
}
func CleanVehicleImageInputs(images []VehicleImageInput) []VehicleImageInput {
	seen := map[string]bool{}
	cleaned := []VehicleImageInput{}
	hasPrimary := false
	for _, image := range images {
		image = CleanVehicleImageInput(image)
		if image.URL == "" {
			continue
		}
		key := strings.ToLower(image.URL)
		if seen[key] {
			continue
		}
		seen[key] = true
		if image.IsPrimary {
			if hasPrimary {
				image.IsPrimary = false
			} else {
				hasPrimary = true
			}
		}
		cleaned = append(cleaned, image)
		if len(cleaned) >= 12 {
			break
		}
	}
	if len(cleaned) > 0 && !hasPrimary {
		cleaned[0].IsPrimary = true
	}
	return cleaned
}
func CleanVehicleImageInput(image VehicleImageInput) VehicleImageInput {
	image.ID = strings.TrimSpace(image.ID)
	image.URL = strings.TrimSpace(image.URL)
	image.Title = strings.TrimSpace(image.Title)
	image.SourceURL = strings.TrimSpace(image.SourceURL)
	image.FileName = strings.TrimSpace(image.FileName)
	image.MimeType = strings.TrimSpace(image.MimeType)
	image.StoragePath = strings.TrimSpace(image.StoragePath)
	image.ThumbnailPath = strings.TrimSpace(image.ThumbnailPath)
	image.BlobID = strings.TrimSpace(image.BlobID)
	image.ThumbnailBlobID = strings.TrimSpace(image.ThumbnailBlobID)
	image.MaintenanceID = strings.TrimSpace(image.MaintenanceID)
	return image
}
func CleanVehicleAttachmentInput(input VehicleAttachmentInput) VehicleAttachmentInput {
	input.FileName = strings.TrimSpace(input.FileName)
	input.OriginalName = strings.TrimSpace(input.OriginalName)
	input.Description = strings.TrimSpace(input.Description)
	input.Category = strings.TrimSpace(input.Category)
	input.MimeType = strings.TrimSpace(input.MimeType)
	input.StoragePath = strings.TrimSpace(input.StoragePath)
	input.BlobID = strings.TrimSpace(input.BlobID)
	input.MaintenanceID = strings.TrimSpace(input.MaintenanceID)
	return input
}
func CleanVehicleFunctionInput(input VehicleFunctionInput) VehicleFunctionInput {
	input.Name = strings.TrimSpace(input.Name)
	input.SymbolKey = strings.TrimSpace(input.SymbolKey)
	input.FunctionType = strings.ToLower(strings.TrimSpace(input.FunctionType))
	input.Mode = strings.ToLower(strings.TrimSpace(input.Mode))
	input.Notes = strings.TrimSpace(input.Notes)
	if input.FunctionType == "" {
		input.FunctionType = "standard"
	}
	if input.Mode == "" {
		input.Mode = "dauer"
	}
	return input
}
func IsValidVehicleFunctionInput(input VehicleFunctionInput) bool {
	if _, ok := allowedFunctionTypes[input.FunctionType]; !ok {
		return false
	}
	if _, ok := allowedFunctionModes[input.Mode]; !ok {
		return false
	}
	return input.FlagsVersion >= 0 && input.FlagsVersion <= 2 && len(input.Name) <= 120 &&
		len(input.SymbolKey) <= 80 &&
		len(input.Notes) <= 1000
}
func NormalizeFunctionKey(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	if !strings.HasPrefix(value, "F") {
		return value
	}
	number, err := strconv.Atoi(strings.TrimPrefix(value, "F"))
	if err != nil {
		return value
	}
	return fmt.Sprintf("F%d", number)
}
func ValidFunctionKey(value string) bool {
	value = NormalizeFunctionKey(value)
	if !strings.HasPrefix(value, "F") {
		return false
	}
	number, err := strconv.Atoi(strings.TrimPrefix(value, "F"))
	return err == nil && number >= 0 && number <= 31
}
func FunctionSortOrder(value string) int {
	number, err := strconv.Atoi(strings.TrimPrefix(NormalizeFunctionKey(value), "F"))
	if err != nil {
		return 999
	}
	return number
}
func ValidateFunctionChanges(changes []VehicleFunctionChange) error {
	if len(changes) > 32 {
		return ErrVehicleFunctionValidation
	}
	seen := map[string]bool{}
	for _, change := range changes {
		key := NormalizeFunctionKey(change.FunctionKey)
		if !ValidFunctionKey(key) || seen[key] || change.Delete == (change.Value != nil) {
			return ErrVehicleFunctionValidation
		}
		seen[key] = true
		if change.ExpectedRevision != "" {
			data, err := hex.DecodeString(change.ExpectedRevision)
			if err != nil || len(data) != sha256.Size {
				return ErrVehicleFunctionValidation
			}
		}
		if change.Delete && change.ExpectedRevision == "" {
			return ErrVehicleFunctionValidation
		}
		if change.Value != nil && (change.Value.FlagsVersion != 2 ||
			!IsValidVehicleFunctionInput(CleanVehicleFunctionInput(*change.Value))) {
			return ErrVehicleFunctionValidation
		}
	}
	return nil
}
func ValidateCreateInput(input CreateVehicleInput) (CreateVehicleInput, error) {
	input = CleanVehicleInput(input)
	if input.Manufacturer == "" || input.Name == "" || input.Gauge == "" || input.Category == "" ||
		input.Gattung == "" {
		return CreateVehicleInput{}, ErrVehicleValidation
	}
	if !IsValidVehicleOperationalInput(input) {
		return CreateVehicleInput{}, ErrVehicleOperationalValidation
	}
	if err := ValidateFunctionChanges(input.FunctionChanges); err != nil {
		return CreateVehicleInput{}, err
	}
	return input, nil
}
