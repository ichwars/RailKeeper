package application

import "context"

func (s *VehicleService) ListMaintenance(ctx context.Context, vehicleID string) ([]VehicleMaintenance, error) {
	return s.workshop.ListMaintenance(ctx, vehicleID)
}
func (s *VehicleService) CreateMaintenance(ctx context.Context, vehicleID string, input VehicleMaintenanceInput) (*VehicleMaintenance, error) {
	return s.workshop.CreateMaintenance(ctx, vehicleID, input)
}
func (s *VehicleService) UpdateMaintenance(ctx context.Context, vehicleID, maintenanceID string, input VehicleMaintenanceInput) (*VehicleMaintenance, error) {
	return s.workshop.UpdateMaintenance(ctx, vehicleID, maintenanceID, input)
}
func (s *VehicleService) GetMaintenance(ctx context.Context, vehicleID, maintenanceID string) (*VehicleMaintenance, error) {
	return s.workshop.GetMaintenance(ctx, vehicleID, maintenanceID)
}
func (s *VehicleService) DeleteMaintenance(ctx context.Context, vehicleID, maintenanceID string) (*VehicleMaintenance, error) {
	return s.workshop.DeleteMaintenance(ctx, vehicleID, maintenanceID)
}
func (s *VehicleService) loadVehicleMaintenance(ctx context.Context, vehicleID string) ([]VehicleMaintenance, error) {
	return s.workshopRows.ListMaintenance(ctx, vehicleID)
}
