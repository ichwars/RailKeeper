package application

import (
	"context"
	"errors"
	"math"
	"strings"

	"railkeeper/backend/internal/domain"
)

var (
	ErrTrackPlanValidation = errors.New("track plan validation failed")
	ErrTrackPlanNotFound   = errors.New("track plan object not found")
	ErrTrackPlanImmutable  = errors.New("track plan revision is immutable")
	ErrTrackPlanConflict   = errors.New("track plan object version conflict")
)

type TrackPlan struct {
	RevisionID string                    `json:"revisionId"`
	Status     domain.PlanRevisionStatus `json:"status"`
	Objects    []domain.PlanTrackObject  `json:"objects"`
}

type CreatePlanTrackObjectInput struct {
	GeometryID      string  `json:"geometryId"`
	PositionXMM     float64 `json:"positionXMm"`
	PositionYMM     float64 `json:"positionYMm"`
	RotationDegrees float64 `json:"rotationDegrees"`
}

type UpdatePlanTrackObjectInput struct {
	PositionXMM     float64 `json:"positionXMm"`
	PositionYMM     float64 `json:"positionYMm"`
	RotationDegrees float64 `json:"rotationDegrees"`
	ExpectedVersion int     `json:"expectedVersion"`
}

type TrackPlannerRepository interface {
	ListGeometries(context.Context, string) ([]domain.TrackGeometryDefinition, error)
	GetPlan(context.Context, string) (*TrackPlan, error)
	CreateObject(
		context.Context, string, CreatePlanTrackObjectInput, string,
	) (*domain.PlanTrackObject, error)
	UpdateObject(
		context.Context, string, UpdatePlanTrackObjectInput, string,
	) (*domain.PlanTrackObject, error)
	DeleteObject(context.Context, string, int, string) error
}

type TrackPlannerService struct {
	repository TrackPlannerRepository
}

func NewTrackPlannerService(repository TrackPlannerRepository) *TrackPlannerService {
	return &TrackPlannerService{repository: repository}
}

func (service *TrackPlannerService) ListGeometries(
	ctx context.Context,
	gauge string,
) ([]domain.TrackGeometryDefinition, error) {
	gauge = strings.TrimSpace(gauge)
	if gauge == "" {
		return nil, ErrTrackPlanValidation
	}
	return service.repository.ListGeometries(ctx, gauge)
}

func (service *TrackPlannerService) GetPlan(ctx context.Context, revisionID string) (*TrackPlan, error) {
	revisionID = strings.TrimSpace(revisionID)
	if revisionID == "" {
		return nil, ErrTrackPlanValidation
	}
	return service.repository.GetPlan(ctx, revisionID)
}

func (service *TrackPlannerService) CreateObject(
	ctx context.Context,
	revisionID string,
	input CreatePlanTrackObjectInput,
	actor string,
) (*domain.PlanTrackObject, error) {
	revisionID = strings.TrimSpace(revisionID)
	input.GeometryID = strings.TrimSpace(input.GeometryID)
	if revisionID == "" || input.GeometryID == "" || !validTrackCoordinates(
		input.PositionXMM, input.PositionYMM, input.RotationDegrees,
	) {
		return nil, ErrTrackPlanValidation
	}
	input.RotationDegrees = domain.NormalizeTrackRotation(input.RotationDegrees)
	return service.repository.CreateObject(ctx, revisionID, input, actor)
}

func (service *TrackPlannerService) UpdateObject(
	ctx context.Context,
	id string,
	input UpdatePlanTrackObjectInput,
	actor string,
) (*domain.PlanTrackObject, error) {
	id = strings.TrimSpace(id)
	if id == "" || input.ExpectedVersion < 1 || !validTrackCoordinates(
		input.PositionXMM, input.PositionYMM, input.RotationDegrees,
	) {
		return nil, ErrTrackPlanValidation
	}
	input.RotationDegrees = domain.NormalizeTrackRotation(input.RotationDegrees)
	return service.repository.UpdateObject(ctx, id, input, actor)
}

func (service *TrackPlannerService) DeleteObject(
	ctx context.Context,
	id string,
	expectedVersion int,
	actor string,
) error {
	id = strings.TrimSpace(id)
	if id == "" || expectedVersion < 1 {
		return ErrTrackPlanValidation
	}
	return service.repository.DeleteObject(ctx, id, expectedVersion, actor)
}

func validTrackCoordinates(values ...float64) bool {
	for _, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return false
		}
	}
	return true
}
