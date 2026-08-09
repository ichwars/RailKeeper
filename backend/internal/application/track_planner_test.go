package application

import (
	"context"
	"errors"
	"math"
	"testing"

	"railkeeper/backend/internal/domain"
)

type trackPlannerRepositorySpy struct {
	listedGauge     string
	createdRevision string
	created         CreatePlanTrackObjectInput
	updated         UpdatePlanTrackObjectInput
	deletedVersion  int
}

func (spy *trackPlannerRepositorySpy) ListGeometries(
	_ context.Context,
	gauge string,
) ([]domain.TrackGeometryDefinition, error) {
	spy.listedGauge = gauge
	return []domain.TrackGeometryDefinition{}, nil
}

func (spy *trackPlannerRepositorySpy) GetPlan(_ context.Context, revisionID string) (*TrackPlan, error) {
	return &TrackPlan{RevisionID: revisionID, Objects: []domain.PlanTrackObject{}}, nil
}

func (spy *trackPlannerRepositorySpy) CreateObject(
	_ context.Context,
	revisionID string,
	input CreatePlanTrackObjectInput,
	_ string,
) (*domain.PlanTrackObject, error) {
	spy.createdRevision = revisionID
	spy.created = input
	return &domain.PlanTrackObject{
		RevisionID: revisionID, GeometryID: input.GeometryID,
		PositionXMM: input.PositionXMM, PositionYMM: input.PositionYMM,
		RotationDegrees: input.RotationDegrees,
	}, nil
}

func (spy *trackPlannerRepositorySpy) UpdateObject(
	_ context.Context,
	_ string,
	input UpdatePlanTrackObjectInput,
	_ string,
) (*domain.PlanTrackObject, error) {
	spy.updated = input
	return &domain.PlanTrackObject{
		PositionXMM: input.PositionXMM, PositionYMM: input.PositionYMM,
		RotationDegrees: input.RotationDegrees,
	}, nil
}

func (spy *trackPlannerRepositorySpy) DeleteObject(
	_ context.Context,
	_ string,
	expectedVersion int,
	_ string,
) error {
	spy.deletedVersion = expectedVersion
	return nil
}

func TestTrackPlannerServiceNormalizesInputs(t *testing.T) {
	repository := &trackPlannerRepositorySpy{}
	service := NewTrackPlannerService(repository)

	if _, err := service.ListGeometries(t.Context(), " TT "); err != nil {
		t.Fatal(err)
	}
	if repository.listedGauge != "TT" {
		t.Fatalf("unexpected normalized gauge: %q", repository.listedGauge)
	}

	created, err := service.CreateObject(t.Context(), " revision-1 ", CreatePlanTrackObjectInput{
		GeometryID: " geometry-1 ", PositionXMM: 20, PositionYMM: 30, RotationDegrees: -15,
	}, "planner")
	if err != nil {
		t.Fatal(err)
	}
	if repository.createdRevision != "revision-1" || repository.created.GeometryID != "geometry-1" ||
		created.RotationDegrees != 345 {
		t.Fatalf("unexpected normalized create: revision=%q input=%#v object=%#v",
			repository.createdRevision, repository.created, created)
	}

	updated, err := service.UpdateObject(t.Context(), " object-1 ", UpdatePlanTrackObjectInput{
		PositionXMM: 25, PositionYMM: 35, RotationDegrees: 735, ExpectedVersion: 2,
	}, "planner")
	if err != nil {
		t.Fatal(err)
	}
	if updated.RotationDegrees != 15 || repository.updated.ExpectedVersion != 2 {
		t.Fatalf("unexpected normalized update: %#v", repository.updated)
	}
}

func TestTrackPlannerServiceRejectsInvalidInputs(t *testing.T) {
	service := NewTrackPlannerService(&trackPlannerRepositorySpy{})

	createInputs := []CreatePlanTrackObjectInput{
		{},
		{GeometryID: "geometry-1", PositionXMM: math.NaN()},
		{GeometryID: "geometry-1", PositionYMM: math.Inf(1)},
		{GeometryID: "geometry-1", RotationDegrees: math.NaN()},
	}
	for _, input := range createInputs {
		if _, err := service.CreateObject(t.Context(), "revision-1", input, "planner"); !errors.Is(err, ErrTrackPlanValidation) {
			t.Fatalf("expected create validation error for %#v, got %v", input, err)
		}
	}

	if _, err := service.GetPlan(t.Context(), " "); !errors.Is(err, ErrTrackPlanValidation) {
		t.Fatalf("expected plan validation error, got %v", err)
	}
	if _, err := service.UpdateObject(t.Context(), "object-1", UpdatePlanTrackObjectInput{}, "planner"); !errors.Is(err, ErrTrackPlanValidation) {
		t.Fatalf("expected update validation error, got %v", err)
	}
	if err := service.DeleteObject(t.Context(), "object-1", 0, "planner"); !errors.Is(err, ErrTrackPlanValidation) {
		t.Fatalf("expected delete validation error, got %v", err)
	}
}
