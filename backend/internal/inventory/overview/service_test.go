package overview

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type priceSource struct {
	vehicles []VehiclePrice
	owned    []OwnedPrice
	buys     []Purchase
	manual   []string
	failed   string
	failure  error
	reads    []string
}

func (s *priceSource) read(area string) error {
	s.reads = append(s.reads, area)
	if s.failed == area {
		return s.failure
	}
	return nil
}
func (s *priceSource) VehiclePrices(context.Context) ([]VehiclePrice, error) {
	return s.vehicles, s.read("vehicles")
}
func (s *priceSource) AccessoryOwnedPrices(context.Context) ([]OwnedPrice, error) {
	return s.owned, s.read("owned")
}
func (s *priceSource) AccessoryPurchases(context.Context) ([]Purchase, error) {
	return s.buys, s.read("purchases")
}
func (s *priceSource) ManualAssetPrices(context.Context) ([]string, error) {
	return s.manual, s.read("manual")
}

func TestOverviewKeepsRecordedPricesAndCurrencyExclusions(t *testing.T) {
	source := &priceSource{
		vehicles: []VehiclePrice{{List: "EUR 1.299,90", Purchase: "95,00 €"}, {List: "unknown", Purchase: ""}},
		owned:    []OwnedPrice{{Price: "10.50", Quantity: 2}, {Price: "unknown", Quantity: 100}},
		buys:     []Purchase{{UnitPrice: "1,25", Quantity: 3, Currency: " eur "}, {UnitPrice: "250", Quantity: 5, Currency: "USD"}},
		manual:   []string{"9.99", "unrecorded"},
	}
	value, err := New(source).Get(t.Context())
	want := OverviewValuation{VehicleListValue: "1299.90", VehiclePurchaseValue: "95.00",
		AccessoryListValue: "21.00", AccessoryPurchaseCost: "13.74", ExcludedForeignCurrencyPurchases: 1}
	if err != nil || value != want || !reflect.DeepEqual(source.reads, []string{"vehicles", "owned", "purchases", "manual"}) {
		t.Fatal("recorded prices, excluded currencies or read order changed", value, err, source.reads)
	}
}

func TestOverviewDoesNotReturnPartialTotalsOnSourceFailure(t *testing.T) {
	for index, phase := range []string{"vehicles", "owned", "purchases", "manual"} {
		t.Run(phase, func(t *testing.T) {
			sentinel := errors.New("synthetic price source failure")
			source := &priceSource{failed: phase, failure: sentinel, vehicles: []VehiclePrice{{List: "100"}},
				owned: []OwnedPrice{{Price: "200", Quantity: 1}}, buys: []Purchase{{UnitPrice: "300", Quantity: 1}}}
			value, err := New(source).Get(t.Context())
			if !errors.Is(err, sentinel) || value != (OverviewValuation{}) || len(source.reads) != index+1 {
				t.Fatal("source failure was turned into partial totals or masked", value, err, source.reads)
			}
		})
	}
}

func TestOverviewRejectsOverflowAndNegativeQuantitiesWithoutPartialTotals(t *testing.T) {
	for name, source := range map[string]*priceSource{
		"vehicle-list":      {vehicles: []VehiclePrice{{List: "92233720368547758.07"}, {List: "0.01"}}},
		"vehicle-purchase":  {vehicles: []VehiclePrice{{Purchase: "92233720368547758.07"}, {Purchase: "0.01"}}},
		"owned":             {owned: []OwnedPrice{{Price: "92233720368547758.07", Quantity: 2}}},
		"purchases":         {buys: []Purchase{{UnitPrice: "92233720368547758.07", Quantity: 2}}},
		"manual":            {manual: []string{"92233720368547758.07", "0.01"}},
		"combined":          {buys: []Purchase{{UnitPrice: "92233720368547758.07", Quantity: 1}}, manual: []string{"0.01"}},
		"negative-owned":    {owned: []OwnedPrice{{Price: "1", Quantity: -1}}},
		"negative-purchase": {buys: []Purchase{{UnitPrice: "1", Quantity: -1}}},
	} {
		t.Run(name, func(t *testing.T) {
			value, err := New(source).Get(t.Context())
			if err == nil || value != (OverviewValuation{}) {
				t.Fatal("invalid arithmetic produced a misleading total", value, err)
			}
		})
	}
}

func TestOverviewUnavailableSourceIsAnError(t *testing.T) {
	for _, service := range []*Service{nil, New(nil)} {
		if value, err := service.Get(t.Context()); err == nil || value != (OverviewValuation{}) {
			t.Fatal("missing source reported a valid zero-valued collection", value, err)
		}
	}
}
