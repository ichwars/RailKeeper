package domain_test

import (
	"testing"

	"railkeeper/backend/internal/domain"
)

func TestAccessoryEnumsAcceptOnlyDefinedValues(t *testing.T) {
	tests := []struct {
		name  string
		valid bool
		want  bool
	}{
		{"tracking quantity", domain.AccessoryTrackingModeQuantity.Valid(), true},
		{"tracking individual", domain.AccessoryTrackingModeIndividual.Valid(), true},
		{"tracking invalid", domain.AccessoryTrackingMode("bulk").Valid(), false},
		{"condition ready", domain.AccessoryConditionReady.Valid(), true},
		{"condition maintenance", domain.AccessoryConditionMaintenanceDue.Valid(), true},
		{"condition defective", domain.AccessoryConditionDefective.Valid(), true},
		{"condition unknown", domain.AccessoryConditionUnknown.Valid(), true},
		{"condition invalid", domain.AccessoryCondition("new").Valid(), false},
		{"lifecycle stored", domain.AccessoryLifecycleStored.Valid(), true},
		{"lifecycle reserved", domain.AccessoryLifecycleReserved.Valid(), true},
		{"lifecycle installed", domain.AccessoryLifecycleInstalled.Valid(), true},
		{"lifecycle maintenance", domain.AccessoryLifecycleMaintenance.Valid(), true},
		{"lifecycle retired", domain.AccessoryLifecycleRetired.Valid(), true},
		{"lifecycle invalid", domain.AccessoryLifecycle("sold").Valid(), false},
		{"reservation active", domain.AccessoryReservationActive.Valid(), true},
		{"reservation fulfilled", domain.AccessoryReservationFulfilled.Valid(), true},
		{"reservation cancelled", domain.AccessoryReservationCancelled.Valid(), true},
		{"reservation invalid", domain.AccessoryReservationStatus("closed").Valid(), false},
		{"removal stored", domain.AccessoryRemovalStored.Valid(), true},
		{"removal maintenance", domain.AccessoryRemovalMaintenance.Valid(), true},
		{"removal defective", domain.AccessoryRemovalDefective.Valid(), true},
		{"removal retired", domain.AccessoryRemovalRetired.Valid(), true},
		{"removal invalid", domain.AccessoryRemovalDisposition("sold").Valid(), false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.valid != test.want {
				t.Fatalf("Valid() = %t, want %t", test.valid, test.want)
			}
		})
	}
}

func TestAccessoryArticleEnumsAcceptOnlyDefinedValues(t *testing.T) {
	strategies := []struct {
		name  string
		value domain.AccessoryInventoryStrategy
		want  bool
	}{
		{"quantity", domain.AccessoryInventoryQuantity, true},
		{"individual", domain.AccessoryInventoryIndividual, true},
		{"quantity later individual", domain.AccessoryInventoryQuantityLaterIndividual, true},
		{"invalid strategy", domain.AccessoryInventoryStrategy("bulk"), false},
	}
	for _, test := range strategies {
		t.Run(test.name, func(t *testing.T) {
			if got := test.value.Valid(); got != test.want {
				t.Fatalf("Valid() = %t, want %t", got, test.want)
			}
		})
	}

	types := []struct {
		name  string
		value domain.AccessoryArticleType
		want  bool
	}{
		{"track", domain.AccessoryArticleTrack, true},
		{"signal", domain.AccessoryArticleSignal, true},
		{"decoder", domain.AccessoryArticleDecoder, true},
		{"electrical control", domain.AccessoryArticleElectricalControl, true},
		{"building equipment", domain.AccessoryArticleBuildingEquipment, true},
		{"landscape consumable", domain.AccessoryArticleLandscapeConsumable, true},
		{"lighting", domain.AccessoryArticleLighting, true},
		{"other", domain.AccessoryArticleOther, true},
		{"invalid article type", domain.AccessoryArticleType("vehicle"), false},
	}
	for _, test := range types {
		t.Run(test.name, func(t *testing.T) {
			if got := test.value.Valid(); got != test.want {
				t.Fatalf("Valid() = %t, want %t", got, test.want)
			}
		})
	}
}

func TestValidateAllocationTargetRequiresExactlyOneTarget(t *testing.T) {
	valid := []domain.AllocationTarget{
		{VehicleID: "vehicle-1"},
		{LayoutID: "layout-1"},
		{LayoutUnitID: "unit-1"},
	}
	for _, target := range valid {
		if err := target.Validate(); err != nil {
			t.Fatalf("expected valid target %#v, got %v", target, err)
		}
	}

	invalid := []domain.AllocationTarget{
		{},
		{VehicleID: "vehicle-1", LayoutID: "layout-1"},
		{VehicleID: "vehicle-1", LayoutUnitID: "unit-1"},
		{LayoutID: "layout-1", LayoutUnitID: "unit-1"},
	}
	for _, target := range invalid {
		if err := target.Validate(); err == nil {
			t.Fatalf("expected invalid target %#v", target)
		}
	}
}
