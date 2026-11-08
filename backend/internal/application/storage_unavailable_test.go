package application

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"testing"

	_ "modernc.org/sqlite"
)

// Storage outages must not look like missing inventory or successful empty results.
func unavailableResult[T any](value T, err error) error {
	var zero T
	if err == nil || !reflect.DeepEqual(value, zero) {
		return fmt.Errorf("storage outage returned data or success: %v", err)
	}
	for _, rejected := range []error{ErrVehicleNotFound, ErrVehicleValidation, ErrVehicleSetNotFound,
		ErrMasterDataNotFound, ErrMasterDataValidation, ErrInventoryNumberNotFound, ErrInventoryNumberValidation} {
		if errors.Is(err, rejected) {
			return fmt.Errorf("storage outage was reported as a missing or invalid record: %w", err)
		}
	}
	return nil
}

func TestInventoryStorageOutageNeverAcknowledgesReadsOrWrites(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	vehicles := NewVehicleService(db)
	masterData := NewMasterDataService(db)
	numbers := NewInventoryNumberService(db)
	maintenance := VehicleMaintenanceInput{Kind: "Reinigung", Status: "geplant"}
	part := VehicleSparePartInput{ArticleNumber: "part-1", Description: "Recorded part"}
	operations := map[string]func() error{
		"inventory":           func() error { return unavailableResult(vehicles.List(t.Context(), "")) },
		"inventory-read-only": func() error { return unavailableResult(vehicles.ListReadOnly(t.Context(), "")) },
		"vehicle":             func() error { return unavailableResult(vehicles.Get(t.Context(), "vehicle")) },
		"set":                 func() error { return unavailableResult(vehicles.GetSet(t.Context(), "set")) },
		"set-image":           func() error { return unavailableResult(vehicles.GetSetImage(t.Context(), "set")) },
		"image":               func() error { return unavailableResult(vehicles.GetImage(t.Context(), "vehicle", "image")) },
		"image-delete":        func() error { return unavailableResult(vehicles.DeleteImage(t.Context(), "vehicle", "image")) },
		"attachment":          func() error { return unavailableResult(vehicles.GetAttachment(t.Context(), "vehicle", "attachment")) },
		"attachment-delete": func() error {
			return unavailableResult(vehicles.DeleteAttachment(t.Context(), "vehicle", "attachment"))
		},
		"file-references": func() error {
			return unavailableResult(vehicles.MediaFileReferenceCount(t.Context(), "images/recorded.png"))
		},
		"maintenance-list": func() error { return unavailableResult(vehicles.ListMaintenance(t.Context(), "vehicle")) },
		"maintenance":      func() error { return unavailableResult(vehicles.GetMaintenance(t.Context(), "vehicle", "maintenance")) },
		"maintenance-create": func() error {
			return unavailableResult(vehicles.CreateMaintenance(t.Context(), "vehicle", maintenance))
		},
		"maintenance-update": func() error {
			return unavailableResult(vehicles.UpdateMaintenance(t.Context(), "vehicle", "maintenance", maintenance))
		},
		"maintenance-delete": func() error {
			return unavailableResult(vehicles.DeleteMaintenance(t.Context(), "vehicle", "maintenance"))
		},
		"spare-parts":           func() error { return unavailableResult(vehicles.ListSpareParts(t.Context(), "vehicle")) },
		"spare-part":            func() error { return unavailableResult(vehicles.GetSparePart(t.Context(), "vehicle", "part")) },
		"spare-part-create":     func() error { return unavailableResult(vehicles.CreateSparePart(t.Context(), "vehicle", part)) },
		"spare-part-update":     func() error { return unavailableResult(vehicles.UpdateSparePart(t.Context(), "vehicle", "part", part)) },
		"spare-part-delete":     func() error { return unavailableResult(vehicles.DeleteSparePart(t.Context(), "vehicle", "part")) },
		"master-data-list":      func() error { return unavailableResult(masterData.List(t.Context(), "gauge", false)) },
		"master-data-all":       func() error { return unavailableResult(masterData.ListAll(t.Context(), false)) },
		"master-data":           func() error { return unavailableResult(masterData.Get(t.Context(), "gauge", "recorded")) },
		"master-data-relations": func() error { return unavailableResult(masterData.Relations(t.Context(), "category", "gattung")) },
		"master-data-export":    func() error { return unavailableResult(masterData.Export(t.Context())) },
		"numbering-list":        func() error { return unavailableResult(numbers.List(t.Context())) },
		"numbering-create": func() error {
			return unavailableResult(numbers.Create(t.Context(), InventoryNumberSchemeCreateInput{Category: "Custom", InventoryNumberSchemeInput: InventoryNumberSchemeInput{Prefix: "RK", Padding: 6, NextNumber: 1}}))
		},
		"numbering-update": func() error {
			return unavailableResult(numbers.Update(t.Context(), "Custom", InventoryNumberSchemeInput{Prefix: "RK", Padding: 6, NextNumber: 1}))
		},
		"overview": func() error { return unavailableResult(NewOverviewValuationService(db).Get(t.Context())) },
	}
	for name, operation := range operations {
		t.Run(name, func(t *testing.T) {
			if err := operation(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
