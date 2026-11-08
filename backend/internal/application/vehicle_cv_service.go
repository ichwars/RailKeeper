package application

import (
	"context"
	"strings"

	"railkeeper/backend/internal/decoder/workflow"
	"railkeeper/backend/internal/infrastructure/vehicledata"
)

func (s *VehicleService) cvService() *workflow.CVService {
	return workflow.NewCVService(vehicledata.NewCVStore(s.db))
}
func (s *VehicleService) ListCVValues(ctx context.Context, vehicleID string) ([]VehicleCVValue, error) {
	vehicleID = strings.TrimSpace(vehicleID)
	if vehicleID == "" {
		return nil, ErrVehicleNotFound
	}
	if _, err := s.Get(ctx, vehicleID); err != nil {
		return nil, err
	}
	return s.loadVehicleCVValues(ctx, vehicleID)
}

func (s *VehicleService) ListCVFiles(ctx context.Context, vehicleID string) ([]VehicleCVFile, error) {
	vehicleID = strings.TrimSpace(vehicleID)
	if vehicleID == "" {
		return nil, ErrVehicleNotFound
	}
	if _, err := s.Get(ctx, vehicleID); err != nil {
		return nil, err
	}
	return s.loadVehicleCVFiles(ctx, vehicleID)
}

func (s *VehicleService) DeleteCVValue(ctx context.Context, vehicleID, cvValueID string, expectedRevision ...string) (*VehicleCVValue, error) {
	if len(expectedRevision) > 1 {
		return nil, ErrVehicleValidation
	}
	var expected *string
	if len(expectedRevision) == 1 {
		expected = &expectedRevision[0]
	}
	return s.DeleteCVValueCompared(ctx, vehicleID, cvValueID, VehicleCVDeletionInput{ExpectedRevision: expected})
}

func (s *VehicleService) DeleteCVFile(ctx context.Context, vehicleID, fileID string, expectedRevision ...string) (*VehicleCVFile, error) {
	if len(expectedRevision) > 1 {
		return nil, ErrVehicleValidation
	}
	var expected *string
	if len(expectedRevision) == 1 {
		expected = &expectedRevision[0]
	}
	return s.DeleteCVFileCompared(ctx, vehicleID, fileID, VehicleCVDeletionInput{ExpectedRevision: expected})
}

func (s *VehicleService) CreateCVValue(ctx context.Context, vehicleID string, input VehicleCVValueInput) (*VehicleCVValue, error) {
	v, err := s.cvService().CreateCVValue(ctx, vehicleID, input)
	return v, cvRuleError(err)
}
func (s *VehicleService) UpdateCVValue(ctx context.Context, vehicleID, cvValueID string, input VehicleCVValueInput) (*VehicleCVValue, error) {
	v, err := s.cvService().UpdateCVValue(ctx, vehicleID, cvValueID, input)
	return v, cvRuleError(err)
}
func (s *VehicleService) GetCVValue(ctx context.Context, vehicleID, cvValueID string) (*VehicleCVValue, error) {
	v, err := s.cvService().GetCVValue(ctx, vehicleID, cvValueID)
	return v, cvRuleError(err)
}
func (s *VehicleService) DeleteCVValueCompared(ctx context.Context, vehicleID, cvValueID string, input VehicleCVDeletionInput) (*VehicleCVValue, error) {
	v, err := s.cvService().DeleteCVValueCompared(ctx, vehicleID, cvValueID, input)
	return v, cvRuleError(err)
}
func (s *VehicleService) CreateCVFile(ctx context.Context, vehicleID string, input VehicleCVFileInput) (*VehicleCVFile, error) {
	v, err := s.cvService().CreateCVFile(ctx, vehicleID, input)
	return v, cvRuleError(err)
}
func (s *VehicleService) GetCVFile(ctx context.Context, vehicleID, fileID string) (*VehicleCVFile, error) {
	v, err := s.cvService().GetCVFile(ctx, vehicleID, fileID)
	return v, cvRuleError(err)
}
func (s *VehicleService) DeleteCVFileCompared(ctx context.Context, vehicleID, fileID string, input VehicleCVDeletionInput) (*VehicleCVFile, error) {
	v, err := s.cvService().DeleteCVFileCompared(ctx, vehicleID, fileID, input)
	return v, cvRuleError(err)
}
func (s *VehicleService) loadVehicleCVValues(ctx context.Context, vehicleID string) ([]VehicleCVValue, error) {
	v, err := s.cvService().Values(ctx, vehicleID)
	return v, cvRuleError(err)
}
func (s *VehicleService) loadVehicleCVFiles(ctx context.Context, vehicleID string) ([]VehicleCVFile, error) {
	v, err := s.cvService().Files(ctx, vehicleID)
	return v, cvRuleError(err)
}
