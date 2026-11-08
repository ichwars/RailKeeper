package vehicle

import (
	"context"
	"strings"

	"railkeeper/backend/internal/credentials"
)

func ReplaceImages(ctx context.Context, tx ImageReplacementMutation, vehicleID string, images []VehicleImageInput, now string) error {
	existing, err := tx.ExistingImages(ctx, vehicleID)
	if err != nil {
		return err
	}
	if err := tx.ClearImages(ctx, vehicleID); err != nil {
		return err
	}
	cleaned := CleanVehicleImageInputs(images)
	for index, image := range cleaned {
		image.MaintenanceID = strings.TrimSpace(image.MaintenanceID)
		if err := EnsureMaintenanceReference(ctx, tx, vehicleID, image.MaintenanceID); err != nil {
			return err
		}
		meta, hasMeta := existing[image.ID]
		if !hasMeta {
			meta = existing[image.URL]
		}
		imageID := credentials.RandomID()
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
		// Zero is a stored order too. Keep it when this image stayed in the same position.
		keepsZero := (hasMeta || meta.ID != "") && meta.SortOrder == 0 && meta.Position == index
		if sortOrder == 0 && !keepsZero {
			sortOrder = index
		}
		imageURL := image.URL
		if image.StoragePath != "" || image.BlobID != "" {
			imageURL = "/api/v1/vehicles/" + vehicleID + "/images/" + imageID + "/file"
		}
		if err := tx.InsertImage(ctx, ImageRecord{
			ID:              imageID,
			VehicleID:       vehicleID,
			URL:             imageURL,
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
			CreatedAt:       createdAt,
			UpdatedAt:       now,
		}); err != nil {
			return err
		}
	}
	return nil
}
