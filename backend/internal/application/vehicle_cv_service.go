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
	tx, err := beginVehicleCVWrite(ctx, s.db, vehicleID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	item, err := writeVehicleCVTx(ctx, tx, vehicleID, nil, input, time.Now().UTC().Format(time.RFC3339Nano))
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit CV create: %w", err)
	}
	value := cvRecordValue(item)
	return &value, nil
}

func (s *VehicleService) UpdateCVValue(ctx context.Context, vehicleID, cvValueID string, input VehicleCVValueInput) (*VehicleCVValue, error) {
	vehicleID, cvValueID = strings.TrimSpace(vehicleID), strings.TrimSpace(cvValueID)
	input = cleanVehicleCVValueInput(input)
	if vehicleID == "" || cvValueID == "" || !isValidVehicleCVValueInput(input) {
		return nil, ErrVehicleValidation
	}
	tx, err := beginVehicleCVWrite(ctx, s.db, vehicleID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	previous, err := vehicledata.ReadCV(ctx, tx, vehicleID, cvValueID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrVehicleNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("read CV before update: %w", err)
	}
	if _, err := writeVehicleCVTx(ctx, tx, vehicleID, &previous, input, time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit CV update: %w", err)
	}
	return s.GetCVValue(ctx, vehicleID, cvValueID)
}

func (s *VehicleService) GetCVValue(ctx context.Context, vehicleID, cvValueID string) (*VehicleCVValue, error) {
	record, err := vehicledata.ScanCV(s.db.QueryRowContext(ctx, vehicledata.CVSelect+` WHERE id=? AND vehicle_id=?`,
		strings.TrimSpace(cvValueID), strings.TrimSpace(vehicleID)))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrVehicleNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get CV value: %w", err)
	}
	item := cvRecordValue(record)
	item.History, err = s.loadVehicleCVValueHistory(ctx, item.ID)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *VehicleService) DeleteCVValue(ctx context.Context, vehicleID, cvValueID string, expectedRevision ...string) (*VehicleCVValue, error) {
	if len(expectedRevision) > 1 {
		return nil, ErrVehicleValidation
	}
	vehicleID, cvValueID = strings.TrimSpace(vehicleID), strings.TrimSpace(cvValueID)
	tx, err := beginVehicleCVWrite(ctx, s.db, vehicleID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	record, err := vehicledata.ReadCV(ctx, tx, vehicleID, cvValueID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrVehicleNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("read CV before delete: %w", err)
	}
	if len(expectedRevision) == 1 {
		if err := checkCVRevision(&record, &expectedRevision[0]); err != nil {
			return nil, err
		}
	}
	item := cvRecordValue(record)
	item.History, err = loadVehicleCVHistory(ctx, tx, item.ID)
	if err != nil {
		return nil, err
	}
	if err := vehicledata.DeleteCV(ctx, tx, vehicleID, cvValueID); err != nil {
		return nil, fmt.Errorf("delete CV: %w", err)
	}
	if err := touchDecoderCVWorkspaceTx(ctx, tx, vehicleID, item.DecoderProfile, time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit CV delete: %w", err)
	}
	return &item, nil
}

func (s *VehicleService) loadVehicleCVValues(ctx context.Context, vehicleID string) ([]VehicleCVValue, error) {
	rows, err := s.db.QueryContext(ctx, vehicledata.CVSelect+`
WHERE vehicle_id=? ORDER BY decoder_profile ASC, index_cv31 ASC, index_cv32 ASC, cv_number ASC`, strings.TrimSpace(vehicleID))
	if err != nil {
		return nil, fmt.Errorf("list CV values: %w", err)
	}
	defer func() { _ = rows.Close() }()
	out := []VehicleCVValue{}
	for rows.Next() {
		item, err := vehicledata.ScanCV(rows)
		if err != nil {
			return nil, fmt.Errorf("scan CV value: %w", err)
		}
		out = append(out, cvRecordValue(item))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate CV values: %w", err)
	}
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("close CV values: %w", err)
	}
	if err := s.loadCVHistories(ctx, vehicleID, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *VehicleService) loadVehicleCVValueHistory(ctx context.Context, cvValueID string) ([]VehicleCVValueHistory, error) {
	return loadVehicleCVHistory(ctx, s.db, cvValueID)
}

func loadVehicleCVHistory(ctx context.Context, query interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}, cvValueID string) ([]VehicleCVValueHistory, error) {
	rows, err := query.QueryContext(ctx, `
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
	tx, err := beginVehicleCVWrite(ctx, s.db, vehicleID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	profile, err := resolveCVWorkspaceProfileTx(ctx, tx, vehicleID, input.WorkspaceID, input.ExpectedWorkspaceRevision, input.DecoderProfile)
	if err != nil {
		return nil, err
	}
	input.DecoderProfile = profile
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
	if _, err := tx.ExecContext(ctx, `
INSERT INTO vehicle_cv_files(id, vehicle_id, file_name, original_name, description, decoder_profile, mime_type, size_bytes, storage_path, blob_id, created_at, updated_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, file.ID, file.VehicleID, file.FileName, file.OriginalName, file.Description, file.DecoderProfile, file.MimeType, file.SizeBytes, file.StoragePath, file.BlobID, file.CreatedAt, file.UpdatedAt); err != nil {
		return nil, fmt.Errorf("create vehicle cv file: %w", err)
	}
	if err := touchDecoderCVWorkspaceTx(ctx, tx, vehicleID, file.DecoderProfile, now); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit CV file creation: %w", err)
	}
	file.Revision = cvFileRevision(file, "[]")
	return &file, nil
}

func (s *VehicleService) GetCVFile(ctx context.Context, vehicleID, fileID string) (*VehicleCVFile, error) {
	return readVehicleCVFile(ctx, s.db, vehicleID, fileID)
}

func (s *VehicleService) DeleteCVFile(ctx context.Context, vehicleID, fileID string, expectedRevision ...string) (*VehicleCVFile, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin CV file delete: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	file, err := readVehicleCVFile(ctx, tx, vehicleID, fileID)
	if err != nil {
		return nil, err
	}
	if err := checkVehicleRevision(file.Revision, optionalVehicleRevision(expectedRevision), ErrVehicleCVFileConflict); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM vehicle_cv_files WHERE id=? AND vehicle_id=?`, file.ID, file.VehicleID); err != nil {
		return nil, fmt.Errorf("delete vehicle CV file: %w", err)
	}
	if err := touchDecoderCVWorkspaceTx(ctx, tx, file.VehicleID, file.DecoderProfile, time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit CV file delete: %w", err)
	}
	return file, nil
}

func (s *VehicleService) loadVehicleCVFiles(ctx context.Context, vehicleID string) ([]VehicleCVFile, error) {
	rows, err := s.db.QueryContext(ctx, vehicleCVFileSelect+`WHERE vehicle_id=? ORDER BY created_at ASC`, strings.TrimSpace(vehicleID))
	if err != nil {
		return nil, fmt.Errorf("list vehicle cv files: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := []VehicleCVFile{}
	for rows.Next() {
		file, err := scanVehicleCVFile(rows)
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
