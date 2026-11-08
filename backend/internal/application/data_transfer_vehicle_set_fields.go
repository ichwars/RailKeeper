package application

import (
	transferpolicy "railkeeper/backend/internal/transfer/policy"
)

func VehicleSetTransferFields() []VehicleTransferField {
	return transferpolicy.VehicleSetTransferFields()
}

func VehicleCSVTransferFields() []VehicleTransferField {
	return transferpolicy.VehicleCSVTransferFields()
}

func VehicleCSVTransferFieldByKey(key string) (VehicleTransferField, bool) {
	return transferpolicy.VehicleCSVTransferFieldByKey(key)
}

func PreserveUnmappedTransferVehicleSetFields(
	incoming TransferVehicleSet,
	existing TransferVehicleSet,
	mapping []DataTransferCSVColumnMapping,
) TransferVehicleSet {
	return transferpolicy.PreserveUnmappedTransferVehicleSetFields(incoming, existing, mapping)
}
