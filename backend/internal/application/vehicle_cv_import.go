package application

import (
	"context"

	"railkeeper/backend/internal/decoder/workflow"
)

const MaxVehicleCVImportValues = workflow.MaxVehicleCVImportValues

type VehicleCVImportInput = workflow.VehicleCVImportInput

func (s *VehicleService) ImportCVValues(ctx context.Context, vehicleID string, input VehicleCVImportInput) (int, error) {
	n, err := s.cvService().ImportCVValues(ctx, vehicleID, input)
	return n, cvRuleError(err)
}
