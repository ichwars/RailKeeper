package application

import "context"

func (s *VehicleService) ListSpareParts(ctx context.Context, vehicleID string) ([]VehicleSparePart, error) {
	return s.workshop.ListSpareParts(ctx, vehicleID)
}
func (s *VehicleService) CreateSparePart(ctx context.Context, vehicleID string, input VehicleSparePartInput) (*VehicleSparePart, error) {
	return s.workshop.CreateSparePart(ctx, vehicleID, input)
}
func (s *VehicleService) UpdateSparePart(ctx context.Context, vehicleID, sparePartID string, input VehicleSparePartInput) (*VehicleSparePart, error) {
	return s.workshop.UpdateSparePart(ctx, vehicleID, sparePartID, input)
}
func (s *VehicleService) GetSparePart(ctx context.Context, vehicleID, sparePartID string) (*VehicleSparePart, error) {
	return s.workshop.GetSparePart(ctx, vehicleID, sparePartID)
}
func (s *VehicleService) DeleteSparePart(ctx context.Context, vehicleID, sparePartID string) (*VehicleSparePart, error) {
	return s.workshop.DeleteSparePart(ctx, vehicleID, sparePartID)
}
func (s *VehicleService) loadVehicleSpareParts(ctx context.Context, vehicleID string) ([]VehicleSparePart, error) {
	return s.workshopRows.ListSpareParts(ctx, vehicleID)
}
