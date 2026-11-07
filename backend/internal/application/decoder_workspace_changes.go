package application

import (
	"context"
	"database/sql"
	"strings"

	"railkeeper/backend/internal/decoder"
	"railkeeper/backend/internal/infrastructure/vehicledata"
)

type DecoderCVWorkspaceInput = decoder.DecoderCVWorkspaceInput

func (s *VehicleService) CreateDecoderCVWorkspace(ctx context.Context, vehicleID string, input DecoderCVWorkspaceInput, actor string) (*DecoderConfigurationContext, error) {
	input, err := decoder.CleanWorkspaceInput(input)
	if err != nil {
		return nil, err
	}
	return s.changeDecoderConfiguration(ctx, vehicleID, input.ExpectedRevision, "DecoderCVWorkspaceCreated", input.ID, actor,
		func(tx *sql.Tx, before *DecoderConfigurationContext, now string) error {
			exists, err := vehicledata.DecoderWorkspaceExists(ctx, tx, input.ID, before.VehicleID, input.Profile)
			if err != nil {
				return err
			}
			item, err := decoder.PlanWorkspace(before, input, now, exists)
			if err != nil {
				return err
			}
			return vehicledata.InsertDecoderWorkspace(ctx, tx, item)
		})
}

type DecoderCVCheckpointInput = decoder.DecoderCVCheckpointInput

func (s *VehicleService) CaptureDecoderCVCheckpoint(ctx context.Context, vehicleID, workspaceID string, input DecoderCVCheckpointInput, sourceRoot, actor string) (*DecoderConfigurationContext, error) {
	if !decoder.ValidCheckpointInput(input) {
		return nil, ErrDecoderValidation
	}
	return s.changeDecoderConfiguration(ctx, vehicleID, input.ExpectedRevision, "DecoderCVCheckpointRecorded", input.ID, actor,
		func(tx *sql.Tx, before *DecoderConfigurationContext, now string) error {
			workspace := decoderWorkspaceIn(before, workspaceID)
			if workspace == nil {
				return ErrDecoderNotFound
			}
			return captureExplicitDecoderCheckpointTx(ctx, tx, *workspace, decoderSnapshotInput{ID: input.ID, Kind: "checkpoint", Note: strings.TrimSpace(input.Note), RecordedAt: now, SourceRoot: sourceRoot})
		})
}

type DecoderCVAssignmentInput = decoder.DecoderCVAssignmentInput

func (s *VehicleService) AssignDecoderCVWorkspace(ctx context.Context, vehicleID, workspaceID string, input DecoderCVAssignmentInput, sourceRoot, actor string) (*DecoderConfigurationContext, error) {
	if !decoder.ValidAssignmentInput(input) {
		return nil, ErrDecoderValidation
	}
	return s.changeDecoderConfiguration(ctx, vehicleID, input.ExpectedRevision, "DecoderCVWorkspaceAssigned", workspaceID, actor,
		func(tx *sql.Tx, before *DecoderConfigurationContext, now string) error {
			assignment, err := decoder.PlanWorkspaceAssignment(before, workspaceID, input, now)
			if err != nil {
				return err
			}
			if err := captureExplicitDecoderCheckpointTx(ctx, tx, assignment.Before, decoderSnapshotInput{ID: input.SnapshotID, Kind: "checkpoint",
				Note: "Arbeitsbestand vor bestätigter Hardwarezuordnung. / Working state before confirmed hardware assignment.", RecordedAt: now, SourceRoot: sourceRoot}); err != nil {
				return err
			}
			return vehicledata.AssignDecoderWorkspace(ctx, tx, assignment.After)
		})
}

func decoderWorkspaceIn(result *DecoderConfigurationContext, id string) *DecoderCVWorkspace {
	return decoder.WorkspaceIn(result, id)
}

func captureExplicitDecoderCheckpointTx(ctx context.Context, tx *sql.Tx, workspace DecoderCVWorkspace, input decoderSnapshotInput) error {
	exists, err := vehicledata.DecoderSnapshotExists(ctx, tx, input.ID)
	if err != nil {
		return err
	}
	if exists {
		return ErrDecoderConflict
	}
	installation := sql.NullString{}
	if workspace.InstallationID != nil {
		installation = sql.NullString{String: *workspace.InstallationID, Valid: true}
	}
	return captureDecoderCVSnapshotTx(ctx, tx, decoderWorkspaceRecord{ID: workspace.ID, VehicleID: workspace.VehicleID, Profile: workspace.Profile, InstallationID: installation}, input)
}
