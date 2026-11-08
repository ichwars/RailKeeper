package application

import (
	"context"
	"database/sql"

	"railkeeper/backend/internal/decoder/workflow"
	"railkeeper/backend/internal/infrastructure/vehicledata"
)

type decoderSnapshotInput struct{ ID, Kind, Note, RecordedAt, SourceRoot string }

func captureDecoderCVSnapshotTx(ctx context.Context, tx *sql.Tx, workspace decoderWorkspaceRecord, input decoderSnapshotInput) error {
	return workflow.CaptureSnapshot(ctx, vehicledata.SnapshotCaptureFor(tx, input.SourceRoot), workflow.SnapshotWorkspace{
		ID: workspace.ID, VehicleID: workspace.VehicleID, Profile: workspace.Profile,
		InstallationID: workflow.OptionalText{Text: workspace.InstallationID.String, Valid: workspace.InstallationID.Valid},
	}, workflow.SnapshotCaptureInput{ID: input.ID, Kind: input.Kind, Note: input.Note, RecordedAt: input.RecordedAt})
}
