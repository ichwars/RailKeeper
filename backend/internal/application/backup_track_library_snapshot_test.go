package application_test

import (
	"database/sql"
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/infrastructure"
)

func TestBackupVersionFourteenRestoresLegacyTrackGeometrySnapshots(t *testing.T) {
	dataDir := t.TempDir()
	db := backupTestDB(t, dataDir)
	ctx := t.Context()
	vehicleID := seedBackupVehicle(t, db)
	seedStageOneBackupData(t, db, vehicleID)
	planner := application.NewTrackPlannerService(infrastructure.NewTrackPlannerRepository(db))
	object, err := planner.CreateObject(ctx, "revision-2", application.CreatePlanTrackObjectInput{
		GeometryID: "tillig-tt-modellgleis-83101-v1", PositionXMM: 100, PositionYMM: 80,
	}, "planner-1")
	if err != nil {
		t.Fatal(err)
	}

	service := application.NewBackupService(db, dataDir)
	backup, err := service.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if backup.Version != 20 {
		t.Fatalf("expected version 20 export, got %d", backup.Version)
	}
	legacy := cloneBackupWithoutTrackGeometrySnapshots(backup)
	if _, err := service.Import(ctx, legacy); err != nil {
		t.Fatal(err)
	}
	var snapshot sql.NullString
	if err := db.QueryRow(`
SELECT geometry_snapshot_json FROM plan_track_objects WHERE id=?`, object.ID).Scan(&snapshot); err != nil {
		t.Fatal(err)
	}
	if !snapshot.Valid || snapshot.String == "" {
		t.Fatal("legacy restore did not backfill track geometry snapshot")
	}
	if _, err := db.Exec(`
UPDATE track_geometry_definitions SET name='Changed', length_mm=999
WHERE id='tillig-tt-modellgleis-83101-v1'`); err != nil {
		t.Fatal(err)
	}
	restored, err := planner.GetPlan(ctx, "revision-2")
	if err != nil {
		t.Fatal(err)
	}
	if len(restored.Objects) != 1 || restored.Objects[0].Geometry.Name != "Gleisstück G1" ||
		restored.Objects[0].Geometry.LengthMM != 166 {
		t.Fatalf("restored plan did not retain snapshot: %#v", restored.Objects)
	}
}

func cloneBackupWithoutTrackGeometrySnapshots(source *application.BackupDocument) *application.BackupDocument {
	legacy := *source
	legacy.Version = 13
	legacy.Tables = make(map[string][]map[string]any, len(source.Tables))
	for table, rows := range source.Tables {
		legacyRows := make([]map[string]any, 0, len(rows))
		for _, row := range rows {
			legacyRow := make(map[string]any, len(row))
			for key, value := range row {
				if table == "plan_track_objects" && key == "geometry_snapshot_json" {
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
