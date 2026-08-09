package domain_test

import (
	"errors"
	"testing"

	"railkeeper/backend/internal/domain"
)

func TestStandardAccessoryAttributeDefinitionsMatchCatalog(t *testing.T) {
	expected := map[domain.AccessoryArticleType]map[string]domain.AccessoryAttributeKind{
		domain.AccessoryArticleTrack: {
			"trackSystem": domain.AccessoryAttributeText, "lengthMm": domain.AccessoryAttributeNumber,
			"radiusMm": domain.AccessoryAttributeNumber, "angleDegrees": domain.AccessoryAttributeNumber,
			"direction": domain.AccessoryAttributeSingleSelect, "frogAngleDegrees": domain.AccessoryAttributeNumber,
			"sleeperType": domain.AccessoryAttributeText, "railHeightMm": domain.AccessoryAttributeNumber,
			"roadbed": domain.AccessoryAttributeBoolean, "connectionCount": domain.AccessoryAttributeNumber,
			"digitalReady": domain.AccessoryAttributeBoolean,
		},
		domain.AccessoryArticleSignal: {
			"prototype": domain.AccessoryAttributeText, "epoch": domain.AccessoryAttributeMultiSelect,
			"aspects": domain.AccessoryAttributeMultiSelect, "ledCount": domain.AccessoryAttributeNumber,
			"heightMm": domain.AccessoryAttributeNumber, "voltageAC": domain.AccessoryAttributeNumber,
			"voltageDC": domain.AccessoryAttributeNumber, "mounting": domain.AccessoryAttributeSingleSelect,
			"driveType": domain.AccessoryAttributeSingleSelect, "integratedDecoder": domain.AccessoryAttributeBoolean,
			"controlModule": domain.AccessoryAttributeText,
		},
		domain.AccessoryArticleDecoder: {
			"interface": domain.AccessoryAttributeSingleSelect, "protocols": domain.AccessoryAttributeMultiSelect,
			"functionOutputs": domain.AccessoryAttributeNumber, "motorCurrentMa": domain.AccessoryAttributeNumber,
			"outputCurrentMa": domain.AccessoryAttributeNumber, "totalCurrentMa": domain.AccessoryAttributeNumber,
			"railCom": domain.AccessoryAttributeBoolean, "susi": domain.AccessoryAttributeBoolean,
			"servoOutputs": domain.AccessoryAttributeNumber, "dimensions": domain.AccessoryAttributeText,
			"firmware": domain.AccessoryAttributeText,
		},
		domain.AccessoryArticleElectricalControl: {
			"inputVoltage": domain.AccessoryAttributeNumber, "outputVoltage": domain.AccessoryAttributeNumber,
			"currentA": domain.AccessoryAttributeNumber, "powerW": domain.AccessoryAttributeNumber,
			"channelCount": domain.AccessoryAttributeNumber, "protocols": domain.AccessoryAttributeMultiSelect,
			"connectors": domain.AccessoryAttributeMultiSelect, "protections": domain.AccessoryAttributeMultiSelect,
			"compatibleArticles": domain.AccessoryAttributeMultiSelect,
		},
		domain.AccessoryArticleBuildingEquipment: {
			"epoch": domain.AccessoryAttributeMultiSelect, "dimensions": domain.AccessoryAttributeText,
			"footprint": domain.AccessoryAttributeText, "material": domain.AccessoryAttributeText,
			"constructionType": domain.AccessoryAttributeSingleSelect, "partCount": domain.AccessoryAttributeNumber,
			"difficulty": domain.AccessoryAttributeSingleSelect, "lightingOptions": domain.AccessoryAttributeMultiSelect,
			"floorPlanAvailable": domain.AccessoryAttributeBoolean,
		},
		domain.AccessoryArticleLandscapeConsumable: {
			"material": domain.AccessoryAttributeText, "color": domain.AccessoryAttributeText,
			"season": domain.AccessoryAttributeText, "content": domain.AccessoryAttributeNumber,
			"contentUnit": domain.AccessoryAttributeSingleSelect, "fiberOrGrainSize": domain.AccessoryAttributeText,
			"coverage": domain.AccessoryAttributeText, "suitableScales": domain.AccessoryAttributeMultiSelect,
			"safetyNotes": domain.AccessoryAttributeText,
		},
		domain.AccessoryArticleLighting: {
			"lightColor": domain.AccessoryAttributeText, "colorTemperatureK": domain.AccessoryAttributeNumber,
			"voltage": domain.AccessoryAttributeNumber, "currentMa": domain.AccessoryAttributeNumber,
			"powerType": domain.AccessoryAttributeSingleSelect, "ledCount": domain.AccessoryAttributeNumber,
			"dimmable": domain.AccessoryAttributeBoolean, "dimensions": domain.AccessoryAttributeText,
			"mounting": domain.AccessoryAttributeSingleSelect,
		},
	}

	for articleType, fields := range expected {
		t.Run(string(articleType), func(t *testing.T) {
			definitions := domain.StandardAccessoryAttributeDefinitions(articleType)
			if len(definitions) != len(fields) {
				t.Fatalf("definition count = %d, want %d", len(definitions), len(fields))
			}
			for _, definition := range definitions {
				if got, ok := fields[definition.Key]; !ok || got != definition.Kind {
					t.Fatalf("unexpected definition %#v", definition)
				}
			}
		})
	}
}

func TestValidateAccessoryAttributeValuesEnforcesTypedUnionAndCatalog(t *testing.T) {
	text := "TT"
	number := 12.5
	boolean := false
	date := "2026-08-08"
	tests := []struct {
		name        string
		articleType domain.AccessoryArticleType
		attributes  []domain.AccessoryAttributeValue
		wantValid   bool
	}{
		{"text", domain.AccessoryArticleOther, []domain.AccessoryAttributeValue{{Key: "custom", Kind: domain.AccessoryAttributeText, TextValue: &text}}, true},
		{"number with unit", domain.AccessoryArticleOther, []domain.AccessoryAttributeValue{{Key: "custom", Kind: domain.AccessoryAttributeNumber, NumberValue: &number, Unit: stringPointer("mm")}}, true},
		{"boolean false", domain.AccessoryArticleOther, []domain.AccessoryAttributeValue{{Key: "custom", Kind: domain.AccessoryAttributeBoolean, BooleanValue: &boolean}}, true},
		{"date", domain.AccessoryArticleOther, []domain.AccessoryAttributeValue{{Key: "custom", Kind: domain.AccessoryAttributeDate, DateValue: &date}}, true},
		{"single select", domain.AccessoryArticleOther, []domain.AccessoryAttributeValue{{Key: "custom", Kind: domain.AccessoryAttributeSingleSelect, OptionValues: []string{"DCC"}}}, true},
		{"multi select", domain.AccessoryArticleOther, []domain.AccessoryAttributeValue{{Key: "custom", Kind: domain.AccessoryAttributeMultiSelect, OptionValues: []string{"DCC", "MM"}}}, true},
		{"number without value", domain.AccessoryArticleOther, []domain.AccessoryAttributeValue{{Key: "custom", Kind: domain.AccessoryAttributeNumber}}, false},
		{"text with number", domain.AccessoryArticleOther, []domain.AccessoryAttributeValue{{Key: "custom", Kind: domain.AccessoryAttributeText, TextValue: &text, NumberValue: &number}}, false},
		{"unit outside number", domain.AccessoryArticleOther, []domain.AccessoryAttributeValue{{Key: "custom", Kind: domain.AccessoryAttributeText, TextValue: &text, Unit: stringPointer("mm")}}, false},
		{"wrong standard kind", domain.AccessoryArticleTrack, []domain.AccessoryAttributeValue{{Key: "lengthMm", Kind: domain.AccessoryAttributeText, TextValue: &text}}, false},
		{"custom field outside other", domain.AccessoryArticleTrack, []domain.AccessoryAttributeValue{{Key: "custom", Kind: domain.AccessoryAttributeText, TextValue: &text}}, false},
		{"duplicate keys", domain.AccessoryArticleOther, []domain.AccessoryAttributeValue{{Key: "custom", Kind: domain.AccessoryAttributeText, TextValue: &text}, {Key: "custom", Kind: domain.AccessoryAttributeText, TextValue: &text}}, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := domain.ValidateAccessoryAttributeValues(test.articleType, test.attributes)
			if (err == nil) != test.wantValid {
				t.Fatalf("ValidateAccessoryAttributeValues() error = %v, want valid %t", err, test.wantValid)
			}
		})
	}

	if got := domain.StandardAccessoryAttributeDefinitions(domain.AccessoryArticleOther); len(got) != 0 {
		t.Fatalf("other definitions = %#v, want nil", got)
	}
}

func stringPointer(value string) *string { return &value }

func TestValidateControlledAccessoryAttributeValuesEnforcesConfiguredDefinitions(t *testing.T) {
	text := "club"
	number := 12.5
	boolean := false
	date := "2026-08-08"
	definitions := []domain.AccessoryAttributeDefinition{
		{Key: "text", Kind: domain.AccessoryAttributeText},
		{Key: "number", Kind: domain.AccessoryAttributeNumber, Unit: "mm", Minimum: floatPointer(10), Maximum: floatPointer(20)},
		{Key: "boolean", Kind: domain.AccessoryAttributeBoolean},
		{Key: "date", Kind: domain.AccessoryAttributeDate},
		{Key: "single", Kind: domain.AccessoryAttributeSingleSelect, Options: []string{"DCC", "MM"}},
		{Key: "multi", Kind: domain.AccessoryAttributeMultiSelect, Options: []string{"DCC", "MM"}},
	}
	valid := []domain.AccessoryAttributeValue{
		{Key: "text", Kind: domain.AccessoryAttributeText, TextValue: &text},
		{Key: "number", Kind: domain.AccessoryAttributeNumber, NumberValue: &number, Unit: stringPointer("mm")},
		{Key: "boolean", Kind: domain.AccessoryAttributeBoolean, BooleanValue: &boolean},
		{Key: "date", Kind: domain.AccessoryAttributeDate, DateValue: &date},
		{Key: "single", Kind: domain.AccessoryAttributeSingleSelect, OptionValues: []string{"DCC"}},
		{Key: "multi", Kind: domain.AccessoryAttributeMultiSelect, OptionValues: []string{"DCC", "MM"}},
	}
	if err := domain.ValidateControlledAccessoryAttributeValues(valid, definitions); err != nil {
		t.Fatalf("valid configured attributes rejected: %v", err)
	}

	wrongUnit := "cm"
	tooSmall := 9.9
	invalidDate := "08.08.2026"
	tests := []struct {
		name  string
		value domain.AccessoryAttributeValue
	}{
		{"undefined key", domain.AccessoryAttributeValue{Key: "unknown", Kind: domain.AccessoryAttributeText, TextValue: &text}},
		{"incompatible kind", domain.AccessoryAttributeValue{Key: "text", Kind: domain.AccessoryAttributeNumber, NumberValue: &number}},
		{"wrong unit", domain.AccessoryAttributeValue{Key: "number", Kind: domain.AccessoryAttributeNumber, NumberValue: &number, Unit: &wrongUnit}},
		{"below bound", domain.AccessoryAttributeValue{Key: "number", Kind: domain.AccessoryAttributeNumber, NumberValue: &tooSmall, Unit: stringPointer("mm")}},
		{"invalid date", domain.AccessoryAttributeValue{Key: "date", Kind: domain.AccessoryAttributeDate, DateValue: &invalidDate}},
		{"unknown single option", domain.AccessoryAttributeValue{Key: "single", Kind: domain.AccessoryAttributeSingleSelect, OptionValues: []string{"MFX"}}},
		{"unknown multi option", domain.AccessoryAttributeValue{Key: "multi", Kind: domain.AccessoryAttributeMultiSelect, OptionValues: []string{"DCC", "MFX"}}},
		{"duplicate multi option", domain.AccessoryAttributeValue{Key: "multi", Kind: domain.AccessoryAttributeMultiSelect, OptionValues: []string{"DCC", "DCC"}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := domain.ValidateControlledAccessoryAttributeValues(
				[]domain.AccessoryAttributeValue{test.value}, definitions)
			if !errors.Is(err, domain.ErrAccessoryAttributeValidation) {
				t.Fatalf("error = %v, want controlled validation failure", err)
			}
		})
	}
}

func floatPointer(value float64) *float64 { return &value }
