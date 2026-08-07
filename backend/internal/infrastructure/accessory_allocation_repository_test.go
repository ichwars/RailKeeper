package infrastructure_test

import (
	"database/sql"
	"errors"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/domain"
	"railkeeper/backend/internal/infrastructure"
)

type allocationFixture struct {
	db                *sql.DB
	accessories       *application.AccessoryService
	allocations       *application.AccessoryAllocationService
	location          *application.StorageLocation
	quantityProduct   *application.AccessoryProduct
	individualProduct *application.AccessoryProduct
	asset             *application.AccessoryAsset
	layout            *application.Layout
	unit              *application.LayoutUnit
	vehicle           *application.Vehicle
}

func TestAccessoryAllocationsReserveAndInstallQuantityAtomically(t *testing.T) {
	fixture := newAllocationFixture(t)
	ctx := t.Context()
	target := application.AllocationTargetInput{LayoutID: fixture.layout.ID}

	reservation, err := fixture.allocations.CreateReservation(ctx, application.CreateAccessoryReservationInput{
		ProductID: fixture.quantityProduct.ID, LocationID: fixture.location.ID, Quantity: 3,
		AllocationTargetInput: target, Note: "Bauabschnitt West", Placement: "Signalbrücke",
		DigitalAddress: "42", DecoderOutput: "A1", Connection: "Klemme 3", WiringNotes: "blau",
	}, "planner-1")
	if err != nil {
		t.Fatal(err)
	}
	stock, err := fixture.accessories.GetStock(ctx, fixture.quantityProduct.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stock.TotalQuantity != 5 {
		t.Fatalf("reservation changed physical stock: %#v", stock)
	}
	assertAllocationSummary(t, fixture.allocations, fixture.quantityProduct.ID, application.AccessoryAllocationSummary{
		Owned: 5, Stored: 5, Reserved: 3, Installed: 0, Available: 2, Missing: 0,
	})
	if _, err := fixture.allocations.CreateReservation(ctx, application.CreateAccessoryReservationInput{
		ProductID: fixture.quantityProduct.ID, LocationID: fixture.location.ID, Quantity: 3,
		AllocationTargetInput: target,
	}, "planner-1"); !errors.Is(err, application.ErrAccessoryInsufficientStock) {
		t.Fatalf("expected over-reservation rejection, got %v", err)
	}

	installation, err := fixture.allocations.Install(ctx, application.CreateAccessoryInstallationInput{
		ReservationID: reservation.ID, ProductID: fixture.quantityProduct.ID,
		SourceLocationID: fixture.location.ID, Quantity: 3, AllocationTargetInput: target,
		Condition: domain.AccessoryConditionReady, Placement: "Signalbrücke montiert",
		DigitalAddress: "43", DecoderOutput: "A2", Connection: "Klemme 4", WiringNotes: "gelb",
	}, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	reservations, err := fixture.allocations.ListReservations(ctx, fixture.quantityProduct.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(reservations) != 1 || reservations[0].Status != domain.AccessoryReservationFulfilled {
		t.Fatalf("reservation was not fulfilled: %#v", reservations)
	}
	if reservations[0].Placement != "Signalbrücke" || reservations[0].DigitalAddress != "42" ||
		reservations[0].DecoderOutput != "A1" || reservations[0].Connection != "Klemme 3" ||
		reservations[0].WiringNotes != "blau" {
		t.Fatalf("reservation technical data was not retained: %#v", reservations[0])
	}
	if installation.Placement != "Signalbrücke montiert" || installation.DigitalAddress != "43" ||
		installation.DecoderOutput != "A2" || installation.Connection != "Klemme 4" ||
		installation.WiringNotes != "gelb" {
		t.Fatalf("installation technical data was not retained: %#v", installation)
	}
	if _, err := fixture.allocations.UpdateInstallationCondition(ctx, installation.ID,
		application.UpdateAccessoryInstallationConditionInput{Condition: domain.AccessoryConditionMaintenanceDue},
		"editor-2"); err != nil {
		t.Fatal(err)
	}
	assertAllocationSummary(t, fixture.allocations, fixture.quantityProduct.ID, application.AccessoryAllocationSummary{
		Owned: 5, Stored: 2, Reserved: 0, Installed: 3, Available: 2, Missing: 0,
	})

	removed, err := fixture.allocations.RemoveInstallation(ctx, installation.ID,
		application.RemoveAccessoryInstallationInput{
			Disposition: domain.AccessoryRemovalStored, StorageLocationID: fixture.location.ID,
			Notes: "nach Prüfung eingelagert",
		}, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	if removed.RemovalDisposition != domain.AccessoryRemovalStored || removed.RemovedAt == "" {
		t.Fatalf("installation was not closed: %#v", removed)
	}
	if _, err := fixture.db.ExecContext(ctx, `UPDATE accessory_reservations SET created_at='2026-01-01T10:00:00Z'
WHERE id=?`, reservation.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.db.ExecContext(ctx, `UPDATE accessory_installations
SET installed_at='2026-01-01T11:00:00Z', removed_at='2026-01-01T13:00:00Z' WHERE id=?`, installation.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.db.ExecContext(ctx, `UPDATE accessory_installation_condition_history
SET changed_at='2026-01-01T12:00:00Z' WHERE installation_id=?`, installation.ID); err != nil {
		t.Fatal(err)
	}
	history, err := fixture.allocations.GetUsageHistory(ctx, fixture.quantityProduct.ID)
	if err != nil {
		t.Fatal(err)
	}
	wantTypes := []application.AccessoryUsageEventType{
		application.AccessoryUsageRemoval,
		application.AccessoryUsageConditionChanged,
		application.AccessoryUsageInstallation,
		application.AccessoryUsageReservation,
	}
	if len(history.Events) != len(wantTypes) {
		t.Fatalf("unexpected usage history: %#v", history.Events)
	}
	for index, want := range wantTypes {
		if history.Events[index].Type != want {
			t.Fatalf("usage event %d: got %q, want %q (%#v)", index, history.Events[index].Type, want, history.Events)
		}
	}
	if history.Events[1].PreviousCondition != domain.AccessoryConditionReady ||
		history.Events[1].Condition != domain.AccessoryConditionMaintenanceDue ||
		history.Events[1].Actor != "editor-2" {
		t.Fatalf("condition history lost domain data: %#v", history.Events[1])
	}
	assertAllocationSummary(t, fixture.allocations, fixture.quantityProduct.ID, application.AccessoryAllocationSummary{
		Owned: 5, Stored: 5, Reserved: 0, Installed: 0, Available: 5, Missing: 0,
	})
	if _, err := fixture.allocations.RemoveInstallation(ctx, installation.ID,
		application.RemoveAccessoryInstallationInput{
			Disposition: domain.AccessoryRemovalStored, StorageLocationID: fixture.location.ID,
		}, "editor-1"); !errors.Is(err, application.ErrAccessoryConflict) {
		t.Fatalf("expected immutable closed installation, got %v", err)
	}
}

func TestAccessoryAllocationsTrackIndividualAssetLifecycle(t *testing.T) {
	fixture := newAllocationFixture(t)
	ctx := t.Context()
	target := application.AllocationTargetInput{VehicleID: fixture.vehicle.ID}

	reservation, err := fixture.allocations.CreateReservation(ctx, application.CreateAccessoryReservationInput{
		ProductID: fixture.individualProduct.ID, AssetID: fixture.asset.ID,
		LocationID: fixture.location.ID, Quantity: 1, AllocationTargetInput: target,
	}, "planner-1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.accessories.UpdateAsset(ctx, fixture.asset.ID, application.UpdateAccessoryAssetInput{
		CreateAccessoryAssetInput: application.CreateAccessoryAssetInput{
			InventoryNumber: fixture.asset.InventoryNumber, Condition: domain.AccessoryConditionReady,
			Lifecycle: domain.AccessoryLifecycleStored, StorageLocationID: fixture.location.ID,
		},
	}, "editor-1"); !errors.Is(err, application.ErrAccessoryConflict) {
		t.Fatalf("expected generic update of reserved asset to fail, got %v", err)
	}
	assertAssetState(t, fixture.accessories, fixture.individualProduct.ID, fixture.asset.ID,
		domain.AccessoryLifecycleReserved, domain.AccessoryConditionReady, fixture.location.ID)
	if _, err := fixture.allocations.CreateReservation(ctx, application.CreateAccessoryReservationInput{
		ProductID: fixture.individualProduct.ID, AssetID: fixture.asset.ID,
		LocationID: fixture.location.ID, Quantity: 1, AllocationTargetInput: target,
	}, "planner-1"); !errors.Is(err, application.ErrAccessoryConflict) {
		t.Fatalf("expected duplicate asset reservation conflict, got %v", err)
	}

	cancelled, err := fixture.allocations.CancelReservation(ctx, reservation.ID, "planner-1")
	if err != nil {
		t.Fatal(err)
	}
	if cancelled.Status != domain.AccessoryReservationCancelled {
		t.Fatalf("unexpected cancelled reservation: %#v", cancelled)
	}
	assertAssetState(t, fixture.accessories, fixture.individualProduct.ID, fixture.asset.ID,
		domain.AccessoryLifecycleStored, domain.AccessoryConditionReady, fixture.location.ID)

	reservation, err = fixture.allocations.CreateReservation(ctx, application.CreateAccessoryReservationInput{
		ProductID: fixture.individualProduct.ID, AssetID: fixture.asset.ID,
		LocationID: fixture.location.ID, Quantity: 1, AllocationTargetInput: target,
	}, "planner-1")
	if err != nil {
		t.Fatal(err)
	}
	installation, err := fixture.allocations.Install(ctx, application.CreateAccessoryInstallationInput{
		ReservationID: reservation.ID, ProductID: fixture.individualProduct.ID, AssetID: fixture.asset.ID,
		SourceLocationID: fixture.location.ID, Quantity: 1, AllocationTargetInput: target,
		Condition: domain.AccessoryConditionReady, Notes: "Decoder eingebaut",
	}, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	assertAssetState(t, fixture.accessories, fixture.individualProduct.ID, fixture.asset.ID,
		domain.AccessoryLifecycleInstalled, domain.AccessoryConditionReady, "")
	if _, err := fixture.allocations.Install(ctx, application.CreateAccessoryInstallationInput{
		ProductID: fixture.individualProduct.ID, AssetID: fixture.asset.ID,
		SourceLocationID: fixture.location.ID, Quantity: 1,
		AllocationTargetInput: application.AllocationTargetInput{LayoutUnitID: fixture.unit.ID},
	}, "editor-1"); !errors.Is(err, application.ErrAccessoryConflict) {
		t.Fatalf("expected active asset installation conflict, got %v", err)
	}

	installation, err = fixture.allocations.UpdateInstallationCondition(ctx, installation.ID,
		application.UpdateAccessoryInstallationConditionInput{Condition: domain.AccessoryConditionDefective},
		"editor-1")
	if err != nil {
		t.Fatal(err)
	}
	if installation.Condition != domain.AccessoryConditionDefective {
		t.Fatalf("installation condition not updated: %#v", installation)
	}
	assertAssetState(t, fixture.accessories, fixture.individualProduct.ID, fixture.asset.ID,
		domain.AccessoryLifecycleInstalled, domain.AccessoryConditionDefective, "")

	removed, err := fixture.allocations.RemoveInstallation(ctx, installation.ID,
		application.RemoveAccessoryInstallationInput{
			Disposition: domain.AccessoryRemovalMaintenance, Notes: "Werkstattprüfung erforderlich",
		}, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	if removed.RemovalDisposition != domain.AccessoryRemovalMaintenance {
		t.Fatalf("unexpected removal disposition: %#v", removed)
	}
	assertAssetState(t, fixture.accessories, fixture.individualProduct.ID, fixture.asset.ID,
		domain.AccessoryLifecycleMaintenance, domain.AccessoryConditionDefective, "")
	assertAllocationSummary(t, fixture.allocations, fixture.individualProduct.ID,
		application.AccessoryAllocationSummary{
			Owned: 1, Stored: 0, Reserved: 0, Installed: 0, Available: 0, Missing: 0,
		})
	installations, err := fixture.allocations.ListInstallations(ctx, fixture.individualProduct.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(installations) != 1 || installations[0].RemovedAt == "" {
		t.Fatalf("installation history was not retained: %#v", installations)
	}
}

func TestAccessoryStockMovementsTrackQuantityInstallationAndRemoval(t *testing.T) {
	fixture := newAllocationFixture(t)
	ctx := t.Context()
	installation, err := fixture.allocations.Install(ctx, application.CreateAccessoryInstallationInput{
		ProductID: fixture.quantityProduct.ID, SourceLocationID: fixture.location.ID, Quantity: 2,
		AllocationTargetInput: application.AllocationTargetInput{LayoutID: fixture.layout.ID},
		Condition:             domain.AccessoryConditionReady,
	}, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.allocations.RemoveInstallation(ctx, installation.ID,
		application.RemoveAccessoryInstallationInput{
			Disposition: domain.AccessoryRemovalStored, StorageLocationID: fixture.location.ID,
		}, "editor-1"); err != nil {
		t.Fatal(err)
	}
	movements, err := fixture.accessories.ListStockMovements(ctx, fixture.quantityProduct.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(movements) != 3 {
		t.Fatalf("expected adjustment, installation, and removal movements: %#v", movements)
	}
	installationMovement := findAllocationMovement(t, movements, "installation")
	removalMovement := findAllocationMovement(t, movements, "removal")
	if installationMovement.Quantity != -2 || removalMovement.Quantity != 2 ||
		installationMovement.SourceType != "installation" || removalMovement.SourceType != "installation" ||
		installationMovement.SourceID != installation.ID || removalMovement.SourceID != installation.ID {
		t.Fatalf("unexpected physical movement journal: install=%#v removal=%#v",
			installationMovement, removalMovement)
	}

	individualInstallation, err := fixture.allocations.Install(ctx, application.CreateAccessoryInstallationInput{
		ProductID: fixture.individualProduct.ID, AssetID: fixture.asset.ID,
		SourceLocationID: fixture.location.ID, Quantity: 1,
		AllocationTargetInput: application.AllocationTargetInput{LayoutID: fixture.layout.ID},
		Condition:             domain.AccessoryConditionReady,
	}, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.allocations.RemoveInstallation(ctx, individualInstallation.ID,
		application.RemoveAccessoryInstallationInput{
			Disposition: domain.AccessoryRemovalStored, StorageLocationID: fixture.location.ID,
		}, "editor-1"); err != nil {
		t.Fatal(err)
	}
	individualMovements, err := fixture.accessories.ListStockMovements(ctx, fixture.individualProduct.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(individualMovements) != 0 {
		t.Fatalf("pure asset lifecycle transitions wrote quantity movements: %#v", individualMovements)
	}
}

func TestAccessoryHybridAllocationsUseAssetAndQuantityPaths(t *testing.T) {
	fixture := newAllocationFixture(t)
	ctx := t.Context()
	hybrid, err := fixture.accessories.CreateProduct(ctx, application.CreateAccessoryProductInput{
		Manufacturer: "ESU", Name: "Hybrid decoder", Category: "Other",
		ArticleType: domain.AccessoryArticleOther, Subtype: "other", PackageQuantity: 1,
		StockUnit: "piece", InventoryStrategy: domain.AccessoryInventoryQuantityLaterIndividual,
	}, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.accessories.AdjustStock(ctx, hybrid.ID, application.StockAdjustmentInput{
		LocationID: fixture.location.ID, Delta: 3,
	}, "editor-1"); err != nil {
		t.Fatal(err)
	}
	asset, err := fixture.accessories.Individualize(ctx, hybrid.ID, application.IndividualizeAccessoryInput{
		LocationID: fixture.location.ID,
		Asset:      application.CreateAccessoryAssetInput{InventoryNumber: "HYB-ALLOC-1"},
	}, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	assertAllocationSummary(t, fixture.allocations, hybrid.ID, application.AccessoryAllocationSummary{
		Owned: 3, Stored: 3, Available: 3,
	})

	assetReservation, err := fixture.allocations.CreateReservation(ctx,
		application.CreateAccessoryReservationInput{
			ProductID: hybrid.ID, AssetID: asset.ID, LocationID: fixture.location.ID, Quantity: 1,
			AllocationTargetInput: application.AllocationTargetInput{LayoutID: fixture.layout.ID},
		}, "planner-1")
	if err != nil {
		t.Fatal(err)
	}
	assertAllocationSummary(t, fixture.allocations, hybrid.ID, application.AccessoryAllocationSummary{
		Owned: 3, Stored: 3, Reserved: 1, Available: 2,
	})
	assetInstallation, err := fixture.allocations.Install(ctx, application.CreateAccessoryInstallationInput{
		ReservationID: assetReservation.ID, ProductID: hybrid.ID, AssetID: asset.ID,
		SourceLocationID: fixture.location.ID, Quantity: 1,
		AllocationTargetInput: application.AllocationTargetInput{LayoutID: fixture.layout.ID},
		Condition:             domain.AccessoryConditionReady,
	}, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	assertAllocationSummary(t, fixture.allocations, hybrid.ID, application.AccessoryAllocationSummary{
		Owned: 3, Stored: 2, Installed: 1, Available: 2,
	})
	if _, err := fixture.allocations.RemoveInstallation(ctx, assetInstallation.ID,
		application.RemoveAccessoryInstallationInput{
			Disposition: domain.AccessoryRemovalStored, StorageLocationID: fixture.location.ID,
		}, "editor-1"); err != nil {
		t.Fatal(err)
	}
	assertAllocationSummary(t, fixture.allocations, hybrid.ID, application.AccessoryAllocationSummary{
		Owned: 3, Stored: 3, Available: 3,
	})
	movements, err := fixture.accessories.ListStockMovements(ctx, hybrid.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(movements) != 2 {
		t.Fatalf("hybrid asset lifecycle wrote physical movements: %#v", movements)
	}

	quantityReservation, err := fixture.allocations.CreateReservation(ctx,
		application.CreateAccessoryReservationInput{
			ProductID: hybrid.ID, LocationID: fixture.location.ID, Quantity: 1,
			AllocationTargetInput: application.AllocationTargetInput{LayoutUnitID: fixture.unit.ID},
		}, "planner-1")
	if err != nil {
		t.Fatal(err)
	}
	quantityInstallation, err := fixture.allocations.Install(ctx, application.CreateAccessoryInstallationInput{
		ReservationID: quantityReservation.ID, ProductID: hybrid.ID,
		SourceLocationID: fixture.location.ID, Quantity: 1,
		AllocationTargetInput: application.AllocationTargetInput{LayoutUnitID: fixture.unit.ID},
		Condition:             domain.AccessoryConditionReady,
	}, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	assertAllocationSummary(t, fixture.allocations, hybrid.ID, application.AccessoryAllocationSummary{
		Owned: 3, Stored: 2, Installed: 1, Available: 2,
	})
	if _, err := fixture.allocations.RemoveInstallation(ctx, quantityInstallation.ID,
		application.RemoveAccessoryInstallationInput{
			Disposition: domain.AccessoryRemovalStored, StorageLocationID: fixture.location.ID,
		}, "editor-1"); err != nil {
		t.Fatal(err)
	}
	assertAllocationSummary(t, fixture.allocations, hybrid.ID, application.AccessoryAllocationSummary{
		Owned: 3, Stored: 3, Available: 3,
	})
	movements, err = fixture.accessories.ListStockMovements(ctx, hybrid.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(movements) != 4 || findAllocationMovement(t, movements, "installation").Quantity != -1 ||
		findAllocationMovement(t, movements, "removal").Quantity != 1 {
		t.Fatalf("hybrid quantity lifecycle did not journal physical movements: %#v", movements)
	}
}

func findAllocationMovement(
	t *testing.T,
	movements []application.AccessoryStockMovement,
	movementType string,
) application.AccessoryStockMovement {
	t.Helper()
	for _, movement := range movements {
		if movement.MovementType == movementType {
			return movement
		}
	}
	t.Fatalf("movement %q not found in %#v", movementType, movements)
	return application.AccessoryStockMovement{}
}

func TestAccessoryAllocationsValidateTargetsAndReservationMatch(t *testing.T) {
	fixture := newAllocationFixture(t)
	ctx := t.Context()
	if _, err := fixture.allocations.CreateReservation(ctx, application.CreateAccessoryReservationInput{
		ProductID: fixture.quantityProduct.ID, LocationID: fixture.location.ID, Quantity: 1,
		AllocationTargetInput: application.AllocationTargetInput{VehicleID: "missing"},
	}, "planner-1"); !errors.Is(err, application.ErrAccessoryNotFound) {
		t.Fatalf("expected missing target error, got %v", err)
	}
	reservation, err := fixture.allocations.CreateReservation(ctx, application.CreateAccessoryReservationInput{
		ProductID: fixture.quantityProduct.ID, LocationID: fixture.location.ID, Quantity: 1,
		AllocationTargetInput: application.AllocationTargetInput{LayoutID: fixture.layout.ID},
	}, "planner-1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.allocations.Install(ctx, application.CreateAccessoryInstallationInput{
		ReservationID: reservation.ID, ProductID: fixture.quantityProduct.ID,
		SourceLocationID: fixture.location.ID, Quantity: 1,
		AllocationTargetInput: application.AllocationTargetInput{LayoutUnitID: fixture.unit.ID},
	}, "editor-1"); !errors.Is(err, application.ErrAccessoryConflict) {
		t.Fatalf("expected mismatched reservation target conflict, got %v", err)
	}
	stock, err := fixture.accessories.GetStock(ctx, fixture.quantityProduct.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stock.TotalQuantity != 5 {
		t.Fatalf("failed installation changed stock: %#v", stock)
	}
}

func TestAccessoryAllocationsProtectReservationsFromDirectInstall(t *testing.T) {
	fixture := newAllocationFixture(t)
	ctx := t.Context()
	if _, err := fixture.allocations.CreateReservation(ctx, application.CreateAccessoryReservationInput{
		ProductID: fixture.quantityProduct.ID, LocationID: fixture.location.ID, Quantity: 4,
		AllocationTargetInput: application.AllocationTargetInput{LayoutID: fixture.layout.ID},
	}, "planner-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.allocations.Install(ctx, application.CreateAccessoryInstallationInput{
		ProductID: fixture.quantityProduct.ID, SourceLocationID: fixture.location.ID, Quantity: 2,
		AllocationTargetInput: application.AllocationTargetInput{LayoutUnitID: fixture.unit.ID},
	}, "editor-1"); !errors.Is(err, application.ErrAccessoryInsufficientStock) {
		t.Fatalf("expected reserved stock protection, got %v", err)
	}
	installation, err := fixture.allocations.Install(ctx, application.CreateAccessoryInstallationInput{
		ProductID: fixture.quantityProduct.ID, SourceLocationID: fixture.location.ID, Quantity: 1,
		AllocationTargetInput: application.AllocationTargetInput{LayoutUnitID: fixture.unit.ID},
	}, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	if installation.Quantity != 1 || installation.LayoutUnitID != fixture.unit.ID {
		t.Fatalf("unexpected direct installation: %#v", installation)
	}
	assertAllocationSummary(t, fixture.allocations, fixture.quantityProduct.ID,
		application.AccessoryAllocationSummary{
			Owned: 5, Stored: 4, Reserved: 4, Installed: 1, Available: 0, Missing: 0,
		})
}

func TestAccessoryAllocationsPreventTrackingModeChangeWithActiveInstallation(t *testing.T) {
	fixture := newAllocationFixture(t)
	ctx := t.Context()
	if _, err := fixture.allocations.Install(ctx, application.CreateAccessoryInstallationInput{
		ProductID: fixture.quantityProduct.ID, SourceLocationID: fixture.location.ID, Quantity: 5,
		AllocationTargetInput: application.AllocationTargetInput{LayoutID: fixture.layout.ID},
	}, "editor-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.accessories.UpdateProduct(ctx, fixture.quantityProduct.ID,
		application.UpdateAccessoryProductInput{CreateAccessoryProductInput: application.CreateAccessoryProductInput{
			Manufacturer:  fixture.quantityProduct.Manufacturer,
			ArticleNumber: fixture.quantityProduct.ArticleNumber,
			Name:          fixture.quantityProduct.Name,
			Category:      fixture.quantityProduct.Category,
			TrackingMode:  domain.AccessoryTrackingModeIndividual,
		}}, "editor-1"); !errors.Is(err, application.ErrAccessoryConflict) {
		t.Fatalf("expected active installation to block tracking mode change, got %v", err)
	}
}

func TestAccessoryAllocationsHandleIndividualRemovalDispositionsAndPrivateNotes(t *testing.T) {
	fixture := newAllocationFixture(t)
	ctx := t.Context()
	tests := []struct {
		name                string
		disposition         domain.AccessoryRemovalDisposition
		storageLocationID   string
		wantLifecycle       domain.AccessoryLifecycle
		wantCondition       domain.AccessoryCondition
		wantStorageLocation string
	}{
		{"stored", domain.AccessoryRemovalStored, fixture.location.ID,
			domain.AccessoryLifecycleStored, domain.AccessoryConditionReady, fixture.location.ID},
		{"defective", domain.AccessoryRemovalDefective, "",
			domain.AccessoryLifecycleMaintenance, domain.AccessoryConditionDefective, ""},
		{"retired", domain.AccessoryRemovalRetired, "",
			domain.AccessoryLifecycleRetired, domain.AccessoryConditionReady, ""},
	}
	for index, test := range tests {
		asset, err := fixture.accessories.CreateAsset(ctx, fixture.individualProduct.ID,
			application.CreateAccessoryAssetInput{
				InventoryNumber: fixture.asset.InventoryNumber + "-" + test.name,
				Condition:       domain.AccessoryConditionReady, Lifecycle: domain.AccessoryLifecycleStored,
				StorageLocationID: fixture.location.ID,
			}, "editor-1")
		if err != nil {
			t.Fatal(err)
		}
		installation, err := fixture.allocations.Install(ctx, application.CreateAccessoryInstallationInput{
			ProductID: fixture.individualProduct.ID, AssetID: asset.ID,
			SourceLocationID: fixture.location.ID, Quantity: 1,
			AllocationTargetInput: application.AllocationTargetInput{LayoutUnitID: fixture.unit.ID},
			Condition:             domain.AccessoryConditionReady, Notes: "private install note",
		}, "editor-1")
		if err != nil {
			t.Fatalf("case %d direct install: %v", index, err)
		}
		removed, err := fixture.allocations.RemoveInstallation(ctx, installation.ID,
			application.RemoveAccessoryInstallationInput{
				Disposition: test.disposition, StorageLocationID: test.storageLocationID,
				Notes: "private removal note",
			}, "editor-1")
		if err != nil {
			t.Fatalf("case %d removal: %v", index, err)
		}
		if removed.Notes != "private install note" || removed.RemovalNotes != "private removal note" {
			t.Fatalf("installation notes were not preserved separately: %#v", removed)
		}
		assertAssetState(t, fixture.accessories, fixture.individualProduct.ID, asset.ID,
			test.wantLifecycle, test.wantCondition, test.wantStorageLocation)
	}
	rows, err := fixture.db.QueryContext(ctx, `
SELECT details_json FROM audit_logs
WHERE action IN ('AccessoryInstalled', 'AccessoryRemoved')`)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var details string
		if err := rows.Scan(&details); err != nil {
			t.Fatal(err)
		}
		if strings.Contains(details, "private") {
			t.Fatalf("allocation audit leaked private notes: %s", details)
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
}

func newAllocationFixture(t *testing.T) allocationFixture {
	t.Helper()
	accessories, db := testAccessoryService(t)
	repository := infrastructure.NewAccessoryRepository(db)
	fixture := allocationFixture{
		db: db, accessories: accessories,
		allocations: application.NewAccessoryAllocationService(repository),
	}
	ctx := t.Context()
	var err error
	fixture.location, err = accessories.CreateLocation(ctx,
		application.CreateStorageLocationInput{Name: "Zentrallager"}, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	fixture.quantityProduct, err = accessories.CreateProduct(ctx, application.CreateAccessoryProductInput{
		Manufacturer: "Tillig", ArticleNumber: "83501", Name: "Schienenverbinder", Category: "Gleismaterial",
		TrackingMode: domain.AccessoryTrackingModeQuantity,
	}, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = accessories.AdjustStock(ctx, fixture.quantityProduct.ID, application.StockAdjustmentInput{
		LocationID: fixture.location.ID, Delta: 5,
	}, "editor-1"); err != nil {
		t.Fatal(err)
	}
	fixture.individualProduct, err = accessories.CreateProduct(ctx, application.CreateAccessoryProductInput{
		Manufacturer: "Lenz", ArticleNumber: "LS150", Name: "Schaltdecoder", Category: "Decoder",
		TrackingMode: domain.AccessoryTrackingModeIndividual,
	}, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	fixture.asset, err = accessories.CreateAsset(ctx, fixture.individualProduct.ID,
		application.CreateAccessoryAssetInput{
			InventoryNumber: "RK-Z-1001", Condition: domain.AccessoryConditionReady,
			Lifecycle: domain.AccessoryLifecycleStored, StorageLocationID: fixture.location.ID,
		}, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	layouts := application.NewLayoutService(infrastructure.NewLayoutRepository(db))
	fixture.layout, err = layouts.CreateLayout(ctx, application.CreateLayoutInput{
		Name: "Testanlage", Kind: domain.LayoutKindPrivate, Gauge: "TT", Scale: "1:120",
	}, "planner-1")
	if err != nil {
		t.Fatal(err)
	}
	fixture.unit, err = layouts.CreateUnit(ctx, fixture.layout.ID, application.CreateLayoutUnitInput{
		Name: "Modul 1", Kind: domain.LayoutUnitKindModule,
	}, "planner-1")
	if err != nil {
		t.Fatal(err)
	}
	fixture.vehicle, err = application.NewVehicleService(db).Create(ctx, application.CreateVehicleInput{
		InventoryNumber: "RK-TEST-1", Manufacturer: "Tillig", Name: "Testlok", Gauge: "TT",
		Category: "Lokomotive", Gattung: "Dampflokomotive",
	}, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	return fixture
}

func assertAllocationSummary(
	t *testing.T,
	service *application.AccessoryAllocationService,
	productID string,
	want application.AccessoryAllocationSummary,
) {
	t.Helper()
	got, err := service.GetAllocationSummary(t.Context(), productID)
	if err != nil {
		t.Fatal(err)
	}
	want.ProductID = productID
	if *got != want {
		t.Fatalf("allocation summary: got %#v, want %#v", *got, want)
	}
}

func assertAssetState(
	t *testing.T,
	service *application.AccessoryService,
	productID, assetID string,
	lifecycle domain.AccessoryLifecycle,
	condition domain.AccessoryCondition,
	locationID string,
) {
	t.Helper()
	assets, err := service.ListAssets(t.Context(), productID)
	if err != nil {
		t.Fatal(err)
	}
	for _, asset := range assets {
		if asset.ID == assetID {
			if asset.Lifecycle != lifecycle || asset.Condition != condition || asset.StorageLocationID != locationID {
				t.Fatalf("unexpected asset state: %#v", asset)
			}
			return
		}
	}
	t.Fatalf("asset %s not found in %#v", assetID, assets)
}
