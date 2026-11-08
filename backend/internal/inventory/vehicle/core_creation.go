package vehicle

import (
	"context"
	"railkeeper/backend/internal/credentials"
	"railkeeper/backend/internal/inventory/numbering"
	"strings"
)

func InsertVehicle(
	ctx context.Context,
	tx VehicleMutation,
	vehicleID string,
	input CreateVehicleInput,
	actorUserID string,
	now string,
) (*Vehicle, error) {
	var err error
	if input.InventoryNumber == "" {
		input.InventoryNumber, err = ReserveVehicleNumber(ctx, tx, input.Category)
		if err != nil {
			return nil, err
		}
	} else if err = EnsureInventoryNumberAvailable(ctx, tx, input.InventoryNumber, ""); err != nil {
		return nil, err
	}

	vehicle := FromCreateInput(vehicleID, input, now)
	if err = tx.InsertVehicle(ctx, WriteRecord(vehicle)); err != nil {
		return nil, err
	}

	if err = tx.AuditCreated(ctx, VehicleAudit{ID: credentials.RandomID(), VehicleID: vehicle.ID, Actor: actorUserID, At: now}); err != nil {
		return nil, err
	}
	if err = ReplaceImages(ctx, tx, vehicle.ID, input.Images, now); err != nil {
		return nil, err
	}
	if err = ApplyFunctionChanges(ctx, tx, vehicle.ID, input.FunctionChanges, now); err != nil {
		return nil, err
	}
	return &vehicle, nil
}

func ReserveVehicleNumber(ctx context.Context, tx VehicleMutation, category string) (string, error) {
	return numbering.Reserve(ctx, tx.Counter(), numbering.CleanCategory(category), numbering.CategoryForVehicle(category), func(candidate string) error { return EnsureInventoryNumberAvailable(ctx, tx, candidate, "") })
}
func EnsureInventoryNumberAvailable(ctx context.Context, tx VehicleMutation, number, excluded string) error {
	number = strings.TrimSpace(number)
	if number == "" {
		return numbering.ErrValidation
	}
	used, err := tx.InventoryNumberInUse(ctx, number, excluded)
	if err != nil {
		return err
	}
	if used {
		return numbering.ErrConflict
	}
	return nil
}
