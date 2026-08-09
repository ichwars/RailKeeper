package application

import (
	"context"
	"errors"
	"testing"
)

func TestAccessoryServiceNormalizesPurchaseInput(t *testing.T) {
	repository := &accessoryRepositorySpy{}
	service := NewAccessoryService(repository)

	if _, err := service.CreatePurchase(t.Context(), " product-1 ", CreateAccessoryPurchaseInput{
		PurchasedAt: "2026-08-01", Supplier: " Händler ", Quantity: 3, UnitPrice: " 12.50 ",
		Currency: " eur ", InvoiceNumber: " INV-42 ", WarrantyUntil: "2028-08-01",
		StorageLocationID: " location-1 ", BookToStock: true, Notes: " delivery ",
	}, "editor-1"); err != nil {
		t.Fatal(err)
	}
	want := CreateAccessoryPurchaseInput{
		PurchasedAt: "2026-08-01", Supplier: "Händler", Quantity: 3, UnitPrice: "12.50",
		Currency: "EUR", InvoiceNumber: "INV-42", WarrantyUntil: "2028-08-01",
		StorageLocationID: "location-1", BookToStock: true, Notes: "delivery",
	}
	if repository.purchase != want {
		t.Fatalf("unexpected normalized purchase: got %#v, want %#v", repository.purchase, want)
	}
}

func TestAccessoryServiceValidatesPurchaseInput(t *testing.T) {
	service := NewAccessoryService(&accessoryRepositorySpy{})
	valid := CreateAccessoryPurchaseInput{PurchasedAt: "2026-08-01", Quantity: 1}
	tests := []struct {
		productID string
		input     CreateAccessoryPurchaseInput
	}{
		{"", valid},
		{"product-1", CreateAccessoryPurchaseInput{PurchasedAt: "2026-08-01"}},
		{"product-1", CreateAccessoryPurchaseInput{PurchasedAt: "01.08.2026", Quantity: 1}},
		{"product-1", CreateAccessoryPurchaseInput{PurchasedAt: "2026-08-01", WarrantyUntil: "2026-02-30", Quantity: 1}},
		{"product-1", CreateAccessoryPurchaseInput{PurchasedAt: "2026-08-01", Currency: "EU", Quantity: 1}},
		{"product-1", CreateAccessoryPurchaseInput{PurchasedAt: "2026-08-01", Currency: "EURO", Quantity: 1}},
		{"product-1", CreateAccessoryPurchaseInput{PurchasedAt: "2026-08-01", Currency: "E1R", Quantity: 1}},
		{"product-1", CreateAccessoryPurchaseInput{PurchasedAt: "2026-08-01", Quantity: 1, BookToStock: true}},
	}
	for _, test := range tests {
		if _, err := service.CreatePurchase(t.Context(), test.productID, test.input, "editor-1"); !errors.Is(err, ErrAccessoryValidation) {
			t.Fatalf("expected purchase validation error for %#v, got %v", test, err)
		}
	}
	if _, err := service.CreatePurchase(t.Context(), "product-1", valid, "editor-1"); err != nil {
		t.Fatalf("unbooked purchase must not require a destination: %v", err)
	}
}

func (spy *accessoryRepositorySpy) CreatePurchase(
	_ context.Context,
	_ string,
	input CreateAccessoryPurchaseInput,
	_ string,
) (*AccessoryPurchase, error) {
	spy.purchase = input
	return &AccessoryPurchase{}, nil
}
