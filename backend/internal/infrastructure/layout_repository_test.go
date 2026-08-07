package infrastructure_test

import (
	"errors"
	"math"
	"path/filepath"
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/domain"
	"railkeeper/backend/internal/infrastructure"
)

func TestLayoutServicePersistsStructureAndRejectsStaleUpdates(t *testing.T) {
	service := testLayoutService(t)
	ctx := t.Context()

	if _, err := service.CreateLayout(ctx, application.CreateLayoutInput{
		Name: "Incomplete",
		Kind: domain.LayoutKindPrivate,
	}, "admin-1"); !errors.Is(err, application.ErrLayoutValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}

	privateLayout, err := service.CreateLayout(ctx, application.CreateLayoutInput{
		Name: "  Heimanlage  ", Kind: domain.LayoutKindPrivate, Gauge: " TT ", Scale: " 1:120 ",
	}, "admin-1")
	if err != nil {
		t.Fatal(err)
	}
	clubLayout, err := service.CreateLayout(ctx, application.CreateLayoutInput{
		Name: "Clubanlage", Kind: domain.LayoutKindClub, Gauge: "TT", Scale: "1:120",
	}, "admin-1")
	if err != nil {
		t.Fatal(err)
	}
	if privateLayout.Name != "Heimanlage" || privateLayout.Gauge != "TT" || privateLayout.Version != 1 {
		t.Fatalf("unexpected normalized layout: %#v", privateLayout)
	}

	updated, err := service.UpdateLayout(ctx, privateLayout.ID, application.UpdateLayoutInput{
		CreateLayoutInput: application.CreateLayoutInput{
			Name: "Heimanlage erweitert", Kind: domain.LayoutKindPrivate, Gauge: "TT", Scale: "1:120",
		},
		ExpectedVersion: 1,
	}, "planner-1")
	if err != nil {
		t.Fatal(err)
	}
	if updated.Version != 2 {
		t.Fatalf("expected version 2, got %d", updated.Version)
	}
	if _, err := service.UpdateLayout(ctx, privateLayout.ID, application.UpdateLayoutInput{
		CreateLayoutInput: application.CreateLayoutInput{
			Name: "Stale", Kind: domain.LayoutKindPrivate, Gauge: "TT", Scale: "1:120",
		},
		ExpectedVersion: 1,
	}, "planner-1"); !errors.Is(err, application.ErrLayoutVersionConflict) {
		t.Fatalf("expected version conflict, got %v", err)
	}
	layouts, err := service.ListLayouts(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(layouts) != 2 || layouts[0].ID != clubLayout.ID || layouts[1].ID != privateLayout.ID {
		t.Fatalf("unexpected layout list: %#v", layouts)
	}
	storedLayout, err := service.GetLayout(ctx, privateLayout.ID)
	if err != nil {
		t.Fatal(err)
	}
	if storedLayout.Name != "Heimanlage erweitert" || storedLayout.Version != 2 {
		t.Fatalf("unexpected stored layout: %#v", storedLayout)
	}

	for _, kind := range []domain.LayoutUnitKind{
		domain.LayoutUnitKindBaseboard,
		domain.LayoutUnitKindModule,
		domain.LayoutUnitKindSegment,
		domain.LayoutUnitKindArea,
	} {
		if _, err := service.CreateUnit(ctx, clubLayout.ID, application.CreateLayoutUnitInput{
			Name: string(kind), Kind: kind, WidthMM: 1000, HeightMM: 500,
		}, "planner-1"); err != nil {
			t.Fatalf("create %s unit: %v", kind, err)
		}
	}
	if _, err := service.CreateUnit(ctx, clubLayout.ID, application.CreateLayoutUnitInput{
		Name: "Invalid", Kind: domain.LayoutUnitKindModule, WidthMM: -1,
	}, "planner-1"); !errors.Is(err, application.ErrLayoutValidation) {
		t.Fatalf("expected invalid dimension error, got %v", err)
	}

	units, err := service.ListUnits(ctx, clubLayout.ID)
	if err != nil {
		t.Fatal(err)
	}
	updatedUnit, err := service.UpdateUnit(ctx, units[0].ID, application.UpdateLayoutUnitInput{
		CreateLayoutUnitInput: application.CreateLayoutUnitInput{
			Name: "Updated unit", Kind: units[0].Kind, WidthMM: 1200, HeightMM: 600,
		},
		ExpectedVersion: units[0].Version,
	}, "planner-1")
	if err != nil {
		t.Fatal(err)
	}
	if updatedUnit.Version != 2 || updatedUnit.WidthMM != 1200 {
		t.Fatalf("unexpected updated unit: %#v", updatedUnit)
	}
	if _, err := service.UpdateUnit(ctx, units[0].ID, application.UpdateLayoutUnitInput{
		CreateLayoutUnitInput: application.CreateLayoutUnitInput{Name: "Stale unit", Kind: units[0].Kind},
		ExpectedVersion:       units[0].Version,
	}, "planner-1"); !errors.Is(err, application.ErrLayoutVersionConflict) {
		t.Fatalf("expected unit version conflict, got %v", err)
	}
	privateUnit, err := service.CreateUnit(ctx, privateLayout.ID, application.CreateLayoutUnitInput{
		Name: "Private segment", Kind: domain.LayoutUnitKindSegment,
	}, "planner-1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.SaveConfiguration(ctx, clubLayout.ID, application.SaveLayoutConfigurationInput{
		Name: "Wrong membership", Units: []application.ConfigurationUnitInput{{UnitID: privateUnit.ID}},
	}, "planner-1"); !errors.Is(err, application.ErrLayoutValidation) {
		t.Fatalf("expected configuration membership validation error, got %v", err)
	}
	if _, err := service.SaveConfiguration(ctx, clubLayout.ID, application.SaveLayoutConfigurationInput{
		Name:  "Invalid coordinates",
		Units: []application.ConfigurationUnitInput{{UnitID: units[0].ID, PositionXMM: math.NaN()}},
	}, "planner-1"); !errors.Is(err, application.ErrLayoutValidation) {
		t.Fatalf("expected finite coordinate validation error, got %v", err)
	}
	configuration, err := service.SaveConfiguration(ctx, clubLayout.ID, application.SaveLayoutConfigurationInput{
		Name: "Ausstellung",
		Units: []application.ConfigurationUnitInput{
			{UnitID: units[0].ID, PositionXMM: 100, RotationDegrees: -30},
			{UnitID: units[1].ID, PositionXMM: 1100, RotationDegrees: 725},
		},
	}, "planner-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(configuration.Units) != 2 || configuration.Units[0].RotationDegrees != 330 ||
		configuration.Units[1].RotationDegrees != 5 {
		t.Fatalf("unexpected normalized configuration: %#v", configuration.Units)
	}
	configurations, err := service.ListConfigurations(ctx, clubLayout.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(configurations) != 1 || configurations[0].ID != configuration.ID {
		t.Fatalf("unexpected configuration list: %#v", configurations)
	}
	updatedConfiguration, err := service.SaveConfiguration(ctx, clubLayout.ID, application.SaveLayoutConfigurationInput{
		ID: configuration.ID, Name: "Ausstellung erweitert", ExpectedVersion: configuration.Version,
		Units: configuration.Units,
	}, "planner-1")
	if err != nil {
		t.Fatal(err)
	}
	if updatedConfiguration.Version != 2 || updatedConfiguration.Name != "Ausstellung erweitert" {
		t.Fatalf("unexpected updated configuration: %#v", updatedConfiguration)
	}
	if _, err := service.SaveConfiguration(ctx, clubLayout.ID, application.SaveLayoutConfigurationInput{
		ID: configuration.ID, Name: "Stale setup", ExpectedVersion: configuration.Version,
		Units: configuration.Units,
	}, "planner-1"); !errors.Is(err, application.ErrLayoutVersionConflict) {
		t.Fatalf("expected configuration version conflict, got %v", err)
	}
}

func testLayoutService(t *testing.T) *application.LayoutService {
	t.Helper()
	db, err := infrastructure.OpenSQLite(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := infrastructure.Migrate(db, filepath.Join("..", "..", "migrations")); err != nil {
		t.Fatal(err)
	}
	return application.NewLayoutService(infrastructure.NewLayoutRepository(db))
}
