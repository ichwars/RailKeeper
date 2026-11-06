package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"
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
	rows, err := query.QueryContext(ctx, `
SELECT id, vehicle_id, url, COALESCE(title, ''), COALESCE(source_url, ''), COALESCE(file_name, ''), COALESCE(mime_type, ''), COALESCE(storage_path, ''), COALESCE(thumbnail_path, ''), COALESCE(blob_id, ''), COALESCE(thumbnail_blob_id, ''), COALESCE(maintenance_id, ''), is_primary, sort_order, created_at, COALESCE(updated_at, '')
FROM vehicle_images
WHERE vehicle_id=?
ORDER BY is_primary DESC, sort_order ASC, created_at ASC
`, vehicleID)
	if err != nil {
		return nil, fmt.Errorf("list vehicle images: %w", err)
	}
	defer func() { _ = rows.Close() }()

	images := []VehicleImage{}
	for rows.Next() {
		var image VehicleImage
		var isPrimary int
		if err := rows.Scan(&image.ID, &image.VehicleID, &image.URL, &image.Title, &image.SourceURL, &image.FileName, &image.MimeType, &image.StoragePath, &image.ThumbnailPath, &image.BlobID, &image.ThumbnailBlobID, &image.MaintenanceID, &isPrimary, &image.SortOrder, &image.CreatedAt, &image.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan vehicle image: %w", err)
		}
		image.IsPrimary = isPrimary == 1
		image.Revision = vehicleImageRevision(image)
		image = withVehicleImageURLs(image)
		images = append(images, image)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate vehicle images: %w", err)
	}
	return images, nil
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
	vehicleID = strings.TrimSpace(vehicleID)
	input = cleanVehicleImageInput(input)
	if vehicleID == "" || input.FileName == "" || input.MimeType == "" || (input.StoragePath == "" && input.BlobID == "") {
		return nil, ErrVehicleValidation
	}
	if _, err := s.Get(ctx, vehicleID); err != nil {
		return nil, err
	}
	if err := ensureVehicleMaintenanceID(ctx, s.db, vehicleID, input.MaintenanceID); err != nil {
		return nil, err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	imageID := randomID()
	image := VehicleImage{
		ID:              imageID,
		VehicleID:       vehicleID,
		URL:             "/api/v1/vehicles/" + vehicleID + "/images/" + imageID + "/file",
		ThumbnailURL:    "/api/v1/vehicles/" + vehicleID + "/images/" + imageID + "/thumbnail",
		Title:           input.Title,
		SourceURL:       input.SourceURL,
		FileName:        input.FileName,
		MimeType:        input.MimeType,
		StoragePath:     input.StoragePath,
		ThumbnailPath:   input.ThumbnailPath,
		BlobID:          input.BlobID,
		ThumbnailBlobID: input.ThumbnailBlobID,
		MaintenanceID:   input.MaintenanceID,
		IsPrimary:       input.IsPrimary,
		SortOrder:       input.SortOrder,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin vehicle image create: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := ensureVehicleMaintenanceID(ctx, tx, vehicleID, input.MaintenanceID); err != nil {
		return nil, err
	}
	var existingCount int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM vehicle_images WHERE vehicle_id=?`, vehicleID).Scan(&existingCount); err != nil {
		return nil, fmt.Errorf("count vehicle images: %w", err)
	}
	if existingCount == 0 {
		image.IsPrimary = true
	}
	if image.IsPrimary {
		if _, err := tx.ExecContext(ctx, `UPDATE vehicle_images SET is_primary=0, updated_at=? WHERE vehicle_id=?`, now, vehicleID); err != nil {
			return nil, fmt.Errorf("clear vehicle image primary flag: %w", err)
		}
	}
	if image.SortOrder == 0 {
		image.SortOrder = existingCount
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO vehicle_images(id, vehicle_id, url, title, source_url, file_name, mime_type, storage_path, thumbnail_path, blob_id, thumbnail_blob_id, maintenance_id, is_primary, sort_order, created_at, updated_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, image.ID, image.VehicleID, image.URL, image.Title, image.SourceURL, image.FileName, image.MimeType, image.StoragePath, image.ThumbnailPath, image.BlobID, image.ThumbnailBlobID, image.MaintenanceID, boolToInt(image.IsPrimary), image.SortOrder, image.CreatedAt, image.UpdatedAt); err != nil {
		return nil, fmt.Errorf("create vehicle image: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit vehicle image create: %w", err)
	}
	image.Revision = vehicleImageRevision(image)
	return &image, nil
}

func (s *VehicleService) GetImage(ctx context.Context, vehicleID, imageID string) (*VehicleImage, error) {
	return readVehicleImage(ctx, s.db, vehicleID, imageID)
}

func readVehicleImage(ctx context.Context, query queryRower, vehicleID, imageID string) (*VehicleImage, error) {
	var image VehicleImage
	var isPrimary int
	err := query.QueryRowContext(ctx, `
SELECT id, vehicle_id, url, COALESCE(title, ''), COALESCE(source_url, ''), COALESCE(file_name, ''), COALESCE(mime_type, ''), COALESCE(storage_path, ''), COALESCE(thumbnail_path, ''), COALESCE(blob_id, ''), COALESCE(thumbnail_blob_id, ''), COALESCE(maintenance_id, ''), is_primary, sort_order, created_at, COALESCE(updated_at, '')
FROM vehicle_images
WHERE id=? AND vehicle_id=?
`, strings.TrimSpace(imageID), strings.TrimSpace(vehicleID)).Scan(
		&image.ID,
		&image.VehicleID,
		&image.URL,
		&image.Title,
		&image.SourceURL,
		&image.FileName,
		&image.MimeType,
		&image.StoragePath,
		&image.ThumbnailPath,
		&image.BlobID,
		&image.ThumbnailBlobID,
		&image.MaintenanceID,
		&isPrimary,
		&image.SortOrder,
		&image.CreatedAt,
		&image.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrVehicleNotFound
		}
		return nil, fmt.Errorf("get vehicle image: %w", err)
	}
	image.IsPrimary = isPrimary == 1
	image.Revision = vehicleImageRevision(image)
	image = withVehicleImageURLs(image)
	return &image, nil
}

func (s *VehicleService) DeleteImage(ctx context.Context, vehicleID, imageID string, expectedRevision ...string) (*VehicleImage, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin vehicle image delete: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	image, err := readVehicleImage(ctx, tx, vehicleID, imageID)
	if err != nil {
		return nil, err
	}
	if err := checkVehicleRevision(image.Revision, optionalVehicleRevision(expectedRevision), ErrVehicleImageConflict); err != nil {
		return nil, err
	}
	if image.MaintenanceID != "" {
		return nil, ErrVehicleImageInUse
	}

	result, err := tx.ExecContext(ctx, `DELETE FROM vehicle_images WHERE id=? AND vehicle_id=?`, strings.TrimSpace(imageID), strings.TrimSpace(vehicleID))
	if err != nil {
		return nil, fmt.Errorf("delete vehicle image: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("read vehicle image delete result: %w", err)
	}
	if affected == 0 {
		return nil, ErrVehicleNotFound
	}
	if image.IsPrimary {
		now := time.Now().UTC().Format(time.RFC3339)
		if _, err := tx.ExecContext(ctx, `
UPDATE vehicle_images
SET is_primary=1, updated_at=?
WHERE id = (
  SELECT id FROM vehicle_images WHERE vehicle_id=? ORDER BY sort_order ASC, created_at ASC LIMIT 1
)
`, now, strings.TrimSpace(vehicleID)); err != nil {
			return nil, fmt.Errorf("promote vehicle image primary flag: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit vehicle image delete: %w", err)
	}
	return image, nil
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
		if stored.Valid && stored.String != "" && filepath.Clean(filepath.FromSlash(stored.String)) == target {
			count++
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
