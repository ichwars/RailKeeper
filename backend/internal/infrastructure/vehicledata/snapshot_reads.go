package vehicledata

import (
	"context"
	"database/sql"
	"errors"

	"railkeeper/backend/internal/decoder"
	"railkeeper/backend/internal/decoder/workflow"
)

type snapshotStore struct{ db *sql.DB }
type snapshotReadTransaction struct{ tx *sql.Tx }

func NewSnapshotStore(db *sql.DB) workflow.SnapshotRepository { return &snapshotStore{db: db} }
func (s *snapshotStore) BeginSnapshots(ctx context.Context) (workflow.SnapshotReadTransaction, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &snapshotReadTransaction{tx: tx}, nil
}
func (m *snapshotReadTransaction) Rollback() error { return m.tx.Rollback() }
func (m *snapshotReadTransaction) Summaries(ctx context.Context, vehicleID string) ([]decoder.DecoderCVSnapshotSummary, error) {
	return ReadDecoderSnapshotSummaries(ctx, m.tx, vehicleID)
}
func (m *snapshotReadTransaction) Snapshot(ctx context.Context, id string) (workflow.SnapshotRecord, bool, error) {
	var record workflow.SnapshotRecord
	err := m.tx.QueryRowContext(ctx, `SELECT id,workspace_id,vehicle_reference_id,vehicle_name,vehicle_inventory_number,
profile_text,installation_id,kind,scope,note,recorded_at,payload_json,hardware_json FROM decoder_cv_snapshots WHERE id=?`, id).
		Scan(&record.Summary.ID, &record.Summary.WorkspaceID, &record.Summary.VehicleReferenceID, &record.Summary.VehicleName, &record.Summary.VehicleInventoryNumber,
			&record.Summary.Profile, &record.Summary.InstallationID, &record.Summary.Kind, &record.Summary.Scope, &record.Summary.Note, &record.Summary.RecordedAt, &record.PayloadJSON, &record.HardwareJSON)
	if errors.Is(err, sql.ErrNoRows) {
		return record, false, nil
	}
	return record, err == nil, err
}
func (m *snapshotReadTransaction) Sources(ctx context.Context, id string) ([]workflow.SnapshotSourceRecord, error) {
	result := []workflow.SnapshotSourceRecord{}
	err := ReadDecoderRows(ctx, m.tx, `SELECT source_file_id,metadata_json,content_state FROM decoder_cv_snapshot_sources WHERE snapshot_id=? ORDER BY source_file_id`, []any{id}, func(rows *sql.Rows) error {
		var item workflow.SnapshotSourceRecord
		if err := rows.Scan(&item.SourceFileID, &item.MetadataJSON, &item.ContentState); err != nil {
			return err
		}
		result = append(result, item)
		return nil
	})
	return result, err
}
func (s *snapshotStore) SourceContent(ctx context.Context, snapshotID, sourceID string) (workflow.SnapshotContent, bool, error) {
	var metadata, state string
	var compression sql.NullString
	var compressed []byte
	err := s.db.QueryRowContext(ctx, `SELECT s.metadata_json,s.content_state,b.compression,b.data
FROM decoder_cv_snapshot_sources s LEFT JOIN file_blobs b ON b.id=s.blob_id
WHERE s.snapshot_id=? AND s.source_file_id=?`, snapshotID, sourceID).Scan(&metadata, &state, &compression, &compressed)
	if errors.Is(err, sql.ErrNoRows) {
		return workflow.SnapshotContent{}, false, nil
	}
	return workflow.SnapshotContent{MetadataJSON: metadata, State: state, Compression: compression.String, HasCompression: compression.Valid, Data: compressed}, err == nil, err
}
