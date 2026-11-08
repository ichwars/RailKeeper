package vehicledata

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	vehiclemodel "railkeeper/backend/internal/inventory/vehicle"
)

type attachmentStore struct{ db *sql.DB }
type attachmentMutation struct{ tx *sql.Tx }

func NewAttachmentStore(db *sql.DB) vehiclemodel.AttachmentRepository {
	return &attachmentStore{db: db}
}
func (s *attachmentStore) BeginAttachments(ctx context.Context) (vehiclemodel.AttachmentMutation, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &attachmentMutation{tx: tx}, nil
}
func (m *attachmentMutation) Commit() error   { return m.tx.Commit() }
func (m *attachmentMutation) Rollback() error { return m.tx.Rollback() }
func (s *attachmentStore) MaintenanceBelongs(ctx context.Context, id, maintenanceID string) (bool, error) {
	return MaintenanceBelongsToVehicle(ctx, s.db, id, maintenanceID)
}
func (m *attachmentMutation) MaintenanceBelongs(ctx context.Context, id, maintenanceID string) (bool, error) {
	return MaintenanceBelongsToVehicle(ctx, m.tx, id, maintenanceID)
}
func (s *attachmentStore) Attachment(ctx context.Context, id, attachmentID string) (*vehiclemodel.VehicleAttachment, error) {
	return ReadAttachment(ctx, s.db, id, attachmentID)
}
func (m *attachmentMutation) Attachment(ctx context.Context, id, attachmentID string) (*vehiclemodel.VehicleAttachment, error) {
	return ReadAttachment(ctx, m.tx, id, attachmentID)
}

func (s *attachmentStore) Attachments(ctx context.Context, vehicleID string) ([]vehiclemodel.VehicleAttachment, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, vehicle_id, file_name, original_name, COALESCE(description, ''), COALESCE(category, ''),
       COALESCE(mime_type, ''), size_bytes, storage_path, COALESCE(blob_id, ''), COALESCE(maintenance_id, ''), created_at, updated_at
FROM vehicle_attachments
WHERE vehicle_id=?
ORDER BY created_at ASC
`, vehicleID)
	if err != nil {
		return nil, fmt.Errorf("list vehicle attachments: %w", err)
	}
	defer func() { _ = rows.Close() }()

	attachments := []vehiclemodel.VehicleAttachment{}
	for rows.Next() {
		var attachment vehiclemodel.VehicleAttachment
		if err := rows.Scan(
			&attachment.ID,
			&attachment.VehicleID,
			&attachment.FileName,
			&attachment.OriginalName,
			&attachment.Description,
			&attachment.Category,
			&attachment.MimeType,
			&attachment.SizeBytes,
			&attachment.StoragePath,
			&attachment.BlobID,
			&attachment.MaintenanceID,
			&attachment.CreatedAt,
			&attachment.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan vehicle attachment: %w", err)
		}
		attachment.Revision = vehiclemodel.AttachmentRevision(attachment)
		attachments = append(attachments, attachment)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate vehicle attachments: %w", err)
	}
	return attachments, nil
}
func ReadAttachment(ctx context.Context, query queryRower, vehicleID, attachmentID string) (*vehiclemodel.VehicleAttachment, error) {
	var attachment vehiclemodel.VehicleAttachment
	err := query.QueryRowContext(ctx, `
SELECT id, vehicle_id, file_name, original_name, COALESCE(description, ''), COALESCE(category, ''),
       COALESCE(mime_type, ''), size_bytes, storage_path, COALESCE(blob_id, ''), COALESCE(maintenance_id, ''), created_at, updated_at
FROM vehicle_attachments
WHERE id=? AND vehicle_id=?
`, strings.TrimSpace(attachmentID), strings.TrimSpace(vehicleID)).Scan(
		&attachment.ID,
		&attachment.VehicleID,
		&attachment.FileName,
		&attachment.OriginalName,
		&attachment.Description,
		&attachment.Category,
		&attachment.MimeType,
		&attachment.SizeBytes,
		&attachment.StoragePath,
		&attachment.BlobID,
		&attachment.MaintenanceID,
		&attachment.CreatedAt,
		&attachment.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, vehiclemodel.ErrVehicleNotFound
		}
		return nil, fmt.Errorf("get vehicle attachment: %w", err)
	}
	attachment.Revision = vehiclemodel.AttachmentRevision(attachment)
	return &attachment, nil
}
