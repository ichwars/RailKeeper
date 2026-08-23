package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

func (s *VehicleService) SetSetMainImage(
	ctx context.Context,
	setID string,
	input VehicleSetMainImageInput,
	actorUserID string,
) (*VehicleSet, error) {
	setID = strings.TrimSpace(setID)
	input.MemberImageID = strings.TrimSpace(input.MemberImageID)
	if setID == "" {
		return nil, ErrVehicleSetNotFound
	}
	if input.Mode != VehicleSetMainImageModeAutomatic && input.Mode != VehicleSetMainImageModeMember &&
		input.Mode != VehicleSetMainImageModeDedicated {
		return nil, ErrVehicleSetImageValidation
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin vehicle set main image update: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	var setExists int
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM vehicle_sets WHERE id=?)`, setID).
		Scan(&setExists); err != nil {
		return nil, fmt.Errorf("check vehicle set for main image update: %w", err)
	}
	if setExists == 0 {
		return nil, ErrVehicleSetNotFound
	}
	if input.Mode == VehicleSetMainImageModeMember {
		var count int
		if err := tx.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM vehicle_images image
JOIN vehicle_set_members member ON member.vehicle_id=image.vehicle_id
WHERE member.vehicle_set_id=? AND image.id=?
`, setID, input.MemberImageID).Scan(&count); err != nil {
			return nil, fmt.Errorf("validate vehicle set member image: %w", err)
		}
		if count == 0 {
			return nil, ErrVehicleSetImageValidation
		}
	}
	if input.Mode == VehicleSetMainImageModeDedicated {
		var blobID string
		if err := tx.QueryRowContext(ctx, `
SELECT COALESCE(set_image_blob_id, '') FROM vehicle_sets WHERE id=?
`, setID).Scan(&blobID); err != nil {
			return nil, fmt.Errorf("read dedicated vehicle set image: %w", err)
		}
		if blobID == "" {
			return nil, ErrVehicleSetImageValidation
		}
	}
	if input.Mode != VehicleSetMainImageModeMember {
		input.MemberImageID = ""
	}
	now := time.Now().UTC().Format(time.RFC3339)
	result, err := tx.ExecContext(ctx, `
UPDATE vehicle_sets SET main_image_mode=?, main_member_image_id=NULLIF(?, ''), updated_at=? WHERE id=?
`, input.Mode, input.MemberImageID, now, setID)
	if err != nil {
		return nil, fmt.Errorf("set vehicle set main image: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("read vehicle set main image update result: %w", err)
	}
	if affected == 0 {
		return nil, ErrVehicleSetNotFound
	}
	if err := auditVehicleSetImageTx(ctx, tx, actorUserID, setID, "VehicleSetMainImageChanged", now); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit vehicle set main image update: %w", err)
	}
	return s.GetSet(ctx, setID)
}

func (s *VehicleService) UpsertSetImage(
	ctx context.Context,
	setID string,
	input VehicleSetImageInput,
	actorUserID string,
) (*VehicleSet, []string, error) {
	setID = strings.TrimSpace(setID)
	input.FileName = strings.TrimSpace(input.FileName)
	input.MimeType = strings.TrimSpace(input.MimeType)
	input.BlobID = strings.TrimSpace(input.BlobID)
	input.ThumbnailBlobID = strings.TrimSpace(input.ThumbnailBlobID)
	if setID == "" {
		return nil, nil, ErrVehicleSetNotFound
	}
	if input.FileName == "" || input.MimeType == "" || input.BlobID == "" {
		return nil, nil, ErrVehicleSetImageValidation
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("begin vehicle set image upload: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	var oldBlobID, oldThumbnailBlobID, createdAt string
	if err := tx.QueryRowContext(ctx, `
SELECT COALESCE(set_image_blob_id, ''), COALESCE(set_image_thumbnail_blob_id, ''),
       COALESCE(set_image_created_at, '')
FROM vehicle_sets WHERE id=?
`, setID).Scan(&oldBlobID, &oldThumbnailBlobID, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, ErrVehicleSetNotFound
		}
		return nil, nil, fmt.Errorf("read vehicle set image: %w", err)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if createdAt == "" {
		createdAt = now
	}
	if _, err := tx.ExecContext(ctx, `
UPDATE vehicle_sets
SET set_image_file_name=?, set_image_mime_type=?, set_image_blob_id=?, set_image_thumbnail_blob_id=NULLIF(?, ''),
    set_image_created_at=?, set_image_updated_at=?, main_image_mode='dedicated', main_member_image_id=NULL, updated_at=?
WHERE id=?
`, input.FileName, input.MimeType, input.BlobID, input.ThumbnailBlobID, createdAt, now, now, setID); err != nil {
		return nil, nil, fmt.Errorf("store vehicle set image: %w", err)
	}
	if err := auditVehicleSetImageTx(ctx, tx, actorUserID, setID, "VehicleSetImageUploaded", now); err != nil {
		return nil, nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, nil, fmt.Errorf("commit vehicle set image upload: %w", err)
	}
	replaced := distinctNonEmpty(oldBlobID, oldThumbnailBlobID)
	set, err := s.GetSet(ctx, setID)
	if err != nil {
		return nil, replaced, err
	}
	return set, replaced, nil
}

func (s *VehicleService) DeleteSetImage(ctx context.Context, setID, actorUserID string) ([]string, error) {
	setID = strings.TrimSpace(setID)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin vehicle set image deletion: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	var blobID, thumbnailBlobID string
	if err := tx.QueryRowContext(ctx, `
SELECT COALESCE(set_image_blob_id, ''), COALESCE(set_image_thumbnail_blob_id, '')
FROM vehicle_sets WHERE id=?
`, setID).Scan(&blobID, &thumbnailBlobID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrVehicleSetNotFound
		}
		return nil, fmt.Errorf("read vehicle set image for deletion: %w", err)
	}
	now := time.Now().UTC().Format(time.RFC3339)
	if _, err := tx.ExecContext(ctx, `
UPDATE vehicle_sets
SET set_image_file_name='', set_image_mime_type='', set_image_blob_id=NULL,
    set_image_thumbnail_blob_id=NULL, set_image_created_at='', set_image_updated_at='',
    main_image_mode=CASE WHEN main_image_mode='dedicated' THEN 'automatic' ELSE main_image_mode END,
    updated_at=?
WHERE id=?
`, now, setID); err != nil {
		return nil, fmt.Errorf("delete vehicle set image: %w", err)
	}
	if err := auditVehicleSetImageTx(ctx, tx, actorUserID, setID, "VehicleSetImageDeleted", now); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit vehicle set image deletion: %w", err)
	}
	return distinctNonEmpty(blobID, thumbnailBlobID), nil
}

func (s *VehicleService) GetSetImage(ctx context.Context, setID string) (*VehicleSetImage, error) {
	var image VehicleSetImage
	err := s.db.QueryRowContext(ctx, `
SELECT COALESCE(set_image_file_name, ''), COALESCE(set_image_mime_type, ''), COALESCE(set_image_blob_id, ''),
       COALESCE(set_image_thumbnail_blob_id, ''), COALESCE(set_image_created_at, ''), COALESCE(set_image_updated_at, '')
FROM vehicle_sets WHERE id=?
`, strings.TrimSpace(setID)).Scan(
		&image.FileName, &image.MimeType, &image.BlobID, &image.ThumbnailBlobID, &image.CreatedAt, &image.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrVehicleSetNotFound
		}
		return nil, fmt.Errorf("get vehicle set image: %w", err)
	}
	if image.BlobID == "" {
		return nil, ErrVehicleSetImageNotFound
	}
	image.URL, image.ThumbnailURL = vehicleSetImageURLs(strings.TrimSpace(setID), image.BlobID)
	return &image, nil
}

func (s *VehicleService) attachSetMainImage(ctx context.Context, set *VehicleSet) error {
	mode, memberImageID, dedicated, resolved, err := s.resolveSetMainImage(ctx, set.ID)
	if err != nil {
		return err
	}
	set.MainImageMode = mode
	set.SelectedMemberImageID = memberImageID
	set.DedicatedImage = dedicated
	set.MainImage = resolved
	return nil
}

func (s *VehicleService) attachVehicleSetMainImages(ctx context.Context, vehicles []Vehicle) error {
	setIDs := make([]string, 0)
	seen := map[string]struct{}{}
	for index := range vehicles {
		if vehicles[index].VehicleSet == nil {
			continue
		}
		setID := vehicles[index].VehicleSet.ID
		if _, exists := seen[setID]; !exists {
			seen[setID] = struct{}{}
			setIDs = append(setIDs, setID)
		}
	}
	images, err := s.loadSetMainImages(ctx, setIDs)
	if err != nil {
		return err
	}
	for index := range vehicles {
		if vehicles[index].VehicleSet != nil {
			vehicles[index].VehicleSet.MainImage = images[vehicles[index].VehicleSet.ID]
		}
	}
	return nil
}

func (s *VehicleService) loadSetMainImages(
	ctx context.Context,
	setIDs []string,
) (map[string]*VehicleSetMainImage, error) {
	const batchSize = 400
	result := make(map[string]*VehicleSetMainImage, len(setIDs))
	for start := 0; start < len(setIDs); start += batchSize {
		end := min(start+batchSize, len(setIDs))
		values := strings.TrimSuffix(strings.Repeat("(?),", end-start), ",")
		rows, err := s.db.QueryContext(ctx, `
WITH requested(id) AS (VALUES `+values+`),
ranked_images AS (
    SELECT member.vehicle_set_id, image.id, image.vehicle_id, image.url,
           COALESCE(image.thumbnail_blob_id, '') AS thumbnail_blob_id,
           COALESCE(image.title, '') AS title, COALESCE(image.file_name, '') AS file_name,
           COALESCE(image.mime_type, '') AS mime_type, COALESCE(image.blob_id, '') AS blob_id,
           image.is_primary, image.sort_order, image.created_at,
           ROW_NUMBER() OVER (
               PARTITION BY member.vehicle_set_id
               ORDER BY CASE
                            WHEN sets.main_image_mode='member' AND image.id=sets.main_member_image_id THEN 0
                            ELSE 1
                        END,
                        member.position ASC, image.is_primary DESC, image.sort_order ASC, image.created_at ASC
           ) AS image_rank
    FROM requested
    JOIN vehicle_set_members member ON member.vehicle_set_id=requested.id
    JOIN vehicle_sets sets ON sets.id=member.vehicle_set_id
    JOIN vehicle_images image ON image.vehicle_id=member.vehicle_id
)
SELECT sets.id, sets.main_image_mode, COALESCE(sets.main_member_image_id, ''),
       COALESCE(sets.set_image_blob_id, ''), COALESCE(sets.set_image_updated_at, ''),
       ranked_images.id, ranked_images.vehicle_id, ranked_images.url, ranked_images.thumbnail_blob_id,
       ranked_images.title, ranked_images.file_name, ranked_images.mime_type, ranked_images.blob_id,
       ranked_images.is_primary, ranked_images.sort_order, ranked_images.created_at
FROM requested
JOIN vehicle_sets sets ON sets.id=requested.id
LEFT JOIN ranked_images ON ranked_images.vehicle_set_id=sets.id AND ranked_images.image_rank=1
`, stringSliceToAny(setIDs[start:end])...)
		if err != nil {
			return nil, fmt.Errorf("load vehicle set main images: %w", err)
		}
		for rows.Next() {
			var setID, memberImageID, dedicatedBlobID, dedicatedUpdatedAt string
			var mode VehicleSetMainImageMode
			var imageID, vehicleID, imageURL, thumbnailBlobID, title, fileName, mimeType, blobID sql.NullString
			var isPrimary, sortOrder sql.NullInt64
			var createdAt sql.NullString
			if err := rows.Scan(&setID, &mode, &memberImageID, &dedicatedBlobID, &dedicatedUpdatedAt,
				&imageID, &vehicleID, &imageURL, &thumbnailBlobID, &title, &fileName, &mimeType, &blobID,
				&isPrimary, &sortOrder, &createdAt); err != nil {
				_ = rows.Close()
				return nil, fmt.Errorf("scan vehicle set main image: %w", err)
			}
			if mode == VehicleSetMainImageModeDedicated && dedicatedBlobID != "" {
				fileURL, thumbnailURL := vehicleSetImageURLs(setID, dedicatedBlobID)
				result[setID] = &VehicleSetMainImage{
					Source: "dedicated", URL: fileURL, ThumbnailURL: thumbnailURL,
				}
				continue
			}
			if !imageID.Valid {
				continue
			}
			image := VehicleImage{
				ID: imageID.String, VehicleID: vehicleID.String, URL: imageURL.String,
				ThumbnailBlobID: thumbnailBlobID.String, Title: title.String, FileName: fileName.String,
				MimeType: mimeType.String, BlobID: blobID.String, IsPrimary: isPrimary.Int64 == 1,
				SortOrder: int(sortOrder.Int64), CreatedAt: createdAt.String,
			}
			image = withVehicleImageURLs(image)
			source := "automatic"
			if mode == VehicleSetMainImageModeMember && image.ID == memberImageID {
				source = "member"
			}
			result[setID] = mainImageFromVehicleImage(source, &image)
		}
		if err := rows.Close(); err != nil {
			return nil, fmt.Errorf("close vehicle set main image rows: %w", err)
		}
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("iterate vehicle set main images: %w", err)
		}
	}
	return result, nil
}

func (s *VehicleService) resolveSetMainImage(
	ctx context.Context,
	setID string,
) (VehicleSetMainImageMode, string, *VehicleSetImage, *VehicleSetMainImage, error) {
	var mode VehicleSetMainImageMode
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
			return "", "", nil, nil, ErrVehicleSetNotFound
		}
		return "", "", nil, nil, fmt.Errorf("read vehicle set image selection: %w", err)
	}
	var dedicated *VehicleSetImage
	if blobID != "" {
		fileURL, thumbnailURL := vehicleSetImageURLs(setID, blobID)
		dedicated = &VehicleSetImage{
			URL:          fileURL,
			ThumbnailURL: thumbnailURL,
			FileName:     fileName, MimeType: mimeType, BlobID: blobID, ThumbnailBlobID: thumbnailBlobID,
			CreatedAt: createdAt, UpdatedAt: updatedAt,
		}
	}
	if mode == VehicleSetMainImageModeDedicated && dedicated != nil {
		return mode, memberImageID, dedicated, &VehicleSetMainImage{
			Source: "dedicated", URL: dedicated.URL, ThumbnailURL: dedicated.ThumbnailURL,
		}, nil
	}
	if mode == VehicleSetMainImageModeMember && memberImageID != "" {
		image, err := s.loadSetMemberImage(ctx, setID, memberImageID)
		if err == nil {
			return mode, memberImageID, dedicated, mainImageFromVehicleImage("member", image), nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return "", "", nil, nil, err
		}
	}
	automatic, err := s.loadAutomaticSetImage(ctx, setID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", "", nil, nil, err
	}
	if errors.Is(err, sql.ErrNoRows) {
		return mode, memberImageID, dedicated, nil, nil
	}
	return mode, memberImageID, dedicated, mainImageFromVehicleImage("automatic", automatic), nil
}

func (s *VehicleService) loadSetMemberImage(ctx context.Context, setID, imageID string) (*VehicleImage, error) {
	return s.scanSetMemberImage(s.db.QueryRowContext(ctx, `
SELECT image.id, image.vehicle_id, image.url, COALESCE(image.thumbnail_blob_id, ''),
       COALESCE(image.title, ''), COALESCE(image.file_name, ''), COALESCE(image.mime_type, ''),
       COALESCE(image.blob_id, ''), image.is_primary, image.sort_order, image.created_at
FROM vehicle_images image
JOIN vehicle_set_members member ON member.vehicle_id=image.vehicle_id
WHERE member.vehicle_set_id=? AND image.id=?
`, setID, imageID))
}

func (s *VehicleService) loadAutomaticSetImage(ctx context.Context, setID string) (*VehicleImage, error) {
	return s.scanSetMemberImage(s.db.QueryRowContext(ctx, `
SELECT image.id, image.vehicle_id, image.url, COALESCE(image.thumbnail_blob_id, ''),
       COALESCE(image.title, ''), COALESCE(image.file_name, ''), COALESCE(image.mime_type, ''),
       COALESCE(image.blob_id, ''), image.is_primary, image.sort_order, image.created_at
FROM vehicle_set_members member
JOIN vehicle_images image ON image.vehicle_id=member.vehicle_id
WHERE member.vehicle_set_id=?
ORDER BY member.position ASC, image.is_primary DESC, image.sort_order ASC, image.created_at ASC
LIMIT 1
`, setID))
}

type rowScanner interface {
	Scan(dest ...any) error
}

func (s *VehicleService) scanSetMemberImage(row rowScanner) (*VehicleImage, error) {
	var image VehicleImage
	var thumbnailBlobID string
	var isPrimary int
	if err := row.Scan(&image.ID, &image.VehicleID, &image.URL, &thumbnailBlobID, &image.Title,
		&image.FileName, &image.MimeType, &image.BlobID, &isPrimary, &image.SortOrder, &image.CreatedAt); err != nil {
		return nil, err
	}
	image.ThumbnailBlobID = thumbnailBlobID
	image.IsPrimary = isPrimary == 1
	image = withVehicleImageURLs(image)
	return &image, nil
}

func mainImageFromVehicleImage(source string, image *VehicleImage) *VehicleSetMainImage {
	return &VehicleSetMainImage{
		Source: source, ImageID: image.ID, VehicleID: image.VehicleID, URL: image.URL,
		ThumbnailURL: image.ThumbnailURL, Title: image.Title,
	}
}

func auditVehicleSetImageTx(
	ctx context.Context,
	tx *sql.Tx,
	actorUserID, setID, action, now string,
) error {
	if _, err := tx.ExecContext(ctx, `
INSERT INTO audit_logs(id, actor_user_id, action, target_type, target_id, created_at, details_json)
VALUES(?, ?, ?, 'vehicle_set', ?, ?, '{}')
`, randomID(), actorUserID, action, setID, now); err != nil {
		return fmt.Errorf("write vehicle set image audit log: %w", err)
	}
	return nil
}

func vehicleSetImageURLs(setID, version string) (string, string) {
	suffix := ""
	if version = strings.TrimSpace(version); version != "" {
		suffix = "?v=" + url.QueryEscape(version)
	}
	base := "/api/v1/vehicle-sets/" + setID + "/image/"
	return base + "file" + suffix, base + "thumbnail" + suffix
}

func stringSliceToAny(values []string) []any {
	result := make([]any, len(values))
	for index := range values {
		result[index] = values[index]
	}
	return result
}

func distinctNonEmpty(values ...string) []string {
	result := []string{}
	seen := map[string]struct{}{}
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, found := seen[value]; found {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
