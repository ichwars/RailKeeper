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

func cleanVehicleImageInputs(images []VehicleImageInput) []VehicleImageInput {
	return vehiclemodel.CleanVehicleImageInputs(images)
}

func cleanVehicleImageInput(image VehicleImageInput) VehicleImageInput {
	return vehiclemodel.CleanVehicleImageInput(image)
}

func cleanVehicleAttachmentInput(input VehicleAttachmentInput) VehicleAttachmentInput {
	return vehiclemodel.CleanVehicleAttachmentInput(input)
}

func cleanVehicleFunctionInput(input VehicleFunctionInput) VehicleFunctionInput {
	return vehiclemodel.CleanVehicleFunctionInput(input)
}

func isValidVehicleFunctionInput(input VehicleFunctionInput) bool {
	return vehiclemodel.IsValidVehicleFunctionInput(input)
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

func normalizeFunctionKey(value string) string { return vehiclemodel.NormalizeFunctionKey(value) }

func validFunctionKey(value string) bool { return vehiclemodel.ValidFunctionKey(value) }

func functionSortOrder(value string) int { return vehiclemodel.FunctionSortOrder(value) }

func validCVNumber(value int) bool { return decoder.ValidCVNumber(value) }

func validCVValue(value int) bool { return decoder.ValidCVValue(value) }
