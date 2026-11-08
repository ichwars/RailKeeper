package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"railkeeper/backend/internal/infrastructure/vehicledata"
	vehiclemodel "railkeeper/backend/internal/inventory/vehicle"
)

func (s *VehicleService) attachImages(ctx context.Context, vehicles []Vehicle) error {
	for index := range vehicles {
		images, err := s.loadVehicleImages(ctx, vehicles[index].ID)
		if err != nil {
			return err
		}
		vehicles[index].Images = images
		vehicles[index].ImagesRevision = vehicleImagesRevision(images)
	}
	return nil
}

func (s *VehicleService) loadVehicleImages(ctx context.Context, vehicleID string) ([]VehicleImage, error) {
	return readVehicleImages(ctx, s.db, vehicleID)
}

type vehicleMediaQuery interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func readVehicleImages(ctx context.Context, query vehicleMediaQuery, vehicleID string) ([]VehicleImage, error) {
	return vehicledata.ReadImages(ctx, query, vehicleID)
}

func (s *VehicleService) attachAttachments(ctx context.Context, vehicles []Vehicle) error {
	for index := range vehicles {
		attachments, err := s.loadVehicleAttachments(ctx, vehicles[index].ID)
		if err != nil {
			return err
		}
		vehicles[index].Attachments = attachments
	}
	return nil
}

func (s *VehicleService) attachMaintenance(ctx context.Context, vehicles []Vehicle) error {
	for index := range vehicles {
		maintenance, err := s.loadVehicleMaintenance(ctx, vehicles[index].ID)
		if err != nil {
			return err
		}
		vehicles[index].Maintenance = maintenance
	}
	return nil
}

func (s *VehicleService) attachSpareParts(ctx context.Context, vehicles []Vehicle) error {
	for index := range vehicles {
		spareParts, err := s.loadVehicleSpareParts(ctx, vehicles[index].ID)
		if err != nil {
			return err
		}
		vehicles[index].SpareParts = spareParts
	}
	return nil
}

func (s *VehicleService) attachFunctions(ctx context.Context, vehicles []Vehicle) error {
	for index := range vehicles {
		functions, err := s.loadVehicleFunctions(ctx, vehicles[index].ID)
		if err != nil {
			return err
		}
		vehicles[index].Functions = functions
	}
	return nil
}

func (s *VehicleService) attachCVData(ctx context.Context, vehicles []Vehicle) error {
	for index := range vehicles {
		values, err := s.loadVehicleCVValues(ctx, vehicles[index].ID)
		if err != nil {
			return err
		}
		vehicles[index].CVValues = values
		files, err := s.loadVehicleCVFiles(ctx, vehicles[index].ID)
		if err != nil {
			return err
		}
		vehicles[index].CVFiles = files
	}
	return nil
}

func (s *VehicleService) CreateImage(ctx context.Context, vehicleID string, input VehicleImageInput) (*VehicleImage, error) {
	return s.imageService().CreateImage(ctx, vehicleID, input)
}

func (s *VehicleService) CreateImageWithRevision(ctx context.Context, vehicleID string, input VehicleImageInput, expectedRevision *string) (*VehicleImageCreation, error) {
	return s.imageService().CreateImageWithRevision(ctx, vehicleID, input, expectedRevision)
}

func (s *VehicleService) GetImage(ctx context.Context, vehicleID, imageID string) (*VehicleImage, error) {
	return s.imageService().GetImage(ctx, vehicleID, imageID)
}

func (s *VehicleService) DeleteImage(ctx context.Context, vehicleID, imageID string, expectedRevision ...string) (*VehicleImage, error) {
	return s.imageService().DeleteImage(ctx, vehicleID, imageID, expectedRevision...)
}

func (s *VehicleService) ImageFileReferenceCount(ctx context.Context, storagePath string) (int, error) {
	return s.MediaFileReferenceCount(ctx, storagePath)
}

func (s *VehicleService) MediaFileReferenceCount(ctx context.Context, storagePath string) (int, error) {
	storagePath = strings.TrimSpace(storagePath)
	if storagePath == "" {
		return 0, nil
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT storage_path FROM vehicle_images UNION ALL SELECT thumbnail_path FROM vehicle_images
UNION ALL SELECT storage_path FROM vehicle_attachments UNION ALL SELECT storage_path FROM vehicle_cv_files
`)
	if err != nil {
		return 0, fmt.Errorf("list media file references: %w", err)
	}
	defer func() { _ = rows.Close() }()
	target := filepath.Clean(filepath.FromSlash(storagePath))
	count := 0
	for rows.Next() {
		var stored sql.NullString
		if err := rows.Scan(&stored); err != nil {
			return 0, fmt.Errorf("scan media file reference: %w", err)
		}
		if stored.Valid && stored.String != "" {
			relative, err := filepath.Rel(target, filepath.Clean(filepath.FromSlash(stored.String)))
			if err == nil && relative == "." {
				count++
			}
		}
	}
	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("iterate media file references: %w", err)
	}
	return count, nil
}

func (s *VehicleService) loadVehicleAttachments(ctx context.Context, vehicleID string) ([]VehicleAttachment, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, vehicle_id, file_name, original_name, COALESCE(description, ''), COALESCE(category, ''),
       COALESCE(mime_type, ''), size_bytes, storage_path, COALESCE(blob_id, ''), COALESCE(maintenance_id, ''), created_at, updated_at
FROM vehicle_attachments
WHERE vehicle_id=?
ORDER BY created_at ASC
`, vehicleID)
	if err != nil {
		return nil, fmt.Errorf("list vehicle attachments: %w", err)
	}
	defer func() { _ = rows.Close() }()

	attachments := []VehicleAttachment{}
	for rows.Next() {
		var attachment VehicleAttachment
		if err := rows.Scan(
			&attachment.ID,
			&attachment.VehicleID,
			&attachment.FileName,
			&attachment.OriginalName,
			&attachment.Description,
			&attachment.Category,
			&attachment.MimeType,
			&attachment.SizeBytes,
			&attachment.StoragePath,
			&attachment.BlobID,
			&attachment.MaintenanceID,
			&attachment.CreatedAt,
			&attachment.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan vehicle attachment: %w", err)
		}
		attachment.Revision = vehicleAttachmentRevision(attachment)
		attachments = append(attachments, attachment)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate vehicle attachments: %w", err)
	}
	return attachments, nil
}

func (s *VehicleService) CreateAttachment(ctx context.Context, vehicleID string, input VehicleAttachmentInput) (*VehicleAttachment, error) {
	vehicleID = strings.TrimSpace(vehicleID)
	input = cleanVehicleAttachmentInput(input)
	if vehicleID == "" || input.FileName == "" || input.OriginalName == "" || (input.StoragePath == "" && input.BlobID == "") {
		return nil, ErrVehicleValidation
	}
	if _, err := s.Get(ctx, vehicleID); err != nil {
		return nil, err
	}
	if err := ensureVehicleMaintenanceID(ctx, s.db, vehicleID, input.MaintenanceID); err != nil {
		return nil, err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	attachment := VehicleAttachment{
		ID:            randomID(),
		VehicleID:     vehicleID,
		FileName:      input.FileName,
		OriginalName:  input.OriginalName,
		Description:   input.Description,
		Category:      input.Category,
		MimeType:      input.MimeType,
		SizeBytes:     input.SizeBytes,
		StoragePath:   input.StoragePath,
		BlobID:        input.BlobID,
		MaintenanceID: input.MaintenanceID,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if _, err := s.db.ExecContext(ctx, `
INSERT INTO vehicle_attachments(id, vehicle_id, file_name, original_name, description, category, mime_type, size_bytes, storage_path, blob_id, maintenance_id, created_at, updated_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, attachment.ID, attachment.VehicleID, attachment.FileName, attachment.OriginalName, attachment.Description, attachment.Category, attachment.MimeType, attachment.SizeBytes, attachment.StoragePath, attachment.BlobID, attachment.MaintenanceID, attachment.CreatedAt, attachment.UpdatedAt); err != nil {
		return nil, fmt.Errorf("create vehicle attachment: %w", err)
	}
	attachment.Revision = vehicleAttachmentRevision(attachment)
	return &attachment, nil
}

func (s *VehicleService) UpdateAttachment(ctx context.Context, vehicleID, attachmentID string, input VehicleAttachmentUpdateInput) (*VehicleAttachment, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin attachment update: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	current, err := readVehicleAttachment(ctx, tx, vehicleID, attachmentID)
	if err != nil {
		return nil, err
	}
	if err := checkVehicleRevision(current.Revision, input.ExpectedRevision, ErrVehicleAttachmentConflict); err != nil {
		return nil, err
	}
	input.MaintenanceID = strings.TrimSpace(input.MaintenanceID)
	if err := ensureVehicleMaintenanceID(ctx, tx, vehicleID, input.MaintenanceID); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `
UPDATE vehicle_attachments SET description=?, category=?, maintenance_id=?, updated_at=? WHERE id=? AND vehicle_id=?
`, strings.TrimSpace(input.Description), strings.TrimSpace(input.Category), input.MaintenanceID,
		time.Now().UTC().Format(time.RFC3339), current.ID, current.VehicleID); err != nil {
		return nil, fmt.Errorf("update vehicle attachment: %w", err)
	}
	updated, err := readVehicleAttachment(ctx, tx, vehicleID, attachmentID)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit attachment update: %w", err)
	}
	return updated, nil
}

func (s *VehicleService) GetAttachment(ctx context.Context, vehicleID, attachmentID string) (*VehicleAttachment, error) {
	return readVehicleAttachment(ctx, s.db, vehicleID, attachmentID)
}

func readVehicleAttachment(ctx context.Context, query queryRower, vehicleID, attachmentID string) (*VehicleAttachment, error) {
	var attachment VehicleAttachment
	err := query.QueryRowContext(ctx, `
SELECT id, vehicle_id, file_name, original_name, COALESCE(description, ''), COALESCE(category, ''),
       COALESCE(mime_type, ''), size_bytes, storage_path, COALESCE(blob_id, ''), COALESCE(maintenance_id, ''), created_at, updated_at
FROM vehicle_attachments
WHERE id=? AND vehicle_id=?
`, strings.TrimSpace(attachmentID), strings.TrimSpace(vehicleID)).Scan(
		&attachment.ID,
		&attachment.VehicleID,
		&attachment.FileName,
		&attachment.OriginalName,
		&attachment.Description,
		&attachment.Category,
		&attachment.MimeType,
		&attachment.SizeBytes,
		&attachment.StoragePath,
		&attachment.BlobID,
		&attachment.MaintenanceID,
		&attachment.CreatedAt,
		&attachment.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrVehicleNotFound
		}
		return nil, fmt.Errorf("get vehicle attachment: %w", err)
	}
	attachment.Revision = vehicleAttachmentRevision(attachment)
	return &attachment, nil
}

func (s *VehicleService) DeleteAttachment(ctx context.Context, vehicleID, attachmentID string, expectedRevision ...string) (*VehicleAttachment, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin attachment delete: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	attachment, err := readVehicleAttachment(ctx, tx, vehicleID, attachmentID)
	if err != nil {
		return nil, err
	}
	if err := checkVehicleRevision(attachment.Revision, optionalVehicleRevision(expectedRevision), ErrVehicleAttachmentConflict); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM vehicle_attachments WHERE id=? AND vehicle_id=?`, attachment.ID, attachment.VehicleID); err != nil {
		return nil, fmt.Errorf("delete vehicle attachment: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit attachment delete: %w", err)
	}
	return attachment, nil
}

func (s *VehicleService) imageService() *vehiclemodel.ImageService {
	return vehiclemodel.NewImageService(vehicledata.NewImageStore(s.db), func(ctx context.Context, id string) error { _, err := s.Get(ctx, id); return err })
}
