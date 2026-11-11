package vehicledata

import (
	"context"
	"database/sql"
	"fmt"

	"railkeeper/backend/internal/decoder"
	"railkeeper/backend/internal/decoder/workflow"
)

type snapshotCaptureMutation struct {
	tx         *sql.Tx
	sourceRoot string
}

func SnapshotCaptureFor(tx *sql.Tx, sourceRoot string) workflow.SnapshotCapture {
	return &snapshotCaptureMutation{tx: tx, sourceRoot: sourceRoot}
}
func LegacySnapshotMutationFor(tx *sql.Tx, sourceRoot string) workflow.LegacySnapshotMutation {
	return &snapshotCaptureMutation{tx: tx, sourceRoot: sourceRoot}
}
func nullableText(value workflow.OptionalText) any {
	if value.Valid {
		return value.Text
	}
	return nil
}
func (m *snapshotCaptureMutation) VehicleIdentity(ctx context.Context, vehicleID string) (workflow.SnapshotVehicle, error) {
	var name, inventory string
	if err := m.tx.QueryRowContext(ctx, `SELECT name, inventory_number FROM vehicles WHERE id=?`, vehicleID).
		Scan(&name, &inventory); err != nil {
		return workflow.SnapshotVehicle{}, fmt.Errorf("read snapshot vehicle identity: %w", err)
	}
	return workflow.SnapshotVehicle{Name: name, InventoryNumber: inventory}, nil
}
func (m *snapshotCaptureMutation) Hardware(ctx context.Context, installationID, vehicleID string) (decoder.DecoderSnapshotHardware, error) {
	var hardware decoder.DecoderSnapshotHardware
	if err := m.tx.QueryRowContext(ctx, `SELECT d.id, d.label, d.manufacturer, d.model, d.serial_number,
i.id, i.state, i.installed_at, i.removed_at
FROM decoder_installations i JOIN decoder_instances d ON d.id=i.decoder_id
WHERE i.id=? AND i.vehicle_reference_id=?`, installationID, vehicleID).
		Scan(&hardware.DecoderID, &hardware.DecoderLabel, &hardware.Manufacturer, &hardware.Model,
			&hardware.SerialNumber, &hardware.InstallationID, &hardware.InstallationState, &hardware.InstalledAt, &hardware.RemovedAt); err != nil {
		return hardware, fmt.Errorf("capture confirmed decoder hardware: %w", err)
	}
	return hardware, nil
}
func (m *snapshotCaptureMutation) InsertPartialSnapshot(ctx context.Context, record workflow.SnapshotWrite) error {
	if _, err := m.tx.ExecContext(ctx, `INSERT INTO decoder_cv_snapshots
(id, workspace_id, vehicle_reference_id, vehicle_name, vehicle_inventory_number, profile_text,
installation_id, kind, scope, payload_json, hardware_json, note, recorded_at)
VALUES(?,?,?,?,?,?,?,?,'partial_working_state',?,?,?,?)`, record.ID, record.WorkspaceID, record.VehicleID,
		record.VehicleName, record.InventoryNumber, record.Profile, nullableText(record.InstallationID), record.Kind, record.Payload, record.Hardware,
		record.Note, record.RecordedAt); err != nil {
		return fmt.Errorf("record decoder working-state snapshot: %w", err)
	}
	return nil
}
func (m *snapshotCaptureMutation) InsertSnapshotSource(ctx context.Context, snapshotID string, source workflow.SnapshotSource) error {
	if _, err := m.tx.ExecContext(ctx, `INSERT INTO decoder_cv_snapshot_sources
(snapshot_id, source_file_id, metadata_json, content_state, blob_id) VALUES(?,?,?,?,?)`, snapshotID, source.ID,
		source.Metadata, source.State, nullableText(source.BlobID)); err != nil {
		return fmt.Errorf("record immutable decoder source: %w", err)
	}
	return nil
}
func (m *snapshotCaptureMutation) Values(ctx context.Context, workspace workflow.SnapshotWorkspace) ([]decoder.VehicleCVValue, error) {
	rows, err := m.tx.QueryContext(ctx, CVSelect+` WHERE vehicle_id=? AND decoder_profile=? ORDER BY id`,
		workspace.VehicleID, workspace.Profile)
	if err != nil {
		return nil, fmt.Errorf("read decoder snapshot values: %w", err)
	}
	defer func() { _ = rows.Close() }()
	records := []decoder.CVRecord{}
	for rows.Next() {
		record, err := ScanCV(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return []decoder.VehicleCVValue{}, nil
	}
	history, err := readCVHistory(ctx, m.tx, `cv_value_id IN (
SELECT id FROM vehicle_cv_values WHERE vehicle_id=? AND decoder_profile=?
)`, workspace.VehicleID, workspace.Profile)
	if err != nil {
		return nil, err
	}
	return workflow.CVValuesWithHistory(records, history), nil
}
