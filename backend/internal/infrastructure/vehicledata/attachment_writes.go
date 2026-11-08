package vehicledata

import (
	"context"
	"fmt"
	"strings"

	vehiclemodel "railkeeper/backend/internal/inventory/vehicle"
)

func (s *attachmentStore) InsertAttachment(ctx context.Context, attachment vehiclemodel.VehicleAttachment) error {
	if _, err := s.db.ExecContext(ctx, `
INSERT INTO vehicle_attachments(id, vehicle_id, file_name, original_name, description, category, mime_type, size_bytes, storage_path, blob_id, maintenance_id, created_at, updated_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, attachment.ID, attachment.VehicleID, attachment.FileName, attachment.OriginalName, attachment.Description, attachment.Category, attachment.MimeType, attachment.SizeBytes, attachment.StoragePath, attachment.BlobID, attachment.MaintenanceID, attachment.CreatedAt, attachment.UpdatedAt); err != nil {
		return fmt.Errorf("create vehicle attachment: %w", err)
	}
	return nil
}

func (m *attachmentMutation) UpdateAttachment(ctx context.Context, vehicleID, attachmentID string, input vehiclemodel.VehicleAttachmentUpdateInput, now string) error {
	if _, err := m.tx.ExecContext(ctx, `
UPDATE vehicle_attachments SET description=?, category=?, maintenance_id=?, updated_at=? WHERE id=? AND vehicle_id=?
`, strings.TrimSpace(input.Description), strings.TrimSpace(input.Category), input.MaintenanceID,
		now, attachmentID, vehicleID); err != nil {
		return fmt.Errorf("update vehicle attachment: %w", err)
	}
	return nil
}

func (m *attachmentMutation) DeleteAttachment(ctx context.Context, vehicleID, attachmentID string) error {
	if _, err := m.tx.ExecContext(ctx, `DELETE FROM vehicle_attachments WHERE id=? AND vehicle_id=?`, attachmentID, vehicleID); err != nil {
		return fmt.Errorf("delete vehicle attachment: %w", err)
	}
	return nil
}
