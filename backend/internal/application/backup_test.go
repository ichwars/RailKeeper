package application_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
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
	if backup.Version != 2 {
		t.Fatalf("expected backup version 2, got %d", backup.Version)
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

func TestBackupVersionTwoRoundTripPreservesStageOneDataReferences(t *testing.T) {
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
	if backup.Version != 2 {
		t.Fatalf("expected version 2 export, got %d", backup.Version)
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
	for _, table := range []string{"users", "user_roles", "sessions", "audit_log", "rate_limit_attempts"} {
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

func TestBackupValidationAllowsMissingOptionalExhibitionTables(t *testing.T) {
	db := backupTestDB(t, t.TempDir())
	service := application.NewBackupService(db, t.TempDir())
	doc := &application.BackupDocument{
		Format:  "railkeeper-backup",
		Version: 1,
		Tables:  backupDocumentTablesWithout("exhibition_lists", "exhibition_entries"),
	}

	result, err := service.Validate(context.Background(), doc)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Compatible {
		t.Fatalf("expected backup without optional exhibition tables to remain compatible, got %#v", result)
	}
	if !containsWarning(result.Warnings, "Optionale Tabelle exhibition_lists fehlt") || !containsWarning(result.Warnings, "Optionale Tabelle exhibition_entries fehlt") {
		t.Fatalf("expected optional table warnings, got %#v", result.Warnings)
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

func seedStageOneBackupData(t *testing.T, db *sql.DB, vehicleID string) {
	t.Helper()
	_, err := db.Exec(`
INSERT INTO storage_locations(id, parent_id, name, description, created_at, updated_at) VALUES
  ('location-root', NULL, 'Club depot', '', '2026-08-07T18:00:00Z', '2026-08-07T18:00:00Z'),
  ('location-child', 'location-root', 'Track box', '', '2026-08-07T18:00:00Z', '2026-08-07T18:00:00Z');
INSERT INTO accessory_products(id, manufacturer, article_number, name, category, tracking_mode, description, created_at, updated_at) VALUES
  ('product-quantity', 'Tillig', '83101', 'Straight track', 'track', 'quantity', '', '2026-08-07T18:00:00Z', '2026-08-07T18:00:00Z'),
  ('product-individual', 'ESU', '59610', 'LokSound decoder', 'decoder', 'individual', '', '2026-08-07T18:00:00Z', '2026-08-07T18:00:00Z');
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
