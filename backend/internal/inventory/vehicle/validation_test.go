package vehicle

import (
	"errors"
	"strings"
	"testing"
)

func TestOperationalValidationCountsUnicodeAndDoesNotRewriteRecordedValues(t *testing.T) {
	speed := 1000
	input := CreateVehicleInput{Manufacturer: " Maker ", Name: " Name ", Gauge: " TT ", Category: " Lokomotive ",
		Gattung: " Diesellok ", HomeBase: strings.Repeat("ä", 200), MaximumSpeedKmh: &speed,
		PurchasePrice: " 12,50 EUR ", CouplingSame: true, CouplingFront: " Front ", CouplingRear: " Other "}
	cleaned, err := ValidateCreateInput(input)
	if err != nil || cleaned.PurchasePrice != "12,50 EUR" || cleaned.HomeBase != input.HomeBase ||
		cleaned.CouplingRear != "Front" || input.CouplingRear != " Other " || *input.MaximumSpeedKmh != 1000 {
		t.Fatal("normalization changed recorded semantics", cleaned, err)
	}
	input.HomeBase += "ä"
	if _, err := ValidateCreateInput(input); !errors.Is(err, ErrVehicleOperationalValidation) {
		t.Fatal("overlong Unicode field accepted", err)
	}
	input.HomeBase = ""
	speed = 0
	if _, err := ValidateCreateInput(input); !errors.Is(err, ErrVehicleOperationalValidation) {
		t.Fatal("zero speed accepted", err)
	}
}

func TestFunctionChangeIdentityRetainsLegacyHexAndRejectsDuplicateNormalizedKeys(t *testing.T) {
	input := VehicleFunctionInput{FlagsVersion: 2}
	change := VehicleFunctionChange{FunctionKey: "f01", ExpectedRevision: strings.Repeat("A", 64), Value: &input}
	if err := ValidateFunctionChanges([]VehicleFunctionChange{change}); err != nil {
		t.Fatal("previously valid token rejected", err)
	}
	duplicate := change
	duplicate.FunctionKey = "F1"
	if err := ValidateFunctionChanges([]VehicleFunctionChange{change, duplicate}); !errors.Is(err, ErrVehicleFunctionValidation) {
		t.Fatal("duplicate normalized function accepted", err)
	}
	change.Value = nil
	change.Delete = true
	change.ExpectedRevision = ""
	if err := ValidateFunctionChanges([]VehicleFunctionChange{change}); !errors.Is(err, ErrVehicleFunctionValidation) {
		t.Fatal("unguarded deletion accepted", err)
	}
}
