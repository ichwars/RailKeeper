package vehicledata

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
)

func (s *attachmentStore) FileReferenceCount(ctx context.Context, storagePath string) (int, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT storage_path FROM vehicle_images UNION ALL SELECT thumbnail_path FROM vehicle_images
UNION ALL SELECT storage_path FROM vehicle_attachments UNION ALL SELECT storage_path FROM vehicle_cv_files
`)
	if err != nil {
		return 0, fmt.Errorf("list media file references: %w", err)
	}
	defer func() { _ = rows.Close() }()
	target := filepath.Clean(filepath.FromSlash(storagePath))
	count := 0
	for rows.Next() {
		var stored sql.NullString
		if err := rows.Scan(&stored); err != nil {
			return 0, fmt.Errorf("scan media file reference: %w", err)
		}
		if stored.Valid && stored.String != "" {
			relative, err := filepath.Rel(target, filepath.Clean(filepath.FromSlash(stored.String)))
			if err == nil && relative == "." {
				count++
			}
		}
	}
	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("iterate media file references: %w", err)
	}
	return count, nil
}
