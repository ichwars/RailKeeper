package vehicle

import (
	"context"
	"fmt"
	"strings"
	"time"

	"railkeeper/backend/internal/credentials"
)

func (s *AttachmentService) CreateAttachment(ctx context.Context, vehicleID string, input VehicleAttachmentInput) (*VehicleAttachment, error) {
	vehicleID = strings.TrimSpace(vehicleID)
	input = CleanVehicleAttachmentInput(input)
	if vehicleID == "" || input.FileName == "" || input.OriginalName == "" || (input.StoragePath == "" && input.BlobID == "") {
		return nil, ErrVehicleValidation
	}
	if err := s.parent(ctx, vehicleID); err != nil {
		return nil, err
	}
	if err := EnsureMaintenanceReference(ctx, s.repository, vehicleID, input.MaintenanceID); err != nil {
		return nil, err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	attachment := VehicleAttachment{
		ID:            credentials.RandomID(),
		VehicleID:     vehicleID,
		FileName:      input.FileName,
		OriginalName:  input.OriginalName,
		Description:   input.Description,
		Category:      input.Category,
		MimeType:      input.MimeType,
		SizeBytes:     input.SizeBytes,
		StoragePath:   input.StoragePath,
		BlobID:        input.BlobID,
		MaintenanceID: input.MaintenanceID,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := s.repository.InsertAttachment(ctx, attachment); err != nil {
		return nil, err
	}
	attachment.Revision = AttachmentRevision(attachment)
	return &attachment, nil
}
func (s *AttachmentService) UpdateAttachment(ctx context.Context, vehicleID, attachmentID string, input VehicleAttachmentUpdateInput) (*VehicleAttachment, error) {
	tx, err := s.repository.BeginAttachments(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin attachment update: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	current, err := tx.Attachment(ctx, vehicleID, attachmentID)
	if err != nil {
		return nil, err
	}
	if err := CheckRevision(current.Revision, input.ExpectedRevision, ErrVehicleAttachmentConflict); err != nil {
		return nil, err
	}
	input.MaintenanceID = strings.TrimSpace(input.MaintenanceID)
	if err := EnsureMaintenanceReference(ctx, tx, vehicleID, input.MaintenanceID); err != nil {
		return nil, err
	}
	if err := tx.UpdateAttachment(ctx, current.VehicleID, current.ID, input, time.Now().UTC().Format(time.RFC3339)); err != nil {
		return nil, err
	}
	updated, err := tx.Attachment(ctx, vehicleID, attachmentID)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit attachment update: %w", err)
	}
	return updated, nil
}
func (s *AttachmentService) DeleteAttachment(ctx context.Context, vehicleID, attachmentID string, expectedRevision ...string) (*VehicleAttachment, error) {
	tx, err := s.repository.BeginAttachments(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin attachment delete: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	attachment, err := tx.Attachment(ctx, vehicleID, attachmentID)
	if err != nil {
		return nil, err
	}
	if err := CheckRevision(attachment.Revision, OptionalRevision(expectedRevision), ErrVehicleAttachmentConflict); err != nil {
		return nil, err
	}
	if err := tx.DeleteAttachment(ctx, attachment.VehicleID, attachment.ID); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit attachment delete: %w", err)
	}
	return attachment, nil
}
