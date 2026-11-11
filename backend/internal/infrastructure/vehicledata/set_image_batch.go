package vehicledata

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	vehiclemodel "railkeeper/backend/internal/inventory/vehicle"
)

func (s *setImageStore) Candidates(
	ctx context.Context,
	setIDs []string,
) ([]vehiclemodel.SetImageCandidate, error) {
	result := []vehiclemodel.SetImageCandidate{}
	if len(setIDs) == 0 {
		return result, nil
	}
	ids, err := json.Marshal(setIDs)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
WITH requested(id) AS (SELECT value FROM json_each(?)),
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
       COALESCE(sets.set_image_blob_id, ''),
       ranked_images.id, ranked_images.vehicle_id, ranked_images.url, ranked_images.thumbnail_blob_id,
       ranked_images.title, ranked_images.file_name, ranked_images.mime_type, ranked_images.blob_id,
       ranked_images.is_primary, ranked_images.sort_order, ranked_images.created_at
FROM requested
JOIN vehicle_sets sets ON sets.id=requested.id
LEFT JOIN ranked_images ON ranked_images.vehicle_set_id=sets.id AND ranked_images.image_rank=1
`, string(ids))
	if err != nil {
		return nil, fmt.Errorf("load vehicle set main images: %w", err)
	}
	for rows.Next() {
		var setID, memberImageID, dedicatedBlobID string
		var mode vehiclemodel.VehicleSetMainImageMode
		var imageID, vehicleID, imageURL, thumbnailBlobID, title, fileName, mimeType, blobID sql.NullString
		var isPrimary, sortOrder sql.NullInt64
		var createdAt sql.NullString
		if err := rows.Scan(&setID, &mode, &memberImageID, &dedicatedBlobID,
			&imageID, &vehicleID, &imageURL, &thumbnailBlobID, &title, &fileName, &mimeType, &blobID,
			&isPrimary, &sortOrder, &createdAt); err != nil {
			_ = rows.Close()
			return nil, fmt.Errorf("scan vehicle set main image: %w", err)
		}
		row := vehiclemodel.SetImageCandidate{SetID: setID, Mode: mode, MemberImageID: memberImageID, DedicatedBlobID: dedicatedBlobID}
		if !imageID.Valid {
			result = append(result, row)
			continue
		}
		image := vehiclemodel.VehicleImage{
			ID: imageID.String, VehicleID: vehicleID.String, URL: imageURL.String,
			ThumbnailBlobID: thumbnailBlobID.String, Title: title.String, FileName: fileName.String,
			MimeType: mimeType.String, BlobID: blobID.String, IsPrimary: isPrimary.Int64 == 1,
			SortOrder: int(sortOrder.Int64), CreatedAt: createdAt.String,
		}
		row.Image = &image
		result = append(result, row)
	}
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("close vehicle set main image rows: %w", err)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate vehicle set main images: %w", err)
	}
	return result, nil
}
