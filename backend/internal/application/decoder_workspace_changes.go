package application

import (
	"context"
	"database/sql"
	"strings"
)

type DecoderCVWorkspaceInput struct {
	ID               string `json:"id"`
	ExpectedRevision string `json:"expectedRevision"`
	Profile          string `json:"profile"`
}

func (s *VehicleService) CreateDecoderCVWorkspace(ctx context.Context, vehicleID string, input DecoderCVWorkspaceInput, actor string) (*DecoderConfigurationContext, error) {
	input.Profile = strings.TrimSpace(input.Profile)
	if !validDecoderID(input.ID) || input.Profile == "" || len(input.Profile) > 160 {
		return nil, ErrDecoderValidation
	}
	return s.changeDecoderConfiguration(ctx, vehicleID, input.ExpectedRevision, "DecoderCVWorkspaceCreated", input.ID, actor,
		func(tx *sql.Tx, before *DecoderConfigurationContext, now string) error {
			var exists bool
			if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM decoder_cv_workspaces WHERE id=? OR (vehicle_id=? AND profile_text=?))`, input.ID, before.VehicleID, input.Profile).Scan(&exists); err != nil {
				return err
			}
			if exists {
				return ErrDecoderConflict
			}
			_, err := tx.ExecContext(ctx, `INSERT INTO decoder_cv_workspaces
(id,vehicle_id,profile_text,origin,installation_id,created_at,updated_at) VALUES(?,?,?,'manual',NULL,?,?)`, input.ID, before.VehicleID, input.Profile, now, now)
			return err
		})
}

type DecoderCVCheckpointInput struct {
	ID               string `json:"id"`
	ExpectedRevision string `json:"expectedRevision"`
	Note             string `json:"note"`
}

func (s *VehicleService) CaptureDecoderCVCheckpoint(ctx context.Context, vehicleID, workspaceID string, input DecoderCVCheckpointInput, sourceRoot, actor string) (*DecoderConfigurationContext, error) {
	if !validDecoderID(input.ID) || len(input.Note) > 2000 {
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

type DecoderCVAssignmentInput struct {
	ExpectedRevision string  `json:"expectedRevision"`
	InstallationID   *string `json:"installationId"`
	SnapshotID       string  `json:"snapshotId"`
}

func (s *VehicleService) AssignDecoderCVWorkspace(ctx context.Context, vehicleID, workspaceID string, input DecoderCVAssignmentInput, sourceRoot, actor string) (*DecoderConfigurationContext, error) {
	if !validDecoderID(input.SnapshotID) || (input.InstallationID != nil && *input.InstallationID == "") {
		return nil, ErrDecoderValidation
	}
	return s.changeDecoderConfiguration(ctx, vehicleID, input.ExpectedRevision, "DecoderCVWorkspaceAssigned", workspaceID, actor,
		func(tx *sql.Tx, before *DecoderConfigurationContext, now string) error {
			workspace := decoderWorkspaceIn(before, workspaceID)
			if workspace == nil {
				return ErrDecoderNotFound
			}
			if (workspace.InstallationID == nil && input.InstallationID == nil) ||
				(workspace.InstallationID != nil && input.InstallationID != nil && *workspace.InstallationID == *input.InstallationID) {
				return ErrDecoderConflict
			}
			if input.InstallationID != nil {
				installation := decoderInstallationIn(before, *input.InstallationID)
				if installation == nil || installation.VehicleID == nil || *installation.VehicleID != before.VehicleID || installation.State != "installed" {
					return ErrDecoderValidation
				}
			}
			if err := captureExplicitDecoderCheckpointTx(ctx, tx, *workspace, decoderSnapshotInput{ID: input.SnapshotID, Kind: "checkpoint",
				Note: "Arbeitsbestand vor bestätigter Hardwarezuordnung. / Working state before confirmed hardware assignment.", RecordedAt: now, SourceRoot: sourceRoot}); err != nil {
				return err
			}
			_, err := tx.ExecContext(ctx, `UPDATE decoder_cv_workspaces SET installation_id=?,updated_at=? WHERE id=?`, input.InstallationID, now, workspace.ID)
			return err
		})
}

func decoderWorkspaceIn(result *DecoderConfigurationContext, id string) *DecoderCVWorkspace {
	for index := range result.Workspaces {
		if result.Workspaces[index].ID == id {
			return &result.Workspaces[index]
		}
	}
	return nil
}

func captureExplicitDecoderCheckpointTx(ctx context.Context, tx *sql.Tx, workspace DecoderCVWorkspace, input decoderSnapshotInput) error {
	var exists bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM decoder_cv_snapshots WHERE id=?)`, input.ID).Scan(&exists); err != nil {
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
