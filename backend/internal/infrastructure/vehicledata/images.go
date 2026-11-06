package vehicledata

import (
	"context"
	"database/sql"
	"fmt"
)

type ImageRecord struct {
	ID              string
	VehicleID       string
	URL             string
	Title           string
	SourceURL       string
	FileName        string
	MimeType        string
	StoragePath     string
	ThumbnailPath   string
	BlobID          string
	ThumbnailBlobID string
	MaintenanceID   string
	IsPrimary       bool
	SortOrder       int
	CreatedAt       string
	UpdatedAt       string
}

type ImageMetadata struct {
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

func ExistingImages(ctx context.Context, tx *sql.Tx, vehicleID string) (map[string]ImageMetadata, error) {
	rows, err := tx.QueryContext(ctx, `
SELECT id, url, COALESCE(file_name, ''), COALESCE(mime_type, ''), COALESCE(storage_path, ''), COALESCE(thumbnail_path, ''), COALESCE(blob_id, ''), COALESCE(thumbnail_blob_id, ''), COALESCE(maintenance_id, ''), created_at
FROM vehicle_images
WHERE vehicle_id=?
`, vehicleID)
	if err != nil {
		return nil, fmt.Errorf("list existing vehicle image metadata: %w", err)
	}
	defer func() { _ = rows.Close() }()
	out := map[string]ImageMetadata{}
	for rows.Next() {
		var meta ImageMetadata
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

func ClearImages(ctx context.Context, tx *sql.Tx, vehicleID string) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM vehicle_images WHERE vehicle_id=?`, vehicleID); err != nil {
		return fmt.Errorf("clear vehicle images: %w", err)
	}
	return nil
}

func InsertImage(ctx context.Context, tx *sql.Tx, image ImageRecord) error {
	if _, err := tx.ExecContext(ctx, `
INSERT INTO vehicle_images(id, vehicle_id, url, title, source_url, file_name, mime_type, storage_path, thumbnail_path, blob_id, thumbnail_blob_id, maintenance_id, is_primary, sort_order, created_at, updated_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, image.ID, image.VehicleID, image.URL, image.Title, image.SourceURL, image.FileName, image.MimeType, image.StoragePath, image.ThumbnailPath, image.BlobID, image.ThumbnailBlobID, image.MaintenanceID, boolToInt(image.IsPrimary), image.SortOrder, image.CreatedAt, image.UpdatedAt); err != nil {
		return fmt.Errorf("insert vehicle image: %w", err)
	}
	return nil
}
