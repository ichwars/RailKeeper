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
	plan            *TrackPlan
	planForObject   *TrackPlan
	materials       []TrackMaterialStatus
}

func (spy *trackPlannerRepositorySpy) ListGeometries(
	_ context.Context,
	gauge string,
) ([]domain.TrackGeometryDefinition, error) {
	spy.listedGauge = gauge
	return []domain.TrackGeometryDefinition{}, nil
}

func (spy *trackPlannerRepositorySpy) GetPlan(_ context.Context, revisionID string) (*TrackPlan, error) {
	if spy.plan != nil {
		return spy.plan, nil
	}
	return &TrackPlan{RevisionID: revisionID, Objects: []domain.PlanTrackObject{}}, nil
}

func (spy *trackPlannerRepositorySpy) GetPlanForObject(_ context.Context, _ string) (*TrackPlan, error) {
	if spy.planForObject == nil {
		return nil, ErrTrackPlanNotFound
	}
	return spy.planForObject, nil
}

func (spy *trackPlannerRepositorySpy) TrackMaterialAvailability(
	_ context.Context,
	_ []domain.TrackBOMLine,
) ([]TrackMaterialStatus, error) {
	return spy.materials, nil
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
	repository.planForObject = &TrackPlan{RevisionID: "revision-1", Status: domain.PlanRevisionDraft,
		Objects: []domain.PlanTrackObject{trackPlannerTestG1("object-1", 25, 35, 15)}}

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

func TestTrackPlannerSnapAppliesNearestCompatiblePose(t *testing.T) {
	moving := trackPlannerTestG1("moving", 172, 2, 2)
	target := trackPlannerTestG1("target", 0, 0, 0)
	repository := &trackPlannerRepositorySpy{planForObject: &TrackPlan{
		RevisionID: "revision-1", Status: domain.PlanRevisionDraft,
		Objects: []domain.PlanTrackObject{target, moving},
	}}
	service := NewTrackPlannerService(repository)

	updated, err := service.UpdateObject(t.Context(), moving.ID, UpdatePlanTrackObjectInput{
		PositionXMM: moving.PositionXMM, PositionYMM: moving.PositionYMM,
		RotationDegrees: moving.RotationDegrees, ExpectedVersion: 1,
	}, "planner")
	if err != nil {
		t.Fatal(err)
	}
	if updated.PositionXMM != 166 || updated.PositionYMM != 0 || updated.RotationDegrees != 0 {
		t.Fatalf("expected authoritative snapped pose, got %#v", updated)
	}
}

func TestTrackPlannerSnapLeavesPoseOutsideTolerance(t *testing.T) {
	moving := trackPlannerTestG1("moving", 174.01, 0, 0)
	target := trackPlannerTestG1("target", 0, 0, 0)
	repository := &trackPlannerRepositorySpy{planForObject: &TrackPlan{
		RevisionID: "revision-1", Status: domain.PlanRevisionDraft,
		Objects: []domain.PlanTrackObject{target, moving},
	}}
	service := NewTrackPlannerService(repository)

	updated, err := service.UpdateObject(t.Context(), moving.ID, UpdatePlanTrackObjectInput{
		PositionXMM: moving.PositionXMM, PositionYMM: moving.PositionYMM, ExpectedVersion: 1,
	}, "planner")
	if err != nil {
		t.Fatal(err)
	}
	if updated.PositionXMM != 174.01 {
		t.Fatalf("pose outside tolerance changed: %#v", updated)
	}
}

func TestTrackPlannerAnalyzePlanCombinesGeometryAndMaterials(t *testing.T) {
	objects := []domain.PlanTrackObject{
		trackPlannerTestG1("track-1", 0, 0, 0),
		trackPlannerTestG1("track-2", 166, 0, 0),
	}
	repository := &trackPlannerRepositorySpy{
		materials: []TrackMaterialStatus{{
			GeometryID: "tillig-g1", Manufacturer: "Tillig", ArticleNumber: "83101",
			RequiredQuantity: 2, PhysicalQuantity: 3, ReservedQuantity: 1,
			AvailableQuantity: 2, MissingQuantity: 0,
		}},
	}
	repositoryPlan := &TrackPlan{RevisionID: "revision-1", Status: domain.PlanRevisionDraft, Objects: objects}
	repository.plan = repositoryPlan
	service := NewTrackPlannerService(repository)
	analysis, err := service.AnalyzePlan(t.Context(), repositoryPlan.RevisionID)
	if err != nil {
		t.Fatal(err)
	}
	if len(analysis.Connections) != 1 || len(analysis.BOM) != 1 || len(analysis.Materials) != 1 ||
		analysis.Materials[0].AvailableQuantity != 2 {
		t.Fatalf("unexpected combined analysis: %#v", analysis)
	}
}

func trackPlannerTestG1(id string, x, y, rotation float64) domain.PlanTrackObject {
	return domain.PlanTrackObject{
		ID: id, GeometryID: "tillig-g1", PositionXMM: x, PositionYMM: y, RotationDegrees: rotation,
		Version: 1,
		Geometry: domain.TrackGeometryDefinition{
			ID: "tillig-g1", LibraryID: "tillig-v1", ArticleNumber: "83101", Name: "Gleisstück G1",
			Kind: domain.TrackGeometryStraight, LengthMM: 166, Status: domain.TrackGeometryVerified,
			Geometry: domain.TrackGeometry{SchemaVersion: 1,
				Ports: []domain.TrackPort{
					{ID: "a", DirectionDegrees: 180}, {ID: "b", XMM: 166, DirectionDegrees: 0},
				},
				Routes: []domain.TrackRoute{{ID: "main", Points: []domain.TrackPoint{{}, {XMM: 166}}}},
			},
		},
	}
}
