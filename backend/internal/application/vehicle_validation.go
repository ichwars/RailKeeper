package application

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func cleanVehicleInput(input CreateVehicleInput) CreateVehicleInput {
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
	input.Images = cleanVehicleImageInputs(input.Images)
	if input.CouplingSame {
		input.CouplingRear = input.CouplingFront
	}
	return input
}

func cleanVehicleImageInputs(images []VehicleImageInput) []VehicleImageInput {
	seen := map[string]bool{}
	cleaned := []VehicleImageInput{}
	hasPrimary := false
	for _, image := range images {
		image = cleanVehicleImageInput(image)
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

func cleanVehicleImageInput(image VehicleImageInput) VehicleImageInput {
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

func cleanVehicleAttachmentInput(input VehicleAttachmentInput) VehicleAttachmentInput {
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

func cleanVehicleMaintenanceInput(input VehicleMaintenanceInput) VehicleMaintenanceInput {
	input.Kind = strings.TrimSpace(input.Kind)
	input.Status = normalizeMaintenanceStatus(input.Status)
	input.ConditionRating = strings.TrimSpace(input.ConditionRating)
	input.DueDate = strings.TrimSpace(input.DueDate)
	input.CompletedAt = strings.TrimSpace(input.CompletedAt)
	input.Cost = cleanMaintenanceCost(input.Cost)
	input.Notes = strings.TrimSpace(input.Notes)
	if input.Kind == "" {
		input.Kind = "Wartung"
	}
	if input.Status == "" {
		input.Status = "geplant"
	}
	return input
}

func normalizeMaintenanceStatus(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "fällig", "faellig":
		return "faellig"
	case "erledigt":
		return "erledigt"
	case "geplant", "":
		return value
	default:
		return value
	}
}

func cleanMaintenanceCost(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimSuffix(value, "€")
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, " ", "")
	return value
}

func cleanVehicleSparePartInput(input VehicleSparePartInput) VehicleSparePartInput {
	input.ArticleNumber = strings.TrimSpace(input.ArticleNumber)
	input.Description = strings.TrimSpace(input.Description)
	input.Price = strings.TrimSpace(input.Price)
	input.URL = strings.TrimSpace(input.URL)
	return input
}

func vehicleSparePartDedupKey(articleNumber, description, link string) string {
	article := strings.ToLower(strings.TrimSpace(articleNumber))
	article = strings.TrimPrefix(article, "et")
	article = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, article)
	if article != "" {
		return "article:" + article
	}
	link = strings.TrimRight(strings.ToLower(strings.TrimSpace(link)), "/")
	if link != "" {
		return "url:" + link
	}
	description = strings.ToLower(strings.TrimSpace(description))
	if description != "" {
		return "description:" + description
	}
	return ""
}

func isValidVehicleSparePartInput(input VehicleSparePartInput) bool {
	return input.ArticleNumber != "" || input.Description != "" || input.URL != ""
}

func isValidVehicleMaintenanceInput(input VehicleMaintenanceInput) bool {
	if _, ok := allowedMaintenanceKinds[input.Kind]; !ok {
		return false
	}
	if _, ok := allowedMaintenanceStatuses[input.Status]; !ok {
		return false
	}
	if input.ConditionRating != "" {
		if _, ok := allowedConditionRatings[input.ConditionRating]; !ok {
			return false
		}
	}
	return isValidDateOnly(input.DueDate) &&
		isValidDateOnly(input.CompletedAt) &&
		isValidMaintenanceCost(input.Cost) &&
		len(input.Notes) <= 4000
}

func isValidDateOnly(value string) bool {
	if value == "" {
		return true
	}
	parsed, err := time.Parse("2006-01-02", value)
	return err == nil && parsed.Format("2006-01-02") == value
}

func isValidMaintenanceCost(value string) bool {
	if value == "" {
		return true
	}
	normalized := strings.ReplaceAll(value, ",", ".")
	amount, err := strconv.ParseFloat(normalized, 64)
	return err == nil && amount >= 0
}

func cleanVehicleFunctionInput(input VehicleFunctionInput) VehicleFunctionInput {
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

func isValidVehicleFunctionInput(input VehicleFunctionInput) bool {
	if _, ok := allowedFunctionTypes[input.FunctionType]; !ok {
		return false
	}
	if _, ok := allowedFunctionModes[input.Mode]; !ok {
		return false
	}
	return len(input.Name) <= 120 &&
		len(input.SymbolKey) <= 80 &&
		len(input.Notes) <= 1000
}

func cleanVehicleCVValueInput(input VehicleCVValueInput) VehicleCVValueInput {
	input.Description = strings.TrimSpace(input.Description)
	input.Category = strings.TrimSpace(input.Category)
	input.Protocol = strings.TrimSpace(input.Protocol)
	input.DecoderProfile = strings.TrimSpace(input.DecoderProfile)
	input.SourceFileID = strings.TrimSpace(input.SourceFileID)
	return input
}

func isValidVehicleCVValueInput(input VehicleCVValueInput) bool {
	return validCVNumber(input.CVNumber) &&
		validCVValue(input.Value) &&
		len(input.Description) <= 1000 &&
		len(input.Category) <= 80 &&
		len(input.Protocol) <= 80 &&
		len(input.DecoderProfile) <= 160 &&
		len(input.SourceFileID) <= 80
}

func cleanVehicleCVFileInput(input VehicleCVFileInput) VehicleCVFileInput {
	input.FileName = strings.TrimSpace(input.FileName)
	input.OriginalName = strings.TrimSpace(input.OriginalName)
	input.Description = strings.TrimSpace(input.Description)
	input.DecoderProfile = strings.TrimSpace(input.DecoderProfile)
	input.MimeType = strings.TrimSpace(input.MimeType)
	input.StoragePath = strings.TrimSpace(input.StoragePath)
	input.BlobID = strings.TrimSpace(input.BlobID)
	return input
}

func normalizeFunctionKey(value string) string {
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

func validFunctionKey(value string) bool {
	value = normalizeFunctionKey(value)
	if !strings.HasPrefix(value, "F") {
		return false
	}
	number, err := strconv.Atoi(strings.TrimPrefix(value, "F"))
	return err == nil && number >= 0 && number <= 31
}

func functionSortOrder(value string) int {
	number, err := strconv.Atoi(strings.TrimPrefix(normalizeFunctionKey(value), "F"))
	if err != nil {
		return 999
	}
	return number
}

func validCVNumber(value int) bool {
	return value >= 1 && value <= 1024
}

func validCVValue(value int) bool {
	return value >= 0 && value <= 255
}
