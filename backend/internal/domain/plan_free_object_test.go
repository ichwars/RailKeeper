package domain

import (
	"errors"
	"math"
	"strings"
	"testing"
)

func TestValidateFreePlanObjectShapeAcceptsEverySupportedKind(t *testing.T) {
	for _, shape := range []FreePlanObjectShape{
		{SchemaVersion: 1, Kind: FreePlanRectangle, WidthMM: freeFloat(300), HeightMM: freeFloat(80)},
		{SchemaVersion: 1, Kind: FreePlanEllipse, WidthMM: freeFloat(120), HeightMM: freeFloat(60)},
		{SchemaVersion: 1, Kind: FreePlanLine, EndXMM: freeFloat(200), EndYMM: freeFloat(-20)},
		{SchemaVersion: 1, Kind: FreePlanLabel, Text: "Gleis 1", FontSizeMM: freeFloat(8)},
	} {
		if err := ValidateFreePlanObjectShape(shape); err != nil {
			t.Fatalf("expected valid shape %#v, got %v", shape, err)
		}
	}
	for _, category := range []FreePlanObjectCategory{
		FreePlanStructure, FreePlanPlatform, FreePlanScenery, FreePlanAnnotation,
	} {
		if !category.Valid() {
			t.Fatalf("expected valid category %q", category)
		}
	}
}

func TestValidateFreePlanObjectShapeRejectsInvalidOrCompetingFields(t *testing.T) {
	validRectangle := FreePlanObjectShape{
		SchemaVersion: 1, Kind: FreePlanRectangle, WidthMM: freeFloat(300), HeightMM: freeFloat(80),
	}
	cases := []FreePlanObjectShape{
		{SchemaVersion: 2, Kind: FreePlanRectangle, WidthMM: freeFloat(300), HeightMM: freeFloat(80)},
		{SchemaVersion: 1, Kind: "polygon", WidthMM: freeFloat(300), HeightMM: freeFloat(80)},
		{SchemaVersion: 1, Kind: FreePlanRectangle, WidthMM: freeFloat(0), HeightMM: freeFloat(80)},
		{SchemaVersion: 1, Kind: FreePlanEllipse, WidthMM: freeFloat(math.Inf(1)), HeightMM: freeFloat(80)},
		{SchemaVersion: 1, Kind: FreePlanLine, EndXMM: freeFloat(0), EndYMM: freeFloat(0)},
		{SchemaVersion: 1, Kind: FreePlanLine, EndXMM: freeFloat(10)},
		{SchemaVersion: 1, Kind: FreePlanLabel, Text: " ", FontSizeMM: freeFloat(8)},
		{SchemaVersion: 1, Kind: FreePlanLabel, Text: strings.Repeat("x", 121), FontSizeMM: freeFloat(8)},
		{SchemaVersion: 1, Kind: FreePlanLabel, Text: "Text", FontSizeMM: freeFloat(1.9)},
		{SchemaVersion: 1, Kind: FreePlanLabel, Text: "Text", FontSizeMM: freeFloat(50.1)},
		{SchemaVersion: 1, Kind: FreePlanRectangle, WidthMM: freeFloat(300), HeightMM: freeFloat(80), Text: "x"},
	}
	for _, shape := range cases {
		if err := ValidateFreePlanObjectShape(shape); !errors.Is(err, ErrInvalidFreePlanObjectShape) {
			t.Fatalf("expected invalid shape error for %#v, got %v", shape, err)
		}
	}
	if err := ValidateFreePlanObjectShape(validRectangle); err != nil {
		t.Fatalf("valid baseline was rejected: %v", err)
	}
	if FreePlanObjectCategory("unknown").Valid() {
		t.Fatal("unknown category must be invalid")
	}
}

func freeFloat(value float64) *float64 { return &value }
