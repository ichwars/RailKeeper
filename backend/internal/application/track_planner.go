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
	RevisionID   string                       `json:"revisionId"`
	Status       domain.PlanRevisionStatus    `json:"status"`
	Connections  []domain.TrackPlanConnection `json:"connections"`
	Issues       []domain.TrackPlanIssue      `json:"issues"`
	BOM          []domain.TrackBOMLine        `json:"bom"`
	Materials    []TrackMaterialStatus        `json:"materials"`
	Reservations []TrackPlanObjectReservation `json:"reservations"`
}

type TrackPlanAffectedConfiguration struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type TrackPlanChangePreview struct {
	RevisionID             string                           `json:"revisionId"`
	BaseRevisionID         string                           `json:"baseRevisionId"`
	ObjectChanges          []domain.TrackPlanObjectChange   `json:"objectChanges"`
	MaterialDeltas         []domain.TrackPlanMaterialDelta  `json:"materialDeltas"`
	Issues                 domain.TrackPlanIssueDiff        `json:"issues"`
	AffectedConfigurations []TrackPlanAffectedConfiguration `json:"affectedConfigurations"`
}

type TrackPlanReservationInput struct {
	TrackObjectID         string `json:"trackObjectId"`
	ProductID             string `json:"productId"`
	LocationID            string `json:"locationId"`
	AssetID               string `json:"assetId,omitempty"`
	ExpectedObjectVersion int    `json:"expectedObjectVersion"`
}

type ReserveTrackPlanMaterialsInput struct {
	Confirmed bool                        `json:"confirmed"`
	Items     []TrackPlanReservationInput `json:"items"`
}

type TrackPlanObjectReservation struct {
	TrackObjectID string               `json:"trackObjectId"`
	Reservation   AccessoryReservation `json:"reservation"`
}

type TrackPlanReservationBatch struct {
	RevisionID   string                       `json:"revisionId"`
	Reservations []TrackPlanObjectReservation `json:"reservations"`
	Materials    []TrackMaterialStatus        `json:"materials"`
}

type TrackPlannerRepository interface {
	ListGeometries(context.Context, string) ([]domain.TrackGeometryDefinition, error)
	GetPlan(context.Context, string) (*TrackPlan, error)
	GetBaseRevisionID(context.Context, string) (string, error)
	ListAffectedConfigurations(context.Context, string) ([]TrackPlanAffectedConfiguration, error)
	ReserveMaterials(
		context.Context, string, ReserveTrackPlanMaterialsInput, string,
	) (*TrackPlanReservationBatch, error)
	GetPlanForObject(context.Context, string) (*TrackPlan, error)
	TrackMaterialAvailability(context.Context, []domain.TrackBOMLine) ([]TrackMaterialStatus, error)
	ListMaterialReservations(context.Context, string) ([]TrackPlanObjectReservation, error)
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
	reservations, err := service.repository.ListMaterialReservations(ctx, revisionID)
	if err != nil {
		return nil, err
	}
	return &TrackPlanAnalysis{
		RevisionID: plan.RevisionID, Status: plan.Status,
		Connections: geometryAnalysis.Connections, Issues: geometryAnalysis.Issues,
		BOM: geometryAnalysis.BOM, Materials: materials, Reservations: reservations,
	}, nil
}

func (service *TrackPlannerService) ChangePreview(
	ctx context.Context,
	revisionID string,
) (*TrackPlanChangePreview, error) {
	revisionID = strings.TrimSpace(revisionID)
	if revisionID == "" {
		return nil, ErrTrackPlanValidation
	}
	current, err := service.repository.GetPlan(ctx, revisionID)
	if err != nil {
		return nil, err
	}
	baseRevisionID, err := service.repository.GetBaseRevisionID(ctx, revisionID)
	if err != nil {
		return nil, err
	}
	base := &TrackPlan{Objects: []domain.PlanTrackObject{}}
	affected := []TrackPlanAffectedConfiguration{}
	if baseRevisionID != "" {
		base, err = service.repository.GetPlan(ctx, baseRevisionID)
		if err != nil {
			return nil, err
		}
		affected, err = service.repository.ListAffectedConfigurations(ctx, baseRevisionID)
		if err != nil {
			return nil, err
		}
	}
	revisionDiff := domain.CompareTrackPlanRevisions(base.Objects, current.Objects)
	baseAnalysis := domain.AnalyzeTrackPlan(base.Objects)
	currentAnalysis := domain.AnalyzeTrackPlan(current.Objects)
	return &TrackPlanChangePreview{
		RevisionID: current.RevisionID, BaseRevisionID: baseRevisionID,
		ObjectChanges: revisionDiff.ObjectChanges, MaterialDeltas: revisionDiff.MaterialDeltas,
		Issues: domain.DiffTrackPlanIssues(
			baseAnalysis.Issues, currentAnalysis.Issues, base.Objects, current.Objects,
		),
		AffectedConfigurations: affected,
	}, nil
}

func (service *TrackPlannerService) ReserveMaterials(
	ctx context.Context,
	revisionID string,
	input ReserveTrackPlanMaterialsInput,
	actor string,
) (*TrackPlanReservationBatch, error) {
	revisionID = strings.TrimSpace(revisionID)
	if revisionID == "" || !input.Confirmed || len(input.Items) == 0 {
		return nil, ErrTrackPlanValidation
	}
	seenObjects := make(map[string]struct{}, len(input.Items))
	for index := range input.Items {
		item := &input.Items[index]
		item.TrackObjectID = strings.TrimSpace(item.TrackObjectID)
		item.ProductID = strings.TrimSpace(item.ProductID)
		item.LocationID = strings.TrimSpace(item.LocationID)
		item.AssetID = strings.TrimSpace(item.AssetID)
		if item.TrackObjectID == "" || item.ProductID == "" || item.LocationID == "" ||
			item.ExpectedObjectVersion < 1 {
			return nil, ErrTrackPlanValidation
		}
		if _, duplicate := seenObjects[item.TrackObjectID]; duplicate {
			return nil, ErrTrackPlanValidation
		}
		seenObjects[item.TrackObjectID] = struct{}{}
	}
	batch, err := service.repository.ReserveMaterials(ctx, revisionID, input, actor)
	if err != nil {
		return nil, err
	}
	analysis, err := service.AnalyzePlan(ctx, revisionID)
	if err != nil {
		return nil, err
	}
	batch.Materials = analysis.Materials
	return batch, nil
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
