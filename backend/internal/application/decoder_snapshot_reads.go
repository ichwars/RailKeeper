package application

import (
	"context"

	"railkeeper/backend/internal/decoder/workflow"
	"railkeeper/backend/internal/infrastructure/vehicledata"
)

var ErrDecoderSourceUnavailable = workflow.ErrDecoderSourceUnavailable

type DecoderCVSnapshotSource = workflow.DecoderCVSnapshotSource
type DecoderCVSnapshot = workflow.DecoderCVSnapshot

func (s *VehicleService) snapshotReader() *workflow.SnapshotReader {
	return workflow.NewSnapshotReader(vehicledata.NewSnapshotStore(s.db))
}
func (s *VehicleService) ListDecoderCVSnapshots(ctx context.Context, vehicleID string) ([]DecoderCVSnapshotSummary, error) {
	return s.snapshotReader().List(ctx, vehicleID)
}
func (s *VehicleService) DecoderCVSnapshot(ctx context.Context, id string) (*DecoderCVSnapshot, error) {
	return s.snapshotReader().Snapshot(ctx, id)
}
func (s *VehicleService) DecoderCVSnapshotSourceContent(ctx context.Context, snapshotID, sourceID string) (*VehicleCVFile, []byte, error) {
	return s.snapshotReader().SourceContent(ctx, snapshotID, sourceID)
}
