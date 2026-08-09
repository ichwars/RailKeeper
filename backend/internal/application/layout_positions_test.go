package application

import (
	"context"
	"errors"
	"math"
	"testing"

	"railkeeper/backend/internal/domain"
)

type layoutPositionRepositorySpy struct {
	LayoutRepository
	created CreateLayoutTechnicalPositionInput
	updated UpdateLayoutTechnicalPositionInput
}

func (spy *layoutPositionRepositorySpy) CreateTechnicalPosition(
	_ context.Context,
	_ string,
	input CreateLayoutTechnicalPositionInput,
	_ string,
) (*LayoutTechnicalPosition, error) {
	spy.created = input
	return &LayoutTechnicalPosition{
		Label: input.Label, Kind: input.Kind, PositionXMM: input.PositionXMM,
		PositionYMM: input.PositionYMM, RotationDegrees: input.RotationDegrees,
		ProductID: input.ProductID, Description: input.Description,
	}, nil
}

func (spy *layoutPositionRepositorySpy) UpdateTechnicalPosition(
	_ context.Context,
	_ string,
	input UpdateLayoutTechnicalPositionInput,
	_ string,
) (*LayoutTechnicalPosition, error) {
	spy.updated = input
	return &LayoutTechnicalPosition{Label: input.Label, RotationDegrees: input.RotationDegrees}, nil
}

func TestCreateTechnicalPositionNormalizesRotationAndTrimsText(t *testing.T) {
	repository := &layoutPositionRepositorySpy{}
	service := NewLayoutService(repository)

	position, err := service.CreateTechnicalPosition(t.Context(), " unit-1 ", CreateLayoutTechnicalPositionInput{
		Label: "  Signal A  ", Kind: domain.LayoutPositionSignal,
		PositionXMM: 120.5, PositionYMM: -4, RotationDegrees: 450,
		ProductID: " product-1 ", Description: "  Einfahrt  ",
	}, "planner")
	if err != nil {
		t.Fatal(err)
	}
	if position.RotationDegrees != 90 || position.Label != "Signal A" || position.Description != "Einfahrt" {
		t.Fatalf("unexpected normalized position: %#v", position)
	}
	if repository.created.ProductID != "product-1" {
		t.Fatalf("unexpected normalized product ID: %q", repository.created.ProductID)
	}
}

func TestTechnicalPositionServiceRejectsInvalidInput(t *testing.T) {
	service := NewLayoutService(&layoutPositionRepositorySpy{})
	tests := map[string]CreateLayoutTechnicalPositionInput{
		"missing label": {Kind: domain.LayoutPositionSignal},
		"invalid kind":  {Label: "Signal", Kind: domain.LayoutTechnicalPositionKind("command")},
		"nan x":         {Label: "Signal", Kind: domain.LayoutPositionSignal, PositionXMM: math.NaN()},
		"infinite y":    {Label: "Signal", Kind: domain.LayoutPositionSignal, PositionYMM: math.Inf(1)},
	}
	for name, input := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := service.CreateTechnicalPosition(t.Context(), "unit-1", input, "planner")
			if !errors.Is(err, ErrLayoutValidation) {
				t.Fatalf("expected validation error, got %v", err)
			}
		})
	}

	_, err := service.UpdateTechnicalPosition(t.Context(), "position-1", UpdateLayoutTechnicalPositionInput{
		CreateLayoutTechnicalPositionInput: CreateLayoutTechnicalPositionInput{
			Label: "Signal", Kind: domain.LayoutPositionSignal,
		},
	}, "planner")
	if !errors.Is(err, ErrLayoutValidation) {
		t.Fatalf("expected missing version validation error, got %v", err)
	}
}
