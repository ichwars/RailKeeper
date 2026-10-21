package application_test

import (
	"database/sql"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestBackupVersionFourteenRestoresLegacyTrackGeometrySnapshots(t *testing.T) {
	dataDir := t.TempDir()
	db := backupTestDB(t, dataDir)
	ctx := t.Context()
	vehicleID := seedBackupVehicle(t, db)
	seedStageOneBackupData(t, db, vehicleID)
	if _, err := db.Exec(`INSERT INTO plan_track_objects(id,revision_id,geometry_id,position_x_mm,position_y_mm,rotation_degrees,created_at,updated_at) VALUES('legacy-object','revision-2','tillig-tt-modellgleis-83101-v1',100,80,0,'now','now')`); err != nil {
		t.Fatal(err)
	}

	service := application.NewBackupService(db, dataDir)
	backup, err := service.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if backup.Version != 24 {
		t.Fatalf("expected version 24 export, got %d", backup.Version)
	}
	legacy := cloneBackupWithoutTrackGeometrySnapshots(backup)
	if _, err := service.Import(ctx, legacy); err != nil {
		t.Fatal(err)
	}
	var snapshot sql.NullString
	if err := db.QueryRow(`
SELECT geometry_snapshot_json FROM plan_track_objects WHERE id=?`, "legacy-object").Scan(&snapshot); err != nil {
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
	if !strings.Contains(snapshot.String, `"lengthMm":166`) {
		t.Fatalf("snapshot lost geometry: %s", snapshot.String)
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
