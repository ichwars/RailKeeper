package masterdata

import (
	"context"
	"fmt"
	"strings"

	"railkeeper/backend/internal/domain"
)

func ValidateImportedAccessoryCustomFieldReferences(
	ctx context.Context,
	referencesFor AttributeReferences,
	entriesByType map[string][]MasterDataEntry,
) error {
	definitions := map[string]domain.AccessoryAttributeDefinition{}
	for typeName, entries := range entriesByType {
		for _, entry := range entries {
			if EffectiveMasterDataType(typeName, entry) != AccessoryCustomField {
				continue
			}
			key := strings.TrimSpace(entry.Key)
			if key == "" {
				key = SlugKey(entry.Label)
			}
			if _, duplicate := definitions[key]; duplicate {
				return fmt.Errorf("%w: duplicate custom field %q", ErrMasterDataValidation, key)
			}
			definition, err := ParseAccessoryCustomAttributeDefinition(key, entry.Active, entry.Metadata)
			if err != nil {
				return err
			}
			definitions[key] = definition
		}
	}

	references, err := referencesFor(ctx, "")
	if err != nil {
		return err
	}
	for _, reference := range references {
		definition, exists := definitions[reference.Value.Key]
		if !exists {
			return fmt.Errorf("%w: referenced custom field %q is missing from import",
				ErrMasterDataValidation, reference.Value.Key)
		}
		if !definition.Active {
			continue
		}
		if err := domain.ValidateControlledAccessoryAttributeValues(
			[]domain.AccessoryAttributeValue{reference.Value},
			[]domain.AccessoryAttributeDefinition{definition},
		); err != nil {
			return fmt.Errorf("%w: referenced custom field %q rejects stored attribute for product %s: %v",
				ErrMasterDataValidation, definition.Key, reference.ProductID, err)
		}
	}
	return nil
}
func ValidateAccessoryCustomFieldValues(
	ctx context.Context,
	referencesFor AttributeReferences,
	definition domain.AccessoryAttributeDefinition,
) error {
	references, err := referencesFor(ctx, definition.Key)
	if err != nil {
		return err
	}
	for _, reference := range references {
		if err := domain.ValidateControlledAccessoryAttributeValues(
			[]domain.AccessoryAttributeValue{reference.Value},
			[]domain.AccessoryAttributeDefinition{definition},
		); err != nil {
			return fmt.Errorf("%w: referenced custom field %q rejects stored attribute for product %s: %v",
				ErrMasterDataValidation, definition.Key, reference.ProductID, err)
		}
	}
	return nil
}
