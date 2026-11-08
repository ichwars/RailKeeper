package workflow

import (
	"context"
	"errors"

	"railkeeper/backend/internal/decoder"
)

var ErrDecoderSourceUnavailable = errors.New("recorded decoder source content is unavailable")

type DecoderCVSnapshotSource struct {
	SourceFileID string                `json:"sourceFileId"`
	Metadata     decoder.VehicleCVFile `json:"metadata"`
	ContentState string                `json:"contentState"`
}

type DecoderCVSnapshot struct {
	decoder.DecoderCVSnapshotSummary
	Payload  decoder.DecoderCVSnapshotPayload `json:"payload"`
	Hardware *decoder.DecoderSnapshotHardware `json:"hardware"`
	Sources  []DecoderCVSnapshotSource        `json:"sources"`
}

type SnapshotRecord struct {
	Summary                   decoder.DecoderCVSnapshotSummary
	PayloadJSON, HardwareJSON string
}
type SnapshotSourceRecord struct{ SourceFileID, MetadataJSON, ContentState string }
type SnapshotContent struct {
	MetadataJSON, State, Compression string
	HasCompression                   bool
	Data                             []byte
}
type SnapshotReadTransaction interface {
	Summaries(context.Context, string) ([]decoder.DecoderCVSnapshotSummary, error)
	Snapshot(context.Context, string) (SnapshotRecord, bool, error)
	Sources(context.Context, string) ([]SnapshotSourceRecord, error)
	Rollback() error
}
type SnapshotRepository interface {
	BeginSnapshots(context.Context) (SnapshotReadTransaction, error)
	SourceContent(context.Context, string, string) (SnapshotContent, bool, error)
}
type SnapshotReader struct{ repository SnapshotRepository }

func NewSnapshotReader(repository SnapshotRepository) *SnapshotReader {
	return &SnapshotReader{repository: repository}
}
