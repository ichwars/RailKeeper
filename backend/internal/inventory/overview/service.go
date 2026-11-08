package overview

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"railkeeper/backend/internal/inventory/money"
)

type Service struct{ source Source }

func New(source Source) *Service { return &Service{source: source} }
func (s *Service) Get(ctx context.Context) (OverviewValuation, error) {
	if s == nil || s.source == nil {
		return OverviewValuation{}, errors.New("calculate overview valuation: database unavailable")
	}
	vehicleList, vehiclePurchase, err := s.vehicleValues(ctx)
	if err != nil {
		return OverviewValuation{}, err
	}
	accessoryList, err := s.accessoryListValue(ctx)
	if err != nil {
		return OverviewValuation{}, err
	}
	accessoryPurchases, excluded, err := s.accessoryPurchaseCost(ctx)
	if err != nil {
		return OverviewValuation{}, err
	}
	manualAssetPurchases, err := s.manualAssetPurchaseCost(ctx)
	if err != nil {
		return OverviewValuation{}, err
	}
	accessoryPurchases, err = money.CheckedAdd(accessoryPurchases, manualAssetPurchases)
	if err != nil {
		return OverviewValuation{}, fmt.Errorf("calculate accessory purchase cost: %w", err)
	}
	return OverviewValuation{
		VehicleListValue:                 money.FormatCents(vehicleList),
		VehiclePurchaseValue:             money.FormatCents(vehiclePurchase),
		AccessoryListValue:               money.FormatCents(accessoryList),
		AccessoryPurchaseCost:            money.FormatCents(accessoryPurchases),
		ExcludedForeignCurrencyPurchases: excluded,
	}, nil
}
func (s *Service) vehicleValues(ctx context.Context) (int64, int64, error) {
	rows, err := s.source.VehiclePrices(ctx)
	if err != nil {
		return 0, 0, err
	}
	var listTotal, purchaseTotal int64
	for _, row := range rows {
		if err := addParsedVehicleMoney(&listTotal, row.List); err != nil {
			return 0, 0, fmt.Errorf("calculate vehicle list value: %w", err)
		}
		if err := addParsedVehicleMoney(&purchaseTotal, row.Purchase); err != nil {
			return 0, 0, fmt.Errorf("calculate vehicle purchase value: %w", err)
		}
	}
	return listTotal, purchaseTotal, nil
}
func (s *Service) accessoryListValue(ctx context.Context) (int64, error) {
	rows, err := s.source.AccessoryOwnedPrices(ctx)
	if err != nil {
		return 0, err
	}
	var total int64
	for _, row := range rows {
		if err := addParsedMoney(&total, row.Price, row.Quantity); err != nil {
			return 0, fmt.Errorf("calculate accessory list value: %w", err)
		}
	}
	return total, nil
}
func (s *Service) accessoryPurchaseCost(ctx context.Context) (int64, int, error) {
	rows, err := s.source.AccessoryPurchases(ctx)
	if err != nil {
		return 0, 0, err
	}
	var total int64
	excluded := 0
	for _, row := range rows {
		currencyCode := strings.ToUpper(strings.TrimSpace(row.Currency))
		if currencyCode != "" && currencyCode != "EUR" {
			if excluded == math.MaxInt {
				return 0, 0, fmt.Errorf("calculate accessory purchase cost: excluded count overflow")
			}
			excluded++
			continue
		}
		if err := addParsedMoney(&total, row.UnitPrice, row.Quantity); err != nil {
			return 0, 0, fmt.Errorf("calculate accessory purchase cost: %w", err)
		}
	}
	return total, excluded, nil
}
func (s *Service) manualAssetPurchaseCost(ctx context.Context) (int64, error) {
	rows, err := s.source.ManualAssetPrices(ctx)
	if err != nil {
		return 0, err
	}
	var total int64
	for _, price := range rows {
		if err := addParsedMoney(&total, price, 1); err != nil {
			return 0, fmt.Errorf("calculate manual asset purchase cost: %w", err)
		}
	}
	return total, nil
}
func addParsedVehicleMoney(total *int64, value string) error {
	cents, ok := money.ParseVehicleCents(value)
	if !ok {
		return nil
	}
	var err error
	*total, err = money.CheckedAdd(*total, cents)
	return err
}
func addParsedMoney(total *int64, value string, quantity int64) error {
	cents, ok := money.ParseCents(value)
	if !ok {
		return nil
	}
	lineTotal, err := money.CheckedMultiply(cents, quantity)
	if err != nil {
		return err
	}
	*total, err = money.CheckedAdd(*total, lineTotal)
	return err
}
