package application

import (
	"context"

	"railkeeper/backend/internal/decoder"
	"railkeeper/backend/internal/decoder/workflow"
	"railkeeper/backend/internal/infrastructure/vehicledata"
)

var (
	ErrDecoderValidation = decoder.ErrValidation
	ErrDecoderNotFound   = decoder.ErrNotFound
	ErrDecoderConflict   = decoder.ErrConflict
)

func (s *VehicleService) configurationService(sourceRoot string) *workflow.ConfigurationService {
	return workflow.NewConfigurationService(vehicledata.NewConfigurationStore(s.db, sourceRoot))
}
func (s *VehicleService) DecoderConfigurationContext(ctx context.Context, id string) (*DecoderConfigurationContext, error) {
	value, err := s.configurationService("").Context(ctx, id)
	return value, cvRuleError(err)
}
