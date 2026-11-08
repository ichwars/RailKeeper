package application

import (
	"context"

	"railkeeper/backend/internal/infrastructure/vehicledata"
	vehiclemodel "railkeeper/backend/internal/inventory/vehicle"
)

const maxVehicleSetMembers = vehiclemodel.MaxVehicleSetMembers

func (s *VehicleService) setService() *vehiclemodel.SetService {
	return vehiclemodel.NewSetService(vehicledata.NewSetStore(s.db), vehicleSetEnvironment{vehicleCoreEnvironment{service: s}})
}
func (s *VehicleService) CreateSet(ctx context.Context, input CreateVehicleSetInput, actor string) (*VehicleSet, error) {
	return s.setService().CreateSet(ctx, input, actor)
}
func (s *VehicleService) GetSet(ctx context.Context, id string) (*VehicleSet, error) {
	return s.setService().GetSet(ctx, id)
}
func (s *VehicleService) UpdateSet(ctx context.Context, id string, input VehicleSetInput, actor string) (*VehicleSet, error) {
	return s.setService().UpdateSet(ctx, id, input, actor)
}

type vehicleSetEnvironment struct{ vehicleCoreEnvironment }

func (e vehicleSetEnvironment) LoadMembers(ctx context.Context, id string) ([]Vehicle, error) {
	if err := e.service.resetExpiredExhibitionFlags(ctx); err != nil {
		return nil, err
	}
	return e.service.list(ctx, "", id)
}
func (e vehicleSetEnvironment) AttachMainImage(ctx context.Context, set *VehicleSet) error {
	return e.service.attachSetMainImage(ctx, set)
}
