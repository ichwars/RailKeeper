package domain

import "strings"

const MaxVehicleSetMembers = 100

func CleanVehicleSetInput(input VehicleSetInput) VehicleSetInput {
	input.Name = strings.TrimSpace(input.Name)
	input.Manufacturer = strings.TrimSpace(input.Manufacturer)
	input.ArticleNumber = strings.TrimSpace(input.ArticleNumber)
	input.ArticleSourceURL = strings.TrimSpace(input.ArticleSourceURL)
	input.Gauge = strings.TrimSpace(input.Gauge)
	input.Epoch = strings.TrimSpace(input.Epoch)
	input.RailwayCompany = strings.TrimSpace(input.RailwayCompany)
	input.Category = strings.TrimSpace(input.Category)
	input.Gattung = strings.TrimSpace(input.Gattung)
	input.Description = strings.TrimSpace(input.Description)
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
	return input
}
func ValidVehicleSetInput(input VehicleSetInput) bool {
	return input.Name != "" && input.Manufacturer != "" && input.Gauge != "" && input.Category != "" &&
		input.Gattung != ""
}
