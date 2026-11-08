// Package overview owns the existing collection value calculations.
// It is separate from immutable valuation-history events and never infers market prices.
package overview

import "context"

type OverviewValuation struct {
	VehicleListValue                 string `json:"vehicleListValue"`
	VehiclePurchaseValue             string `json:"vehiclePurchaseValue"`
	AccessoryListValue               string `json:"accessoryListValue"`
	AccessoryPurchaseCost            string `json:"accessoryPurchaseCost"`
	ExcludedForeignCurrencyPurchases int    `json:"excludedForeignCurrencyPurchases"`
}
type VehiclePrice struct{ List, Purchase string }
type OwnedPrice struct {
	Price    string
	Quantity int64
}
type Purchase struct {
	UnitPrice, Currency string
	Quantity            int64
}
type Source interface {
	VehiclePrices(context.Context) ([]VehiclePrice, error)
	AccessoryOwnedPrices(context.Context) ([]OwnedPrice, error)
	AccessoryPurchases(context.Context) ([]Purchase, error)
	ManualAssetPrices(context.Context) ([]string, error)
}
