package application

import (
	"context"
	"database/sql"

	"railkeeper/backend/internal/infrastructure/vehicledata"
	vehiclemodel "railkeeper/backend/internal/inventory/vehicle"
)

func (s *VehicleService) attachImages(ctx context.Context, vehicles []Vehicle) error {
	for index := range vehicles {
		images, err := s.loadVehicleImages(ctx, vehicles[index].ID)
		if err != nil {
			return err
		}
		vehicles[index].Images = images
		vehicles[index].ImagesRevision = vehicleImagesRevision(images)
	}
	return nil
}

func (s *VehicleService) loadVehicleImages(ctx context.Context, vehicleID string) ([]VehicleImage, error) {
	return readVehicleImages(ctx, s.db, vehicleID)
}

type vehicleMediaQuery interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func readVehicleImages(ctx context.Context, query vehicleMediaQuery, vehicleID string) ([]VehicleImage, error) {
	return vehicledata.ReadImages(ctx, query, vehicleID)
}

func (s *VehicleService) attachAttachments(ctx context.Context, vehicles []Vehicle) error {
	for index := range vehicles {
		attachments, err := s.loadVehicleAttachments(ctx, vehicles[index].ID)
		if err != nil {
			return err
		}
		vehicles[index].Attachments = attachments
	}
	return nil
}

func (s *VehicleService) attachMaintenance(ctx context.Context, vehicles []Vehicle) error {
	for index := range vehicles {
		maintenance, err := s.loadVehicleMaintenance(ctx, vehicles[index].ID)
		if err != nil {
			return err
		}
		vehicles[index].Maintenance = maintenance
	}
	return nil
}

func (s *VehicleService) attachSpareParts(ctx context.Context, vehicles []Vehicle) error {
	for index := range vehicles {
		spareParts, err := s.loadVehicleSpareParts(ctx, vehicles[index].ID)
		if err != nil {
			return err
		}
		vehicles[index].SpareParts = spareParts
	}
	return nil
}

func (s *VehicleService) attachFunctions(ctx context.Context, vehicles []Vehicle) error {
	for index := range vehicles {
		functions, err := s.loadVehicleFunctions(ctx, vehicles[index].ID)
		if err != nil {
			return err
		}
		vehicles[index].Functions = functions
	}
	return nil
}

func (s *VehicleService) attachCVData(ctx context.Context, vehicles []Vehicle) error {
	for index := range vehicles {
		values, err := s.loadVehicleCVValues(ctx, vehicles[index].ID)
		if err != nil {
			return err
		}
		vehicles[index].CVValues = values
		files, err := s.loadVehicleCVFiles(ctx, vehicles[index].ID)
		if err != nil {
			return err
		}
		vehicles[index].CVFiles = files
	}
	return nil
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
