package domain

import "testing"

func TestSetNormalizationRetainsRecordedPricesAndUnknownDates(t *testing.T) {
	input := VehicleSetInput{Name: " Synthetic set ", Manufacturer: " Synthetic ", Gauge: " H0 ", Category: " Train ", Gattung: " Passenger ", PurchasePrice: " 129,99 EUR ", ListPrice: " 150.00 ", PurchaseDate: " ", ArticleSourceURL: " https://example.invalid/source "}
	clean := CleanVehicleSetInput(input)
	if !ValidVehicleSetInput(clean) || clean.PurchasePrice != "129,99 EUR" || clean.ListPrice != "150.00" || clean.PurchaseDate != "" || clean.ArticleSourceURL != "https://example.invalid/source" {
		t.Fatal("set normalization changed recorded values or invented a date")
	}
	if input.Name != " Synthetic set " {
		t.Fatal("normalization mutated the caller's draft")
	}
	if ValidVehicleSetInput(VehicleSetInput{Name: "incomplete"}) {
		t.Fatal("incomplete aggregate accepted")
	}
	if MaxVehicleSetMembers != 100 {
		t.Fatal("compatible set capacity changed")
	}
}
