package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"railkeeper/backend/internal/infrastructure/vehicledata"
)

type decoderSnapshotInput struct {
	ID, Kind, Note, RecordedAt, SourceRoot string
}

func captureDecoderCVSnapshotTx(ctx context.Context, tx *sql.Tx, workspace decoderWorkspaceRecord, input decoderSnapshotInput) error {
	values, err := loadDecoderSnapshotValuesTx(ctx, tx, workspace)
	if err != nil {
		return err
	}
	sources, err := loadDecoderSnapshotSourcesTx(ctx, tx, workspace, values, input.SourceRoot)
	if err != nil {
		return err
	}
	sourceIDs := map[string]bool{}
	for _, source := range sources {
		sourceIDs[source.ID] = true
	}
	payload, err := json.Marshal(DecoderCVSnapshotPayload{Values: values})
	if err != nil {
		return fmt.Errorf("encode decoder working-state snapshot: %w", err)
	}
	if err := checkDecoderSnapshotPayload(string(payload), workspace.VehicleID, workspace.Profile, sourceIDs); err != nil {
		return err
	}
	hardware, err := captureDecoderHardwareTx(ctx, tx, workspace)
	if err != nil {
		return err
	}
	var name, inventory string
	if err := tx.QueryRowContext(ctx, `SELECT name, inventory_number FROM vehicles WHERE id=?`, workspace.VehicleID).
		Scan(&name, &inventory); err != nil {
		return fmt.Errorf("read snapshot vehicle identity: %w", err)
	}
	if input.ID == "" {
		input.ID = randomID()
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO decoder_cv_snapshots
(id, workspace_id, vehicle_reference_id, vehicle_name, vehicle_inventory_number, profile_text,
installation_id, kind, scope, payload_json, hardware_json, note, recorded_at)
VALUES(?,?,?,?,?,?,?,?,'partial_working_state',?,?,?,?)`, input.ID, workspace.ID, workspace.VehicleID,
		name, inventory, workspace.Profile, workspace.InstallationID, input.Kind, string(payload), hardware,
		input.Note, input.RecordedAt); err != nil {
		return fmt.Errorf("record decoder working-state snapshot: %w", err)
	}
	for _, source := range sources {
		if _, err := tx.ExecContext(ctx, `INSERT INTO decoder_cv_snapshot_sources
(snapshot_id, source_file_id, metadata_json, content_state, blob_id) VALUES(?,?,?,?,?)`, input.ID, source.ID,
			source.Metadata, source.State, source.BlobID); err != nil {
			return fmt.Errorf("record immutable decoder source: %w", err)
		}
	}
	return nil
}

func loadDecoderSnapshotValuesTx(ctx context.Context, tx *sql.Tx, workspace decoderWorkspaceRecord) ([]VehicleCVValue, error) {
	rows, err := tx.QueryContext(ctx, vehicledata.CVSelect+` WHERE vehicle_id=? AND decoder_profile=? ORDER BY id`,
		workspace.VehicleID, workspace.Profile)
	if err != nil {
		return nil, fmt.Errorf("read decoder snapshot values: %w", err)
	}
	defer func() { _ = rows.Close() }()
	values := []VehicleCVValue{}
	for rows.Next() {
		record, err := vehicledata.ScanCV(rows)
		if err != nil {
			return nil, err
		}
		value := cvRecordValue(record)
		value.Revision = "" // Transport comparison tokens are not historical evidence.
		values = append(values, value)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	for i := range values {
		values[i].History, err = loadVehicleCVHistory(ctx, tx, values[i].ID)
		if err != nil {
			return nil, err
		}
	}
	return values, nil
}

func captureDecoderHardwareTx(ctx context.Context, tx *sql.Tx, workspace decoderWorkspaceRecord) (string, error) {
	if !workspace.InstallationID.Valid {
		return "{}", nil
	}
	var hardware DecoderSnapshotHardware
	if err := tx.QueryRowContext(ctx, `SELECT d.id, d.label, d.manufacturer, d.model, d.serial_number,
i.id, i.state, i.installed_at, i.removed_at
FROM decoder_installations i JOIN decoder_instances d ON d.id=i.decoder_id
WHERE i.id=? AND i.vehicle_reference_id=?`, workspace.InstallationID.String, workspace.VehicleID).
		Scan(&hardware.DecoderID, &hardware.DecoderLabel, &hardware.Manufacturer, &hardware.Model,
			&hardware.SerialNumber, &hardware.InstallationID, &hardware.InstallationState, &hardware.InstalledAt, &hardware.RemovedAt); err != nil {
		return "", fmt.Errorf("capture confirmed decoder hardware: %w", err)
	}
	encoded, err := json.Marshal(hardware)
	if err != nil {
		return "", err
	}
	if err := checkDecoderSnapshotHardware(string(encoded), workspace.InstallationID.String,
		map[string]map[string]any{hardware.InstallationID: {"decoder_id": hardware.DecoderID}}); err != nil {
		return "", err
	}
	return string(encoded), nil
}
