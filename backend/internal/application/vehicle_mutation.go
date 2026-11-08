package application

import (
	"context"

	"railkeeper/backend/internal/infrastructure/vehicledata"
	vehiclemodel "railkeeper/backend/internal/inventory/vehicle"
)

func (s *VehicleService) coreVehicleService() *vehiclemodel.CoreService {
	return vehiclemodel.NewCoreService(vehicledata.NewCoreStore(s.db), vehicleCoreEnvironment{service: s})
}
func (s *VehicleService) Create(ctx context.Context, input CreateVehicleInput, actor string) (*Vehicle, error) {
	return s.coreVehicleService().Create(ctx, input, actor)
}
func (s *VehicleService) CreateWithExternalMapping(ctx context.Context, input CreateVehicleInput, mapping VehicleExternalMapInput, actor string) (*Vehicle, error) {
	return s.coreVehicleService().CreateWithExternalMapping(ctx, input, mapping, actor)
}
func (s *VehicleService) CreateWithOptions(ctx context.Context, input CreateVehicleInput, options VehicleCreationOptions, actor string) (*Vehicle, error) {
	return s.coreVehicleService().CreateWithOptions(ctx, input, options, actor)
}
func (s *VehicleService) Update(ctx context.Context, id string, input CreateVehicleInput, actor string) (*Vehicle, error) {
	return s.coreVehicleService().Update(ctx, id, input, actor)
}

type vehicleCoreEnvironment struct{ service *VehicleService }

func (e vehicleCoreEnvironment) Existing(ctx context.Context, id string) (*Vehicle, error) {
	return e.service.get(ctx, id)
}
func (e vehicleCoreEnvironment) LocalizeImages(ctx context.Context, id string, images []VehicleImageInput) ([]VehicleImageInput, error) {
	if e.service.imageLocalizer == nil {
		return images, nil
	}
	return e.service.imageLocalizer(ctx, id, images)
}
func (e vehicleCoreEnvironment) CleanupImages(ctx context.Context, images []VehicleImageInput) {
	e.service.cleanupLocalizedVehicleImages(ctx, images)
}
func (e vehicleCoreEnvironment) CompleteCreated(ctx context.Context, vehicle *Vehicle) error {
	functions, err := e.service.loadVehicleFunctions(ctx, vehicle.ID)
	if err != nil {
		return err
	}
	vehicle.Functions = functions
	return nil
}
func (e vehicleCoreEnvironment) CompleteUpdated(ctx context.Context, vehicle *Vehicle) error {
	s := e.service
	if vehicle.VehicleSet != nil {
		_, _, _, mainImage, err := s.resolveSetMainImage(ctx, vehicle.VehicleSet.ID)
		if err != nil {
			return err
		}
		vehicle.VehicleSet.MainImage = mainImage
	}
	attachments, err := s.loadVehicleAttachments(ctx, vehicle.ID)
	if err != nil {
		return err
	}
	vehicle.Attachments = attachments
	maintenance, err := s.loadVehicleMaintenance(ctx, vehicle.ID)
	if err != nil {
		return err
	}
	vehicle.Maintenance = maintenance
	functions, err := s.loadVehicleFunctions(ctx, vehicle.ID)
	if err != nil {
		return err
	}
	vehicle.Functions = functions
	cvValues, err := s.loadVehicleCVValues(ctx, vehicle.ID)
	if err != nil {
		return err
	}
	vehicle.CVValues = cvValues
	cvFiles, err := s.loadVehicleCVFiles(ctx, vehicle.ID)
	if err != nil {
		return err
	}
	vehicle.CVFiles = cvFiles
	externalMappings, err := s.loadVehicleExternalMappings(ctx, vehicle.ID)
	if err != nil {
		return err
	}
	vehicle.ExternalMappings = externalMappings
	vehicle.ECoSDecoderProfile, err = s.loadECoSDecoderProfile(ctx, vehicle.ID)
	if err != nil {
		return err
	}
	return nil
}
