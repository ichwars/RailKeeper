package application_test

import (
	"database/sql"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestBackupVersionElevenPreservesFlexTrackFoundation(t *testing.T) {
	dataDir := t.TempDir()
	db := backupTestDB(t, dataDir)
	ctx := t.Context()
	vehicleID := seedBackupVehicle(t, db)
	seedStageOneBackupData(t, db, vehicleID)
	if _, err := db.ExecContext(ctx, `
UPDATE layouts SET minimum_flex_radius_mm=700 WHERE id='layout-1';
INSERT INTO plan_track_objects(
  id, revision_id, geometry_id, position_x_mm, position_y_mm, rotation_degrees,
  flex_path_json, created_at, updated_at
) VALUES(
  'track-flex', 'revision-2', 'tillig-tt-modellgleis-83125-v1', 100, 80, 0,
  '{"schemaVersion":1,"end":{"xMm":500,"yMm":120},"startHandleMm":180,"endHandleMm":180}',
  '2026-08-10T10:00:00Z', '2026-08-10T10:00:00Z'
)`); err != nil {
		t.Fatal(err)
	}

	service := application.NewBackupService(db, dataDir)
	backup, err := service.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if backup.Version != 17 {
		t.Fatalf("expected version 17 export, got %d", backup.Version)
	}
	if _, err := db.ExecContext(ctx, `
UPDATE layouts SET minimum_flex_radius_mm=NULL WHERE id='layout-1';
UPDATE track_geometry_definitions SET minimum_radius_mm=NULL
WHERE id='tillig-tt-modellgleis-83125-v1';
UPDATE plan_track_objects SET flex_path_json=NULL WHERE id='track-flex'`); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Import(ctx, backup); err != nil {
		t.Fatal(err)
	}
	assertFlexTrackFoundation(t, db, true)

	legacy := cloneBackupWithoutFlexColumns(backup)
	if _, err := service.Import(ctx, legacy); err != nil {
		t.Fatal(err)
	}
	assertFlexTrackFoundation(t, db, false)
}

func seedBackupVehicle(t *testing.T, db *sql.DB) string {
	t.Helper()
	vehicle, err := application.NewVehicleService(db).Create(t.Context(), application.CreateVehicleInput{
		Manufacturer: "Tillig", Name: "V 100", Gauge: "TT", Category: "Lokomotive", Gattung: "Diesellok",
	}, "actor-1")
	if err != nil {
		t.Fatal(err)
	}
	return vehicle.ID
}

func cloneBackupWithoutFlexColumns(source *application.BackupDocument) *application.BackupDocument {
	legacy := *source
	legacy.Version = 10
	legacy.Tables = make(map[string][]map[string]any, len(source.Tables))
	for table, rows := range source.Tables {
		legacyRows := make([]map[string]any, 0, len(rows))
		for _, row := range rows {
			legacyRow := make(map[string]any, len(row))
			for key, value := range row {
				if (table == "track_geometry_definitions" && key == "minimum_radius_mm") ||
					(table == "plan_track_objects" && key == "flex_path_json") ||
					(table == "layouts" && key == "minimum_flex_radius_mm") {
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

func assertFlexTrackFoundation(t *testing.T, db *sql.DB, populated bool) {
	t.Helper()
	var layoutRadius, productRadius sql.NullFloat64
	var path sql.NullString
	if err := db.QueryRow(`SELECT minimum_flex_radius_mm FROM layouts WHERE id='layout-1'`).
		Scan(&layoutRadius); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`
SELECT minimum_radius_mm FROM track_geometry_definitions
WHERE id='tillig-tt-modellgleis-83125-v1'`).Scan(&productRadius); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT flex_path_json FROM plan_track_objects WHERE id='track-flex'`).
		Scan(&path); err != nil {
		t.Fatal(err)
	}
	if populated {
		if !layoutRadius.Valid || layoutRadius.Float64 != 700 ||
			!productRadius.Valid || productRadius.Float64 != 543 || !path.Valid {
			t.Fatalf("flex foundation not restored: layout=%#v product=%#v path=%#v",
				layoutRadius, productRadius, path)
		}
		return
	}
	if layoutRadius.Valid || productRadius.Valid || path.Valid {
		t.Fatalf("legacy flex columns not normalized to NULL: layout=%#v product=%#v path=%#v",
			layoutRadius, productRadius, path)
	}
}
