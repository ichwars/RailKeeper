package infrastructure_test

import (
	"database/sql"
	"encoding/json"
	"path/filepath"
	"testing"

	"railkeeper/backend/internal/domain"
	"railkeeper/backend/internal/infrastructure"
)

func TestTrackPlannerMigrationCreatesVerifiedTilligG1(t *testing.T) {
	db := openTrackPlannerSchemaDB(t)

	for _, table := range []string{
		"track_geometry_libraries",
		"track_geometry_definitions",
		"plan_track_objects",
		"plan_track_object_reservations",
	} {
		var count int
		if err := db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?`, table).
			Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 1 {
			t.Fatalf("missing track planner table %q", table)
		}
	}

	var articleNumber, sourceURL, status, geometryJSON string
	var lengthMM float64
	err := db.QueryRow(`
SELECT article_number, length_mm, source_url, status, geometry_json
FROM track_geometry_definitions
WHERE article_number='83101'`).Scan(&articleNumber, &lengthMM, &sourceURL, &status, &geometryJSON)
	if err != nil {
		t.Fatal(err)
	}
	if articleNumber != "83101" || lengthMM != 166 || status != "verified" ||
		sourceURL != "https://www.tillig.com/Produkte/produktinfo-83101.html" {
		t.Fatalf("unexpected Tillig G1 seed: %q %v %q %q", articleNumber, lengthMM, sourceURL, status)
	}
	var geometry domain.TrackGeometry
	if err := json.Unmarshal([]byte(geometryJSON), &geometry); err != nil {
		t.Fatal(err)
	}
	if geometry.SchemaVersion != 1 || len(geometry.Ports) != 2 ||
		geometry.Ports[0].XMM != 0 || geometry.Ports[1].XMM != 166 {
		t.Fatalf("unexpected Tillig G1 geometry: %#v", geometry)
	}
}

func TestTrackPlannerMigrationEnforcesReferencesAndRotation(t *testing.T) {
	db := openTrackPlannerSchemaDB(t)
	seedTrackPlanRevision(t, db)

	if _, err := db.Exec(`
INSERT INTO plan_track_objects(
  id, revision_id, geometry_id, position_x_mm, position_y_mm, rotation_degrees,
  created_at, updated_at
) VALUES('track-1', 'revision-track-1', 'tillig-tt-modellgleis-83101-v1', 10, 20, 90, 'now', 'now')`); err != nil {
		t.Fatal(err)
	}
	var lineageID string
	if err := db.QueryRow(`SELECT lineage_id FROM plan_track_objects WHERE id='track-1'`).Scan(&lineageID); err != nil {
		t.Fatal(err)
	}
	if lineageID != "track-1" {
		t.Fatalf("expected existing object id as default lineage, got %q", lineageID)
	}
	expectConstraintFailure(t, db, `
INSERT INTO plan_track_objects(
  id, revision_id, geometry_id, position_x_mm, position_y_mm, rotation_degrees,
  lineage_id, created_at, updated_at
) VALUES('duplicate-lineage', 'revision-track-1', 'tillig-tt-modellgleis-83101-v1',
  20, 20, 0, 'track-1', 'now', 'now')`)

	expectConstraintFailure(t, db, `
INSERT INTO plan_track_objects(
  id, revision_id, geometry_id, position_x_mm, position_y_mm, rotation_degrees,
  created_at, updated_at
) VALUES('invalid-rotation', 'revision-track-1', 'tillig-tt-modellgleis-83101-v1', 0, 0, 360, 'now', 'now')`)
	expectConstraintFailure(t, db, `
INSERT INTO plan_track_objects(
  id, revision_id, geometry_id, position_x_mm, position_y_mm, rotation_degrees,
  created_at, updated_at
) VALUES('missing-revision', 'missing', 'tillig-tt-modellgleis-83101-v1', 0, 0, 0, 'now', 'now')`)
	expectConstraintFailure(t, db, `DELETE FROM track_geometry_definitions
WHERE id='tillig-tt-modellgleis-83101-v1'`)
}

func openTrackPlannerSchemaDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := infrastructure.OpenSQLite(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := infrastructure.Migrate(db, filepath.Join("..", "..", "migrations")); err != nil {
		t.Fatal(err)
	}
	return db
}

func seedTrackPlanRevision(t *testing.T, db *sql.DB) {
	t.Helper()
	statements := []string{
		`INSERT INTO layouts(id, name, kind, gauge, scale, created_at, updated_at)
         VALUES('layout-track-1', 'Testanlage', 'private', 'TT', '1:120', 'now', 'now')`,
		`INSERT INTO layout_units(id, layout_id, name, kind, created_at, updated_at)
         VALUES('unit-track-1', 'layout-track-1', 'Bahnhof', 'module', 'now', 'now')`,
		`INSERT INTO plan_variants(id, layout_unit_id, name, created_at, updated_at)
         VALUES('variant-track-1', 'unit-track-1', 'Standard', 'now', 'now')`,
		`INSERT INTO plan_revisions(id, variant_id, revision_number, status, created_by, created_at, updated_at)
         VALUES('revision-track-1', 'variant-track-1', 1, 'draft', 'planner-1', 'now', 'now')`,
	}
	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
}
