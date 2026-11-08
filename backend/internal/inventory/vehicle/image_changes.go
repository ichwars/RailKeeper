package vehicle

import (
	"context"
	"fmt"
	"strings"
	"time"

	"railkeeper/backend/internal/credentials"
)

func (s *ImageService) CreateImage(ctx context.Context, vehicleID string, input VehicleImageInput) (*VehicleImage, error) {
	created, err := s.CreateImageWithRevision(ctx, vehicleID, input, nil)
	if err != nil {
		return nil, err
	}
	return &created.VehicleImage, nil
}
func (s *ImageService) CreateImageWithRevision(ctx context.Context, vehicleID string, input VehicleImageInput, expectedRevision *string) (*VehicleImageCreation, error) {
	vehicleID = strings.TrimSpace(vehicleID)
	input = CleanVehicleImageInput(input)
	if vehicleID == "" || input.FileName == "" || input.MimeType == "" || (input.StoragePath == "" && input.BlobID == "") {
		return nil, ErrVehicleValidation
	}
	if err := s.parent(ctx, vehicleID); err != nil {
		return nil, err
	}
	if err := EnsureMaintenanceReference(ctx, s.repository, vehicleID, input.MaintenanceID); err != nil {
		return nil, err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	imageID := credentials.RandomID()
	image := VehicleImage{
		ID:              imageID,
		VehicleID:       vehicleID,
		URL:             "/api/v1/vehicles/" + vehicleID + "/images/" + imageID + "/file",
		ThumbnailURL:    "/api/v1/vehicles/" + vehicleID + "/images/" + imageID + "/thumbnail",
		Title:           input.Title,
		SourceURL:       input.SourceURL,
		FileName:        input.FileName,
		MimeType:        input.MimeType,
		StoragePath:     input.StoragePath,
		ThumbnailPath:   input.ThumbnailPath,
		BlobID:          input.BlobID,
		ThumbnailBlobID: input.ThumbnailBlobID,
		MaintenanceID:   input.MaintenanceID,
		IsPrimary:       input.IsPrimary,
		SortOrder:       input.SortOrder,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	tx, err := s.repository.BeginImages(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin vehicle image create: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := EnsureMaintenanceReference(ctx, tx, vehicleID, input.MaintenanceID); err != nil {
		return nil, err
	}
	if expectedRevision != nil {
		images, err := tx.Images(ctx, vehicleID)
		if err != nil {
			return nil, err
		}
		if err := CheckRevision(ImagesRevision(images), expectedRevision, ErrVehicleImageConflict); err != nil {
			return nil, err
		}
	}
	existingCount, err := tx.CountImages(ctx, vehicleID)
	if err != nil {
		return nil, err
	}
	if existingCount == 0 {
		image.IsPrimary = true
	}
	if image.IsPrimary {
		if err := tx.ClearPrimary(ctx, vehicleID, now); err != nil {
			return nil, err
		}
	}
	if image.SortOrder == 0 {
		image.SortOrder = existingCount
	}
	if err := tx.AppendImage(ctx, ImageWriteRecord(image)); err != nil {
		return nil, err
	}
	images, err := tx.Images(ctx, vehicleID)
	if err != nil {
		return nil, err
	}
	imagesRevision := ImagesRevision(images)
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit vehicle image create: %w", err)
	}
	image.Revision = ImageRevision(image)
	return &VehicleImageCreation{VehicleImage: image, ImagesRevision: imagesRevision}, nil
}
func (s *ImageService) DeleteImage(ctx context.Context, vehicleID, imageID string, expectedRevision ...string) (*VehicleImage, error) {
	tx, err := s.repository.BeginImages(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin vehicle image delete: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	image, err := tx.Image(ctx, vehicleID, imageID)
	if err != nil {
		return nil, err
	}
	if err := CheckRevision(image.Revision, OptionalRevision(expectedRevision), ErrVehicleImageConflict); err != nil {
		return nil, err
	}
	if image.MaintenanceID != "" {
		return nil, ErrVehicleImageInUse
	}

	if err := tx.DeleteImage(ctx, vehicleID, imageID); err != nil {
		return nil, err
	}
	if image.IsPrimary {
		now := time.Now().UTC().Format(time.RFC3339)
		if err := tx.PromotePrimary(ctx, vehicleID, now); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit vehicle image delete: %w", err)
	}
	return image, nil
}
func (s *ImageService) GetImage(ctx context.Context, vehicleID, imageID string) (*VehicleImage, error) {
	return s.repository.Image(ctx, vehicleID, imageID)
}
