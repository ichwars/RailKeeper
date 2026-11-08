package application

import (
	"context"
	"railkeeper/backend/internal/infrastructure/vehicledata"
	vehiclemodel "railkeeper/backend/internal/inventory/vehicle"
)

func (s *VehicleService) SetSetMainImage(
	ctx context.Context,
	setID string,
	input VehicleSetMainImageInput,
	actorUserID string,
) (*VehicleSet, error) {
	return s.setImageService().SetSetMainImage(ctx, setID, input, actorUserID)
}

func (s *VehicleService) UpsertSetImage(
	ctx context.Context,
	setID string,
	input VehicleSetImageInput,
	actorUserID string,
) (*VehicleSet, []string, error) {
	return s.setImageService().UpsertSetImage(ctx, setID, input, actorUserID)
}

func (s *VehicleService) DeleteSetImage(ctx context.Context, setID, actorUserID string) ([]string, error) {
	return s.setImageService().DeleteSetImage(ctx, setID, actorUserID)
}

func (s *VehicleService) GetSetImage(ctx context.Context, setID string) (*VehicleSetImage, error) {
	return s.setImageService().GetSetImage(ctx, setID)
}

func (s *VehicleService) attachSetMainImage(ctx context.Context, set *VehicleSet) error {
	mode, memberImageID, dedicated, resolved, err := s.resolveSetMainImage(ctx, set.ID)
	if err != nil {
		return err
	}
	set.MainImageMode = mode
	set.SelectedMemberImageID = memberImageID
	set.DedicatedImage = dedicated
	set.MainImage = resolved
	return nil
}
func (s *VehicleService) attachVehicleSetMainImages(ctx context.Context, vehicles []Vehicle) error {
	setIDs := make([]string, 0)
	seen := map[string]struct{}{}
	for index := range vehicles {
		if vehicles[index].VehicleSet == nil {
			continue
		}
		setID := vehicles[index].VehicleSet.ID
		if _, exists := seen[setID]; !exists {
			seen[setID] = struct{}{}
			setIDs = append(setIDs, setID)
		}
	}
	images, err := s.loadSetMainImages(ctx, setIDs)
	if err != nil {
		return err
	}
	for index := range vehicles {
		if vehicles[index].VehicleSet != nil {
			vehicles[index].VehicleSet.MainImage = images[vehicles[index].VehicleSet.ID]
		}
	}
	return nil
}
func (s *VehicleService) loadSetMainImages(
	ctx context.Context,
	setIDs []string,
) (map[string]*VehicleSetMainImage, error) {
	return s.setImageService().LoadMainImages(ctx, setIDs)
}

func (s *VehicleService) resolveSetMainImage(
	ctx context.Context,
	setID string,
) (VehicleSetMainImageMode, string, *VehicleSetImage, *VehicleSetMainImage, error) {
	return s.setImageService().Resolve(ctx, setID)
}

func (s *VehicleService) setImageService() *vehiclemodel.SetImageService {
	return vehiclemodel.NewSetImageService(vehicledata.NewSetImageStore(s.db), s.GetSet)
}
