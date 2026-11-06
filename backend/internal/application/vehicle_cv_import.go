package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"railkeeper/backend/internal/infrastructure/vehicledata"
)

const MaxVehicleCVImportValues = 10000

type VehicleCVImportInput struct {
	Values []VehicleCVValueInput `json:"values"`
}

// ImportCVValues applies the reviewed selection as one transaction, including value history.
func (s *VehicleService) ImportCVValues(ctx context.Context, vehicleID string, input VehicleCVImportInput) (int, error) {
	vehicleID = strings.TrimSpace(vehicleID)
	if len(input.Values) == 0 || len(input.Values) > MaxVehicleCVImportValues {
		return 0, ErrVehicleValidation
	}
	seen := map[string]bool{}
	for i, value := range input.Values {
		value = cleanVehicleCVValueInput(value)
		key := vehicleCVIdentity(value)
		if !isValidVehicleCVValueInput(value) || seen[key] {
			return 0, ErrVehicleValidation
		}
		seen[key], input.Values[i] = true, value
	}
	tx, err := beginVehicleCVWrite(ctx, s.db, vehicleID)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	for _, value := range input.Values {
		if err := importCVValueTx(ctx, tx, vehicleID, value, now); err != nil {
			return 0, err
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit cv import: %w", err)
	}
	return len(input.Values), nil
}

func cvIndexKey(value *int) int {
	if value == nil {
		return -1
	}
	return *value
}

func importCVValueTx(ctx context.Context, tx *sql.Tx, vehicleID string, value VehicleCVValueInput, now string) error {
	previous, err := vehicledata.FindCV(ctx, tx, cvRecordInput(vehicleID, value))
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("read imported CV: %w", err)
	}
	var existing *vehicledata.CVRecord
	if err == nil {
		existing = &previous
	}
	_, err = writeVehicleCVTx(ctx, tx, vehicleID, existing, value, now)
	return err
}

// A single history query keeps large imported CV sets usable in the vehicle detail view.
func (s *VehicleService) loadCVHistories(ctx context.Context, vehicleID string, values []VehicleCVValue) error {
	rows, err := s.db.QueryContext(ctx, `SELECT id, cv_value_id, vehicle_id, old_value, new_value, changed_at
FROM vehicle_cv_value_history WHERE vehicle_id=? ORDER BY changed_at DESC, id DESC`, strings.TrimSpace(vehicleID))
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()
	positions := make(map[string]int, len(values))
	for i := range values {
		positions[values[i].ID] = i
		values[i].History = []VehicleCVValueHistory{}
	}
	for rows.Next() {
		var h VehicleCVValueHistory
		if err := rows.Scan(&h.ID, &h.CVValueID, &h.VehicleID, &h.OldValue, &h.NewValue, &h.ChangedAt); err != nil {
			return err
		}
		if i, ok := positions[h.CVValueID]; ok {
			values[i].History = append(values[i].History, h)
		}
	}
	return rows.Err()
}
