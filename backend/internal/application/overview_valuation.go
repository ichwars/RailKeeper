package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
)

type OverviewValuation struct {
	VehicleListValue                 string `json:"vehicleListValue"`
	VehiclePurchaseValue             string `json:"vehiclePurchaseValue"`
	AccessoryListValue               string `json:"accessoryListValue"`
	AccessoryPurchaseCost            string `json:"accessoryPurchaseCost"`
	ExcludedForeignCurrencyPurchases int    `json:"excludedForeignCurrencyPurchases"`
}

type OverviewValuationService struct {
	db *sql.DB
}

func NewOverviewValuationService(db *sql.DB) *OverviewValuationService {
	return &OverviewValuationService{db: db}
}

func (s *OverviewValuationService) Get(ctx context.Context) (OverviewValuation, error) {
	if s == nil || s.db == nil {
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
	accessoryPurchases, err = checkedMoneyAdd(accessoryPurchases, manualAssetPurchases)
	if err != nil {
		return OverviewValuation{}, fmt.Errorf("calculate accessory purchase cost: %w", err)
	}
	return OverviewValuation{
		VehicleListValue:                 formatMoneyCents(vehicleList),
		VehiclePurchaseValue:             formatMoneyCents(vehiclePurchase),
		AccessoryListValue:               formatMoneyCents(accessoryList),
		AccessoryPurchaseCost:            formatMoneyCents(accessoryPurchases),
		ExcludedForeignCurrencyPurchases: excluded,
	}, nil
}

func (s *OverviewValuationService) vehicleValues(ctx context.Context) (int64, int64, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT list_price, purchase_price FROM vehicles`)
	if err != nil {
		return 0, 0, fmt.Errorf("calculate vehicle values: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var listTotal, purchaseTotal int64
	for rows.Next() {
		var listPrice, purchasePrice sql.NullString
		if err := rows.Scan(&listPrice, &purchasePrice); err != nil {
			return 0, 0, fmt.Errorf("calculate vehicle values: %w", err)
		}
		if err := addParsedVehicleMoney(&listTotal, listPrice.String); err != nil {
			return 0, 0, fmt.Errorf("calculate vehicle list value: %w", err)
		}
		if err := addParsedVehicleMoney(&purchaseTotal, purchasePrice.String); err != nil {
			return 0, 0, fmt.Errorf("calculate vehicle purchase value: %w", err)
		}
	}
	if err := rows.Err(); err != nil {
		return 0, 0, fmt.Errorf("calculate vehicle values: %w", err)
	}
	return listTotal, purchaseTotal, nil
}

func addParsedVehicleMoney(total *int64, value string) error {
	cents, ok := parseVehicleMoneyCents(value)
	if !ok {
		return nil
	}
	var err error
	*total, err = checkedMoneyAdd(*total, cents)
	return err
}

func (s *OverviewValuationService) accessoryListValue(ctx context.Context) (int64, error) {
	rows, err := s.db.QueryContext(ctx, `
WITH stock AS (
  SELECT product_id, SUM(quantity) AS quantity FROM accessory_stock GROUP BY product_id
), assets AS (
  SELECT product_id, SUM(CASE WHEN lifecycle_state <> 'retired' THEN 1 ELSE 0 END) AS quantity
  FROM accessory_assets GROUP BY product_id
), installations AS (
  SELECT product_id, SUM(CASE WHEN removed_at IS NULL AND asset_id IS NULL THEN quantity ELSE 0 END) AS quantity
  FROM accessory_installations GROUP BY product_id
)
SELECT product.list_price,
       CASE product.inventory_strategy
         WHEN 'individual' THEN COALESCE(assets.quantity, 0)
         WHEN 'quantity_later_individual' THEN
           COALESCE(stock.quantity, 0) + COALESCE(assets.quantity, 0) + COALESCE(installations.quantity, 0)
         ELSE COALESCE(stock.quantity, 0) + COALESCE(installations.quantity, 0)
       END AS owned
FROM accessory_products product
LEFT JOIN stock ON stock.product_id=product.id
LEFT JOIN assets ON assets.product_id=product.id
LEFT JOIN installations ON installations.product_id=product.id`)
	if err != nil {
		return 0, fmt.Errorf("calculate accessory list value: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var total int64
	for rows.Next() {
		var listPrice sql.NullString
		var quantity int64
		if err := rows.Scan(&listPrice, &quantity); err != nil {
			return 0, fmt.Errorf("calculate accessory list value: %w", err)
		}
		if err := addParsedMoney(&total, listPrice.String, quantity); err != nil {
			return 0, fmt.Errorf("calculate accessory list value: %w", err)
		}
	}
	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("calculate accessory list value: %w", err)
	}
	return total, nil
}

func (s *OverviewValuationService) accessoryPurchaseCost(ctx context.Context) (int64, int, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT quantity, unit_price, currency FROM accessory_purchases`)
	if err != nil {
		return 0, 0, fmt.Errorf("calculate accessory purchase cost: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var total int64
	excluded := 0
	for rows.Next() {
		var quantity int64
		var unitPrice, currency sql.NullString
		if err := rows.Scan(&quantity, &unitPrice, &currency); err != nil {
			return 0, 0, fmt.Errorf("calculate accessory purchase cost: %w", err)
		}
		currencyCode := strings.ToUpper(strings.TrimSpace(currency.String))
		if currencyCode != "" && currencyCode != "EUR" {
			if excluded == math.MaxInt {
				return 0, 0, fmt.Errorf("calculate accessory purchase cost: excluded count overflow")
			}
			excluded++
			continue
		}
		if err := addParsedMoney(&total, unitPrice.String, quantity); err != nil {
			return 0, 0, fmt.Errorf("calculate accessory purchase cost: %w", err)
		}
	}
	if err := rows.Err(); err != nil {
		return 0, 0, fmt.Errorf("calculate accessory purchase cost: %w", err)
	}
	return total, excluded, nil
}

func (s *OverviewValuationService) manualAssetPurchaseCost(ctx context.Context) (int64, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT asset.purchase_price
FROM accessory_assets asset
JOIN accessory_products product ON product.id=asset.product_id
WHERE asset.purchase_id IS NULL AND product.inventory_strategy='individual'`)
	if err != nil {
		return 0, fmt.Errorf("calculate manual asset purchase cost: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var total int64
	for rows.Next() {
		var purchasePrice sql.NullString
		if err := rows.Scan(&purchasePrice); err != nil {
			return 0, fmt.Errorf("calculate manual asset purchase cost: %w", err)
		}
		if err := addParsedMoney(&total, purchasePrice.String, 1); err != nil {
			return 0, fmt.Errorf("calculate manual asset purchase cost: %w", err)
		}
	}
	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("calculate manual asset purchase cost: %w", err)
	}
	return total, nil
}

func addParsedMoney(total *int64, value string, quantity int64) error {
	cents, ok := parseMoneyCents(value)
	if !ok {
		return nil
	}
	lineTotal, err := checkedMoneyMultiply(cents, quantity)
	if err != nil {
		return err
	}
	*total, err = checkedMoneyAdd(*total, lineTotal)
	return err
}

func checkedMoneyMultiply(value, quantity int64) (int64, error) {
	if value < 0 || quantity < 0 {
		return 0, errors.New("negative money operand")
	}
	if value != 0 && quantity > math.MaxInt64/value {
		return 0, errors.New("money multiplication overflow")
	}
	return value * quantity, nil
}

func checkedMoneyAdd(left, right int64) (int64, error) {
	if left < 0 || right < 0 {
		return 0, errors.New("negative money operand")
	}
	if left > math.MaxInt64-right {
		return 0, errors.New("money addition overflow")
	}
	return left + right, nil
}
