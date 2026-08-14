package infrastructure_test

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"railkeeper/backend/internal/domain"
	"railkeeper/backend/internal/infrastructure"
)

func TestTrackLibrarySnapshotMigrationBackfillsExistingPlanObjects(t *testing.T) {
	root := t.TempDir()
	partialMigrations := filepath.Join(root, "migrations-through-0054")
	if err := os.Mkdir(partialMigrations, 0700); err != nil {
		t.Fatal(err)
	}
	copyMigrationsThrough(t, filepath.Join("..", "..", "migrations"), partialMigrations,
		"0054_free_plan_objects.sql")
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
) VALUES('existing-track', 'revision-track-1', 'tillig-tt-modellgleis-83101-v1', 10, 20, 0,
  '2026-08-10T10:00:00Z', '2026-08-10T10:00:00Z')`); err != nil {
		t.Fatal(err)
	}

	if err := infrastructure.Migrate(db, filepath.Join("..", "..", "migrations")); err != nil {
		t.Fatal(err)
	}
	var snapshot string
	if err := db.QueryRow(`
SELECT geometry_snapshot_json FROM plan_track_objects WHERE id='existing-track'`).Scan(&snapshot); err != nil {
		t.Fatal(err)
	}
	var geometry domain.TrackGeometryDefinition
	if err := json.Unmarshal([]byte(snapshot), &geometry); err != nil {
		t.Fatal(err)
	}
	if geometry.ID != "tillig-tt-modellgleis-83101-v1" || geometry.ArticleNumber != "83101" ||
		geometry.Name != "Gleisstück G1" || len(geometry.Geometry.Ports) != 2 {
		t.Fatalf("unexpected backfilled snapshot: %#v", geometry)
	}
	var note string
	var verifiedAt, verifiedBy sql.NullString
	if err := db.QueryRow(`
SELECT verification_note, verified_at, verified_by
FROM track_geometry_libraries WHERE id='tillig-tt-modellgleis-v1'`).
		Scan(&note, &verifiedAt, &verifiedBy); err != nil {
		t.Fatal(err)
	}
	if note != "" || verifiedAt.Valid || verifiedBy.Valid {
		t.Fatalf("unexpected verification metadata: %q %#v %#v", note, verifiedAt, verifiedBy)
	}
	assertForeignKeyCheck(t, db)
}
