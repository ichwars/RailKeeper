package application

import (
	"context"

	"railkeeper/backend/internal/decoder/workflow"
	ecospkg "railkeeper/backend/internal/ecos"
	"railkeeper/backend/internal/ecosmodel"
)

type ECoSStoredCVSource = ecosmodel.ECoSStoredCVSource
type ECoSCVImportResult = workflow.ECoSCVImportResult

func (s *VehicleService) ImportECoSStoredCVs(ctx context.Context, vehicleID string, source ECoSStoredCVSource) (ECoSCVImportResult, error) {
	plan, err := ecosmodel.PlanStoredCVAdoption(source, ecosmodel.StoredCVLimits{MaximumValues: ecospkg.CVDatabaseLimit, MaximumObjectID: maxDigitalCenterObjectID})
	if err != nil {
		return ECoSCVImportResult{}, cvRuleError(err)
	}
	result, err := s.cvService().ImportECoSStoredCVs(ctx, vehicleID, workflow.ECoSStoredCVSource{Host: plan.Host, ObjectID: plan.ObjectID, ReadAt: plan.ReadAt, Protocol: plan.Protocol, Values: plan.Values})
	return result, cvRuleError(err)
}
