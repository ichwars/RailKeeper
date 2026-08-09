package domain

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

var ErrAccessoryAttributeValidation = errors.New("accessory attribute validation failed")

type AccessoryAttributeKind string

const (
	AccessoryAttributeText         AccessoryAttributeKind = "text"
	AccessoryAttributeNumber       AccessoryAttributeKind = "number"
	AccessoryAttributeBoolean      AccessoryAttributeKind = "boolean"
	AccessoryAttributeDate         AccessoryAttributeKind = "date"
	AccessoryAttributeSingleSelect AccessoryAttributeKind = "single_select"
	AccessoryAttributeMultiSelect  AccessoryAttributeKind = "multi_select"
)

func (kind AccessoryAttributeKind) Valid() bool {
	switch kind {
	case AccessoryAttributeText, AccessoryAttributeNumber, AccessoryAttributeBoolean,
		AccessoryAttributeDate, AccessoryAttributeSingleSelect, AccessoryAttributeMultiSelect:
		return true
	default:
		return false
	}
}

type AccessoryAttributeValue struct {
	Key          string                 `json:"key"`
	Kind         AccessoryAttributeKind `json:"kind"`
	TextValue    *string                `json:"textValue,omitempty"`
	NumberValue  *float64               `json:"numberValue,omitempty"`
	BooleanValue *bool                  `json:"booleanValue,omitempty"`
	DateValue    *string                `json:"dateValue,omitempty"`
	OptionValues []string               `json:"optionValues,omitempty"`
	Unit         *string                `json:"unit,omitempty"`
}

type AccessoryAttributeDefinition struct {
	Key     string                 `json:"key"`
	Kind    AccessoryAttributeKind `json:"kind"`
	Active  bool                   `json:"active"`
	Unit    string                 `json:"unit,omitempty"`
	Minimum *float64               `json:"minimum,omitempty"`
	Maximum *float64               `json:"maximum,omitempty"`
	Options []string               `json:"options,omitempty"`
}

type standardAccessoryAttributeDefinition struct {
	Key  string
	Kind AccessoryAttributeKind
}

var standardAccessoryAttributeDefinitions = map[AccessoryArticleType][]standardAccessoryAttributeDefinition{
	AccessoryArticleTrack: {
		{"trackSystem", AccessoryAttributeText}, {"lengthMm", AccessoryAttributeNumber},
		{"radiusMm", AccessoryAttributeNumber}, {"angleDegrees", AccessoryAttributeNumber},
		{"direction", AccessoryAttributeSingleSelect}, {"frogAngleDegrees", AccessoryAttributeNumber},
		{"sleeperType", AccessoryAttributeText}, {"railHeightMm", AccessoryAttributeNumber},
		{"roadbed", AccessoryAttributeBoolean}, {"connectionCount", AccessoryAttributeNumber},
		{"digitalReady", AccessoryAttributeBoolean},
	},
	AccessoryArticleSignal: {
		{"prototype", AccessoryAttributeText}, {"epoch", AccessoryAttributeMultiSelect},
		{"aspects", AccessoryAttributeMultiSelect}, {"ledCount", AccessoryAttributeNumber},
		{"heightMm", AccessoryAttributeNumber}, {"voltageAC", AccessoryAttributeNumber},
		{"voltageDC", AccessoryAttributeNumber}, {"mounting", AccessoryAttributeSingleSelect},
		{"driveType", AccessoryAttributeSingleSelect}, {"integratedDecoder", AccessoryAttributeBoolean},
		{"controlModule", AccessoryAttributeText},
	},
	AccessoryArticleDecoder: {
		{"interface", AccessoryAttributeSingleSelect}, {"protocols", AccessoryAttributeMultiSelect},
		{"functionOutputs", AccessoryAttributeNumber}, {"motorCurrentMa", AccessoryAttributeNumber},
		{"outputCurrentMa", AccessoryAttributeNumber}, {"totalCurrentMa", AccessoryAttributeNumber},
		{"railCom", AccessoryAttributeBoolean}, {"susi", AccessoryAttributeBoolean},
		{"servoOutputs", AccessoryAttributeNumber}, {"dimensions", AccessoryAttributeText},
		{"firmware", AccessoryAttributeText},
	},
	AccessoryArticleElectricalControl: {
		{"inputVoltage", AccessoryAttributeNumber}, {"outputVoltage", AccessoryAttributeNumber},
		{"currentA", AccessoryAttributeNumber}, {"powerW", AccessoryAttributeNumber},
		{"channelCount", AccessoryAttributeNumber}, {"protocols", AccessoryAttributeMultiSelect},
		{"connectors", AccessoryAttributeMultiSelect}, {"protections", AccessoryAttributeMultiSelect},
		{"compatibleArticles", AccessoryAttributeMultiSelect},
	},
	AccessoryArticleBuildingEquipment: {
		{"epoch", AccessoryAttributeMultiSelect}, {"dimensions", AccessoryAttributeText},
		{"footprint", AccessoryAttributeText}, {"material", AccessoryAttributeText},
		{"constructionType", AccessoryAttributeSingleSelect}, {"partCount", AccessoryAttributeNumber},
		{"difficulty", AccessoryAttributeSingleSelect}, {"lightingOptions", AccessoryAttributeMultiSelect},
		{"floorPlanAvailable", AccessoryAttributeBoolean},
	},
	AccessoryArticleLandscapeConsumable: {
		{"material", AccessoryAttributeText}, {"color", AccessoryAttributeText},
		{"season", AccessoryAttributeText}, {"content", AccessoryAttributeNumber},
		{"contentUnit", AccessoryAttributeSingleSelect}, {"fiberOrGrainSize", AccessoryAttributeText},
		{"coverage", AccessoryAttributeText}, {"suitableScales", AccessoryAttributeMultiSelect},
		{"safetyNotes", AccessoryAttributeText},
	},
	AccessoryArticleLighting: {
		{"lightColor", AccessoryAttributeText}, {"colorTemperatureK", AccessoryAttributeNumber},
		{"voltage", AccessoryAttributeNumber}, {"currentMa", AccessoryAttributeNumber},
		{"powerType", AccessoryAttributeSingleSelect}, {"ledCount", AccessoryAttributeNumber},
		{"dimmable", AccessoryAttributeBoolean}, {"dimensions", AccessoryAttributeText},
		{"mounting", AccessoryAttributeSingleSelect},
	},
}

func StandardAccessoryAttributeDefinitions(articleType AccessoryArticleType) []AccessoryAttributeDefinition {
	definitions := standardAccessoryAttributeDefinitions[articleType]
	out := make([]AccessoryAttributeDefinition, len(definitions))
	for index, definition := range definitions {
		out[index] = AccessoryAttributeDefinition{Key: definition.Key, Kind: definition.Kind}
	}
	return out
}

func ValidateAccessoryAttributeValues(articleType AccessoryArticleType, values []AccessoryAttributeValue) error {
	if !articleType.Valid() {
		return fmt.Errorf("%w: invalid article type", ErrAccessoryAttributeValidation)
	}
	definitions := standardAttributeKinds(articleType)
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if err := value.Validate(); err != nil {
			return err
		}
		if _, exists := seen[value.Key]; exists {
			return fmt.Errorf("%w: duplicate key %q", ErrAccessoryAttributeValidation, value.Key)
		}
		seen[value.Key] = struct{}{}
		expectedKind, standard := definitions[value.Key]
		if articleType == AccessoryArticleOther {
			continue
		}
		if !standard {
			return fmt.Errorf("%w: unsupported key %q", ErrAccessoryAttributeValidation, value.Key)
		}
		if value.Kind != expectedKind {
			return fmt.Errorf("%w: key %q requires %q", ErrAccessoryAttributeValidation, value.Key, expectedKind)
		}
	}
	return nil
}

func ValidateControlledAccessoryAttributeValues(
	values []AccessoryAttributeValue,
	definitions []AccessoryAttributeDefinition,
) error {
	definitionsByKey := make(map[string]AccessoryAttributeDefinition, len(definitions))
	for _, definition := range definitions {
		if err := definition.Validate(); err != nil {
			return err
		}
		if _, exists := definitionsByKey[definition.Key]; exists {
			return fmt.Errorf("%w: duplicate definition %q", ErrAccessoryAttributeValidation, definition.Key)
		}
		definitionsByKey[definition.Key] = definition
	}

	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if err := value.Validate(); err != nil {
			return err
		}
		if _, exists := seen[value.Key]; exists {
			return fmt.Errorf("%w: duplicate key %q", ErrAccessoryAttributeValidation, value.Key)
		}
		seen[value.Key] = struct{}{}
		definition, exists := definitionsByKey[value.Key]
		if !exists {
			return fmt.Errorf("%w: undefined controlled key %q", ErrAccessoryAttributeValidation, value.Key)
		}
		if value.Kind != definition.Kind {
			return fmt.Errorf("%w: key %q requires %q", ErrAccessoryAttributeValidation, value.Key, definition.Kind)
		}
		if err := definition.validateValue(value); err != nil {
			return err
		}
	}
	return nil
}

func (definition AccessoryAttributeDefinition) Validate() error {
	if strings.TrimSpace(definition.Key) == "" || definition.Key != strings.TrimSpace(definition.Key) ||
		!definition.Kind.Valid() {
		return fmt.Errorf("%w: invalid controlled definition", ErrAccessoryAttributeValidation)
	}
	if definition.Kind != AccessoryAttributeNumber &&
		(definition.Unit != "" || definition.Minimum != nil || definition.Maximum != nil) {
		return fmt.Errorf("%w: numeric constraints require number kind", ErrAccessoryAttributeValidation)
	}
	if definition.Kind == AccessoryAttributeNumber {
		if definition.Unit != strings.TrimSpace(definition.Unit) ||
			!validAccessoryAttributeBound(definition.Minimum) || !validAccessoryAttributeBound(definition.Maximum) {
			return fmt.Errorf("%w: invalid numeric definition", ErrAccessoryAttributeValidation)
		}
		if definition.Minimum != nil && definition.Maximum != nil && *definition.Minimum > *definition.Maximum {
			return fmt.Errorf("%w: minimum exceeds maximum", ErrAccessoryAttributeValidation)
		}
	}
	isSelect := definition.Kind == AccessoryAttributeSingleSelect || definition.Kind == AccessoryAttributeMultiSelect
	if isSelect != (len(definition.Options) > 0) {
		return fmt.Errorf("%w: selection options are invalid", ErrAccessoryAttributeValidation)
	}
	seen := make(map[string]struct{}, len(definition.Options))
	for _, option := range definition.Options {
		if option == "" || option != strings.TrimSpace(option) {
			return fmt.Errorf("%w: selection option is invalid", ErrAccessoryAttributeValidation)
		}
		if _, exists := seen[option]; exists {
			return fmt.Errorf("%w: duplicate selection option %q", ErrAccessoryAttributeValidation, option)
		}
		seen[option] = struct{}{}
	}
	return nil
}

func (definition AccessoryAttributeDefinition) validateValue(value AccessoryAttributeValue) error {
	switch value.Kind {
	case AccessoryAttributeNumber:
		if definition.Unit == "" {
			if value.Unit != nil && *value.Unit != "" {
				return fmt.Errorf("%w: key %q does not accept a unit", ErrAccessoryAttributeValidation, value.Key)
			}
		} else if value.Unit == nil || *value.Unit != definition.Unit {
			return fmt.Errorf("%w: key %q requires unit %q", ErrAccessoryAttributeValidation, value.Key, definition.Unit)
		}
		if value.NumberValue == nil || math.IsNaN(*value.NumberValue) || math.IsInf(*value.NumberValue, 0) ||
			(definition.Minimum != nil && *value.NumberValue < *definition.Minimum) ||
			(definition.Maximum != nil && *value.NumberValue > *definition.Maximum) {
			return fmt.Errorf("%w: key %q is outside configured bounds", ErrAccessoryAttributeValidation, value.Key)
		}
	case AccessoryAttributeDate:
		if value.DateValue == nil {
			return fmt.Errorf("%w: key %q requires a date", ErrAccessoryAttributeValidation, value.Key)
		}
		if parsed, err := time.Parse("2006-01-02", *value.DateValue); err != nil ||
			parsed.Format("2006-01-02") != *value.DateValue {
			return fmt.Errorf("%w: key %q requires an ISO date", ErrAccessoryAttributeValidation, value.Key)
		}
	case AccessoryAttributeSingleSelect, AccessoryAttributeMultiSelect:
		allowed := make(map[string]struct{}, len(definition.Options))
		for _, option := range definition.Options {
			allowed[option] = struct{}{}
		}
		seen := make(map[string]struct{}, len(value.OptionValues))
		for _, option := range value.OptionValues {
			if _, exists := allowed[option]; !exists {
				return fmt.Errorf("%w: key %q contains unsupported option %q", ErrAccessoryAttributeValidation, value.Key, option)
			}
			if _, exists := seen[option]; exists {
				return fmt.Errorf("%w: key %q contains duplicate option %q", ErrAccessoryAttributeValidation, value.Key, option)
			}
			seen[option] = struct{}{}
		}
	}
	return nil
}

func validAccessoryAttributeBound(value *float64) bool {
	return value == nil || (!math.IsNaN(*value) && !math.IsInf(*value, 0))
}

func (value AccessoryAttributeValue) Validate() error {
	if strings.TrimSpace(value.Key) == "" || !value.Kind.Valid() {
		return ErrAccessoryAttributeValidation
	}
	valueCount := 0
	if value.TextValue != nil {
		valueCount++
	}
	if value.NumberValue != nil {
		valueCount++
	}
	if value.BooleanValue != nil {
		valueCount++
	}
	if value.DateValue != nil {
		valueCount++
	}
	if value.OptionValues != nil {
		valueCount++
	}
	if valueCount != 1 || (value.Unit != nil && value.Kind != AccessoryAttributeNumber) {
		return ErrAccessoryAttributeValidation
	}
	switch value.Kind {
	case AccessoryAttributeText:
		if value.TextValue == nil {
			return ErrAccessoryAttributeValidation
		}
	case AccessoryAttributeNumber:
		if value.NumberValue == nil {
			return ErrAccessoryAttributeValidation
		}
	case AccessoryAttributeBoolean:
		if value.BooleanValue == nil {
			return ErrAccessoryAttributeValidation
		}
	case AccessoryAttributeDate:
		if value.DateValue == nil {
			return ErrAccessoryAttributeValidation
		}
	case AccessoryAttributeSingleSelect:
		if len(value.OptionValues) != 1 {
			return ErrAccessoryAttributeValidation
		}
	case AccessoryAttributeMultiSelect:
		if len(value.OptionValues) == 0 {
			return ErrAccessoryAttributeValidation
		}
	}
	return nil
}

func standardAttributeKinds(articleType AccessoryArticleType) map[string]AccessoryAttributeKind {
	definitions := standardAccessoryAttributeDefinitions[articleType]
	kinds := make(map[string]AccessoryAttributeKind, len(definitions))
	for _, definition := range definitions {
		kinds[definition.Key] = definition.Kind
	}
	return kinds
}
