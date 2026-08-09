package application

import (
	"context"
	"errors"
	"math"
	"strings"

	"railkeeper/backend/internal/domain"
)

var (
	ErrLayoutValidation      = errors.New("layout validation failed")
	ErrLayoutNotFound        = errors.New("layout not found")
	ErrLayoutVersionConflict = errors.New("layout version conflict")
	ErrPlanRevisionImmutable = errors.New("plan revision is immutable")
	ErrPlanRevisionConflict  = errors.New("plan revision conflict")
)

type Layout struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Kind        domain.LayoutKind `json:"kind"`
	Gauge       string            `json:"gauge"`
	Scale       string            `json:"scale"`
	Description string            `json:"description,omitempty"`
	Version     int               `json:"version"`
	Archived    bool              `json:"archived"`
	CreatedAt   string            `json:"createdAt"`
	UpdatedAt   string            `json:"updatedAt"`
}

type CreateLayoutInput struct {
	Name        string            `json:"name"`
	Kind        domain.LayoutKind `json:"kind"`
	Gauge       string            `json:"gauge"`
	Scale       string            `json:"scale"`
	Description string            `json:"description"`
	Archived    bool              `json:"archived"`
}

type UpdateLayoutInput struct {
	CreateLayoutInput
	ExpectedVersion int `json:"expectedVersion"`
}

type LayoutUnit struct {
	ID         string                `json:"id"`
	LayoutID   string                `json:"layoutId"`
	Name       string                `json:"name"`
	Kind       domain.LayoutUnitKind `json:"kind"`
	OwnerLabel string                `json:"ownerLabel,omitempty"`
	WidthMM    float64               `json:"widthMm"`
	HeightMM   float64               `json:"heightMm"`
	Version    int                   `json:"version"`
	Archived   bool                  `json:"archived"`
	CreatedAt  string                `json:"createdAt"`
	UpdatedAt  string                `json:"updatedAt"`
}

type CreateLayoutUnitInput struct {
	Name       string                `json:"name"`
	Kind       domain.LayoutUnitKind `json:"kind"`
	OwnerLabel string                `json:"ownerLabel"`
	WidthMM    float64               `json:"widthMm"`
	HeightMM   float64               `json:"heightMm"`
	Archived   bool                  `json:"archived"`
}

type UpdateLayoutUnitInput struct {
	CreateLayoutUnitInput
	ExpectedVersion int `json:"expectedVersion"`
}

type LayoutConfiguration struct {
	ID          string              `json:"id"`
	LayoutID    string              `json:"layoutId"`
	Name        string              `json:"name"`
	Description string              `json:"description,omitempty"`
	Version     int                 `json:"version"`
	Archived    bool                `json:"archived"`
	Units       []ConfigurationUnit `json:"units"`
	CreatedAt   string              `json:"createdAt"`
	UpdatedAt   string              `json:"updatedAt"`
}

type ConfigurationUnit struct {
	UnitID          string  `json:"unitId"`
	PlanRevisionID  string  `json:"planRevisionId,omitempty"`
	PositionXMM     float64 `json:"positionXMm"`
	PositionYMM     float64 `json:"positionYMm"`
	RotationDegrees float64 `json:"rotationDegrees"`
	SortOrder       int     `json:"sortOrder"`
}

type ConfigurationUnitInput = ConfigurationUnit

type SaveLayoutConfigurationInput struct {
	ID              string                   `json:"id"`
	Name            string                   `json:"name"`
	Description     string                   `json:"description"`
	ExpectedVersion int                      `json:"expectedVersion"`
	Archived        bool                     `json:"archived"`
	Units           []ConfigurationUnitInput `json:"units"`
}

type PlanVariant struct {
	ID           string         `json:"id"`
	LayoutUnitID string         `json:"layoutUnitId"`
	Name         string         `json:"name"`
	Description  string         `json:"description,omitempty"`
	Archived     bool           `json:"archived"`
	Revisions    []PlanRevision `json:"revisions"`
	CreatedAt    string         `json:"createdAt"`
	UpdatedAt    string         `json:"updatedAt"`
}

type CreatePlanVariantInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type PlanRevision struct {
	ID             string                    `json:"id"`
	VariantID      string                    `json:"variantId"`
	RevisionNumber int                       `json:"revisionNumber"`
	Status         domain.PlanRevisionStatus `json:"status"`
	BaseRevisionID string                    `json:"baseRevisionId,omitempty"`
	Version        int                       `json:"version"`
	CreatedBy      string                    `json:"createdBy"`
	PublishedBy    string                    `json:"publishedBy,omitempty"`
	PublishedAt    string                    `json:"publishedAt,omitempty"`
	CreatedAt      string                    `json:"createdAt"`
	UpdatedAt      string                    `json:"updatedAt"`
}

type CreatePlanRevisionInput struct {
	BaseRevisionID string `json:"baseRevisionId"`
}

type LayoutRepository interface {
	ListLayouts(context.Context) ([]Layout, error)
	GetLayout(context.Context, string) (*Layout, error)
	CreateLayout(context.Context, CreateLayoutInput, string) (*Layout, error)
	UpdateLayout(context.Context, string, UpdateLayoutInput, string) (*Layout, error)
	ListUnits(context.Context, string) ([]LayoutUnit, error)
	CreateUnit(context.Context, string, CreateLayoutUnitInput, string) (*LayoutUnit, error)
	UpdateUnit(context.Context, string, UpdateLayoutUnitInput, string) (*LayoutUnit, error)
	ListTechnicalPositions(context.Context, string) ([]LayoutTechnicalPosition, error)
	CreateTechnicalPosition(
		context.Context, string, CreateLayoutTechnicalPositionInput, string,
	) (*LayoutTechnicalPosition, error)
	UpdateTechnicalPosition(
		context.Context, string, UpdateLayoutTechnicalPositionInput, string,
	) (*LayoutTechnicalPosition, error)
	ListConfigurations(context.Context, string) ([]LayoutConfiguration, error)
	SaveConfiguration(context.Context, string, SaveLayoutConfigurationInput, string) (*LayoutConfiguration, error)
	ListVariants(context.Context, string) ([]PlanVariant, error)
	CreateVariant(context.Context, string, CreatePlanVariantInput, string) (*PlanVariant, error)
	CreateDraft(context.Context, string, CreatePlanRevisionInput, string) (*PlanRevision, error)
	SubmitRevision(context.Context, string, int, string) (*PlanRevision, error)
	PublishRevision(context.Context, string, int, string) (*PlanRevision, error)
}

type LayoutService struct {
	repository LayoutRepository
}

func NewLayoutService(repository LayoutRepository) *LayoutService {
	return &LayoutService{repository: repository}
}

func (s *LayoutService) ListLayouts(ctx context.Context) ([]Layout, error) {
	return s.repository.ListLayouts(ctx)
}

func (s *LayoutService) GetLayout(ctx context.Context, id string) (*Layout, error) {
	return s.repository.GetLayout(ctx, strings.TrimSpace(id))
}

func (s *LayoutService) CreateLayout(ctx context.Context, input CreateLayoutInput, actor string) (*Layout, error) {
	input = cleanLayoutInput(input)
	if !validLayoutInput(input) {
		return nil, ErrLayoutValidation
	}
	return s.repository.CreateLayout(ctx, input, actor)
}

func (s *LayoutService) UpdateLayout(ctx context.Context, id string, input UpdateLayoutInput, actor string) (*Layout, error) {
	input.CreateLayoutInput = cleanLayoutInput(input.CreateLayoutInput)
	if strings.TrimSpace(id) == "" || input.ExpectedVersion < 1 || !validLayoutInput(input.CreateLayoutInput) {
		return nil, ErrLayoutValidation
	}
	return s.repository.UpdateLayout(ctx, strings.TrimSpace(id), input, actor)
}

func (s *LayoutService) ListUnits(ctx context.Context, layoutID string) ([]LayoutUnit, error) {
	return s.repository.ListUnits(ctx, strings.TrimSpace(layoutID))
}

func (s *LayoutService) CreateUnit(ctx context.Context, layoutID string, input CreateLayoutUnitInput, actor string) (*LayoutUnit, error) {
	input = cleanLayoutUnitInput(input)
	if strings.TrimSpace(layoutID) == "" || !validLayoutUnitInput(input) {
		return nil, ErrLayoutValidation
	}
	return s.repository.CreateUnit(ctx, strings.TrimSpace(layoutID), input, actor)
}

func (s *LayoutService) UpdateUnit(ctx context.Context, id string, input UpdateLayoutUnitInput, actor string) (*LayoutUnit, error) {
	input.CreateLayoutUnitInput = cleanLayoutUnitInput(input.CreateLayoutUnitInput)
	if strings.TrimSpace(id) == "" || input.ExpectedVersion < 1 || !validLayoutUnitInput(input.CreateLayoutUnitInput) {
		return nil, ErrLayoutValidation
	}
	return s.repository.UpdateUnit(ctx, strings.TrimSpace(id), input, actor)
}

func (s *LayoutService) ListConfigurations(ctx context.Context, layoutID string) ([]LayoutConfiguration, error) {
	return s.repository.ListConfigurations(ctx, strings.TrimSpace(layoutID))
}

func (s *LayoutService) SaveConfiguration(ctx context.Context, layoutID string, input SaveLayoutConfigurationInput, actor string) (*LayoutConfiguration, error) {
	input.ID = strings.TrimSpace(input.ID)
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	seen := map[string]struct{}{}
	for index := range input.Units {
		unit := &input.Units[index]
		unit.UnitID = strings.TrimSpace(unit.UnitID)
		unit.PlanRevisionID = strings.TrimSpace(unit.PlanRevisionID)
		if !finite(unit.PositionXMM) || !finite(unit.PositionYMM) || !finite(unit.RotationDegrees) {
			return nil, ErrLayoutValidation
		}
		unit.RotationDegrees = normalizeRotation(unit.RotationDegrees)
		unit.SortOrder = index
		if unit.UnitID == "" {
			return nil, ErrLayoutValidation
		}
		if _, duplicate := seen[unit.UnitID]; duplicate {
			return nil, ErrLayoutValidation
		}
		seen[unit.UnitID] = struct{}{}
	}
	layoutID = strings.TrimSpace(layoutID)
	if (layoutID == "" && input.ID == "") || input.Name == "" || (input.ID != "" && input.ExpectedVersion < 1) {
		return nil, ErrLayoutValidation
	}
	return s.repository.SaveConfiguration(ctx, layoutID, input, actor)
}

func cleanLayoutInput(input CreateLayoutInput) CreateLayoutInput {
	input.Name = strings.TrimSpace(input.Name)
	input.Gauge = strings.TrimSpace(input.Gauge)
	input.Scale = strings.TrimSpace(input.Scale)
	input.Description = strings.TrimSpace(input.Description)
	return input
}

func validLayoutInput(input CreateLayoutInput) bool {
	return input.Name != "" && input.Gauge != "" && input.Scale != "" && input.Kind.Valid()
}

func cleanLayoutUnitInput(input CreateLayoutUnitInput) CreateLayoutUnitInput {
	input.Name = strings.TrimSpace(input.Name)
	input.OwnerLabel = strings.TrimSpace(input.OwnerLabel)
	return input
}

func validLayoutUnitInput(input CreateLayoutUnitInput) bool {
	return input.Name != "" && input.Kind.Valid() && finite(input.WidthMM) && finite(input.HeightMM) &&
		input.WidthMM >= 0 && input.HeightMM >= 0
}

func normalizeRotation(value float64) float64 {
	value = math.Mod(value, 360)
	if value < 0 {
		value += 360
	}
	return value
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
