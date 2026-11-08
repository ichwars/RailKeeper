package workflow

import (
	"context"
	"railkeeper/backend/internal/decoder"
	"railkeeper/backend/internal/media"
)

type OptionalText struct {
	Text  string
	Valid bool
}
type SnapshotWorkspace struct {
	ID, VehicleID, Profile string
	InstallationID         OptionalText
}
type SnapshotCaptureInput struct{ ID, Kind, Note, RecordedAt string }
type SnapshotSource struct {
	ID, Metadata, State string
	BlobID              OptionalText
}
type SnapshotVehicle struct{ Name, InventoryNumber string }
type SnapshotWrite struct {
	ID, WorkspaceID, VehicleID, VehicleName, InventoryNumber, Profile string
	InstallationID                                                    OptionalText
	Kind, Payload, Hardware, Note, RecordedAt                         string
}
type OriginalCursor interface {
	Next() bool
	Blob() (media.StoredBlob, error)
	Err() error
	Close() error
}
type SnapshotCapture interface {
	Values(context.Context, SnapshotWorkspace) ([]decoder.VehicleCVValue, error)
	SourceIDs(context.Context, string, string) ([]string, error)
	SourceFile(context.Context, string) (*decoder.VehicleCVFile, bool, error)
	VehicleIdentity(context.Context, string) (SnapshotVehicle, error)
	Hardware(context.Context, string, string) (decoder.DecoderSnapshotHardware, error)
	InsertPartialSnapshot(context.Context, SnapshotWrite) error
	InsertSnapshotSource(context.Context, string, SnapshotSource) error
	ArchivedVariants(context.Context, string, string) (int, error)
	ArchivedOriginals(context.Context, string, string) (OriginalCursor, error)
	LatestArchivedSource(context.Context, string, string) (*SnapshotSource, error)
	ReadSourceBytes(string) ([]byte, error)
	StoreSourceBytes(context.Context, []byte) (string, error)
}
type LegacySnapshotMutation interface {
	SnapshotCapture
	AdoptLegacyGroups(context.Context, string) error
	PendingLegacyWorkspaces(context.Context) ([]SnapshotWorkspace, error)
	Commit() error
	Rollback() error
}
type LegacySnapshotRepository interface {
	BeginLegacy(context.Context) (LegacySnapshotMutation, error)
}
