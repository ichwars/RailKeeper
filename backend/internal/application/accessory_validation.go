package application

import (
	"fmt"
	"railkeeper/backend/internal/domain"
	"reflect"
	"strings"
)

func cleanAccessoryProductInput(input CreateAccessoryProductInput) CreateAccessoryProductInput {
	legacyInput := input.ArticleType == "" && input.Subtype == "" && input.InventoryStrategy == ""
	input.Manufacturer = strings.TrimSpace(input.Manufacturer)
	input.ArticleNumber = strings.TrimSpace(input.ArticleNumber)
	input.Name = strings.TrimSpace(input.Name)
	input.Category = strings.TrimSpace(input.Category)
	input.Description = strings.TrimSpace(input.Description)
	input.EAN = strings.TrimSpace(input.EAN)
	input.ManufacturerStatus = strings.TrimSpace(input.ManufacturerStatus)
	input.Subtype = strings.TrimSpace(input.Subtype)
	input.Scale = strings.TrimSpace(input.Scale)
	input.ListPrice = strings.TrimSpace(input.ListPrice)
	input.StockUnit = strings.TrimSpace(input.StockUnit)
	input.ManufacturerURL = strings.TrimSpace(input.ManufacturerURL)
	input.ProductURL = strings.TrimSpace(input.ProductURL)
	input.CompatibilityNotes = strings.TrimSpace(input.CompatibilityNotes)
	input.InternalNotes = strings.TrimSpace(input.InternalNotes)
	input.Gauges = cleanStringArray(input.Gauges)
	input.AlternativeNumbers = cleanStringArray(input.AlternativeNumbers)
	input.Keywords = cleanStringArray(input.Keywords)
	input.Attributes = cleanAccessoryAttributes(input.Attributes, input.ArticleType == domain.AccessoryArticleOther)
	if legacyInput {
		input.ArticleType = domain.AccessoryArticleOther
		input.Subtype = "other:other"
		input.InventoryStrategy = domain.InventoryStrategyFromTrackingMode(input.TrackingMode)
		input.PackageQuantity = 1
		input.StockUnit = "piece"
		return input
	}
	if input.InventoryStrategy.Valid() {
		input.TrackingMode = input.InventoryStrategy.TrackingMode()
	}
	if input.Category == "" {
		input.Category = input.Subtype
	}
	input.Subtype = normalizeAccessorySubtype(input.ArticleType, input.Subtype)
	return input
}

func validateAccessoryProductStructure(input CreateAccessoryProductInput) error {
	if !validAccessoryProductInput(input) {
		return ErrAccessoryValidation
	}
	if input.ArticleType != domain.AccessoryArticleOther {
		if err := domain.ValidateAccessoryAttributeValues(input.ArticleType, input.Attributes); err != nil {
			return ErrAccessoryValidation
		}
	}
	return nil
}

func validateAccessoryProductInput(
	input CreateAccessoryProductInput,
	state AccessoryProductMutationState,
) error {
	current := state.Current
	currentArticleType := domain.AccessoryArticleType("")
	currentSubtype := ""
	if current != nil {
		currentArticleType = domain.AccessoryArticleType(strings.TrimSpace(string(current.ArticleType)))
		currentSubtype = normalizeAccessorySubtype(currentArticleType, strings.TrimSpace(current.Subtype))
	}
	typeUnchanged := current != nil && currentArticleType == input.ArticleType
	if !typeUnchanged && !state.ArticleTypeActive {
		return ErrAccessoryValidation
	}
	if input.ArticleType == domain.AccessoryArticleOther {
		if err := validateControlledAccessoryAttributes(
			input.Attributes, current, state.CustomAttributeDefinitions,
		); err != nil {
			return err
		}
	}
	subtypeUnchanged := typeUnchanged && currentSubtype == input.Subtype
	if subtypeUnchanged {
		return nil
	}
	if !state.SubtypeActive {
		return ErrAccessoryValidation
	}
	return nil
}

func validateControlledAccessoryAttributes(
	values []domain.AccessoryAttributeValue,
	current *AccessoryProduct,
	definitions []domain.AccessoryAttributeDefinition,
) error {
	if err := domain.ValidateAccessoryAttributeValues(domain.AccessoryArticleOther, values); err != nil {
		return fmt.Errorf("%w: %v", ErrAccessoryValidation, err)
	}
	active := make([]domain.AccessoryAttributeDefinition, 0, len(definitions))
	inactiveByKey := make(map[string]domain.AccessoryAttributeDefinition)
	for _, definition := range definitions {
		if definition.Active {
			active = append(active, definition)
		} else {
			inactiveByKey[definition.Key] = definition
		}
	}

	valuesToValidate := values
	if current != nil && current.ArticleType == domain.AccessoryArticleOther {
		currentByKey := accessoryAttributesByKey(current.Attributes)
		valuesByKey := accessoryAttributesByKey(values)
		for key := range inactiveByKey {
			currentValue, existed := currentByKey[key]
			value, preserved := valuesByKey[key]
			if existed && (!preserved || !reflect.DeepEqual(value, currentValue)) {
				return fmt.Errorf("%w: inactive custom attribute %q must remain unchanged", ErrAccessoryValidation, key)
			}
			if preserved && (!existed || !reflect.DeepEqual(value, currentValue)) {
				return fmt.Errorf("%w: inactive custom attribute %q cannot be added or changed", ErrAccessoryValidation, key)
			}
		}
		valuesToValidate = make([]domain.AccessoryAttributeValue, 0, len(values))
		for _, value := range values {
			if _, inactive := inactiveByKey[value.Key]; !inactive {
				valuesToValidate = append(valuesToValidate, value)
			}
		}
	}
	if err := domain.ValidateControlledAccessoryAttributeValues(valuesToValidate, active); err != nil {
		return fmt.Errorf("%w: %v", ErrAccessoryValidation, err)
	}
	return nil
}

func accessoryAttributesByKey(values []domain.AccessoryAttributeValue) map[string]domain.AccessoryAttributeValue {
	out := make(map[string]domain.AccessoryAttributeValue, len(values))
	for _, value := range values {
		out[value.Key] = value
	}
	return out
}

func validAccessoryProductInput(input CreateAccessoryProductInput) bool {
	return input.Manufacturer != "" && input.Name != "" && input.Category != "" && input.TrackingMode.Valid() &&
		input.ArticleType.Valid() && accessorySubtypeMatchesType(input.ArticleType, input.Subtype) &&
		input.InventoryStrategy.Valid() && input.PackageQuantity > 0 && input.StockUnit != "" &&
		input.MinimumStock >= 0 && validAccessoryListPrice(input.ListPrice)
}

func validAccessoryListPrice(value string) bool {
	if value == "" {
		return true
	}
	_, ok := parseMoneyCents(value)
	return ok
}

func normalizeAccessorySubtype(articleType domain.AccessoryArticleType, subtype string) string {
	if subtype == "" || strings.Contains(subtype, ":") {
		return subtype
	}
	return string(articleType) + ":" + subtype
}

func accessorySubtypeMatchesType(articleType domain.AccessoryArticleType, subtype string) bool {
	if subtype == "" {
		return false
	}
	prefix := string(articleType) + ":"
	return strings.HasPrefix(subtype, prefix) && len(subtype) > len(prefix)
}

func cleanStringArray(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	cleaned := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || containsFold(cleaned, value) {
			continue
		}
		cleaned = append(cleaned, value)
	}
	return cleaned
}

func cleanAccessoryAttributes(
	values []domain.AccessoryAttributeValue,
	preserveOptionDuplicates bool,
) []domain.AccessoryAttributeValue {
	for index := range values {
		value := &values[index]
		value.Key = strings.TrimSpace(value.Key)
		if value.TextValue != nil {
			trimmed := strings.TrimSpace(*value.TextValue)
			value.TextValue = &trimmed
		}
		if value.DateValue != nil {
			trimmed := strings.TrimSpace(*value.DateValue)
			value.DateValue = &trimmed
		}
		if value.Unit != nil {
			trimmed := strings.TrimSpace(*value.Unit)
			value.Unit = &trimmed
		}
		if preserveOptionDuplicates {
			for optionIndex := range value.OptionValues {
				value.OptionValues[optionIndex] = strings.TrimSpace(value.OptionValues[optionIndex])
			}
		} else {
			value.OptionValues = cleanStringArray(value.OptionValues)
		}
	}
	return values
}

func containsFold(values []string, candidate string) bool {
	for _, value := range values {
		if strings.EqualFold(value, candidate) {
			return true
		}
	}
	return false
}

func cleanStorageLocationInput(input CreateStorageLocationInput) CreateStorageLocationInput {
	input.ParentID = strings.TrimSpace(input.ParentID)
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	return input
}
