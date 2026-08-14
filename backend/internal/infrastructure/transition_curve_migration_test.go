package infrastructure_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"railkeeper/backend/internal/infrastructure"
)

func TestTransitionCurveMigrationPreservesExistingFlexPaths(t *testing.T) {
	root := t.TempDir()
	partialMigrations := filepath.Join(root, "migrations-through-0052")
	if err := os.Mkdir(partialMigrations, 0700); err != nil {
		t.Fatal(err)
	}
	copyMigrationsThrough(t, filepath.Join("..", "..", "migrations"), partialMigrations,
		"0052_flex_track_paths.sql")
	db, err := infrastructure.OpenSQLite(filepath.Join(root, "data"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := infrastructure.Migrate(db, partialMigrations); err != nil {
		t.Fatal(err)
	}
	seedTrackPlanRevision(t, db)
	const flexJSON = `{"schemaVersion":1,"endXMm":500,"endYMm":100,"endDirectionDegrees":20,"startHandleMm":180,"endHandleMm":170}`
	if _, err := db.Exec(`
INSERT INTO plan_track_objects(
  id, revision_id, geometry_id, position_x_mm, position_y_mm, rotation_degrees,
  flex_path_json, created_at, updated_at
) VALUES('existing-flex', 'revision-track-1', 'tillig-tt-modellgleis-83125-v1', 10, 20, 0, ?,
  '2026-08-10T10:00:00Z', '2026-08-10T10:00:00Z')`, flexJSON); err != nil {
		t.Fatal(err)
	}

	if err := infrastructure.Migrate(db, filepath.Join("..", "..", "migrations")); err != nil {
		t.Fatal(err)
	}
	var storedFlex string
	var transition sql.NullString
	if err := db.QueryRow(`
SELECT flex_path_json, transition_path_json
FROM plan_track_objects WHERE id='existing-flex'`).Scan(&storedFlex, &transition); err != nil {
		t.Fatal(err)
	}
	if storedFlex != flexJSON || transition.Valid {
		t.Fatalf("existing flex path changed: %q / %#v", storedFlex, transition)
	}
	assertForeignKeyCheck(t, db)
}
