package overviewdata

import (
	"context"
	"database/sql"
	"fmt"

	"railkeeper/backend/internal/inventory/overview"
)

type Store struct{ db *sql.DB }

func New(db *sql.DB) *Store { return &Store{db: db} }

var _ overview.Source = (*Store)(nil)

func (s *Store) VehiclePrices(ctx context.Context) ([]overview.VehiclePrice, error) {
	return readRows(ctx, s.db, `
SELECT vehicles.list_price, vehicles.purchase_price
FROM vehicles
LEFT JOIN vehicle_set_members ON vehicle_set_members.vehicle_id=vehicles.id
WHERE vehicle_set_members.vehicle_id IS NULL
UNION ALL
SELECT vehicle_sets.list_price, vehicle_sets.purchase_price
FROM vehicle_sets
WHERE EXISTS (SELECT 1 FROM vehicle_set_members WHERE vehicle_set_members.vehicle_set_id=vehicle_sets.id)
`, "calculate vehicle values", func(rows *sql.Rows) (overview.VehiclePrice, error) {

		var a, b sql.NullString
		err := rows.Scan(&a, &b)
		return overview.VehiclePrice{List: a.String, Purchase: b.String}, err

	})
}
func (s *Store) AccessoryOwnedPrices(ctx context.Context) ([]overview.OwnedPrice, error) {
	return readRows(ctx, s.db, `
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
LEFT JOIN installations ON installations.product_id=product.id`, "calculate accessory list value", func(rows *sql.Rows) (overview.OwnedPrice, error) {

		var price sql.NullString
		var quantity int64
		err := rows.Scan(&price, &quantity)
		return overview.OwnedPrice{Price: price.String, Quantity: quantity}, err

	})
}
func (s *Store) AccessoryPurchases(ctx context.Context) ([]overview.Purchase, error) {
	return readRows(ctx, s.db, `SELECT quantity, unit_price, currency FROM accessory_purchases`, "calculate accessory purchase cost", func(rows *sql.Rows) (overview.Purchase, error) {

		var price, currency sql.NullString
		var quantity int64
		err := rows.Scan(&quantity, &price, &currency)
		return overview.Purchase{UnitPrice: price.String, Currency: currency.String, Quantity: quantity}, err

	})
}
func (s *Store) ManualAssetPrices(ctx context.Context) ([]string, error) {
	return readRows(ctx, s.db, `
SELECT asset.purchase_price
FROM accessory_assets asset
JOIN accessory_products product ON product.id=asset.product_id
WHERE asset.purchase_id IS NULL AND product.inventory_strategy='individual'`, "calculate manual asset purchase cost", func(rows *sql.Rows) (string, error) {

		var price sql.NullString
		err := rows.Scan(&price)
		return price.String, err

	})
}

func readRows[T any](ctx context.Context, db *sql.DB, query, operation string,
	scan func(*sql.Rows) (T, error)) ([]T, error) {
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}
	defer func() { _ = rows.Close() }()
	values := []T{}
	for rows.Next() {
		value, err := scan(rows)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", operation, err)
		}
		values = append(values, value)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}
	return values, nil
}
