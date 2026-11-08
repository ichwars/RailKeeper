package vehicledata

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	vehiclemodel "railkeeper/backend/internal/inventory/vehicle"
)

type imageQuery interface {
	queryRower
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}
type imageReader struct{ query imageQuery }
type imageStore struct {
	*imageReader
	db *sql.DB
}
type imageMutation struct {
	*imageReader
	tx *sql.Tx
}

func NewImageStore(db *sql.DB) vehiclemodel.ImageRepository {
	return &imageStore{imageReader: &imageReader{query: db}, db: db}
}
func ImageMutationFor(tx *sql.Tx) vehiclemodel.ImageMutation {
	return &imageMutation{imageReader: &imageReader{query: tx}, tx: tx}
}
func (s *imageStore) BeginImages(ctx context.Context) (vehiclemodel.ImageMutation, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return ImageMutationFor(tx), nil
}
func (m *imageMutation) Commit() error   { return m.tx.Commit() }
func (m *imageMutation) Rollback() error { return m.tx.Rollback() }
func (r *imageReader) Images(ctx context.Context, id string) ([]vehiclemodel.VehicleImage, error) {
	return ReadImages(ctx, r.query, id)
}
func (r *imageReader) Image(ctx context.Context, id, imageID string) (*vehiclemodel.VehicleImage, error) {
	return ReadImage(ctx, r.query, id, imageID)
}
func (r *imageReader) MaintenanceBelongs(ctx context.Context, id, maintenanceID string) (bool, error) {
	return MaintenanceBelongsToVehicle(ctx, r.query, id, maintenanceID)
}

type vehicleMediaQuery interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func ReadImages(ctx context.Context, query vehicleMediaQuery, vehicleID string) ([]vehiclemodel.VehicleImage, error) {
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

	images := []vehiclemodel.VehicleImage{}
	for rows.Next() {
		var image vehiclemodel.VehicleImage
		var isPrimary int
		if err := rows.Scan(&image.ID, &image.VehicleID, &image.URL, &image.Title, &image.SourceURL, &image.FileName, &image.MimeType, &image.StoragePath, &image.ThumbnailPath, &image.BlobID, &image.ThumbnailBlobID, &image.MaintenanceID, &isPrimary, &image.SortOrder, &image.CreatedAt, &image.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan vehicle image: %w", err)
		}
		image.IsPrimary = isPrimary == 1
		image.Revision = vehiclemodel.ImageRevision(image)
		image = vehiclemodel.WithImageURLs(image)
		images = append(images, image)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate vehicle images: %w", err)
	}
	return images, nil
}
func ReadImage(ctx context.Context, query queryRower, vehicleID, imageID string) (*vehiclemodel.VehicleImage, error) {
	var image vehiclemodel.VehicleImage
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
			return nil, vehiclemodel.ErrVehicleNotFound
		}
		return nil, fmt.Errorf("get vehicle image: %w", err)
	}
	image.IsPrimary = isPrimary == 1
	image.Revision = vehiclemodel.ImageRevision(image)
	image = vehiclemodel.WithImageURLs(image)
	return &image, nil
}
