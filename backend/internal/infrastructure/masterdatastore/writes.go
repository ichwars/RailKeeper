package masterdatastore

import (
	"context"
	"fmt"
	"time"

	"railkeeper/backend/internal/inventory/masterdata"
)

var _ masterdata.Repository = (*Store)(nil)

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
func (s *Store) Create(ctx context.Context, mutation masterdata.Mutation) error {
	entry := mutation.Entry
	now := time.Now().UTC().Format(time.RFC3339)
	if _, err := s.db.ExecContext(ctx, `
INSERT INTO master_data_entries(
  id, type, key, label, active, sort_order, source_url, metadata_json,
  created_at, updated_at, origin
)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'custom')
`, entry.ID, entry.Type, entry.Key, entry.Label, boolToInt(entry.Active), entry.SortOrder,
		entry.SourceURL, string(mutation.MetadataJSON), now, now); err != nil {
		return fmt.Errorf("create master data: %w", err)
	}
	return nil
}

func (s *Store) Update(ctx context.Context, mutation masterdata.Mutation) error {
	entry := mutation.Entry
	result, err := s.db.ExecContext(ctx, `
UPDATE master_data_entries
SET label=?, active=?, sort_order=?, source_url=?, metadata_json=?, updated_at=?
WHERE type=? AND key=?
`, entry.Label, boolToInt(entry.Active), entry.SortOrder, entry.SourceURL,
		string(mutation.MetadataJSON), time.Now().UTC().Format(time.RFC3339), entry.Type, entry.Key)
	if err != nil {
		return fmt.Errorf("update master data: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read master data update result: %w", err)
	}
	if affected == 0 {
		return masterdata.ErrMasterDataNotFound
	}
	return nil
}
