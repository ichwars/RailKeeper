package application

import (
	"context"
	"strings"
	"unicode/utf8"

	"railkeeper/backend/internal/domain"
)

func (service *TrackPlannerService) CreateFreeObject(
	ctx context.Context,
	revisionID string,
	input CreateFreePlanObjectInput,
	actor string,
) (*domain.PlanFreeObject, error) {
	revisionID = strings.TrimSpace(revisionID)
	input = normalizeFreePlanObjectInput(input)
	if revisionID == "" || !validFreePlanObjectInput(input) {
		return nil, ErrTrackPlanValidation
	}
	return service.repository.CreateFreeObject(ctx, revisionID, input, actor)
}

func (service *TrackPlannerService) UpdateFreeObject(
	ctx context.Context,
	id string,
	input UpdateFreePlanObjectInput,
	actor string,
) (*domain.PlanFreeObject, error) {
	id = strings.TrimSpace(id)
	input.CreateFreePlanObjectInput = normalizeFreePlanObjectInput(input.CreateFreePlanObjectInput)
	if id == "" || input.ExpectedVersion < 1 || !validFreePlanObjectInput(input.CreateFreePlanObjectInput) {
		return nil, ErrTrackPlanValidation
	}
	return service.repository.UpdateFreeObject(ctx, id, input, actor)
}

func (service *TrackPlannerService) DeleteFreeObject(
	ctx context.Context,
	id string,
	expectedVersion int,
	actor string,
) error {
	id = strings.TrimSpace(id)
	if id == "" || expectedVersion < 1 {
		return ErrTrackPlanValidation
	}
	return service.repository.DeleteFreeObject(ctx, id, expectedVersion, actor)
}

func normalizeFreePlanObjectInput(input CreateFreePlanObjectInput) CreateFreePlanObjectInput {
	input.Name = strings.TrimSpace(input.Name)
	input.Shape.Text = strings.TrimSpace(input.Shape.Text)
	input.RotationDegrees = domain.NormalizeTrackRotation(input.RotationDegrees)
	return input
}

func validFreePlanObjectInput(input CreateFreePlanObjectInput) bool {
	nameLength := utf8.RuneCountInString(input.Name)
	return nameLength >= 1 && nameLength <= 80 && input.Category.Valid() &&
		validTrackCoordinates(input.PositionXMM, input.PositionYMM, input.RotationDegrees) &&
		domain.ValidateFreePlanObjectShape(input.Shape) == nil
}
