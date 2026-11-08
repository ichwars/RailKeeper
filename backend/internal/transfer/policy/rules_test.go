package policy

import (
	"errors"
	"testing"

	"railkeeper/backend/internal/transfer"
)

func TestPreviewFailsClosedWithoutAggregateValidation(t *testing.T) {
	vehicle := TransferVehicle{InventoryNumber: "RK-1", Manufacturer: "Synthetic", Name: "Synthetic vehicle", Gauge: "H0", Category: "Locomotive", Gattung: "Diesel"}
	accessory := TransferAccessory{InventoryNumber: "ACC-1", Manufacturer: "Synthetic", Name: "Synthetic accessory"}
	input := DataTransferSnapshot{Vehicles: []TransferVehicle{vehicle}, Accessories: []TransferAccessory{accessory}}
	for _, rules := range []AggregateRules{{}, {Vehicle: func(TransferVehicle) error { return errors.New("synthetic vehicle rejection") }, Accessory: func(TransferAccessory) error { return errors.New("synthetic accessory rejection") }}} {
		records, issues := ClassifyDataTransferImport("job", input, DataTransferSnapshot{}, nil, rules)
		if len(records) != 2 || records[0].Classification != "error" || records[1].Classification != "error" {
			t.Fatal("missing or rejected aggregate rules yielded ready records")
		}
		found := map[string]bool{}
		for _, issue := range issues {
			found[issue.Code] = true
		}
		if !found["invalid_vehicle"] || !found["invalid_accessory"] {
			t.Fatal("aggregate rejection lost its actionable diagnostic")
		}
	}
	calls := 0
	records, issues := ClassifyDataTransferImport("job", input, DataTransferSnapshot{}, nil, AggregateRules{Vehicle: func(v TransferVehicle) error {
		calls++
		if v.InventoryNumber != "RK-1" {
			t.Fatal("wrong vehicle projection")
		}
		return nil
	}, Accessory: func(a TransferAccessory) error {
		calls++
		if a.InventoryNumber != "ACC-1" {
			t.Fatal("wrong accessory projection")
		}
		return nil
	}})
	if calls != 2 || len(issues) != 0 || records[0].Classification != "ready" || records[1].Classification != "ready" {
		t.Fatal("approved aggregates did not become reviewable records")
	}
}

func TestPreviewAndApplyShareTheSameResolutionPolicy(t *testing.T) {
	for _, example := range []struct {
		code, resolution string
		allowed          bool
	}{
		{"duplicate_inventory_number", "replace", true}, {"duplicate_inventory_number", "copy", true}, {"duplicate_inventory_number", "skip", true},
		{"locked_exhibition_list", "replace", false}, {"locked_exhibition_list", "copy", true},
		{"missing_vehicle_reference", "link", false}, {"missing_vehicle_reference", "skip", true},
		{"vehicle_set_member_external_conflict", "replace", false}, {"vehicle_set_member_external_conflict", "copy", true},
		{"unknown", "skip", false}, {"duplicate_inventory_number", "", false},
	} {
		preview := ValidDataTransferResolution(transfer.DataTransferIssue{Code: example.code}, example.resolution)
		apply := ValidTransferApplyResolution(example.code, example.resolution)
		if preview != example.allowed || apply != example.allowed {
			t.Fatal("preview/apply resolution mismatch", example.code, example.resolution)
		}
	}
}

func TestPreviewRowIdentityDistinguishesAbsentFromZero(t *testing.T) {
	zero, one := 0, 1
	for _, row := range []struct {
		left, right *int
		equal       bool
	}{{nil, nil, true}, {nil, &zero, false}, {&zero, nil, false}, {&zero, &zero, true}, {&zero, &one, false}} {
		if transferPreviewRowsEqual(row.left, row.right) != row.equal || transferRowNumbersEqual(row.left, row.right) != row.equal {
			t.Fatal("row identity collapsed an absent row into zero")
		}
	}
}
