package application

import (
	"context"
	"railkeeper/backend/internal/infrastructure/vehicledata"
	vehiclemodel "railkeeper/backend/internal/inventory/vehicle"
	"strings"
)

func (s *VehicleService) functionService() *vehiclemodel.FunctionService {
	return vehiclemodel.NewFunctionService(vehicledata.NewFunctionStore(s.db), func(ctx context.Context, id string) error { _, err := s.Get(ctx, id); return err })
}
func (s *VehicleService) ListFunctions(ctx context.Context, vehicleID string) ([]VehicleFunction, error) {
	return s.functionService().ListFunctions(ctx, vehicleID)
}
func (s *VehicleService) UpsertFunction(ctx context.Context, vehicleID, key string, input VehicleFunctionInput) (*VehicleFunction, error) {
	return s.functionService().UpsertFunction(ctx, vehicleID, key, input)
}
func (s *VehicleService) GetFunction(ctx context.Context, vehicleID, key string) (*VehicleFunction, error) {
	return s.functionService().GetFunction(ctx, vehicleID, key)
}
func (s *VehicleService) DeleteFunction(ctx context.Context, vehicleID, key string) (*VehicleFunction, error) {
	return s.functionService().DeleteFunction(ctx, vehicleID, key)
}
func (s *VehicleService) loadVehicleFunctions(ctx context.Context, vehicleID string) ([]VehicleFunction, error) {
	return vehicledata.NewFunctionStore(s.db).Functions(ctx, strings.TrimSpace(vehicleID))
}
