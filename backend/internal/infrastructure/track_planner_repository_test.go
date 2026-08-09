package infrastructure_test

import (
	"errors"
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/domain"
	"railkeeper/backend/internal/infrastructure"
)

func TestTrackPlannerRepositoryPersistsVersionsAndClonesDraftObjects(t *testing.T) {
	db := openTrackPlannerSchemaDB(t)
	layouts := application.NewLayoutService(infrastructure.NewLayoutRepository(db))
	planner := application.NewTrackPlannerService(infrastructure.NewTrackPlannerRepository(db))
	ctx := t.Context()

	layout, err := layouts.CreateLayout(ctx, application.CreateLayoutInput{
		Name: "Clubanlage", Kind: domain.LayoutKindClub, Gauge: "TT", Scale: "1:120",
	}, "admin")
	if err != nil {
		t.Fatal(err)
	}
	unit, err := layouts.CreateUnit(ctx, layout.ID, application.CreateLayoutUnitInput{
		Name: "Bahnhof", Kind: domain.LayoutUnitKindModule, WidthMM: 1200, HeightMM: 500,
	}, "planner")
	if err != nil {
		t.Fatal(err)
	}
	variant, err := layouts.CreateVariant(ctx, unit.ID, application.CreatePlanVariantInput{Name: "Standard"}, "planner")
	if err != nil {
		t.Fatal(err)
	}
	draft, err := layouts.CreateDraft(ctx, variant.ID, application.CreatePlanRevisionInput{}, "planner")
	if err != nil {
		t.Fatal(err)
	}

	geometries, err := planner.ListGeometries(ctx, "TT")
	if err != nil {
		t.Fatal(err)
	}
	if len(geometries) != 1 || geometries[0].ArticleNumber != "83101" {
		t.Fatalf("unexpected verified geometries: %#v", geometries)
	}

	created, err := planner.CreateObject(ctx, draft.ID, application.CreatePlanTrackObjectInput{
		GeometryID: geometries[0].ID, PositionXMM: 100, PositionYMM: 50, RotationDegrees: -15,
	}, "planner")
	if err != nil {
		t.Fatal(err)
	}
	if created.Version != 1 || created.RotationDegrees != 345 || created.Geometry.ArticleNumber != "83101" {
		t.Fatalf("unexpected created track object: %#v", created)
	}

	updated, err := planner.UpdateObject(ctx, created.ID, application.UpdatePlanTrackObjectInput{
		PositionXMM: 110, PositionYMM: 55, RotationDegrees: 15, ExpectedVersion: created.Version,
	}, "planner")
	if err != nil {
		t.Fatal(err)
	}
	if updated.Version != 2 || updated.PositionXMM != 110 || updated.RotationDegrees != 15 {
		t.Fatalf("unexpected updated track object: %#v", updated)
	}
	if _, err := planner.UpdateObject(ctx, created.ID, application.UpdatePlanTrackObjectInput{
		PositionXMM: 120, PositionYMM: 60, ExpectedVersion: created.Version,
	}, "planner"); !errors.Is(err, application.ErrTrackPlanConflict) {
		t.Fatalf("expected track object conflict, got %v", err)
	}

	published, err := layouts.PublishRevision(ctx, draft.ID, draft.Version, "planner")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := planner.CreateObject(ctx, published.ID, application.CreatePlanTrackObjectInput{
		GeometryID: geometries[0].ID,
	}, "planner"); !errors.Is(err, application.ErrTrackPlanImmutable) {
		t.Fatalf("expected immutable published plan, got %v", err)
	}

	clone, err := layouts.CreateDraft(ctx, variant.ID, application.CreatePlanRevisionInput{
		BaseRevisionID: published.ID,
	}, "planner")
	if err != nil {
		t.Fatal(err)
	}
	clonedPlan, err := planner.GetPlan(ctx, clone.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(clonedPlan.Objects) != 1 || clonedPlan.Objects[0].ID == updated.ID ||
		clonedPlan.Objects[0].PositionXMM != updated.PositionXMM || clonedPlan.Objects[0].Version != 1 {
		t.Fatalf("unexpected cloned track plan: %#v", clonedPlan)
	}
	if err := planner.DeleteObject(ctx, clonedPlan.Objects[0].ID, clonedPlan.Objects[0].Version, "planner"); err != nil {
		t.Fatal(err)
	}
	clonedPlan, err = planner.GetPlan(ctx, clone.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(clonedPlan.Objects) != 0 {
		t.Fatalf("deleted track object remains: %#v", clonedPlan.Objects)
	}

	for _, action := range []string{"PlanTrackObjectCreated", "PlanTrackObjectUpdated", "PlanTrackObjectDeleted"} {
		var count int
		if err := db.QueryRow(`SELECT COUNT(*) FROM audit_logs WHERE action=?`, action).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 1 {
			t.Fatalf("unexpected %s audit count: %d", action, count)
		}
	}
}

func TestTrackPlannerRepositoryRejectsUnverifiedGeometry(t *testing.T) {
	db := openTrackPlannerSchemaDB(t)
	seedTrackPlanRevision(t, db)
	if _, err := db.Exec(`
INSERT INTO track_geometry_definitions(
  id, library_id, article_number, name, kind, length_mm, geometry_json,
  source_url, status, created_at
) SELECT 'draft-geometry', library_id, 'draft', 'Entwurf', kind, length_mm, geometry_json,
         source_url, 'draft', created_at
  FROM track_geometry_definitions WHERE article_number='83101'`); err != nil {
		t.Fatal(err)
	}
	planner := application.NewTrackPlannerService(infrastructure.NewTrackPlannerRepository(db))
	_, err := planner.CreateObject(t.Context(), "revision-track-1", application.CreatePlanTrackObjectInput{
		GeometryID: "draft-geometry",
	}, "planner")
	if !errors.Is(err, application.ErrTrackPlanValidation) {
		t.Fatalf("expected unverified geometry validation error, got %v", err)
	}
}
