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

type TrackMaterialStatus struct {
	GeometryID        string   `json:"geometryId"`
	Manufacturer      string   `json:"manufacturer"`
	ArticleNumber     string   `json:"articleNumber"`
	Name              string   `json:"name"`
	RequiredQuantity  int      `json:"requiredQuantity"`
	ProductIDs        []string `json:"productIds"`
	InventoryNumbers  []string `json:"inventoryNumbers"`
	PhysicalQuantity  int      `json:"physicalQuantity"`
	ReservedQuantity  int      `json:"reservedQuantity"`
	AvailableQuantity int      `json:"availableQuantity"`
	MissingQuantity   int      `json:"missingQuantity"`
}

type TrackPlanAnalysis struct {
	RevisionID  string                       `json:"revisionId"`
	Status      domain.PlanRevisionStatus    `json:"status"`
	Connections []domain.TrackPlanConnection `json:"connections"`
	Issues      []domain.TrackPlanIssue      `json:"issues"`
	BOM         []domain.TrackBOMLine        `json:"bom"`
	Materials   []TrackMaterialStatus        `json:"materials"`
}

type TrackPlannerRepository interface {
	ListGeometries(context.Context, string) ([]domain.TrackGeometryDefinition, error)
	GetPlan(context.Context, string) (*TrackPlan, error)
	GetPlanForObject(context.Context, string) (*TrackPlan, error)
	TrackMaterialAvailability(context.Context, []domain.TrackBOMLine) ([]TrackMaterialStatus, error)
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

func (service *TrackPlannerService) AnalyzePlan(
	ctx context.Context,
	revisionID string,
) (*TrackPlanAnalysis, error) {
	revisionID = strings.TrimSpace(revisionID)
	if revisionID == "" {
		return nil, ErrTrackPlanValidation
	}
	plan, err := service.repository.GetPlan(ctx, revisionID)
	if err != nil {
		return nil, err
	}
	geometryAnalysis := domain.AnalyzeTrackPlan(plan.Objects)
	materials, err := service.repository.TrackMaterialAvailability(ctx, geometryAnalysis.BOM)
	if err != nil {
		return nil, err
	}
	return &TrackPlanAnalysis{
		RevisionID: plan.RevisionID, Status: plan.Status,
		Connections: geometryAnalysis.Connections, Issues: geometryAnalysis.Issues,
		BOM: geometryAnalysis.BOM, Materials: materials,
	}, nil
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
	plan, err := service.repository.GetPlanForObject(ctx, id)
	if err != nil {
		return nil, err
	}
	var moving *domain.PlanTrackObject
	for index := range plan.Objects {
		if plan.Objects[index].ID == id {
			moving = &plan.Objects[index]
			break
		}
	}
	if moving == nil {
		return nil, ErrTrackPlanNotFound
	}
	moving.PositionXMM = input.PositionXMM
	moving.PositionYMM = input.PositionYMM
	moving.RotationDegrees = input.RotationDegrees
	if snap := domain.FindTrackSnap(*moving, plan.Objects); snap.Snapped {
		input.PositionXMM = snap.Pose.PositionXMM
		input.PositionYMM = snap.Pose.PositionYMM
		input.RotationDegrees = snap.Pose.RotationDegrees
	}
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
