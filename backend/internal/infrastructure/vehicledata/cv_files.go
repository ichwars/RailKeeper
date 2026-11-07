package vehicledata

import (
	"encoding/json"
	"fmt"

	"railkeeper/backend/internal/decoder"
)

const CVFileSelect = `
SELECT id, vehicle_id, file_name, original_name, COALESCE(description, ''), COALESCE(decoder_profile, ''),
       COALESCE(mime_type, ''), size_bytes, storage_path, COALESCE(blob_id, ''), created_at, updated_at,
       (SELECT json_group_array(id) FROM (
          SELECT id FROM vehicle_cv_values WHERE source_file_id=vehicle_cv_files.id ORDER BY id
       ))
FROM vehicle_cv_files
`

func ScanCVFile(row interface{ Scan(...any) error }) (*decoder.VehicleCVFile, error) {
	var file decoder.VehicleCVFile
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
	file.Revision = decoder.CVFileRevision(file, references)
	return &file, nil
}
