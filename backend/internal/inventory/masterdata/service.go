package masterdata

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

type Service struct {
	repository     Repository
	cacheMu        sync.RWMutex
	cache          map[bool]map[string][]MasterDataEntry
	snapshotLoader func(context.Context) (map[string][]MasterDataEntry, error)
}

func New(repository Repository, loader func(context.Context) (map[string][]MasterDataEntry, error)) *Service {
	if loader == nil {
		loader = repository.Snapshot
	}
	return &Service{repository: repository, cache: map[bool]map[string][]MasterDataEntry{}, snapshotLoader: loader}
}
func (s *Service) List(ctx context.Context, typeName string, activeOnly bool) ([]MasterDataEntry, error) {
	typeName = strings.TrimSpace(typeName)
	if typeName == "" {
		return nil, ErrMasterDataValidation
	}
	return s.repository.List(ctx, typeName, activeOnly)
}
func (s *Service) Get(ctx context.Context, typeName, key string) (*MasterDataEntry, error) {
	return s.repository.Get(ctx, typeName, key)
}
func (s *Service) Relations(ctx context.Context, parentType, childType string) ([]MasterDataRelation, error) {
	return s.repository.Relations(ctx, parentType, childType)
}
func (s *Service) WarmCache(ctx context.Context) error {
	return s.RefreshCache(ctx)
}
func (s *Service) RefreshCache(ctx context.Context) error {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	snapshot, err := s.snapshotLoader(ctx)
	if err != nil {
		return err
	}
	s.cache = MasterDataCaches(snapshot)
	return nil
}
func (s *Service) ListAll(ctx context.Context, activeOnly bool) (map[string][]MasterDataEntry, error) {
	s.cacheMu.RLock()
	if cached, ok := s.cache[activeOnly]; ok {
		out := CloneMasterDataMap(cached)
		s.cacheMu.RUnlock()
		return out, nil
	}
	s.cacheMu.RUnlock()

	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	if cached, ok := s.cache[activeOnly]; ok {
		return CloneMasterDataMap(cached), nil
	}
	snapshot, err := s.snapshotLoader(ctx)
	if err != nil {
		return nil, err
	}
	s.cache = MasterDataCaches(snapshot)
	return CloneMasterDataMap(s.cache[activeOnly]), nil
}
func (s *Service) InvalidateCache() {
	s.cacheMu.Lock()
	s.cache = map[bool]map[string][]MasterDataEntry{}
	s.cacheMu.Unlock()
}
func (s *Service) Export(ctx context.Context) (*MasterDataDocument, error) {
	entries, err := s.ListAll(ctx, false)
	if err != nil {
		return nil, err
	}
	relations, err := s.repository.AllRelations(ctx)
	if err != nil {
		return nil, err
	}
	return &MasterDataDocument{
		Format:    ExportFormat,
		Version:   ExportVersion,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		Entries:   entries,
		Relations: relations,
	}, nil
}
func (s *Service) Create(ctx context.Context, typeName string, input MasterDataInput) (*MasterDataEntry, error) {
	typeName = strings.TrimSpace(typeName)
	input = CleanMasterDataInput(input)
	if typeName == StandardArticleType {
		return nil, ErrMasterDataProtected
	}
	if typeName == "" || input.Label == "" {
		return nil, ErrMasterDataValidation
	}
	if input.Key == "" {
		input.Key = SlugKey(input.Label)
	}
	if err := ValidateDecoderProfile(typeName, input.Key, input.Metadata); err != nil {
		return nil, err
	}
	if typeName == AccessoryCustomField {
		metadata, err := NormalizeAccessoryCustomFieldMetadata(input.Key, true, input.Metadata)
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

	mutation := Mutation{Entry: MasterDataEntry{ID: typeName + ":" + input.Key, Type: typeName, Key: input.Key, Label: input.Label,
		Active: active, SortOrder: sortOrder, SourceURL: input.SourceURL, Metadata: input.Metadata}, MetadataJSON: metadata}
	if err := s.repository.Create(ctx, mutation); err != nil {
		return nil, err
	}
	s.InvalidateCache()
	return s.Get(ctx, typeName, input.Key)
}
func (s *Service) Update(ctx context.Context, typeName, key string, input MasterDataInput) (*MasterDataEntry, error) {
	typeName = strings.TrimSpace(typeName)
	key = strings.TrimSpace(key)
	input = CleanMasterDataInput(input)
	if typeName == "" || key == "" || input.Label == "" {
		return nil, ErrMasterDataValidation
	}
	if typeName == StandardArticleType && input.Key != "" && input.Key != key {
		return nil, ErrMasterDataProtected
	}
	if typeName == AccessoryCustomField {
		return s.updateAccessoryCustomField(ctx, key, input)
	}
	if err := ValidateDecoderProfile(typeName, key, input.Metadata); err != nil {
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
	mutation := Mutation{Entry: MasterDataEntry{ID: typeName + ":" + key, Type: typeName, Key: key, Label: input.Label,
		Active: active, SortOrder: sortOrder, SourceURL: input.SourceURL, Metadata: input.Metadata}, MetadataJSON: metadata}
	if err := s.repository.Update(ctx, mutation); err != nil {
		return nil, err
	}
	s.InvalidateCache()
	return s.Get(ctx, typeName, key)
}
func (s *Service) updateAccessoryCustomField(
	ctx context.Context,
	key string,
	input MasterDataInput,
) (*MasterDataEntry, error) {
	active := true
	if input.Active != nil {
		active = *input.Active
	}
	sortOrder := 0
	if input.SortOrder != nil {
		sortOrder = *input.SortOrder
	}
	metadata, err := NormalizeAccessoryCustomFieldMetadata(key, active, input.Metadata)
	if err != nil {
		return nil, err
	}
	definition, err := ParseAccessoryCustomAttributeDefinition(key, active, metadata)
	if err != nil {
		return nil, err
	}
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("marshal master data metadata: %w", err)
	}

	mutation := Mutation{Entry: MasterDataEntry{ID: AccessoryCustomField + ":" + key, Type: AccessoryCustomField, Key: key, Label: input.Label,
		Active: active, SortOrder: sortOrder, SourceURL: input.SourceURL, Metadata: metadata}, MetadataJSON: metadataJSON, Definition: &definition}
	if err := s.repository.UpdateCustomField(ctx, mutation); err != nil {
		return nil, err
	}
	s.InvalidateCache()
	return s.Get(ctx, AccessoryCustomField, key)
}
func (s *Service) Import(ctx context.Context, doc *MasterDataDocument) (*MasterDataImportResult, error) {
	if doc == nil || doc.Format != ExportFormat || doc.Version < 1 ||
		doc.Version > ExportVersion {
		return nil, ErrMasterDataValidation
	}
	for typeName, entries := range doc.Entries {
		for _, entry := range entries {
			if err := ValidateDecoderProfile(EffectiveMasterDataType(typeName, entry), entry.Key, entry.Metadata); err != nil {
				return nil, err
			}
		}
	}
	importedArticleTypes, err := ValidateImportedArticleTypes(doc.Entries)
	if err != nil {
		return nil, err
	}
	currentArticleTypes, err := s.List(ctx, StandardArticleType, false)
	if err != nil {
		return nil, fmt.Errorf("read authoritative article types: %w", err)
	}
	articleTypes, err := PrepareImportedArticleTypes(currentArticleTypes, importedArticleTypes)
	if err != nil {
		return nil, err
	}
	currentAccessorySubtypes, err := s.List(ctx, StandardAccessorySubtype, false)
	if err != nil {
		return nil, fmt.Errorf("read authoritative accessory subtypes: %w", err)
	}
	entriesByType := make(map[string][]MasterDataEntry, len(doc.Entries)+1)
	for typeName, entries := range doc.Entries {
		for _, entry := range entries {
			if EffectiveMasterDataType(typeName, entry) != StandardArticleType {
				entriesByType[typeName] = append(entriesByType[typeName], entry)
			}
		}
	}
	entriesByType[StandardArticleType] = append(entriesByType[StandardArticleType], articleTypes...)
	accessorySubtypesWereImported := ContainsEffectiveMasterDataType(doc.Entries, StandardAccessorySubtype)
	if !accessorySubtypesWereImported {
		entriesByType[StandardAccessorySubtype] = append(
			entriesByType[StandardAccessorySubtype], currentAccessorySubtypes...,
		)
	}
	articleTypesWereImported := len(importedArticleTypes) > 0
	if err := NormalizeImportedAccessoryCustomFields(entriesByType); err != nil {
		return nil, err
	}

	result, err := s.repository.ImportReconciled(ctx, ImportRequest{Document: doc, EntriesByType: entriesByType,
		ArticleTypesWereImported: articleTypesWereImported, AccessorySubtypesWereImported: accessorySubtypesWereImported})
	if err != nil {
		return nil, err
	}
	s.InvalidateCache()
	return result, nil
}
