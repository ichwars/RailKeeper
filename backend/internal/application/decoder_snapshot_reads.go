package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
)

var ErrDecoderSourceUnavailable = errors.New("recorded decoder source content is unavailable")

type DecoderCVSnapshotSource struct {
	SourceFileID string        `json:"sourceFileId"`
	Metadata     VehicleCVFile `json:"metadata"`
	ContentState string        `json:"contentState"`
}

type DecoderCVSnapshot struct {
	DecoderCVSnapshotSummary
	Payload  DecoderCVSnapshotPayload  `json:"payload"`
	Hardware *DecoderSnapshotHardware  `json:"hardware"`
	Sources  []DecoderCVSnapshotSource `json:"sources"`
}

// Historical evidence stays readable independently of the current vehicle or workspace.
func (s *VehicleService) ListDecoderCVSnapshots(ctx context.Context, vehicleID string) ([]DecoderCVSnapshotSummary, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	return readDecoderSnapshotSummaries(ctx, tx, vehicleID)
}

func (s *VehicleService) DecoderCVSnapshot(ctx context.Context, id string) (*DecoderCVSnapshot, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	result := &DecoderCVSnapshot{Sources: []DecoderCVSnapshotSource{}}
	var payload, hardware string
	err = tx.QueryRowContext(ctx, `SELECT id,workspace_id,vehicle_reference_id,vehicle_name,vehicle_inventory_number,
profile_text,installation_id,kind,scope,note,recorded_at,payload_json,hardware_json FROM decoder_cv_snapshots WHERE id=?`, id).
		Scan(&result.ID, &result.WorkspaceID, &result.VehicleReferenceID, &result.VehicleName, &result.VehicleInventoryNumber,
			&result.Profile, &result.InstallationID, &result.Kind, &result.Scope, &result.Note, &result.RecordedAt, &payload, &hardware)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrDecoderNotFound
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(payload), &result.Payload); err != nil {
		return nil, err
	}
	result.ValueCount = len(result.Payload.Values)
	var fields map[string]json.RawMessage
	if err := json.Unmarshal([]byte(hardware), &fields); err != nil {
		return nil, err
	}
	if len(fields) > 0 {
		result.Hardware = &DecoderSnapshotHardware{}
		if err := json.Unmarshal([]byte(hardware), result.Hardware); err != nil {
			return nil, err
		}
	}
	if err := readDecoderRows(ctx, tx, `SELECT source_file_id,metadata_json,content_state FROM decoder_cv_snapshot_sources WHERE snapshot_id=? ORDER BY source_file_id`, []any{id}, func(rows *sql.Rows) error {
		var item DecoderCVSnapshotSource
		var metadata string
		if err := rows.Scan(&item.SourceFileID, &metadata, &item.ContentState); err != nil {
			return err
		}
		if err := json.Unmarshal([]byte(metadata), &item.Metadata); err != nil {
			return err
		}
		result.Sources = append(result.Sources, item)
		return nil
	}); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *VehicleService) DecoderCVSnapshotSourceContent(ctx context.Context, snapshotID, sourceID string) (*VehicleCVFile, []byte, error) {
	var metadata, state string
	var compression sql.NullString
	var compressed []byte
	err := s.db.QueryRowContext(ctx, `SELECT s.metadata_json,s.content_state,b.compression,b.data
FROM decoder_cv_snapshot_sources s LEFT JOIN file_blobs b ON b.id=s.blob_id
WHERE s.snapshot_id=? AND s.source_file_id=?`, snapshotID, sourceID).Scan(&metadata, &state, &compression, &compressed)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil, ErrDecoderNotFound
	}
	if err != nil {
		return nil, nil, err
	}
	if state != "stored" || !compression.Valid {
		return nil, nil, ErrDecoderSourceUnavailable
	}
	var file VehicleCVFile
	if err := json.Unmarshal([]byte(metadata), &file); err != nil {
		return nil, nil, err
	}
	data, err := decompressBlobData(compressed, compression.String)
	if err != nil {
		return nil, nil, err
	}
	return &file, data, nil
}
