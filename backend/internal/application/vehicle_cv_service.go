package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

func (s *VehicleService) ListCVValues(ctx context.Context, vehicleID string) ([]VehicleCVValue, error) {
	vehicleID = strings.TrimSpace(vehicleID)
	if vehicleID == "" {
		return nil, ErrVehicleNotFound
	}
	if _, err := s.Get(ctx, vehicleID); err != nil {
		return nil, err
	}
	return s.loadVehicleCVValues(ctx, vehicleID)
}

func (s *VehicleService) CreateCVValue(ctx context.Context, vehicleID string, input VehicleCVValueInput) (*VehicleCVValue, error) {
	vehicleID = strings.TrimSpace(vehicleID)
	input = cleanVehicleCVValueInput(input)
	if vehicleID == "" || !isValidVehicleCVValueInput(input) {
		return nil, ErrVehicleValidation
	}
	if _, err := s.Get(ctx, vehicleID); err != nil {
		return nil, err
	}
	if input.SourceFileID != "" {
		if _, err := s.GetCVFile(ctx, vehicleID, input.SourceFileID); err != nil {
			if errors.Is(err, ErrVehicleNotFound) {
				return nil, ErrVehicleValidation
			}
			return nil, err
		}
	}
	now := time.Now().UTC().Format(time.RFC3339)
	item := VehicleCVValue{
		ID:             randomID(),
		VehicleID:      vehicleID,
		CVNumber:       input.CVNumber,
		Value:          input.Value,
		Description:    input.Description,
		Category:       input.Category,
		Protocol:       input.Protocol,
		DecoderProfile: input.DecoderProfile,
		SourceFileID:   input.SourceFileID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if _, err := s.db.ExecContext(ctx, `
INSERT INTO vehicle_cv_values(id, vehicle_id, cv_number, value, description, category, protocol, decoder_profile, source_file_id, created_at, updated_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, item.ID, item.VehicleID, item.CVNumber, item.Value, item.Description, item.Category, item.Protocol, item.DecoderProfile, item.SourceFileID, item.CreatedAt, item.UpdatedAt); err != nil {
		return nil, fmt.Errorf("create vehicle cv value: %w", err)
	}
	return &item, nil
}

func (s *VehicleService) UpdateCVValue(ctx context.Context, vehicleID, cvValueID string, input VehicleCVValueInput) (*VehicleCVValue, error) {
	vehicleID = strings.TrimSpace(vehicleID)
	cvValueID = strings.TrimSpace(cvValueID)
	input = cleanVehicleCVValueInput(input)
	if vehicleID == "" || cvValueID == "" || !isValidVehicleCVValueInput(input) {
		return nil, ErrVehicleValidation
	}
	existing, err := s.GetCVValue(ctx, vehicleID, cvValueID)
	if err != nil {
		return nil, err
	}
	if input.SourceFileID != "" {
		if _, err := s.GetCVFile(ctx, vehicleID, input.SourceFileID); err != nil {
			if errors.Is(err, ErrVehicleNotFound) {
				return nil, ErrVehicleValidation
			}
			return nil, err
		}
	}
	now := time.Now().UTC().Format(time.RFC3339)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin cv value update: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	result, err := tx.ExecContext(ctx, `
UPDATE vehicle_cv_values
SET cv_number=?, value=?, description=?, category=?, protocol=?, decoder_profile=?, source_file_id=?, updated_at=?
WHERE id=? AND vehicle_id=?
`, input.CVNumber, input.Value, input.Description, input.Category, input.Protocol, input.DecoderProfile, input.SourceFileID, now, cvValueID, vehicleID)
	if err != nil {
		return nil, fmt.Errorf("update vehicle cv value: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("read vehicle cv value update result: %w", err)
	}
	if affected == 0 {
		return nil, ErrVehicleNotFound
	}
	if existing.Value != input.Value {
		if _, err = tx.ExecContext(ctx, `
INSERT INTO vehicle_cv_value_history(id, cv_value_id, vehicle_id, old_value, new_value, changed_at)
VALUES(?, ?, ?, ?, ?, ?)
`, randomID(), cvValueID, vehicleID, existing.Value, input.Value, now); err != nil {
			return nil, fmt.Errorf("write cv value history: %w", err)
		}
	}
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit cv value update: %w", err)
	}
	return s.GetCVValue(ctx, vehicleID, cvValueID)
}

func (s *VehicleService) GetCVValue(ctx context.Context, vehicleID, cvValueID string) (*VehicleCVValue, error) {
	var item VehicleCVValue
	err := s.db.QueryRowContext(ctx, `
SELECT id, vehicle_id, cv_number, value, COALESCE(description, ''), COALESCE(category, ''),
       COALESCE(protocol, ''), COALESCE(decoder_profile, ''), COALESCE(source_file_id, ''), created_at, updated_at
FROM vehicle_cv_values
WHERE id=? AND vehicle_id=?
`, strings.TrimSpace(cvValueID), strings.TrimSpace(vehicleID)).Scan(
		&item.ID,
		&item.VehicleID,
		&item.CVNumber,
		&item.Value,
		&item.Description,
		&item.Category,
		&item.Protocol,
		&item.DecoderProfile,
		&item.SourceFileID,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrVehicleNotFound
		}
		return nil, fmt.Errorf("get vehicle cv value: %w", err)
	}
	history, err := s.loadVehicleCVValueHistory(ctx, item.ID)
	if err != nil {
		return nil, err
	}
	item.History = history
	return &item, nil
}

func (s *VehicleService) DeleteCVValue(ctx context.Context, vehicleID, cvValueID string) (*VehicleCVValue, error) {
	item, err := s.GetCVValue(ctx, vehicleID, cvValueID)
	if err != nil {
		return nil, err
	}
	result, err := s.db.ExecContext(ctx, `DELETE FROM vehicle_cv_values WHERE id=? AND vehicle_id=?`, strings.TrimSpace(cvValueID), strings.TrimSpace(vehicleID))
	if err != nil {
		return nil, fmt.Errorf("delete vehicle cv value: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("read vehicle cv value delete result: %w", err)
	}
	if affected == 0 {
		return nil, ErrVehicleNotFound
	}
	return item, nil
}

func (s *VehicleService) loadVehicleCVValues(ctx context.Context, vehicleID string) ([]VehicleCVValue, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, vehicle_id, cv_number, value, COALESCE(description, ''), COALESCE(category, ''),
       COALESCE(protocol, ''), COALESCE(decoder_profile, ''), COALESCE(source_file_id, ''), created_at, updated_at
FROM vehicle_cv_values
WHERE vehicle_id=?
ORDER BY protocol ASC, decoder_profile ASC, cv_number ASC
`, strings.TrimSpace(vehicleID))
	if err != nil {
		return nil, fmt.Errorf("list vehicle cv values: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := []VehicleCVValue{}
	for rows.Next() {
		var item VehicleCVValue
		if err := rows.Scan(
			&item.ID,
			&item.VehicleID,
			&item.CVNumber,
			&item.Value,
			&item.Description,
			&item.Category,
			&item.Protocol,
			&item.DecoderProfile,
			&item.SourceFileID,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan vehicle cv value: %w", err)
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate vehicle cv values: %w", err)
	}
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("close vehicle cv values: %w", err)
	}
	for index := range out {
		history, err := s.loadVehicleCVValueHistory(ctx, out[index].ID)
		if err != nil {
			return nil, err
		}
		out[index].History = history
	}
	return out, nil
}

func (s *VehicleService) loadVehicleCVValueHistory(ctx context.Context, cvValueID string) ([]VehicleCVValueHistory, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, cv_value_id, vehicle_id, old_value, new_value, changed_at
FROM vehicle_cv_value_history
WHERE cv_value_id=?
ORDER BY changed_at DESC
`, strings.TrimSpace(cvValueID))
	if err != nil {
		return nil, fmt.Errorf("list vehicle cv value history: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := []VehicleCVValueHistory{}
	for rows.Next() {
		var item VehicleCVValueHistory
		if err := rows.Scan(
			&item.ID,
			&item.CVValueID,
			&item.VehicleID,
			&item.OldValue,
			&item.NewValue,
			&item.ChangedAt,
		); err != nil {
			return nil, fmt.Errorf("scan vehicle cv value history: %w", err)
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate vehicle cv value history: %w", err)
	}
	return out, nil
}

func (s *VehicleService) ListCVFiles(ctx context.Context, vehicleID string) ([]VehicleCVFile, error) {
	vehicleID = strings.TrimSpace(vehicleID)
	if vehicleID == "" {
		return nil, ErrVehicleNotFound
	}
	if _, err := s.Get(ctx, vehicleID); err != nil {
		return nil, err
	}
	return s.loadVehicleCVFiles(ctx, vehicleID)
}

func (s *VehicleService) CreateCVFile(ctx context.Context, vehicleID string, input VehicleCVFileInput) (*VehicleCVFile, error) {
	vehicleID = strings.TrimSpace(vehicleID)
	input = cleanVehicleCVFileInput(input)
	if vehicleID == "" || input.FileName == "" || input.OriginalName == "" || (input.StoragePath == "" && input.BlobID == "") {
		return nil, ErrVehicleValidation
	}
	if _, err := s.Get(ctx, vehicleID); err != nil {
		return nil, err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	file := VehicleCVFile{
		ID:             randomID(),
		VehicleID:      vehicleID,
		FileName:       input.FileName,
		OriginalName:   input.OriginalName,
		Description:    input.Description,
		DecoderProfile: input.DecoderProfile,
		MimeType:       input.MimeType,
		SizeBytes:      input.SizeBytes,
		StoragePath:    input.StoragePath,
		BlobID:         input.BlobID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if _, err := s.db.ExecContext(ctx, `
INSERT INTO vehicle_cv_files(id, vehicle_id, file_name, original_name, description, decoder_profile, mime_type, size_bytes, storage_path, blob_id, created_at, updated_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, file.ID, file.VehicleID, file.FileName, file.OriginalName, file.Description, file.DecoderProfile, file.MimeType, file.SizeBytes, file.StoragePath, file.BlobID, file.CreatedAt, file.UpdatedAt); err != nil {
		return nil, fmt.Errorf("create vehicle cv file: %w", err)
	}
	return &file, nil
}

func (s *VehicleService) GetCVFile(ctx context.Context, vehicleID, fileID string) (*VehicleCVFile, error) {
	var file VehicleCVFile
	err := s.db.QueryRowContext(ctx, `
SELECT id, vehicle_id, file_name, original_name, COALESCE(description, ''), COALESCE(decoder_profile, ''),
       COALESCE(mime_type, ''), size_bytes, storage_path, COALESCE(blob_id, ''), created_at, updated_at
FROM vehicle_cv_files
WHERE id=? AND vehicle_id=?
`, strings.TrimSpace(fileID), strings.TrimSpace(vehicleID)).Scan(
		&file.ID,
		&file.VehicleID,
		&file.FileName,
		&file.OriginalName,
		&file.Description,
		&file.DecoderProfile,
		&file.MimeType,
		&file.SizeBytes,
		&file.StoragePath,
		&file.BlobID,
		&file.CreatedAt,
		&file.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrVehicleNotFound
		}
		return nil, fmt.Errorf("get vehicle cv file: %w", err)
	}
	return &file, nil
}

func (s *VehicleService) DeleteCVFile(ctx context.Context, vehicleID, fileID string) (*VehicleCVFile, error) {
	file, err := s.GetCVFile(ctx, vehicleID, fileID)
	if err != nil {
		return nil, err
	}
	result, err := s.db.ExecContext(ctx, `DELETE FROM vehicle_cv_files WHERE id=? AND vehicle_id=?`, strings.TrimSpace(fileID), strings.TrimSpace(vehicleID))
	if err != nil {
		return nil, fmt.Errorf("delete vehicle cv file: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("read vehicle cv file delete result: %w", err)
	}
	if affected == 0 {
		return nil, ErrVehicleNotFound
	}
	return file, nil
}

func (s *VehicleService) loadVehicleCVFiles(ctx context.Context, vehicleID string) ([]VehicleCVFile, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, vehicle_id, file_name, original_name, COALESCE(description, ''), COALESCE(decoder_profile, ''),
       COALESCE(mime_type, ''), size_bytes, storage_path, COALESCE(blob_id, ''), created_at, updated_at
FROM vehicle_cv_files
WHERE vehicle_id=?
ORDER BY created_at ASC
`, strings.TrimSpace(vehicleID))
	if err != nil {
		return nil, fmt.Errorf("list vehicle cv files: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := []VehicleCVFile{}
	for rows.Next() {
		var file VehicleCVFile
		if err := rows.Scan(
			&file.ID,
			&file.VehicleID,
			&file.FileName,
			&file.OriginalName,
			&file.Description,
			&file.DecoderProfile,
			&file.MimeType,
			&file.SizeBytes,
			&file.StoragePath,
			&file.BlobID,
			&file.CreatedAt,
			&file.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan vehicle cv file: %w", err)
		}
		out = append(out, file)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate vehicle cv files: %w", err)
	}
	return out, nil
}
