package application

import (
	"context"
	"database/sql"
	"railkeeper/backend/internal/infrastructure/vehicledata"
	vehiclemodel "railkeeper/backend/internal/inventory/vehicle"
)

var (
	ErrVehicleFunctionValidation = vehiclemodel.ErrVehicleFunctionValidation
	ErrVehicleFunctionConflict   = vehiclemodel.ErrVehicleFunctionConflict
)

func validateVehicleFunctionChanges(changes []VehicleFunctionChange) error {
	return vehiclemodel.ValidateFunctionChanges(changes)
}
func applyVehicleFunctionChanges(ctx context.Context, tx *sql.Tx, vehicleID string, changes []VehicleFunctionChange, now string) error {
	return vehiclemodel.ApplyFunctionChanges(ctx, vehicledata.FunctionMutationFor(tx), vehicleID, changes, now)
}
