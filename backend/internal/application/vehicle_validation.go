package application

import (
	"railkeeper/backend/internal/decoder"
	vehiclemodel "railkeeper/backend/internal/inventory/vehicle"
)

func cleanVehicleInput(input CreateVehicleInput) CreateVehicleInput {
	return vehiclemodel.CleanVehicleInput(input)
}

func isValidVehicleOperationalInput(input CreateVehicleInput) bool {
	return vehiclemodel.IsValidVehicleOperationalInput(input)
}

func cleanVehicleCVValueInput(input VehicleCVValueInput) VehicleCVValueInput {
	return decoder.CleanCVValueInput(input)
}

func isValidVehicleCVValueInput(input VehicleCVValueInput) bool {
	return decoder.ValidCVValueInput(input)
}

func validCVExportTimestamp(value string) bool { return decoder.ValidCVExportTimestamp(value) }

func cleanVehicleCVFileInput(input VehicleCVFileInput) VehicleCVFileInput {
	return decoder.CleanCVFileInput(input)
}

func validCVNumber(value int) bool { return decoder.ValidCVNumber(value) }

func validCVValue(value int) bool { return decoder.ValidCVValue(value) }
