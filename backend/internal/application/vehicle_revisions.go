package application

import (
	"context"
	"time"

	vehiclemodel "railkeeper/backend/internal/inventory/vehicle"
)

var (
	ErrVehicleCVFileConflict     = vehiclemodel.ErrVehicleCVFileConflict
	ErrVehicleConflict           = vehiclemodel.ErrVehicleConflict
	ErrVehicleImageConflict      = vehiclemodel.ErrVehicleImageConflict
	ErrVehicleAttachmentConflict = vehiclemodel.ErrVehicleAttachmentConflict
)

func vehicleImagesRevision(images []VehicleImage) string { return vehiclemodel.ImagesRevision(images) }

func checkVehicleRevision(actual string, expected *string, conflict error) error {
	return vehiclemodel.CheckRevision(actual, expected, conflict)
}

func (s *VehicleService) cleanupLocalizedVehicleImages(ctx context.Context, images []VehicleImageInput) {
	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	blobs := NewFileBlobService(s.db, "")
	for _, image := range images {
		_ = blobs.DeleteIfUnreferenced(cleanupCtx, image.BlobID)
		_ = blobs.DeleteIfUnreferenced(cleanupCtx, image.ThumbnailBlobID)
	}
}
