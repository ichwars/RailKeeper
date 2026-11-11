package vehicledata

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	vehiclemodel "railkeeper/backend/internal/inventory/vehicle"
	"strings"
)

type functionStore struct{ db *sql.DB }
type functionMutation struct{ tx *sql.Tx }

func NewFunctionStore(db *sql.DB) vehiclemodel.FunctionRepository { return &functionStore{db: db} }

// FunctionMutationFor participates in the caller's vehicle transaction without committing it.
func FunctionMutationFor(tx *sql.Tx) vehiclemodel.FunctionMutation { return &functionMutation{tx: tx} }
func (s *functionStore) FindFunction(ctx context.Context, id, key string) (vehiclemodel.VehicleFunction, bool, error) {
	return findFunction(ctx, s.db, id, key)
}
func (m *functionMutation) FindFunction(ctx context.Context, id, key string) (vehiclemodel.VehicleFunction, bool, error) {
	return findFunction(ctx, m.tx, id, key)
}
func findFunction(ctx context.Context, query queryRower, id, key string) (vehiclemodel.VehicleFunction, bool, error) {
	item, err := scanFunction(query.QueryRowContext(ctx, functionSelect+" WHERE vehicle_id=? AND function_key=?", id, key))
	if errors.Is(err, sql.ErrNoRows) {
		return item, false, nil
	}
	return item, err == nil, err
}
func (s *functionStore) Functions(ctx context.Context, vehicleID string) ([]vehiclemodel.VehicleFunction, error) {
	return readFunctions(ctx, s.db, "vehicle_id=?", strings.TrimSpace(vehicleID))
}

func ReadFunctionsForVehicles(ctx context.Context, query vehicleMediaQuery, vehicleIDsJSON string) ([]vehiclemodel.VehicleFunction, error) {
	return readFunctions(ctx, query, "vehicle_id IN (SELECT value FROM json_each(?))", vehicleIDsJSON)
}

func readFunctions(ctx context.Context, query vehicleMediaQuery, filter string, arg string) ([]vehiclemodel.VehicleFunction, error) {
	rows, err := query.QueryContext(ctx, functionSelect+`WHERE `+filter+`
ORDER BY sort_order ASC, function_key ASC
`, arg)
	if err != nil {
		return nil, fmt.Errorf("list vehicle functions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := []vehiclemodel.VehicleFunction{}
	for rows.Next() {
		item, err := scanFunction(rows)
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

const functionSelect = `
SELECT id, vehicle_id, function_key, COALESCE(name, ''), COALESCE(symbol_key, ''), function_type, mode,
       direction_dependent, inverted, legacy_direction_dependent, COALESCE(notes, ''), sort_order, created_at, updated_at
FROM vehicle_functions
`

type functionScanner interface{ Scan(...any) error }

func scanFunction(row functionScanner) (vehiclemodel.VehicleFunction, error) {
	var item vehiclemodel.VehicleFunction
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
	item.Revision = vehiclemodel.FunctionRevision(item)
	return item, nil
}
