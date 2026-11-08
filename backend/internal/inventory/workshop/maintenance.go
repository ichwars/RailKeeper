package workshop

import (
	"context"
	"strings"
	"time"

	"railkeeper/backend/internal/credentials"
)

func (s *Service) ListMaintenance(ctx context.Context, vehicleID string) ([]VehicleMaintenance, error) {
	vehicleID = strings.TrimSpace(vehicleID)
	if vehicleID == "" {
		return nil, ErrVehicleNotFound
	}
	if err := s.parent(ctx, vehicleID); err != nil {
		return nil, err
	}
	return s.repository.ListMaintenance(ctx, vehicleID)
}
func (s *Service) CreateMaintenance(ctx context.Context, vehicleID string, input VehicleMaintenanceInput) (*VehicleMaintenance, error) {
	vehicleID = strings.TrimSpace(vehicleID)
	input = cleanVehicleMaintenanceInput(input)
	if vehicleID == "" || !isValidVehicleMaintenanceInput(input) {
		return nil, ErrVehicleValidation
	}
	if err := s.parent(ctx, vehicleID); err != nil {
		return nil, err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	maintenance := VehicleMaintenance{
		ID:              credentials.RandomID(),
		VehicleID:       vehicleID,
		Kind:            input.Kind,
		Status:          input.Status,
		ConditionRating: input.ConditionRating,
		DueDate:         input.DueDate,
		CompletedAt:     input.CompletedAt,
		Cost:            input.Cost,
		Notes:           input.Notes,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := s.repository.CreateMaintenance(ctx, maintenance); err != nil {
		return nil, err
	}
	return &maintenance, nil
}
func (s *Service) UpdateMaintenance(ctx context.Context, vehicleID, maintenanceID string, input VehicleMaintenanceInput) (*VehicleMaintenance, error) {
	vehicleID = strings.TrimSpace(vehicleID)
	maintenanceID = strings.TrimSpace(maintenanceID)
	input = cleanVehicleMaintenanceInput(input)
	if vehicleID == "" || maintenanceID == "" || !isValidVehicleMaintenanceInput(input) {
		return nil, ErrVehicleValidation
	}
	now := time.Now().UTC().Format(time.RFC3339)
	if err := s.repository.UpdateMaintenance(ctx, vehicleID, maintenanceID, input, now); err != nil {
		return nil, err
	}
	return s.GetMaintenance(ctx, vehicleID, maintenanceID)
}
func (s *Service) GetMaintenance(ctx context.Context, vehicleID, id string) (*VehicleMaintenance, error) {
	return s.repository.GetMaintenance(ctx, vehicleID, id)
}
func (s *Service) DeleteMaintenance(ctx context.Context, vehicleID, maintenanceID string) (*VehicleMaintenance, error) {
	maintenance, err := s.GetMaintenance(ctx, vehicleID, maintenanceID)
	if err != nil {
		return nil, err
	}
	if err := s.repository.DeleteMaintenance(ctx, vehicleID, maintenanceID); err != nil {
		return nil, err
	}
	return maintenance, nil
}
