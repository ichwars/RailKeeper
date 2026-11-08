package vehicle

import (
	"context"
	"strings"
)

type MaintenanceLookup interface {
	MaintenanceBelongs(context.Context, string, string) (bool, error)
}
type ImageReader interface {
	Images(context.Context, string) ([]VehicleImage, error)
	Image(context.Context, string, string) (*VehicleImage, error)
}
type ImageReplacementMutation interface {
	MaintenanceLookup
	ExistingImages(context.Context, string) (map[string]ImageMetadata, error)
	ClearImages(context.Context, string) error
	InsertImage(context.Context, ImageRecord) error
}
type ImageMutation interface {
	ImageReader
	ImageReplacementMutation
	CountImages(context.Context, string) (int, error)
	ClearPrimary(context.Context, string, string) error
	AppendImage(context.Context, ImageRecord) error
	DeleteImage(context.Context, string, string) error
	PromotePrimary(context.Context, string, string) error
	Commit() error
	Rollback() error
}
type ImageRepository interface {
	ImageReader
	MaintenanceLookup
	BeginImages(context.Context) (ImageMutation, error)
}
type ImageService struct {
	repository ImageRepository
	parent     func(context.Context, string) error
}

func NewImageService(repository ImageRepository, parent func(context.Context, string) error) *ImageService {
	return &ImageService{repository: repository, parent: parent}
}
func EnsureMaintenanceReference(ctx context.Context, lookup MaintenanceLookup, vehicleID, maintenanceID string) error {
	maintenanceID = strings.TrimSpace(maintenanceID)
	if maintenanceID == "" {
		return nil
	}
	exists, err := lookup.MaintenanceBelongs(ctx, vehicleID, maintenanceID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrVehicleValidation
	}
	return nil
}
