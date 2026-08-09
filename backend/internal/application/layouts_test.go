package application

import (
	"context"
	"errors"
	"math"
	"testing"

	"railkeeper/backend/internal/domain"
)

type layoutRepositorySpy struct {
	LayoutRepository
	createdLayout      CreateLayoutInput
	savedConfiguration SaveLayoutConfigurationInput
	savedOutline       UpdateLayoutUnitOutlineInput
	unit               LayoutUnit
	createdPort        CreateLayoutUnitPortInput
	updatedPort        UpdateLayoutUnitPortInput
	configurationPorts []domain.ModulePortPlacement
}

func (spy *layoutRepositorySpy) CreateLayout(
	_ context.Context,
	input CreateLayoutInput,
	_ string,
) (*Layout, error) {
	spy.createdLayout = input
	return &Layout{Name: input.Name, Gauge: input.Gauge, Scale: input.Scale}, nil
}

func (spy *layoutRepositorySpy) SaveConfiguration(
	_ context.Context,
	_ string,
	input SaveLayoutConfigurationInput,
	_ string,
) (*LayoutConfiguration, error) {
	spy.savedConfiguration = input
	return &LayoutConfiguration{Name: input.Name, Units: input.Units}, nil
}

func (spy *layoutRepositorySpy) CreateUnit(
	_ context.Context,
	_ string,
	input CreateLayoutUnitInput,
	_ string,
) (*LayoutUnit, error) {
	return &LayoutUnit{Name: input.Name, WidthMM: input.WidthMM, HeightMM: input.HeightMM}, nil
}

func (spy *layoutRepositorySpy) GetUnit(_ context.Context, id string) (*LayoutUnit, error) {
	unit := spy.unit
	unit.ID = id
	return &unit, nil
}

func (spy *layoutRepositorySpy) GetUnitForPort(_ context.Context, _ string) (*LayoutUnit, error) {
	unit := spy.unit
	return &unit, nil
}

func (spy *layoutRepositorySpy) ListUnitPorts(_ context.Context, _ string) ([]LayoutUnitPort, error) {
	return []LayoutUnitPort{}, nil
}

func (spy *layoutRepositorySpy) CreateUnitPort(
	_ context.Context,
	unitID string,
	input CreateLayoutUnitPortInput,
	_ string,
) (*LayoutUnitPort, error) {
	spy.createdPort = input
	return &LayoutUnitPort{LayoutUnitID: unitID, Name: input.Name, Kind: input.Kind,
		InterfaceKey: input.InterfaceKey, XMM: input.XMM, YMM: input.YMM,
		DirectionDegrees: input.DirectionDegrees}, nil
}

func (spy *layoutRepositorySpy) UpdateUnitPort(
	_ context.Context,
	id string,
	input UpdateLayoutUnitPortInput,
	_ string,
) (*LayoutUnitPort, error) {
	spy.updatedPort = input
	return &LayoutUnitPort{ID: id, Name: input.Name, Kind: input.Kind,
		InterfaceKey: input.InterfaceKey, XMM: input.XMM, YMM: input.YMM,
		DirectionDegrees: input.DirectionDegrees, Version: input.ExpectedVersion + 1}, nil
}

func (spy *layoutRepositorySpy) LoadConfigurationPortPlacements(
	_ context.Context,
	_ string,
) ([]domain.ModulePortPlacement, error) {
	return spy.configurationPorts, nil
}

func (spy *layoutRepositorySpy) UpdateUnitOutline(
	_ context.Context,
	unitID string,
	input UpdateLayoutUnitOutlineInput,
	_ string,
) (*LayoutUnitOutline, error) {
	spy.savedOutline = input
	return &LayoutUnitOutline{LayoutUnitID: unitID, Points: input.Points, Version: input.ExpectedVersion + 1}, nil
}

func TestLayoutServiceValidatesUnitOutline(t *testing.T) {
	repository := &layoutRepositorySpy{}
	service := NewLayoutService(repository)
	valid := UpdateLayoutUnitOutlineInput{ExpectedVersion: 2, Points: []LayoutTwinPoint{
		{XMM: 0, YMM: 0}, {XMM: 100, YMM: 0}, {XMM: 50, YMM: 80},
	}}
	outline, err := service.UpdateUnitOutline(t.Context(), " unit-1 ", valid, "planner")
	if err != nil {
		t.Fatal(err)
	}
	if outline.LayoutUnitID != "unit-1" || outline.Version != 3 || len(repository.savedOutline.Points) != 3 {
		t.Fatalf("unexpected outline: %#v", outline)
	}
	invalid := []UpdateLayoutUnitOutlineInput{
		{ExpectedVersion: 1, Points: valid.Points[:2]},
		{ExpectedVersion: 0, Points: valid.Points},
		{ExpectedVersion: 1, Points: []LayoutTwinPoint{{XMM: 0}, {XMM: 1}, {XMM: 2}}},
		{ExpectedVersion: 1, Points: []LayoutTwinPoint{{XMM: math.NaN()}, {XMM: 1}, {YMM: 1}}},
	}
	for _, input := range invalid {
		if _, err := service.UpdateUnitOutline(t.Context(), "unit-1", input, "planner"); !errors.Is(err, ErrLayoutValidation) {
			t.Fatalf("expected invalid outline rejection for %#v, got %v", input, err)
		}
	}
}

func TestLayoutServiceNormalizesInputsBeforePersistence(t *testing.T) {
	repository := &layoutRepositorySpy{}
	service := NewLayoutService(repository)

	if _, err := service.CreateLayout(t.Context(), CreateLayoutInput{
		Name: " Home ", Kind: domain.LayoutKindPrivate, Gauge: " TT ", Scale: " 1:120 ",
	}, "planner-1"); err != nil {
		t.Fatal(err)
	}
	if repository.createdLayout.Name != "Home" || repository.createdLayout.Gauge != "TT" ||
		repository.createdLayout.Scale != "1:120" {
		t.Fatalf("unexpected normalized layout input: %#v", repository.createdLayout)
	}

	if _, err := service.SaveConfiguration(t.Context(), "layout-1", SaveLayoutConfigurationInput{
		Name: " Setup ",
		Units: []ConfigurationUnitInput{
			{UnitID: " unit-1 ", RotationDegrees: -15, SortOrder: 9},
			{UnitID: "unit-2", RotationDegrees: 721, SortOrder: 9},
		},
	}, "planner-1"); err != nil {
		t.Fatal(err)
	}
	got := repository.savedConfiguration
	if got.Name != "Setup" || got.Units[0].UnitID != "unit-1" || got.Units[0].RotationDegrees != 345 ||
		got.Units[0].SortOrder != 0 || got.Units[1].RotationDegrees != 1 || got.Units[1].SortOrder != 1 {
		t.Fatalf("unexpected normalized configuration input: %#v", got)
	}
}

func TestLayoutServiceRejectsNonFiniteConfigurationCoordinates(t *testing.T) {
	service := NewLayoutService(&layoutRepositorySpy{})
	for name, value := range map[string]float64{
		"not a number":      math.NaN(),
		"positive infinity": math.Inf(1),
		"negative infinity": math.Inf(-1),
	} {
		t.Run(name, func(t *testing.T) {
			_, err := service.SaveConfiguration(t.Context(), "layout-1", SaveLayoutConfigurationInput{
				Name: "Setup", Units: []ConfigurationUnitInput{{UnitID: "unit-1", PositionXMM: value}},
			}, "planner-1")
			if !errors.Is(err, ErrLayoutValidation) {
				t.Fatalf("expected layout validation error, got %v", err)
			}
		})
	}
}

func TestLayoutServiceRejectsNonFiniteUnitDimensions(t *testing.T) {
	service := NewLayoutService(&layoutRepositorySpy{})
	for name, value := range map[string]float64{
		"not a number": math.NaN(),
		"infinity":     math.Inf(1),
	} {
		t.Run(name, func(t *testing.T) {
			_, err := service.CreateUnit(t.Context(), "layout-1", CreateLayoutUnitInput{
				Name: "Module", Kind: domain.LayoutUnitKindModule, WidthMM: value,
			}, "planner-1")
			if !errors.Is(err, ErrLayoutValidation) {
				t.Fatalf("expected layout validation error, got %v", err)
			}
		})
	}
}

func TestLayoutUnitPortKinds(t *testing.T) {
	for _, kind := range []domain.LayoutUnitPortKind{
		domain.LayoutUnitPortTrack,
		domain.LayoutUnitPortPower,
		domain.LayoutUnitPortDigital,
		domain.LayoutUnitPortFeedback,
		domain.LayoutUnitPortAccessory,
		domain.LayoutUnitPortOther,
	} {
		if !kind.Valid() {
			t.Fatalf("expected valid layout unit port kind %q", kind)
		}
	}
	if domain.LayoutUnitPortKind("video").Valid() {
		t.Fatal("unexpected valid unknown layout unit port kind")
	}
}

func TestLayoutUnitPortServiceNormalizesBeforePersistence(t *testing.T) {
	repository := &layoutRepositorySpy{unit: LayoutUnit{WidthMM: 1000, HeightMM: 500}}
	service := NewLayoutService(repository)

	port, err := service.CreateUnitPort(t.Context(), " unit-1 ", CreateLayoutUnitPortInput{
		Name: " West ", Kind: domain.LayoutUnitPortTrack,
		InterfaceKey: " TRACK:Tillig-TT-Modellgleis ", XMM: 0, YMM: 250,
		DirectionDegrees: -180, Notes: " Übergang zum Schattenbahnhof ",
	}, "planner")
	if err != nil {
		t.Fatal(err)
	}
	if port.LayoutUnitID != "unit-1" || repository.createdPort.Name != "West" ||
		repository.createdPort.InterfaceKey != "track:tillig-tt-modellgleis" ||
		repository.createdPort.DirectionDegrees != 180 ||
		repository.createdPort.Notes != "Übergang zum Schattenbahnhof" {
		t.Fatalf("unexpected normalized module port: %#v / %#v", port, repository.createdPort)
	}

	updated, err := service.UpdateUnitPort(t.Context(), " port-1 ", UpdateLayoutUnitPortInput{
		CreateLayoutUnitPortInput: repository.createdPort,
		ExpectedVersion:           2,
	}, "planner")
	if err != nil {
		t.Fatal(err)
	}
	if updated.ID != "port-1" || updated.Version != 3 || repository.updatedPort.ExpectedVersion != 2 {
		t.Fatalf("unexpected updated module port: %#v", updated)
	}
}

func TestLayoutUnitPortServiceRejectsInvalidInput(t *testing.T) {
	repository := &layoutRepositorySpy{unit: LayoutUnit{WidthMM: 1000, HeightMM: 500}}
	service := NewLayoutService(repository)
	valid := CreateLayoutUnitPortInput{
		Name: "West", Kind: domain.LayoutUnitPortTrack,
		InterfaceKey: "track:tillig-tt-modellgleis", XMM: 0, YMM: 250,
	}
	tests := map[string]CreateLayoutUnitPortInput{
		"blank name":        {Kind: valid.Kind, InterfaceKey: valid.InterfaceKey},
		"invalid kind":      {Name: valid.Name, Kind: "video", InterfaceKey: valid.InterfaceKey},
		"blank interface":   {Name: valid.Name, Kind: valid.Kind},
		"negative x":        {Name: valid.Name, Kind: valid.Kind, InterfaceKey: valid.InterfaceKey, XMM: -1},
		"outside width":     {Name: valid.Name, Kind: valid.Kind, InterfaceKey: valid.InterfaceKey, XMM: 1001},
		"outside height":    {Name: valid.Name, Kind: valid.Kind, InterfaceKey: valid.InterfaceKey, YMM: 501},
		"not a number":      {Name: valid.Name, Kind: valid.Kind, InterfaceKey: valid.InterfaceKey, XMM: math.NaN()},
		"positive infinity": {Name: valid.Name, Kind: valid.Kind, InterfaceKey: valid.InterfaceKey, YMM: math.Inf(1)},
	}
	for name, input := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := service.CreateUnitPort(t.Context(), "unit-1", input, "planner"); !errors.Is(err, ErrLayoutValidation) {
				t.Fatalf("expected layout validation error, got %v", err)
			}
		})
	}
	if _, err := service.UpdateUnitPort(t.Context(), "port-1", UpdateLayoutUnitPortInput{
		CreateLayoutUnitPortInput: valid,
	}, "planner"); !errors.Is(err, ErrLayoutValidation) {
		t.Fatalf("expected version validation error, got %v", err)
	}
}

func TestLayoutServiceAnalyzesConfigurationPortsAndPreviewsWithoutMutation(t *testing.T) {
	repository := &layoutRepositorySpy{configurationPorts: []domain.ModulePortPlacement{
		{UnitID: "moving", UnitName: "West", PortID: "east", PortName: "Ost",
			Kind: domain.LayoutUnitPortTrack, InterfaceKey: "track:tillig-tt-modellgleis",
			XMM: 100, DirectionDegrees: 0, UnitPose: domain.TrackPose{PositionXMM: 5}},
		{UnitID: "target", UnitName: "Ost", PortID: "west", PortName: "West",
			Kind: domain.LayoutUnitPortTrack, InterfaceKey: "track:tillig-tt-modellgleis",
			DirectionDegrees: 180, UnitPose: domain.TrackPose{PositionXMM: 112}},
	}}
	service := NewLayoutService(repository)

	analysis, err := service.AnalyzeConfigurationPorts(t.Context(), " configuration-1 ")
	if err != nil {
		t.Fatal(err)
	}
	if len(analysis.Issues) != 2 || len(analysis.Connections) != 0 {
		t.Fatalf("unexpected analysis before snap: %#v", analysis)
	}

	preview, err := service.PreviewConfigurationUnitSnap(t.Context(), " configuration-1 ",
		PreviewConfigurationUnitSnapInput{UnitID: " moving ", PositionXMM: 5})
	if err != nil {
		t.Fatal(err)
	}
	if !preview.Snapped || preview.Pose.PositionXMM != 12 || preview.TargetUnitID != "target" {
		t.Fatalf("unexpected snap preview: %#v", preview)
	}
	if repository.configurationPorts[0].UnitPose.PositionXMM != 5 {
		t.Fatalf("preview mutated repository context: %#v", repository.configurationPorts[0])
	}
}

func TestLayoutServiceRejectsInvalidConfigurationPortRequests(t *testing.T) {
	service := NewLayoutService(&layoutRepositorySpy{configurationPorts: []domain.ModulePortPlacement{
		{UnitID: "unit-1", PortID: "port-1", Kind: domain.LayoutUnitPortTrack,
			InterfaceKey: "track:tillig-tt-modellgleis"},
	}})
	if _, err := service.AnalyzeConfigurationPorts(t.Context(), " "); !errors.Is(err, ErrLayoutValidation) {
		t.Fatalf("expected blank configuration rejection, got %v", err)
	}
	invalid := []PreviewConfigurationUnitSnapInput{
		{},
		{UnitID: "missing"},
		{UnitID: "unit-1", PositionXMM: math.NaN()},
		{UnitID: "unit-1", PositionYMM: math.Inf(1)},
	}
	for _, input := range invalid {
		if _, err := service.PreviewConfigurationUnitSnap(t.Context(), "configuration-1", input); !errors.Is(err, ErrLayoutValidation) {
			t.Fatalf("expected invalid preview rejection for %#v, got %v", input, err)
		}
	}
}
