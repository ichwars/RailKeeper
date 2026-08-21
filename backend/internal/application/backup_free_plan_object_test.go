package application_test

import (
	"testing"

	"railkeeper/backend/internal/application"
)

func TestBackupVersionThirteenPreservesFreePlanObjects(t *testing.T) {
	dataDir := t.TempDir()
	db := backupTestDB(t, dataDir)
	ctx := t.Context()
	vehicleID := seedBackupVehicle(t, db)
	seedStageOneBackupData(t, db, vehicleID)
	const shapeJSON = `{"schemaVersion":1,"kind":"label","text":"Bahnhof","fontSizeMm":8}`
	if _, err := db.ExecContext(ctx, `
INSERT INTO plan_free_objects(
  id, lineage_id, revision_id, name, category, position_x_mm, position_y_mm,
  rotation_degrees, shape_json, version, created_at, updated_at
) VALUES('free-label', 'free-lineage', 'revision-2', 'Bahnhof', 'annotation', 120, 80,
  5, ?, 2, '2026-08-10T10:00:00Z', '2026-08-10T10:00:00Z')`, shapeJSON); err != nil {
		t.Fatal(err)
	}

	service := application.NewBackupService(db, dataDir)
	backup, err := service.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if backup.Version != 19 || len(backup.Tables["plan_free_objects"]) != 1 {
		t.Fatalf("expected version 19 free-object export, got version=%d rows=%d",
			backup.Version, len(backup.Tables["plan_free_objects"]))
	}
	if _, err := db.ExecContext(ctx, `DELETE FROM plan_free_objects`); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Import(ctx, backup); err != nil {
		t.Fatal(err)
	}
	var restoredShape string
	var restoredVersion int
	if err := db.QueryRowContext(ctx, `
SELECT shape_json, version FROM plan_free_objects WHERE id='free-label'`).
		Scan(&restoredShape, &restoredVersion); err != nil {
		t.Fatal(err)
	}
	if restoredShape != shapeJSON || restoredVersion != 2 {
		t.Fatalf("free plan object not restored: shape=%q version=%d", restoredShape, restoredVersion)
	}

	legacy := *backup
	legacy.Version = 12
	legacy.Tables = make(map[string][]map[string]any, len(backup.Tables)-1)
	for table, rows := range backup.Tables {
		if table != "plan_free_objects" {
			legacy.Tables[table] = rows
		}
	}
	if _, err := service.Import(ctx, &legacy); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM plan_free_objects`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("version 12 restore must normalize free objects to empty, got %d", count)
	}
}
