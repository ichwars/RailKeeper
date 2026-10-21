package application

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
)

var (
	ErrVehicleFunctionValidation = errors.New("invalid vehicle function changes")
	ErrVehicleFunctionConflict   = errors.New("vehicle function changed concurrently")
)

func vehicleFunctionRevision(item VehicleFunction) string {
	item.Revision = ""
	data, _ := json.Marshal(item)
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

func validateVehicleFunctionChanges(changes []VehicleFunctionChange) error {
	if len(changes) > 32 {
		return ErrVehicleFunctionValidation
	}
	seen := map[string]bool{}
	for _, change := range changes {
		key := normalizeFunctionKey(change.FunctionKey)
		if !validFunctionKey(key) || seen[key] || change.Delete == (change.Value != nil) {
			return ErrVehicleFunctionValidation
		}
		seen[key] = true
		if change.ExpectedRevision != "" {
			data, err := hex.DecodeString(change.ExpectedRevision)
			if err != nil || len(data) != sha256.Size {
				return ErrVehicleFunctionValidation
			}
		}
		if change.Delete && change.ExpectedRevision == "" {
			return ErrVehicleFunctionValidation
		}
		if change.Value != nil && (change.Value.FlagsVersion != 2 ||
			!isValidVehicleFunctionInput(cleanVehicleFunctionInput(*change.Value))) {
			return ErrVehicleFunctionValidation
		}
	}
	return nil
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
