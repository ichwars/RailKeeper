package application_test

import (
	"database/sql"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestBackupVersionTwelvePreservesTransitionCurvePaths(t *testing.T) {
	dataDir := t.TempDir()
	db := backupTestDB(t, dataDir)
	ctx := t.Context()
	vehicleID := seedBackupVehicle(t, db)
	seedStageOneBackupData(t, db, vehicleID)
	const transitionJSON = `{"schemaVersion":1,"lengthMm":500,"endRadiusMm":700,"direction":"left"}`
	if _, err := db.ExecContext(ctx, `
INSERT INTO plan_track_objects(
  id, revision_id, geometry_id, position_x_mm, position_y_mm, rotation_degrees,
  transition_path_json, created_at, updated_at
) VALUES('track-transition', 'revision-2', 'tillig-tt-modellgleis-83125-v1', 100, 80, 0, ?,
  '2026-08-10T10:00:00Z', '2026-08-10T10:00:00Z')`, transitionJSON); err != nil {
		t.Fatal(err)
	}

	service := application.NewBackupService(db, dataDir)
	backup, err := service.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if backup.Version != 14 {
		t.Fatalf("expected version 14 export, got %d", backup.Version)
	}
	if _, err := db.ExecContext(ctx,
		`UPDATE plan_track_objects SET transition_path_json=NULL WHERE id='track-transition'`); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Import(ctx, backup); err != nil {
		t.Fatal(err)
	}
	assertTransitionPath(t, db, transitionJSON)

	legacy := cloneBackupWithoutTransitionPath(backup)
	if _, err := service.Import(ctx, legacy); err != nil {
		t.Fatal(err)
	}
	assertTransitionPath(t, db, "")
}

func cloneBackupWithoutTransitionPath(source *application.BackupDocument) *application.BackupDocument {
	legacy := *source
	legacy.Version = 11
	legacy.Tables = make(map[string][]map[string]any, len(source.Tables))
	for table, rows := range source.Tables {
		legacyRows := make([]map[string]any, 0, len(rows))
		for _, row := range rows {
			legacyRow := make(map[string]any, len(row))
			for key, value := range row {
				if table == "plan_track_objects" && key == "transition_path_json" {
					continue
				}
				legacyRow[key] = value
			}
			legacyRows = append(legacyRows, legacyRow)
		}
		legacy.Tables[table] = legacyRows
	}
	return &legacy
}

func assertTransitionPath(t *testing.T, db *sql.DB, expected string) {
	t.Helper()
	var path sql.NullString
	if err := db.QueryRow(`
SELECT transition_path_json FROM plan_track_objects WHERE id='track-transition'`).Scan(&path); err != nil {
		t.Fatal(err)
	}
	if expected == "" && path.Valid {
		t.Fatalf("legacy transition path was not normalized to NULL: %#v", path)
	}
	if expected != "" && (!path.Valid || path.String != expected) {
		t.Fatalf("transition path not restored: %#v", path)
	}
}
