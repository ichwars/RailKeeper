package application

import (
	"context"
	"strings"
	"time"
)

type CreateAccessoryPurchaseInput struct {
	PurchasedAt       string `json:"purchasedAt"`
	Supplier          string `json:"supplier"`
	Quantity          int    `json:"quantity"`
	UnitPrice         string `json:"unitPrice"`
	Currency          string `json:"currency"`
	InvoiceNumber     string `json:"invoiceNumber"`
	WarrantyUntil     string `json:"warrantyUntil"`
	StorageLocationID string `json:"storageLocationId"`
	BookToStock       bool   `json:"bookToStock"`
	Notes             string `json:"notes"`
}

type AccessoryPurchase struct {
	ID                string `json:"id"`
	ProductID         string `json:"productId"`
	StorageLocationID string `json:"storageLocationId,omitempty"`
	Quantity          int    `json:"quantity"`
	PurchasedAt       string `json:"purchasedAt"`
	Supplier          string `json:"supplier,omitempty"`
	UnitPrice         string `json:"unitPrice,omitempty"`
	Currency          string `json:"currency,omitempty"`
	InvoiceNumber     string `json:"invoiceNumber,omitempty"`
	WarrantyUntil     string `json:"warrantyUntil,omitempty"`
	BookToStock       bool   `json:"bookToStock"`
	Notes             string `json:"notes,omitempty"`
	CreatedBy         string `json:"createdBy,omitempty"`
	CreatedAt         string `json:"createdAt"`
	UpdatedAt         string `json:"updatedAt"`
}

func (s *AccessoryService) ListPurchases(ctx context.Context, productID string) ([]AccessoryPurchase, error) {
	productID = strings.TrimSpace(productID)
	if productID == "" {
		return nil, ErrAccessoryValidation
	}
	return s.repository.ListPurchases(ctx, productID)
}

func (s *AccessoryService) CreatePurchase(
	ctx context.Context,
	productID string,
	input CreateAccessoryPurchaseInput,
	actor string,
) (*AccessoryPurchase, error) {
	productID = strings.TrimSpace(productID)
	input = cleanAccessoryPurchaseInput(input)
	if productID == "" || input.Quantity <= 0 || !validAccessoryDate(input.PurchasedAt) ||
		!validOptionalAccessoryDate(input.WarrantyUntil) || !validAccessoryCurrency(input.Currency) ||
		(input.BookToStock && input.StorageLocationID == "") {
		return nil, ErrAccessoryValidation
	}
	return s.repository.CreatePurchase(ctx, productID, input, actor)
}

func cleanAccessoryPurchaseInput(input CreateAccessoryPurchaseInput) CreateAccessoryPurchaseInput {
	input.PurchasedAt = strings.TrimSpace(input.PurchasedAt)
	input.Supplier = strings.TrimSpace(input.Supplier)
	input.UnitPrice = strings.TrimSpace(input.UnitPrice)
	input.Currency = strings.ToUpper(strings.TrimSpace(input.Currency))
	input.InvoiceNumber = strings.TrimSpace(input.InvoiceNumber)
	input.WarrantyUntil = strings.TrimSpace(input.WarrantyUntil)
	input.StorageLocationID = strings.TrimSpace(input.StorageLocationID)
	input.Notes = strings.TrimSpace(input.Notes)
	return input
}

func validAccessoryCurrency(currency string) bool {
	if currency == "" {
		return true
	}
	if len(currency) != 3 {
		return false
	}
	for _, character := range currency {
		if character < 'A' || character > 'Z' {
			return false
		}
	}
	return true
}

func validAccessoryDate(value string) bool {
	parsed, err := time.Parse("2006-01-02", value)
	return err == nil && parsed.Format("2006-01-02") == value
}

func validOptionalAccessoryDate(value string) bool {
	return value == "" || validAccessoryDate(value)
}
