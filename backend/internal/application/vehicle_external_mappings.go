package application

import (
	"context"
	"railkeeper/backend/internal/infrastructure/vehicledata"
	vehiclemodel "railkeeper/backend/internal/inventory/vehicle"
)

func (s *VehicleService) externalMappingService() *vehiclemodel.ExternalMappingService {
	return vehiclemodel.NewExternalMappingService(vehicledata.NewExternalMappingStore(s.db), func(ctx context.Context, id string) error { _, err := s.get(ctx, id); return err })
}
func (s *VehicleService) UpsertExternalMapping(ctx context.Context, id string, input VehicleExternalMapInput, actor string) (*VehicleExternalMap, error) {
	return s.externalMappingService().UpsertExternalMapping(ctx, id, input, actor)
}
func (s *VehicleService) RebindExternalMapping(ctx context.Context, id, previous string, input VehicleExternalMapInput, actor string) (*VehicleExternalMap, error) {
	return s.externalMappingService().RebindExternalMapping(ctx, id, previous, input, actor)
}
func cleanVehicleExternalMapInput(input VehicleExternalMapInput) VehicleExternalMapInput {
	return vehiclemodel.CleanExternalMappingInput(input)
}
func (s *VehicleService) attachExternalMappings(ctx context.Context, vehicles []Vehicle) error {
	for index := range vehicles {
		mappings, err := s.loadVehicleExternalMappings(ctx, vehicles[index].ID)
		if err != nil {
			return err
		}
		vehicles[index].ExternalMappings = mappings
	}
	return nil
}

func (s *VehicleService) loadVehicleExternalMappings(ctx context.Context, id string) ([]VehicleExternalMap, error) {
	return vehicledata.NewExternalMappingStore(s.db).Mappings(ctx, id)
}
