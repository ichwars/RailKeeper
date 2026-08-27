package infrastructure_test

import (
	"errors"
	"path/filepath"
	"sync"
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/infrastructure"
)

func TestDataTransferConcurrentExportClaimsJobOnce(t *testing.T) {
	db := testDB(t)
	repository := infrastructure.NewDataTransferRepository(db)
	dataDir := t.TempDir()
	service := application.NewDataTransferService(repository, dataDir)
	profile, err := service.CreateProfile(t.Context(), application.CreateDataTransferProfileInput{
		Name: "Vehicles", Direction: application.TransferExport, Format: application.TransferCSV,
		Areas: []application.TransferArea{application.TransferVehicles},
	}, "viewer-1")
	if err != nil {
		t.Fatal(err)
	}
	job, err := service.CreateExportJob(t.Context(), profile.ID, "viewer-1")
	if err != nil {
		t.Fatal(err)
	}

	type execution struct {
		result application.DataTransferExportResult
		err    error
	}
	start := make(chan struct{})
	results := make(chan execution, 2)
	var wait sync.WaitGroup
	for range 2 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			<-start
			result, err := service.ExecuteExport(t.Context(), job.ID)
			results <- execution{result: result, err: err}
		}()
	}
	close(start)
	wait.Wait()
	close(results)

	successes := 0
	conflicts := 0
	for execution := range results {
		switch {
		case execution.err == nil:
			successes++
		case errors.Is(execution.err, application.ErrDataTransferConflict):
			conflicts++
		default:
			t.Fatalf("concurrent export returned unexpected error: %v", execution.err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("concurrent executions: successes=%d conflicts=%d", successes, conflicts)
	}
	var artifactCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM data_transfer_artifacts WHERE job_id=?`, job.ID).
		Scan(&artifactCount); err != nil {
		t.Fatal(err)
	}
	if artifactCount != 1 {
		t.Fatalf("artifact rows = %d, want 1", artifactCount)
	}
	artifacts, err := filepath.Glob(filepath.Join(dataDir, "exports", "*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(artifacts) != 1 {
		t.Fatalf("published artifacts = %v, want exactly one", artifacts)
	}
}

func TestDataTransferSnapshotReadsSelectedAreasInStableOrder(t *testing.T) {
	db := testDB(t)
	for _, statement := range []string{
		`INSERT INTO vehicles(id, inventory_number, manufacturer, name, gauge, created_at, updated_at)
         VALUES('v-2', 'RK-010', 'Roco', 'Zweite', 'H0', '2026-08-20T10:00:00Z', '2026-08-20T10:00:00Z')`,
		`INSERT INTO vehicles(id, inventory_number, manufacturer, name, gauge, created_at, updated_at)
         VALUES('v-1', 'RK-002', 'Märklin', 'Erste', 'H0', '2026-08-20T10:00:00Z', '2026-08-20T10:00:00Z')`,
		`INSERT INTO exhibition_lists(id, designation, list_date, locked, created_at, updated_at)
         VALUES('list-2', 'Zweite', '2026-08-21', 0, '2026-08-20T10:00:00Z', '2026-08-20T10:00:00Z')`,
		`INSERT INTO exhibition_lists(
			id, designation, list_date, end_date, location, description, organization_notes,
			status, locked, lock_reason, locked_at, created_at, updated_at
		 ) VALUES(
			'list-1', 'Erste', '2026-08-20', '2026-08-22', 'Köln', 'Clubmesse', 'Aufbau Freitag',
			'locked', 1, 'Freigegeben', '2026-08-19T10:00:00Z',
			'2026-08-20T10:00:00Z', '2026-08-20T10:00:00Z'
		 )`,
		`INSERT INTO exhibition_entries(
			id, list_id, owner, locomotive_name, day_scope, interface_name, availability,
			created_at, updated_at
		 ) VALUES(
			'entry-1', 'list-1', 'Club', 'BR 103', 'day1,day2', 'ECoS', 'unavailable',
			'2026-08-20T10:00:00Z', '2026-08-20T10:00:00Z'
		 )`,
	} {
		if _, err := db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	repository := infrastructure.NewDataTransferRepository(db)
	snapshot, err := repository.Snapshot(t.Context(), []application.TransferArea{
		application.TransferVehicles, application.TransferExhibitionLists,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Vehicles) != 2 || snapshot.Vehicles[0].InventoryNumber != "RK-002" {
		t.Fatalf("vehicles are not stable: %#v", snapshot.Vehicles)
	}
	if snapshot.Accessories != nil {
		t.Fatalf("unselected accessories were queried: %#v", snapshot.Accessories)
	}
	if len(snapshot.ExhibitionLists) != 2 || snapshot.ExhibitionLists[0].ID != "list-1" {
		t.Fatalf("exhibition lists are not stable: %#v", snapshot.ExhibitionLists)
	}
	list := snapshot.ExhibitionLists[0]
	if list.EndDate != "2026-08-22" || list.Location != "Köln" || list.Description != "Clubmesse" ||
		list.OrganizationNotes != "Aufbau Freitag" || list.Status != application.ExhibitionStatusLocked ||
		list.LockReason != "Freigegeben" || list.LockedAt != "2026-08-19T10:00:00Z" {
		t.Fatalf("exhibition metadata was not preserved: %#v", list)
	}
	if len(list.Entries) != 1 || list.Entries[0].InterfaceName != "ECoS" ||
		list.Entries[0].Availability != "unavailable" {
		t.Fatalf("exhibition entry workspace data was not preserved: %#v", list.Entries)
	}
}

func TestDataTransferSnapshotIncludesExtendedVehicleFields(t *testing.T) {
	db := testDB(t)
	if _, err := db.Exec(`
INSERT INTO vehicles(
  id, inventory_number, manufacturer, name, gauge, category, gattung, length_mm, weight_g,
  coupling_same, coupling_front, coupling_rear, drive_enabled, drive_description,
  sound_generator_enabled, sound_generator_description, additional_info, qr_code_enabled,
  created_at, updated_at
) VALUES(
  'vehicle-full', 'RK-FULL', 'Roco', 'BR 218', 'H0', 'Lokomotive', 'Diesellokomotive', '181', '540',
  1, 'KKK', 'KKK', 1, 'Kardan', 1, 'Diesel', 'Clubbestand', 1,
  '2026-08-20T10:00:00Z', '2026-08-20T10:00:00Z'
)`); err != nil {
		t.Fatal(err)
	}
	repository := infrastructure.NewDataTransferRepository(db)
	snapshot, err := repository.Snapshot(t.Context(), []application.TransferArea{application.TransferVehicles})
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Vehicles) != 1 {
		t.Fatalf("snapshot vehicles = %d, want 1", len(snapshot.Vehicles))
	}
	vehicle := snapshot.Vehicles[0]
	if vehicle.LengthMM != "181" || vehicle.WeightG != "540" || !vehicle.CouplingSame ||
		vehicle.CouplingRear != "KKK" || !vehicle.DriveEnabled || vehicle.DriveDescription != "Kardan" ||
		!vehicle.SoundGeneratorEnabled || vehicle.AdditionalInfo != "Clubbestand" || !vehicle.QRCodeEnabled {
		t.Fatalf("snapshot lost extended vehicle fields: %#v", vehicle)
	}
}

func TestDataTransferSnapshotIncludesVehicleSets(t *testing.T) {
	db := testDB(t)
	for _, statement := range []string{
		`INSERT INTO vehicles(id, inventory_number, manufacturer, name, gauge, category, gattung, created_at, updated_at)
         VALUES('vehicle-a', 'RK-A', 'Roco', 'Wagen A', 'H0', 'Wagen', 'Reisezugwagen',
           '2026-08-27T10:00:00Z', '2026-08-27T10:00:00Z')`,
		`INSERT INTO vehicles(id, inventory_number, manufacturer, name, gauge, category, gattung, created_at, updated_at)
         VALUES('vehicle-b', 'RK-B', 'Roco', 'Wagen B', 'H0', 'Wagen', 'Reisezugwagen',
           '2026-08-27T10:00:00Z', '2026-08-27T10:00:00Z')`,
		`INSERT INTO vehicle_sets(
           id, inventory_number, name, manufacturer, article_number, article_source_url, gauge, epoch,
           railway_company, category, gattung, description, ean, production_period, list_price,
           acquisition_type, acquired_from, purchase_price, purchase_date, storage_location,
           storage_details, condition, condition_details, packaging, created_at, updated_at
         ) VALUES(
           'set-1', 'Set-001', 'Rheingold', 'Roco', '43000', 'https://example.invalid/43000', 'H0', 'III',
           'DB', 'Set', 'Reisezug', 'Komplettzug', '4000000000001', '2020', '499.00',
           'purchase', 'Fachhandel', '450.00', '2026-08-01', 'Vitrine', 'Fach 2',
           'very_good', 'Vollständig', 'original', '2026-08-27T10:00:00Z', '2026-08-27T10:00:00Z'
         )`,
		`INSERT INTO vehicle_set_members(vehicle_set_id, vehicle_id, position, label)
         VALUES('set-1', 'vehicle-b', 2, 'Speisewagen')`,
		`INSERT INTO vehicle_set_members(vehicle_set_id, vehicle_id, position, label)
         VALUES('set-1', 'vehicle-a', 1, 'Steuerwagen')`,
	} {
		if _, err := db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	repository := infrastructure.NewDataTransferRepository(db)
	snapshot, err := repository.Snapshot(t.Context(), []application.TransferArea{application.TransferVehicles})
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.VehicleSets) != 1 || len(snapshot.VehicleSets[0].Members) != 2 {
		t.Fatalf("vehicle sets = %#v", snapshot.VehicleSets)
	}
	set := snapshot.VehicleSets[0]
	if set.InventoryNumber != "Set-001" || set.Name != "Rheingold" || set.ArticleNumber != "43000" ||
		set.StorageLocation != "Vitrine" {
		t.Fatalf("vehicle set metadata = %#v", set)
	}
	if set.Members[0].Position != 1 || set.Members[0].SourceVehicleID != "vehicle-a" ||
		set.Members[0].VehicleInventoryNumber != "RK-A" || set.Members[0].Label != "Steuerwagen" ||
		set.Members[1].Position != 2 {
		t.Fatalf("ordered members = %#v", set.Members)
	}
}

func TestDataTransferSnapshotNormalizesReachableVehicleSetMemberships(t *testing.T) {
	db := testDB(t)
	for _, statement := range []string{
		`INSERT INTO vehicles(id, inventory_number, manufacturer, name, gauge, category, gattung, created_at, updated_at)
         VALUES('vehicle-a', 'RK-A', 'Roco', 'Wagen A', 'H0', 'Wagen', 'Reisezugwagen',
           '2026-08-27T10:00:00Z', '2026-08-27T10:00:00Z')`,
		`INSERT INTO vehicles(id, inventory_number, manufacturer, name, gauge, category, gattung, created_at, updated_at)
         VALUES('vehicle-c', 'RK-C', 'Roco', 'Wagen C', 'H0', 'Wagen', 'Reisezugwagen',
           '2026-08-27T10:00:00Z', '2026-08-27T10:00:00Z')`,
		`INSERT INTO vehicles(id, inventory_number, manufacturer, name, gauge, category, gattung, created_at, updated_at)
         VALUES('vehicle-d', 'RK-D', 'Roco', 'Wagen D', 'H0', 'Wagen', 'Reisezugwagen',
           '2026-08-27T10:00:00Z', '2026-08-27T10:00:00Z')`,
		`INSERT INTO vehicle_sets(id, inventory_number, name, manufacturer, gauge, category, gattung, created_at, updated_at)
         VALUES('set-gap', 'Set-Gap', 'Set mit Lücke', 'Roco', 'H0', 'Set', 'Reisezug',
           '2026-08-27T10:00:00Z', '2026-08-27T10:00:00Z')`,
		`INSERT INTO vehicle_sets(id, inventory_number, name, manufacturer, gauge, category, gattung, created_at, updated_at)
         VALUES('set-single', 'Set-Single', 'Einzelrest', 'Roco', 'H0', 'Set', 'Reisezug',
           '2026-08-27T10:00:00Z', '2026-08-27T10:00:00Z')`,
		`INSERT INTO vehicle_set_members(vehicle_set_id, vehicle_id, position, label)
         VALUES('set-gap', 'vehicle-a', 1, 'A')`,
		`INSERT INTO vehicle_set_members(vehicle_set_id, vehicle_id, position, label)
         VALUES('set-gap', 'vehicle-c', 3, 'C')`,
		`INSERT INTO vehicle_set_members(vehicle_set_id, vehicle_id, position, label)
         VALUES('set-single', 'vehicle-d', 2, 'D')`,
	} {
		if _, err := db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	repository := infrastructure.NewDataTransferRepository(db)
	snapshot, err := repository.Snapshot(t.Context(), []application.TransferArea{application.TransferVehicles})
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.VehicleSets) != 1 || snapshot.VehicleSets[0].ID != "set-gap" {
		t.Fatalf("exportable vehicle sets = %#v", snapshot.VehicleSets)
	}
	set := snapshot.VehicleSets[0]
	if len(set.Members) != 2 || set.Members[0].Position != 1 || set.Members[1].Position != 2 {
		t.Fatalf("normalized vehicle set members = %#v", set.Members)
	}
	if err := application.ValidateTransferVehicleSet(set); err != nil {
		t.Fatalf("normalized vehicle set is not importable: %v", err)
	}
}

func TestDataTransferSnapshotAccessoryIncludesCurrentStateOnly(t *testing.T) {
	db := testDB(t)
	for _, statement := range []string{
		`INSERT INTO storage_locations(id, name, created_at, updated_at)
         VALUES('location-1', 'Schrank', '2026-08-20T10:00:00Z', '2026-08-20T10:00:00Z')`,
		`INSERT INTO accessory_products(id, inventory_number, manufacturer, article_number, name, category,
           tracking_mode, inventory_strategy, created_at, updated_at)
         VALUES('product-1', 'RK-ART-1', 'Viessmann', '4011', 'Signal', 'signal', 'individual', 'individual',
           '2026-08-20T10:00:00Z', '2026-08-20T10:00:00Z')`,
		`INSERT INTO accessory_stock(product_id, location_id, quantity, updated_at)
         VALUES('product-1', 'location-1', 2, '2026-08-20T10:00:00Z')`,
		`INSERT INTO accessory_assets(id, product_id, inventory_number, serial_number, condition_state,
           lifecycle_state, storage_location_id, created_at, updated_at)
         VALUES('asset-1', 'product-1', 'RK-ASSET-1', 'S1', 'ready', 'stored', 'location-1',
           '2026-08-20T10:00:00Z', '2026-08-20T10:00:00Z')`,
	} {
		if _, err := db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	repository := infrastructure.NewDataTransferRepository(db)
	snapshot, err := repository.Snapshot(t.Context(), []application.TransferArea{application.TransferAccessories})
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Accessories) != 1 || len(snapshot.Accessories[0].Stock) != 1 ||
		len(snapshot.Accessories[0].Assets) != 1 {
		t.Fatalf("unexpected accessory snapshot: %#v", snapshot.Accessories)
	}
	if snapshot.Accessories[0].Stock[0].LocationName != "Schrank" ||
		snapshot.Accessories[0].Assets[0].InventoryNumber != "RK-ASSET-1" {
		t.Fatalf("unexpected accessory current state: %#v", snapshot.Accessories[0])
	}
}
