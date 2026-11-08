package vehicledata

import (
	"context"
	"database/sql"
	"fmt"
	vehiclemodel "railkeeper/backend/internal/inventory/vehicle"
)

type functionExecutor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

func (s *functionStore) StoreFunction(ctx context.Context, write vehiclemodel.FunctionWrite) error {
	return storeFunction(ctx, s.db, write)
}
func (m *functionMutation) StoreFunction(ctx context.Context, write vehiclemodel.FunctionWrite) error {
	return storeFunction(ctx, m.tx, write)
}
func storeFunction(ctx context.Context, db functionExecutor, write vehiclemodel.FunctionWrite) error {
	item := write.Function
	if _, err := db.ExecContext(ctx, `
INSERT INTO vehicle_functions(id, vehicle_id, function_key, name, symbol_key, function_type, mode, direction_dependent, inverted, legacy_direction_dependent, notes, sort_order, created_at, updated_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(vehicle_id, function_key) DO UPDATE SET
  name=excluded.name,
  symbol_key=excluded.symbol_key,
  function_type=excluded.function_type,
  mode=excluded.mode,
  direction_dependent=CASE WHEN ? THEN excluded.direction_dependent ELSE vehicle_functions.direction_dependent END,
  inverted=CASE WHEN ? THEN excluded.inverted ELSE vehicle_functions.inverted END,
  legacy_direction_dependent=COALESCE(vehicle_functions.legacy_direction_dependent, excluded.legacy_direction_dependent),
  notes=excluded.notes,
  sort_order=excluded.sort_order,
  updated_at=excluded.updated_at
`, item.ID, item.VehicleID, item.FunctionKey, item.Name, item.SymbolKey, item.FunctionType, item.Mode, item.DirectionDependent, item.Inverted, item.LegacyDirectionDependent, item.Notes, item.SortOrder, item.CreatedAt, item.UpdatedAt, write.ConfirmedFlags, write.ConfirmedFlags); err != nil {
		return fmt.Errorf("upsert vehicle function: %w", err)
	}
	return nil
}

func (s *functionStore) DeleteFunction(ctx context.Context, vehicleID, functionKey string) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM vehicle_functions WHERE vehicle_id=? AND function_key=?`, vehicleID, functionKey)
	if err != nil {
		return fmt.Errorf("delete vehicle function: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read vehicle function delete result: %w", err)
	}
	if affected == 0 {
		return vehiclemodel.ErrVehicleNotFound
	}
	return nil
}

func (m *functionMutation) DeleteDraftFunction(ctx context.Context, vehicleID, key string) error {
	if _, err := m.tx.ExecContext(ctx, "DELETE FROM vehicle_functions WHERE vehicle_id=? AND function_key=?", vehicleID, key); err != nil {
		return fmt.Errorf("delete draft function: %w", err)
	}
	return nil
}
