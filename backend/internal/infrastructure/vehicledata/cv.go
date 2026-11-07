package vehicledata

import (
	"context"
	"database/sql"
	"fmt"
)

// CVRecord contains only persisted fields, including nullable index/source identities.
type CVRecord struct {
	ID, VehicleID                   string
	CVNumber                        int
	IndexCV31, IndexCV32            *int
	Value                           int
	Description, Category, Protocol string
	DecoderProfile, SourceFileID    string
	SourceLine                      *int
	SourceExportedAt                string
	CreatedAt, UpdatedAt            string
}

const CVSelect = `SELECT id, vehicle_id, cv_number, value, COALESCE(description, ''), COALESCE(category, ''),
COALESCE(protocol, ''), COALESCE(decoder_profile, ''), COALESCE(source_file_id, ''), created_at, updated_at,
index_cv31, index_cv32, source_line, source_exported_at FROM vehicle_cv_values`

func ScanCV(row interface{ Scan(...any) error }) (CVRecord, error) {
	var item CVRecord
	err := row.Scan(&item.ID, &item.VehicleID, &item.CVNumber, &item.Value, &item.Description,
		&item.Category, &item.Protocol, &item.DecoderProfile, &item.SourceFileID, &item.CreatedAt,
		&item.UpdatedAt, &item.IndexCV31, &item.IndexCV32, &item.SourceLine, &item.SourceExportedAt)
	return item, err
}

func ReadCV(ctx context.Context, tx *sql.Tx, vehicleID, id string) (CVRecord, error) {
	return ScanCV(tx.QueryRowContext(ctx, CVSelect+` WHERE id=? AND vehicle_id=?`, id, vehicleID))
}

func FindCV(ctx context.Context, tx *sql.Tx, item CVRecord) (CVRecord, error) {
	return ScanCV(tx.QueryRowContext(ctx, CVSelect+`
WHERE vehicle_id=? AND cv_number=? AND decoder_profile=? AND index_cv31 IS ? AND index_cv32 IS ?`,
		item.VehicleID, item.CVNumber, item.DecoderProfile, item.IndexCV31, item.IndexCV32))
}

func InsertCV(ctx context.Context, tx *sql.Tx, item CVRecord) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO vehicle_cv_values
(id, vehicle_id, cv_number, value, description, category, protocol, decoder_profile, source_file_id,
created_at, updated_at, index_cv31, index_cv32, source_line, source_exported_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, item.ID, item.VehicleID, item.CVNumber,
		item.Value, item.Description, item.Category, item.Protocol, item.DecoderProfile, item.SourceFileID,
		item.CreatedAt, item.UpdatedAt, item.IndexCV31, item.IndexCV32, item.SourceLine, item.SourceExportedAt)
	return err
}

func UpdateCV(ctx context.Context, tx *sql.Tx, item CVRecord) error {
	result, err := tx.ExecContext(ctx, `UPDATE vehicle_cv_values SET cv_number=?, value=?, description=?,
category=?, protocol=?, decoder_profile=?, source_file_id=?, updated_at=?, index_cv31=?, index_cv32=?,
source_line=?, source_exported_at=? WHERE id=? AND vehicle_id=?`, item.CVNumber, item.Value,
		item.Description, item.Category, item.Protocol, item.DecoderProfile, item.SourceFileID, item.UpdatedAt,
		item.IndexCV31, item.IndexCV32, item.SourceLine, item.SourceExportedAt, item.ID, item.VehicleID)
	return cvAffected(result, err)
}

func DeleteCV(ctx context.Context, tx *sql.Tx, vehicleID, id string) error {
	result, err := tx.ExecContext(ctx, `DELETE FROM vehicle_cv_values WHERE id=? AND vehicle_id=?`, id, vehicleID)
	return cvAffected(result, err)
}

func InsertCVHistory(ctx context.Context, tx *sql.Tx, id, vehicleID, cvID string, oldValue, newValue int, now string) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO vehicle_cv_value_history
(id, cv_value_id, vehicle_id, old_value, new_value, changed_at) VALUES(?, ?, ?, ?, ?, ?)`,
		id, cvID, vehicleID, oldValue, newValue, now)
	return err
}

func cvAffected(result sql.Result, err error) error {
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read cv write result: %w", err)
	}
	if count != 1 {
		return sql.ErrNoRows
	}
	return nil
}
