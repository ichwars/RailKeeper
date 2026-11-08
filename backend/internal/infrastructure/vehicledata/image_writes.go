package vehicledata

import (
	"context"
	"fmt"
	"strings"

	vehiclemodel "railkeeper/backend/internal/inventory/vehicle"
)

func (m *imageMutation) CountImages(ctx context.Context, vehicleID string) (int, error) {
	var existingCount int
	if err := m.tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM vehicle_images WHERE vehicle_id=?`, vehicleID).Scan(&existingCount); err != nil {
		return 0, fmt.Errorf("count vehicle images: %w", err)
	}
	return existingCount, nil
}

func (m *imageMutation) ClearPrimary(ctx context.Context, vehicleID, now string) error {
	if _, err := m.tx.ExecContext(ctx, `UPDATE vehicle_images SET is_primary=0, updated_at=? WHERE vehicle_id=?`, now, vehicleID); err != nil {
		return fmt.Errorf("clear vehicle image primary flag: %w", err)
	}
	return nil
}

func (m *imageMutation) AppendImage(ctx context.Context, image ImageRecord) error {
	if _, err := m.tx.ExecContext(ctx, `
INSERT INTO vehicle_images(id, vehicle_id, url, title, source_url, file_name, mime_type, storage_path, thumbnail_path, blob_id, thumbnail_blob_id, maintenance_id, is_primary, sort_order, created_at, updated_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, image.ID, image.VehicleID, image.URL, image.Title, image.SourceURL, image.FileName, image.MimeType, image.StoragePath, image.ThumbnailPath, image.BlobID, image.ThumbnailBlobID, image.MaintenanceID, boolToInt(image.IsPrimary), image.SortOrder, image.CreatedAt, image.UpdatedAt); err != nil {
		return fmt.Errorf("create vehicle image: %w", err)
	}
	return nil
}

func (m *imageMutation) DeleteImage(ctx context.Context, vehicleID, imageID string) error {
	result, err := m.tx.ExecContext(ctx, `DELETE FROM vehicle_images WHERE id=? AND vehicle_id=?`, strings.TrimSpace(imageID), strings.TrimSpace(vehicleID))
	if err != nil {
		return fmt.Errorf("delete vehicle image: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read vehicle image delete result: %w", err)
	}
	if affected == 0 {
		return vehiclemodel.ErrVehicleNotFound
	}
	return nil
}

func (m *imageMutation) PromotePrimary(ctx context.Context, vehicleID, now string) error {
	if _, err := m.tx.ExecContext(ctx, `
UPDATE vehicle_images
SET is_primary=1, updated_at=?
WHERE id = (
  SELECT id FROM vehicle_images WHERE vehicle_id=? ORDER BY sort_order ASC, created_at ASC LIMIT 1
)
`, now, strings.TrimSpace(vehicleID)); err != nil {
		return fmt.Errorf("promote vehicle image primary flag: %w", err)
	}
	return nil
}

func (m *imageMutation) ExistingImages(ctx context.Context, id string) (map[string]ImageMetadata, error) {
	return ExistingImages(ctx, m.tx, id)
}
func (m *imageMutation) ClearImages(ctx context.Context, id string) error {
	return ClearImages(ctx, m.tx, id)
}
func (m *imageMutation) InsertImage(ctx context.Context, image ImageRecord) error {
	return InsertImage(ctx, m.tx, image)
}
