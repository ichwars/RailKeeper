package application

import (
	"context"
	"database/sql"
	"strings"

	"railkeeper/backend/internal/infrastructure/vehicledata"
)

func saveVehicleImages(ctx context.Context, tx *sql.Tx, vehicleID string, images []VehicleImageInput, now string) error {
	existing, err := existingVehicleImageMeta(ctx, tx, vehicleID)
	if err != nil {
		return err
	}
	if err := vehicledata.ClearImages(ctx, tx, vehicleID); err != nil {
		return err
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
		// Zero is a stored order too. Keep it when this image stayed in the same position.
		keepsZero := (hasMeta || meta.ID != "") && meta.SortOrder == 0 && meta.Position == index
		if sortOrder == 0 && !keepsZero {
			sortOrder = index
		}
		imageURL := image.URL
		if image.StoragePath != "" || image.BlobID != "" {
			imageURL = "/api/v1/vehicles/" + vehicleID + "/images/" + imageID + "/file"
		}
		if err := vehicledata.InsertImage(ctx, tx, vehicledata.ImageRecord{
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

type queryRower interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func ensureVehicleMaintenanceID(ctx context.Context, query queryRower, vehicleID, maintenanceID string) error {
	maintenanceID = strings.TrimSpace(maintenanceID)
	if maintenanceID == "" {
		return nil
	}
	exists, err := vehicledata.MaintenanceBelongsToVehicle(ctx, query, vehicleID, maintenanceID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrVehicleValidation
	}
	return nil
}

type vehicleImageMeta = vehicledata.ImageMetadata

func existingVehicleImageMeta(ctx context.Context, tx *sql.Tx, vehicleID string) (map[string]vehicleImageMeta, error) {
	return vehicledata.ExistingImages(ctx, tx, vehicleID)
}

func (s *VehicleService) nextInventoryNumber(ctx context.Context, tx *sql.Tx, vehicleCategory string) (string, error) {
	requestedCategory := cleanInventoryCategory(vehicleCategory)
	fallbackCategory := inventoryCategoryForVehicle(vehicleCategory)
	return ReserveInventoryNumber(ctx, tx, requestedCategory, fallbackCategory, func(candidate string) error {
		return s.ensureInventoryNumberAvailable(ctx, tx, candidate, "")
	})
}

func (s *VehicleService) ensureInventoryNumberAvailable(ctx context.Context, tx *sql.Tx, inventoryNumber, excludeVehicleID string) error {
	inventoryNumber = strings.TrimSpace(inventoryNumber)
	if inventoryNumber == "" {
		return ErrInventoryNumberValidation
	}
	used, err := vehicledata.InventoryNumberInUse(ctx, tx, inventoryNumber, excludeVehicleID)
	if err != nil {
		return err
	}
	if used {
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
