package masterdatastore

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"railkeeper/backend/internal/inventory/masterdata"
)

func (s *Store) ImportReconciled(ctx context.Context, request masterdata.ImportRequest) (*masterdata.MasterDataImportResult, error) {
	doc := request.Document
	entriesByType := request.EntriesByType
	articleTypesWereImported := request.ArticleTypesWereImported
	accessorySubtypesWereImported := request.AccessorySubtypesWereImported

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin master data import: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := reserveMasterDataWriteTransaction(ctx, tx); err != nil {
		return nil, err
	}

	current, err := loadCurrentMasterData(ctx, tx)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	desired, importedEntries, err := masterdata.BuildDesiredEntries(
		ctx,
		func(ctx context.Context, entries []masterdata.MasterDataEntry) (map[masterdata.Identity]bool, error) {
			return readMasterDataUsage(ctx, tx, entries)
		},
		current,
		entriesByType,
		now,
		articleTypesWereImported,
		accessorySubtypesWereImported,
	)
	if err != nil {
		return nil, err
	}
	desiredByType := masterdata.EntriesByType(desired)
	if err := masterdata.ValidateImportedAccessoryCustomFieldReferences(ctx, func(ctx context.Context, key string) ([]masterdata.ReferencedAttribute, error) {
		return loadReferencedAccessoryCustomAttributes(ctx, tx, key)
	}, desiredByType); err != nil {
		return nil, err
	}

	currentRelations, err := loadCurrentMasterDataRelations(ctx, tx)
	if err != nil {
		return nil, err
	}
	desiredRelations, importedRelations, err := masterdata.BuildDesiredRelations(
		doc.Relations,
		currentRelations,
		current,
		desired,
	)
	if err != nil {
		return nil, err
	}
	if err := replaceMasterData(ctx, tx, desired, desiredRelations, now); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit master data import: %w", err)
	}
	return &masterdata.MasterDataImportResult{
		ImportedTypes:     len(doc.Entries),
		ImportedEntries:   importedEntries,
		ImportedRelations: importedRelations,
	}, nil
}
func loadCurrentMasterData(
	ctx context.Context,
	tx *sql.Tx,
) (map[masterdata.Identity]masterdata.MasterDataEntry, error) {
	rows, err := tx.QueryContext(ctx, `
SELECT id, type, key, label, active, sort_order, COALESCE(source_url, ''), metadata_json,
       created_at, updated_at, origin
FROM master_data_entries`)
	if err != nil {
		return nil, fmt.Errorf("load current master data: %w", err)
	}
	defer func() { _ = rows.Close() }()
	entries := map[masterdata.Identity]masterdata.MasterDataEntry{}
	for rows.Next() {
		entry, err := ScanEntry(rows)
		if err != nil {
			return nil, err
		}
		entries[masterdata.IdentityOf(entry)] = entry
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate current master data: %w", err)
	}
	return entries, nil
}
func loadCurrentMasterDataRelations(
	ctx context.Context,
	tx *sql.Tx,
) ([]masterdata.MasterDataRelation, error) {
	rows, err := tx.QueryContext(ctx, `
SELECT id, parent_type, parent_key, child_type, child_key, sort_order
FROM master_data_relations`)
	if err != nil {
		return nil, fmt.Errorf("load current master data relations: %w", err)
	}
	defer func() { _ = rows.Close() }()
	relations := []masterdata.MasterDataRelation{}
	for rows.Next() {
		var relation masterdata.MasterDataRelation
		if err := rows.Scan(
			&relation.ID,
			&relation.ParentType,
			&relation.ParentKey,
			&relation.ChildType,
			&relation.ChildKey,
			&relation.SortOrder,
		); err != nil {
			return nil, fmt.Errorf("scan current master data relation: %w", err)
		}
		relations = append(relations, relation)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate current master data relations: %w", err)
	}
	return relations, nil
}
func replaceMasterData(
	ctx context.Context,
	tx *sql.Tx,
	entries map[masterdata.Identity]masterdata.MasterDataEntry,
	relations []masterdata.MasterDataRelation,
	now string,
) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM master_data_relations`); err != nil {
		return fmt.Errorf("clear master data relations: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM master_data_entries`); err != nil {
		return fmt.Errorf("clear master data entries: %w", err)
	}
	ordered := make([]masterdata.MasterDataEntry, 0, len(entries))
	for _, entry := range entries {
		ordered = append(ordered, entry)
	}
	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].Type != ordered[j].Type {
			return ordered[i].Type < ordered[j].Type
		}
		return ordered[i].Key < ordered[j].Key
	})
	for _, entry := range ordered {

		metadataJSON, err := json.Marshal(entry.Metadata)
		if err != nil {
			return fmt.Errorf("marshal imported master data metadata: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO master_data_entries(
  id, type, key, label, active, sort_order, source_url, metadata_json,
  created_at, updated_at, origin
)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, entry.ID, entry.Type, entry.Key, entry.Label, boolToInt(entry.Active), entry.SortOrder,
			entry.SourceURL, string(metadataJSON), entry.CreatedAt, entry.UpdatedAt, entry.Origin); err != nil {
			return fmt.Errorf("insert imported master data entry: %w", err)
		}
	}
	for _, relation := range relations {
		if _, err := tx.ExecContext(ctx, `
INSERT INTO master_data_relations(
  id, parent_type, parent_key, child_type, child_key, sort_order, created_at
)
VALUES(?, ?, ?, ?, ?, ?, ?)
`, relation.ID, relation.ParentType, relation.ParentKey, relation.ChildType,
			relation.ChildKey, relation.SortOrder, now); err != nil {
			return fmt.Errorf("insert imported master data relation: %w", err)
		}
	}
	return nil
}
