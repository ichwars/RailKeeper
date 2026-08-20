package infrastructure_test

import (
	"database/sql"
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
				record.ProposedAction = "replace"
				record.TargetID = "vehicle-existing"
				record.TargetUpdatedAt = "2026-01-01T00:00:00Z"
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
	record.TargetID = "vehicle-replace"
	record.TargetUpdatedAt = "2026-01-01T00:00:00Z"
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
	record := application.DataTransferPreviewRecord{
		Area: application.TransferAccessories, RecordKey: "RK-ART-COPY", Classification: "warning",
		ProposedAction: "replace", TargetID: "product-existing", TargetUpdatedAt: "2026-01-01T00:00:00Z", Data: data,
	}
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
		Classification: "warning", ProposedAction: "create", Data: listData,
	}
	job := createApplyJob(t, repository, "sha-references", []application.DataTransferPreviewRecord{record})
	if err := repository.ReplaceIssues(t.Context(), job.ID, []application.DataTransferIssue{
		{ID: "issue-1", JobID: job.ID, Area: record.Area, RecordKey: record.RecordKey,
			Severity: application.TransferIssueWarning, Code: "exhibition_vehicle_reference", SelectedResolution: "link"},
		{ID: "issue-2", JobID: job.ID, Area: record.Area, RecordKey: record.RecordKey,
			Severity: application.TransferIssueWarning, Code: "exhibition_vehicle_reference", SelectedResolution: "skip"},
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
