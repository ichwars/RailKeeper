package application

import (
	"context"

	"railkeeper/backend/internal/infrastructure/vehicledata"
	vehiclemodel "railkeeper/backend/internal/inventory/vehicle"
)

func (s *VehicleService) Delete(ctx context.Context, id, actor string) error {
	return vehiclemodel.Delete(ctx, vehicledata.NewDeleteStore(s.db), id, actor)
}

func readVehicleCore(ctx context.Context, query queryRower, id string) (*Vehicle, error) {
	return vehicledata.ReadCore(ctx, query, id)
}

func (s *VehicleService) get(ctx context.Context, id string) (*Vehicle, error) {
	vehicle, err := readVehicleCore(ctx, s.db, id)
	if err != nil {
		return nil, err
	}

	images, err := s.loadVehicleImages(ctx, id)
	if err != nil {
		return nil, err
	}
	vehicle.Images = images
	vehicle.ImagesRevision = vehicleImagesRevision(images)
	if vehicle.VehicleSet != nil {
		_, _, _, mainImage, err := s.resolveSetMainImage(ctx, vehicle.VehicleSet.ID)
		if err != nil {
			return nil, err
		}
		vehicle.VehicleSet.MainImage = mainImage
	}
	attachments, err := s.loadVehicleAttachments(ctx, id)
	if err != nil {
		return nil, err
	}
	vehicle.Attachments = attachments
	maintenance, err := s.loadVehicleMaintenance(ctx, id)
	if err != nil {
		return nil, err
	}
	vehicle.Maintenance = maintenance
	spareParts, err := s.loadVehicleSpareParts(ctx, id)
	if err != nil {
		return nil, err
	}
	vehicle.SpareParts = spareParts
	functions, err := s.loadVehicleFunctions(ctx, id)
	if err != nil {
		return nil, err
	}
	vehicle.Functions = functions
	cvValues, err := s.loadVehicleCVValues(ctx, id)
	if err != nil {
		return nil, err
	}
	vehicle.CVValues = cvValues
	cvFiles, err := s.loadVehicleCVFiles(ctx, id)
	if err != nil {
		return nil, err
	}
	vehicle.CVFiles = cvFiles

	return vehicle, nil
}
