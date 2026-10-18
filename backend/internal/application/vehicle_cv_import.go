package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

const MaxVehicleCVImportValues = 10000

type VehicleCVImportInput struct {
	Values []VehicleCVValueInput `json:"values"`
}

// ImportCVValues applies the reviewed selection as one transaction, including value history.
func (s *VehicleService) ImportCVValues(ctx context.Context, vehicleID string, input VehicleCVImportInput) (int, error) {
	if len(input.Values) == 0 || len(input.Values) > MaxVehicleCVImportValues {
		return 0, ErrVehicleValidation
	}
	seen := map[string]bool{}
	for i, value := range input.Values {
		value = cleanVehicleCVValueInput(value)
		key := fmt.Sprintf("%s:%d:%v:%v", value.DecoderProfile, value.CVNumber,
			cvIndexKey(value.IndexCV31), cvIndexKey(value.IndexCV32))
		if !isValidVehicleCVValueInput(value) || seen[key] {
			return 0, ErrVehicleValidation
		}
		seen[key] = true
		input.Values[i] = value
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin cv import: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	var exists int
	if err := tx.QueryRowContext(ctx, `SELECT 1 FROM vehicles WHERE id=?`, vehicleID).Scan(&exists); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrVehicleNotFound
		}
		return 0, err
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	checkedFiles := map[string]bool{}
	for _, value := range input.Values {
		if value.SourceFileID != "" && !checkedFiles[value.SourceFileID] {
			if err := tx.QueryRowContext(ctx, `SELECT 1 FROM vehicle_cv_files WHERE id=? AND vehicle_id=?`,
				value.SourceFileID, vehicleID).Scan(&exists); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return 0, ErrVehicleValidation
				}
				return 0, err
			}
			checkedFiles[value.SourceFileID] = true
		}
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
	var id string
	var oldValue int
	err := tx.QueryRowContext(ctx, `SELECT id, value FROM vehicle_cv_values
WHERE vehicle_id=? AND cv_number=? AND decoder_profile=? AND index_cv31 IS ? AND index_cv32 IS ?`,
		vehicleID, value.CVNumber, value.DecoderProfile, value.IndexCV31, value.IndexCV32).Scan(&id, &oldValue)
	if errors.Is(err, sql.ErrNoRows) {
		_, err = tx.ExecContext(ctx, `INSERT INTO vehicle_cv_values
(id, vehicle_id, cv_number, value, description, category, protocol, decoder_profile, source_file_id,
 created_at, updated_at, index_cv31, index_cv32, source_line, source_exported_at) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			randomID(), vehicleID, value.CVNumber, value.Value, value.Description, value.Category,
			value.Protocol, value.DecoderProfile, value.SourceFileID, now, now, value.IndexCV31, value.IndexCV32, value.SourceLine, value.SourceExportedAt)
		return err
	}
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `UPDATE vehicle_cv_values
SET value=?, description=?, category=?, protocol=?, source_file_id=?, updated_at=?,
    source_line=COALESCE(?, source_line), source_exported_at=CASE WHEN ?='' THEN source_exported_at ELSE ? END WHERE id=? AND vehicle_id=?`,
		value.Value, value.Description, value.Category, value.Protocol, value.SourceFileID, now, value.SourceLine, value.SourceExportedAt, value.SourceExportedAt, id, vehicleID)
	if err != nil || oldValue == value.Value {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO vehicle_cv_value_history
(id, cv_value_id, vehicle_id, old_value, new_value, changed_at) VALUES(?, ?, ?, ?, ?, ?)`,
		randomID(), id, vehicleID, oldValue, value.Value, now)
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
