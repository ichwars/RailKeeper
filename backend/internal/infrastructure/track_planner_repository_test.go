package infrastructure_test

import (
	"database/sql"
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

	maxGradePercent := 3.5
	layout, err := layouts.CreateLayout(ctx, application.CreateLayoutInput{
		Name: "Clubanlage", Kind: domain.LayoutKindClub, Gauge: "TT", Scale: "1:120",
		MaxGradePercent: &maxGradePercent,
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
		ElevationStartMM: -3, ElevationEndMM: 1.15,
	}, "planner")
	if err != nil {
		t.Fatal(err)
	}
	if created.Version != 1 || created.RotationDegrees != 345 || created.Geometry.ArticleNumber != "83101" ||
		created.LineageID != created.ID || created.ElevationStartMM != -3 || created.ElevationEndMM != 1.15 {
		t.Fatalf("unexpected created track object: %#v", created)
	}

	updated, err := planner.UpdateObject(ctx, created.ID, application.UpdatePlanTrackObjectInput{
		PositionXMM: 110, PositionYMM: 55, RotationDegrees: 15, ElevationStartMM: 4,
		ElevationEndMM: 8.15, ExpectedVersion: created.Version,
	}, "planner")
	if err != nil {
		t.Fatal(err)
	}
	if updated.Version != 2 || updated.PositionXMM != 110 || updated.RotationDegrees != 15 ||
		updated.ElevationStartMM != 4 || updated.ElevationEndMM != 8.15 {
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
	configuration, err := layouts.SaveConfiguration(ctx, layout.ID, application.SaveLayoutConfigurationInput{
		Name:  "Ausstellung",
		Units: []application.ConfigurationUnitInput{{UnitID: unit.ID, PlanRevisionID: published.ID}},
	}, "planner")
	if err != nil {
		t.Fatal(err)
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
		clonedPlan.Objects[0].LineageID != updated.LineageID ||
		clonedPlan.Objects[0].PositionXMM != updated.PositionXMM || clonedPlan.Objects[0].Version != 1 ||
		clonedPlan.Objects[0].ElevationStartMM != 4 || clonedPlan.Objects[0].ElevationEndMM != 8.15 {
		t.Fatalf("unexpected cloned track plan: %#v", clonedPlan)
	}
	if clonedPlan.Limits.MaxGradePercent == nil || *clonedPlan.Limits.MaxGradePercent != maxGradePercent {
		t.Fatalf("layout grade limit missing from track plan: %#v", clonedPlan.Limits)
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
	preview, err := planner.ChangePreview(ctx, clone.ID)
	if err != nil {
		t.Fatal(err)
	}
	if preview.BaseRevisionID != published.ID || len(preview.ObjectChanges) != 1 ||
		preview.ObjectChanges[0].Type != domain.TrackPlanObjectRemoved ||
		len(preview.MaterialDeltas) != 1 || preview.MaterialDeltas[0].Delta != -1 ||
		len(preview.AffectedConfigurations) != 1 ||
		preview.AffectedConfigurations[0].ID != configuration.ID {
		t.Fatalf("unexpected persisted change preview: %#v", preview)
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

func TestTrackPlannerSnapPersistsAuthoritativePose(t *testing.T) {
	db := openTrackPlannerSchemaDB(t)
	seedTrackPlanRevision(t, db)
	planner := application.NewTrackPlannerService(infrastructure.NewTrackPlannerRepository(db))
	geometryID := "tillig-tt-modellgleis-83101-v1"
	if _, err := planner.CreateObject(t.Context(), "revision-track-1", application.CreatePlanTrackObjectInput{
		GeometryID: geometryID, PositionXMM: 0, PositionYMM: 0,
	}, "planner"); err != nil {
		t.Fatal(err)
	}
	moving, err := planner.CreateObject(t.Context(), "revision-track-1", application.CreatePlanTrackObjectInput{
		GeometryID: geometryID, PositionXMM: 172, PositionYMM: 2, RotationDegrees: 2,
	}, "planner")
	if err != nil {
		t.Fatal(err)
	}

	updated, err := planner.UpdateObject(t.Context(), moving.ID, application.UpdatePlanTrackObjectInput{
		PositionXMM: 172, PositionYMM: 2, RotationDegrees: 2, ExpectedVersion: moving.Version,
	}, "planner")
	if err != nil {
		t.Fatal(err)
	}
	if updated.PositionXMM != 166 || updated.PositionYMM != 0 || updated.RotationDegrees != 0 {
		t.Fatalf("unexpected authoritative snapped pose: %#v", updated)
	}
}

func TestTrackPlannerAnalysisAggregatesLocalMaterialAvailability(t *testing.T) {
	db := openTrackPlannerSchemaDB(t)
	seedTrackPlanRevision(t, db)
	if _, err := db.Exec(`
INSERT INTO storage_locations(id, name, created_at, updated_at)
VALUES('track-location', 'Gleislager', 'now', 'now');
INSERT INTO accessory_products(
  id, inventory_number, manufacturer, article_number, name, category, tracking_mode,
  created_at, updated_at
) VALUES(
  'track-product', 'RK-ART-0083101', 'Tillig', '83101', 'Gleisstück G1', 'track', 'quantity',
  'now', 'now'
);
INSERT INTO accessory_stock(product_id, location_id, quantity, updated_at)
VALUES('track-product', 'track-location', 3, 'now');
INSERT INTO accessory_reservations(
  id, product_id, location_id, quantity, layout_unit_id, status, created_by, created_at, updated_at
) VALUES(
  'track-reservation', 'track-product', 'track-location', 1, 'unit-track-1',
  'active', 'planner', 'now', 'now'
);
INSERT INTO plan_track_objects(
  id, revision_id, geometry_id, position_x_mm, position_y_mm, rotation_degrees,
  version, created_at, updated_at
) VALUES
  ('track-material-1', 'revision-track-1', 'tillig-tt-modellgleis-83101-v1', 0, 0, 0, 1, 'now', 'now'),
  ('track-material-2', 'revision-track-1', 'tillig-tt-modellgleis-83101-v1', 166, 0, 0, 1, 'now', 'now');
`); err != nil {
		t.Fatal(err)
	}

	planner := application.NewTrackPlannerService(infrastructure.NewTrackPlannerRepository(db))
	analysis, err := planner.AnalyzePlan(t.Context(), "revision-track-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(analysis.Materials) != 1 {
		t.Fatalf("unexpected material lines: %#v", analysis.Materials)
	}
	material := analysis.Materials[0]
	if material.Manufacturer != "Tillig" || material.ArticleNumber != "83101" ||
		material.RequiredQuantity != 2 || material.PhysicalQuantity != 3 || material.ReservedQuantity != 1 ||
		material.AvailableQuantity != 2 || material.MissingQuantity != 0 ||
		len(material.ProductIDs) != 1 || material.ProductIDs[0] != "track-product" {
		t.Fatalf("unexpected material availability: %#v", material)
	}
}

func TestTrackPlannerReservesQuantityMaterialsByObject(t *testing.T) {
	db := openTrackPlannerSchemaDB(t)
	seedTrackPlanRevision(t, db)
	seedTrackReservationInventory(t, db, 2)
	planner := application.NewTrackPlannerService(infrastructure.NewTrackPlannerRepository(db))

	batch, err := planner.ReserveMaterials(t.Context(), "revision-track-1",
		application.ReserveTrackPlanMaterialsInput{Confirmed: true, Items: []application.TrackPlanReservationInput{
			{TrackObjectID: "track-material-1", ProductID: "track-product", LocationID: "track-location",
				ExpectedObjectVersion: 1},
			{TrackObjectID: "track-material-2", ProductID: "track-product", LocationID: "track-location",
				ExpectedObjectVersion: 1},
		}}, "planner")
	if err != nil {
		t.Fatal(err)
	}
	if len(batch.Reservations) != 2 || batch.Reservations[0].Reservation.LayoutUnitID != "unit-track-1" {
		t.Fatalf("unexpected plan material reservations: %#v", batch)
	}
	var activeLinks, activeReservations int
	if err := db.QueryRow(`SELECT COUNT(*) FROM plan_track_object_reservations WHERE active=1`).Scan(&activeLinks); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM accessory_reservations WHERE status='active'`).
		Scan(&activeReservations); err != nil {
		t.Fatal(err)
	}
	if activeLinks != 2 || activeReservations != 2 {
		t.Fatalf("unexpected persisted reservations: links=%d reservations=%d", activeLinks, activeReservations)
	}
	analysis, err := planner.AnalyzePlan(t.Context(), "revision-track-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(analysis.Reservations) != 2 {
		t.Fatalf("active plan reservations missing from analysis: %#v", analysis.Reservations)
	}
	if _, err := planner.ReserveMaterials(t.Context(), "revision-track-1",
		application.ReserveTrackPlanMaterialsInput{Confirmed: true, Items: []application.TrackPlanReservationInput{{
			TrackObjectID: "track-material-1", ProductID: "track-product", LocationID: "track-location",
			ExpectedObjectVersion: 1,
		}}}, "planner"); !errors.Is(err, application.ErrTrackPlanConflict) {
		t.Fatalf("expected duplicate plan object conflict, got %v", err)
	}
}

func TestTrackPlannerReservationBatchIsAllOrNothing(t *testing.T) {
	db := openTrackPlannerSchemaDB(t)
	seedTrackPlanRevision(t, db)
	seedTrackReservationInventory(t, db, 1)
	planner := application.NewTrackPlannerService(infrastructure.NewTrackPlannerRepository(db))

	_, err := planner.ReserveMaterials(t.Context(), "revision-track-1",
		application.ReserveTrackPlanMaterialsInput{Confirmed: true, Items: []application.TrackPlanReservationInput{
			{TrackObjectID: "track-material-1", ProductID: "track-product", LocationID: "track-location",
				ExpectedObjectVersion: 1},
			{TrackObjectID: "track-material-2", ProductID: "track-product", LocationID: "track-location",
				ExpectedObjectVersion: 1},
		}}, "planner")
	if !errors.Is(err, application.ErrAccessoryInsufficientStock) {
		t.Fatalf("expected insufficient stock, got %v", err)
	}
	var reservations, links int
	if err := db.QueryRow(`SELECT COUNT(*) FROM accessory_reservations`).Scan(&reservations); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM plan_track_object_reservations`).Scan(&links); err != nil {
		t.Fatal(err)
	}
	if reservations != 0 || links != 0 {
		t.Fatalf("failed batch persisted partial work: reservations=%d links=%d", reservations, links)
	}
}

func TestTrackPlannerReservesIndividualAssetAndReleasesLinkOnCancellation(t *testing.T) {
	db := openTrackPlannerSchemaDB(t)
	seedTrackPlanRevision(t, db)
	seedTrackReservationInventory(t, db, 0)
	if _, err := db.Exec(`
UPDATE accessory_products SET tracking_mode='individual', inventory_strategy='individual'
WHERE id='track-product';
INSERT INTO accessory_assets(
  id, product_id, inventory_number, condition_state, lifecycle_state, storage_location_id,
  created_at, updated_at
) VALUES(
  'track-asset', 'track-product', 'RK-GL-000001', 'ready', 'stored', 'track-location', 'now', 'now'
);`); err != nil {
		t.Fatal(err)
	}
	planner := application.NewTrackPlannerService(infrastructure.NewTrackPlannerRepository(db))
	batch, err := planner.ReserveMaterials(t.Context(), "revision-track-1",
		application.ReserveTrackPlanMaterialsInput{Confirmed: true, Items: []application.TrackPlanReservationInput{{
			TrackObjectID: "track-material-1", ProductID: "track-product", LocationID: "track-location",
			AssetID: "track-asset", ExpectedObjectVersion: 1,
		}}}, "planner")
	if err != nil {
		t.Fatal(err)
	}
	if len(batch.Reservations) != 1 || batch.Reservations[0].Reservation.AssetID != "track-asset" {
		t.Fatalf("unexpected individual reservation: %#v", batch)
	}
	var lifecycle string
	if err := db.QueryRow(`SELECT lifecycle_state FROM accessory_assets WHERE id='track-asset'`).Scan(&lifecycle); err != nil {
		t.Fatal(err)
	}
	if lifecycle != "reserved" {
		t.Fatalf("expected reserved asset lifecycle, got %q", lifecycle)
	}
	allocations := application.NewAccessoryAllocationService(infrastructure.NewAccessoryRepository(db))
	if _, err := allocations.CancelReservation(t.Context(), batch.Reservations[0].Reservation.ID, "planner"); err != nil {
		t.Fatal(err)
	}
	var active int
	if err := db.QueryRow(`
SELECT active FROM plan_track_object_reservations WHERE reservation_id=?`,
		batch.Reservations[0].Reservation.ID).Scan(&active); err != nil {
		t.Fatal(err)
	}
	if active != 0 {
		t.Fatalf("cancelled reservation link remains active: %d", active)
	}
	if err := planner.DeleteObject(t.Context(), "track-material-1", 1, "planner"); !errors.Is(err, application.ErrTrackPlanConflict) {
		t.Fatalf("expected reservation history to protect its plan object, got %v", err)
	}
}

func TestTrackPlannerReservationRejectsStaleObjectAndMismatchedProduct(t *testing.T) {
	db := openTrackPlannerSchemaDB(t)
	seedTrackPlanRevision(t, db)
	seedTrackReservationInventory(t, db, 2)
	if _, err := db.Exec(`
INSERT INTO accessory_products(
  id, inventory_number, manufacturer, article_number, name, category, tracking_mode,
  created_at, updated_at
) VALUES(
  'wrong-product', 'RK-ART-WRONG', 'Tillig', '99999', 'Falsches Gleis', 'track', 'quantity',
  'now', 'now'
);`); err != nil {
		t.Fatal(err)
	}
	planner := application.NewTrackPlannerService(infrastructure.NewTrackPlannerRepository(db))
	for name, item := range map[string]application.TrackPlanReservationInput{
		"stale": {
			TrackObjectID: "track-material-1", ProductID: "track-product", LocationID: "track-location",
			ExpectedObjectVersion: 2,
		},
		"mismatched product": {
			TrackObjectID: "track-material-1", ProductID: "wrong-product", LocationID: "track-location",
			ExpectedObjectVersion: 1,
		},
	} {
		_, err := planner.ReserveMaterials(t.Context(), "revision-track-1",
			application.ReserveTrackPlanMaterialsInput{Confirmed: true,
				Items: []application.TrackPlanReservationInput{item}}, "planner")
		want := application.ErrTrackPlanConflict
		if name == "mismatched product" {
			want = application.ErrTrackPlanValidation
		}
		if !errors.Is(err, want) {
			t.Fatalf("%s: expected %v, got %v", name, want, err)
		}
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM accessory_reservations`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("rejected reservations persisted: %d", count)
	}
}

func seedTrackReservationInventory(t *testing.T, db *sql.DB, quantity int) {
	t.Helper()
	if _, err := db.Exec(`
INSERT INTO storage_locations(id, name, created_at, updated_at)
VALUES('track-location', 'Gleislager', 'now', 'now');
INSERT INTO accessory_products(
  id, inventory_number, manufacturer, article_number, name, category, tracking_mode,
  created_at, updated_at
) VALUES(
  'track-product', 'RK-ART-0083101', 'Tillig', '83101', 'Gleisstück G1', 'track', 'quantity',
  'now', 'now'
);
INSERT INTO accessory_stock(product_id, location_id, quantity, updated_at)
VALUES('track-product', 'track-location', ?, 'now');
INSERT INTO plan_track_objects(
  id, revision_id, geometry_id, position_x_mm, position_y_mm, rotation_degrees,
  version, created_at, updated_at
) VALUES
  ('track-material-1', 'revision-track-1', 'tillig-tt-modellgleis-83101-v1', 0, 0, 0, 1, 'now', 'now'),
  ('track-material-2', 'revision-track-1', 'tillig-tt-modellgleis-83101-v1', 166, 0, 0, 1, 'now', 'now');
`, quantity); err != nil {
		t.Fatal(err)
	}
}
