package workflow

import (
	"context"
	"strings"

	"railkeeper/backend/internal/decoder"
)

func (s *ConfigurationService) CreateDecoderCVWorkspace(ctx context.Context, vehicleID string, input decoder.DecoderCVWorkspaceInput, actor string) (*decoder.DecoderConfigurationContext, error) {
	input, err := decoder.CleanWorkspaceInput(input)
	if err != nil {
		return nil, err
	}
	return s.change(ctx, vehicleID, input.ExpectedRevision, "DecoderCVWorkspaceCreated", input.ID, actor,
		func(tx ConfigurationMutation, before *decoder.DecoderConfigurationContext, now string) error {
			exists, err := tx.DecoderWorkspaceExists(ctx, input.ID, before.VehicleID, input.Profile)
			if err != nil {
				return err
			}
			item, err := decoder.PlanWorkspace(before, input, now, exists)
			if err != nil {
				return err
			}
			return tx.InsertDecoderWorkspace(ctx, item)
		})
}

func (s *ConfigurationService) CaptureDecoderCVCheckpoint(ctx context.Context, vehicleID, workspaceID string, input decoder.DecoderCVCheckpointInput, actor string) (*decoder.DecoderConfigurationContext, error) {
	if !decoder.ValidCheckpointInput(input) {
		return nil, decoder.ErrValidation
	}
	return s.change(ctx, vehicleID, input.ExpectedRevision, "DecoderCVCheckpointRecorded", input.ID, actor,
		func(tx ConfigurationMutation, before *decoder.DecoderConfigurationContext, now string) error {
			workspace := decoder.WorkspaceIn(before, workspaceID)
			if workspace == nil {
				return decoder.ErrNotFound
			}
			return captureExplicitCheckpoint(ctx, tx, *workspace, SnapshotCaptureInput{ID: input.ID, Kind: "checkpoint", Note: strings.TrimSpace(input.Note), RecordedAt: now})
		})
}

func (s *ConfigurationService) AssignDecoderCVWorkspace(ctx context.Context, vehicleID, workspaceID string, input decoder.DecoderCVAssignmentInput, actor string) (*decoder.DecoderConfigurationContext, error) {
	if !decoder.ValidAssignmentInput(input) {
		return nil, decoder.ErrValidation
	}
	return s.change(ctx, vehicleID, input.ExpectedRevision, "DecoderCVWorkspaceAssigned", workspaceID, actor,
		func(tx ConfigurationMutation, before *decoder.DecoderConfigurationContext, now string) error {
			assignment, err := decoder.PlanWorkspaceAssignment(before, workspaceID, input, now)
			if err != nil {
				return err
			}
			if err := captureExplicitCheckpoint(ctx, tx, assignment.Before, SnapshotCaptureInput{ID: input.SnapshotID, Kind: "checkpoint",
				Note: "Arbeitsbestand vor bestätigter Hardwarezuordnung. / Working state before confirmed hardware assignment.", RecordedAt: now}); err != nil {
				return err
			}
			return tx.AssignDecoderWorkspace(ctx, assignment.After)
		})
}

func captureExplicitCheckpoint(ctx context.Context, tx ConfigurationMutation, workspace decoder.DecoderCVWorkspace, input SnapshotCaptureInput) error {
	exists, err := tx.DecoderSnapshotExists(ctx, input.ID)
	if err != nil {
		return err
	}
	if exists {
		return decoder.ErrConflict
	}
	installation := OptionalText{}
	if workspace.InstallationID != nil {
		installation = OptionalText{Text: *workspace.InstallationID, Valid: true}
	}
	return CaptureSnapshot(ctx, tx, SnapshotWorkspace{ID: workspace.ID, VehicleID: workspace.VehicleID, Profile: workspace.Profile, InstallationID: installation}, input)
}
