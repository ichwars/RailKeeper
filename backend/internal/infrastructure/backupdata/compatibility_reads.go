package backupdata

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"railkeeper/backend/internal/backup"
	"railkeeper/backend/internal/inventory/numbering"
)

func (m *legacyRestoreStore) LegacyArticleMasterRows(
	ctx context.Context,
) ([]backup.StoredMasterRow, error) {
	rows, err := m.tx.QueryContext(ctx, `
SELECT id, type, key, label, active, sort_order, source_url, metadata_json,
       created_at, updated_at, origin
FROM master_data_entries
WHERE type IN ('article_type', 'accessory_subtype')
ORDER BY type, key
`)
	if err != nil {
		return nil, fmt.Errorf("read current article master data: %w", err)
	}
	defer func() { _ = rows.Close() }()

	entries := []backup.StoredMasterRow{}
	for rows.Next() {
		var entry backup.StoredMasterRow
		if err := rows.Scan(
			&entry.ID, &entry.TypeName, &entry.Key, &entry.Label, &entry.Active, &entry.SortOrder,
			&entry.SourceURL, &entry.MetadataJSON, &entry.CreatedAt, &entry.UpdatedAt, &entry.Origin,
		); err != nil {
			return nil, fmt.Errorf("scan current article master data: %w", err)
		}
		entries = append(entries, entry)

	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate current article master data: %w", err)
	}
	return entries, nil
}

func (m *legacyRestoreStore) BundledMasterIdentities(
	ctx context.Context,
) ([]backup.BundledIdentity, error) {
	rows, err := m.tx.QueryContext(ctx, `
SELECT type, key FROM master_data_entries WHERE origin='bundled' ORDER BY type, key`)
	if err != nil {
		return nil, fmt.Errorf("read bundled master data identities: %w", err)
	}
	defer func() { _ = rows.Close() }()
	identities := []backup.BundledIdentity{}
	for rows.Next() {
		var identity backup.BundledIdentity
		if err := rows.Scan(&identity.TypeName, &identity.Key); err != nil {
			return nil, fmt.Errorf("scan bundled master data identity: %w", err)
		}
		identities = append(identities, identity)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate bundled master data identities: %w", err)
	}
	return identities, nil
}

func (m *legacyRestoreStore) BundledFunctionSymbols(
	ctx context.Context,
) ([]backup.StoredMasterRow, error) {
	rows, err := m.tx.QueryContext(ctx, `
SELECT id, type, key, label, active, sort_order, source_url, metadata_json,
       created_at, updated_at, origin
FROM master_data_entries
WHERE type='symbols'
  AND origin='bundled'
  AND json_extract(metadata_json, '$.library')='railkeeper-workshop-line'
ORDER BY key`)
	if err != nil {
		return nil, fmt.Errorf("read bundled function symbols: %w", err)
	}
	defer func() { _ = rows.Close() }()
	symbols := []backup.StoredMasterRow{}
	for rows.Next() {
		var symbol backup.StoredMasterRow
		if err := rows.Scan(
			&symbol.ID,
			&symbol.TypeName,
			&symbol.Key,
			&symbol.Label,
			&symbol.Active,
			&symbol.SortOrder,
			&symbol.SourceURL,
			&symbol.MetadataJSON,
			&symbol.CreatedAt,
			&symbol.UpdatedAt,
			&symbol.Origin,
		); err != nil {
			return nil, fmt.Errorf("scan bundled function symbol: %w", err)
		}
		symbols = append(symbols, symbol)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate bundled function symbols: %w", err)
	}
	return symbols, nil
}

func (m *legacyRestoreStore) ArticleScheme(
	ctx context.Context,
) (*backup.ArticleInventoryScheme, error) {
	var scheme backup.ArticleInventoryScheme
	var active int
	err := m.tx.QueryRowContext(ctx, `
SELECT id, prefix, next_number, padding, active, created_at, updated_at
FROM inventory_number_schemes
WHERE category='Artikel'
`).Scan(
		&scheme.ID,
		&scheme.Prefix,
		&scheme.NextNumber,
		&scheme.Padding,
		&active,
		&scheme.CreatedAt,
		&scheme.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read article inventory scheme before restore: %w", err)
	}
	scheme.Active = active == 1
	return &scheme, nil
}

func (m *legacyRestoreStore) VehicleSetScheme(ctx context.Context) (*numbering.Scheme, error) {
	var scheme numbering.Scheme
	var active int
	err := m.tx.QueryRowContext(ctx, `
SELECT id, category, prefix, next_number, padding, active, created_at, updated_at
FROM inventory_number_schemes
WHERE category='Set'
`).Scan(
		&scheme.ID,
		&scheme.Category,
		&scheme.Prefix,
		&scheme.NextNumber,
		&scheme.Padding,
		&active,
		&scheme.CreatedAt,
		&scheme.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, numbering.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("read inventory number scheme: %w", err)
	}
	scheme.Active = active == 1
	return &scheme, nil
}
