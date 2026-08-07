package domain

import (
	"errors"
	"fmt"
	"strings"
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
	Key  string                 `json:"key"`
	Kind AccessoryAttributeKind `json:"kind"`
}

var standardAccessoryAttributeDefinitions = map[AccessoryArticleType][]AccessoryAttributeDefinition{
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
	return append([]AccessoryAttributeDefinition(nil), definitions...)
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
