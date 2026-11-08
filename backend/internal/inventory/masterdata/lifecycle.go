package masterdata

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrMasterDataBundled      = errors.New("bundled master data cannot be deleted")
	ErrMasterDataInUse        = fmt.Errorf("%w: master data entry is in use", ErrMasterDataValidation)
	ErrMasterDataUsageUnknown = errors.New("master data usage is unknown")
)

const MaxMasterDataActiveBatchSize = 5000

func normalizeMasterDataBatchKeys(keys []string) ([]string, error) {
	if len(keys) == 0 {
		return nil, ErrMasterDataValidation
	}
	normalized := make([]string, 0, len(keys))
	seen := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key == "" {
			return nil, ErrMasterDataValidation
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		normalized = append(normalized, key)
		if len(normalized) > MaxMasterDataActiveBatchSize {
			return nil, ErrMasterDataValidation
		}
	}
	return normalized, nil
}
func isStandardArticleTypeKey(key string) bool {
	for _, standardKey := range standardArticleTypeKeys {
		if key == standardKey {
			return true
		}
	}
	return false
}
func (s *Service) ListForManagement(
	ctx context.Context,
	typeName string,
) ([]MasterDataEntry, error) {
	entries, err := s.List(ctx, typeName, false)
	if err != nil {
		return nil, err
	}
	for index := range entries {
		capabilities, err := masterDataCapabilities(ctx, s.repository.Usage, entries[index])
		if err != nil {
			return nil, err
		}
		entries[index].Capabilities = &capabilities
	}
	return entries, nil
}
func (s *Service) ListAllForManagement(
	ctx context.Context,
) (map[string][]MasterDataEntry, error) {
	entriesByType, err := s.ListAll(ctx, false)
	if err != nil {
		return nil, err
	}
	for typeName, entries := range entriesByType {
		for index := range entries {
			capabilities, err := masterDataCapabilities(ctx, s.repository.Usage, entries[index])
			if err != nil {
				return nil, err
			}
			entries[index].Capabilities = &capabilities
		}
		entriesByType[typeName] = entries
	}
	return entriesByType, nil
}
func masterDataCapabilities(
	ctx context.Context,
	usage Usage,
	entry MasterDataEntry,
) (MasterDataCapabilities, error) {
	capabilities := MasterDataCapabilities{
		CanDeactivate: entry.Active,
		CanReactivate: !entry.Active,
	}
	if entry.Origin != MasterDataOriginCustom {
		return capabilities, nil
	}
	used, err := usage(ctx, entry)
	if errors.Is(err, ErrMasterDataUsageUnknown) {
		return capabilities, nil
	}
	if err != nil {
		return capabilities, err
	}
	capabilities.CanDelete = !used
	return capabilities, nil
}
func (s *Service) SetActive(
	ctx context.Context,
	typeName, key string,
	active bool,
) (*MasterDataEntry, error) {
	typeName = strings.TrimSpace(typeName)
	key = strings.TrimSpace(key)
	if typeName == "" || key == "" {
		return nil, ErrMasterDataValidation
	}
	if err := s.repository.SetActive(ctx, typeName, key, active); err != nil {
		return nil, err
	}
	s.InvalidateCache()
	return s.Get(ctx, typeName, key)
}

func (s *Service) SetActiveMany(
	ctx context.Context,
	typeName string,
	keys []string,
	active bool,
) ([]MasterDataEntry, error) {
	typeName = strings.TrimSpace(typeName)
	normalizedKeys, err := normalizeMasterDataBatchKeys(keys)
	if typeName == "" || err != nil {
		return nil, ErrMasterDataValidation
	}

	if err := s.repository.SetActiveMany(ctx, typeName, normalizedKeys, active); err != nil {
		return nil, err
	}

	s.InvalidateCache()
	entries, err := s.ListForManagement(ctx, typeName)
	if err != nil {
		return nil, err
	}
	byKey := make(map[string]MasterDataEntry, len(entries))
	for _, entry := range entries {
		byKey[entry.Key] = entry
	}
	updated := make([]MasterDataEntry, 0, len(normalizedKeys))
	for _, key := range normalizedKeys {
		entry, exists := byKey[key]
		if !exists {
			return nil, fmt.Errorf("%w: %s/%s", ErrMasterDataNotFound, typeName, key)
		}
		updated = append(updated, entry)
	}
	return updated, nil
}
func (s *Service) Delete(ctx context.Context, typeName, key string) error {
	typeName = strings.TrimSpace(typeName)
	key = strings.TrimSpace(key)
	if typeName == "" || key == "" {
		return ErrMasterDataValidation
	}
	if typeName == StandardArticleType && isStandardArticleTypeKey(key) {
		return ErrMasterDataProtected
	}

	if err := s.repository.Delete(ctx, typeName, key); err != nil {
		return err
	}
	s.InvalidateCache()
	return nil
}

// ValidateDeletion is also rechecked against the transaction's stored identity.
func ValidateDeletion(entry MasterDataEntry, used bool) error {
	if entry.Origin == MasterDataOriginBundled {
		return fmt.Errorf("%w: %s/%s", ErrMasterDataBundled, entry.Type, entry.Key)
	}
	if used {
		return fmt.Errorf("%w: referenced master data %s/%s cannot be deleted", ErrMasterDataInUse, entry.Type, entry.Key)
	}
	return nil
}
