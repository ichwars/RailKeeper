package infrastructure_test

import (
	"os"
	"path/filepath"
	"testing"

	"railkeeper/backend/internal/infrastructure"
)

func TestFlexTrackMigrationPreservesExistingPlanObjects(t *testing.T) {
	root := t.TempDir()
	partialMigrations := filepath.Join(root, "migrations-through-0051")
	if err := os.Mkdir(partialMigrations, 0700); err != nil {
		t.Fatal(err)
	}
	copyMigrationsThrough(t, filepath.Join("..", "..", "migrations"), partialMigrations,
		"0051_layout_minimum_track_clearance.sql")
	db, err := infrastructure.OpenSQLite(filepath.Join(root, "data"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := infrastructure.Migrate(db, partialMigrations); err != nil {
		t.Fatal(err)
	}
	seedTrackPlanRevision(t, db)
	if _, err := db.Exec(`
INSERT INTO plan_track_objects(
  id, revision_id, geometry_id, position_x_mm, position_y_mm, rotation_degrees,
  created_at, updated_at
) VALUES(
  'existing-g1', 'revision-track-1', 'tillig-tt-modellgleis-83101-v1', 10, 20, 0,
  '2026-08-10T10:00:00Z', '2026-08-10T10:00:00Z'
)`); err != nil {
		t.Fatal(err)
	}

	if err := infrastructure.Migrate(db, filepath.Join("..", "..", "migrations")); err != nil {
		t.Fatal(err)
	}

	var kind string
	var lengthMM, minimumRadiusMM float64
	if err := db.QueryRow(`
SELECT kind, length_mm, minimum_radius_mm
FROM track_geometry_definitions
WHERE id='tillig-tt-modellgleis-83125-v1'`).
		Scan(&kind, &lengthMM, &minimumRadiusMM); err != nil {
		t.Fatal(err)
	}
	if kind != "flex" || lengthMM != 664 || minimumRadiusMM != 543 {
		t.Fatalf("unexpected flex definition: %q %.2f %.2f", kind, lengthMM, minimumRadiusMM)
	}

	var existingKind string
	if err := db.QueryRow(`
SELECT geometry.kind
FROM plan_track_objects object
JOIN track_geometry_definitions geometry ON geometry.id=object.geometry_id
WHERE object.id='existing-g1'`).Scan(&existingKind); err != nil {
		t.Fatal(err)
	}
	if existingKind != "straight" {
		t.Fatalf("existing G1 changed kind: %q", existingKind)
	}
	assertForeignKeyCheck(t, db)
}
