package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"railkeeper/backend/internal/domain"
	"railkeeper/backend/internal/infrastructure/masterdatastore"
	"railkeeper/backend/internal/inventory/masterdata"
)

type MasterDataService struct {
	db             *sql.DB
	reader         masterdata.Reader
	cacheMu        sync.RWMutex
	cache          map[bool]map[string][]MasterDataEntry
	snapshotLoader func(context.Context) (map[string][]MasterDataEntry, error)
}

func NewMasterDataService(db *sql.DB) *MasterDataService {
	return newMasterDataService(db, nil)
}

func newMasterDataService(
	db *sql.DB,
	loader func(context.Context) (map[string][]MasterDataEntry, error),
) *MasterDataService {
	reader := masterdatastore.New(db)
	if loader == nil {
		loader = reader.Snapshot
	}
	return &MasterDataService{db: db, reader: reader, cache: map[bool]map[string][]MasterDataEntry{}, snapshotLoader: loader}
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
	if err := validateDecoderProfile(typeName, input.Key, input.Metadata); err != nil {
		return nil, err
	}
	if typeName == accessoryCustomField {
		metadata, err := normalizeAccessoryCustomFieldMetadata(input.Key, true, input.Metadata)
		if err != nil {
			return nil, err
		}
		input.Metadata = metadata
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
INSERT INTO master_data_entries(
  id, type, key, label, active, sort_order, source_url, metadata_json,
  created_at, updated_at, origin
)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'custom')
`, id, typeName, input.Key, input.Label, boolToInt(active), sortOrder,
		input.SourceURL, string(metadata), now, now); err != nil {
		return nil, fmt.Errorf("create master data: %w", err)
	}
	s.invalidateCache()
	return s.Get(ctx, typeName, input.Key)
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
	if typeName == accessoryCustomField {
		return s.updateAccessoryCustomField(ctx, key, input)
	}
	if err := validateDecoderProfile(typeName, key, input.Metadata); err != nil {
		return nil, err
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
		Version:   masterDataExportVersion,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		Entries:   entries,
		Relations: relations,
	}, nil
}

func (s *MasterDataService) Import(ctx context.Context, doc *MasterDataDocument) (*MasterDataImportResult, error) {
	if doc == nil || doc.Format != masterDataExportFormat || doc.Version < 1 ||
		doc.Version > masterDataExportVersion {
		return nil, ErrMasterDataValidation
	}
	for typeName, entries := range doc.Entries {
		for _, entry := range entries {
			if err := validateDecoderProfile(effectiveMasterDataType(typeName, entry), entry.Key, entry.Metadata); err != nil {
				return nil, err
			}
		}
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
	if err := normalizeImportedAccessoryCustomFields(entriesByType); err != nil {
		return nil, err
	}

	return s.importReconciledMasterData(
		ctx,
		doc,
		entriesByType,
		articleTypesWereImported,
		accessorySubtypesWereImported,
	)
}

func (s *MasterDataService) invalidateCache() {
	s.cacheMu.Lock()
	s.cache = map[bool]map[string][]MasterDataEntry{}
	s.cacheMu.Unlock()
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

type MasterDataOrigin = masterdata.MasterDataOrigin

const MasterDataOriginBundled = masterdata.MasterDataOriginBundled

const MasterDataOriginCustom = masterdata.MasterDataOriginCustom

type MasterDataCapabilities = masterdata.MasterDataCapabilities

const masterDataExportFormat = masterdata.ExportFormat

const masterDataExportVersion = masterdata.ExportVersion

const standardArticleType = masterdata.StandardArticleType

const standardAccessorySubtype = masterdata.StandardAccessorySubtype

const legacyArticleSubtype = masterdata.LegacyArticleSubtype

const accessoryCustomField = masterdata.AccessoryCustomField

type MasterDataEntry = masterdata.MasterDataEntry

type MasterDataInput = masterdata.MasterDataInput

type MasterDataRelation = masterdata.MasterDataRelation

type MasterDataDocument = masterdata.MasterDataDocument

type MasterDataImportResult = masterdata.MasterDataImportResult

var ErrMasterDataValidation = masterdata.ErrMasterDataValidation

var ErrMasterDataNotFound = masterdata.ErrMasterDataNotFound

var ErrMasterDataProtected = masterdata.ErrMasterDataProtected

var standardArticleTypeKeys = masterdata.StandardArticleTypeKeys()

func (s *MasterDataService) List(ctx context.Context, typeName string, activeOnly bool) ([]MasterDataEntry, error) {
	typeName = strings.TrimSpace(typeName)
	if typeName == "" {
		return nil, ErrMasterDataValidation
	}
	return s.reader.List(ctx, typeName, activeOnly)
}

func (s *MasterDataService) Get(ctx context.Context, typeName, key string) (*MasterDataEntry, error) {
	return s.reader.Get(ctx, typeName, key)
}

func (s *MasterDataService) Relations(ctx context.Context, parentType, childType string) ([]MasterDataRelation, error) {
	return s.reader.Relations(ctx, parentType, childType)
}

func (s *MasterDataService) listAllRelations(ctx context.Context) ([]MasterDataRelation, error) {
	return s.reader.AllRelations(ctx)
}

func cleanMasterDataInput(input MasterDataInput) MasterDataInput {
	return masterdata.CleanMasterDataInput(input)
}

func slugKey(value string) string { return masterdata.SlugKey(value) }

func ParseAccessoryCustomAttributeDefinition(
	key string,
	active bool,
	metadata map[string]any,
) (domain.AccessoryAttributeDefinition, error) {
	return masterdata.ParseAccessoryCustomAttributeDefinition(key, active, metadata)
}

func normalizeAccessoryCustomFieldMetadata(
	key string,
	active bool,
	metadata map[string]any,
) (map[string]any, error) {
	return masterdata.NormalizeAccessoryCustomFieldMetadata(key, active, metadata)
}

func normalizeImportedAccessoryCustomFields(entriesByType map[string][]MasterDataEntry) error {
	return masterdata.NormalizeImportedAccessoryCustomFields(entriesByType)
}

func validateImportedArticleTypes(entriesByType map[string][]MasterDataEntry) ([]MasterDataEntry, error) {
	return masterdata.ValidateImportedArticleTypes(entriesByType)
}

func validateProtectedArticleTypes(entries []MasterDataEntry, required bool) error {
	return masterdata.ValidateProtectedArticleTypes(entries, required)
}

func effectiveMasterDataType(bucketType string, entry MasterDataEntry) string {
	return masterdata.EffectiveMasterDataType(bucketType, entry)
}

func containsEffectiveMasterDataType(entriesByType map[string][]MasterDataEntry, typeName string) bool {
	return masterdata.ContainsEffectiveMasterDataType(entriesByType, typeName)
}

func prepareImportedArticleTypes(
	current []MasterDataEntry,
	imported []MasterDataEntry,
) ([]MasterDataEntry, error) {
	return masterdata.PrepareImportedArticleTypes(current, imported)
}

func cloneMasterDataMap(input map[string][]MasterDataEntry) map[string][]MasterDataEntry {
	return masterdata.CloneMasterDataMap(input)
}

func masterDataCaches(snapshot map[string][]MasterDataEntry) map[bool]map[string][]MasterDataEntry {
	return masterdata.MasterDataCaches(snapshot)
}
