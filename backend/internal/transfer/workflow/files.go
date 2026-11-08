package workflow

import (
	"context"
	"io"
	"time"
)

type UploadMetadata struct{ Name, SHA256, DetectedMIME string }
type Upload interface {
	io.Reader
	io.Closer
	Metadata() UploadMetadata
}
type ArtifactMetadata struct {
	Size       int64
	ModifiedAt time.Time
}
type ArtifactReader interface {
	io.Reader
	io.Seeker
	io.Closer
	Metadata() (ArtifactMetadata, error)
}

// ArtifactRecorder is deliberately narrower than the optional export repository.
type ArtifactRecorder interface {
	CreateArtifact(context.Context, DataTransferArtifact) (DataTransferArtifact, error)
}
type FileStore interface {
	StageUpload(context.Context, string, TransferFormat, string, string, io.Reader) (Upload, error)
	WriteExport(context.Context, ArtifactRecorder, string, string, string, []byte) (DataTransferArtifact, error)
	OpenArtifact(string) (ArtifactReader, error)
	RemoveArtifact(string) error
	ExportDirectory() (string, error)
}
