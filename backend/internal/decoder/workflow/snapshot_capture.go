package workflow

import (
	"context"
	"encoding/json"
	"fmt"

	"railkeeper/backend/internal/credentials"
	"railkeeper/backend/internal/decoder"
)

// CaptureSnapshot records a partial working state, not a complete decoder readout.
func CaptureSnapshot(ctx context.Context, repository SnapshotCapture, workspace SnapshotWorkspace, input SnapshotCaptureInput) error {
	values, err := repository.Values(ctx, workspace)
	if err != nil {
		return err
	}
	for i := range values {
		values[i].Revision = ""
	}
	sources, err := captureSnapshotSources(ctx, repository, workspace, values)
	if err != nil {
		return err
	}
	sourceIDs := map[string]bool{}
	for _, source := range sources {
		sourceIDs[source.ID] = true
	}
	payload, err := json.Marshal(decoder.DecoderCVSnapshotPayload{Values: values})
	if err != nil {
		return fmt.Errorf("encode decoder working-state snapshot: %w", err)
	}
	if err := decoder.CheckSnapshotPayload(string(payload), workspace.VehicleID, workspace.Profile, sourceIDs); err != nil {
		return err
	}
	hardware, err := captureHardware(ctx, repository, workspace)
	if err != nil {
		return err
	}
	identity, err := repository.VehicleIdentity(ctx, workspace.VehicleID)
	if err != nil {
		return err
	}
	if input.ID == "" {
		input.ID = credentials.RandomID()
	}
	if err := repository.InsertPartialSnapshot(ctx, SnapshotWrite{
		ID: input.ID, WorkspaceID: workspace.ID, VehicleID: workspace.VehicleID, VehicleName: identity.Name,
		InventoryNumber: identity.InventoryNumber, Profile: workspace.Profile, InstallationID: workspace.InstallationID,
		Kind: input.Kind, Payload: string(payload), Hardware: hardware, Note: input.Note, RecordedAt: input.RecordedAt,
	}); err != nil {
		return err
	}
	for _, source := range sources {
		if err := repository.InsertSnapshotSource(ctx, input.ID, source); err != nil {
			return err
		}
	}
	return nil
}
func captureHardware(ctx context.Context, repository SnapshotCapture, workspace SnapshotWorkspace) (string, error) {
	if !workspace.InstallationID.Valid {
		return "{}", nil
	}
	hardware, err := repository.Hardware(ctx, workspace.InstallationID.Text, workspace.VehicleID)
	if err != nil {
		return "", err
	}
	encoded, err := json.Marshal(hardware)
	if err != nil {
		return "", err
	}
	if err := decoder.CheckSnapshotHardware(string(encoded), workspace.InstallationID.Text, &hardware.DecoderID); err != nil {
		return "", err
	}
	return string(encoded), nil
}
