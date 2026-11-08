package application

import (
	"context"
	"database/sql"
	"strings"

	"railkeeper/backend/internal/infrastructure/vehicledata"
	vehiclemodel "railkeeper/backend/internal/inventory/vehicle"
)

func saveVehicleImages(ctx context.Context, tx *sql.Tx, vehicleID string, images []VehicleImageInput, now string) error {
	return vehiclemodel.ReplaceImages(ctx, vehicledata.ImageMutationFor(tx), vehicleID, images, now)
}

type queryRower interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func (s *VehicleService) nextInventoryNumber(ctx context.Context, tx *sql.Tx, vehicleCategory string) (string, error) {
	requestedCategory := cleanInventoryCategory(vehicleCategory)
	fallbackCategory := inventoryCategoryForVehicle(vehicleCategory)
	return ReserveInventoryNumber(ctx, tx, requestedCategory, fallbackCategory, func(candidate string) error {
		return s.ensureInventoryNumberAvailable(ctx, tx, candidate, "")
	})
}

func (s *VehicleService) ensureInventoryNumberAvailable(ctx context.Context, tx *sql.Tx, inventoryNumber, excludeVehicleID string) error {
	inventoryNumber = strings.TrimSpace(inventoryNumber)
	if inventoryNumber == "" {
		return ErrInventoryNumberValidation
	}
	used, err := vehicledata.InventoryNumberInUse(ctx, tx, inventoryNumber, excludeVehicleID)
	if err != nil {
		return err
	}
	if used {
		return ErrInventoryNumberConflict
	}
	return nil
}

func vehicleImagesFromInput(vehicleID string, images []VehicleImageInput, now string) []VehicleImage {
	return vehiclemodel.ImagesFromInput(vehicleID, images, now)
}
