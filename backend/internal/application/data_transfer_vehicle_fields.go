package application

import (
	"railkeeper/backend/internal/transfer"
	transferpolicy "railkeeper/backend/internal/transfer/policy"
)

type VehicleTransferValueKind = transfer.VehicleTransferValueKind
type VehicleTransferField = transfer.VehicleTransferField
type DataTransferCSVMappingOrigin = transfer.DataTransferCSVMappingOrigin
type DataTransferCSVColumnMapping = transfer.DataTransferCSVColumnMapping
type DataTransferCSVMappingInput = transfer.DataTransferCSVMappingInput

const (
	VehicleTransferString  = transfer.VehicleTransferString
	VehicleTransferInteger = transfer.VehicleTransferInteger
	VehicleTransferBoolean = transfer.VehicleTransferBoolean
	CSVMappingAlias        = transfer.CSVMappingAlias
	CSVMappingProfile      = transfer.CSVMappingProfile
	CSVMappingManual       = transfer.CSVMappingManual
	CSVMappingIgnored      = transfer.CSVMappingIgnored
	CSVMappingUnmapped     = transfer.CSVMappingUnmapped
)

func VehicleTransferFields() []VehicleTransferField { return transferpolicy.VehicleTransferFields() }

func VehicleTransferFieldByKey(key string) (VehicleTransferField, bool) {
	return transferpolicy.VehicleTransferFieldByKey(key)
}

func PreserveUnmappedTransferVehicleFields(
	incoming TransferVehicle,
	existing TransferVehicle,
	mapping []DataTransferCSVColumnMapping,
) (TransferVehicle, error) {
	return transferpolicy.PreserveUnmappedTransferVehicleFields(incoming, existing, mapping)
}
