package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

type masterDataIdentity struct {
	typeName string
	key      string
}

type masterDataRelationIdentity struct {
	parent masterDataIdentity
	child  masterDataIdentity
}

func (s *MasterDataService) importReconciledMasterData(
	ctx context.Context,
	doc *MasterDataDocument,
	entriesByType map[string][]MasterDataEntry,
	articleTypesWereImported bool,
	accessorySubtypesWereImported bool,
) (*MasterDataImportResult, error) {
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
	desired, importedEntries, err := buildDesiredMasterData(
		ctx,
		tx,
		current,
		entriesByType,
		now,
		articleTypesWereImported,
		accessorySubtypesWereImported,
	)
	if err != nil {
		return nil, err
	}
	desiredByType := masterDataEntriesByType(desired)
	if err := validateImportedAccessoryCustomFieldReferences(ctx, tx, desiredByType); err != nil {
		return nil, err
	}

	currentRelations, err := loadCurrentMasterDataRelations(ctx, tx)
	if err != nil {
		return nil, err
	}
	desiredRelations, importedRelations, err := buildDesiredMasterDataRelations(
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
	s.invalidateCache()
	return &MasterDataImportResult{
		ImportedTypes:     len(doc.Entries),
		ImportedEntries:   importedEntries,
		ImportedRelations: importedRelations,
	}, nil
}

func loadCurrentMasterData(
	ctx context.Context,
	tx *sql.Tx,
) (map[masterDataIdentity]MasterDataEntry, error) {
	rows, err := tx.QueryContext(ctx, `
SELECT id, type, key, label, active, sort_order, COALESCE(source_url, ''), metadata_json,
       created_at, updated_at, origin
FROM master_data_entries`)
	if err != nil {
		return nil, fmt.Errorf("load current master data: %w", err)
	}
	defer func() { _ = rows.Close() }()
	entries := map[masterDataIdentity]MasterDataEntry{}
	for rows.Next() {
		entry, err := scanMasterDataEntry(rows)
		if err != nil {
			return nil, err
		}
		entries[masterDataIdentity{entry.Type, entry.Key}] = entry
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate current master data: %w", err)
	}
	return entries, nil
}

func buildDesiredMasterData(
	ctx context.Context,
	tx *sql.Tx,
	current map[masterDataIdentity]MasterDataEntry,
	entriesByType map[string][]MasterDataEntry,
	now string,
	articleTypesWereImported bool,
	accessorySubtypesWereImported bool,
) (map[masterDataIdentity]MasterDataEntry, int, error) {
	desired := map[masterDataIdentity]MasterDataEntry{}
	importedCount := 0
	for bucketType, entries := range entriesByType {
		for _, entry := range entries {
			normalized, identity, err := normalizeMasterDataImportEntry(bucketType, entry, current, now)
			if err != nil {
				return nil, 0, err
			}
			if _, duplicate := desired[identity]; duplicate {
				return nil, 0, fmt.Errorf("%w: duplicate master data %s/%s",
					ErrMasterDataValidation, identity.typeName, identity.key)
			}
			desired[identity] = normalized
			preservedArticleType := identity.typeName == standardArticleType && !articleTypesWereImported
			preservedSubtype := identity.typeName == standardAccessorySubtype && !accessorySubtypesWereImported
			if !preservedArticleType && !preservedSubtype {
				importedCount++
			}
		}
	}
	for identity, entry := range current {
		if _, imported := desired[identity]; imported {
			continue
		}
		if entry.Origin == MasterDataOriginBundled {
			desired[identity] = entry
			continue
		}
		used, err := masterDataIsUsed(ctx, tx, entry)
		if err != nil {
			return nil, 0, fmt.Errorf("check omitted master data %s/%s: %w",
				identity.typeName, identity.key, err)
		}
		if used {
			return nil, 0, fmt.Errorf("%w: omitted referenced master data %s/%s",
				ErrMasterDataInUse, identity.typeName, identity.key)
		}
	}
	return desired, importedCount, nil
}

func normalizeMasterDataImportEntry(
	bucketType string,
	entry MasterDataEntry,
	current map[masterDataIdentity]MasterDataEntry,
	now string,
) (MasterDataEntry, masterDataIdentity, error) {
	entry.Type = effectiveMasterDataType(strings.TrimSpace(bucketType), entry)
	entry.Key = strings.TrimSpace(entry.Key)
	entry.Label = strings.TrimSpace(entry.Label)
	if entry.Type == "" || entry.Label == "" {
		return entry, masterDataIdentity{}, ErrMasterDataValidation
	}
	if entry.Key == "" {
		entry.Key = slugKey(entry.Label)
	}
	identity := masterDataIdentity{entry.Type, entry.Key}
	entry.ID = strings.TrimSpace(entry.ID)
	if entry.ID == "" {
		entry.ID = entry.Type + ":" + entry.Key
	}
	entry.SourceURL = strings.TrimSpace(entry.SourceURL)
	if entry.Metadata == nil {
		entry.Metadata = map[string]any{}
	}
	entry.CreatedAt = strings.TrimSpace(entry.CreatedAt)
	if entry.CreatedAt == "" {
		entry.CreatedAt = now
	}
	entry.UpdatedAt = strings.TrimSpace(entry.UpdatedAt)
	if entry.UpdatedAt == "" {
		entry.UpdatedAt = now
	}
	entry.Capabilities = nil
	if currentEntry, exists := current[identity]; exists && currentEntry.Origin == MasterDataOriginBundled {
		entry.Origin = MasterDataOriginBundled
	} else {
		entry.Origin = MasterDataOriginCustom
	}
	return entry, identity, nil
}

func masterDataEntriesByType(
	entries map[masterDataIdentity]MasterDataEntry,
) map[string][]MasterDataEntry {
	byType := map[string][]MasterDataEntry{}
	for _, entry := range entries {
		byType[entry.Type] = append(byType[entry.Type], entry)
	}
	return byType
}

func loadCurrentMasterDataRelations(
	ctx context.Context,
	tx *sql.Tx,
) ([]MasterDataRelation, error) {
	rows, err := tx.QueryContext(ctx, `
SELECT id, parent_type, parent_key, child_type, child_key, sort_order
FROM master_data_relations`)
	if err != nil {
		return nil, fmt.Errorf("load current master data relations: %w", err)
	}
	defer func() { _ = rows.Close() }()
	relations := []MasterDataRelation{}
	for rows.Next() {
		var relation MasterDataRelation
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

func buildDesiredMasterDataRelations(
	imported []MasterDataRelation,
	currentRelations []MasterDataRelation,
	currentEntries map[masterDataIdentity]MasterDataEntry,
	desiredEntries map[masterDataIdentity]MasterDataEntry,
) ([]MasterDataRelation, int, error) {
	relations := map[masterDataRelationIdentity]MasterDataRelation{}
	for _, relation := range imported {
		normalized, identity, err := normalizeMasterDataRelation(relation)
		if err != nil {
			return nil, 0, err
		}
		if _, duplicate := relations[identity]; duplicate {
			return nil, 0, fmt.Errorf("%w: duplicate master data relation", ErrMasterDataValidation)
		}
		relations[identity] = normalized
	}
	for _, relation := range currentRelations {
		identity := masterDataRelationKey(relation)
		if _, imported := relations[identity]; imported {
			continue
		}
		parent, parentExists := currentEntries[identity.parent]
		child, childExists := currentEntries[identity.child]
		_, parentRetained := desiredEntries[identity.parent]
		_, childRetained := desiredEntries[identity.child]
		if parentExists && childExists && parentRetained && childRetained &&
			parent.Origin == MasterDataOriginBundled && child.Origin == MasterDataOriginBundled {
			relations[identity] = relation
		}
	}
	ordered := make([]MasterDataRelation, 0, len(relations))
	for identity, relation := range relations {
		if _, exists := desiredEntries[identity.parent]; !exists {
			return nil, 0, fmt.Errorf("%w: missing master data relation parent %s/%s",
				ErrMasterDataValidation, identity.parent.typeName, identity.parent.key)
		}
		if _, exists := desiredEntries[identity.child]; !exists {
			return nil, 0, fmt.Errorf("%w: missing master data relation child %s/%s",
				ErrMasterDataValidation, identity.child.typeName, identity.child.key)
		}
		ordered = append(ordered, relation)
	}
	sort.Slice(ordered, func(i, j int) bool {
		left := masterDataRelationKey(ordered[i])
		right := masterDataRelationKey(ordered[j])
		if left.parent.typeName != right.parent.typeName {
			return left.parent.typeName < right.parent.typeName
		}
		if left.parent.key != right.parent.key {
			return left.parent.key < right.parent.key
		}
		if left.child.typeName != right.child.typeName {
			return left.child.typeName < right.child.typeName
		}
		return left.child.key < right.child.key
	})
	return ordered, len(imported), nil
}

func normalizeMasterDataRelation(
	relation MasterDataRelation,
) (MasterDataRelation, masterDataRelationIdentity, error) {
	relation.ParentType = strings.TrimSpace(relation.ParentType)
	relation.ParentKey = strings.TrimSpace(relation.ParentKey)
	relation.ChildType = strings.TrimSpace(relation.ChildType)
	relation.ChildKey = strings.TrimSpace(relation.ChildKey)
	if relation.ParentType == "" || relation.ParentKey == "" ||
		relation.ChildType == "" || relation.ChildKey == "" {
		return relation, masterDataRelationIdentity{}, ErrMasterDataValidation
	}
	relation.ID = strings.TrimSpace(relation.ID)
	if relation.ID == "" {
		relation.ID = randomID()
	}
	return relation, masterDataRelationKey(relation), nil
}

func masterDataRelationKey(relation MasterDataRelation) masterDataRelationIdentity {
	return masterDataRelationIdentity{
		parent: masterDataIdentity{relation.ParentType, relation.ParentKey},
		child:  masterDataIdentity{relation.ChildType, relation.ChildKey},
	}
}

func replaceMasterData(
	ctx context.Context,
	tx *sql.Tx,
	entries map[masterDataIdentity]MasterDataEntry,
	relations []MasterDataRelation,
	now string,
) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM master_data_relations`); err != nil {
		return fmt.Errorf("clear master data relations: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM master_data_entries`); err != nil {
		return fmt.Errorf("clear master data entries: %w", err)
	}
	identities := make([]masterDataIdentity, 0, len(entries))
	for identity := range entries {
		identities = append(identities, identity)
	}
	sort.Slice(identities, func(i, j int) bool {
		if identities[i].typeName != identities[j].typeName {
			return identities[i].typeName < identities[j].typeName
		}
		return identities[i].key < identities[j].key
	})
	for _, identity := range identities {
		entry := entries[identity]
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
