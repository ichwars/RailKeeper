package application

import (
	"railkeeper/backend/internal/transfer"
	transferpolicy "railkeeper/backend/internal/transfer/policy"
)

type DataTransferVehicleSetPreview = transfer.DataTransferVehicleSetPreview

func ValidateTransferVehicleSet(set TransferVehicleSet) error {
	return transferpolicy.ValidateTransferVehicleSet(set)
}
