package vehicle

import (
	"fmt"
	"strings"

	"railkeeper/backend/internal/domain"
)

const MaxVehicleSetMembers = domain.MaxVehicleSetMembers

func cleanVehicleSetInput(input VehicleSetInput) VehicleSetInput {
	return domain.CleanVehicleSetInput(input)
}

func isValidVehicleSetInput(input VehicleSetInput) bool { return domain.ValidVehicleSetInput(input) }

func mergeVehicleSetMember(setInput VehicleSetInput, member CreateVehicleInput, index int) CreateVehicleInput {
	if strings.TrimSpace(member.Name) == "" {
		member.Name = fmt.Sprintf("%s (%d)", setInput.Name, index+1)
	}
	return applyVehicleSetDefaults(setInput, member)
}

func applyVehicleSetDefaults(setInput VehicleSetInput, member CreateVehicleInput) CreateVehicleInput {
	member.Manufacturer = vehicleSetDefault(member.Manufacturer, setInput.Manufacturer)
	member.ArticleNumber = vehicleSetDefault(member.ArticleNumber, setInput.ArticleNumber)
	member.ArticleSourceURL = vehicleSetDefault(member.ArticleSourceURL, setInput.ArticleSourceURL)
	member.Gauge = vehicleSetDefault(member.Gauge, setInput.Gauge)
	member.Epoch = vehicleSetDefault(member.Epoch, setInput.Epoch)
	member.RailwayCompany = vehicleSetDefault(member.RailwayCompany, setInput.RailwayCompany)
	member.Category = vehicleSetDefault(member.Category, setInput.Category)
	member.Gattung = vehicleSetDefault(member.Gattung, setInput.Gattung)
	member.Description = vehicleSetDefault(member.Description, setInput.Description)
	member.EAN = vehicleSetDefault(member.EAN, setInput.EAN)
	member.ProductionPeriod = vehicleSetDefault(member.ProductionPeriod, setInput.ProductionPeriod)
	member.ListPrice = vehicleSetDefault(member.ListPrice, setInput.ListPrice)
	member.AcquisitionType = vehicleSetDefault(member.AcquisitionType, setInput.AcquisitionType)
	member.AcquiredFrom = vehicleSetDefault(member.AcquiredFrom, setInput.AcquiredFrom)
	member.PurchasePrice = vehicleSetDefault(member.PurchasePrice, setInput.PurchasePrice)
	member.PurchaseDate = vehicleSetDefault(member.PurchaseDate, setInput.PurchaseDate)
	member.StorageLocation = vehicleSetDefault(member.StorageLocation, setInput.StorageLocation)
	member.StorageDetails = vehicleSetDefault(member.StorageDetails, setInput.StorageDetails)
	member.Condition = vehicleSetDefault(member.Condition, setInput.Condition)
	member.ConditionDetails = vehicleSetDefault(member.ConditionDetails, setInput.ConditionDetails)
	member.Packaging = vehicleSetDefault(member.Packaging, setInput.Packaging)
	return member
}

func vehicleSetDefault(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func vehicleSetFromInput(
	setID string,
	inventoryNumber string,
	input VehicleSetInput,
	members []Vehicle,
	now string,
) *VehicleSet {
	return &VehicleSet{
		ID: setID, InventoryNumber: inventoryNumber, Name: input.Name, Manufacturer: input.Manufacturer,
		ArticleNumber:    input.ArticleNumber,
		ArticleSourceURL: input.ArticleSourceURL, Gauge: input.Gauge, Epoch: input.Epoch,
		RailwayCompany: input.RailwayCompany, Category: input.Category, Gattung: input.Gattung,
		Description: input.Description, EAN: input.EAN, ProductionPeriod: input.ProductionPeriod,
		ListPrice: input.ListPrice, AcquisitionType: input.AcquisitionType, AcquiredFrom: input.AcquiredFrom,
		PurchasePrice: input.PurchasePrice, PurchaseDate: input.PurchaseDate,
		StorageLocation: input.StorageLocation, StorageDetails: input.StorageDetails,
		Condition: input.Condition, ConditionDetails: input.ConditionDetails, Packaging: input.Packaging,
		Members: members, CreatedAt: now, UpdatedAt: now, MainImageMode: VehicleSetMainImageModeAutomatic,
	}
}

func vehicleSetSummary(
	setID string,
	inventoryNumber string,
	input VehicleSetInput,
	memberCount int,
	position int,
) *VehicleSetSummary {
	return &VehicleSetSummary{
		ID: setID, InventoryNumber: inventoryNumber, Name: input.Name, Manufacturer: input.Manufacturer,
		ArticleNumber: input.ArticleNumber, Gauge: input.Gauge, Epoch: input.Epoch,
		AcquisitionType: input.AcquisitionType, PurchaseDate: input.PurchaseDate,
		PurchasePrice: input.PurchasePrice, Condition: input.Condition,
		MemberCount: memberCount, Position: position,
	}
}
