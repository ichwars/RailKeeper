package infrastructure_test

import (
	"os"
	"path/filepath"
	"testing"

	"railkeeper/backend/internal/infrastructure"
)

func TestFreePlanObjectMigrationAddsIsolatedRevisionTable(t *testing.T) {
	root := t.TempDir()
	partialMigrations := filepath.Join(root, "migrations-through-0053")
	if err := os.Mkdir(partialMigrations, 0700); err != nil {
		t.Fatal(err)
	}
	copyMigrationsThrough(t, filepath.Join("..", "..", "migrations"), partialMigrations,
		"0053_transition_curve_paths.sql")
	db, err := infrastructure.OpenSQLite(filepath.Join(root, "data"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := infrastructure.Migrate(db, partialMigrations); err != nil {
		t.Fatal(err)
	}
	seedTrackPlanRevision(t, db)
	if err := infrastructure.Migrate(db, filepath.Join("..", "..", "migrations")); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO plan_free_objects(
id, lineage_id, revision_id, name, category, position_x_mm, position_y_mm, rotation_degrees,
shape_json, version, created_at, updated_at
) VALUES('free-migration', 'free-migration', 'revision-track-1', 'Text', 'annotation', 1, 2, 0,
'{"schemaVersion":1,"kind":"label","text":"Test","fontSizeMm":8}', 1, 'now', 'now')`); err != nil {
		t.Fatal(err)
	}
	assertForeignKeyCheck(t, db)
}
