package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

func saveVehicleImages(ctx context.Context, tx *sql.Tx, vehicleID string, images []VehicleImageInput, now string) error {
	existing, err := existingVehicleImageMeta(ctx, tx, vehicleID)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM vehicle_images WHERE vehicle_id=?`, vehicleID); err != nil {
		return fmt.Errorf("clear vehicle images: %w", err)
	}
	cleaned := cleanVehicleImageInputs(images)
	for index, image := range cleaned {
		image.MaintenanceID = strings.TrimSpace(image.MaintenanceID)
		if err := ensureVehicleMaintenanceID(ctx, tx, vehicleID, image.MaintenanceID); err != nil {
			return err
		}
		meta, hasMeta := existing[image.ID]
		if !hasMeta {
			meta = existing[image.URL]
		}
		imageID := randomID()
		createdAt := now
		if hasMeta || meta.ID != "" {
			imageID = meta.ID
			createdAt = meta.CreatedAt
			image.FileName = meta.FileName
			image.MimeType = meta.MimeType
			image.StoragePath = meta.StoragePath
			image.ThumbnailPath = meta.ThumbnailPath
			image.BlobID = meta.BlobID
			image.ThumbnailBlobID = meta.ThumbnailBlobID
		}
		sortOrder := image.SortOrder
		if sortOrder == 0 {
			sortOrder = index
		}
		imageURL := image.URL
		if image.StoragePath != "" || image.BlobID != "" {
			imageURL = "/api/v1/vehicles/" + vehicleID + "/images/" + imageID + "/file"
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO vehicle_images(id, vehicle_id, url, title, source_url, file_name, mime_type, storage_path, thumbnail_path, blob_id, thumbnail_blob_id, maintenance_id, is_primary, sort_order, created_at, updated_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, imageID, vehicleID, imageURL, image.Title, image.SourceURL, image.FileName, image.MimeType, image.StoragePath, image.ThumbnailPath, image.BlobID, image.ThumbnailBlobID, image.MaintenanceID, boolToInt(image.IsPrimary), sortOrder, createdAt, now); err != nil {
			return fmt.Errorf("insert vehicle image: %w", err)
		}
	}
	return nil
}

type queryRower interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func ensureVehicleMaintenanceID(ctx context.Context, query queryRower, vehicleID, maintenanceID string) error {
	maintenanceID = strings.TrimSpace(maintenanceID)
	if maintenanceID == "" {
		return nil
	}
	var count int
	if err := query.QueryRowContext(ctx, `SELECT COUNT(*) FROM vehicle_maintenance WHERE id=? AND vehicle_id=?`, maintenanceID, vehicleID).Scan(&count); err != nil {
		return fmt.Errorf("validate vehicle maintenance link: %w", err)
	}
	if count == 0 {
		return ErrVehicleValidation
	}
	return nil
}

type vehicleImageMeta struct {
	ID              string
	URL             string
	FileName        string
	MimeType        string
	StoragePath     string
	ThumbnailPath   string
	BlobID          string
	ThumbnailBlobID string
	MaintenanceID   string
	CreatedAt       string
}

func existingVehicleImageMeta(ctx context.Context, tx *sql.Tx, vehicleID string) (map[string]vehicleImageMeta, error) {
	rows, err := tx.QueryContext(ctx, `
SELECT id, url, COALESCE(file_name, ''), COALESCE(mime_type, ''), COALESCE(storage_path, ''), COALESCE(thumbnail_path, ''), COALESCE(blob_id, ''), COALESCE(thumbnail_blob_id, ''), COALESCE(maintenance_id, ''), created_at
FROM vehicle_images
WHERE vehicle_id=?
`, vehicleID)
	if err != nil {
		return nil, fmt.Errorf("list existing vehicle image metadata: %w", err)
	}
	defer func() { _ = rows.Close() }()
	out := map[string]vehicleImageMeta{}
	for rows.Next() {
		var meta vehicleImageMeta
		if err := rows.Scan(&meta.ID, &meta.URL, &meta.FileName, &meta.MimeType, &meta.StoragePath, &meta.ThumbnailPath, &meta.BlobID, &meta.ThumbnailBlobID, &meta.MaintenanceID, &meta.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan existing vehicle image metadata: %w", err)
		}
		out[meta.ID] = meta
		out[meta.URL] = meta
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate existing vehicle image metadata: %w", err)
	}
	return out, nil
}

func (s *VehicleService) nextInventoryNumber(ctx context.Context, tx *sql.Tx, vehicleCategory string) (string, error) {
	requestedCategory := cleanInventoryCategory(vehicleCategory)
	fallbackCategory := inventoryCategoryForVehicle(vehicleCategory)
	scheme, err := s.inventoryNumberSchemeForUpdate(ctx, tx, requestedCategory, fallbackCategory)
	if err != nil {
		return "", err
	}

	next := scheme.NextNumber
	for attempts := 0; attempts < 500; attempts++ {
		candidate := formatInventoryNumber(scheme.Prefix, next, scheme.Padding)
		if err := s.ensureInventoryNumberAvailable(ctx, tx, candidate, ""); err == nil {
			if _, err = tx.ExecContext(ctx, `
UPDATE inventory_number_schemes
SET next_number=?, updated_at=?
WHERE category=?
`, next+1, time.Now().UTC().Format(time.RFC3339), scheme.Category); err != nil {
				return "", fmt.Errorf("advance inventory number scheme: %w", err)
			}
			return candidate, nil
		} else if !errors.Is(err, ErrInventoryNumberConflict) {
			return "", err
		}
		next++
	}

	return "", fmt.Errorf("next inventory number: exhausted attempts for %s", scheme.Category)
}

func (s *VehicleService) inventoryNumberSchemeForUpdate(ctx context.Context, tx *sql.Tx, category string, fallbackCategory string) (*InventoryNumberScheme, error) {
	if category != "" {
		if scheme, err := s.readActiveInventoryNumberScheme(ctx, tx, category); err == nil {
			return scheme, nil
		} else if !errors.Is(err, ErrInventoryNumberNotFound) {
			return nil, err
		}
	}
	if fallbackCategory != "" && fallbackCategory != category {
		if scheme, err := s.readActiveInventoryNumberScheme(ctx, tx, fallbackCategory); err == nil {
			return scheme, nil
		} else if !errors.Is(err, ErrInventoryNumberNotFound) {
			return nil, err
		}
	}
	if fallbackCategory != "Fahrzeug" && category != "Fahrzeug" {
		return s.readActiveInventoryNumberScheme(ctx, tx, "Fahrzeug")
	}
	return nil, ErrInventoryNumberNotFound
}

func (s *VehicleService) readActiveInventoryNumberScheme(ctx context.Context, tx *sql.Tx, category string) (*InventoryNumberScheme, error) {
	var scheme InventoryNumberScheme
	var active int
	err := tx.QueryRowContext(ctx, `
SELECT id, category, prefix, next_number, padding, active, created_at, updated_at
FROM inventory_number_schemes
WHERE category=? AND active=1
`, category).Scan(
		&scheme.ID,
		&scheme.Category,
		&scheme.Prefix,
		&scheme.NextNumber,
		&scheme.Padding,
		&active,
		&scheme.CreatedAt,
		&scheme.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInventoryNumberNotFound
		}
		return nil, fmt.Errorf("read inventory number scheme: %w", err)
	}
	scheme.Active = active == 1
	return &scheme, nil
}

func (s *VehicleService) ensureInventoryNumberAvailable(ctx context.Context, tx *sql.Tx, inventoryNumber, excludeVehicleID string) error {
	inventoryNumber = strings.TrimSpace(inventoryNumber)
	if inventoryNumber == "" {
		return ErrInventoryNumberValidation
	}
	var count int
	if err := tx.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM vehicles
WHERE inventory_number=? AND (? = '' OR id <> ?)
`, inventoryNumber, excludeVehicleID, excludeVehicleID).Scan(&count); err != nil {
		return fmt.Errorf("check inventory number availability: %w", err)
	}
	if count > 0 {
		return ErrInventoryNumberConflict
	}
	return nil
}

func vehicleImagesFromInput(vehicleID string, images []VehicleImageInput, now string) []VehicleImage {
	cleaned := cleanVehicleImageInputs(images)
	out := make([]VehicleImage, 0, len(cleaned))
	for index, image := range cleaned {
		sortOrder := image.SortOrder
		if sortOrder == 0 {
			sortOrder = index
		}
		out = append(out, VehicleImage{
			ID:              image.ID,
			VehicleID:       vehicleID,
			URL:             image.URL,
			Title:           image.Title,
			SourceURL:       image.SourceURL,
			FileName:        image.FileName,
			MimeType:        image.MimeType,
			StoragePath:     image.StoragePath,
			ThumbnailPath:   image.ThumbnailPath,
			BlobID:          image.BlobID,
			ThumbnailBlobID: image.ThumbnailBlobID,
			MaintenanceID:   image.MaintenanceID,
			IsPrimary:       image.IsPrimary,
			SortOrder:       sortOrder,
			CreatedAt:       now,
		})
	}
	return out
}

func withVehicleImageURLs(image VehicleImage) VehicleImage {
	if image.StoragePath != "" || image.BlobID != "" {
		image.URL = "/api/v1/vehicles/" + image.VehicleID + "/images/" + image.ID + "/file"
	}
	if image.ThumbnailPath != "" || image.ThumbnailBlobID != "" {
		image.ThumbnailURL = "/api/v1/vehicles/" + image.VehicleID + "/images/" + image.ID + "/thumbnail"
	}
	return image
}
