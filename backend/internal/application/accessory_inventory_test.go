package application

import (
	"context"
	"errors"
	"testing"

	"railkeeper/backend/internal/domain"
)

func TestAccessoryServiceNormalizesTransferAndIndividualizationInputs(t *testing.T) {
	repository := &accessoryRepositorySpy{}
	service := NewAccessoryService(repository)

	if _, err := service.TransferStock(t.Context(), " product-1 ", TransferAccessoryStockInput{
		FromLocationID: " source-1 ", ToLocationID: " destination-1 ", Quantity: 2, Note: " relocation ",
	}, "editor-1"); err != nil {
		t.Fatal(err)
	}
	if repository.stockTransfer != (TransferAccessoryStockInput{
		FromLocationID: "source-1", ToLocationID: "destination-1", Quantity: 2, Note: "relocation",
	}) {
		t.Fatalf("unexpected normalized transfer: %#v", repository.stockTransfer)
	}

	if _, err := service.Individualize(t.Context(), " product-1 ", IndividualizeAccessoryInput{
		LocationID: " source-1 ", Asset: CreateAccessoryAssetInput{
			InventoryNumber: " A-1 ", SerialNumber: " serial-1 ", StorageLocationID: " ignored ",
			PurchaseDate: "2026-08-01", WarrantyUntil: "2028-08-01", Notes: " stored unit ",
		},
	}, "editor-1"); err != nil {
		t.Fatal(err)
	}
	individualized := repository.individualization
	if individualized.LocationID != "source-1" || individualized.Asset.InventoryNumber != "A-1" ||
		individualized.Asset.SerialNumber != "serial-1" || individualized.Asset.StorageLocationID != "source-1" ||
		individualized.Asset.Condition != domain.AccessoryConditionUnknown ||
		individualized.Asset.Lifecycle != domain.AccessoryLifecycleStored || individualized.Asset.Notes != "stored unit" {
		t.Fatalf("unexpected normalized individualization: %#v", individualized)
	}
}

func TestAccessoryServiceValidatesTransfersAndIndividualization(t *testing.T) {
	service := NewAccessoryService(&accessoryRepositorySpy{})
	validTransfer := TransferAccessoryStockInput{
		FromLocationID: "source-1", ToLocationID: "destination-1", Quantity: 1,
	}
	for _, test := range []struct {
		productID string
		input     TransferAccessoryStockInput
	}{
		{"", validTransfer},
		{"product-1", TransferAccessoryStockInput{ToLocationID: "destination-1", Quantity: 1}},
		{"product-1", TransferAccessoryStockInput{FromLocationID: "source-1", Quantity: 1}},
		{"product-1", TransferAccessoryStockInput{FromLocationID: "source-1", ToLocationID: "source-1", Quantity: 1}},
		{"product-1", TransferAccessoryStockInput{FromLocationID: "source-1", ToLocationID: "destination-1"}},
	} {
		if _, err := service.TransferStock(t.Context(), test.productID, test.input, "editor-1"); !errors.Is(err, ErrAccessoryValidation) {
			t.Fatalf("expected transfer validation error for %#v, got %v", test, err)
		}
	}

	validAsset := CreateAccessoryAssetInput{PurchaseDate: "2026-08-01", WarrantyUntil: "2028-08-01"}
	for _, test := range []struct {
		productID string
		input     IndividualizeAccessoryInput
	}{
		{"", IndividualizeAccessoryInput{LocationID: "source-1", Asset: validAsset}},
		{"product-1", IndividualizeAccessoryInput{Asset: validAsset}},
		{"product-1", IndividualizeAccessoryInput{LocationID: "source-1", Asset: CreateAccessoryAssetInput{PurchaseDate: "08/01/2026"}}},
		{"product-1", IndividualizeAccessoryInput{LocationID: "source-1", Asset: CreateAccessoryAssetInput{WarrantyUntil: "2026-02-30"}}},
		{"product-1", IndividualizeAccessoryInput{LocationID: "source-1", Asset: CreateAccessoryAssetInput{Lifecycle: domain.AccessoryLifecycleInstalled}}},
	} {
		if _, err := service.Individualize(t.Context(), test.productID, test.input, "editor-1"); !errors.Is(err, ErrAccessoryValidation) {
			t.Fatalf("expected individualization validation error for %#v, got %v", test, err)
		}
	}
}

func (spy *accessoryRepositorySpy) TransferStock(
	_ context.Context,
	_ string,
	input TransferAccessoryStockInput,
	_ string,
) (*AccessoryStockSummary, error) {
	spy.stockTransfer = input
	return &AccessoryStockSummary{}, nil
}

func (spy *accessoryRepositorySpy) Individualize(
	_ context.Context,
	_ string,
	input IndividualizeAccessoryInput,
	_ string,
) (*AccessoryAsset, error) {
	spy.individualization = input
	return &AccessoryAsset{}, nil
}
