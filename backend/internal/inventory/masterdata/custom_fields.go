package masterdata

import (
	"encoding/json"
	"fmt"
	"strings"

	"railkeeper/backend/internal/domain"
)

func ParseAccessoryCustomAttributeDefinition(
	key string,
	active bool,
	metadata map[string]any,
) (domain.AccessoryAttributeDefinition, error) {
	definition := domain.AccessoryAttributeDefinition{Key: strings.TrimSpace(key), Active: active}
	kind, ok := metadata["kind"].(string)
	if !ok {
		return definition, fmt.Errorf("%w: custom field %q requires a valid kind", ErrMasterDataValidation, key)
	}
	definition.Kind = domain.AccessoryAttributeKind(strings.TrimSpace(kind))

	unitValue, hasUnit := metadata["unit"]
	minimumValue, hasMinimum := metadata["min"]
	maximumValue, hasMaximum := metadata["max"]
	optionsValue, hasOptions := metadata["options"]
	if definition.Kind != domain.AccessoryAttributeNumber && (hasUnit || hasMinimum || hasMaximum) {
		return definition, fmt.Errorf("%w: custom field %q has numeric metadata for a non-number kind", ErrMasterDataValidation, key)
	}
	if hasUnit {
		unit, ok := unitValue.(string)
		if !ok {
			return definition, fmt.Errorf("%w: custom field %q unit must be a string", ErrMasterDataValidation, key)
		}
		definition.Unit = strings.TrimSpace(unit)
	}
	if hasMinimum {
		minimum, ok := masterDataFloat(minimumValue)
		if !ok {
			return definition, fmt.Errorf("%w: custom field %q minimum must be a number", ErrMasterDataValidation, key)
		}
		definition.Minimum = &minimum
	}
	if hasMaximum {
		maximum, ok := masterDataFloat(maximumValue)
		if !ok {
			return definition, fmt.Errorf("%w: custom field %q maximum must be a number", ErrMasterDataValidation, key)
		}
		definition.Maximum = &maximum
	}

	isSelect := definition.Kind == domain.AccessoryAttributeSingleSelect ||
		definition.Kind == domain.AccessoryAttributeMultiSelect
	if isSelect != hasOptions {
		return definition, fmt.Errorf("%w: custom field %q selection options are invalid", ErrMasterDataValidation, key)
	}
	if hasOptions {
		options, ok := masterDataStrings(optionsValue)
		if !ok {
			return definition, fmt.Errorf("%w: custom field %q options must be strings", ErrMasterDataValidation, key)
		}
		definition.Options = options
	}
	if err := definition.Validate(); err != nil {
		return definition, fmt.Errorf("%w: custom field %q: %v", ErrMasterDataValidation, key, err)
	}
	return definition, nil
}

func NormalizeAccessoryCustomFieldMetadata(
	key string,
	active bool,
	metadata map[string]any,
) (map[string]any, error) {
	definition, err := ParseAccessoryCustomAttributeDefinition(key, active, metadata)
	if err != nil {
		return nil, err
	}
	normalized := make(map[string]any, len(metadata))
	for name, value := range metadata {
		normalized[name] = value
	}
	normalized["kind"] = string(definition.Kind)
	if _, exists := metadata["unit"]; exists {
		normalized["unit"] = definition.Unit
	}
	if definition.Minimum != nil {
		normalized["min"] = *definition.Minimum
	}
	if definition.Maximum != nil {
		normalized["max"] = *definition.Maximum
	}
	if definition.Options != nil {
		normalized["options"] = append([]string(nil), definition.Options...)
	}
	return normalized, nil
}

func masterDataStrings(value any) ([]string, bool) {
	var raw []any
	switch typed := value.(type) {
	case []string:
		out := make([]string, len(typed))
		for index, option := range typed {
			out[index] = strings.TrimSpace(option)
		}
		return out, true
	case []any:
		raw = typed
	default:
		return nil, false
	}
	out := make([]string, len(raw))
	for index, value := range raw {
		option, ok := value.(string)
		if !ok {
			return nil, false
		}
		out[index] = strings.TrimSpace(option)
	}
	return out, true
}

func masterDataFloat(value any) (float64, bool) {
	switch typed := value.(type) {
	case float64:
		return typed, true
	case float32:
		return float64(typed), true
	case int:
		return float64(typed), true
	case int8:
		return float64(typed), true
	case int16:
		return float64(typed), true
	case int32:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case uint:
		return float64(typed), true
	case uint8:
		return float64(typed), true
	case uint16:
		return float64(typed), true
	case uint32:
		return float64(typed), true
	case uint64:
		return float64(typed), true
	case json.Number:
		number, err := typed.Float64()
		return number, err == nil
	default:
		return 0, false
	}
}
