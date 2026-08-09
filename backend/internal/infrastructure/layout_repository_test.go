package infrastructure_test

import (
	"database/sql"
	"errors"
	"math"
	"path/filepath"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/domain"
	"railkeeper/backend/internal/infrastructure"
)

func TestLayoutUnitPortRepositoryPersistsAndRejectsStaleUpdates(t *testing.T) {
	db, service := testLayoutServiceWithDB(t)
	ctx := t.Context()
	layout, err := service.CreateLayout(ctx, application.CreateLayoutInput{
		Name: "Club", Kind: domain.LayoutKindClub, Gauge: "TT", Scale: "1:120",
	}, "admin-1")
	if err != nil {
		t.Fatal(err)
	}
	unit, err := service.CreateUnit(ctx, layout.ID, application.CreateLayoutUnitInput{
		Name: "Module A", Kind: domain.LayoutUnitKindModule, WidthMM: 1000, HeightMM: 500,
	}, "planner-1")
	if err != nil {
		t.Fatal(err)
	}

	east, err := service.CreateUnitPort(ctx, unit.ID, application.CreateLayoutUnitPortInput{
		Name: "East", Kind: domain.LayoutUnitPortTrack, InterfaceKey: "track:tillig-tt-modellgleis",
		XMM: 1000, YMM: 250, DirectionDegrees: 0,
	}, "planner-1")
	if err != nil {
		t.Fatal(err)
	}
	west, err := service.CreateUnitPort(ctx, unit.ID, application.CreateLayoutUnitPortInput{
		Name: "west", Kind: domain.LayoutUnitPortPower, InterfaceKey: "power:16v-ac",
		XMM: 0, YMM: 250, DirectionDegrees: 180, Archived: true,
	}, "planner-1")
	if err != nil {
		t.Fatal(err)
	}
	ports, err := service.ListUnitPorts(ctx, unit.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(ports) != 2 || ports[0].ID != east.ID || ports[1].ID != west.ID {
		t.Fatalf("unexpected active-first module port order: %#v", ports)
	}

	updated, err := service.UpdateUnitPort(ctx, east.ID, application.UpdateLayoutUnitPortInput{
		CreateLayoutUnitPortInput: application.CreateLayoutUnitPortInput{
			Name: "East main", Kind: domain.LayoutUnitPortTrack,
			InterfaceKey: "track:tillig-tt-modellgleis", XMM: 1000, YMM: 250, DirectionDegrees: 360,
		},
		ExpectedVersion: east.Version,
	}, "planner-1")
	if err != nil {
		t.Fatal(err)
	}
	if updated.Version != 2 || updated.DirectionDegrees != 0 || updated.Name != "East main" {
		t.Fatalf("unexpected updated module port: %#v", updated)
	}
	if _, err := service.UpdateUnitPort(ctx, east.ID, application.UpdateLayoutUnitPortInput{
		CreateLayoutUnitPortInput: application.CreateLayoutUnitPortInput{
			Name: "Stale", Kind: domain.LayoutUnitPortTrack, InterfaceKey: "track:tillig-tt-modellgleis",
		},
		ExpectedVersion: east.Version,
	}, "planner-1"); !errors.Is(err, application.ErrLayoutVersionConflict) {
		t.Fatalf("expected module port version conflict, got %v", err)
	}
	if _, err := service.CreateUnitPort(ctx, unit.ID, application.CreateLayoutUnitPortInput{
		Name: "EAST MAIN", Kind: domain.LayoutUnitPortTrack, InterfaceKey: "track:tillig-tt-modellgleis",
	}, "planner-1"); err == nil {
		t.Fatal("expected case-insensitive duplicate port name rejection")
	}

	var auditCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM audit_logs WHERE target_type='layout_unit_port' AND target_id=?`,
		east.ID).Scan(&auditCount); err != nil {
		t.Fatal(err)
	}
	if auditCount != 2 {
		t.Fatalf("expected create and update audit records, got %d", auditCount)
	}
}

func TestLayoutUnitPortMigrationSchema(t *testing.T) {
	db, _ := testLayoutServiceWithDB(t)
	var tableSQL string
	if err := db.QueryRow(`SELECT sql FROM sqlite_master WHERE type='table' AND name='layout_unit_ports'`).Scan(&tableSQL); err != nil {
		t.Fatal(err)
	}
	for _, fragment := range []string{"layout_unit_id", "interface_key", "direction_degrees"} {
		if !strings.Contains(strings.ToLower(tableSQL), fragment) {
			t.Fatalf("module port schema missing %q: %s", fragment, tableSQL)
		}
	}
}

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
	_, service := testLayoutServiceWithDB(t)
	return service
}

func testLayoutServiceWithDB(t *testing.T) (*sql.DB, *application.LayoutService) {
	t.Helper()
	db, err := infrastructure.OpenSQLite(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := infrastructure.Migrate(db, filepath.Join("..", "..", "migrations")); err != nil {
		t.Fatal(err)
	}
	return db, application.NewLayoutService(infrastructure.NewLayoutRepository(db))
}
