package application

import (
	"context"

	"railkeeper/backend/internal/decoder"
	"railkeeper/backend/internal/decoder/workflow"
)

const MaxVehicleCVChanges = workflow.MaxVehicleCVChanges

var ErrVehicleCVConflict = decoder.ErrCVConflict

type VehicleCVChange = decoder.VehicleCVChange
type VehicleCVChangesInput = decoder.VehicleCVChangesInput

func (s *VehicleService) ApplyCVChanges(ctx context.Context, vehicleID string, input VehicleCVChangesInput) (int, error) {
	n, err := s.cvService().ApplyCVChanges(ctx, vehicleID, input)
	return n, cvRuleError(err)
}
