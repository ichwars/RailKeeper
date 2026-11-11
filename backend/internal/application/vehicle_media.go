package application

import (
	"context"
	"database/sql"

	"railkeeper/backend/internal/infrastructure/vehicledata"
	vehiclemodel "railkeeper/backend/internal/inventory/vehicle"
)

func (s *VehicleService) loadVehicleImages(ctx context.Context, vehicleID string) ([]VehicleImage, error) {
	return readVehicleImages(ctx, s.db, vehicleID)
}

type vehicleMediaQuery interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func readVehicleImages(ctx context.Context, query vehicleMediaQuery, vehicleID string) ([]VehicleImage, error) {
	return vehicledata.ReadImages(ctx, query, vehicleID)
}

func (s *VehicleService) CreateImage(ctx context.Context, vehicleID string, input VehicleImageInput) (*VehicleImage, error) {
	return s.imageService().CreateImage(ctx, vehicleID, input)
}

func (s *VehicleService) CreateImageWithRevision(ctx context.Context, vehicleID string, input VehicleImageInput, expectedRevision *string) (*VehicleImageCreation, error) {
	return s.imageService().CreateImageWithRevision(ctx, vehicleID, input, expectedRevision)
}

func (s *VehicleService) GetImage(ctx context.Context, vehicleID, imageID string) (*VehicleImage, error) {
	return s.imageService().GetImage(ctx, vehicleID, imageID)
}

func (s *VehicleService) DeleteImage(ctx context.Context, vehicleID, imageID string, expectedRevision ...string) (*VehicleImage, error) {
	return s.imageService().DeleteImage(ctx, vehicleID, imageID, expectedRevision...)
}

func (s *VehicleService) ImageFileReferenceCount(ctx context.Context, storagePath string) (int, error) {
	return s.MediaFileReferenceCount(ctx, storagePath)
}

func (s *VehicleService) MediaFileReferenceCount(ctx context.Context, storagePath string) (int, error) {
	return s.attachmentService().MediaFileReferenceCount(ctx, storagePath)
}

func (s *VehicleService) loadVehicleAttachments(ctx context.Context, vehicleID string) ([]VehicleAttachment, error) {
	return s.attachmentService().Attachments(ctx, vehicleID)
}

func (s *VehicleService) CreateAttachment(ctx context.Context, vehicleID string, input VehicleAttachmentInput) (*VehicleAttachment, error) {
	return s.attachmentService().CreateAttachment(ctx, vehicleID, input)
}

func (s *VehicleService) UpdateAttachment(ctx context.Context, vehicleID, attachmentID string, input VehicleAttachmentUpdateInput) (*VehicleAttachment, error) {
	return s.attachmentService().UpdateAttachment(ctx, vehicleID, attachmentID, input)
}

func (s *VehicleService) GetAttachment(ctx context.Context, vehicleID, attachmentID string) (*VehicleAttachment, error) {
	return s.attachmentService().GetAttachment(ctx, vehicleID, attachmentID)
}

func (s *VehicleService) DeleteAttachment(ctx context.Context, vehicleID, attachmentID string, expectedRevision ...string) (*VehicleAttachment, error) {
	return s.attachmentService().DeleteAttachment(ctx, vehicleID, attachmentID, expectedRevision...)
}

func (s *VehicleService) imageService() *vehiclemodel.ImageService {
	return vehiclemodel.NewImageService(vehicledata.NewImageStore(s.db), func(ctx context.Context, id string) error { _, err := s.Get(ctx, id); return err })
}

func (s *VehicleService) attachmentService() *vehiclemodel.AttachmentService {
	return vehiclemodel.NewAttachmentService(vehicledata.NewAttachmentStore(s.db), func(ctx context.Context, id string) error { _, err := s.Get(ctx, id); return err })
}
