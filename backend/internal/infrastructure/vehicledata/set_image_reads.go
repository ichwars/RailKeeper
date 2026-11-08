package vehicledata

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	vehiclemodel "railkeeper/backend/internal/inventory/vehicle"
)

type setImageStore struct{ db *sql.DB }
type setImageMutation struct{ tx *sql.Tx }

func NewSetImageStore(db *sql.DB) vehiclemodel.SetImageRepository { return &setImageStore{db: db} }
func (s *setImageStore) BeginSetImages(ctx context.Context) (vehiclemodel.SetImageMutation, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &setImageMutation{tx: tx}, nil
}
func (m *setImageMutation) Commit() error   { return m.tx.Commit() }
func (m *setImageMutation) Rollback() error { return m.tx.Rollback() }

func (s *setImageStore) Selection(ctx context.Context, setID string) (vehiclemodel.SetImageSelection, error) {
	var mode vehiclemodel.VehicleSetMainImageMode
	var memberImageID, fileName, mimeType, blobID, thumbnailBlobID, createdAt, updatedAt string
	err := s.db.QueryRowContext(ctx, `
SELECT main_image_mode, COALESCE(main_member_image_id, ''), COALESCE(set_image_file_name, ''),
       COALESCE(set_image_mime_type, ''), COALESCE(set_image_blob_id, ''),
       COALESCE(set_image_thumbnail_blob_id, ''), COALESCE(set_image_created_at, ''),
       COALESCE(set_image_updated_at, '')
FROM vehicle_sets WHERE id=?
`, setID).Scan(&mode, &memberImageID, &fileName, &mimeType, &blobID, &thumbnailBlobID, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return vehiclemodel.SetImageSelection{}, vehiclemodel.ErrVehicleSetNotFound
		}
		return vehiclemodel.SetImageSelection{}, fmt.Errorf("read vehicle set image selection: %w", err)
	}
	return vehiclemodel.SetImageSelection{Mode: mode, MemberImageID: memberImageID, Dedicated: vehiclemodel.VehicleSetImage{FileName: fileName, MimeType: mimeType, BlobID: blobID, ThumbnailBlobID: thumbnailBlobID, CreatedAt: createdAt, UpdatedAt: updatedAt}}, nil
}

func (s *setImageStore) DedicatedImage(ctx context.Context, setID string) (vehiclemodel.VehicleSetImage, error) {
	var image vehiclemodel.VehicleSetImage
	err := s.db.QueryRowContext(ctx, `
SELECT COALESCE(set_image_file_name, ''), COALESCE(set_image_mime_type, ''), COALESCE(set_image_blob_id, ''),
       COALESCE(set_image_thumbnail_blob_id, ''), COALESCE(set_image_created_at, ''), COALESCE(set_image_updated_at, '')
FROM vehicle_sets WHERE id=?
`, strings.TrimSpace(setID)).Scan(
		&image.FileName, &image.MimeType, &image.BlobID, &image.ThumbnailBlobID, &image.CreatedAt, &image.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return vehiclemodel.VehicleSetImage{}, vehiclemodel.ErrVehicleSetNotFound
		}
		return vehiclemodel.VehicleSetImage{}, fmt.Errorf("get vehicle set image: %w", err)
	}
	return image, nil
}

func (s *setImageStore) MemberImage(ctx context.Context, setID, imageID string) (*vehiclemodel.VehicleImage, error) {
	image, err := s.scanSetMemberImage(s.db.QueryRowContext(ctx, `
SELECT image.id, image.vehicle_id, image.url, COALESCE(image.thumbnail_blob_id, ''),
       COALESCE(image.title, ''), COALESCE(image.file_name, ''), COALESCE(image.mime_type, ''),
       COALESCE(image.blob_id, ''), image.is_primary, image.sort_order, image.created_at
FROM vehicle_images image
JOIN vehicle_set_members member ON member.vehicle_id=image.vehicle_id
WHERE member.vehicle_set_id=? AND image.id=?
`, setID, imageID))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return image, err
}

func (s *setImageStore) AutomaticImage(ctx context.Context, setID string) (*vehiclemodel.VehicleImage, error) {
	image, err := s.scanSetMemberImage(s.db.QueryRowContext(ctx, `
SELECT image.id, image.vehicle_id, image.url, COALESCE(image.thumbnail_blob_id, ''),
       COALESCE(image.title, ''), COALESCE(image.file_name, ''), COALESCE(image.mime_type, ''),
       COALESCE(image.blob_id, ''), image.is_primary, image.sort_order, image.created_at
FROM vehicle_set_members member
JOIN vehicle_images image ON image.vehicle_id=member.vehicle_id
WHERE member.vehicle_set_id=?
ORDER BY member.position ASC, image.is_primary DESC, image.sort_order ASC, image.created_at ASC
LIMIT 1
`, setID))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return image, err
}

type rowScanner interface {
	Scan(dest ...any) error
}

func (s *setImageStore) scanSetMemberImage(row rowScanner) (*vehiclemodel.VehicleImage, error) {
	var image vehiclemodel.VehicleImage
	var thumbnailBlobID string
	var isPrimary int
	if err := row.Scan(&image.ID, &image.VehicleID, &image.URL, &thumbnailBlobID, &image.Title,
		&image.FileName, &image.MimeType, &image.BlobID, &isPrimary, &image.SortOrder, &image.CreatedAt); err != nil {
		return nil, err
	}
	image.ThumbnailBlobID = thumbnailBlobID
	image.IsPrimary = isPrimary == 1
	image = vehiclemodel.WithImageURLs(image)
	return &image, nil
}
