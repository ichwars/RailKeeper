package application

import (
	"context"
	"errors"
	"strings"

	"railkeeper/backend/internal/domain"
)

var (
	ErrLayoutPositionNotFound        = errors.New("layout technical position not found")
	ErrLayoutPositionVersionConflict = errors.New("layout technical position version conflict")
	ErrLayoutPositionProductNotFound = errors.New("layout technical position product not found")
)

type LayoutTechnicalPosition struct {
	ID              string                             `json:"id"`
	LayoutUnitID    string                             `json:"layoutUnitId"`
	Label           string                             `json:"label"`
	Kind            domain.LayoutTechnicalPositionKind `json:"kind"`
	PositionXMM     float64                            `json:"positionXMm"`
	PositionYMM     float64                            `json:"positionYMm"`
	RotationDegrees float64                            `json:"rotationDegrees"`
	ProductID       string                             `json:"productId,omitempty"`
	Description     string                             `json:"description,omitempty"`
	Version         int                                `json:"version"`
	Archived        bool                               `json:"archived"`
	CreatedAt       string                             `json:"createdAt"`
	UpdatedAt       string                             `json:"updatedAt"`
}

type CreateLayoutTechnicalPositionInput struct {
	Label           string                             `json:"label"`
	Kind            domain.LayoutTechnicalPositionKind `json:"kind"`
	PositionXMM     float64                            `json:"positionXMm"`
	PositionYMM     float64                            `json:"positionYMm"`
	RotationDegrees float64                            `json:"rotationDegrees"`
	ProductID       string                             `json:"productId"`
	Description     string                             `json:"description"`
	Archived        bool                               `json:"archived"`
}

type UpdateLayoutTechnicalPositionInput struct {
	CreateLayoutTechnicalPositionInput
	ExpectedVersion int `json:"expectedVersion"`
}

func (s *LayoutService) ListTechnicalPositions(
	ctx context.Context,
	layoutUnitID string,
) ([]LayoutTechnicalPosition, error) {
	layoutUnitID = strings.TrimSpace(layoutUnitID)
	if layoutUnitID == "" {
		return nil, ErrLayoutValidation
	}
	return s.repository.ListTechnicalPositions(ctx, layoutUnitID)
}

func (s *LayoutService) CreateTechnicalPosition(
	ctx context.Context,
	layoutUnitID string,
	input CreateLayoutTechnicalPositionInput,
	actor string,
) (*LayoutTechnicalPosition, error) {
	layoutUnitID = strings.TrimSpace(layoutUnitID)
	input = cleanTechnicalPositionInput(input)
	if layoutUnitID == "" || !validTechnicalPositionInput(input) {
		return nil, ErrLayoutValidation
	}
	return s.repository.CreateTechnicalPosition(ctx, layoutUnitID, input, actor)
}

func (s *LayoutService) UpdateTechnicalPosition(
	ctx context.Context,
	id string,
	input UpdateLayoutTechnicalPositionInput,
	actor string,
) (*LayoutTechnicalPosition, error) {
	id = strings.TrimSpace(id)
	input.CreateLayoutTechnicalPositionInput = cleanTechnicalPositionInput(
		input.CreateLayoutTechnicalPositionInput,
	)
	if id == "" || input.ExpectedVersion < 1 || !validTechnicalPositionInput(input.CreateLayoutTechnicalPositionInput) {
		return nil, ErrLayoutValidation
	}
	return s.repository.UpdateTechnicalPosition(ctx, id, input, actor)
}

func cleanTechnicalPositionInput(input CreateLayoutTechnicalPositionInput) CreateLayoutTechnicalPositionInput {
	input.Label = strings.TrimSpace(input.Label)
	input.ProductID = strings.TrimSpace(input.ProductID)
	input.Description = strings.TrimSpace(input.Description)
	input.RotationDegrees = normalizeRotation(input.RotationDegrees)
	return input
}

func validTechnicalPositionInput(input CreateLayoutTechnicalPositionInput) bool {
	return input.Label != "" && input.Kind.Valid() && finite(input.PositionXMM) && finite(input.PositionYMM) &&
		finite(input.RotationDegrees)
}
