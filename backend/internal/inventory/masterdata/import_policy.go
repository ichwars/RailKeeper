package masterdata

import (
	"strings"
)

func NormalizeImportedAccessoryCustomFields(entriesByType map[string][]MasterDataEntry) error {
	for typeName, entries := range entriesByType {
		for index := range entries {
			entry := &entries[index]
			if EffectiveMasterDataType(typeName, *entry) != AccessoryCustomField {
				continue
			}
			key := strings.TrimSpace(entry.Key)
			if key == "" {
				key = SlugKey(entry.Label)
			}
			metadata := entry.Metadata
			if metadata == nil {
				metadata = map[string]any{}
			}
			normalized, err := NormalizeAccessoryCustomFieldMetadata(key, entry.Active, metadata)
			if err != nil {
				return err
			}
			entry.Metadata = normalized
		}
		entriesByType[typeName] = entries
	}
	return nil
}

func ValidateImportedArticleTypes(entriesByType map[string][]MasterDataEntry) ([]MasterDataEntry, error) {
	entries := []MasterDataEntry{}
	for typeName, typedEntries := range entriesByType {
		for _, entry := range typedEntries {
			if EffectiveMasterDataType(typeName, entry) == StandardArticleType {
				entries = append(entries, entry)
			}
		}
	}
	if err := ValidateProtectedArticleTypes(entries, false); err != nil {
		return nil, err
	}
	return entries, nil
}

func ValidateProtectedArticleTypes(entries []MasterDataEntry, required bool) error {
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

func EffectiveMasterDataType(bucketType string, entry MasterDataEntry) string {
	entryType := strings.TrimSpace(bucketType)
	if entry.Type != "" {
		entryType = strings.TrimSpace(entry.Type)
	}
	if entryType == LegacyArticleSubtype {
		return StandardAccessorySubtype
	}
	return entryType
}

func ContainsEffectiveMasterDataType(entriesByType map[string][]MasterDataEntry, typeName string) bool {
	for bucketType, entries := range entriesByType {
		for _, entry := range entries {
			if EffectiveMasterDataType(bucketType, entry) == typeName {
				return true
			}
		}
	}
	return false
}

func PrepareImportedArticleTypes(
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
