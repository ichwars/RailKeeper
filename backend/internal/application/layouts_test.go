package application

import (
	"context"
	"errors"
	"math"
	"testing"

	"railkeeper/backend/internal/domain"
)

type layoutRepositorySpy struct {
	LayoutRepository
	createdLayout      CreateLayoutInput
	savedConfiguration SaveLayoutConfigurationInput
	savedOutline       UpdateLayoutUnitOutlineInput
}

func (spy *layoutRepositorySpy) CreateLayout(
	_ context.Context,
	input CreateLayoutInput,
	_ string,
) (*Layout, error) {
	spy.createdLayout = input
	return &Layout{Name: input.Name, Gauge: input.Gauge, Scale: input.Scale}, nil
}

func (spy *layoutRepositorySpy) SaveConfiguration(
	_ context.Context,
	_ string,
	input SaveLayoutConfigurationInput,
	_ string,
) (*LayoutConfiguration, error) {
	spy.savedConfiguration = input
	return &LayoutConfiguration{Name: input.Name, Units: input.Units}, nil
}

func (spy *layoutRepositorySpy) CreateUnit(
	_ context.Context,
	_ string,
	input CreateLayoutUnitInput,
	_ string,
) (*LayoutUnit, error) {
	return &LayoutUnit{Name: input.Name, WidthMM: input.WidthMM, HeightMM: input.HeightMM}, nil
}

func (spy *layoutRepositorySpy) UpdateUnitOutline(
	_ context.Context,
	unitID string,
	input UpdateLayoutUnitOutlineInput,
	_ string,
) (*LayoutUnitOutline, error) {
	spy.savedOutline = input
	return &LayoutUnitOutline{LayoutUnitID: unitID, Points: input.Points, Version: input.ExpectedVersion + 1}, nil
}

func TestLayoutServiceValidatesUnitOutline(t *testing.T) {
	repository := &layoutRepositorySpy{}
	service := NewLayoutService(repository)
	valid := UpdateLayoutUnitOutlineInput{ExpectedVersion: 2, Points: []LayoutTwinPoint{
		{XMM: 0, YMM: 0}, {XMM: 100, YMM: 0}, {XMM: 50, YMM: 80},
	}}
	outline, err := service.UpdateUnitOutline(t.Context(), " unit-1 ", valid, "planner")
	if err != nil {
		t.Fatal(err)
	}
	if outline.LayoutUnitID != "unit-1" || outline.Version != 3 || len(repository.savedOutline.Points) != 3 {
		t.Fatalf("unexpected outline: %#v", outline)
	}
	invalid := []UpdateLayoutUnitOutlineInput{
		{ExpectedVersion: 1, Points: valid.Points[:2]},
		{ExpectedVersion: 0, Points: valid.Points},
		{ExpectedVersion: 1, Points: []LayoutTwinPoint{{XMM: 0}, {XMM: 1}, {XMM: 2}}},
		{ExpectedVersion: 1, Points: []LayoutTwinPoint{{XMM: math.NaN()}, {XMM: 1}, {YMM: 1}}},
	}
	for _, input := range invalid {
		if _, err := service.UpdateUnitOutline(t.Context(), "unit-1", input, "planner"); !errors.Is(err, ErrLayoutValidation) {
			t.Fatalf("expected invalid outline rejection for %#v, got %v", input, err)
		}
	}
}

func TestLayoutServiceNormalizesInputsBeforePersistence(t *testing.T) {
	repository := &layoutRepositorySpy{}
	service := NewLayoutService(repository)

	if _, err := service.CreateLayout(t.Context(), CreateLayoutInput{
		Name: " Home ", Kind: domain.LayoutKindPrivate, Gauge: " TT ", Scale: " 1:120 ",
	}, "planner-1"); err != nil {
		t.Fatal(err)
	}
	if repository.createdLayout.Name != "Home" || repository.createdLayout.Gauge != "TT" ||
		repository.createdLayout.Scale != "1:120" {
		t.Fatalf("unexpected normalized layout input: %#v", repository.createdLayout)
	}

	if _, err := service.SaveConfiguration(t.Context(), "layout-1", SaveLayoutConfigurationInput{
		Name: " Setup ",
		Units: []ConfigurationUnitInput{
			{UnitID: " unit-1 ", RotationDegrees: -15, SortOrder: 9},
			{UnitID: "unit-2", RotationDegrees: 721, SortOrder: 9},
		},
	}, "planner-1"); err != nil {
		t.Fatal(err)
	}
	got := repository.savedConfiguration
	if got.Name != "Setup" || got.Units[0].UnitID != "unit-1" || got.Units[0].RotationDegrees != 345 ||
		got.Units[0].SortOrder != 0 || got.Units[1].RotationDegrees != 1 || got.Units[1].SortOrder != 1 {
		t.Fatalf("unexpected normalized configuration input: %#v", got)
	}
}

func TestLayoutServiceRejectsNonFiniteConfigurationCoordinates(t *testing.T) {
	service := NewLayoutService(&layoutRepositorySpy{})
	for name, value := range map[string]float64{
		"not a number":      math.NaN(),
		"positive infinity": math.Inf(1),
		"negative infinity": math.Inf(-1),
	} {
		t.Run(name, func(t *testing.T) {
			_, err := service.SaveConfiguration(t.Context(), "layout-1", SaveLayoutConfigurationInput{
				Name: "Setup", Units: []ConfigurationUnitInput{{UnitID: "unit-1", PositionXMM: value}},
			}, "planner-1")
			if !errors.Is(err, ErrLayoutValidation) {
				t.Fatalf("expected layout validation error, got %v", err)
			}
		})
	}
}

func TestLayoutServiceRejectsNonFiniteUnitDimensions(t *testing.T) {
	service := NewLayoutService(&layoutRepositorySpy{})
	for name, value := range map[string]float64{
		"not a number": math.NaN(),
		"infinity":     math.Inf(1),
	} {
		t.Run(name, func(t *testing.T) {
			_, err := service.CreateUnit(t.Context(), "layout-1", CreateLayoutUnitInput{
				Name: "Module", Kind: domain.LayoutUnitKindModule, WidthMM: value,
			}, "planner-1")
			if !errors.Is(err, ErrLayoutValidation) {
				t.Fatalf("expected layout validation error, got %v", err)
			}
		})
	}
}
