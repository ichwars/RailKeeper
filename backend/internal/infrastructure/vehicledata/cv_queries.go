package vehicledata

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"railkeeper/backend/internal/decoder"
	"railkeeper/backend/internal/decoder/workflow"
)

func (q *cvQueries) CVValue(ctx context.Context, vehicleID, id string) (decoder.CVRecord, bool, error) {
	record, err := ScanCV(q.query.QueryRowContext(ctx, CVSelect+` WHERE id=? AND vehicle_id=?`, id, vehicleID))
	if errors.Is(err, sql.ErrNoRows) {
		return record, false, nil
	}
	return record, err == nil, err
}
func (q *cvQueries) CVValues(ctx context.Context, vehicleID string) ([]decoder.CVRecord, error) {
	rows, err := q.query.QueryContext(ctx, CVSelect+`
WHERE vehicle_id=? ORDER BY decoder_profile ASC, index_cv31 ASC, index_cv32 ASC, cv_number ASC`, strings.TrimSpace(vehicleID))
	if err != nil {
		return nil, fmt.Errorf("list CV values: %w", err)
	}
	defer func() { _ = rows.Close() }()
	out := []decoder.CVRecord{}
	for rows.Next() {
		item, err := ScanCV(rows)
		if err != nil {
			return nil, fmt.Errorf("scan CV value: %w", err)
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate CV values: %w", err)
	}
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("close CV values: %w", err)
	}
	return out, nil
}
func (q *cvQueries) CVHistory(ctx context.Context, id string) ([]decoder.VehicleCVValueHistory, error) {
	return ReadCVHistory(ctx, q.query, id)
}
func (q *cvQueries) CVHistories(ctx context.Context, vehicleID string) ([]decoder.VehicleCVValueHistory, error) {
	rows, err := q.query.QueryContext(ctx, `SELECT id, cv_value_id, vehicle_id, old_value, new_value, changed_at
FROM vehicle_cv_value_history WHERE vehicle_id=? ORDER BY changed_at DESC, id DESC`, strings.TrimSpace(vehicleID))
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := []decoder.VehicleCVValueHistory{}
	for rows.Next() {
		var h decoder.VehicleCVValueHistory
		if err := rows.Scan(&h.ID, &h.CVValueID, &h.VehicleID, &h.OldValue, &h.NewValue, &h.ChangedAt); err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, rows.Err()
}
func (q *cvQueries) CVFile(ctx context.Context, vehicleID, id string) (*decoder.VehicleCVFile, error) {
	file, err := ScanCVFile(q.query.QueryRowContext(ctx, CVFileSelect+`WHERE id=? AND vehicle_id=?`,
		strings.TrimSpace(id), strings.TrimSpace(vehicleID)))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, workflow.ErrCVNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get vehicle CV file: %w", err)
	}
	return file, nil
}
func (q *cvQueries) CVFiles(ctx context.Context, vehicleID string) ([]decoder.VehicleCVFile, error) {
	rows, err := q.query.QueryContext(ctx, CVFileSelect+`WHERE vehicle_id=? ORDER BY created_at ASC`, strings.TrimSpace(vehicleID))
	if err != nil {
		return nil, fmt.Errorf("list vehicle cv files: %w", err)
	}
	defer func() { _ = rows.Close() }()
	out := []decoder.VehicleCVFile{}
	for rows.Next() {
		file, err := ScanCVFile(rows)
		if err != nil {
			return nil, fmt.Errorf("scan vehicle CV file: %w", err)
		}
		out = append(out, *file)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate vehicle cv files: %w", err)
	}
	return out, nil
}
