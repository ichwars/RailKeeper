package application

import (
	"context"

	"railkeeper/backend/internal/decoder"
)

type DecoderCVWorkspaceInput = decoder.DecoderCVWorkspaceInput
type DecoderCVCheckpointInput = decoder.DecoderCVCheckpointInput
type DecoderCVAssignmentInput = decoder.DecoderCVAssignmentInput

func (s *VehicleService) CreateDecoderCVWorkspace(ctx context.Context, id string, input DecoderCVWorkspaceInput, actor string) (*DecoderConfigurationContext, error) {
	value, err := s.configurationService("").CreateDecoderCVWorkspace(ctx, id, input, actor)
	return value, cvRuleError(err)
}
func (s *VehicleService) CaptureDecoderCVCheckpoint(ctx context.Context, id, workspaceID string, input DecoderCVCheckpointInput, sourceRoot, actor string) (*DecoderConfigurationContext, error) {
	value, err := s.configurationService(sourceRoot).CaptureDecoderCVCheckpoint(ctx, id, workspaceID, input, actor)
	return value, cvRuleError(err)
}
func (s *VehicleService) AssignDecoderCVWorkspace(ctx context.Context, id, workspaceID string, input DecoderCVAssignmentInput, sourceRoot, actor string) (*DecoderConfigurationContext, error) {
	value, err := s.configurationService(sourceRoot).AssignDecoderCVWorkspace(ctx, id, workspaceID, input, actor)
	return value, cvRuleError(err)
}
