package application_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/domain"
	"railkeeper/backend/internal/infrastructure"
)

func TestBackupExportsAndRestoresAppDataAndUploads(t *testing.T) {
	dataDir := t.TempDir()
	db := backupTestDB(t, dataDir)
	ctx := context.Background()

	vehicles := application.NewVehicleService(db)
	exhibitions := application.NewExhibitionService(db)
	created, err := vehicles.Create(ctx, application.CreateVehicleInput{
		Manufacturer: "Piko",
		Name:         "BR 118",
		Gauge:        "H0",
		Category:     "Lokomotive",
		Gattung:      "Diesellok",
	}, "actor-1")
	if err != nil {
		t.Fatal(err)
	}
	_, err = vehicles.CreateAttachment(ctx, created.ID, application.VehicleAttachmentInput{
		FileName:     "manual.pdf",
		OriginalName: "manual.pdf",
		MimeType:     "application/pdf",
		SizeBytes:    6,
		StoragePath:  "uploads/vehicles/" + created.ID + "/manual.pdf",
	})
	if err != nil {
		t.Fatal(err)
	}
	uploadPath := filepath.Join(dataDir, "uploads", "vehicles", created.ID, "manual.pdf")
	if err := os.MkdirAll(filepath.Dir(uploadPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(uploadPath, []byte("manual"), 0o600); err != nil {
		t.Fatal(err)
	}
	list, err := exhibitions.Create(ctx, application.ExhibitionListInput{
		Designation: "Leipzig 2026",
		Date:        "2026-05-12",
	})
	if err != nil {
		t.Fatal(err)
	}
	entry, err := exhibitions.CreateEntry(ctx, list.ID, application.ExhibitionEntryInput{
		Owner:          "Daniel",
		LocomotiveName: "V180",
		DTDecoder:      true,
		DecoderNumber:  "1001",
		FunctionKeys:   `[{"key":"F0","name":"Licht","type":"licht","symbolKey":"esu-f006-spitzensignal"}]`,
	})
	if err != nil {
		t.Fatal(err)
	}

	backupService := application.NewBackupService(db, dataDir)
	backup, err := backupService.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if backup.Version != 18 {
		t.Fatalf("expected backup version 18, got %d", backup.Version)
	}
	if len(backup.Tables["vehicles"]) != 1 {
		t.Fatalf("expected one vehicle in backup, got %d", len(backup.Tables["vehicles"]))
	}
	if len(backup.Files) != 1 {
		t.Fatalf("expected one file in backup, got %d", len(backup.Files))
	}
	if len(backup.Tables["exhibition_lists"]) != 1 || len(backup.Tables["exhibition_entries"]) != 1 {
		t.Fatalf("expected exhibition data in backup, got lists=%d entries=%d", len(backup.Tables["exhibition_lists"]), len(backup.Tables["exhibition_entries"]))
	}
	validation, err := backupService.Validate(ctx, backup)
	if err != nil {
		t.Fatal(err)
	}
	if !validation.Compatible || validation.RowCount == 0 || validation.FileCount != 1 {
		t.Fatalf("expected backup to validate, got %#v", validation)
	}

	if _, err := db.Exec(`DELETE FROM vehicle_attachments`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`DELETE FROM exhibition_entries`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`DELETE FROM exhibition_lists`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`DELETE FROM vehicles`); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(filepath.Join(dataDir, "uploads")); err != nil {
		t.Fatal(err)
	}

	result, err := backupService.Import(ctx, backup)
	if err != nil {
		t.Fatal(err)
	}
	if result.RestoredRows == 0 || result.RestoredFiles != 1 {
		t.Fatalf("unexpected restore result: %#v", result)
	}

	restored, err := vehicles.Get(ctx, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if restored.InventoryNumber != created.InventoryNumber || len(restored.Attachments) != 1 {
		t.Fatalf("unexpected restored vehicle: %#v", restored)
	}
	if _, err := os.Stat(uploadPath); err != nil {
		t.Fatalf("expected upload file restored: %v", err)
	}
	restoredList, err := exhibitions.Get(ctx, list.ID)
	if err != nil {
		t.Fatal(err)
	}
	if restoredList.Designation != list.Designation || len(restoredList.Entries) != 1 || restoredList.Entries[0].ID != entry.ID {
		t.Fatalf("unexpected restored exhibition list: %#v", restoredList)
	}
}

func TestBackupVersionEighteenPreservesVehicleSetInventoryNumber(t *testing.T) {
	dataDir := t.TempDir()
	db := backupTestDB(t, dataDir)
	ctx := context.Background()
	vehicles := application.NewVehicleService(db)
	created, err := vehicles.CreateSet(ctx, application.CreateVehicleSetInput{
		Set: application.VehicleSetInput{
			Name: "Rheingold", Manufacturer: "Roco", Gauge: "H0", Category: "Wagen",
			Gattung: "Reisezugwagen",
		},
		Members: []application.CreateVehicleInput{{Name: "Wagen 1"}, {Name: "Wagen 2"}},
	}, "actor-1")
	if err != nil {
		t.Fatal(err)
	}

	backup, err := application.NewBackupService(db, dataDir).Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if backup.Version != 18 {
		t.Fatalf("expected backup version 18, got %d", backup.Version)
	}
	if got := backup.Tables["vehicle_sets"][0]["inventory_number"]; got != created.InventoryNumber {
		t.Fatalf("set inventory number missing from backup: %v", got)
	}
}

func TestLegacyVersionSeventeenVehicleSetNumbersAreNormalizedDeterministically(t *testing.T) {
	dataDir := t.TempDir()
	db := backupTestDB(t, dataDir)
	ctx := context.Background()
	service := application.NewBackupService(db, dataDir)
	doc, err := service.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	doc.Version = 17
	legacyRows := []map[string]any{
		{
			"id": "set-b", "name": "Später", "manufacturer": "Roco", "gauge": "H0",
			"category": "Wagen", "gattung": "Reisezugwagen",
			"created_at": "2026-08-02T00:00:00Z", "updated_at": "2026-08-02T00:00:00Z",
		},
		{
			"id": "set-a", "name": "Früher", "manufacturer": "Roco", "gauge": "H0",
			"category": "Wagen", "gattung": "Reisezugwagen",
			"created_at": "2026-08-01T00:00:00Z", "updated_at": "2026-08-01T00:00:00Z",
		},
	}
	doc.Tables["vehicle_sets"] = legacyRows
	doc.Tables["vehicle_set_members"] = []map[string]any{}
	setSchemes := make([]map[string]any, 0, len(doc.Tables["inventory_number_schemes"]))
	for _, row := range doc.Tables["inventory_number_schemes"] {
		if row["category"] != "Set" {
			setSchemes = append(setSchemes, row)
		}
	}
	doc.Tables["inventory_number_schemes"] = setSchemes

	if _, err := service.Import(ctx, doc); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		id   string
		want string
	}{{"set-a", "RK-SET-000001"}, {"set-b", "RK-SET-000002"}} {
		var got string
		if err := db.QueryRow(`SELECT inventory_number FROM vehicle_sets WHERE id=?`, test.id).Scan(&got); err != nil {
			t.Fatal(err)
		}
		if got != test.want {
			t.Fatalf("set %s inventory number=%q, want %q", test.id, got, test.want)
		}
	}
	var nextNumber int
	if err := db.QueryRow(`SELECT next_number FROM inventory_number_schemes WHERE category='Set'`).Scan(&nextNumber); err != nil {
		t.Fatal(err)
	}
	if nextNumber != 3 {
		t.Fatalf("set scheme next number=%d, want 3", nextNumber)
	}
	for index, row := range legacyRows {
		if _, mutated := row["inventory_number"]; mutated {
			t.Fatalf("legacy input row %d was mutated: %#v", index, row)
		}
	}
}

func TestLegacyVersionSeventeenVehicleSetNumbersActivateExistingInactiveScheme(t *testing.T) {
	dataDir := t.TempDir()
	db := backupTestDB(t, dataDir)
	ctx := context.Background()
	service := application.NewBackupService(db, dataDir)
	doc, err := service.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	doc.Version = 17
	doc.Tables["vehicle_sets"] = []map[string]any{{
		"id": "set-a", "name": "Altset", "manufacturer": "Roco", "gauge": "H0",
		"category": "Wagen", "gattung": "Reisezugwagen",
		"created_at": "2026-08-01T00:00:00Z", "updated_at": "2026-08-01T00:00:00Z",
	}}
	doc.Tables["vehicle_set_members"] = []map[string]any{}
	for _, row := range doc.Tables["inventory_number_schemes"] {
		if row["category"] == "Set" {
			row["active"] = float64(0)
			row["prefix"] = "CLUB-SET"
			row["next_number"] = float64(7)
		}
	}

	if _, err := service.Import(ctx, doc); err != nil {
		t.Fatal(err)
	}
	var inventoryNumber string
	var active int
	if err := db.QueryRow(`SELECT inventory_number FROM vehicle_sets WHERE id='set-a'`).Scan(&inventoryNumber); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT active FROM inventory_number_schemes WHERE category='Set'`).Scan(&active); err != nil {
		t.Fatal(err)
	}
	if inventoryNumber != "CLUB-SET-000007" || active != 1 {
		t.Fatalf("restored inventory number=%q active=%d", inventoryNumber, active)
	}
}

func TestLegacyVersionSeventeenWithoutVehicleSetsRestoresSetNumberScheme(t *testing.T) {
	dataDir := t.TempDir()
	db := backupTestDB(t, dataDir)
	ctx := context.Background()
	service := application.NewBackupService(db, dataDir)
	doc, err := service.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	doc.Version = 17
	doc.Tables["vehicle_sets"] = []map[string]any{}
	doc.Tables["vehicle_set_members"] = []map[string]any{}
	setSchemes := make([]map[string]any, 0, len(doc.Tables["inventory_number_schemes"]))
	for _, row := range doc.Tables["inventory_number_schemes"] {
		if row["category"] != "Set" {
			setSchemes = append(setSchemes, row)
		}
	}
	doc.Tables["inventory_number_schemes"] = setSchemes

	if _, err := service.Import(ctx, doc); err != nil {
		t.Fatal(err)
	}
	var prefix string
	var active int
	if err := db.QueryRow(`
SELECT prefix, active FROM inventory_number_schemes WHERE category='Set'
`).Scan(&prefix, &active); err != nil {
		t.Fatal(err)
	}
	if prefix != "RK-SET" || active != 1 {
		t.Fatalf("restored Set scheme prefix=%q active=%d", prefix, active)
	}
}

func TestBackupRestorePreservesInactiveBundledAndCustomOrigins(t *testing.T) {
	dataDir := t.TempDir()
	db := backupTestDB(t, dataDir)
	if _, err := db.Exec(`
UPDATE master_data_entries
SET origin='bundled', active=0, label='Gleismaterial'
WHERE type='article_type' AND key='track'`); err != nil {
		t.Fatal(err)
	}
	masterData := application.NewMasterDataService(db)
	if _, err := masterData.Create(t.Context(), "manufacturer", application.MasterDataInput{
		Key: "club", Label: "Club",
	}); err != nil {
		t.Fatal(err)
	}
	service := application.NewBackupService(db, dataDir)
	doc, err := service.Export(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if doc.Version != 18 {
		t.Fatalf("version=%d", doc.Version)
	}
	if _, err := service.Import(t.Context(), doc); err != nil {
		t.Fatal(err)
	}
	assertMasterDataOrigin(t, db, "article_type", "track", "bundled", false)
	assertMasterDataOrigin(t, db, "manufacturer", "club", "custom", true)
}

func TestBackupRestoreVersion15ReconcilesCurrentBundledKeys(t *testing.T) {
	dataDir := t.TempDir()
	db := backupTestDB(t, dataDir)
	if _, err := db.Exec(`
UPDATE master_data_entries SET origin='bundled', active=0
WHERE type='article_type' AND key='track'`); err != nil {
		t.Fatal(err)
	}
	service := application.NewBackupService(db, dataDir)
	doc, err := service.Export(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	doc.Version = 15
	for _, row := range doc.Tables["master_data_entries"] {
		delete(row, "origin")
	}
	if _, err := service.Import(t.Context(), doc); err != nil {
		t.Fatal(err)
	}
	assertMasterDataOrigin(t, db, "article_type", "track", "bundled", false)
}

func TestBackupRestoreDoesNotTrustUnknownBundledOrigin(t *testing.T) {
	dataDir := t.TempDir()
	db := backupTestDB(t, dataDir)
	service := application.NewBackupService(db, dataDir)
	doc, err := service.Export(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	doc.Tables["master_data_entries"] = append(doc.Tables["master_data_entries"], map[string]any{
		"id": "manufacturer:unknown", "type": "manufacturer", "key": "unknown",
		"label": "Unknown", "active": 1, "sort_order": 0, "source_url": "",
		"metadata_json": "{}", "created_at": "now", "updated_at": "now", "origin": "bundled",
	})
	if _, err := service.Import(t.Context(), doc); err != nil {
		t.Fatal(err)
	}
	assertMasterDataOrigin(t, db, "manufacturer", "unknown", "custom", true)
}

func assertMasterDataOrigin(
	t *testing.T,
	db *sql.DB,
	typeName, key, wantOrigin string,
	wantActive bool,
) {
	t.Helper()
	var origin string
	var active int
	if err := db.QueryRow(`
SELECT origin, active FROM master_data_entries WHERE type=? AND key=?`, typeName, key).Scan(
		&origin,
		&active,
	); err != nil {
		t.Fatal(err)
	}
	if origin != wantOrigin || (active == 1) != wantActive {
		t.Fatalf("master data %s/%s origin=%q active=%d", typeName, key, origin, active)
	}
}

func TestBackupOperationalFieldsAndListPriceRemainCompatible(t *testing.T) {
	ctx := context.Background()
	sourceDir := t.TempDir()
	sourceDB := backupTestDB(t, sourceDir)
	speed := 120
	vehicle, err := application.NewVehicleService(sourceDB).Create(ctx, application.CreateVehicleInput{
		Manufacturer: "Piko", Name: "BR 118", Gauge: "H0", Category: "Lokomotive", Gattung: "Diesellok",
		MaximumSpeedKmh: &speed, HomeBase: "Bw Leipzig",
	}, "actor-1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := sourceDB.Exec(`INSERT INTO accessory_products(
  id, inventory_number, manufacturer, name, category, tracking_mode, article_type, subtype,
  gauges_json, package_quantity, stock_unit, minimum_stock, inventory_strategy, list_price,
  created_at, updated_at
) VALUES(
  'product-valued', 'RK-ART-BACKUP-000001', 'Tillig', 'Gleis', 'other', 'quantity',
  'other', 'other:other', '[]', 1, 'piece', 0, 'quantity', '129,90', 'now', 'now'
)`); err != nil {
		t.Fatal(err)
	}

	document, err := application.NewBackupService(sourceDB, sourceDir).Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if document.Version != 18 {
		t.Fatalf("expected backup version 18 for operational fields and accessory list price, got %d",
			document.Version)
	}
	targetDir := t.TempDir()
	targetDB := backupTestDB(t, targetDir)
	if _, err := application.NewBackupService(targetDB, targetDir).Import(ctx, document); err != nil {
		t.Fatal(err)
	}
	assertRestoredOperationalValues(t, targetDB, vehicle.ID, &speed, "Bw Leipzig", "129,90")

	encoded, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	legacy := &application.BackupDocument{}
	if err := json.Unmarshal(encoded, legacy); err != nil {
		t.Fatal(err)
	}
	for _, row := range legacy.Tables["vehicles"] {
		delete(row, "maximum_speed_kmh")
		delete(row, "home_base")
	}
	for _, row := range legacy.Tables["accessory_products"] {
		delete(row, "list_price")
	}
	legacy.Version = 14
	legacyDir := t.TempDir()
	legacyDB := backupTestDB(t, legacyDir)
	if _, err := application.NewBackupService(legacyDB, legacyDir).Import(ctx, legacy); err != nil {
		t.Fatalf("restore older backup without new columns: %v", err)
	}
	assertRestoredOperationalValues(t, legacyDB, vehicle.ID, nil, "", "")
}

func assertRestoredOperationalValues(
	t *testing.T,
	db *sql.DB,
	vehicleID string,
	wantSpeed *int,
	wantHomeBase string,
	wantListPrice string,
) {
	t.Helper()
	vehicle, err := application.NewVehicleService(db).Get(t.Context(), vehicleID)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(vehicle.MaximumSpeedKmh, wantSpeed) || vehicle.HomeBase != wantHomeBase {
		t.Fatalf("unexpected restored operational fields: speed=%v home=%q",
			vehicle.MaximumSpeedKmh, vehicle.HomeBase)
	}
	var listPrice string
	if err := db.QueryRow(`SELECT list_price FROM accessory_products WHERE id='product-valued'`).Scan(&listPrice); err != nil {
		t.Fatal(err)
	}
	if listPrice != wantListPrice {
		t.Fatalf("restored list price = %q, want %q", listPrice, wantListPrice)
	}
}

func TestBackupVersionFourRoundTripPreservesStageOneDataReferences(t *testing.T) {
	dataDir := t.TempDir()
	db := backupTestDB(t, dataDir)
	ctx := context.Background()
	vehicle, err := application.NewVehicleService(db).Create(ctx, application.CreateVehicleInput{
		Manufacturer: "Tillig", Name: "V 100", Gauge: "TT", Category: "Lokomotive", Gattung: "Diesellok",
	}, "actor-1")
	if err != nil {
		t.Fatal(err)
	}
	seedStageOneBackupData(t, db, vehicle.ID)

	service := application.NewBackupService(db, dataDir)
	backup, err := service.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if backup.Version != 18 {
		t.Fatalf("expected version 18 export, got %d", backup.Version)
	}
	for _, table := range stageOneBackupTableNames() {
		if len(backup.Tables[table]) == 0 {
			t.Fatalf("expected stage-one table %q in backup", table)
		}
	}
	if _, err := db.ExecContext(ctx, `UPDATE layouts SET name = 'Changed after export' WHERE id = 'layout-1'`); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Import(ctx, backup); err != nil {
		t.Fatal(err)
	}

	var layoutName, baseRevisionID, configurationRevisionID, reservationVehicleID, installationUnitID string
	if err := db.QueryRowContext(ctx, `SELECT name FROM layouts WHERE id = 'layout-1'`).Scan(&layoutName); err != nil {
		t.Fatal(err)
	}
	if layoutName != "Clubanlage Falkenstein" {
		t.Fatalf("expected exported layout name after restore, got %q", layoutName)
	}
	if err := db.QueryRowContext(ctx, `SELECT base_revision_id FROM plan_revisions WHERE id = 'revision-2'`).Scan(&baseRevisionID); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT plan_revision_id FROM layout_configuration_units WHERE configuration_id = 'configuration-1'`).Scan(&configurationRevisionID); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT vehicle_id FROM accessory_reservations WHERE id = 'reservation-1'`).Scan(&reservationVehicleID); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT layout_unit_id FROM accessory_installations WHERE id = 'installation-1'`).Scan(&installationUnitID); err != nil {
		t.Fatal(err)
	}
	if baseRevisionID != "revision-1" || configurationRevisionID != "revision-1" ||
		reservationVehicleID != vehicle.ID || installationUnitID != "unit-1" {
		t.Fatalf("stage-one references changed after restore: base=%q configuration=%q reservation=%q installation=%q",
			baseRevisionID, configurationRevisionID, reservationVehicleID, installationUnitID)
	}
	rows, err := db.QueryContext(ctx, `PRAGMA foreign_key_check`)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	if rows.Next() {
		t.Fatal("expected no foreign-key violations after version 2 restore")
	}
}

func TestBackupVersionFourRoundTripPreservesLayoutTwinPositions(t *testing.T) {
	dataDir := t.TempDir()
	db := backupTestDB(t, dataDir)
	ctx := t.Context()
	vehicle, err := application.NewVehicleService(db).Create(ctx, application.CreateVehicleInput{
		Manufacturer: "Tillig", Name: "V 100", Gauge: "TT", Category: "Lokomotive", Gattung: "Diesellok",
	}, "actor-1")
	if err != nil {
		t.Fatal(err)
	}
	seedStageOneBackupData(t, db, vehicle.ID)
	if _, err := db.ExecContext(ctx, `
INSERT INTO layout_unit_outline_points(layout_unit_id, point_index, position_x_mm, position_y_mm) VALUES
  ('unit-1', 0, 0, 0), ('unit-1', 1, 1200, 0), ('unit-1', 2, 1200, 500), ('unit-1', 3, 0, 500);
INSERT INTO layout_technical_positions(
  id, layout_unit_id, label, kind, position_x_mm, position_y_mm, rotation_degrees,
  product_id, description, version, archived, created_at, updated_at
) VALUES (
  'position-1', 'unit-1', 'Signal A', 'signal', 250, 80, 90,
  'product-quantity', 'Einfahrsignal', 1, 0, '2026-08-09T10:00:00Z', '2026-08-09T10:00:00Z'
);
INSERT INTO accessory_reservation_positions(reservation_id, position_id)
  VALUES ('reservation-1', 'position-1');
INSERT INTO accessory_installation_positions(installation_id, position_id)
  VALUES ('installation-1', 'position-1');
`); err != nil {
		t.Fatal(err)
	}

	service := application.NewBackupService(db, dataDir)
	backup, err := service.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if backup.Version != 18 {
		t.Fatalf("expected version 18 export, got %d", backup.Version)
	}
	for _, table := range []string{
		"layout_unit_outline_points", "layout_technical_positions",
		"accessory_reservation_positions", "accessory_installation_positions",
	} {
		if len(backup.Tables[table]) == 0 {
			t.Fatalf("expected layout-twin table %q in backup", table)
		}
	}
	if _, err := db.ExecContext(ctx, `UPDATE layout_technical_positions SET label='Changed' WHERE id='position-1'`); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Import(ctx, backup); err != nil {
		t.Fatal(err)
	}

	var label, reservationPositionID, installationPositionID string
	if err := db.QueryRowContext(ctx, `SELECT label FROM layout_technical_positions WHERE id='position-1'`).Scan(&label); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `
SELECT position_id FROM accessory_reservation_positions WHERE reservation_id='reservation-1'
`).Scan(&reservationPositionID); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `
SELECT position_id FROM accessory_installation_positions WHERE installation_id='installation-1'
`).Scan(&installationPositionID); err != nil {
		t.Fatal(err)
	}
	if label != "Signal A" || reservationPositionID != "position-1" || installationPositionID != "position-1" {
		t.Fatalf("layout-twin references changed: label=%q reservation=%q installation=%q",
			label, reservationPositionID, installationPositionID)
	}
}

func TestBackupVersionSixRoundTripPreservesTrackPlannerData(t *testing.T) {
	dataDir := t.TempDir()
	db := backupTestDB(t, dataDir)
	ctx := t.Context()
	vehicle, err := application.NewVehicleService(db).Create(ctx, application.CreateVehicleInput{
		Manufacturer: "Tillig", Name: "V 100", Gauge: "TT", Category: "Lokomotive", Gattung: "Diesellok",
	}, "actor-1")
	if err != nil {
		t.Fatal(err)
	}
	seedStageOneBackupData(t, db, vehicle.ID)
	if _, err := db.ExecContext(ctx, `
INSERT INTO plan_track_objects(
  id, revision_id, geometry_id, position_x_mm, position_y_mm, rotation_degrees,
	elevation_start_mm, elevation_end_mm, version, created_at, updated_at
) VALUES
	('track-published', 'revision-1', 'tillig-tt-modellgleis-83101-v1', 125.5, 80, 0, 0, 0,
	 1, '2026-08-09T10:00:00Z', '2026-08-09T10:00:00Z'),
	('track-draft', 'revision-2', 'tillig-tt-modellgleis-83101-v1', 291.5, 80, 15, -2, 2.15,
	 3, '2026-08-09T10:00:00Z', '2026-08-09T11:00:00Z');
INSERT INTO accessory_reservations(
  id, product_id, location_id, quantity, layout_unit_id, status, note, created_by, created_at, updated_at
) VALUES(
  'track-plan-reservation', 'product-quantity', 'location-child', 1, 'unit-1', 'active',
  'Gleisplanobjekt track-draft', 'planner-1', '2026-08-09T11:00:00Z', '2026-08-09T11:00:00Z'
);
INSERT INTO plan_track_object_reservations(
  reservation_id, track_object_id, active, created_at, updated_at
) VALUES(
  'track-plan-reservation', 'track-draft', 1,
  '2026-08-09T11:00:00Z', '2026-08-09T11:00:00Z'
)`); err != nil {
		t.Fatal(err)
	}

	service := application.NewBackupService(db, dataDir)
	backup, err := service.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if backup.Version != 18 {
		t.Fatalf("expected version 18 export, got %d", backup.Version)
	}
	for _, table := range versionFiveBackupTableNames() {
		if len(backup.Tables[table]) == 0 {
			t.Fatalf("expected track-planner table %q in backup", table)
		}
	}
	for _, table := range versionSixBackupTableNames() {
		if len(backup.Tables[table]) == 0 {
			t.Fatalf("expected track-reservation table %q in backup", table)
		}
	}
	if _, err := db.ExecContext(ctx, `
UPDATE track_geometry_libraries SET manufacturer='Changed';
UPDATE track_geometry_definitions SET name='Changed', geometry_json='{}';
UPDATE plan_track_objects SET position_x_mm=999, rotation_degrees=90,
       elevation_start_mm=99, elevation_end_mm=100;
UPDATE plan_track_object_reservations SET active=0;
`); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Import(ctx, backup); err != nil {
		t.Fatal(err)
	}

	var manufacturer, name, geometryJSON, revisionID, geometryID, lineageID string
	var positionX, positionY, rotation, elevationStart, elevationEnd float64
	var version, reservationActive int
	if err := db.QueryRowContext(ctx, `
SELECT l.manufacturer, g.name, g.geometry_json, o.revision_id, o.geometry_id,
	   o.position_x_mm, o.position_y_mm, o.rotation_degrees,
	   o.elevation_start_mm, o.elevation_end_mm, o.version, o.lineage_id
FROM plan_track_objects o
JOIN track_geometry_definitions g ON g.id=o.geometry_id
JOIN track_geometry_libraries l ON l.id=g.library_id
WHERE o.id='track-draft'
`).Scan(&manufacturer, &name, &geometryJSON, &revisionID, &geometryID,
		&positionX, &positionY, &rotation, &elevationStart, &elevationEnd, &version, &lineageID); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `
SELECT active FROM plan_track_object_reservations WHERE reservation_id='track-plan-reservation'
`).Scan(&reservationActive); err != nil {
		t.Fatal(err)
	}
	if manufacturer != "Tillig" || name != "Gleisstück G1" ||
		!strings.Contains(geometryJSON, `"xMm":166`) || revisionID != "revision-2" ||
		geometryID != "tillig-tt-modellgleis-83101-v1" || positionX != 291.5 || positionY != 80 ||
		rotation != 15 || elevationStart != -2 || elevationEnd != 2.15 || version != 3 ||
		lineageID != "track-draft" || reservationActive != 1 {
		t.Fatalf("track-planner data changed after restore: manufacturer=%q name=%q geometry=%q "+
			"revision=%q geometryID=%q position=%v/%v rotation=%v elevation=%v/%v version=%d",
			manufacturer, name, geometryJSON, revisionID, geometryID, positionX, positionY, rotation,
			elevationStart, elevationEnd, version)
	}

	legacy := *backup
	legacy.Version = 7
	legacy.Tables = make(map[string][]map[string]any, len(backup.Tables))
	for table, rows := range backup.Tables {
		legacy.Tables[table] = rows
	}
	legacyTrackRows := make([]map[string]any, 0, len(backup.Tables["plan_track_objects"]))
	for _, row := range backup.Tables["plan_track_objects"] {
		legacyRow := make(map[string]any, len(row))
		for key, value := range row {
			if key != "elevation_start_mm" && key != "elevation_end_mm" {
				legacyRow[key] = value
			}
		}
		legacyTrackRows = append(legacyTrackRows, legacyRow)
	}
	legacy.Tables["plan_track_objects"] = legacyTrackRows
	if _, err := service.Import(ctx, &legacy); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `
SELECT elevation_start_mm, elevation_end_mm FROM plan_track_objects WHERE id='track-draft'
`).Scan(&elevationStart, &elevationEnd); err != nil {
		t.Fatal(err)
	}
	if elevationStart != 0 || elevationEnd != 0 {
		t.Fatalf("expected version 7 restore defaults, got %v/%v", elevationStart, elevationEnd)
	}
}

func TestBackupVersionSevenRoundTripPreservesLayoutUnitPorts(t *testing.T) {
	dataDir := t.TempDir()
	db := backupTestDB(t, dataDir)
	ctx := t.Context()
	if _, err := db.ExecContext(ctx, `
INSERT INTO layouts(id, name, kind, gauge, scale, description, version, archived, created_at, updated_at)
VALUES('layout-port', 'Module layout', 'club', 'TT', '1:120', '', 1, 0,
       '2026-08-10T10:00:00Z', '2026-08-10T10:00:00Z');
INSERT INTO layout_units(id, layout_id, name, kind, owner_label, width_mm, height_mm, version, archived, created_at, updated_at)
VALUES('unit-port', 'layout-port', 'Module A', 'module', '', 1000, 500, 1, 0,
       '2026-08-10T10:00:00Z', '2026-08-10T10:00:00Z');
INSERT INTO layout_unit_ports(
  id, layout_unit_id, name, kind, interface_key, x_mm, y_mm, direction_degrees,
  notes, version, archived, created_at, updated_at
) VALUES(
  'port-west', 'unit-port', 'West', 'track', 'track:tillig-tt-modellgleis', 0, 250, 180,
  'Main line', 2, 0, '2026-08-10T10:00:00Z', '2026-08-10T11:00:00Z'
);`); err != nil {
		t.Fatal(err)
	}

	service := application.NewBackupService(db, dataDir)
	backup, err := service.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if backup.Version != 18 || len(backup.Tables["layout_unit_ports"]) != 1 {
		t.Fatalf("expected version 18 module-port export, got version=%d rows=%d",
			backup.Version, len(backup.Tables["layout_unit_ports"]))
	}
	if _, err := db.ExecContext(ctx, `UPDATE layout_unit_ports SET name='Changed', x_mm=100`); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Import(ctx, backup); err != nil {
		t.Fatal(err)
	}
	var name, unitID, interfaceKey string
	var xMM, direction float64
	if err := db.QueryRowContext(ctx, `
SELECT name, layout_unit_id, interface_key, x_mm, direction_degrees
FROM layout_unit_ports WHERE id='port-west'`).Scan(&name, &unitID, &interfaceKey, &xMM, &direction); err != nil {
		t.Fatal(err)
	}
	if name != "West" || unitID != "unit-port" || interfaceKey != "track:tillig-tt-modellgleis" ||
		xMM != 0 || direction != 180 {
		t.Fatalf("module port changed after restore: name=%q unit=%q interface=%q x=%v direction=%v",
			name, unitID, interfaceKey, xMM, direction)
	}

	legacyTables := make(map[string][]map[string]any, len(backup.Tables)-1)
	for table, rows := range backup.Tables {
		if table != "layout_unit_ports" {
			legacyTables[table] = rows
		}
	}
	legacy := &application.BackupDocument{
		Format: backup.Format, Version: 6, CreatedAt: backup.CreatedAt,
		Tables: legacyTables, Files: backup.Files,
	}
	validation, err := service.Validate(ctx, legacy)
	if err != nil || !validation.Compatible {
		t.Fatalf("expected version 6 without module ports to remain compatible: %#v, %v", validation, err)
	}
}

func TestBackupVersionNinePreservesAndLegacyVersionEightOmitsLayoutGradeLimit(t *testing.T) {
	dataDir := t.TempDir()
	db := backupTestDB(t, dataDir)
	ctx := t.Context()
	if _, err := db.ExecContext(ctx, `
INSERT INTO layouts(id, name, kind, gauge, scale, description, max_grade_percent,
                    version, archived, created_at, updated_at)
VALUES('layout-grade', 'Steigungsanlage', 'private', 'TT', '1:120', '', 3.5,
       1, 0, '2026-08-10T10:00:00Z', '2026-08-10T10:00:00Z')`); err != nil {
		t.Fatal(err)
	}

	service := application.NewBackupService(db, dataDir)
	backup, err := service.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if backup.Version != 18 {
		t.Fatalf("expected version 18 export, got %d", backup.Version)
	}
	if _, err := db.ExecContext(ctx, `UPDATE layouts SET max_grade_percent=NULL WHERE id='layout-grade'`); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Import(ctx, backup); err != nil {
		t.Fatal(err)
	}
	var restored sql.NullFloat64
	if err := db.QueryRowContext(ctx, `SELECT max_grade_percent FROM layouts WHERE id='layout-grade'`).
		Scan(&restored); err != nil {
		t.Fatal(err)
	}
	if !restored.Valid || restored.Float64 != 3.5 {
		t.Fatalf("expected restored 3.5 percent limit, got %#v", restored)
	}

	legacy := *backup
	legacy.Version = 8
	legacy.Tables = make(map[string][]map[string]any, len(backup.Tables))
	for table, rows := range backup.Tables {
		legacyRows := make([]map[string]any, 0, len(rows))
		for _, row := range rows {
			legacyRow := make(map[string]any, len(row))
			for key, value := range row {
				if table != "layouts" || key != "max_grade_percent" {
					legacyRow[key] = value
				}
			}
			legacyRows = append(legacyRows, legacyRow)
		}
		legacy.Tables[table] = legacyRows
	}
	if _, err := service.Import(ctx, &legacy); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT max_grade_percent FROM layouts WHERE id='layout-grade'`).
		Scan(&restored); err != nil {
		t.Fatal(err)
	}
	if restored.Valid {
		t.Fatalf("expected legacy version 8 restore without limit, got %#v", restored)
	}
}

func TestBackupVersionTenPreservesAndLegacyVersionNineOmitsTrackClearance(t *testing.T) {
	dataDir := t.TempDir()
	db := backupTestDB(t, dataDir)
	ctx := t.Context()
	if _, err := db.ExecContext(ctx, `
INSERT INTO layouts(id, name, kind, gauge, scale, description, max_grade_percent,
                    minimum_track_clearance_mm, version, archived, created_at, updated_at)
VALUES('layout-clearance', 'Abstandsanlage', 'private', 'TT', '1:120', '', NULL,
       40, 1, 0, '2026-08-10T10:00:00Z', '2026-08-10T10:00:00Z')`); err != nil {
		t.Fatal(err)
	}

	service := application.NewBackupService(db, dataDir)
	backup, err := service.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if backup.Version != 18 {
		t.Fatalf("expected version 18 export, got %d", backup.Version)
	}
	if _, err := db.ExecContext(ctx, `
UPDATE layouts SET minimum_track_clearance_mm=NULL WHERE id='layout-clearance'`); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Import(ctx, backup); err != nil {
		t.Fatal(err)
	}
	var restored sql.NullFloat64
	if err := db.QueryRowContext(ctx, `
SELECT minimum_track_clearance_mm FROM layouts WHERE id='layout-clearance'`).Scan(&restored); err != nil {
		t.Fatal(err)
	}
	if !restored.Valid || restored.Float64 != 40 {
		t.Fatalf("expected restored 40 mm limit, got %#v", restored)
	}

	legacy := *backup
	legacy.Version = 9
	legacy.Tables = make(map[string][]map[string]any, len(backup.Tables))
	for table, rows := range backup.Tables {
		legacyRows := make([]map[string]any, 0, len(rows))
		for _, row := range rows {
			legacyRow := make(map[string]any, len(row))
			for key, value := range row {
				if table != "layouts" || key != "minimum_track_clearance_mm" {
					legacyRow[key] = value
				}
			}
			legacyRows = append(legacyRows, legacyRow)
		}
		legacy.Tables[table] = legacyRows
	}
	if _, err := service.Import(ctx, &legacy); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `
SELECT minimum_track_clearance_mm FROM layouts WHERE id='layout-clearance'`).Scan(&restored); err != nil {
		t.Fatal(err)
	}
	if restored.Valid {
		t.Fatalf("expected legacy version 9 restore without limit, got %#v", restored)
	}
}

func TestBackupVersionTwoRestoreBackfillsIndividualInventoryStrategy(t *testing.T) {
	dataDir := t.TempDir()
	db := backupTestDB(t, dataDir)
	ctx := t.Context()
	tables := backupDocumentTablesThroughVersion(2)
	tables["storage_locations"] = []map[string]any{{
		"id": "legacy-location", "name": "Legacy shelf", "description": "", "archived": 0,
		"created_at": "2026-08-01T00:00:00Z", "updated_at": "2026-08-01T00:00:00Z",
	}}
	tables["accessory_products"] = []map[string]any{
		{
			"id": "legacy-quantity", "manufacturer": "Tillig", "article_number": "83101",
			"name": "Legacy track", "category": "track", "tracking_mode": "quantity", "description": "",
			"created_at": "2026-08-01T00:00:00Z", "updated_at": "2026-08-01T00:00:00Z",
		},
		{
			"id": "legacy-individual", "manufacturer": "ESU", "article_number": "59610",
			"name": "Legacy decoder", "category": "decoder", "tracking_mode": "individual", "description": "",
			"created_at": "2026-08-01T00:00:00Z", "updated_at": "2026-08-01T00:00:00Z",
		},
		{
			"id": "interim-hybrid", "manufacturer": "ESU", "article_number": "51830",
			"name": "Interim hybrid", "category": "decoder", "tracking_mode": "quantity", "description": "",
			"article_type": "decoder", "subtype": "decoder:accessory", "package_quantity": 1,
			"stock_unit": "piece", "inventory_strategy": "quantity_later_individual",
			"created_at": "2026-08-01T00:00:00Z", "updated_at": "2026-08-01T00:00:00Z",
		},
	}
	tables["accessory_stock"] = []map[string]any{{
		"product_id": "legacy-quantity", "location_id": "legacy-location", "quantity": 4,
		"updated_at": "2026-08-01T00:00:00Z",
	}}
	tables["accessory_assets"] = []map[string]any{{
		"id": "legacy-asset", "product_id": "legacy-individual", "inventory_number": "LEGACY-1",
		"serial_number": "SER-1", "condition_state": "ready", "lifecycle_state": "stored",
		"storage_location_id": "legacy-location", "purchase_date": "", "purchase_price": "",
		"warranty_until": "", "notes": "", "created_at": "2026-08-01T00:00:00Z",
		"updated_at": "2026-08-01T00:00:00Z",
	}}

	service := application.NewBackupService(db, dataDir)
	if _, err := service.Import(ctx, &application.BackupDocument{
		Format: "railkeeper-backup", Version: 2, Tables: tables,
	}); err != nil {
		t.Fatal(err)
	}

	repository := infrastructure.NewAccessoryRepository(db)
	accessories := application.NewAccessoryService(repository)
	allocations := application.NewAccessoryAllocationService(repository)
	individual, err := accessories.GetProduct(ctx, "legacy-individual")
	if err != nil {
		t.Fatal(err)
	}
	quantity, err := accessories.GetProduct(ctx, "legacy-quantity")
	if err != nil {
		t.Fatal(err)
	}
	if individual.InventoryStrategy != domain.AccessoryInventoryIndividual ||
		quantity.InventoryStrategy != domain.AccessoryInventoryQuantity {
		t.Fatalf("legacy strategies not backfilled: individual=%q quantity=%q",
			individual.InventoryStrategy, quantity.InventoryStrategy)
	}
	if individual.InventoryNumber == "" || quantity.InventoryNumber == "" ||
		individual.InventoryNumber == quantity.InventoryNumber {
		t.Fatalf("legacy articles did not receive distinct inventory numbers: individual=%q quantity=%q",
			individual.InventoryNumber, quantity.InventoryNumber)
	}
	interim, err := accessories.GetProduct(ctx, "interim-hybrid")
	if err != nil {
		t.Fatal(err)
	}
	if interim.InventoryStrategy != domain.AccessoryInventoryQuantityLaterIndividual ||
		interim.ArticleType != domain.AccessoryArticleDecoder || interim.Subtype != "decoder:accessory" {
		t.Fatalf("restore overwrote explicit interim v2 article fields: %#v", interim)
	}
	assets, err := accessories.ListAssets(ctx, individual.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(assets) != 1 || assets[0].ID != "legacy-asset" {
		t.Fatalf("legacy assets not accessible after restore: %#v", assets)
	}
	summary, err := allocations.GetAllocationSummary(ctx, individual.ID)
	if err != nil {
		t.Fatal(err)
	}
	if summary.Owned != 1 || summary.Stored != 1 || summary.Available != 1 {
		t.Fatalf("unexpected restored individual totals: %#v", summary)
	}
	articles, err := accessories.ListArticles(ctx, application.AccessoryArticleListQuery{})
	if err != nil {
		t.Fatal(err)
	}
	foundIndividual := false
	for _, item := range articles.Items {
		if item.ID == individual.ID {
			foundIndividual = true
			if item.Owned != 1 || item.Available != 1 {
				t.Fatalf("unexpected restored individual article totals: %#v", item)
			}
		}
	}
	if !foundIndividual {
		t.Fatalf("restored individual article missing from list: %#v", articles.Items)
	}
}

func TestBackupVersionFourRoundTripPreservesArticleManagementData(t *testing.T) {
	dataDir := t.TempDir()
	db := backupTestDB(t, dataDir)
	ctx := context.Background()
	repository := infrastructure.NewAccessoryRepository(db)
	accessories := application.NewAccessoryService(repository)
	allocations := application.NewAccessoryAllocationService(repository)
	blobs := application.NewFileBlobService(db, dataDir)
	documents := application.NewAccessoryDocumentService(repository, blobs)
	layouts := application.NewLayoutService(infrastructure.NewLayoutRepository(db))

	location, err := accessories.CreateLocation(ctx, application.CreateStorageLocationInput{
		Name: "Main article store", Description: "Dry and locked",
	}, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	length := 166.0
	unit := "mm"
	product, err := accessories.CreateProduct(ctx, application.CreateAccessoryProductInput{
		Manufacturer: "Tillig", ArticleNumber: "83101", Name: "Straight track", Category: "straight",
		Description: "Bedding track", EAN: "4012500831012", ManufacturerStatus: "available",
		ArticleType: domain.AccessoryArticleTrack, Subtype: "straight", Gauges: []string{"TT"}, Scale: "1:120",
		PackageQuantity: 6, StockUnit: "piece", MinimumStock: 4,
		InventoryStrategy: domain.AccessoryInventoryQuantity,
		ManufacturerURL:   "https://example.test/tillig", ProductURL: "https://example.test/83101",
		AlternativeNumbers: []string{"83101-A"}, Keywords: []string{"track", "straight"},
		CompatibilityNotes: "Code 83", InternalNotes: "Club standard",
		Attributes: []domain.AccessoryAttributeValue{{
			Key: "lengthMm", Kind: domain.AccessoryAttributeNumber, NumberValue: &length, Unit: &unit,
		}},
	}, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	purchase, err := accessories.CreatePurchase(ctx, product.ID, application.CreateAccessoryPurchaseInput{
		PurchasedAt: "2026-08-08", Supplier: "Model shop", Quantity: 5, UnitPrice: "3.25", Currency: "EUR",
		InvoiceNumber: "INV-300", WarrantyUntil: "2028-08-08", StorageLocationID: location.ID,
		BookToStock: true, Notes: "Stage one stock",
	}, "buyer-1")
	if err != nil {
		t.Fatal(err)
	}

	individualProduct, err := accessories.CreateProduct(ctx, application.CreateAccessoryProductInput{
		Manufacturer: "ESU", ArticleNumber: "51820", Name: "SwitchPilot", Category: "decoder",
		ArticleType: domain.AccessoryArticleOther, Subtype: "other", PackageQuantity: 1,
		StockUnit: "piece", InventoryStrategy: domain.AccessoryInventoryIndividual,
	}, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	individualPurchase, err := accessories.CreatePurchase(ctx, individualProduct.ID,
		application.CreateAccessoryPurchaseInput{
			PurchasedAt: "2026-08-07", Supplier: "Decoder shop", Quantity: 1, UnitPrice: "79.90", Currency: "EUR",
			InvoiceNumber: "DEC-1", StorageLocationID: location.ID, BookToStock: true,
		}, "buyer-1")
	if err != nil {
		t.Fatal(err)
	}

	manualBytes := []byte("RailKeeper article manual bytes")
	blobID, err := blobs.Store(ctx, manualBytes)
	if err != nil {
		t.Fatal(err)
	}
	document, err := documents.CreateDocument(ctx, application.CreateAccessoryDocumentInput{
		ProductID: product.ID, FileBlobID: blobID,
		AccessoryDocumentUploadMetadata: application.AccessoryDocumentUploadMetadata{
			FileName: "manual.pdf", OriginalName: "Tillig 83101.pdf", Category: application.AccessoryDocumentManual,
			MimeType: "application/pdf", SizeBytes: int64(len(manualBytes)),
		},
		Description: "Official manual",
	}, 1024, "editor-1")
	if err != nil {
		t.Fatal(err)
	}

	layout, err := layouts.CreateLayout(ctx, application.CreateLayoutInput{
		Name: "Club layout", Kind: domain.LayoutKindClub, Gauge: "TT", Scale: "1:120",
	}, "planner-1")
	if err != nil {
		t.Fatal(err)
	}
	unitRecord, err := layouts.CreateUnit(ctx, layout.ID, application.CreateLayoutUnitInput{
		Name: "Station module", Kind: domain.LayoutUnitKindModule,
	}, "planner-1")
	if err != nil {
		t.Fatal(err)
	}
	reservation, err := allocations.CreateReservation(ctx, application.CreateAccessoryReservationInput{
		ProductID: product.ID, LocationID: location.ID, Quantity: 1,
		AllocationTargetInput: application.AllocationTargetInput{LayoutUnitID: unitRecord.ID},
		Placement:             "Track 2", DigitalAddress: "17", DecoderOutput: "A", Connection: "J", WiringNotes: "blue wire",
		Note: "Autumn exhibition",
	}, "planner-1")
	if err != nil {
		t.Fatal(err)
	}
	installation, err := allocations.Install(ctx, application.CreateAccessoryInstallationInput{
		ReservationID: reservation.ID, ProductID: product.ID, SourceLocationID: location.ID, Quantity: 1,
		AllocationTargetInput: application.AllocationTargetInput{LayoutUnitID: unitRecord.ID},
		Placement:             "Track 2", DigitalAddress: "17", DecoderOutput: "A", Connection: "J", WiringNotes: "blue wire",
		Condition: domain.AccessoryConditionReady, Notes: "Installed for acceptance",
	}, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := allocations.UpdateInstallationCondition(ctx, installation.ID,
		application.UpdateAccessoryInstallationConditionInput{Condition: domain.AccessoryConditionDefective},
		"maintainer-1"); err != nil {
		t.Fatal(err)
	}
	removed, err := allocations.RemoveInstallation(ctx, installation.ID,
		application.RemoveAccessoryInstallationInput{
			Disposition: domain.AccessoryRemovalStored, StorageLocationID: location.ID, Notes: "Needs repair",
		}, "editor-2")
	if err != nil {
		t.Fatal(err)
	}
	if removed.RemovedAt == "" || removed.RemovalDisposition != domain.AccessoryRemovalStored {
		t.Fatalf("expected completed removal before export, got %#v", removed)
	}

	service := application.NewBackupService(db, dataDir)
	backup, err := service.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if backup.Version != 18 {
		t.Fatalf("expected version 18 export, got %d", backup.Version)
	}
	for _, table := range versionThreeBackupTableNames() {
		if len(backup.Tables[table]) == 0 {
			t.Fatalf("expected article data table %q in backup", table)
		}
	}
	encoded, err := json.Marshal(backup)
	if err != nil {
		t.Fatal(err)
	}
	backup, err = application.DecodeBackup(encoded)
	if err != nil {
		t.Fatal(err)
	}
	validation, err := service.Validate(ctx, backup)
	if err != nil {
		t.Fatal(err)
	}
	if !validation.Compatible {
		t.Fatalf("expected complete version 8 backup, got %#v", validation)
	}
	if _, err := db.ExecContext(ctx, `UPDATE accessory_products SET name='Changed after export' WHERE id=?`, product.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `UPDATE file_blobs SET data=x'00' WHERE id=?`, blobID); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Import(ctx, backup); err != nil {
		t.Fatal(err)
	}

	restoredProduct, err := accessories.GetProduct(ctx, product.ID)
	if err != nil {
		t.Fatal(err)
	}
	if restoredProduct.Name != "Straight track" || restoredProduct.EAN != "4012500831012" ||
		restoredProduct.ManufacturerStatus != "available" || restoredProduct.Scale != "1:120" ||
		restoredProduct.PackageQuantity != 6 || restoredProduct.MinimumStock != 4 ||
		len(restoredProduct.Attributes) != 1 || restoredProduct.Attributes[0].NumberValue == nil ||
		*restoredProduct.Attributes[0].NumberValue != length || restoredProduct.Attributes[0].Unit == nil ||
		*restoredProduct.Attributes[0].Unit != unit {
		t.Fatalf("article fields or typed attribute changed after restore: %#v", restoredProduct)
	}
	restoredPurchases, err := accessories.ListPurchases(ctx, product.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(restoredPurchases) != 1 || restoredPurchases[0].ID != purchase.ID ||
		restoredPurchases[0].InvoiceNumber != "INV-300" || !restoredPurchases[0].BookToStock {
		t.Fatalf("purchase changed after restore: %#v", restoredPurchases)
	}
	stock, err := accessories.GetStock(ctx, product.ID)
	if err != nil {
		t.Fatal(err)
	}
	movements, err := accessories.ListStockMovements(ctx, product.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stock.TotalQuantity != 5 || len(movements) != 3 {
		t.Fatalf("quantity stock or movement history changed: stock=%#v movements=%#v", stock, movements)
	}
	assets, err := accessories.ListAssets(ctx, individualProduct.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(assets) != 1 || assets[0].PurchaseID != individualPurchase.ID || assets[0].StorageLocationID != location.ID {
		t.Fatalf("individual asset changed after restore: %#v", assets)
	}
	restoredDocuments, err := documents.ListDocuments(ctx, product.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(restoredDocuments) != 1 || restoredDocuments[0].ID != document.ID ||
		restoredDocuments[0].Description != "Official manual" || restoredDocuments[0].FileBlobID != blobID {
		t.Fatalf("document metadata changed after restore: %#v", restoredDocuments)
	}
	if restoredBytes, err := blobs.Load(ctx, blobID); err != nil || string(restoredBytes) != string(manualBytes) {
		t.Fatalf("document blob changed after restore: %q, %v", restoredBytes, err)
	}
	restoredReservations, err := allocations.ListReservations(ctx, product.ID)
	if err != nil {
		t.Fatal(err)
	}
	restoredInstallations, err := allocations.ListInstallations(ctx, product.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(restoredReservations) != 1 || restoredReservations[0].ID != reservation.ID ||
		restoredReservations[0].Placement != "Track 2" || len(restoredInstallations) != 1 ||
		restoredInstallations[0].ID != installation.ID || restoredInstallations[0].RemovedAt == "" ||
		restoredInstallations[0].Condition != domain.AccessoryConditionDefective ||
		restoredInstallations[0].RemovalDisposition != domain.AccessoryRemovalStored ||
		restoredInstallations[0].RemovalNotes != "Needs repair" {
		t.Fatalf("allocation history changed: reservations=%#v installations=%#v",
			restoredReservations, restoredInstallations)
	}
	history, err := allocations.GetUsageHistory(ctx, product.ID)
	if err != nil {
		t.Fatal(err)
	}
	eventTypes := map[application.AccessoryUsageEventType]bool{}
	for _, event := range history.Events {
		eventTypes[event.Type] = true
	}
	for _, eventType := range []application.AccessoryUsageEventType{
		application.AccessoryUsageReservation, application.AccessoryUsageInstallation,
		application.AccessoryUsageConditionChanged, application.AccessoryUsageRemoval,
	} {
		if !eventTypes[eventType] {
			t.Fatalf("usage history lacks %q after restore: %#v", eventType, history.Events)
		}
	}
}

func TestBackupExcludesAuthenticationTables(t *testing.T) {
	dataDir := t.TempDir()
	db := backupTestDB(t, dataDir)
	ctx := context.Background()
	setup := application.NewSetupService(db)
	auth := application.NewAuthService(db)
	if err := setup.CreateAdmin(ctx, application.CreateAdminInput{
		Username: "admin",
		Email:    "admin@example.test",
		Password: "very-secure-password",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := auth.Login(ctx, application.LoginInput{
		Username: "admin",
		Password: "very-secure-password",
	}); err != nil {
		t.Fatal(err)
	}

	backupService := application.NewBackupService(db, dataDir)
	backup, err := backupService.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, table := range []string{
		"users", "roles", "user_roles", "sessions", "password_reset_requests", "rate_limit_attempts", "audit_logs",
	} {
		if _, ok := backup.Tables[table]; ok {
			t.Fatalf("backup should not export authentication table %q", table)
		}
	}
	data, err := json.Marshal(backup)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "password_hash") {
		t.Fatal("backup should not contain password hashes")
	}
}

func TestBackupExcludesLocalSettingsAndCredentials(t *testing.T) {
	dataDir := t.TempDir()
	db := backupTestDB(t, dataDir)
	ctx := context.Background()
	setup := application.NewSetupService(db)
	if err := setup.CreateAdmin(ctx, application.CreateAdminInput{
		Username: "admin",
		Email:    "admin@example.test",
		Password: "very-secure-password",
	}); err != nil {
		t.Fatal(err)
	}
	var userID string
	if err := db.QueryRowContext(ctx, `SELECT id FROM users WHERE username='admin'`).Scan(&userID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `
INSERT INTO app_settings(key, value, updated_at)
VALUES
  ('smtp.password', 'secret-smtp-password', '2026-06-21T00:00:00Z'),
  ('digital.ecos.host', '192.168.178.44', '2026-06-21T00:00:00Z')
`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `
INSERT INTO user_settings(user_id, key, value, updated_at)
VALUES(?, 'sidebar.order', '["settings","vehicles"]', '2026-06-21T00:00:00Z')
`, userID); err != nil {
		t.Fatal(err)
	}

	backupService := application.NewBackupService(db, dataDir)
	backup, err := backupService.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, table := range []string{"app_settings", "user_settings"} {
		if _, ok := backup.Tables[table]; ok {
			t.Fatalf("backup should not export local settings table %q", table)
		}
	}
	data, err := json.Marshal(backup)
	if err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{"smtp.password", "secret-smtp-password", "digital.ecos.host", "192.168.178.44", "sidebar.order"} {
		if strings.Contains(string(data), secret) {
			t.Fatalf("backup should not contain local setting %q", secret)
		}
	}
}

func TestBackupCoversAllApplicationDataTables(t *testing.T) {
	dataDir := t.TempDir()
	db := backupTestDB(t, dataDir)
	backupService := application.NewBackupService(db, dataDir)
	backup, err := backupService.Export(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	rows, err := db.Query(`
SELECT name
FROM sqlite_master
WHERE type='table'
  AND name NOT LIKE 'sqlite_%'
ORDER BY name
`)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()

	excluded := map[string]bool{
		"app_settings":            true,
		"audit_logs":              true,
		"data_transfer_artifacts": true,
		"data_transfer_job_issues": true,
		"data_transfer_jobs":      true,
		"data_transfer_profiles":  true,
		"password_reset_requests": true,
		"rate_limit_attempts":     true,
		"roles":                   true,
		"schema_migrations":       true,
		"sessions":                true,
		"user_roles":              true,
		"user_settings":           true,
		"users":                   true,
	}
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			t.Fatal(err)
		}
		if excluded[table] {
			continue
		}
		if _, ok := backup.Tables[table]; !ok {
			t.Fatalf("application data table %q is missing from backup export", table)
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
}

func TestBackupValidationWarnsAboutIgnoredAuthenticationTables(t *testing.T) {
	db := backupTestDB(t, t.TempDir())
	service := application.NewBackupService(db, t.TempDir())
	doc := &application.BackupDocument{
		Format:  "railkeeper-backup",
		Version: 1,
		Tables:  map[string][]map[string]any{},
	}
	for _, table := range []string{
		"master_data_entries",
		"master_data_relations",
		"inventory_number_schemes",
		"vehicles",
		"inventory_number_history",
		"vehicle_external_mappings",
		"file_blobs",
		"vehicle_images",
		"vehicle_attachments",
		"vehicle_maintenance",
		"vehicle_spare_parts",
		"vehicle_functions",
		"vehicle_cv_files",
		"vehicle_cv_values",
		"vehicle_cv_value_history",
		"exhibition_lists",
		"exhibition_entries",
	} {
		doc.Tables[table] = []map[string]any{}
	}
	doc.Tables["users"] = []map[string]any{{"id": "user-1", "password_hash": "secret"}}
	doc.Tables["app_settings"] = []map[string]any{{"key": "smtp.password", "value": "secret"}}
	doc.Tables["user_settings"] = []map[string]any{{"user_id": "user-1", "key": "sidebar.order", "value": "[]"}}

	result, err := service.Validate(context.Background(), doc)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Compatible {
		t.Fatalf("expected backup to remain compatible with ignored auth table, got %#v", result)
	}
	if !containsWarning(result.Warnings, "Unbekannte Tabelle users") {
		t.Fatalf("expected ignored users table warning, got %#v", result.Warnings)
	}
	if !containsWarning(result.Warnings, "Unbekannte Tabelle app_settings") || !containsWarning(result.Warnings, "Unbekannte Tabelle user_settings") {
		t.Fatalf("expected ignored local settings table warning, got %#v", result.Warnings)
	}
}

func backupDocumentTablesThroughVersion(version int) map[string][]map[string]any {
	tables := backupDocumentTablesWithout()
	if version >= 2 {
		for _, table := range stageOneBackupTableNames() {
			tables[table] = []map[string]any{}
		}
	}
	if version >= 3 {
		for _, table := range versionThreeBackupTableNames() {
			tables[table] = []map[string]any{}
		}
	}
	if version >= 4 {
		for _, table := range versionFourBackupTableNames() {
			tables[table] = []map[string]any{}
		}
	}
	if version >= 5 {
		for _, table := range versionFiveBackupTableNames() {
			tables[table] = []map[string]any{}
		}
	}
	if version >= 6 {
		for _, table := range versionSixBackupTableNames() {
			tables[table] = []map[string]any{}
		}
	}
	if version >= 7 {
		for _, table := range versionSevenBackupTableNames() {
			tables[table] = []map[string]any{}
		}
	}
	return tables
}

func TestBackupLegacyOptionalTablesRemainCompatibleUntilVersionThree(t *testing.T) {
	legacyOptionalTables := legacyOptionalBackupTableNames()
	for _, test := range []struct {
		name           string
		version        int
		wantCompatible bool
	}{
		{name: "version one", version: 1, wantCompatible: true},
		{name: "version two", version: 2, wantCompatible: true},
		{name: "version three", version: 3, wantCompatible: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			db := backupTestDB(t, t.TempDir())
			service := application.NewBackupService(db, t.TempDir())
			tables := backupDocumentTablesThroughVersion(test.version)
			for _, table := range legacyOptionalTables {
				delete(tables, table)
			}
			doc := &application.BackupDocument{
				Format: "railkeeper-backup", Version: test.version, Tables: tables,
			}

			result, err := service.Validate(context.Background(), doc)
			if err != nil {
				t.Fatal(err)
			}
			if result.Compatible != test.wantCompatible {
				t.Fatalf("compatibility=%t, want %t: %#v", result.Compatible, test.wantCompatible, result)
			}
			for _, table := range legacyOptionalTables {
				messages := result.Warnings
				prefix := "Optionale Tabelle "
				if !test.wantCompatible {
					messages = result.Errors
					prefix = "Tabelle "
				}
				if !containsWarning(messages, prefix+table+" fehlt") {
					t.Fatalf("expected version %d missing-table result for %s: %#v", test.version, table, result)
				}
			}
			if test.wantCompatible {
				if _, err := service.Import(context.Background(), doc); err != nil {
					t.Fatalf("expected version %d legacy backup to import: %v", test.version, err)
				}
			}
		})
	}
}

func TestBackupVersionOneWithoutStageOneTablesRemainsImportable(t *testing.T) {
	db := backupTestDB(t, t.TempDir())
	service := application.NewBackupService(db, t.TempDir())
	doc := &application.BackupDocument{
		Format:  "railkeeper-backup",
		Version: 1,
		Tables:  backupDocumentTablesWithout(),
	}

	result, err := service.Validate(context.Background(), doc)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Compatible {
		t.Fatalf("expected version 1 backup without stage-one tables to remain compatible, got %#v", result)
	}
	for _, table := range stageOneBackupTableNames() {
		if !containsWarning(result.Warnings, "Optionale Tabelle "+table+" fehlt") {
			t.Fatalf("expected compatibility warning for %s, got %#v", table, result.Warnings)
		}
	}
	for _, table := range versionThreeBackupTableNames() {
		if !containsWarning(result.Warnings, "Optionale Tabelle "+table+" fehlt") {
			t.Fatalf("expected version-three compatibility warning for %s, got %#v", table, result.Warnings)
		}
	}
	for _, table := range versionFourBackupTableNames() {
		if !containsWarning(result.Warnings, "Optionale Tabelle "+table+" fehlt") {
			t.Fatalf("expected version-four compatibility warning for %s, got %#v", table, result.Warnings)
		}
	}
	if _, err := service.Import(context.Background(), doc); err != nil {
		t.Fatalf("expected compatible version 1 backup to import, got %v", err)
	}
}

func TestBackupVersionTwoRequiresStageOneTables(t *testing.T) {
	db := backupTestDB(t, t.TempDir())
	service := application.NewBackupService(db, t.TempDir())
	doc := &application.BackupDocument{
		Format:  "railkeeper-backup",
		Version: 2,
		Tables:  backupDocumentTablesWithout(),
	}

	result, err := service.Validate(context.Background(), doc)
	if err != nil {
		t.Fatal(err)
	}
	if result.Compatible {
		t.Fatalf("expected incomplete version 2 backup to be rejected")
	}
	if !containsWarning(result.Errors, "Tabelle storage_locations fehlt") {
		t.Fatalf("expected missing stage-one table error, got %#v", result.Errors)
	}
}

func TestBackupVersionTwoWithoutVersionThreeTablesRemainsCompatible(t *testing.T) {
	db := backupTestDB(t, t.TempDir())
	service := application.NewBackupService(db, t.TempDir())
	doc := &application.BackupDocument{
		Format: "railkeeper-backup", Version: 2, Tables: backupDocumentTablesThroughVersion(2),
	}

	result, err := service.Validate(context.Background(), doc)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Compatible {
		t.Fatalf("expected version 2 without version-three tables to remain compatible, got %#v", result)
	}
	for _, table := range versionThreeBackupTableNames() {
		if !containsWarning(result.Warnings, "Optionale Tabelle "+table+" fehlt") {
			t.Fatalf("expected version-three compatibility warning for %s, got %#v", table, result.Warnings)
		}
	}
	for _, table := range versionThreeBackupTableNames() {
		doc.Tables[table] = []map[string]any{}
	}
	result, err = service.Validate(context.Background(), doc)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Compatible {
		t.Fatalf("expected interim version 2 containing version-three tables to validate, got %#v", result)
	}
}

func TestBackupVersionThreeRequiresVersionThreeTables(t *testing.T) {
	db := backupTestDB(t, t.TempDir())
	service := application.NewBackupService(db, t.TempDir())
	doc := &application.BackupDocument{
		Format: "railkeeper-backup", Version: 3, Tables: backupDocumentTablesThroughVersion(2),
	}

	result, err := service.Validate(context.Background(), doc)
	if err != nil {
		t.Fatal(err)
	}
	if result.Compatible || !containsWarning(result.Errors, "Tabelle accessory_product_attributes fehlt") {
		t.Fatalf("expected missing version-three table error, got %#v", result)
	}
}

func TestBackupVersionThreeWithoutVersionFourTablesRemainsCompatible(t *testing.T) {
	db := backupTestDB(t, t.TempDir())
	service := application.NewBackupService(db, t.TempDir())
	doc, err := service.Export(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	doc.Version = 3
	for _, table := range versionFourBackupTableNames() {
		delete(doc.Tables, table)
	}

	result, err := service.Validate(t.Context(), doc)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Compatible {
		t.Fatalf("expected version 3 without version-four tables to remain compatible, got %#v", result)
	}
	for _, table := range versionFourBackupTableNames() {
		if !containsWarning(result.Warnings, "Optionale Tabelle "+table+" fehlt") {
			t.Fatalf("expected version-four compatibility warning for %s, got %#v", table, result.Warnings)
		}
	}
}

func TestBackupVersionFourRequiresVersionFourTables(t *testing.T) {
	db := backupTestDB(t, t.TempDir())
	service := application.NewBackupService(db, t.TempDir())
	doc := &application.BackupDocument{
		Format: "railkeeper-backup", Version: 4, Tables: backupDocumentTablesThroughVersion(3),
	}

	result, err := service.Validate(t.Context(), doc)
	if err != nil {
		t.Fatal(err)
	}
	if result.Compatible || !containsWarning(result.Errors, "Tabelle layout_technical_positions fehlt") {
		t.Fatalf("expected missing version-four table error, got %#v", result)
	}
}

func TestBackupVersionFourWithoutTrackPlannerTablesRemainsCompatible(t *testing.T) {
	db := backupTestDB(t, t.TempDir())
	service := application.NewBackupService(db, t.TempDir())
	doc, err := service.Export(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	doc.Version = 4
	for _, table := range versionFiveBackupTableNames() {
		delete(doc.Tables, table)
	}

	result, err := service.Validate(t.Context(), doc)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Compatible {
		t.Fatalf("expected version 4 without track-planner tables to remain compatible, got %#v", result)
	}
	for _, table := range versionFiveBackupTableNames() {
		if !containsWarning(result.Warnings, "Optionale Tabelle "+table+" fehlt") {
			t.Fatalf("expected version-five compatibility warning for %s, got %#v", table, result.Warnings)
		}
	}
}

func TestBackupVersionFiveRequiresTrackPlannerTables(t *testing.T) {
	db := backupTestDB(t, t.TempDir())
	service := application.NewBackupService(db, t.TempDir())
	doc, err := service.Export(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	doc.Version = 5
	for _, table := range versionFiveBackupTableNames() {
		delete(doc.Tables, table)
	}

	result, err := service.Validate(t.Context(), doc)
	if err != nil {
		t.Fatal(err)
	}
	if result.Compatible || !containsWarning(result.Errors, "Tabelle track_geometry_libraries fehlt") {
		t.Fatalf("expected missing version-five track-planner table error, got %#v", result)
	}
}

func TestBackupVersionTwoRestoresRowsUsingVersionThreeColumnDefaults(t *testing.T) {
	db := backupTestDB(t, t.TempDir())
	service := application.NewBackupService(db, t.TempDir())
	doc := &application.BackupDocument{
		Format: "railkeeper-backup", Version: 2, Tables: backupDocumentTablesThroughVersion(2),
	}
	doc.Tables["accessory_products"] = []map[string]any{{
		"id": "legacy-product", "manufacturer": "Piko", "article_number": "55200", "name": "Legacy track",
		"category": "track", "tracking_mode": "quantity", "description": "old row",
		"created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z",
	}}

	if _, err := service.Import(context.Background(), doc); err != nil {
		t.Fatalf("expected old row to use new-column defaults: %v", err)
	}
	var inventoryNumber, articleType, stockUnit, inventoryStrategy string
	var packageQuantity, minimumStock int
	if err := db.QueryRow(`
SELECT inventory_number, article_type, package_quantity, stock_unit, minimum_stock, inventory_strategy
FROM accessory_products WHERE id='legacy-product'
`).Scan(&inventoryNumber, &articleType, &packageQuantity, &stockUnit, &minimumStock, &inventoryStrategy); err != nil {
		t.Fatal(err)
	}
	if inventoryNumber == "" || articleType != "other" || packageQuantity != 1 || stockUnit != "piece" || minimumStock != 0 ||
		inventoryStrategy != "quantity" {
		t.Fatalf("unexpected defaults: article=%q package=%d unit=%q minimum=%d strategy=%q",
			articleType, packageQuantity, stockUnit, minimumStock, inventoryStrategy)
	}
}

func TestBackupPreflightFailuresLeaveExistingDataUntouched(t *testing.T) {
	dataDir := t.TempDir()
	db := backupTestDB(t, dataDir)
	ctx := context.Background()
	vehicleService := application.NewVehicleService(db)
	sentinel, err := vehicleService.Create(ctx, application.CreateVehicleInput{
		Manufacturer: "Piko", Name: "Sentinel locomotive", Gauge: "H0", Category: "Lokomotive", Gattung: "Diesellok",
	}, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	repository := infrastructure.NewAccessoryRepository(db)
	accessories := application.NewAccessoryService(repository)
	value := 12.5
	unit := "mm"
	product, err := accessories.CreateProduct(ctx, application.CreateAccessoryProductInput{
		Manufacturer: "Test", Name: "Sentinel article", Category: "straight", ArticleType: domain.AccessoryArticleTrack,
		Subtype: "straight", PackageQuantity: 1, StockUnit: "piece", InventoryStrategy: domain.AccessoryInventoryQuantity,
		Attributes: []domain.AccessoryAttributeValue{{
			Key: "lengthMm", Kind: domain.AccessoryAttributeNumber, NumberValue: &value, Unit: &unit,
		}},
	}, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	service := application.NewBackupService(db, dataDir)

	t.Run("future version", func(t *testing.T) {
		doc, err := service.Export(ctx)
		if err != nil {
			t.Fatal(err)
		}
		doc.Version = 19
		if _, err := service.Import(ctx, doc); !errors.Is(err, application.ErrBackupInvalid) {
			t.Fatalf("expected future backup rejection, got %v", err)
		}
		assertBackupSentinels(t, ctx, vehicleService, accessories, sentinel.ID, product.ID)
	})

	for _, test := range []struct {
		name   string
		mutate func(map[string]any)
	}{
		{
			name: "structural union mismatch",
			mutate: func(row map[string]any) {
				row["value_type"] = "text"
			},
		},
		{
			name: "standard key kind mismatch",
			mutate: func(row map[string]any) {
				row["value_type"] = "text"
				row["text_value"] = "12.5"
				row["number_value"] = nil
				row["unit"] = nil
			},
		},
		{
			name: "custom key on standard article type",
			mutate: func(row map[string]any) {
				row["attribute_key"] = "customLength"
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			doc, err := service.Export(ctx)
			if err != nil {
				t.Fatal(err)
			}
			rows := doc.Tables["accessory_product_attributes"]
			if len(rows) != 1 {
				t.Fatalf("expected one attribute row, got %#v", rows)
			}
			test.mutate(rows[0])
			validation, err := service.Validate(ctx, doc)
			if err != nil {
				t.Fatal(err)
			}
			if validation.Compatible || !containsWarning(validation.Errors, "accessory_product_attributes") {
				t.Fatalf("expected semantic attribute validation error, got %#v", validation)
			}
			if _, err := service.Import(ctx, doc); !errors.Is(err, application.ErrBackupInvalid) {
				t.Fatalf("expected malformed attribute rejection, got %v", err)
			}
			assertBackupSentinels(t, ctx, vehicleService, accessories, sentinel.ID, product.ID)
		})
	}
}

func TestBackupRoundTripPreservesValidInactiveControlledAttributes(t *testing.T) {
	value, unit := 12.5, "mm"
	for _, test := range []struct {
		name             string
		originalMetadata map[string]any
		changedMetadata  map[string]any
		attribute        domain.AccessoryAttributeValue
	}{
		{
			name: "kind changed", originalMetadata: map[string]any{"kind": "number", "unit": "mm"},
			changedMetadata: map[string]any{"kind": "text"},
			attribute: domain.AccessoryAttributeValue{
				Key: "customField", Kind: domain.AccessoryAttributeNumber, NumberValue: &value, Unit: &unit,
			},
		},
		{
			name: "unit changed", originalMetadata: map[string]any{"kind": "number", "unit": "mm"},
			changedMetadata: map[string]any{"kind": "number", "unit": "cm"},
			attribute: domain.AccessoryAttributeValue{
				Key: "customField", Kind: domain.AccessoryAttributeNumber, NumberValue: &value, Unit: &unit,
			},
		},
		{
			name:             "bounds changed",
			originalMetadata: map[string]any{"kind": "number", "unit": "mm", "min": 10.0, "max": 20.0},
			changedMetadata:  map[string]any{"kind": "number", "unit": "mm", "min": 100.0, "max": 200.0},
			attribute: domain.AccessoryAttributeValue{
				Key: "customField", Kind: domain.AccessoryAttributeNumber, NumberValue: &value, Unit: &unit,
			},
		},
		{
			name:             "options changed",
			originalMetadata: map[string]any{"kind": "single_select", "options": []string{"red", "green"}},
			changedMetadata:  map[string]any{"kind": "single_select", "options": []string{"blue", "yellow"}},
			attribute: domain.AccessoryAttributeValue{
				Key: "customField", Kind: domain.AccessoryAttributeSingleSelect, OptionValues: []string{"red"},
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			dataDir := t.TempDir()
			db := backupTestDB(t, dataDir)
			ctx := t.Context()
			active := true
			masterData := application.NewMasterDataService(db)
			definition, err := masterData.Create(ctx, "accessory_custom_field", application.MasterDataInput{
				Key: "customField", Label: "Custom field", Active: &active, Metadata: test.originalMetadata,
			})
			if err != nil {
				t.Fatal(err)
			}
			repository := infrastructure.NewAccessoryRepository(db)
			accessories := application.NewAccessoryService(repository)
			product, err := accessories.CreateProduct(ctx, application.CreateAccessoryProductInput{
				Manufacturer: "Test", Name: "Historical custom article", Category: "other",
				ArticleType: domain.AccessoryArticleOther, Subtype: "other:other", PackageQuantity: 1,
				StockUnit: "piece", InventoryStrategy: domain.AccessoryInventoryQuantity,
				Attributes: []domain.AccessoryAttributeValue{test.attribute},
			}, "editor-1")
			if err != nil {
				t.Fatal(err)
			}
			historical, err := accessories.GetProduct(ctx, product.ID)
			if err != nil {
				t.Fatal(err)
			}
			inactive := false
			if _, err := masterData.Update(ctx, "accessory_custom_field", definition.Key,
				application.MasterDataInput{
					Label: definition.Label, Active: &inactive, Metadata: definition.Metadata,
				}); err != nil {
				t.Fatal(err)
			}
			changedDefinition, err := masterData.Update(ctx, "accessory_custom_field", definition.Key,
				application.MasterDataInput{
					Label: definition.Label, Active: &inactive, Metadata: test.changedMetadata,
				})
			if err != nil {
				t.Fatal(err)
			}

			service := application.NewBackupService(db, dataDir)
			doc, err := service.Export(ctx)
			if err != nil {
				t.Fatal(err)
			}
			validation, err := service.Validate(ctx, doc)
			if err != nil {
				t.Fatal(err)
			}
			if !validation.Compatible {
				t.Fatalf("valid inactive historical attribute backup rejected: %#v", validation.Errors)
			}
			if _, err := db.ExecContext(ctx,
				`UPDATE accessory_products SET name='Changed after export' WHERE id=?`, product.ID); err != nil {
				t.Fatal(err)
			}
			if _, err := service.Import(ctx, doc); err != nil {
				t.Fatal(err)
			}

			restored, err := accessories.GetProduct(ctx, product.ID)
			if err != nil {
				t.Fatal(err)
			}
			if restored.Name != historical.Name || !reflect.DeepEqual(restored.Attributes, historical.Attributes) {
				t.Fatalf("inactive historical attribute changed during restore: before=%#v after=%#v",
					historical, restored)
			}
			restoredDefinition, err := masterData.Get(ctx, "accessory_custom_field", definition.Key)
			if err != nil {
				t.Fatal(err)
			}
			if restoredDefinition.Active || !reflect.DeepEqual(restoredDefinition.Metadata, changedDefinition.Metadata) {
				t.Fatalf("inactive changed definition not restored: %#v", restoredDefinition)
			}
		})
	}
}

func TestBackupPreflightRejectsInvalidControlledCustomAttributesBeforeMutation(t *testing.T) {
	dataDir := t.TempDir()
	db := backupTestDB(t, dataDir)
	ctx := t.Context()
	active := true
	masterData := application.NewMasterDataService(db)
	if _, err := masterData.Create(ctx, "accessory_custom_field", application.MasterDataInput{
		Key: "customLength", Label: "Custom length", Active: &active,
		Metadata: map[string]any{"kind": "number", "unit": "mm", "min": 10.0, "max": 20.0},
	}); err != nil {
		t.Fatal(err)
	}
	repository := infrastructure.NewAccessoryRepository(db)
	accessories := application.NewAccessoryService(repository)
	value, unit := 12.5, "mm"
	product, err := accessories.CreateProduct(ctx, application.CreateAccessoryProductInput{
		Manufacturer: "Test", Name: "Sentinel article", Category: "other",
		ArticleType: domain.AccessoryArticleOther, Subtype: "other:other", PackageQuantity: 1,
		StockUnit: "piece", InventoryStrategy: domain.AccessoryInventoryQuantity,
		Attributes: []domain.AccessoryAttributeValue{{
			Key: "customLength", Kind: domain.AccessoryAttributeNumber, NumberValue: &value, Unit: &unit,
		}},
	}, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	vehicles := application.NewVehicleService(db)
	vehicle, err := vehicles.Create(ctx, application.CreateVehicleInput{
		Manufacturer: "Piko", Name: "Sentinel locomotive", Gauge: "H0",
		Category: "Lokomotive", Gattung: "Diesellok",
	}, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	service := application.NewBackupService(db, dataDir)

	customDefinition := func(t *testing.T, doc *application.BackupDocument) map[string]any {
		t.Helper()
		for _, row := range doc.Tables["master_data_entries"] {
			if row["type"] == "accessory_custom_field" && row["key"] == "customLength" {
				return row
			}
		}
		t.Fatal("custom definition missing from backup")
		return nil
	}
	customAttribute := func(t *testing.T, doc *application.BackupDocument) map[string]any {
		t.Helper()
		for _, row := range doc.Tables["accessory_product_attributes"] {
			if row["product_id"] == product.ID && row["attribute_key"] == "customLength" {
				return row
			}
		}
		t.Fatal("custom attribute missing from backup")
		return nil
	}
	cloneRow := func(row map[string]any) map[string]any {
		clone := make(map[string]any, len(row))
		for key, value := range row {
			clone[key] = value
		}
		return clone
	}
	setSelectAttribute := func(row map[string]any, kind, encoded string) {
		row["value_type"] = kind
		row["text_value"] = nil
		row["number_value"] = nil
		row["unit"] = nil
		row["boolean_value"] = nil
		row["date_value"] = nil
		row["single_select_value"] = nil
		row["multi_select_value"] = nil
		if kind == "single_select" {
			row["single_select_value"] = encoded
		} else {
			row["multi_select_value"] = encoded
		}
	}

	for _, test := range []struct {
		name   string
		mutate func(*testing.T, *application.BackupDocument)
	}{
		{name: "malformed definition", mutate: func(t *testing.T, doc *application.BackupDocument) {
			customDefinition(t, doc)["metadata_json"] = `{"kind":"json"}`
		}},
		{name: "undefined key", mutate: func(t *testing.T, doc *application.BackupDocument) {
			customAttribute(t, doc)["attribute_key"] = "undefined"
		}},
		{name: "kind mismatch", mutate: func(t *testing.T, doc *application.BackupDocument) {
			row := customAttribute(t, doc)
			row["value_type"], row["text_value"], row["number_value"], row["unit"] = "text", "12.5", nil, nil
		}},
		{name: "unit mismatch", mutate: func(t *testing.T, doc *application.BackupDocument) {
			customAttribute(t, doc)["unit"] = "cm"
		}},
		{name: "bounds violation", mutate: func(t *testing.T, doc *application.BackupDocument) {
			customAttribute(t, doc)["number_value"] = 25.0
		}},
		{name: "unsupported option", mutate: func(t *testing.T, doc *application.BackupDocument) {
			customDefinition(t, doc)["metadata_json"] = `{"kind":"single_select","options":["red","green"]}`
			setSelectAttribute(customAttribute(t, doc), "single_select", "blue")
		}},
		{name: "duplicate attribute key", mutate: func(t *testing.T, doc *application.BackupDocument) {
			duplicate := cloneRow(customAttribute(t, doc))
			duplicate["id"] = "duplicate-attribute"
			doc.Tables["accessory_product_attributes"] = append(
				doc.Tables["accessory_product_attributes"], duplicate,
			)
		}},
		{name: "duplicate option value", mutate: func(t *testing.T, doc *application.BackupDocument) {
			customDefinition(t, doc)["metadata_json"] = `{"kind":"multi_select","options":["red","green"]}`
			setSelectAttribute(customAttribute(t, doc), "multi_select", `["red","red"]`)
		}},
		{name: "duplicate definition", mutate: func(t *testing.T, doc *application.BackupDocument) {
			duplicate := cloneRow(customDefinition(t, doc))
			duplicate["id"] = "duplicate-definition"
			doc.Tables["master_data_entries"] = append(doc.Tables["master_data_entries"], duplicate)
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			doc, err := service.Export(ctx)
			if err != nil {
				t.Fatal(err)
			}
			test.mutate(t, doc)
			validation, err := service.Validate(ctx, doc)
			if err != nil {
				t.Fatal(err)
			}
			if validation.Compatible || !containsWarning(validation.Errors, "accessory_product_attributes") {
				t.Fatalf("expected controlled attribute preflight error, got %#v", validation)
			}
			if _, err := service.Import(ctx, doc); !errors.Is(err, application.ErrBackupInvalid) {
				t.Fatalf("expected controlled attribute restore rejection, got %v", err)
			}
			assertBackupSentinels(t, ctx, vehicles, accessories, vehicle.ID, product.ID)
		})
	}
}

func TestBackupRejectsUnsafeFilePath(t *testing.T) {
	db := testDB(t)
	service := application.NewBackupService(db, t.TempDir())

	_, err := service.Import(context.Background(), &application.BackupDocument{
		Format:  "railkeeper-backup",
		Version: 1,
		Tables:  map[string][]map[string]any{},
		Files: []application.BackupFile{{
			Path:          "../outside.txt",
			ContentBase64: "dGVzdA==",
		}},
	})
	if !errors.Is(err, application.ErrBackupPath) {
		t.Fatalf("expected unsafe backup path error, got %v", err)
	}
}

func TestBackupRestoreLeavesDatabaseAndUploadsUntouchedWhenFileStagingFails(t *testing.T) {
	dataDir := t.TempDir()
	db := backupTestDB(t, dataDir)
	ctx := context.Background()
	vehicles := application.NewVehicleService(db)
	existing, err := vehicles.Create(ctx, application.CreateVehicleInput{
		Manufacturer: "Piko",
		Name:         "Bestehende Lok",
		Gauge:        "H0",
		Category:     "Lokomotive",
		Gattung:      "Diesellok",
	}, "actor-1")
	if err != nil {
		t.Fatal(err)
	}
	existingUpload := filepath.Join(dataDir, "uploads", "vehicles", existing.ID, "manual.pdf")
	if err := os.MkdirAll(filepath.Dir(existingUpload), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(existingUpload, []byte("existing manual"), 0o600); err != nil {
		t.Fatal(err)
	}

	service := application.NewBackupService(db, dataDir)
	backup, err := service.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	vehicleRows := backup.Tables["vehicles"]
	if len(vehicleRows) != 1 {
		t.Fatalf("expected exported vehicle row, got %#v", vehicleRows)
	}
	restoredRow := map[string]any{}
	for key, value := range vehicleRows[0] {
		restoredRow[key] = value
	}
	restoredRow["id"] = "restored-vehicle"
	restoredRow["name"] = "Restored Lok"
	backup.Tables["vehicles"] = []map[string]any{restoredRow}
	backup.Files = []application.BackupFile{
		{
			Path:          "uploads/conflict",
			SizeBytes:     4,
			ContentBase64: "ZmlsZQ==",
		},
		{
			Path:          "uploads/conflict/nested.txt",
			SizeBytes:     6,
			ContentBase64: "bmVzdGVk",
		},
	}

	_, err = service.Import(ctx, backup)
	if err == nil {
		t.Fatal("expected conflicting backup file paths to fail restore")
	}
	restored, err := vehicles.Get(ctx, existing.ID)
	if err != nil {
		t.Fatalf("existing vehicle should remain after failed restore: %v", err)
	}
	if restored.Name != "Bestehende Lok" {
		t.Fatalf("existing vehicle changed after failed restore: %#v", restored)
	}
	if _, err := vehicles.Get(ctx, "restored-vehicle"); !errors.Is(err, application.ErrVehicleNotFound) {
		t.Fatalf("restored vehicle should not be committed after failed restore, got %v", err)
	}
	data, err := os.ReadFile(existingUpload)
	if err != nil {
		t.Fatalf("existing upload should remain after failed restore: %v", err)
	}
	if string(data) != "existing manual" {
		t.Fatalf("existing upload changed after failed restore: %q", string(data))
	}
}

func TestBackupValidationReportsIncompatibleDocuments(t *testing.T) {
	db := testDB(t)
	service := application.NewBackupService(db, t.TempDir())

	result, err := service.Validate(context.Background(), &application.BackupDocument{
		Format:  "other",
		Version: 99,
		Tables:  map[string][]map[string]any{"vehicles": {}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Compatible {
		t.Fatalf("expected incompatible backup")
	}
	if len(result.Errors) == 0 {
		t.Fatalf("expected validation errors")
	}
}

func containsWarning(warnings []string, needle string) bool {
	for _, warning := range warnings {
		if strings.Contains(warning, needle) {
			return true
		}
	}
	return false
}

func backupDocumentTablesWithout(excludedTables ...string) map[string][]map[string]any {
	excluded := map[string]bool{}
	for _, table := range excludedTables {
		excluded[table] = true
	}
	tables := map[string][]map[string]any{}
	for _, table := range []string{
		"master_data_entries",
		"master_data_relations",
		"inventory_number_schemes",
		"vehicles",
		"inventory_number_history",
		"vehicle_external_mappings",
		"file_blobs",
		"vehicle_images",
		"vehicle_attachments",
		"vehicle_maintenance",
		"vehicle_spare_parts",
		"vehicle_functions",
		"vehicle_cv_files",
		"vehicle_cv_values",
		"vehicle_cv_value_history",
		"exhibition_lists",
		"exhibition_entries",
	} {
		if !excluded[table] {
			tables[table] = []map[string]any{}
		}
	}
	return tables
}

func stageOneBackupTableNames() []string {
	return []string{
		"storage_locations", "accessory_products", "accessory_stock", "accessory_assets", "layouts",
		"layout_units", "plan_variants", "plan_revisions", "layout_configurations",
		"layout_configuration_units", "accessory_reservations", "accessory_installations",
	}
}

func versionThreeBackupTableNames() []string {
	return []string{
		"accessory_product_attributes", "accessory_purchases", "accessory_documents",
		"accessory_stock_movements", "accessory_installation_condition_history",
	}
}

func versionFourBackupTableNames() []string {
	return []string{
		"layout_unit_outline_points", "layout_technical_positions",
		"accessory_reservation_positions", "accessory_installation_positions",
	}
}

func versionFiveBackupTableNames() []string {
	return []string{"track_geometry_libraries", "track_geometry_definitions", "plan_track_objects"}
}

func versionSixBackupTableNames() []string {
	return []string{"plan_track_object_reservations"}
}

func versionSevenBackupTableNames() []string {
	return []string{"layout_unit_ports"}
}

func legacyOptionalBackupTableNames() []string {
	return []string{
		"file_blobs", "vehicle_external_mappings", "vehicle_spare_parts",
		"exhibition_lists", "exhibition_entries",
	}
}

func assertBackupSentinels(
	t *testing.T,
	ctx context.Context,
	vehicles *application.VehicleService,
	accessories *application.AccessoryService,
	vehicleID string,
	productID string,
) {
	t.Helper()
	vehicle, err := vehicles.Get(ctx, vehicleID)
	if err != nil || vehicle.Name != "Sentinel locomotive" {
		t.Fatalf("sentinel vehicle changed after failed preflight: vehicle=%#v err=%v", vehicle, err)
	}
	product, err := accessories.GetProduct(ctx, productID)
	if err != nil || product.Name != "Sentinel article" {
		t.Fatalf("sentinel article changed after failed preflight: product=%#v err=%v", product, err)
	}
}

func seedStageOneBackupData(t *testing.T, db *sql.DB, vehicleID string) {
	t.Helper()
	_, err := db.Exec(`
INSERT INTO storage_locations(id, parent_id, name, description, created_at, updated_at) VALUES
  ('location-root', NULL, 'Club depot', '', '2026-08-07T18:00:00Z', '2026-08-07T18:00:00Z'),
  ('location-child', 'location-root', 'Track box', '', '2026-08-07T18:00:00Z', '2026-08-07T18:00:00Z');
INSERT INTO accessory_products(id, inventory_number, manufacturer, article_number, name, category, tracking_mode, description, created_at, updated_at) VALUES
  ('product-quantity', 'RK-ART-000001', 'Tillig', '83101', 'Straight track', 'track', 'quantity', '', '2026-08-07T18:00:00Z', '2026-08-07T18:00:00Z'),
  ('product-individual', 'RK-ART-000002', 'ESU', '59610', 'LokSound decoder', 'decoder', 'individual', '', '2026-08-07T18:00:00Z', '2026-08-07T18:00:00Z');
INSERT INTO accessory_stock(product_id, location_id, quantity, updated_at)
  VALUES ('product-quantity', 'location-child', 12, '2026-08-07T18:00:00Z');
INSERT INTO accessory_assets(id, product_id, inventory_number, condition_state, lifecycle_state, storage_location_id, created_at, updated_at)
  VALUES ('asset-1', 'product-individual', 'RK-ZUB-0001', 'ready', 'stored', 'location-child', '2026-08-07T18:00:00Z', '2026-08-07T18:00:00Z');
INSERT INTO layouts(id, name, kind, gauge, scale, description, version, archived, created_at, updated_at)
  VALUES ('layout-1', 'Clubanlage Falkenstein', 'club', 'TT', '1:120', '', 1, 0, '2026-08-07T18:00:00Z', '2026-08-07T18:00:00Z');
INSERT INTO layout_units(id, layout_id, name, kind, owner_label, width_mm, height_mm, version, archived, created_at, updated_at)
  VALUES ('unit-1', 'layout-1', 'Station module', 'module', 'Club', 1200, 500, 1, 0, '2026-08-07T18:00:00Z', '2026-08-07T18:00:00Z');
INSERT INTO plan_variants(id, layout_unit_id, name, description, archived, created_at, updated_at)
  VALUES ('variant-1', 'unit-1', 'Exhibition operation', '', 0, '2026-08-07T18:00:00Z', '2026-08-07T18:00:00Z');
INSERT INTO plan_revisions(id, variant_id, revision_number, status, base_revision_id, version, created_by, published_by, published_at, created_at, updated_at) VALUES
  ('revision-1', 'variant-1', 1, 'published', NULL, 2, 'planner-1', 'planner-1', '2026-08-07T18:00:00Z', '2026-08-07T18:00:00Z', '2026-08-07T18:00:00Z'),
  ('revision-2', 'variant-1', 2, 'draft', 'revision-1', 1, 'planner-1', NULL, NULL, '2026-08-07T18:00:00Z', '2026-08-07T18:00:00Z');
INSERT INTO layout_configurations(id, layout_id, name, description, version, archived, created_at, updated_at)
  VALUES ('configuration-1', 'layout-1', 'Autumn exhibition', '', 1, 0, '2026-08-07T18:00:00Z', '2026-08-07T18:00:00Z');
INSERT INTO layout_configuration_units(configuration_id, unit_id, plan_revision_id, position_x_mm, position_y_mm, rotation_degrees, sort_order)
  VALUES ('configuration-1', 'unit-1', 'revision-1', 250, 100, 90, 0);
INSERT INTO accessory_reservations(id, product_id, location_id, quantity, vehicle_id, status, note, created_by, created_at, updated_at)
  VALUES ('reservation-1', 'product-quantity', 'location-child', 2, ?, 'active', '', 'planner-1', '2026-08-07T18:00:00Z', '2026-08-07T18:00:00Z');
INSERT INTO accessory_installations(id, product_id, source_location_id, quantity, layout_unit_id, condition_state, installed_by, installed_at, notes, removal_notes)
  VALUES ('installation-1', 'product-quantity', 'location-child', 3, 'unit-1', 'ready', 'editor-1', '2026-08-07T18:00:00Z', '', '');
`, vehicleID)
	if err != nil {
		t.Fatal(err)
	}
}

func backupTestDB(t *testing.T, dataDir string) *sql.DB {
	t.Helper()

	db, err := infrastructure.OpenSQLite(dataDir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	migrationsDir := filepath.Join("..", "..", "migrations")
	if err := infrastructure.Migrate(db, migrationsDir); err != nil {
		t.Fatal(err)
	}
	if err := infrastructure.SeedRoles(db); err != nil {
		t.Fatal(err)
	}

	return db
}
