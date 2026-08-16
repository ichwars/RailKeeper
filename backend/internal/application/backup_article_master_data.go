package application

import (
	"context"
	"database/sql"
	"fmt"
)

type backupArticleMasterDataEntry struct {
	id           string
	typeName     string
	key          string
	label        string
	active       int
	sortOrder    int
	sourceURL    sql.NullString
	metadataJSON string
	createdAt    string
	updatedAt    string
	origin       string
}

func readLegacyRestoreArticleMasterData(
	ctx context.Context,
	tx *sql.Tx,
	backupVersion int,
) ([]backupArticleMasterDataEntry, error) {
	if backupVersion >= 3 {
		return nil, nil
	}
	rows, err := tx.QueryContext(ctx, `
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

	entries := []backupArticleMasterDataEntry{}
	articleTypeKeys := map[string]bool{}
	for rows.Next() {
		var entry backupArticleMasterDataEntry
		if err := rows.Scan(
			&entry.id, &entry.typeName, &entry.key, &entry.label, &entry.active, &entry.sortOrder,
			&entry.sourceURL, &entry.metadataJSON, &entry.createdAt, &entry.updatedAt, &entry.origin,
		); err != nil {
			return nil, fmt.Errorf("scan current article master data: %w", err)
		}
		entries = append(entries, entry)
		if entry.typeName == standardArticleType {
			articleTypeKeys[entry.key] = true
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate current article master data: %w", err)
	}
	if len(articleTypeKeys) != len(standardArticleTypeKeys) {
		return nil, ErrBackupInvalid
	}
	for _, key := range standardArticleTypeKeys {
		if !articleTypeKeys[key] {
			return nil, ErrBackupInvalid
		}
	}
	return entries, nil
}

func restoreLegacyArticleMasterData(
	ctx context.Context,
	tx *sql.Tx,
	entries []backupArticleMasterDataEntry,
) error {
	if len(entries) == 0 {
		return nil
	}
	if _, err := tx.ExecContext(ctx, `
DELETE FROM master_data_entries
WHERE type IN ('article_type', 'accessory_subtype', 'article_subtype')
`); err != nil {
		return fmt.Errorf("clear restored legacy article master data: %w", err)
	}
	for _, entry := range entries {
		if _, err := tx.ExecContext(ctx, `
INSERT INTO master_data_entries(
  id, type, key, label, active, sort_order, source_url, metadata_json,
  created_at, updated_at, origin
) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, entry.id, entry.typeName, entry.key, entry.label, entry.active, entry.sortOrder, entry.sourceURL,
			entry.metadataJSON, entry.createdAt, entry.updatedAt, entry.origin); err != nil {
			return fmt.Errorf("restore current article master data: %w", err)
		}
	}
	return nil
}
