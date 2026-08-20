package infrastructure_test

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sync"
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/infrastructure"
)

func TestDataTransferApplyCommitsTwoVehiclesAndAudit(t *testing.T) {
	db := testDB(t)
	repository := infrastructure.NewDataTransferRepository(db)
	job := createApplyJob(t, repository, "sha-two", []application.DataTransferPreviewRecord{
		applyVehicleRecord(t, "RK-TX-1", "Roco"),
		applyVehicleRecord(t, "RK-TX-2", "Piko"),
	})

	if err := repository.ApplyImport(t.Context(), job, "editor-1"); err != nil {
		t.Fatal(err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM vehicles WHERE inventory_number IN ('RK-TX-1', 'RK-TX-2')").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("imported vehicle count = %d, want 2", count)
	}
	loaded, err := repository.GetJob(t.Context(), job.ID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.State != application.TransferJobCompleted || loaded.ConfirmedByUserID != "editor-1" ||
		loaded.ConfirmedAt == "" || loaded.CompletedAt == "" || loaded.Revision != job.Revision+1 {
		t.Fatalf("unexpected completed job: %#v", loaded)
	}
	var actor, action, targetType, targetID, details string
	if err := db.QueryRow(`SELECT COALESCE(actor_user_id, ''), action, target_type, target_id, details_json
		FROM audit_logs WHERE target_id=?`, job.ID).Scan(&actor, &action, &targetType, &targetID, &details); err != nil {
		t.Fatal(err)
	}
	if actor != "editor-1" || action != "data_transfer.import_applied" ||
		targetType != "data_transfer_job" || targetID != job.ID || details == "" {
		t.Fatalf("unexpected audit: %q %q %q %q %q", actor, action, targetType, targetID, details)
	}
}

func TestDataTransferApplyCommitsAllSelectedAreasInOneTransaction(t *testing.T) {
	db := testDB(t)
	repository := infrastructure.NewDataTransferRepository(db)
	vehicle := applyVehicleRecord(t, "RK-ALL-1", "Roco")
	accessoryData, err := json.Marshal(application.TransferAccessory{
		InventoryNumber: "RK-ART-ALL-1", Manufacturer: "Viessmann", Name: "Signal", Category: "Signal",
		TrackingMode: "quantity", ManufacturerStatus: "unknown", ArticleType: "other", Gauges: []string{"H0"},
		PackageQuantity: 1, StockUnit: "piece", InventoryStrategy: "quantity",
	})
	if err != nil {
		t.Fatal(err)
	}
	listData, err := json.Marshal(application.TransferExhibitionList{
		Designation: "All areas", Date: "2026-08-20", Entries: []application.TransferExhibitionEntry{{
			Owner: "Club", LocomotiveName: "Guest", DayScope: "all",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	job := createApplyJob(t, repository, "sha-all", []application.DataTransferPreviewRecord{
		vehicle,
		{Area: application.TransferAccessories, RecordKey: "RK-ART-ALL-1", Classification: "ready", ProposedAction: "create", Data: accessoryData},
		{Area: application.TransferExhibitionLists, RecordKey: "All areas|2026-08-20", Classification: "ready", ProposedAction: "create", Data: listData},
	})

	if err := repository.ApplyImport(t.Context(), job, "editor-1"); err != nil {
		t.Fatal(err)
	}
	for table, want := range map[string]int{"vehicles": 1, "accessory_products": 1, "exhibition_lists": 1, "exhibition_entries": 1} {
		var count int
		if err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != want {
			t.Fatalf("%s count = %d, want %d", table, count, want)
		}
	}
}

func TestDataTransferApplyRollsBackWhenSecondRowFails(t *testing.T) {
	db := testDB(t)
	repository := infrastructure.NewDataTransferRepository(db)
	job := createApplyJob(t, repository, "sha-rollback", []application.DataTransferPreviewRecord{
		applyVehicleRecord(t, "RK-TX-1", "Roco"),
		applyVehicleRecord(t, "RK-TX-1", "Piko"),
	})

	if err := repository.ApplyImport(t.Context(), job, "editor-1"); err == nil {
		t.Fatal("expected second-row constraint failure")
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM vehicles WHERE inventory_number IN ('RK-TX-1', 'RK-TX-2')").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("partial import committed %d vehicles", count)
	}
	assertApplyJobStillReady(t, repository, job.ID)
}

func TestDataTransferApplyCASAllowsOneConcurrentConfirmation(t *testing.T) {
	db := testDB(t)
	repository := infrastructure.NewDataTransferRepository(db)
	job := createApplyJob(t, repository, "sha-cas", []application.DataTransferPreviewRecord{
		applyVehicleRecord(t, "RK-CAS-1", "Roco"),
	})
	start := make(chan struct{})
	errorsByWorker := make([]error, 2)
	var workers sync.WaitGroup
	for index := range errorsByWorker {
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			errorsByWorker[index] = repository.ApplyImport(t.Context(), job, "editor-1")
		}()
	}
	close(start)
	workers.Wait()
	successes, conflicts := 0, 0
	for _, err := range errorsByWorker {
		if err == nil {
			successes++
		} else if errors.Is(err, application.ErrDataTransferConflict) {
			conflicts++
		} else {
			t.Fatalf("unexpected concurrent apply error: %v", err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("concurrent results: success=%d conflict=%d errors=%v", successes, conflicts, errorsByWorker)
	}
	var vehicles, audits int
	if err := db.QueryRow(`SELECT COUNT(*) FROM vehicles WHERE inventory_number='RK-CAS-1'`).Scan(&vehicles); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM audit_logs WHERE target_id=?`, job.ID).Scan(&audits); err != nil {
		t.Fatal(err)
	}
	if vehicles != 1 || audits != 1 {
		t.Fatalf("concurrent apply committed vehicles=%d audits=%d", vehicles, audits)
	}
}

func TestDataTransferApplyRejectsStaleTargetChangedSourceAndUnresolvedIssues(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*testing.T, *sql.DB, *infrastructure.DataTransferRepository) application.DataTransferJob
	}{
		{
			name: "stale target updated_at",
			setup: func(t *testing.T, db *sql.DB, repository *infrastructure.DataTransferRepository) application.DataTransferJob {
				if _, err := db.Exec(`INSERT INTO vehicles(id, inventory_number, manufacturer, name, gauge, created_at, updated_at)
					VALUES('vehicle-existing', 'RK-OLD', 'Roco', 'Old', 'H0', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z')`); err != nil {
					t.Fatal(err)
				}
				record := applyVehicleRecord(t, "RK-OLD", "Piko")
				snapshot, err := repository.Snapshot(t.Context(), []application.TransferArea{application.TransferVehicles})
				if err != nil {
					t.Fatal(err)
				}
				record.ProposedAction = "replace"
				record.TargetID = "vehicle-existing"
				record.TargetUpdatedAt = "2026-01-01T00:00:00Z"
				record.TargetFingerprint = applyTargetFingerprint(t, snapshot.Vehicles[0])
				job := createApplyJob(t, repository, "sha-stale", []application.DataTransferPreviewRecord{record})
				if err := repository.ReplaceIssues(t.Context(), job.ID, []application.DataTransferIssue{{
					JobID: job.ID, Area: application.TransferVehicles, RecordKey: "RK-OLD",
					Severity: application.TransferIssueWarning, Code: "duplicate_inventory_number",
					SelectedResolution: "replace",
				}}); err != nil {
					t.Fatal(err)
				}
				if _, err := db.Exec(`UPDATE vehicles SET updated_at='2026-02-01T00:00:00Z' WHERE id='vehicle-existing'`); err != nil {
					t.Fatal(err)
				}
				return job
			},
		},
		{
			name: "changed upload hash",
			setup: func(t *testing.T, _ *sql.DB, repository *infrastructure.DataTransferRepository) application.DataTransferJob {
				job := createApplyJob(t, repository, "sha-original", []application.DataTransferPreviewRecord{
					applyVehicleRecord(t, "RK-HASH", "Roco"),
				})
				job.SourceSHA256 = "sha-changed"
				updated, err := repository.UpdateJob(t.Context(), job)
				if err != nil {
					t.Fatal(err)
				}
				return updated
			},
		},
		{
			name: "unresolved issue",
			setup: func(t *testing.T, _ *sql.DB, repository *infrastructure.DataTransferRepository) application.DataTransferJob {
				job := createApplyJob(t, repository, "sha-issue", []application.DataTransferPreviewRecord{
					applyVehicleRecord(t, "RK-ISSUE", "Roco"),
				})
				if err := repository.ReplaceIssues(t.Context(), job.ID, []application.DataTransferIssue{{
					JobID: job.ID, Area: application.TransferVehicles, RecordKey: "RK-ISSUE",
					Severity: application.TransferIssueWarning, Code: "duplicate_inventory_number",
				}}); err != nil {
					t.Fatal(err)
				}
				return job
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := testDB(t)
			repository := infrastructure.NewDataTransferRepository(db)
			job := test.setup(t, db, repository)
			err := repository.ApplyImport(t.Context(), job, "editor-1")
			if !errors.Is(err, application.ErrDataTransferConflict) {
				t.Fatalf("ApplyImport() error = %v, want conflict", err)
			}
			assertApplyJobStillReady(t, repository, job.ID)
		})
	}
}

func TestDataTransferApplyRejectsSameSecondVehicleMutation(t *testing.T) {
	db := testDB(t)
	repository := infrastructure.NewDataTransferRepository(db)
	const timestamp = "2026-01-01T00:00:00Z"
	if _, err := db.Exec(`INSERT INTO vehicles(id, inventory_number, manufacturer, name, gauge, created_at, updated_at)
		VALUES('vehicle-fingerprint', 'RK-FP', 'Roco', 'Original', 'H0', ?, ?)`, timestamp, timestamp); err != nil {
		t.Fatal(err)
	}
	snapshot, err := repository.Snapshot(t.Context(), []application.TransferArea{application.TransferVehicles})
	if err != nil {
		t.Fatal(err)
	}
	record := applyVehicleRecord(t, "RK-FP", "Piko")
	record.TargetID = "vehicle-fingerprint"
	record.TargetUpdatedAt = timestamp
	record.TargetFingerprint = applyTargetFingerprint(t, snapshot.Vehicles[0])
	record.ProposedAction = "replace"
	record.Classification = "warning"
	job := createApplyJob(t, repository, "sha-vehicle-fingerprint", []application.DataTransferPreviewRecord{record})
	resolveApplyIssue(t, repository, job.ID, record, "duplicate_inventory_number", "replace")
	if _, err := db.Exec(`UPDATE vehicles SET name='Changed in same second', updated_at=? WHERE id=?`,
		timestamp, record.TargetID); err != nil {
		t.Fatal(err)
	}

	if err := repository.ApplyImport(t.Context(), job, "editor-1"); !errors.Is(err, application.ErrDataTransferConflict) {
		t.Fatalf("ApplyImport() error = %v, want same-second fingerprint conflict", err)
	}
	assertApplyJobStillReady(t, repository, job.ID)
}

func TestDataTransferApplyRejectsSameSecondAccessoryChildMutation(t *testing.T) {
	db := testDB(t)
	repository := infrastructure.NewDataTransferRepository(db)
	const timestamp = "2026-01-01T00:00:00Z"
	if _, err := db.Exec(`INSERT INTO storage_locations(id, name, created_at, updated_at)
		VALUES('location-fingerprint', 'Shelf', ?, ?)`, timestamp, timestamp); err != nil {
		t.Fatal(err)
	}
	insertApplyAccessoryProduct(t, db, "product-fingerprint", "RK-ART-FP", timestamp)
	if _, err := db.Exec(`INSERT INTO accessory_stock(product_id, location_id, quantity, updated_at)
		VALUES('product-fingerprint', 'location-fingerprint', 2, ?)`, timestamp); err != nil {
		t.Fatal(err)
	}
	snapshot, err := repository.Snapshot(t.Context(), []application.TransferArea{application.TransferAccessories})
	if err != nil {
		t.Fatal(err)
	}
	target := snapshot.Accessories[0]
	data, err := json.Marshal(target)
	if err != nil {
		t.Fatal(err)
	}
	record := application.DataTransferPreviewRecord{
		Area: application.TransferAccessories, RecordKey: target.InventoryNumber, Classification: "warning",
		ProposedAction: "replace", TargetID: target.ID, TargetUpdatedAt: timestamp,
		TargetFingerprint: applyTargetFingerprint(t, target), Data: data,
	}
	job := createApplyJob(t, repository, "sha-accessory-fingerprint", []application.DataTransferPreviewRecord{record})
	resolveApplyIssue(t, repository, job.ID, record, "duplicate_inventory_number", "replace")
	if _, err := db.Exec(`UPDATE accessory_stock SET quantity=9, updated_at=? WHERE product_id=?`,
		timestamp, target.ID); err != nil {
		t.Fatal(err)
	}

	if err := repository.ApplyImport(t.Context(), job, "editor-1"); !errors.Is(err, application.ErrDataTransferConflict) {
		t.Fatalf("ApplyImport() error = %v, want accessory aggregate conflict", err)
	}
	assertApplyJobStillReady(t, repository, job.ID)
}

func TestDataTransferApplyRejectsSameSecondExhibitionEntryMutation(t *testing.T) {
	db := testDB(t)
	repository := infrastructure.NewDataTransferRepository(db)
	const timestamp = "2026-01-01T00:00:00Z"
	if _, err := db.Exec(`INSERT INTO exhibition_lists(id, designation, list_date, created_at, updated_at)
		VALUES('list-fingerprint', 'Show', '2026-08-20', ?, ?)`, timestamp, timestamp); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO exhibition_entries(
		id, list_id, vehicle_id, owner, locomotive_name, day_scope, notes, created_at, updated_at
	) VALUES('entry-fingerprint', 'list-fingerprint', '', 'Club', 'BR 01', 'all', 'Original', ?, ?)`,
		timestamp, timestamp); err != nil {
		t.Fatal(err)
	}
	snapshot, err := repository.Snapshot(t.Context(), []application.TransferArea{application.TransferExhibitionLists})
	if err != nil {
		t.Fatal(err)
	}
	target := snapshot.ExhibitionLists[0]
	data, err := json.Marshal(target)
	if err != nil {
		t.Fatal(err)
	}
	record := application.DataTransferPreviewRecord{
		Area: application.TransferExhibitionLists, RecordKey: target.ID, Classification: "warning",
		ProposedAction: "replace", TargetID: target.ID, TargetUpdatedAt: timestamp,
		TargetFingerprint: applyTargetFingerprint(t, target), Data: data,
	}
	job := createApplyJob(t, repository, "sha-list-fingerprint", []application.DataTransferPreviewRecord{record})
	resolveApplyIssue(t, repository, job.ID, record, "duplicate_exhibition_list", "replace")
	if _, err := db.Exec(`UPDATE exhibition_entries SET notes='Changed in same second', updated_at=? WHERE id=?`,
		timestamp, "entry-fingerprint"); err != nil {
		t.Fatal(err)
	}

	if err := repository.ApplyImport(t.Context(), job, "editor-1"); !errors.Is(err, application.ErrDataTransferConflict) {
		t.Fatalf("ApplyImport() error = %v, want exhibition aggregate conflict", err)
	}
	assertApplyJobStillReady(t, repository, job.ID)
}

func TestDataTransferApplyNeverReplacesLockedExhibitionList(t *testing.T) {
	db := testDB(t)
	repository := infrastructure.NewDataTransferRepository(db)
	if _, err := db.Exec(`INSERT INTO exhibition_lists(id, designation, list_date, locked, created_at, updated_at)
		VALUES('list-locked', 'Clubtag', '2026-08-20', 1, '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z')`); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(application.TransferExhibitionList{Designation: "Clubtag", Date: "2026-08-20"})
	if err != nil {
		t.Fatal(err)
	}
	record := application.DataTransferPreviewRecord{
		Area: application.TransferExhibitionLists, RecordKey: "Clubtag|2026-08-20",
		Classification: "error", ProposedAction: "copy", TargetID: "list-locked",
		TargetUpdatedAt: "2026-01-01T00:00:00Z", Data: data,
	}
	job := createApplyJob(t, repository, "sha-list", []application.DataTransferPreviewRecord{record})
	if err := repository.ReplaceIssues(t.Context(), job.ID, []application.DataTransferIssue{{
		JobID: job.ID, Area: application.TransferExhibitionLists, RecordKey: record.RecordKey,
		Severity: application.TransferIssueError, Code: "locked_exhibition_list", SelectedResolution: "replace",
	}}); err != nil {
		t.Fatal(err)
	}

	if err := repository.ApplyImport(t.Context(), job, "editor-1"); !errors.Is(err, application.ErrDataTransferConflict) {
		t.Fatalf("ApplyImport() error = %v, want conflict", err)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM exhibition_lists`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("exhibition list count = %d, want 1", count)
	}
}

func TestDataTransferApplyUsesApprovedVehicleReplace(t *testing.T) {
	db := testDB(t)
	repository := infrastructure.NewDataTransferRepository(db)
	if _, err := db.Exec(`INSERT INTO vehicles(id, inventory_number, manufacturer, name, gauge, created_at, updated_at)
		VALUES('vehicle-replace', 'RK-REPLACE', 'Roco', 'Old', 'H0', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z')`); err != nil {
		t.Fatal(err)
	}
	record := applyVehicleRecord(t, "RK-REPLACE", "Piko")
	snapshot, err := repository.Snapshot(t.Context(), []application.TransferArea{application.TransferVehicles})
	if err != nil {
		t.Fatal(err)
	}
	record.TargetID = "vehicle-replace"
	record.TargetUpdatedAt = "2026-01-01T00:00:00Z"
	record.TargetFingerprint = applyTargetFingerprint(t, snapshot.Vehicles[0])
	record.ProposedAction = "replace"
	record.Classification = "warning"
	job := createApplyJob(t, repository, "sha-replace", []application.DataTransferPreviewRecord{record})
	if err := repository.ReplaceIssues(t.Context(), job.ID, []application.DataTransferIssue{{
		JobID: job.ID, Area: record.Area, RecordKey: record.RecordKey,
		Severity: application.TransferIssueWarning, Code: "duplicate_inventory_number", SelectedResolution: "replace",
	}}); err != nil {
		t.Fatal(err)
	}
	if err := repository.ApplyImport(t.Context(), job, "editor-1"); err != nil {
		t.Fatal(err)
	}
	var manufacturer, name string
	if err := db.QueryRow(`SELECT manufacturer, name FROM vehicles WHERE id='vehicle-replace'`).Scan(
		&manufacturer, &name,
	); err != nil {
		t.Fatal(err)
	}
	if manufacturer != "Piko" || name != "RK-REPLACE" {
		t.Fatalf("replaced vehicle = %q %q", manufacturer, name)
	}
}

func TestDataTransferApplyCopiesAccessoryAssetInventoryNumbersSafely(t *testing.T) {
	db := testDB(t)
	repository := infrastructure.NewDataTransferRepository(db)
	if _, err := db.Exec(`INSERT INTO accessory_products(
		id, inventory_number, manufacturer, name, category, tracking_mode, manufacturer_status, article_type,
		package_quantity, stock_unit, minimum_stock, inventory_strategy, created_at, updated_at
	) VALUES('product-existing', 'RK-ART-COPY', 'Viessmann', 'Signal', 'Signal', 'individual', 'unknown',
		'other', 1, 'piece', 0, 'individual', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z')`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO accessory_assets(
		id, product_id, inventory_number, serial_number, condition_state, lifecycle_state, created_at, updated_at
	) VALUES('asset-existing', 'product-existing', 'RK-ASSET-COPY', '', 'ready', 'stored',
		'2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z')`); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(application.TransferAccessory{
		InventoryNumber: "RK-ART-COPY", Manufacturer: "Viessmann", Name: "Signal", Category: "Signal",
		TrackingMode: "individual", ManufacturerStatus: "unknown", ArticleType: "other", PackageQuantity: 1,
		StockUnit: "piece", InventoryStrategy: "individual", Assets: []application.TransferAccessoryAsset{{
			InventoryNumber: "RK-ASSET-COPY", Condition: "ready", Lifecycle: "stored",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := repository.Snapshot(t.Context(), []application.TransferArea{application.TransferAccessories})
	if err != nil {
		t.Fatal(err)
	}
	record := application.DataTransferPreviewRecord{
		Area: application.TransferAccessories, RecordKey: "RK-ART-COPY", Classification: "warning",
		ProposedAction: "replace", TargetID: "product-existing", TargetUpdatedAt: "2026-01-01T00:00:00Z", Data: data,
	}
	record.TargetFingerprint = applyTargetFingerprint(t, snapshot.Accessories[0])
	job := createApplyJob(t, repository, "sha-accessory-copy", []application.DataTransferPreviewRecord{record})
	if err := repository.ReplaceIssues(t.Context(), job.ID, []application.DataTransferIssue{{
		JobID: job.ID, Area: record.Area, RecordKey: record.RecordKey, Severity: application.TransferIssueWarning,
		Code: "duplicate_inventory_number", SelectedResolution: "copy",
	}}); err != nil {
		t.Fatal(err)
	}
	if err := repository.ApplyImport(t.Context(), job, "editor-1"); err != nil {
		t.Fatal(err)
	}
	var products, assets int
	if err := db.QueryRow(`SELECT COUNT(*) FROM accessory_products`).Scan(&products); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM accessory_assets`).Scan(&assets); err != nil {
		t.Fatal(err)
	}
	if products != 2 || assets != 2 {
		t.Fatalf("copied accessory products=%d assets=%d", products, assets)
	}
}

func TestDataTransferApplyAccessoryReplacePreservesPurchaseAndRelationships(t *testing.T) {
	db := testDB(t)
	repository := infrastructure.NewDataTransferRepository(db)
	const timestamp = "2026-01-01T00:00:00Z"
	if _, err := db.Exec(`INSERT INTO storage_locations(id, name, created_at, updated_at)
		VALUES('location-merge', 'Merge shelf', ?, ?)`, timestamp, timestamp); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO vehicles(id, inventory_number, manufacturer, name, gauge, created_at, updated_at)
		VALUES('vehicle-merge', 'RK-MERGE-VEHICLE', 'Roco', 'BR 01', 'H0', ?, ?)`, timestamp, timestamp); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO accessory_products(
		id, inventory_number, manufacturer, name, category, tracking_mode, manufacturer_status, article_type,
		package_quantity, stock_unit, minimum_stock, inventory_strategy, created_at, updated_at
	) VALUES('product-merge', 'RK-ART-MERGE', 'Viessmann', 'Signal', 'Signal', 'individual', 'unknown',
		'other', 1, 'piece', 0, 'individual', ?, ?)`, timestamp, timestamp); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO accessory_purchases(
		id, product_id, destination_location_id, quantity, purchased_at, created_at, updated_at
	) VALUES('purchase-merge', 'product-merge', 'location-merge', 2, '2025-12-01', ?, ?)`,
		timestamp, timestamp); err != nil {
		t.Fatal(err)
	}
	for _, asset := range []struct {
		id, inventoryNumber, lifecycle, location string
	}{
		{"asset-reserved", "RK-ASSET-RES", "reserved", "location-merge"},
		{"asset-installed", "RK-ASSET-INST", "installed", ""},
		{"asset-local", "RK-ASSET-LOCAL", "stored", "location-merge"},
	} {
		if _, err := db.Exec(`INSERT INTO accessory_assets(
			id, product_id, purchase_id, inventory_number, condition_state, lifecycle_state, storage_location_id,
			created_at, updated_at
		) VALUES(?, 'product-merge', 'purchase-merge', ?, 'ready', ?, NULLIF(?, ''), ?, ?)`,
			asset.id, asset.inventoryNumber, asset.lifecycle, asset.location, timestamp, timestamp); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec(`INSERT INTO accessory_reservations(
		id, product_id, asset_id, location_id, quantity, vehicle_id, status, created_by, created_at, updated_at
	) VALUES('reservation-merge', 'product-merge', 'asset-reserved', 'location-merge', 1, 'vehicle-merge',
		'active', 'editor-1', ?, ?)`, timestamp, timestamp); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO accessory_installations(
		id, product_id, asset_id, source_location_id, quantity, vehicle_id, condition_state, installed_by,
		installed_at
	) VALUES('installation-merge', 'product-merge', 'asset-installed', 'location-merge', 1,
		'vehicle-merge', 'ready', 'editor-1', ?)`, timestamp); err != nil {
		t.Fatal(err)
	}
	snapshot, err := repository.Snapshot(t.Context(), []application.TransferArea{application.TransferAccessories})
	if err != nil {
		t.Fatal(err)
	}
	target := snapshot.Accessories[0]
	incoming := target
	incoming.Name = "Imported signal"
	incoming.Assets = []application.TransferAccessoryAsset{}
	for _, asset := range target.Assets {
		if asset.ID != "asset-local" {
			incoming.Assets = append(incoming.Assets, asset)
		}
	}
	for index := range incoming.Assets {
		incoming.Assets[index].Notes = "updated by import"
	}
	data, err := json.Marshal(incoming)
	if err != nil {
		t.Fatal(err)
	}
	record := application.DataTransferPreviewRecord{
		Area: application.TransferAccessories, RecordKey: target.InventoryNumber, Classification: "warning",
		ProposedAction: "replace", TargetID: target.ID, TargetUpdatedAt: target.UpdatedAt,
		TargetFingerprint: applyTargetFingerprint(t, target), Data: data,
	}
	job := createApplyJob(t, repository, "sha-accessory-merge", []application.DataTransferPreviewRecord{record})
	resolveApplyIssue(t, repository, job.ID, record, "duplicate_inventory_number", "replace")

	if err := repository.ApplyImport(t.Context(), job, "editor-1"); err != nil {
		t.Fatal(err)
	}
	for _, assetID := range []string{"asset-reserved", "asset-installed"} {
		var purchaseID, notes string
		if err := db.QueryRow(`SELECT COALESCE(purchase_id, ''), notes FROM accessory_assets WHERE id=?`,
			assetID).Scan(&purchaseID, &notes); err != nil {
			t.Fatal(err)
		}
		if purchaseID != "purchase-merge" || notes != "updated by import" {
			t.Fatalf("asset %s provenance changed: purchase=%q notes=%q", assetID, purchaseID, notes)
		}
	}
	for table, id := range map[string]string{
		"accessory_assets":        "asset-local",
		"accessory_reservations":  "reservation-merge",
		"accessory_installations": "installation-merge",
	} {
		var count int
		if err := db.QueryRow("SELECT COUNT(*) FROM "+table+" WHERE id=?", id).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 1 {
			t.Fatalf("%s relationship %s was not preserved", table, id)
		}
	}
	if _, err := db.Exec(`INSERT INTO storage_locations(id, name, created_at, updated_at)
		VALUES('location-other', 'Other shelf', ?, ?)`, timestamp, timestamp); err != nil {
		t.Fatal(err)
	}
	refreshed, err := repository.Snapshot(t.Context(), []application.TransferArea{application.TransferAccessories})
	if err != nil {
		t.Fatal(err)
	}
	refreshedFingerprint := applyTargetFingerprint(t, refreshed.Accessories[0])
	conflicting := refreshed.Accessories[0]
	conflicting.Assets = append([]application.TransferAccessoryAsset(nil), conflicting.Assets...)
	for index := range conflicting.Assets {
		if conflicting.Assets[index].ID == "asset-reserved" {
			conflicting.Assets[index].StorageLocationID = "location-other"
			conflicting.Assets[index].StorageLocationName = "Other shelf"
		}
	}
	conflictingData, err := json.Marshal(conflicting)
	if err != nil {
		t.Fatal(err)
	}
	conflictingRecord := application.DataTransferPreviewRecord{
		Area: application.TransferAccessories, RecordKey: conflicting.InventoryNumber, Classification: "warning",
		ProposedAction: "replace", TargetID: conflicting.ID, TargetUpdatedAt: conflicting.UpdatedAt,
		TargetFingerprint: refreshedFingerprint, Data: conflictingData,
	}
	conflictingJob := createApplyJob(t, repository, "sha-accessory-invariant", []application.DataTransferPreviewRecord{
		conflictingRecord,
	})
	resolveApplyIssue(t, repository, conflictingJob.ID, conflictingRecord, "duplicate_inventory_number", "replace")
	if err := repository.ApplyImport(t.Context(), conflictingJob, "editor-1"); !errors.Is(err, application.ErrDataTransferConflict) {
		t.Fatalf("ApplyImport() invariant error = %v, want conflict", err)
	}
	var reservedLocation string
	if err := db.QueryRow(`SELECT COALESCE(storage_location_id, '') FROM accessory_assets WHERE id='asset-reserved'`).
		Scan(&reservedLocation); err != nil {
		t.Fatal(err)
	}
	if reservedLocation != "location-merge" {
		t.Fatalf("failed invariant import changed reserved asset location to %q", reservedLocation)
	}
	latest, err := repository.Snapshot(t.Context(), []application.TransferArea{application.TransferAccessories})
	if err != nil {
		t.Fatal(err)
	}
	orphaned := latest.Accessories[0]
	orphaned.Assets = append(append([]application.TransferAccessoryAsset(nil), orphaned.Assets...),
		application.TransferAccessoryAsset{
			InventoryNumber: "RK-ASSET-ORPHAN", Condition: "ready", Lifecycle: "installed",
		})
	orphanedData, err := json.Marshal(orphaned)
	if err != nil {
		t.Fatal(err)
	}
	orphanedRecord := application.DataTransferPreviewRecord{
		Area: application.TransferAccessories, RecordKey: orphaned.InventoryNumber, Classification: "warning",
		ProposedAction: "replace", TargetID: orphaned.ID, TargetUpdatedAt: orphaned.UpdatedAt,
		TargetFingerprint: applyTargetFingerprint(t, latest.Accessories[0]), Data: orphanedData,
	}
	orphanedJob := createApplyJob(t, repository, "sha-accessory-orphan", []application.DataTransferPreviewRecord{
		orphanedRecord,
	})
	resolveApplyIssue(t, repository, orphanedJob.ID, orphanedRecord, "duplicate_inventory_number", "replace")
	if err := repository.ApplyImport(t.Context(), orphanedJob, "editor-1"); !errors.Is(err, application.ErrDataTransferConflict) {
		t.Fatalf("ApplyImport() orphan lifecycle error = %v, want conflict", err)
	}
	var orphanCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM accessory_assets WHERE inventory_number='RK-ASSET-ORPHAN'`).
		Scan(&orphanCount); err != nil {
		t.Fatal(err)
	}
	if orphanCount != 0 {
		t.Fatalf("failed invariant import created %d orphan installed assets", orphanCount)
	}
}

func TestDataTransferApplyUsesEachExhibitionReferenceResolutionOnce(t *testing.T) {
	db := testDB(t)
	repository := infrastructure.NewDataTransferRepository(db)
	for _, values := range [][]string{{"vehicle-a", "RK-LINK-A"}, {"vehicle-b", "RK-LINK-B"}} {
		if _, err := db.Exec(`INSERT INTO vehicles(id, inventory_number, manufacturer, name, gauge, created_at, updated_at)
			VALUES(?, ?, 'Roco', ?, 'H0', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z')`,
			values[0], values[1], values[1]); err != nil {
			t.Fatal(err)
		}
	}
	listData, err := json.Marshal(application.TransferExhibitionList{
		Designation: "References", Date: "2026-08-20", Entries: []application.TransferExhibitionEntry{
			{VehicleID: "foreign-a", VehicleInventoryNumber: "RK-LINK-A", Owner: "Club", LocomotiveName: "A", DayScope: "all", SortOrder: 1},
			{VehicleID: "foreign-b", VehicleInventoryNumber: "RK-LINK-B", Owner: "Club", LocomotiveName: "B", DayScope: "all", SortOrder: 2},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	record := application.DataTransferPreviewRecord{
		Area: application.TransferExhibitionLists, RecordKey: "References|2026-08-20",
		RowNumber: intPointer(1), Classification: "warning", ProposedAction: "create", Data: listData,
	}
	job := createApplyJob(t, repository, "sha-references", []application.DataTransferPreviewRecord{record})
	if err := repository.ReplaceIssues(t.Context(), job.ID, []application.DataTransferIssue{
		{ID: "issue-z", JobID: job.ID, Area: record.Area, RecordKey: record.RecordKey,
			RowNumber: record.RowNumber, Field: "entries[0].vehicleReference", Severity: application.TransferIssueWarning,
			Code: "exhibition_vehicle_reference", SelectedResolution: "link"},
		{ID: "issue-a", JobID: job.ID, Area: record.Area, RecordKey: record.RecordKey,
			RowNumber: record.RowNumber, Field: "entries[1].vehicleReference", Severity: application.TransferIssueWarning,
			Code: "exhibition_vehicle_reference", SelectedResolution: "skip"},
	}); err != nil {
		t.Fatal(err)
	}
	if err := repository.ApplyImport(t.Context(), job, "editor-1"); err != nil {
		t.Fatal(err)
	}
	rows, err := db.Query(`SELECT vehicle_id FROM exhibition_entries ORDER BY sort_order`)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	vehicleIDs := []string{}
	for rows.Next() {
		var vehicleID string
		if err := rows.Scan(&vehicleID); err != nil {
			t.Fatal(err)
		}
		vehicleIDs = append(vehicleIDs, vehicleID)
	}
	if len(vehicleIDs) != 2 || vehicleIDs[0] != "vehicle-a" || vehicleIDs[1] != "" {
		t.Fatalf("resolved exhibition vehicle IDs = %#v, want [vehicle-a empty]", vehicleIDs)
	}
}

func TestDataTransferApplyBindsDuplicateRecordResolutionsToRowNumber(t *testing.T) {
	db := testDB(t)
	repository := infrastructure.NewDataTransferRepository(db)
	const timestamp = "2026-01-01T00:00:00Z"
	if _, err := db.Exec(`INSERT INTO vehicles(id, inventory_number, manufacturer, name, gauge, created_at, updated_at)
		VALUES('vehicle-duplicate-row', 'RK-DUP-ROW', 'Roco', 'Original', 'H0', ?, ?)`, timestamp, timestamp); err != nil {
		t.Fatal(err)
	}
	snapshot, err := repository.Snapshot(t.Context(), []application.TransferArea{application.TransferVehicles})
	if err != nil {
		t.Fatal(err)
	}
	first := applyVehicleRecord(t, "RK-DUP-ROW", "Piko")
	first.RowNumber = intPointer(1)
	first.Classification = "warning"
	first.ProposedAction = "replace"
	first.TargetID = "vehicle-duplicate-row"
	first.TargetUpdatedAt = timestamp
	first.TargetFingerprint = applyTargetFingerprint(t, snapshot.Vehicles[0])
	second := first
	second.RowNumber = intPointer(2)
	job := createApplyJob(t, repository, "sha-duplicate-row", []application.DataTransferPreviewRecord{first, second})
	if err := repository.ReplaceIssues(t.Context(), job.ID, []application.DataTransferIssue{
		{ID: "row-one", JobID: job.ID, Area: first.Area, RecordKey: first.RecordKey, RowNumber: first.RowNumber,
			Severity: application.TransferIssueWarning, Code: "duplicate_inventory_number", SelectedResolution: "skip"},
		{ID: "row-two", JobID: job.ID, Area: second.Area, RecordKey: second.RecordKey, RowNumber: second.RowNumber,
			Severity: application.TransferIssueWarning, Code: "duplicate_inventory_number", SelectedResolution: "copy"},
	}); err != nil {
		t.Fatal(err)
	}

	if err := repository.ApplyImport(t.Context(), job, "editor-1"); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM vehicles WHERE inventory_number LIKE 'RK-DUP-ROW%'`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("row-scoped resolutions created %d vehicles, want original plus one copy", count)
	}
}

func createApplyJob(
	t *testing.T,
	repository *infrastructure.DataTransferRepository,
	sourceSHA string,
	records []application.DataTransferPreviewRecord,
) application.DataTransferJob {
	t.Helper()
	preview := map[string]any{"sourceSha256": sourceSHA, "records": records}
	areas := []application.TransferArea{}
	for _, record := range records {
		found := false
		for _, area := range areas {
			found = found || area == record.Area
		}
		if !found {
			areas = append(areas, record.Area)
		}
	}
	job, err := repository.CreateJob(t.Context(), application.DataTransferJob{
		ProfileName: "Apply test", Direction: application.TransferImport, Format: application.TransferJSON,
		Areas: areas, State: application.TransferJobReady,
		Stage: "preview", SourceName: "test.json", SourceSHA256: sourceSHA,
		PackageVersion: application.DataTransferPackageVersion, TotalRecords: len(records), ReadyRecords: len(records),
		Preview: preview, CreatedByUserID: "editor-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	return job
}

func applyVehicleRecord(t *testing.T, inventoryNumber, manufacturer string) application.DataTransferPreviewRecord {
	t.Helper()
	data, err := json.Marshal(application.TransferVehicle{
		InventoryNumber: inventoryNumber, Manufacturer: manufacturer, Name: inventoryNumber, Gauge: "H0",
	})
	if err != nil {
		t.Fatal(err)
	}
	return application.DataTransferPreviewRecord{
		Area: application.TransferVehicles, RecordKey: inventoryNumber, Classification: "ready",
		ProposedAction: "create", Data: data,
	}
}

func assertApplyJobStillReady(
	t *testing.T,
	repository *infrastructure.DataTransferRepository,
	jobID string,
) {
	t.Helper()
	job, err := repository.GetJob(t.Context(), jobID)
	if err != nil {
		t.Fatal(err)
	}
	if job.State != application.TransferJobReady || job.ConfirmedAt != "" || job.CompletedAt != "" {
		t.Fatalf("job changed after failed apply: %#v", job)
	}
}

func applyTargetFingerprint(t *testing.T, value any) string {
	t.Helper()
	payload, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(payload)
	return hex.EncodeToString(digest[:])
}

func resolveApplyIssue(
	t *testing.T,
	repository *infrastructure.DataTransferRepository,
	jobID string,
	record application.DataTransferPreviewRecord,
	code string,
	resolution string,
) {
	t.Helper()
	if err := repository.ReplaceIssues(t.Context(), jobID, []application.DataTransferIssue{{
		JobID: jobID, Area: record.Area, RecordKey: record.RecordKey, RowNumber: record.RowNumber,
		Severity: application.TransferIssueWarning, Code: code, SelectedResolution: resolution,
	}}); err != nil {
		t.Fatal(err)
	}
}

func insertApplyAccessoryProduct(t *testing.T, db *sql.DB, id, inventoryNumber, timestamp string) {
	t.Helper()
	if _, err := db.Exec(`INSERT INTO accessory_products(
		id, inventory_number, manufacturer, name, category, tracking_mode, manufacturer_status, article_type,
		package_quantity, stock_unit, minimum_stock, inventory_strategy, created_at, updated_at
	) VALUES(?, ?, 'Viessmann', 'Signal', 'Signal', 'quantity', 'unknown', 'other', 1, 'piece', 0,
		'quantity', ?, ?)`, id, inventoryNumber, timestamp, timestamp); err != nil {
		t.Fatal(err)
	}
}

func intPointer(value int) *int {
	return &value
}
