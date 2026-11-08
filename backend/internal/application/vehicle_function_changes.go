package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	vehiclemodel "railkeeper/backend/internal/inventory/vehicle"
)

var (
	ErrVehicleFunctionValidation = vehiclemodel.ErrVehicleFunctionValidation
	ErrVehicleFunctionConflict   = vehiclemodel.ErrVehicleFunctionConflict
)

func vehicleFunctionRevision(item VehicleFunction) string { return vehiclemodel.FunctionRevision(item) }

func validateVehicleFunctionChanges(changes []VehicleFunctionChange) error {
	return vehiclemodel.ValidateFunctionChanges(changes)
}

func applyVehicleFunctionChanges(ctx context.Context, tx *sql.Tx, vehicleID string,
	changes []VehicleFunctionChange, now string) error {
	if err := validateVehicleFunctionChanges(changes); err != nil {
		return err
	}
	for _, change := range changes {
		key := normalizeFunctionKey(change.FunctionKey)
		current, err := scanVehicleFunction(tx.QueryRowContext(ctx,
			vehicleFunctionSelect+" WHERE vehicle_id=? AND function_key=?", vehicleID, key))
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("read function revision: %w", err)
		}
		exists := err == nil
		if (exists && current.Revision != change.ExpectedRevision) ||
			(!exists && change.ExpectedRevision != "") {
			return ErrVehicleFunctionConflict
		}
		if change.Delete {
			if _, err := tx.ExecContext(ctx,
				"DELETE FROM vehicle_functions WHERE vehicle_id=? AND function_key=?", vehicleID, key); err != nil {
				return fmt.Errorf("delete draft function: %w", err)
			}
		} else if err := upsertVehicleFunction(ctx, tx, vehicleID, key,
			cleanVehicleFunctionInput(*change.Value), now); err != nil {
			return err
		}
	}
	return nil
}
