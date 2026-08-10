package infrastructure_test

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/domain"
	"railkeeper/backend/internal/infrastructure"
)

func TestLayoutServicePublishesImmutableRevisionHistory(t *testing.T) {
	service, _ := testRevisionServiceWithDB(t)
	ctx := t.Context()
	layout, err := service.CreateLayout(ctx, application.CreateLayoutInput{
		Name: "Heimanlage", Kind: domain.LayoutKindPrivate, Gauge: "TT", Scale: "1:120",
	}, "admin-1")
	if err != nil {
		t.Fatal(err)
	}
	unit, err := service.CreateUnit(ctx, layout.ID, application.CreateLayoutUnitInput{
		Name: "Grundplatte", Kind: domain.LayoutUnitKindBaseboard, WidthMM: 2000, HeightMM: 1000,
	}, "planner-1")
	if err != nil {
		t.Fatal(err)
	}
	variant, err := service.CreateVariant(ctx, unit.ID, application.CreatePlanVariantInput{Name: "Standard"}, "planner-1")
	if err != nil {
		t.Fatal(err)
	}

	first, err := service.CreateDraft(ctx, variant.ID, application.CreatePlanRevisionInput{}, "planner-1")
	if err != nil {
		t.Fatal(err)
	}
	if first.RevisionNumber != 1 || first.Status != domain.PlanRevisionDraft {
		t.Fatalf("unexpected first draft: %#v", first)
	}
	first, err = service.SubmitRevision(ctx, first.ID, first.Version, "planner-1")
	if err != nil {
		t.Fatal(err)
	}
	first, err = service.PublishRevision(ctx, first.ID, first.Version, "planner-1")
	if err != nil {
		t.Fatal(err)
	}
	if first.Status != domain.PlanRevisionPublished {
		t.Fatalf("expected published revision, got %#v", first)
	}
	if _, err := service.SubmitRevision(ctx, first.ID, first.Version, "planner-1"); !errors.Is(err, application.ErrPlanRevisionImmutable) {
		t.Fatalf("expected immutable revision error, got %v", err)
	}
	configuration, err := service.SaveConfiguration(ctx, layout.ID, application.SaveLayoutConfigurationInput{
		Name:  "Published setup",
		Units: []application.ConfigurationUnitInput{{UnitID: unit.ID, PlanRevisionID: first.ID}},
	}, "planner-1")
	if err != nil {
		t.Fatal(err)
	}

	second, err := service.CreateDraft(ctx, variant.ID, application.CreatePlanRevisionInput{
		BaseRevisionID: first.ID,
	}, "planner-1")
	if err != nil {
		t.Fatal(err)
	}
	if second.RevisionNumber != 2 || second.BaseRevisionID != first.ID {
		t.Fatalf("unexpected second draft: %#v", second)
	}
	if _, err := service.SaveConfiguration(ctx, layout.ID, application.SaveLayoutConfigurationInput{
		Name:  "Draft setup",
		Units: []application.ConfigurationUnitInput{{UnitID: unit.ID, PlanRevisionID: second.ID}},
	}, "planner-1"); !errors.Is(err, application.ErrLayoutValidation) {
		t.Fatalf("expected unpublished revision validation error, got %v", err)
	}
	if _, err := service.PublishRevision(ctx, second.ID, second.Version+1, "planner-1"); !errors.Is(err, application.ErrPlanRevisionConflict) {
		t.Fatalf("expected plan revision conflict, got %v", err)
	}
	if _, err := service.PublishRevision(ctx, second.ID, second.Version, "planner-1"); err != nil {
		t.Fatal(err)
	}
	variants, err := service.ListVariants(ctx, unit.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(variants) != 1 || len(variants[0].Revisions) != 2 {
		t.Fatalf("unexpected revision history: %#v", variants)
	}
	if variants[0].Revisions[0].Status != domain.PlanRevisionArchived ||
		variants[0].Revisions[1].Status != domain.PlanRevisionPublished {
		t.Fatalf("unexpected revision statuses: %#v", variants[0].Revisions)
	}
	configuration, err = service.SaveConfiguration(ctx, layout.ID, application.SaveLayoutConfigurationInput{
		ID: configuration.ID, Name: configuration.Name, ExpectedVersion: configuration.Version,
		Units: configuration.Units,
	}, "planner-1")
	if err != nil {
		t.Fatalf("archived published revision must remain usable in an existing setup: %v", err)
	}
	if configuration.Version != 2 {
		t.Fatalf("expected configuration version 2, got %d", configuration.Version)
	}
}

func TestLayoutServiceWritesAuditTrail(t *testing.T) {
	service, db := testRevisionServiceWithDB(t)
	ctx := t.Context()
	layout, err := service.CreateLayout(ctx, application.CreateLayoutInput{
		Name: "Audit layout", Kind: domain.LayoutKindPrivate, Gauge: "TT", Scale: "1:120",
	}, "admin-1")
	if err != nil {
		t.Fatal(err)
	}
	unit, err := service.CreateUnit(ctx, layout.ID, application.CreateLayoutUnitInput{
		Name: "Audit module", Kind: domain.LayoutUnitKindModule,
	}, "planner-1")
	if err != nil {
		t.Fatal(err)
	}
	variant, err := service.CreateVariant(ctx, unit.ID, application.CreatePlanVariantInput{Name: "Audit plan"}, "planner-1")
	if err != nil {
		t.Fatal(err)
	}
	revision, err := service.CreateDraft(ctx, variant.ID, application.CreatePlanRevisionInput{}, "planner-1")
	if err != nil {
		t.Fatal(err)
	}
	revision, err = service.SubmitRevision(ctx, revision.ID, revision.Version, "planner-1")
	if err != nil {
		t.Fatal(err)
	}
	revision, err = service.PublishRevision(ctx, revision.ID, revision.Version, "planner-1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.SaveConfiguration(ctx, layout.ID, application.SaveLayoutConfigurationInput{
		Name:  "Audit setup",
		Units: []application.ConfigurationUnitInput{{UnitID: unit.ID, PlanRevisionID: revision.ID}},
	}, "planner-1"); err != nil {
		t.Fatal(err)
	}

	rows, err := db.QueryContext(ctx, `SELECT action FROM audit_logs ORDER BY rowid`)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	actions := []string{}
	for rows.Next() {
		var action string
		if err := rows.Scan(&action); err != nil {
			t.Fatal(err)
		}
		actions = append(actions, action)
	}
	want := []string{
		"LayoutCreated", "LayoutUnitCreated", "PlanVariantCreated", "PlanDraftCreated",
		"PlanRevisionSubmitted", "PlanRevisionPublished", "LayoutConfigurationSaved",
	}
	if len(actions) != len(want) {
		t.Fatalf("unexpected audit actions: %#v", actions)
	}
	for index := range want {
		if actions[index] != want[index] {
			t.Fatalf("unexpected audit action %d: got %q, want %q", index, actions[index], want[index])
		}
	}
}

func TestLayoutRevisionClonePreservesFlexiblePaths(t *testing.T) {
	service, db := testRevisionServiceWithDB(t)
	planner := application.NewTrackPlannerService(infrastructure.NewTrackPlannerRepository(db))
	ctx := t.Context()
	layout, err := service.CreateLayout(ctx, application.CreateLayoutInput{
		Name: "Flexanlage", Kind: domain.LayoutKindPrivate, Gauge: "TT", Scale: "1:120",
	}, "admin")
	if err != nil {
		t.Fatal(err)
	}
	unit, err := service.CreateUnit(ctx, layout.ID, application.CreateLayoutUnitInput{
		Name: "Segment", Kind: domain.LayoutUnitKindModule,
	}, "planner")
	if err != nil {
		t.Fatal(err)
	}
	variant, err := service.CreateVariant(ctx, unit.ID, application.CreatePlanVariantInput{Name: "Flex"}, "planner")
	if err != nil {
		t.Fatal(err)
	}
	base, err := service.CreateDraft(ctx, variant.ID, application.CreatePlanRevisionInput{}, "planner")
	if err != nil {
		t.Fatal(err)
	}
	path := domain.FlexTrackPath{
		SchemaVersion: 1, EndXMM: 500, EndYMM: 100,
		StartHandleMM: 180, EndHandleMM: 180,
	}
	created, err := planner.CreateObject(ctx, base.ID, application.CreatePlanTrackObjectInput{
		GeometryID: "tillig-tt-modellgleis-83125-v1", FlexPath: &path,
	}, "planner")
	if err != nil {
		t.Fatal(err)
	}
	transitionPath := domain.TransitionCurvePath{
		SchemaVersion: 1, LengthMM: 450, EndRadiusMM: 700, Direction: domain.TransitionRight,
	}
	createdTransition, err := planner.CreateObject(ctx, base.ID, application.CreatePlanTrackObjectInput{
		GeometryID: "tillig-tt-modellgleis-83125-v1", TransitionPath: &transitionPath,
	}, "planner")
	if err != nil {
		t.Fatal(err)
	}
	base, err = service.SubmitRevision(ctx, base.ID, base.Version, "planner")
	if err != nil {
		t.Fatal(err)
	}
	base, err = service.PublishRevision(ctx, base.ID, base.Version, "planner")
	if err != nil {
		t.Fatal(err)
	}
	clone, err := service.CreateDraft(ctx, variant.ID, application.CreatePlanRevisionInput{
		BaseRevisionID: base.ID,
	}, "planner")
	if err != nil {
		t.Fatal(err)
	}
	plan, err := planner.GetPlan(ctx, clone.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Objects) != 2 {
		t.Fatalf("flexible paths not cloned: %#v", plan.Objects)
	}
	cloned := map[string]domain.PlanTrackObject{}
	for _, object := range plan.Objects {
		cloned[object.LineageID] = object
	}
	if cloned[created.LineageID].FlexPath == nil || cloned[created.LineageID].FlexPath.EndYMM != 100 {
		t.Fatalf("flex path not cloned: %#v", plan.Objects)
	}
	transitionClone := cloned[createdTransition.LineageID]
	if transitionClone.TransitionPath == nil ||
		transitionClone.TransitionPath.Direction != domain.TransitionRight || transitionClone.FlexPath != nil {
		t.Fatalf("transition path not cloned: %#v", plan.Objects)
	}
}

func testRevisionServiceWithDB(t *testing.T) (*application.LayoutService, *sql.DB) {
	t.Helper()
	db, err := infrastructure.OpenSQLite(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := infrastructure.Migrate(db, filepath.Join("..", "..", "migrations")); err != nil {
		t.Fatal(err)
	}
	return application.NewLayoutService(infrastructure.NewLayoutRepository(db)), db
}
