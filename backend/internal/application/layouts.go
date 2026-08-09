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

type LayoutUnitPort struct {
	ID               string                    `json:"id"`
	LayoutUnitID     string                    `json:"layoutUnitId"`
	Name             string                    `json:"name"`
	Kind             domain.LayoutUnitPortKind `json:"kind"`
	InterfaceKey     string                    `json:"interfaceKey"`
	XMM              float64                   `json:"xMm"`
	YMM              float64                   `json:"yMm"`
	DirectionDegrees float64                   `json:"directionDegrees"`
	Notes            string                    `json:"notes,omitempty"`
	Version          int                       `json:"version"`
	Archived         bool                      `json:"archived"`
	CreatedAt        string                    `json:"createdAt"`
	UpdatedAt        string                    `json:"updatedAt"`
}

type CreateLayoutUnitPortInput struct {
	Name             string                    `json:"name"`
	Kind             domain.LayoutUnitPortKind `json:"kind"`
	InterfaceKey     string                    `json:"interfaceKey"`
	XMM              float64                   `json:"xMm"`
	YMM              float64                   `json:"yMm"`
	DirectionDegrees float64                   `json:"directionDegrees"`
	Notes            string                    `json:"notes"`
	Archived         bool                      `json:"archived"`
}

type UpdateLayoutUnitPortInput struct {
	CreateLayoutUnitPortInput
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
	UpdateUnitOutline(context.Context, string, UpdateLayoutUnitOutlineInput, string) (*LayoutUnitOutline, error)
	ListTechnicalPositions(context.Context, string) ([]LayoutTechnicalPosition, error)
	CreateTechnicalPosition(
		context.Context, string, CreateLayoutTechnicalPositionInput, string,
	) (*LayoutTechnicalPosition, error)
	UpdateTechnicalPosition(
		context.Context, string, UpdateLayoutTechnicalPositionInput, string,
	) (*LayoutTechnicalPosition, error)
	GetTwin(context.Context, string, LayoutTwinSelection) (*LayoutTwin, error)
	ListConfigurations(context.Context, string) ([]LayoutConfiguration, error)
	SaveConfiguration(context.Context, string, SaveLayoutConfigurationInput, string) (*LayoutConfiguration, error)
	ListVariants(context.Context, string) ([]PlanVariant, error)
	CreateVariant(context.Context, string, CreatePlanVariantInput, string) (*PlanVariant, error)
	CreateDraft(context.Context, string, CreatePlanRevisionInput, string) (*PlanRevision, error)
	SubmitRevision(context.Context, string, int, string) (*PlanRevision, error)
	PublishRevision(context.Context, string, int, string) (*PlanRevision, error)
}

type LayoutUnitPortRepository interface {
	GetUnit(context.Context, string) (*LayoutUnit, error)
	ListUnitPorts(context.Context, string) ([]LayoutUnitPort, error)
	GetUnitForPort(context.Context, string) (*LayoutUnit, error)
	CreateUnitPort(context.Context, string, CreateLayoutUnitPortInput, string) (*LayoutUnitPort, error)
	UpdateUnitPort(context.Context, string, UpdateLayoutUnitPortInput, string) (*LayoutUnitPort, error)
}

type LayoutConfigurationPortRepository interface {
	LoadConfigurationPortPlacements(context.Context, string) ([]domain.ModulePortPlacement, error)
}

type PreviewConfigurationUnitSnapInput struct {
	UnitID          string  `json:"unitId"`
	PositionXMM     float64 `json:"positionXMm"`
	PositionYMM     float64 `json:"positionYMm"`
	RotationDegrees float64 `json:"rotationDegrees"`
}

type LayoutService struct {
	repository                  LayoutRepository
	portRepository              LayoutUnitPortRepository
	configurationPortRepository LayoutConfigurationPortRepository
}

func NewLayoutService(repository LayoutRepository) *LayoutService {
	portRepository, _ := repository.(LayoutUnitPortRepository)
	configurationPortRepository, _ := repository.(LayoutConfigurationPortRepository)
	return &LayoutService{repository: repository, portRepository: portRepository,
		configurationPortRepository: configurationPortRepository}
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

func (s *LayoutService) ListUnitPorts(ctx context.Context, unitID string) ([]LayoutUnitPort, error) {
	unitID = strings.TrimSpace(unitID)
	if unitID == "" {
		return nil, ErrLayoutValidation
	}
	return s.portRepository.ListUnitPorts(ctx, unitID)
}

func (s *LayoutService) CreateUnitPort(
	ctx context.Context,
	unitID string,
	input CreateLayoutUnitPortInput,
	actor string,
) (*LayoutUnitPort, error) {
	unitID = strings.TrimSpace(unitID)
	input = cleanLayoutUnitPortInput(input)
	if unitID == "" || !validLayoutUnitPortInput(input) {
		return nil, ErrLayoutValidation
	}
	unit, err := s.portRepository.GetUnit(ctx, unitID)
	if err != nil {
		return nil, err
	}
	if !layoutUnitPortWithinBounds(input, *unit) {
		return nil, ErrLayoutValidation
	}
	return s.portRepository.CreateUnitPort(ctx, unitID, input, actor)
}

func (s *LayoutService) UpdateUnitPort(
	ctx context.Context,
	id string,
	input UpdateLayoutUnitPortInput,
	actor string,
) (*LayoutUnitPort, error) {
	id = strings.TrimSpace(id)
	input.CreateLayoutUnitPortInput = cleanLayoutUnitPortInput(input.CreateLayoutUnitPortInput)
	if id == "" || input.ExpectedVersion < 1 || !validLayoutUnitPortInput(input.CreateLayoutUnitPortInput) {
		return nil, ErrLayoutValidation
	}
	unit, err := s.portRepository.GetUnitForPort(ctx, id)
	if err != nil {
		return nil, err
	}
	if !layoutUnitPortWithinBounds(input.CreateLayoutUnitPortInput, *unit) {
		return nil, ErrLayoutValidation
	}
	return s.portRepository.UpdateUnitPort(ctx, id, input, actor)
}

func (s *LayoutService) ListConfigurations(ctx context.Context, layoutID string) ([]LayoutConfiguration, error) {
	return s.repository.ListConfigurations(ctx, strings.TrimSpace(layoutID))
}

func (s *LayoutService) AnalyzeConfigurationPorts(
	ctx context.Context,
	configurationID string,
) (*domain.ModulePortAnalysis, error) {
	configurationID = strings.TrimSpace(configurationID)
	if configurationID == "" {
		return nil, ErrLayoutValidation
	}
	placements, err := s.configurationPortRepository.LoadConfigurationPortPlacements(ctx, configurationID)
	if err != nil {
		return nil, err
	}
	analysis := domain.AnalyzeModulePorts(placements)
	return &analysis, nil
}

func (s *LayoutService) PreviewConfigurationUnitSnap(
	ctx context.Context,
	configurationID string,
	input PreviewConfigurationUnitSnapInput,
) (*domain.ModulePortSnapResult, error) {
	configurationID = strings.TrimSpace(configurationID)
	input.UnitID = strings.TrimSpace(input.UnitID)
	if configurationID == "" || input.UnitID == "" || !finite(input.PositionXMM) ||
		!finite(input.PositionYMM) || !finite(input.RotationDegrees) {
		return nil, ErrLayoutValidation
	}
	placements, err := s.configurationPortRepository.LoadConfigurationPortPlacements(ctx, configurationID)
	if err != nil {
		return nil, err
	}
	found := false
	for _, placement := range placements {
		if placement.UnitID == input.UnitID {
			found = true
			break
		}
	}
	if !found {
		return nil, ErrLayoutValidation
	}
	result := domain.FindModulePortSnap(input.UnitID, domain.TrackPose{
		PositionXMM: input.PositionXMM, PositionYMM: input.PositionYMM,
		RotationDegrees: input.RotationDegrees,
	}, placements)
	return &result, nil
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

func cleanLayoutUnitPortInput(input CreateLayoutUnitPortInput) CreateLayoutUnitPortInput {
	input.Name = strings.TrimSpace(input.Name)
	input.InterfaceKey = strings.ToLower(strings.TrimSpace(input.InterfaceKey))
	input.Notes = strings.TrimSpace(input.Notes)
	input.DirectionDegrees = normalizeRotation(input.DirectionDegrees)
	return input
}

func validLayoutUnitPortInput(input CreateLayoutUnitPortInput) bool {
	return input.Name != "" && input.Kind.Valid() && input.InterfaceKey != "" &&
		finite(input.XMM) && finite(input.YMM) && finite(input.DirectionDegrees) &&
		input.XMM >= 0 && input.YMM >= 0
}

func layoutUnitPortWithinBounds(input CreateLayoutUnitPortInput, unit LayoutUnit) bool {
	return (unit.WidthMM <= 0 || input.XMM <= unit.WidthMM) &&
		(unit.HeightMM <= 0 || input.YMM <= unit.HeightMM)
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
