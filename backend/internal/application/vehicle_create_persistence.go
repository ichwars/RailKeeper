package application

import (
	"context"
	"database/sql"
	"fmt"

	"railkeeper/backend/internal/infrastructure/vehicledata"
	vehiclemodel "railkeeper/backend/internal/inventory/vehicle"
)

func validateCreateVehicleInput(input CreateVehicleInput) (CreateVehicleInput, error) {
	return vehiclemodel.ValidateCreateInput(input)
}

func (s *VehicleService) insertVehicleTx(
	ctx context.Context,
	tx *sql.Tx,
	vehicleID string,
	input CreateVehicleInput,
	actorUserID string,
	now string,
) (*Vehicle, error) {
	var err error
	if input.InventoryNumber == "" {
		input.InventoryNumber, err = s.nextInventoryNumber(ctx, tx, input.Category)
		if err != nil {
			return nil, err
		}
	} else if err = s.ensureInventoryNumberAvailable(ctx, tx, input.InventoryNumber, ""); err != nil {
		return nil, err
	}

	vehicle := vehicleFromCreateInput(vehicleID, input, now)
	if err = vehicledata.Insert(ctx, tx, vehicleWriteRecord(vehicle)); err != nil {
		return nil, err
	}

	if _, err = tx.ExecContext(ctx, `
INSERT INTO audit_logs(id, actor_user_id, action, target_type, target_id, created_at, details_json)
VALUES(?, ?, 'VehicleCreated', 'vehicle', ?, ?, '{}')
`, randomID(), actorUserID, vehicle.ID, now); err != nil {
		return nil, fmt.Errorf("write vehicle audit log: %w", err)
	}
	if err = saveVehicleImages(ctx, tx, vehicle.ID, input.Images, now); err != nil {
		return nil, err
	}
	if err = applyVehicleFunctionChanges(ctx, tx, vehicle.ID, input.FunctionChanges, now); err != nil {
		return nil, err
	}
	return &vehicle, nil
}

func vehicleFromCreateInput(vehicleID string, input CreateVehicleInput, now string) Vehicle {
	return vehiclemodel.FromCreateInput(vehicleID, input, now)
}
