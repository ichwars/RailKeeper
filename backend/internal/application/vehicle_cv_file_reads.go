package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const vehicleCVFileSelect = `
SELECT id, vehicle_id, file_name, original_name, COALESCE(description, ''), COALESCE(decoder_profile, ''),
       COALESCE(mime_type, ''), size_bytes, storage_path, COALESCE(blob_id, ''), created_at, updated_at,
       (SELECT json_group_array(id) FROM (
          SELECT id FROM vehicle_cv_values WHERE source_file_id=vehicle_cv_files.id ORDER BY id
       ))
FROM vehicle_cv_files
`

func cvFileRevision(file VehicleCVFile, sourceReferences string) string {
	file.Revision = ""
	return recordRevision(struct {
		File                                  VehicleCVFile
		StoragePath, BlobID, SourceReferences string
	}{file, file.StoragePath, file.BlobID, sourceReferences})
}

func scanVehicleCVFile(row interface{ Scan(...any) error }) (*VehicleCVFile, error) {
	var file VehicleCVFile
	var references string
	if err := row.Scan(&file.ID, &file.VehicleID, &file.FileName, &file.OriginalName,
		&file.Description, &file.DecoderProfile, &file.MimeType, &file.SizeBytes,
		&file.StoragePath, &file.BlobID, &file.CreatedAt, &file.UpdatedAt, &references); err != nil {
		return nil, err
	}
	var sourceIDs []string
	if err := json.Unmarshal([]byte(references), &sourceIDs); err != nil {
		return nil, fmt.Errorf("read CV file source references: %w", err)
	}
	file.SourceReferenceCount = len(sourceIDs)
	file.Revision = cvFileRevision(file, references)
	return &file, nil
}

func readVehicleCVFile(ctx context.Context, query queryRower, vehicleID, fileID string) (*VehicleCVFile, error) {
	file, err := scanVehicleCVFile(query.QueryRowContext(ctx, vehicleCVFileSelect+`WHERE id=? AND vehicle_id=?`,
		strings.TrimSpace(fileID), strings.TrimSpace(vehicleID)))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrVehicleNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get vehicle CV file: %w", err)
	}
	return file, nil
}
