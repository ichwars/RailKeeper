package application

import (
	"context"
	"database/sql"
	"railkeeper/backend/internal/infrastructure/vehicledata"
	vehiclemodel "railkeeper/backend/internal/inventory/vehicle"
)

func validateCreateVehicleInput(input CreateVehicleInput) (CreateVehicleInput, error) {
	return vehiclemodel.ValidateCreateInput(input)
}
func (s *VehicleService) insertVehicleTx(ctx context.Context, tx *sql.Tx, id string, input CreateVehicleInput, actor, now string) (*Vehicle, error) {
	return vehiclemodel.InsertVehicle(ctx, vehicledata.VehicleMutationFor(tx), id, input, actor, now)
}
