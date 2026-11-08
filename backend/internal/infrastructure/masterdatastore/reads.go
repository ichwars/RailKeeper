package masterdatastore

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"railkeeper/backend/internal/inventory/masterdata"
)

type Store struct{ db *sql.DB }

func New(db *sql.DB) *Store { return &Store{db: db} }

var _ masterdata.Reader = (*Store)(nil)

func (s *Store) List(ctx context.Context, typeName string, activeOnly bool) ([]masterdata.MasterDataEntry, error) {
	query := `
SELECT id, type, key, label, active, sort_order, COALESCE(source_url, ''), metadata_json,
       created_at, updated_at, origin
FROM master_data_entries
WHERE type=?`
	args := []any{typeName}
	if activeOnly {
		query += " AND active=1"
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list master data: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := []masterdata.MasterDataEntry{}
	for rows.Next() {
		item, err := ScanEntry(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate master data: %w", err)
	}
	masterdata.SortEntries(out)
	return out, nil
}

func (s *Store) Snapshot(ctx context.Context) (map[string][]masterdata.MasterDataEntry, error) {
	query := `
SELECT id, type, key, label, active, sort_order, COALESCE(source_url, ''), metadata_json,
       created_at, updated_at, origin
FROM master_data_entries`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list all master data: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := map[string][]masterdata.MasterDataEntry{}
	for rows.Next() {
		item, err := ScanEntry(rows)
		if err != nil {
			return nil, err
		}
		out[item.Type] = append(out[item.Type], item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate all master data: %w", err)
	}
	for _, entries := range out {
		masterdata.SortEntries(entries)
	}

	return out, nil
}

func (s *Store) Get(ctx context.Context, typeName, key string) (*masterdata.MasterDataEntry, error) {
	var metadataJSON string
	var active int
	var item masterdata.MasterDataEntry
	err := s.db.QueryRowContext(ctx, `
SELECT id, type, key, label, active, sort_order, COALESCE(source_url, ''), metadata_json,
       created_at, updated_at, origin
FROM master_data_entries
WHERE type=? AND key=?
`, strings.TrimSpace(typeName), strings.TrimSpace(key)).Scan(
		&item.ID,
		&item.Type,
		&item.Key,
		&item.Label,
		&active,
		&item.SortOrder,
		&item.SourceURL,
		&metadataJSON,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.Origin,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, masterdata.ErrMasterDataNotFound
		}
		return nil, fmt.Errorf("get master data: %w", err)
	}
	item.Active = active == 1
	item.Metadata = map[string]any{}
	if err := json.Unmarshal([]byte(metadataJSON), &item.Metadata); err != nil {
		return nil, fmt.Errorf("parse master data metadata: %w", err)
	}
	return &item, nil
}

func (s *Store) Relations(ctx context.Context, parentType, childType string) ([]masterdata.MasterDataRelation, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, parent_type, parent_key, child_type, child_key, sort_order
FROM master_data_relations
WHERE parent_type=? AND child_type=?
ORDER BY sort_order ASC
`, strings.TrimSpace(parentType), strings.TrimSpace(childType))
	if err != nil {
		return nil, fmt.Errorf("list master data relations: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := []masterdata.MasterDataRelation{}
	for rows.Next() {
		var relation masterdata.MasterDataRelation
		if err := rows.Scan(&relation.ID, &relation.ParentType, &relation.ParentKey, &relation.ChildType, &relation.ChildKey, &relation.SortOrder); err != nil {
			return nil, fmt.Errorf("scan master data relation: %w", err)
		}
		out = append(out, relation)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate master data relations: %w", err)
	}
	return out, nil
}

func (s *Store) AllRelations(ctx context.Context) ([]masterdata.MasterDataRelation, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, parent_type, parent_key, child_type, child_key, sort_order
FROM master_data_relations
ORDER BY parent_type ASC, parent_key ASC, child_type ASC, sort_order ASC
`)
	if err != nil {
		return nil, fmt.Errorf("list all master data relations: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := []masterdata.MasterDataRelation{}
	for rows.Next() {
		var relation masterdata.MasterDataRelation
		if err := rows.Scan(&relation.ID, &relation.ParentType, &relation.ParentKey, &relation.ChildType, &relation.ChildKey, &relation.SortOrder); err != nil {
			return nil, fmt.Errorf("scan master data relation: %w", err)
		}
		out = append(out, relation)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate all master data relations: %w", err)
	}
	return out, nil
}

type EntryScanner interface {
	Scan(dest ...any) error
}

func ScanEntry(scanner EntryScanner) (masterdata.MasterDataEntry, error) {
	var item masterdata.MasterDataEntry
	var active int
	var metadataJSON string
	if err := scanner.Scan(
		&item.ID,
		&item.Type,
		&item.Key,
		&item.Label,
		&active,
		&item.SortOrder,
		&item.SourceURL,
		&metadataJSON,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.Origin,
	); err != nil {
		return item, fmt.Errorf("scan master data: %w", err)
	}
	item.Active = active == 1
	item.Metadata = map[string]any{}
	if err := json.Unmarshal([]byte(metadataJSON), &item.Metadata); err != nil {
		return item, fmt.Errorf("parse master data metadata: %w", err)
	}
	return item, nil
}
