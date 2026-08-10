package infrastructure_test

import (
	"errors"
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/domain"
	"railkeeper/backend/internal/infrastructure"
)

func TestFreePlanObjectRepositoryPersistsDraftObjectsAndVersions(t *testing.T) {
	db := openTrackPlannerSchemaDB(t)
	seedTrackPlanRevision(t, db)
	repository := infrastructure.NewTrackPlannerRepository(db)
	rectangle := domain.FreePlanObjectShape{
		SchemaVersion: 1, Kind: domain.FreePlanRectangle,
		WidthMM: freePlanFloat(500), HeightMM: freePlanFloat(70),
	}

	created, err := repository.CreateFreeObject(t.Context(), "revision-track-1",
		application.CreateFreePlanObjectInput{
			Name: "Bahnsteig 1", Category: domain.FreePlanPlatform,
			PositionXMM: 200, PositionYMM: 100, RotationDegrees: 15, Shape: rectangle,
		}, "planner-1")
	if err != nil {
		t.Fatal(err)
	}
	if created.Version != 1 || created.LineageID != created.ID || created.Shape.WidthMM == nil ||
		*created.Shape.WidthMM != 500 {
		t.Fatalf("unexpected created free object: %#v", created)
	}
	plan, err := repository.GetPlan(t.Context(), "revision-track-1")
	if err != nil {
		t.Fatal(err)
	}
	if plan.FreeObjects == nil || len(plan.FreeObjects) != 1 || plan.FreeObjects[0].ID != created.ID {
		t.Fatalf("free object missing from plan: %#v", plan.FreeObjects)
	}

	updated, err := repository.UpdateFreeObject(t.Context(), created.ID,
		application.UpdateFreePlanObjectInput{
			CreateFreePlanObjectInput: application.CreateFreePlanObjectInput{
				Name: "Gebäude", Category: domain.FreePlanStructure,
				PositionXMM: 250, PositionYMM: 110, RotationDegrees: 30,
				Shape: domain.FreePlanObjectShape{
					SchemaVersion: 1, Kind: domain.FreePlanEllipse,
					WidthMM: freePlanFloat(120), HeightMM: freePlanFloat(80),
				},
			},
			ExpectedVersion: created.Version,
		}, "planner-1")
	if err != nil {
		t.Fatal(err)
	}
	if updated.Version != 2 || updated.Name != "Gebäude" || updated.Shape.Kind != domain.FreePlanEllipse {
		t.Fatalf("unexpected updated free object: %#v", updated)
	}
	if _, err := repository.UpdateFreeObject(t.Context(), created.ID,
		application.UpdateFreePlanObjectInput{
			CreateFreePlanObjectInput: application.CreateFreePlanObjectInput{
				Name: "Veraltet", Category: domain.FreePlanPlatform, Shape: rectangle,
			},
			ExpectedVersion: created.Version,
		}, "planner-1"); !errors.Is(err, application.ErrTrackPlanConflict) {
		t.Fatalf("expected stale free-object conflict, got %v", err)
	}
	if err := repository.DeleteFreeObject(t.Context(), created.ID, created.Version, "planner-1"); !errors.Is(err, application.ErrTrackPlanConflict) {
		t.Fatalf("expected stale delete conflict, got %v", err)
	}
	if err := repository.DeleteFreeObject(t.Context(), created.ID, updated.Version, "planner-1"); err != nil {
		t.Fatal(err)
	}
	plan, err = repository.GetPlan(t.Context(), "revision-track-1")
	if err != nil || plan.FreeObjects == nil || len(plan.FreeObjects) != 0 {
		t.Fatalf("free object was not deleted: %#v, %v", plan.FreeObjects, err)
	}
	var audits int
	if err := db.QueryRow(`SELECT COUNT(*) FROM audit_logs
WHERE target_type='plan_free_object' AND target_id=?`, created.ID).Scan(&audits); err != nil {
		t.Fatal(err)
	}
	if audits != 3 {
		t.Fatalf("expected three free-object audit events, got %d", audits)
	}
}

func TestFreePlanObjectRepositoryEnforcesDraftAndStrictShapes(t *testing.T) {
	db := openTrackPlannerSchemaDB(t)
	seedTrackPlanRevision(t, db)
	repository := infrastructure.NewTrackPlannerRepository(db)
	if _, err := db.Exec(`UPDATE plan_revisions SET status='published' WHERE id='revision-track-1'`); err != nil {
		t.Fatal(err)
	}
	if _, err := repository.CreateFreeObject(t.Context(), "revision-track-1",
		application.CreateFreePlanObjectInput{
			Name: "Linie", Category: domain.FreePlanAnnotation,
			Shape: domain.FreePlanObjectShape{
				SchemaVersion: 1, Kind: domain.FreePlanLine,
				EndXMM: freePlanFloat(100), EndYMM: freePlanFloat(0),
			},
		}, "planner-1"); !errors.Is(err, application.ErrTrackPlanImmutable) {
		t.Fatalf("expected immutable revision, got %v", err)
	}
	if _, err := db.Exec(`UPDATE plan_revisions SET status='draft' WHERE id='revision-track-1'`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO plan_free_objects(
id, lineage_id, revision_id, name, category, position_x_mm, position_y_mm, rotation_degrees,
shape_json, version, created_at, updated_at
) VALUES('invalid-free', 'invalid-free', 'revision-track-1', 'Ungültig', 'structure', 0, 0, 0,
'{}', 1, 'now', 'now')`); err != nil {
		t.Fatal(err)
	}
	if _, err := repository.GetPlan(t.Context(), "revision-track-1"); err == nil {
		t.Fatal("expected invalid persisted shape to fail strict hydration")
	}
}

func TestFreePlanObjectRepositoryCascadesWithRevision(t *testing.T) {
	db := openTrackPlannerSchemaDB(t)
	seedTrackPlanRevision(t, db)
	if _, err := db.Exec(`INSERT INTO plan_free_objects(
id, lineage_id, revision_id, name, category, position_x_mm, position_y_mm, rotation_degrees,
shape_json, version, created_at, updated_at
) VALUES('free-1', 'free-1', 'revision-track-1', 'Fläche', 'structure', 0, 0, 0,
'{"schemaVersion":1,"kind":"rectangle","widthMm":100,"heightMm":50}', 1, 'now', 'now')`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`DELETE FROM plan_revisions WHERE id='revision-track-1'`); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM plan_free_objects`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("expected cascading free-object deletion, got %d", count)
	}
}

func freePlanFloat(value float64) *float64 { return &value }
