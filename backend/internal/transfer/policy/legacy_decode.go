package policy

import (
	"railkeeper/backend/internal/transfer"
)

func transferVehicleFromV1(vehicle transferVehicleV1) TransferVehicle {
	return transfer.TransferVehicleFromV1(vehicle)
}
