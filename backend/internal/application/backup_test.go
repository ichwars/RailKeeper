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
	if backup.Version != 3 {
		t.Fatalf("expected backup version 3, got %d", backup.Version)
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

func TestBackupVersionThreeRoundTripPreservesStageOneDataReferences(t *testing.T) {
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
	if backup.Version != 3 {
		t.Fatalf("expected version 3 export, got %d", backup.Version)
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

func TestBackupVersionThreeRoundTripPreservesArticleManagementData(t *testing.T) {
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
		ArticleType: domain.AccessoryArticleOther, Subtype: "decoder", PackageQuantity: 1,
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
	if backup.Version != 3 {
		t.Fatalf("expected version 3 export, got %d", backup.Version)
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
		t.Fatalf("expected complete version 3 backup, got %#v", validation)
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
	var articleType, stockUnit, inventoryStrategy string
	var packageQuantity, minimumStock int
	if err := db.QueryRow(`
SELECT article_type, package_quantity, stock_unit, minimum_stock, inventory_strategy
FROM accessory_products WHERE id='legacy-product'
`).Scan(&articleType, &packageQuantity, &stockUnit, &minimumStock, &inventoryStrategy); err != nil {
		t.Fatal(err)
	}
	if articleType != "other" || packageQuantity != 1 || stockUnit != "piece" || minimumStock != 0 ||
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
		doc.Version = 4
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
