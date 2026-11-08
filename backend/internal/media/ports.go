package media

import (
	"context"
	"errors"
)

var (
	ErrBlobNotFound  = errors.New("file blob not found")
	ErrSourceMissing = errors.New("legacy file source is missing")
)

type StoredBlob struct {
	Compression string
	Data        []byte
}
type LegacyTarget int

const (
	ImageOriginal LegacyTarget = iota
	ImageThumbnail
	VehicleAttachment
	VehicleCVSource
)

type LegacyPath struct{ RowID, Path string }
type Repository interface {
	Insert(context.Context, Blob) error
	Read(context.Context, string) (StoredBlob, error)
	DeleteIfUnreferenced(context.Context, string) error
	PendingLegacy(context.Context, LegacyTarget) ([]LegacyPath, error)
	BindLegacy(context.Context, LegacyTarget, string, string) error
}
type LegacyFiles interface{ Read(string) ([]byte, error) }
type Service struct {
	repository Repository
	files      LegacyFiles
}

func New(repository Repository, files LegacyFiles) *Service {
	return &Service{repository: repository, files: files}
}
