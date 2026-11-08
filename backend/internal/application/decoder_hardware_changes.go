package application

import (
	"context"

	"railkeeper/backend/internal/decoder"
)

type DecoderInstanceInput = decoder.DecoderInstanceInput
type DecoderInstallationInput = decoder.DecoderInstallationInput
type DecoderRemovalInput = decoder.DecoderRemovalInput
type DecoderInstallationDatesInput = decoder.DecoderInstallationDatesInput

func (s *VehicleService) CreateDecoderInstance(ctx context.Context, vehicleID string, input DecoderInstanceInput, actor string) (*DecoderConfigurationContext, error) {
	value, err := s.configurationService("").CreateDecoderInstance(ctx, vehicleID, input, actor)
	return value, cvRuleError(err)
}
func (s *VehicleService) UpdateDecoderInstance(ctx context.Context, vehicleID string, input DecoderInstanceInput, actor string) (*DecoderConfigurationContext, error) {
	value, err := s.configurationService("").UpdateDecoderInstance(ctx, vehicleID, input, actor)
	return value, cvRuleError(err)
}
func (s *VehicleService) InstallDecoder(ctx context.Context, vehicleID string, input DecoderInstallationInput, actor string) (*DecoderConfigurationContext, error) {
	value, err := s.configurationService("").InstallDecoder(ctx, vehicleID, input, actor)
	return value, cvRuleError(err)
}
func (s *VehicleService) RemoveDecoder(ctx context.Context, vehicleID string, installationID string, input DecoderRemovalInput, actor string) (*DecoderConfigurationContext, error) {
	value, err := s.configurationService("").RemoveDecoder(ctx, vehicleID, installationID, input, actor)
	return value, cvRuleError(err)
}
func (s *VehicleService) CorrectDecoderInstallationDates(ctx context.Context, vehicleID string, installationID string, input DecoderInstallationDatesInput, actor string) (*DecoderConfigurationContext, error) {
	value, err := s.configurationService("").CorrectDecoderInstallationDates(ctx, vehicleID, installationID, input, actor)
	return value, cvRuleError(err)
}
