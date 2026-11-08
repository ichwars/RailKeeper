package vehicledata

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	vehiclemodel "railkeeper/backend/internal/inventory/vehicle"
)

func (m *setImageMutation) Exists(ctx context.Context, setID string) (bool, error) {
	var setExists int
	if err := m.tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM vehicle_sets WHERE id=?)`, setID).
		Scan(&setExists); err != nil {
		return false, fmt.Errorf("check vehicle set for main image update: %w", err)
	}
	return setExists != 0, nil
}

func (m *setImageMutation) MemberBelongs(ctx context.Context, setID, imageID string) (bool, error) {
	var count int
	if err := m.tx.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM vehicle_images image
JOIN vehicle_set_members member ON member.vehicle_id=image.vehicle_id
WHERE member.vehicle_set_id=? AND image.id=?
`, setID, imageID).Scan(&count); err != nil {
		return false, fmt.Errorf("validate vehicle set member image: %w", err)
	}
	return count != 0, nil
}

func (m *setImageMutation) DedicatedBlob(ctx context.Context, setID string) (string, error) {
	var blobID string
	if err := m.tx.QueryRowContext(ctx, `
SELECT COALESCE(set_image_blob_id, '') FROM vehicle_sets WHERE id=?
`, setID).Scan(&blobID); err != nil {
		return "", fmt.Errorf("read dedicated vehicle set image: %w", err)
	}
	return blobID, nil
}

func (m *setImageMutation) UpdateSelection(ctx context.Context, setID string, input vehiclemodel.VehicleSetMainImageInput, now string) error {
	result, err := m.tx.ExecContext(ctx, `
UPDATE vehicle_sets SET main_image_mode=?, main_member_image_id=NULLIF(?, ''), updated_at=? WHERE id=?
`, input.Mode, input.MemberImageID, now, setID)
	if err != nil {
		return fmt.Errorf("set vehicle set main image: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read vehicle set main image update result: %w", err)
	}
	if affected == 0 {
		return vehiclemodel.ErrVehicleSetNotFound
	}
	return nil
}

func (m *setImageMutation) PreviousForUpload(ctx context.Context, setID string) (vehiclemodel.PreviousSetImage, error) {
	var oldBlobID, oldThumbnailBlobID, createdAt string
	if err := m.tx.QueryRowContext(ctx, `
SELECT COALESCE(set_image_blob_id, ''), COALESCE(set_image_thumbnail_blob_id, ''),
       COALESCE(set_image_created_at, '')
FROM vehicle_sets WHERE id=?
`, setID).Scan(&oldBlobID, &oldThumbnailBlobID, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return vehiclemodel.PreviousSetImage{}, vehiclemodel.ErrVehicleSetNotFound
		}
		return vehiclemodel.PreviousSetImage{}, fmt.Errorf("read vehicle set image: %w", err)
	}
	return vehiclemodel.PreviousSetImage{BlobID: oldBlobID, ThumbnailBlobID: oldThumbnailBlobID, CreatedAt: createdAt}, nil
}

func (m *setImageMutation) PreviousForDelete(ctx context.Context, setID string) (vehiclemodel.PreviousSetImage, error) {
	var blobID, thumbnailBlobID string
	if err := m.tx.QueryRowContext(ctx, `
SELECT COALESCE(set_image_blob_id, ''), COALESCE(set_image_thumbnail_blob_id, '')
FROM vehicle_sets WHERE id=?
`, setID).Scan(&blobID, &thumbnailBlobID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return vehiclemodel.PreviousSetImage{}, vehiclemodel.ErrVehicleSetNotFound
		}
		return vehiclemodel.PreviousSetImage{}, fmt.Errorf("read vehicle set image for deletion: %w", err)
	}
	return vehiclemodel.PreviousSetImage{BlobID: blobID, ThumbnailBlobID: thumbnailBlobID}, nil
}

func (m *setImageMutation) StoreDedicated(ctx context.Context, setID string, input vehiclemodel.VehicleSetImageInput, createdAt, now string) error {
	if _, err := m.tx.ExecContext(ctx, `
UPDATE vehicle_sets
SET set_image_file_name=?, set_image_mime_type=?, set_image_blob_id=?, set_image_thumbnail_blob_id=NULLIF(?, ''),
    set_image_created_at=?, set_image_updated_at=?, main_image_mode='dedicated', main_member_image_id=NULL, updated_at=?
WHERE id=?
`, input.FileName, input.MimeType, input.BlobID, input.ThumbnailBlobID, createdAt, now, now, setID); err != nil {
		return fmt.Errorf("store vehicle set image: %w", err)
	}
	return nil
}

func (m *setImageMutation) DeleteDedicated(ctx context.Context, setID, now string) error {
	if _, err := m.tx.ExecContext(ctx, `
UPDATE vehicle_sets
SET set_image_file_name='', set_image_mime_type='', set_image_blob_id=NULL,
    set_image_thumbnail_blob_id=NULL, set_image_created_at='', set_image_updated_at='',
    main_image_mode=CASE WHEN main_image_mode='dedicated' THEN 'automatic' ELSE main_image_mode END,
    updated_at=?
WHERE id=?
`, now, setID); err != nil {
		return fmt.Errorf("delete vehicle set image: %w", err)
	}
	return nil
}

func (m *setImageMutation) Audit(
	ctx context.Context,
	id, actorUserID, setID, action, now string,
) error {
	if _, err := m.tx.ExecContext(ctx, `
INSERT INTO audit_logs(id, actor_user_id, action, target_type, target_id, created_at, details_json)
VALUES(?, ?, ?, 'vehicle_set', ?, ?, '{}')
`, id, actorUserID, action, setID, now); err != nil {
		return fmt.Errorf("write vehicle set image audit log: %w", err)
	}
	return nil
}
