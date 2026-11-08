package vehicle

import (
	"context"
	"strings"
)

type AttachmentMutation interface {
	MaintenanceLookup
	Attachment(context.Context, string, string) (*VehicleAttachment, error)
	UpdateAttachment(context.Context, string, string, VehicleAttachmentUpdateInput, string) error
	DeleteAttachment(context.Context, string, string) error
	Commit() error
	Rollback() error
}
type AttachmentRepository interface {
	MaintenanceLookup
	Attachments(context.Context, string) ([]VehicleAttachment, error)
	Attachment(context.Context, string, string) (*VehicleAttachment, error)
	InsertAttachment(context.Context, VehicleAttachment) error
	BeginAttachments(context.Context) (AttachmentMutation, error)
	FileReferenceCount(context.Context, string) (int, error)
}
type AttachmentService struct {
	repository AttachmentRepository
	parent     func(context.Context, string) error
}

func NewAttachmentService(repository AttachmentRepository, parent func(context.Context, string) error) *AttachmentService {
	return &AttachmentService{repository: repository, parent: parent}
}
func (s *AttachmentService) GetAttachment(ctx context.Context, id, attachmentID string) (*VehicleAttachment, error) {
	return s.repository.Attachment(ctx, id, attachmentID)
}
func (s *AttachmentService) Attachments(ctx context.Context, id string) ([]VehicleAttachment, error) {
	return s.repository.Attachments(ctx, id)
}
func (s *AttachmentService) MediaFileReferenceCount(ctx context.Context, path string) (int, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return 0, nil
	}
	return s.repository.FileReferenceCount(ctx, path)
}
