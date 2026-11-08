package media

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

func (s *Service) Store(ctx context.Context, data []byte) (string, error) {
	if s == nil || s.repository == nil {
		return "", errors.New("file blob service is not configured")
	}
	blob, err := NewBlob(data)
	if err != nil {
		return "", err
	}
	if err := s.repository.Insert(ctx, blob); err != nil {
		return "", err
	}
	return blob.ID, nil
}
func (s *Service) Load(ctx context.Context, id string) ([]byte, error) {
	if s == nil || s.repository == nil {
		return nil, errors.New("file blob service is not configured")
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrBlobNotFound
	}
	stored, err := s.repository.Read(ctx, id)
	if err != nil {
		return nil, err
	}
	data, err := Decode(stored.Data, stored.Compression)
	if err != nil {
		return nil, fmt.Errorf("load file blob: %w", err)
	}
	return data, nil
}
func (s *Service) DeleteIfUnreferenced(ctx context.Context, id string) error {
	if s == nil || s.repository == nil {
		return nil
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return nil
	}
	return s.repository.DeleteIfUnreferenced(ctx, id)
}
func (s *Service) MigrateFilesystemBlobs(ctx context.Context) error {
	if s == nil || s.repository == nil || s.files == nil {
		return nil
	}
	for _, target := range []LegacyTarget{ImageOriginal, ImageThumbnail, VehicleAttachment, VehicleCVSource} {
		if err := s.migratePaths(ctx, target); err != nil {
			return err
		}
	}
	return nil
}
func (s *Service) migratePaths(ctx context.Context, target LegacyTarget) error {
	pending, err := s.repository.PendingLegacy(ctx, target)
	if err != nil {
		return err
	}
	blobByPath := map[string]string{}
	for _, item := range pending {
		blobID := blobByPath[item.Path]
		if blobID == "" {
			data, err := s.files.Read(item.Path)
			if errors.Is(err, ErrSourceMissing) {
				continue
			}
			if err != nil {
				return err
			}
			blobID, err = s.Store(ctx, data)
			if err != nil {
				return err
			}
			blobByPath[item.Path] = blobID
		}
		if err := s.repository.BindLegacy(ctx, target, item.RowID, blobID); err != nil {
			return err
		}
	}
	return nil
}
