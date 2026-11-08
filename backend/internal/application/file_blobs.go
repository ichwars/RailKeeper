package application

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"railkeeper/backend/internal/infrastructure/blobdata"
	"railkeeper/backend/internal/media"
)

type FileBlob = media.Blob
type FileBlobService struct {
	service *media.Service
}

func NewFileBlobService(db *sql.DB, dataDir string) *FileBlobService {
	var repository media.Repository
	if db != nil {
		repository = blobdata.New(db)
	}
	var files media.LegacyFiles
	if strings.TrimSpace(dataDir) != "" {
		files = blobdata.NewLegacyFiles(dataDir)
	}
	return &FileBlobService{service: media.New(repository, files)}
}
func (s *FileBlobService) Store(ctx context.Context, data []byte) (string, error) {
	if s == nil {
		return "", errors.New("file blob service is not configured")
	}
	return s.service.Store(ctx, data)
}
func (s *FileBlobService) Load(ctx context.Context, id string) ([]byte, error) {
	if s == nil {
		return nil, errors.New("file blob service is not configured")
	}
	data, err := s.service.Load(ctx, id)
	if errors.Is(err, media.ErrBlobNotFound) {
		return nil, sql.ErrNoRows
	}
	return data, err
}
func (s *FileBlobService) DeleteIfUnreferenced(ctx context.Context, id string) error {
	if s == nil {
		return nil
	}
	return s.service.DeleteIfUnreferenced(ctx, id)
}
func (s *FileBlobService) MigrateFilesystemBlobs(ctx context.Context) error {
	if s == nil {
		return nil
	}
	return s.service.MigrateFilesystemBlobs(ctx)
}
