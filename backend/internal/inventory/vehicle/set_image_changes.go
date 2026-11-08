package vehicle

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func (s *SetImageService) SetSetMainImage(
	ctx context.Context,
	setID string,
	input VehicleSetMainImageInput,
	actorUserID string,
) (*VehicleSet, error) {
	setID = strings.TrimSpace(setID)
	input.MemberImageID = strings.TrimSpace(input.MemberImageID)
	if setID == "" {
		return nil, ErrVehicleSetNotFound
	}
	if input.Mode != VehicleSetMainImageModeAutomatic && input.Mode != VehicleSetMainImageModeMember &&
		input.Mode != VehicleSetMainImageModeDedicated {
		return nil, ErrVehicleSetImageValidation
	}
	tx, err := s.repository.BeginSetImages(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin vehicle set main image update: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	setExists, err := tx.Exists(ctx, setID)
	if err != nil {
		return nil, err
	}
	if !setExists {
		return nil, ErrVehicleSetNotFound
	}
	if input.Mode == VehicleSetMainImageModeMember {
		memberBelongs, err := tx.MemberBelongs(ctx, setID, input.MemberImageID)
		if err != nil {
			return nil, err
		}
		if !memberBelongs {
			return nil, ErrVehicleSetImageValidation
		}
	}
	if input.Mode == VehicleSetMainImageModeDedicated {
		blobID, err := tx.DedicatedBlob(ctx, setID)
		if err != nil {
			return nil, err
		}
		if blobID == "" {
			return nil, ErrVehicleSetImageValidation
		}
	}
	if input.Mode != VehicleSetMainImageModeMember {
		input.MemberImageID = ""
	}
	now := time.Now().UTC().Format(time.RFC3339)
	if err := tx.UpdateSelection(ctx, setID, input, now); err != nil {
		return nil, err
	}
	if err := auditSetImage(ctx, tx, actorUserID, setID, "VehicleSetMainImageChanged", now); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit vehicle set main image update: %w", err)
	}
	return s.readSet(ctx, setID)
}
func (s *SetImageService) UpsertSetImage(
	ctx context.Context,
	setID string,
	input VehicleSetImageInput,
	actorUserID string,
) (*VehicleSet, []string, error) {
	setID = strings.TrimSpace(setID)
	input.FileName = strings.TrimSpace(input.FileName)
	input.MimeType = strings.TrimSpace(input.MimeType)
	input.BlobID = strings.TrimSpace(input.BlobID)
	input.ThumbnailBlobID = strings.TrimSpace(input.ThumbnailBlobID)
	if setID == "" {
		return nil, nil, ErrVehicleSetNotFound
	}
	if input.FileName == "" || input.MimeType == "" || input.BlobID == "" {
		return nil, nil, ErrVehicleSetImageValidation
	}
	tx, err := s.repository.BeginSetImages(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("begin vehicle set image upload: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	previous, err := tx.PreviousForUpload(ctx, setID)
	if err != nil {
		return nil, nil, err
	}
	oldBlobID, oldThumbnailBlobID, createdAt := previous.BlobID, previous.ThumbnailBlobID, previous.CreatedAt
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if createdAt == "" {
		createdAt = now
	}
	if err := tx.StoreDedicated(ctx, setID, input, createdAt, now); err != nil {
		return nil, nil, err
	}
	if err := auditSetImage(ctx, tx, actorUserID, setID, "VehicleSetImageUploaded", now); err != nil {
		return nil, nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, nil, fmt.Errorf("commit vehicle set image upload: %w", err)
	}
	replaced := distinctSetImageIDs(oldBlobID, oldThumbnailBlobID)
	set, err := s.readSet(ctx, setID)
	if err != nil {
		return nil, replaced, err
	}
	return set, replaced, nil
}
func (s *SetImageService) DeleteSetImage(ctx context.Context, setID, actorUserID string) ([]string, error) {
	setID = strings.TrimSpace(setID)
	tx, err := s.repository.BeginSetImages(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin vehicle set image deletion: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	previous, err := tx.PreviousForDelete(ctx, setID)
	if err != nil {
		return nil, err
	}
	blobID, thumbnailBlobID := previous.BlobID, previous.ThumbnailBlobID
	now := time.Now().UTC().Format(time.RFC3339)
	if err := tx.DeleteDedicated(ctx, setID, now); err != nil {
		return nil, err
	}
	if err := auditSetImage(ctx, tx, actorUserID, setID, "VehicleSetImageDeleted", now); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit vehicle set image deletion: %w", err)
	}
	return distinctSetImageIDs(blobID, thumbnailBlobID), nil
}
