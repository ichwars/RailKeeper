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

	if err := upsertVehicleFunction(ctx, s.db, vehicleID, functionKey, input, time.Now().UTC().Format(time.RFC3339)); err != nil {
		return nil, err
	}
	return s.GetFunction(ctx, vehicleID, functionKey)
}

type vehicleFunctionExecutor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

func upsertVehicleFunction(ctx context.Context, db vehicleFunctionExecutor, vehicleID, functionKey string,
	input VehicleFunctionInput, now string) error {
	// Requests from the old, mislabelled UI cannot establish either meaning.
	confirmedFlags := input.FlagsVersion == 2
	if !confirmedFlags {
		if input.DirectionDependent != nil {
			input.LegacyDirectionDependent = input.DirectionDependent
		}
		input.DirectionDependent, input.Inverted = nil, nil
	}
	sortOrder := functionSortOrder(functionKey)
	id := vehicleID + ":" + functionKey
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
`, id, vehicleID, functionKey, input.Name, input.SymbolKey, input.FunctionType, input.Mode, input.DirectionDependent, input.Inverted, input.LegacyDirectionDependent, input.Notes, sortOrder, now, now, confirmedFlags, confirmedFlags); err != nil {
		return fmt.Errorf("upsert vehicle function: %w", err)
	}
	return nil
}

func (s *VehicleService) GetFunction(ctx context.Context, vehicleID, functionKey string) (*VehicleFunction, error) {
	item, err := scanVehicleFunction(s.db.QueryRowContext(ctx,
		vehicleFunctionSelect+" WHERE vehicle_id=? AND function_key=?",
		strings.TrimSpace(vehicleID), normalizeFunctionKey(functionKey)))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrVehicleNotFound
		}
		return nil, fmt.Errorf("get vehicle function: %w", err)
	}
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
       direction_dependent, inverted, legacy_direction_dependent, COALESCE(notes, ''), sort_order, created_at, updated_at
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
		item, err := scanVehicleFunction(rows)
		if err != nil {
			return nil, fmt.Errorf("scan vehicle function: %w", err)
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate vehicle functions: %w", err)
	}
	return out, nil
}

const vehicleFunctionSelect = `
SELECT id, vehicle_id, function_key, COALESCE(name, ''), COALESCE(symbol_key, ''), function_type, mode,
       direction_dependent, inverted, legacy_direction_dependent, COALESCE(notes, ''), sort_order, created_at, updated_at
FROM vehicle_functions
`

type vehicleFunctionScanner interface{ Scan(...any) error }

func scanVehicleFunction(row vehicleFunctionScanner) (VehicleFunction, error) {
	var item VehicleFunction
	if err := row.Scan(
		&item.ID,
		&item.VehicleID,
		&item.FunctionKey,
		&item.Name,
		&item.SymbolKey,
		&item.FunctionType,
		&item.Mode,
		&item.DirectionDependent,
		&item.Inverted,
		&item.LegacyDirectionDependent,
		&item.Notes,
		&item.SortOrder,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return item, err
	}
	item.FlagsVersion = 2
	item.Revision = vehicleFunctionRevision(item)
	return item, nil
}
