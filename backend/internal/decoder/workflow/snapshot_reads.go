package workflow

import (
	"context"
	"encoding/json"
	"fmt"

	"railkeeper/backend/internal/decoder"
	"railkeeper/backend/internal/media"
)

// Historical evidence remains readable independently of current vehicles or workspaces.
func (s *SnapshotReader) List(ctx context.Context, vehicleID string) ([]decoder.DecoderCVSnapshotSummary, error) {
	tx, err := s.repository.BeginSnapshots(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	return tx.Summaries(ctx, vehicleID)
}
func (s *SnapshotReader) Snapshot(ctx context.Context, id string) (*DecoderCVSnapshot, error) {
	tx, err := s.repository.BeginSnapshots(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	record, found, err := tx.Snapshot(ctx, id)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, decoder.ErrNotFound
	}
	result := &DecoderCVSnapshot{DecoderCVSnapshotSummary: record.Summary, Sources: []DecoderCVSnapshotSource{}}
	if err := json.Unmarshal([]byte(record.PayloadJSON), &result.Payload); err != nil {
		return nil, err
	}
	result.ValueCount = len(result.Payload.Values)
	var fields map[string]json.RawMessage
	if err := json.Unmarshal([]byte(record.HardwareJSON), &fields); err != nil {
		return nil, err
	}
	if len(fields) > 0 {
		result.Hardware = &decoder.DecoderSnapshotHardware{}
		if err := json.Unmarshal([]byte(record.HardwareJSON), result.Hardware); err != nil {
			return nil, err
		}
	}
	sources, err := tx.Sources(ctx, id)
	if err != nil {
		return nil, err
	}
	for _, source := range sources {
		item := DecoderCVSnapshotSource{SourceFileID: source.SourceFileID, ContentState: source.ContentState}
		if err := json.Unmarshal([]byte(source.MetadataJSON), &item.Metadata); err != nil {
			return nil, fmt.Errorf("scan decoder configuration: %w", err)
		}
		result.Sources = append(result.Sources, item)
	}
	return result, nil
}
func (s *SnapshotReader) SourceContent(ctx context.Context, snapshotID, sourceID string) (*decoder.VehicleCVFile, []byte, error) {
	stored, found, err := s.repository.SourceContent(ctx, snapshotID, sourceID)
	if err != nil {
		return nil, nil, err
	}
	if !found {
		return nil, nil, decoder.ErrNotFound
	}
	if stored.State != "stored" || !stored.HasCompression {
		return nil, nil, ErrDecoderSourceUnavailable
	}
	var file decoder.VehicleCVFile
	if err := json.Unmarshal([]byte(stored.MetadataJSON), &file); err != nil {
		return nil, nil, err
	}
	data, err := media.Decode(stored.Data, stored.Compression)
	if err != nil {
		return nil, nil, err
	}
	return &file, data, nil
}
