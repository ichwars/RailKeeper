package backupdata

import (
	"context"
	"fmt"

	"railkeeper/backend/internal/backup"
)

func (m *legacyRestoreStore) RestoreLegacyArticleMasterRows(
	ctx context.Context,
	entries []backup.StoredMasterRow,
) error {
	if len(entries) == 0 {
		return nil
	}
	if _, err := m.tx.ExecContext(ctx, `
DELETE FROM master_data_entries
WHERE type IN ('article_type', 'accessory_subtype', 'article_subtype')
`); err != nil {
		return fmt.Errorf("clear restored legacy article master data: %w", err)
	}
	for _, entry := range entries {
		if _, err := m.tx.ExecContext(ctx, `
INSERT INTO master_data_entries(
  id, type, key, label, active, sort_order, source_url, metadata_json,
  created_at, updated_at, origin
) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, entry.ID, entry.TypeName, entry.Key, entry.Label, entry.Active, entry.SortOrder, entry.SourceURL,
			entry.MetadataJSON, entry.CreatedAt, entry.UpdatedAt, entry.Origin); err != nil {
			return fmt.Errorf("restore current article master data: %w", err)
		}
	}
	return nil
}

func (m *legacyRestoreStore) RestoreBundledFunctionSymbols(
	ctx context.Context,
	symbols []backup.StoredMasterRow,
) error {
	for _, symbol := range symbols {
		if _, err := m.tx.ExecContext(ctx, `
INSERT INTO master_data_entries(
  id, type, key, label, active, sort_order, source_url, metadata_json,
  created_at, updated_at, origin
) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'bundled')
ON CONFLICT(type, key) DO UPDATE SET
  id=excluded.id,
  label=excluded.label,
  active=excluded.active,
  sort_order=excluded.sort_order,
  source_url=excluded.source_url,
  metadata_json=excluded.metadata_json,
  created_at=excluded.created_at,
  updated_at=excluded.updated_at,
  origin='bundled'`,
			symbol.ID,
			symbol.TypeName,
			symbol.Key,
			symbol.Label,
			symbol.Active,
			symbol.SortOrder,
			symbol.SourceURL,
			symbol.MetadataJSON,
			symbol.CreatedAt,
			symbol.UpdatedAt,
		); err != nil {
			return fmt.Errorf("restore bundled function symbol %s: %w", symbol.Key, err)
		}
	}
	return nil
}

func (m *legacyRestoreStore) ReconcileBundledMasterIdentities(
	ctx context.Context,
	identities []backup.BundledIdentity,
) error {
	for _, identity := range identities {
		if _, err := m.tx.ExecContext(ctx, `
UPDATE master_data_entries SET origin='bundled' WHERE type=? AND key=?
`, identity.TypeName, identity.Key); err != nil {
			return fmt.Errorf("reconcile bundled master data %s/%s: %w",
				identity.TypeName, identity.Key, err)
		}
	}
	return nil
}
