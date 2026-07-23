package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

func (s *VehicleService) ListFunctions(ctx context.Context, vehicleID string) ([]VehicleFunction, error) {
	vehicleID = strings.TrimSpace(vehicleID)
	if vehicleID == "" {
		return nil, ErrVehicleNotFound
	}
	if _, err := s.Get(ctx, vehicleID); err != nil {
		return nil, err
	}
	return s.loadVehicleFunctions(ctx, vehicleID)
}

func (s *VehicleService) UpsertFunction(ctx context.Context, vehicleID, functionKey string, input VehicleFunctionInput) (*VehicleFunction, error) {
	vehicleID = strings.TrimSpace(vehicleID)
	functionKey = normalizeFunctionKey(functionKey)
	input = cleanVehicleFunctionInput(input)
	if vehicleID == "" || !validFunctionKey(functionKey) || !isValidVehicleFunctionInput(input) {
		return nil, ErrVehicleValidation
	}
	if _, err := s.Get(ctx, vehicleID); err != nil {
		return nil, err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	sortOrder := functionSortOrder(functionKey)
	id := vehicleID + ":" + functionKey
	if _, err := s.db.ExecContext(ctx, `
INSERT INTO vehicle_functions(id, vehicle_id, function_key, name, symbol_key, function_type, mode, direction_dependent, notes, sort_order, created_at, updated_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(vehicle_id, function_key) DO UPDATE SET
  name=excluded.name,
  symbol_key=excluded.symbol_key,
  function_type=excluded.function_type,
  mode=excluded.mode,
  direction_dependent=excluded.direction_dependent,
  notes=excluded.notes,
  sort_order=excluded.sort_order,
  updated_at=excluded.updated_at
`, id, vehicleID, functionKey, input.Name, input.SymbolKey, input.FunctionType, input.Mode, boolToInt(input.DirectionDependent), input.Notes, sortOrder, now, now); err != nil {
		return nil, fmt.Errorf("upsert vehicle function: %w", err)
	}
	return s.GetFunction(ctx, vehicleID, functionKey)
}

func (s *VehicleService) GetFunction(ctx context.Context, vehicleID, functionKey string) (*VehicleFunction, error) {
	var item VehicleFunction
	var directionDependent int
	err := s.db.QueryRowContext(ctx, `
SELECT id, vehicle_id, function_key, COALESCE(name, ''), COALESCE(symbol_key, ''), function_type, mode,
       direction_dependent, COALESCE(notes, ''), sort_order, created_at, updated_at
FROM vehicle_functions
WHERE vehicle_id=? AND function_key=?
`, strings.TrimSpace(vehicleID), normalizeFunctionKey(functionKey)).Scan(
		&item.ID,
		&item.VehicleID,
		&item.FunctionKey,
		&item.Name,
		&item.SymbolKey,
		&item.FunctionType,
		&item.Mode,
		&directionDependent,
		&item.Notes,
		&item.SortOrder,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrVehicleNotFound
		}
		return nil, fmt.Errorf("get vehicle function: %w", err)
	}
	item.DirectionDependent = directionDependent == 1
	return &item, nil
}

func (s *VehicleService) DeleteFunction(ctx context.Context, vehicleID, functionKey string) (*VehicleFunction, error) {
	function, err := s.GetFunction(ctx, vehicleID, functionKey)
	if err != nil {
		return nil, err
	}
	result, err := s.db.ExecContext(ctx, `DELETE FROM vehicle_functions WHERE vehicle_id=? AND function_key=?`, strings.TrimSpace(vehicleID), normalizeFunctionKey(functionKey))
	if err != nil {
		return nil, fmt.Errorf("delete vehicle function: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("read vehicle function delete result: %w", err)
	}
	if affected == 0 {
		return nil, ErrVehicleNotFound
	}
	return function, nil
}

func (s *VehicleService) loadVehicleFunctions(ctx context.Context, vehicleID string) ([]VehicleFunction, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, vehicle_id, function_key, COALESCE(name, ''), COALESCE(symbol_key, ''), function_type, mode,
       direction_dependent, COALESCE(notes, ''), sort_order, created_at, updated_at
FROM vehicle_functions
WHERE vehicle_id=?
ORDER BY sort_order ASC, function_key ASC
`, strings.TrimSpace(vehicleID))
	if err != nil {
		return nil, fmt.Errorf("list vehicle functions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := []VehicleFunction{}
	for rows.Next() {
		var item VehicleFunction
		var directionDependent int
		if err := rows.Scan(
			&item.ID,
			&item.VehicleID,
			&item.FunctionKey,
			&item.Name,
			&item.SymbolKey,
			&item.FunctionType,
			&item.Mode,
			&directionDependent,
			&item.Notes,
			&item.SortOrder,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan vehicle function: %w", err)
		}
		item.DirectionDependent = directionDependent == 1
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate vehicle functions: %w", err)
	}
	return out, nil
}
