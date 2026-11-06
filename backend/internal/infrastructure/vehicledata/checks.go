package vehicledata

import (
	"context"
	"database/sql"
	"fmt"
)

type queryRower interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func MaintenanceBelongsToVehicle(ctx context.Context, query queryRower, vehicleID, maintenanceID string) (bool, error) {
	var count int
	if err := query.QueryRowContext(ctx, `SELECT COUNT(*) FROM vehicle_maintenance WHERE id=? AND vehicle_id=?`, maintenanceID, vehicleID).Scan(&count); err != nil {
		return false, fmt.Errorf("validate vehicle maintenance link: %w", err)
	}
	return count != 0, nil
}

func InventoryNumberInUse(ctx context.Context, tx *sql.Tx, inventoryNumber, excludeVehicleID string) (bool, error) {
	var count int
	if err := tx.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM vehicles
WHERE inventory_number=? AND (? = '' OR id <> ?)
`, inventoryNumber, excludeVehicleID, excludeVehicleID).Scan(&count); err != nil {
		return false, fmt.Errorf("check inventory number availability: %w", err)
	}
	return count > 0, nil
}
