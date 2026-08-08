package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

var (
	ErrMasterDataValidation = errors.New("master data validation failed")
	ErrMasterDataNotFound   = errors.New("master data not found")
	ErrMasterDataProtected  = fmt.Errorf("%w: standard article type keys are protected", ErrMasterDataValidation)
)

const masterDataExportFormat = "railkeeper-master-data"
const standardArticleType = "article_type"
const standardAccessorySubtype = "accessory_subtype"
const legacyArticleSubtype = "article_subtype"

var standardArticleTypeKeys = []string{
	"track", "signal", "decoder", "electrical_control", "building_equipment",
	"landscape_consumable", "lighting", "other",
}

type MasterDataService struct {
	db             *sql.DB
	cacheMu        sync.RWMutex
	cache          map[bool]map[string][]MasterDataEntry
	snapshotLoader func(context.Context) (map[string][]MasterDataEntry, error)
}

type MasterDataEntry struct {
	ID        string         `json:"id"`
	Type      string         `json:"type"`
	Key       string         `json:"key"`
	Label     string         `json:"label"`
	Active    bool           `json:"active"`
	SortOrder int            `json:"sortOrder"`
	SourceURL string         `json:"sourceUrl,omitempty"`
	Metadata  map[string]any `json:"metadata"`
	CreatedAt string         `json:"createdAt"`
	UpdatedAt string         `json:"updatedAt"`
}

type MasterDataInput struct {
	Key       string         `json:"key"`
	Label     string         `json:"label"`
	Active    *bool          `json:"active"`
	SortOrder *int           `json:"sortOrder"`
	SourceURL string         `json:"sourceUrl"`
	Metadata  map[string]any `json:"metadata"`
}

type MasterDataRelation struct {
	ID         string `json:"id"`
	ParentType string `json:"parentType"`
	ParentKey  string `json:"parentKey"`
	ChildType  string `json:"childType"`
	ChildKey   string `json:"childKey"`
	SortOrder  int    `json:"sortOrder"`
}

type MasterDataDocument struct {
	Format    string                       `json:"format"`
	Version   int                          `json:"version"`
	CreatedAt string                       `json:"createdAt"`
	Entries   map[string][]MasterDataEntry `json:"entries"`
	Relations []MasterDataRelation         `json:"relations"`
}

type MasterDataImportResult struct {
	ImportedTypes     int `json:"importedTypes"`
	ImportedEntries   int `json:"importedEntries"`
	ImportedRelations int `json:"importedRelations"`
}

func NewMasterDataService(db *sql.DB) *MasterDataService {
	return newMasterDataService(db, nil)
}

func newMasterDataService(
	db *sql.DB,
	loader func(context.Context) (map[string][]MasterDataEntry, error),
) *MasterDataService {
	if loader == nil {
		loader = func(ctx context.Context) (map[string][]MasterDataEntry, error) {
			return loadMasterDataSnapshot(ctx, db)
		}
	}
	return &MasterDataService{db: db, cache: map[bool]map[string][]MasterDataEntry{}, snapshotLoader: loader}
}

func (s *MasterDataService) WarmCache(ctx context.Context) error {
	return s.RefreshCache(ctx)
}

func (s *MasterDataService) RefreshCache(ctx context.Context) error {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	snapshot, err := s.snapshotLoader(ctx)
	if err != nil {
		return err
	}
	s.cache = masterDataCaches(snapshot)
	return nil
}

func (s *MasterDataService) List(ctx context.Context, typeName string, activeOnly bool) ([]MasterDataEntry, error) {
	typeName = strings.TrimSpace(typeName)
	if typeName == "" {
		return nil, ErrMasterDataValidation
	}

	query := `
SELECT id, type, key, label, active, sort_order, COALESCE(source_url, ''), metadata_json, created_at, updated_at
FROM master_data_entries
WHERE type=?`
	args := []any{typeName}
	if activeOnly {
		query += " AND active=1"
	}
	query += " ORDER BY active DESC, sort_order ASC, label ASC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list master data: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := []MasterDataEntry{}
	for rows.Next() {
		item, err := scanMasterDataEntry(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate master data: %w", err)
	}
	return out, nil
}

func (s *MasterDataService) ListAll(ctx context.Context, activeOnly bool) (map[string][]MasterDataEntry, error) {
	s.cacheMu.RLock()
	if cached, ok := s.cache[activeOnly]; ok {
		out := cloneMasterDataMap(cached)
		s.cacheMu.RUnlock()
		return out, nil
	}
	s.cacheMu.RUnlock()

	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	if cached, ok := s.cache[activeOnly]; ok {
		return cloneMasterDataMap(cached), nil
	}
	snapshot, err := s.snapshotLoader(ctx)
	if err != nil {
		return nil, err
	}
	s.cache = masterDataCaches(snapshot)
	return cloneMasterDataMap(s.cache[activeOnly]), nil
}

func loadMasterDataSnapshot(ctx context.Context, db *sql.DB) (map[string][]MasterDataEntry, error) {
	query := `
SELECT id, type, key, label, active, sort_order, COALESCE(source_url, ''), metadata_json, created_at, updated_at
FROM master_data_entries`
	query += " ORDER BY type ASC, active DESC, sort_order ASC, label ASC"

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list all master data: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := map[string][]MasterDataEntry{}
	for rows.Next() {
		item, err := scanMasterDataEntry(rows)
		if err != nil {
			return nil, err
		}
		out[item.Type] = append(out[item.Type], item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate all master data: %w", err)
	}

	return out, nil
}

func (s *MasterDataService) Create(ctx context.Context, typeName string, input MasterDataInput) (*MasterDataEntry, error) {
	typeName = strings.TrimSpace(typeName)
	input = cleanMasterDataInput(input)
	if typeName == standardArticleType {
		return nil, ErrMasterDataProtected
	}
	if typeName == "" || input.Label == "" {
		return nil, ErrMasterDataValidation
	}
	if input.Key == "" {
		input.Key = slugKey(input.Label)
	}
	active := true
	if input.Active != nil {
		active = *input.Active
	}
	sortOrder := 0
	if input.SortOrder != nil {
		sortOrder = *input.SortOrder
	}
	metadata, err := json.Marshal(input.Metadata)
	if err != nil {
		return nil, fmt.Errorf("marshal master data metadata: %w", err)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	id := typeName + ":" + input.Key
	if _, err := s.db.ExecContext(ctx, `
INSERT INTO master_data_entries(id, type, key, label, active, sort_order, source_url, metadata_json, created_at, updated_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, id, typeName, input.Key, input.Label, boolToInt(active), sortOrder, input.SourceURL, string(metadata), now, now); err != nil {
		return nil, fmt.Errorf("create master data: %w", err)
	}
	s.invalidateCache()
	return s.Get(ctx, typeName, input.Key)
}

func (s *MasterDataService) Get(ctx context.Context, typeName, key string) (*MasterDataEntry, error) {
	var metadataJSON string
	var active int
	var item MasterDataEntry
	err := s.db.QueryRowContext(ctx, `
SELECT id, type, key, label, active, sort_order, COALESCE(source_url, ''), metadata_json, created_at, updated_at
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
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrMasterDataNotFound
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

func (s *MasterDataService) Update(ctx context.Context, typeName, key string, input MasterDataInput) (*MasterDataEntry, error) {
	typeName = strings.TrimSpace(typeName)
	key = strings.TrimSpace(key)
	input = cleanMasterDataInput(input)
	if typeName == "" || key == "" || input.Label == "" {
		return nil, ErrMasterDataValidation
	}
	if typeName == standardArticleType && input.Key != "" && input.Key != key {
		return nil, ErrMasterDataProtected
	}
	active := true
	if input.Active != nil {
		active = *input.Active
	}
	sortOrder := 0
	if input.SortOrder != nil {
		sortOrder = *input.SortOrder
	}
	metadata, err := json.Marshal(input.Metadata)
	if err != nil {
		return nil, fmt.Errorf("marshal master data metadata: %w", err)
	}
	result, err := s.db.ExecContext(ctx, `
UPDATE master_data_entries
SET label=?, active=?, sort_order=?, source_url=?, metadata_json=?, updated_at=?
WHERE type=? AND key=?
`, input.Label, boolToInt(active), sortOrder, input.SourceURL, string(metadata), time.Now().UTC().Format(time.RFC3339), typeName, key)
	if err != nil {
		return nil, fmt.Errorf("update master data: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("read master data update result: %w", err)
	}
	if affected == 0 {
		return nil, ErrMasterDataNotFound
	}
	s.invalidateCache()
	return s.Get(ctx, typeName, key)
}

func (s *MasterDataService) Delete(ctx context.Context, typeName, key string) error {
	typeName = strings.TrimSpace(typeName)
	key = strings.TrimSpace(key)
	if typeName == standardArticleType {
		return ErrMasterDataProtected
	}
	result, err := s.db.ExecContext(ctx, `DELETE FROM master_data_entries WHERE type=? AND key=?`, typeName, key)
	if err != nil {
		return fmt.Errorf("delete master data: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read master data delete result: %w", err)
	}
	if affected == 0 {
		return ErrMasterDataNotFound
	}
	s.invalidateCache()
	return nil
}

func (s *MasterDataService) Relations(ctx context.Context, parentType, childType string) ([]MasterDataRelation, error) {
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

	out := []MasterDataRelation{}
	for rows.Next() {
		var relation MasterDataRelation
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

func (s *MasterDataService) Export(ctx context.Context) (*MasterDataDocument, error) {
	entries, err := s.ListAll(ctx, false)
	if err != nil {
		return nil, err
	}
	relations, err := s.listAllRelations(ctx)
	if err != nil {
		return nil, err
	}
	return &MasterDataDocument{
		Format:    masterDataExportFormat,
		Version:   1,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		Entries:   entries,
		Relations: relations,
	}, nil
}

func (s *MasterDataService) Import(ctx context.Context, doc *MasterDataDocument) (*MasterDataImportResult, error) {
	if doc == nil || doc.Format != masterDataExportFormat || doc.Version < 1 {
		return nil, ErrMasterDataValidation
	}
	importedArticleTypes, err := validateImportedArticleTypes(doc.Entries)
	if err != nil {
		return nil, err
	}
	currentArticleTypes, err := s.List(ctx, standardArticleType, false)
	if err != nil {
		return nil, fmt.Errorf("read authoritative article types: %w", err)
	}
	articleTypes, err := prepareImportedArticleTypes(currentArticleTypes, importedArticleTypes)
	if err != nil {
		return nil, err
	}
	currentAccessorySubtypes, err := s.List(ctx, standardAccessorySubtype, false)
	if err != nil {
		return nil, fmt.Errorf("read authoritative accessory subtypes: %w", err)
	}
	entriesByType := make(map[string][]MasterDataEntry, len(doc.Entries)+1)
	for typeName, entries := range doc.Entries {
		for _, entry := range entries {
			if effectiveMasterDataType(typeName, entry) != standardArticleType {
				entriesByType[typeName] = append(entriesByType[typeName], entry)
			}
		}
	}
	entriesByType[standardArticleType] = append(entriesByType[standardArticleType], articleTypes...)
	accessorySubtypesWereImported := containsEffectiveMasterDataType(doc.Entries, standardAccessorySubtype)
	if !accessorySubtypesWereImported {
		entriesByType[standardAccessorySubtype] = append(
			entriesByType[standardAccessorySubtype], currentAccessorySubtypes...,
		)
	}
	articleTypesWereImported := len(importedArticleTypes) > 0

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin master data import: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `DELETE FROM master_data_relations`); err != nil {
		return nil, fmt.Errorf("clear master data relations: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM master_data_entries`); err != nil {
		return nil, fmt.Errorf("clear master data entries: %w", err)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	result := &MasterDataImportResult{ImportedTypes: len(doc.Entries)}
	for typeName, entries := range entriesByType {
		typeName = strings.TrimSpace(typeName)
		for _, entry := range entries {
			entryType := effectiveMasterDataType(typeName, entry)
			key := strings.TrimSpace(entry.Key)
			label := strings.TrimSpace(entry.Label)
			if entryType == "" || label == "" {
				return nil, ErrMasterDataValidation
			}
			if key == "" {
				key = slugKey(label)
			}
			id := strings.TrimSpace(entry.ID)
			if id == "" {
				id = entryType + ":" + key
			}
			metadata := entry.Metadata
			if metadata == nil {
				metadata = map[string]any{}
			}
			metadataJSON, err := json.Marshal(metadata)
			if err != nil {
				return nil, fmt.Errorf("marshal imported master data metadata: %w", err)
			}
			createdAt := strings.TrimSpace(entry.CreatedAt)
			if createdAt == "" {
				createdAt = now
			}
			updatedAt := strings.TrimSpace(entry.UpdatedAt)
			if updatedAt == "" {
				updatedAt = now
			}
			if _, err := tx.ExecContext(ctx, `
INSERT INTO master_data_entries(id, type, key, label, active, sort_order, source_url, metadata_json, created_at, updated_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, id, entryType, key, label, boolToInt(entry.Active), entry.SortOrder, strings.TrimSpace(entry.SourceURL), string(metadataJSON), createdAt, updatedAt); err != nil {
				return nil, fmt.Errorf("insert imported master data entry: %w", err)
			}
			preservedArticleType := entryType == standardArticleType && !articleTypesWereImported
			preservedAccessorySubtype := entryType == standardAccessorySubtype && !accessorySubtypesWereImported
			if !preservedArticleType && !preservedAccessorySubtype {
				result.ImportedEntries++
			}
		}
	}

	for _, relation := range doc.Relations {
		relation.ParentType = strings.TrimSpace(relation.ParentType)
		relation.ParentKey = strings.TrimSpace(relation.ParentKey)
		relation.ChildType = strings.TrimSpace(relation.ChildType)
		relation.ChildKey = strings.TrimSpace(relation.ChildKey)
		if relation.ParentType == "" || relation.ParentKey == "" || relation.ChildType == "" || relation.ChildKey == "" {
			return nil, ErrMasterDataValidation
		}
		id := strings.TrimSpace(relation.ID)
		if id == "" {
			id = randomID()
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO master_data_relations(id, parent_type, parent_key, child_type, child_key, sort_order, created_at)
VALUES(?, ?, ?, ?, ?, ?, ?)
`, id, relation.ParentType, relation.ParentKey, relation.ChildType, relation.ChildKey, relation.SortOrder, now); err != nil {
			return nil, fmt.Errorf("insert imported master data relation: %w", err)
		}
		result.ImportedRelations++
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit master data import: %w", err)
	}
	s.invalidateCache()
	return result, nil
}

func validateImportedArticleTypes(entriesByType map[string][]MasterDataEntry) ([]MasterDataEntry, error) {
	entries := []MasterDataEntry{}
	for typeName, typedEntries := range entriesByType {
		for _, entry := range typedEntries {
			if effectiveMasterDataType(typeName, entry) == standardArticleType {
				entries = append(entries, entry)
			}
		}
	}
	if err := validateProtectedArticleTypes(entries, false); err != nil {
		return nil, err
	}
	return entries, nil
}

func validateProtectedArticleTypes(entries []MasterDataEntry, required bool) error {
	if len(entries) == 0 && !required {
		return nil
	}
	if len(entries) != len(standardArticleTypeKeys) {
		return ErrMasterDataProtected
	}
	expected := make(map[string]bool, len(standardArticleTypeKeys))
	for _, key := range standardArticleTypeKeys {
		expected[key] = true
	}
	seen := make(map[string]bool, len(entries))
	for _, entry := range entries {
		key := strings.TrimSpace(entry.Key)
		if !expected[key] || seen[key] {
			return ErrMasterDataProtected
		}
		if strings.TrimSpace(entry.Label) == "" {
			return ErrMasterDataValidation
		}
		seen[key] = true
	}
	return nil
}

func effectiveMasterDataType(bucketType string, entry MasterDataEntry) string {
	entryType := strings.TrimSpace(bucketType)
	if entry.Type != "" {
		entryType = strings.TrimSpace(entry.Type)
	}
	if entryType == legacyArticleSubtype {
		return standardAccessorySubtype
	}
	return entryType
}

func containsEffectiveMasterDataType(entriesByType map[string][]MasterDataEntry, typeName string) bool {
	for bucketType, entries := range entriesByType {
		for _, entry := range entries {
			if effectiveMasterDataType(bucketType, entry) == typeName {
				return true
			}
		}
	}
	return false
}

func prepareImportedArticleTypes(
	current []MasterDataEntry,
	imported []MasterDataEntry,
) ([]MasterDataEntry, error) {
	if len(current) != len(standardArticleTypeKeys) {
		return nil, ErrMasterDataProtected
	}
	byKey := make(map[string]MasterDataEntry, len(current))
	for _, entry := range current {
		byKey[entry.Key] = entry
	}
	for _, key := range standardArticleTypeKeys {
		if _, ok := byKey[key]; !ok {
			return nil, ErrMasterDataProtected
		}
	}
	if len(imported) == 0 {
		return current, nil
	}
	for _, entry := range imported {
		currentEntry := byKey[strings.TrimSpace(entry.Key)]
		currentEntry.Label = strings.TrimSpace(entry.Label)
		currentEntry.Active = entry.Active
		currentEntry.SortOrder = entry.SortOrder
		currentEntry.Metadata = entry.Metadata
		currentEntry.UpdatedAt = ""
		byKey[currentEntry.Key] = currentEntry
	}
	prepared := make([]MasterDataEntry, 0, len(standardArticleTypeKeys))
	for _, key := range standardArticleTypeKeys {
		prepared = append(prepared, byKey[key])
	}
	return prepared, nil
}

type masterDataScanner interface {
	Scan(dest ...any) error
}

func (s *MasterDataService) listAllRelations(ctx context.Context) ([]MasterDataRelation, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, parent_type, parent_key, child_type, child_key, sort_order
FROM master_data_relations
ORDER BY parent_type ASC, parent_key ASC, child_type ASC, sort_order ASC
`)
	if err != nil {
		return nil, fmt.Errorf("list all master data relations: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := []MasterDataRelation{}
	for rows.Next() {
		var relation MasterDataRelation
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

func (s *MasterDataService) invalidateCache() {
	s.cacheMu.Lock()
	s.cache = map[bool]map[string][]MasterDataEntry{}
	s.cacheMu.Unlock()
}

func cloneMasterDataMap(input map[string][]MasterDataEntry) map[string][]MasterDataEntry {
	out := make(map[string][]MasterDataEntry, len(input))
	for key, entries := range input {
		out[key] = append([]MasterDataEntry(nil), entries...)
	}
	return out
}

func masterDataCaches(snapshot map[string][]MasterDataEntry) map[bool]map[string][]MasterDataEntry {
	active := make(map[string][]MasterDataEntry, len(snapshot))
	for typeName, entries := range snapshot {
		for _, entry := range entries {
			if entry.Active {
				active[typeName] = append(active[typeName], entry)
			}
		}
	}
	return map[bool]map[string][]MasterDataEntry{
		false: cloneMasterDataMap(snapshot),
		true:  active,
	}
}

func scanMasterDataEntry(scanner masterDataScanner) (MasterDataEntry, error) {
	var item MasterDataEntry
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

func cleanMasterDataInput(input MasterDataInput) MasterDataInput {
	input.Key = strings.TrimSpace(input.Key)
	input.Label = strings.TrimSpace(input.Label)
	input.SourceURL = strings.TrimSpace(input.SourceURL)
	if input.Metadata == nil {
		input.Metadata = map[string]any{}
	}
	return input
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

var slugPattern = regexp.MustCompile(`[^a-z0-9]+`)

func slugKey(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.NewReplacer("\u00e4", "ae", "\u00f6", "oe", "\u00fc", "ue", "\u00df", "ss").Replace(value)
	value = slugPattern.ReplaceAllString(value, "-")
	return strings.Trim(value, "-")
}
