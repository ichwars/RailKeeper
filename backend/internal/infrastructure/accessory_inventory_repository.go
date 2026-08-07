package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/domain"
)

func (r *AccessoryRepository) AdjustStock(
	ctx context.Context,
	productID string,
	input application.StockAdjustmentInput,
	actor string,
) (*application.AccessoryStockSummary, error) {
	now := timestamp()
	err := r.withTx(ctx, func(tx *sql.Tx) error {
		mode, err := accessoryTrackingMode(ctx, tx, productID)
		if err != nil {
			return err
		}
		if mode != domain.AccessoryTrackingModeQuantity {
			return application.ErrAccessoryTrackingMode
		}
		if err := requireStorageLocation(ctx, tx, input.LocationID); err != nil {
			return err
		}
		current := 0
		if err := tx.QueryRowContext(ctx, `
SELECT quantity FROM accessory_stock WHERE product_id=? AND location_id=?`, productID, input.LocationID).
			Scan(&current); err != nil && !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("read accessory stock level: %w", err)
		}
		if current+input.Delta < 0 {
			return application.ErrAccessoryInsufficientStock
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO accessory_stock(product_id, location_id, quantity, updated_at)
VALUES(?, ?, ?, ?)
ON CONFLICT(product_id, location_id) DO UPDATE SET quantity=excluded.quantity, updated_at=excluded.updated_at`,
			productID, input.LocationID, current+input.Delta, now); err != nil {
			return fmt.Errorf("adjust accessory stock: %w", err)
		}
		details, err := json.Marshal(map[string]any{"delta": input.Delta, "locationId": input.LocationID})
		if err != nil {
			return fmt.Errorf("encode accessory stock audit details: %w", err)
		}
		return writeAccessoryAudit(ctx, tx, "AccessoryStockAdjusted", "accessory_product", productID,
			actor, now, string(details))
	})
	if err != nil {
		return nil, err
	}
	return r.GetStock(ctx, productID)
}

func (r *AccessoryRepository) GetStock(ctx context.Context, productID string) (*application.AccessoryStockSummary, error) {
	mode, err := accessoryTrackingMode(ctx, r.db, productID)
	if err != nil {
		return nil, err
	}
	summary := &application.AccessoryStockSummary{
		ProductID: productID, TrackingMode: mode, Locations: []application.AccessoryStockLevel{},
	}
	rows, err := r.db.QueryContext(ctx, `
SELECT stock.location_id, location.name, stock.quantity, stock.updated_at
FROM accessory_stock stock
JOIN storage_locations location ON location.id=stock.location_id
WHERE stock.product_id=?
ORDER BY location.name COLLATE NOCASE, stock.location_id`, productID)
	if err != nil {
		return nil, fmt.Errorf("list accessory stock: %w", err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		level := application.AccessoryStockLevel{}
		if err := rows.Scan(&level.LocationID, &level.LocationName, &level.Quantity, &level.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan accessory stock: %w", err)
		}
		summary.TotalQuantity += level.Quantity
		summary.Locations = append(summary.Locations, level)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate accessory stock: %w", err)
	}
	return summary, nil
}

func (r *AccessoryRepository) ListAssets(ctx context.Context, productID string) ([]application.AccessoryAsset, error) {
	mode, err := accessoryTrackingMode(ctx, r.db, productID)
	if err != nil {
		return nil, err
	}
	if mode != domain.AccessoryTrackingModeIndividual {
		return nil, application.ErrAccessoryTrackingMode
	}
	rows, err := r.db.QueryContext(ctx, accessoryAssetSelect+` WHERE product_id=? ORDER BY inventory_number COLLATE NOCASE, id`,
		productID)
	if err != nil {
		return nil, fmt.Errorf("list accessory assets: %w", err)
	}
	defer func() { _ = rows.Close() }()
	assets := []application.AccessoryAsset{}
	for rows.Next() {
		asset, err := scanAccessoryAsset(rows)
		if err != nil {
			return nil, fmt.Errorf("scan accessory asset: %w", err)
		}
		assets = append(assets, *asset)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate accessory assets: %w", err)
	}
	return assets, nil
}

func (r *AccessoryRepository) CreateAsset(
	ctx context.Context,
	productID string,
	input application.CreateAccessoryAssetInput,
	actor string,
) (*application.AccessoryAsset, error) {
	now := timestamp()
	asset := &application.AccessoryAsset{
		ID: randomID(), ProductID: productID, InventoryNumber: input.InventoryNumber,
		SerialNumber: input.SerialNumber, Condition: input.Condition, Lifecycle: input.Lifecycle,
		StorageLocationID: input.StorageLocationID, PurchaseDate: input.PurchaseDate,
		PurchasePrice: input.PurchasePrice, WarrantyUntil: input.WarrantyUntil, Notes: input.Notes,
		CreatedAt: now, UpdatedAt: now,
	}
	err := r.withTx(ctx, func(tx *sql.Tx) error {
		mode, err := accessoryTrackingMode(ctx, tx, productID)
		if err != nil {
			return err
		}
		if mode != domain.AccessoryTrackingModeIndividual {
			return application.ErrAccessoryTrackingMode
		}
		if err := requireStorageLocation(ctx, tx, input.StorageLocationID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO accessory_assets(
  id, product_id, inventory_number, serial_number, condition_state, lifecycle_state,
  storage_location_id, purchase_date, purchase_price, warranty_until, notes, created_at, updated_at
) VALUES(?, ?, NULLIF(?, ''), ?, ?, ?, NULLIF(?, ''), ?, ?, ?, ?, ?, ?)`, asset.ID, asset.ProductID,
			asset.InventoryNumber, asset.SerialNumber, asset.Condition, asset.Lifecycle, asset.StorageLocationID,
			asset.PurchaseDate, asset.PurchasePrice, asset.WarrantyUntil, asset.Notes, now, now); err != nil {
			if isSQLiteConstraint(err) {
				return application.ErrAccessoryConflict
			}
			return fmt.Errorf("insert accessory asset: %w", err)
		}
		return writeAccessoryAudit(ctx, tx, "AccessoryAssetCreated", "accessory_asset", asset.ID, actor, now, "{}")
	})
	if err != nil {
		return nil, err
	}
	return asset, nil
}

func (r *AccessoryRepository) UpdateAsset(
	ctx context.Context,
	id string,
	input application.UpdateAccessoryAssetInput,
	actor string,
) (*application.AccessoryAsset, error) {
	now := timestamp()
	err := r.withTx(ctx, func(tx *sql.Tx) error {
		var productID string
		if err := tx.QueryRowContext(ctx, `SELECT product_id FROM accessory_assets WHERE id=?`, id).Scan(&productID); errors.Is(err, sql.ErrNoRows) {
			return application.ErrAccessoryNotFound
		} else if err != nil {
			return fmt.Errorf("read accessory asset product: %w", err)
		}
		if err := requireStorageLocation(ctx, tx, input.StorageLocationID); err != nil {
			return err
		}
		result, err := tx.ExecContext(ctx, `
UPDATE accessory_assets
SET inventory_number=NULLIF(?, ''), serial_number=?, condition_state=?, lifecycle_state=?,
    storage_location_id=NULLIF(?, ''), purchase_date=?, purchase_price=?, warranty_until=?, notes=?, updated_at=?
WHERE id=?`, input.InventoryNumber, input.SerialNumber, input.Condition, input.Lifecycle,
			input.StorageLocationID, input.PurchaseDate, input.PurchasePrice, input.WarrantyUntil, input.Notes, now, id)
		if err != nil {
			if isSQLiteConstraint(err) {
				return application.ErrAccessoryConflict
			}
			return fmt.Errorf("update accessory asset: %w", err)
		}
		if err := requireAccessoryUpdated(result); err != nil {
			return err
		}
		return writeAccessoryAudit(ctx, tx, "AccessoryAssetUpdated", "accessory_asset", id, actor, now, "{}")
	})
	if err != nil {
		return nil, err
	}
	asset, err := scanAccessoryAsset(r.db.QueryRowContext(ctx, accessoryAssetSelect+` WHERE id=?`, id))
	if err != nil {
		return nil, fmt.Errorf("get updated accessory asset: %w", err)
	}
	return asset, nil
}

const accessoryAssetSelect = `SELECT id, product_id, COALESCE(inventory_number, ''), serial_number, condition_state, lifecycle_state, COALESCE(storage_location_id, ''), purchase_date, purchase_price, warranty_until, notes, created_at, updated_at FROM accessory_assets`

func scanAccessoryAsset(scanner rowScanner) (*application.AccessoryAsset, error) {
	asset := &application.AccessoryAsset{}
	err := scanner.Scan(&asset.ID, &asset.ProductID, &asset.InventoryNumber, &asset.SerialNumber,
		&asset.Condition, &asset.Lifecycle, &asset.StorageLocationID, &asset.PurchaseDate,
		&asset.PurchasePrice, &asset.WarrantyUntil, &asset.Notes, &asset.CreatedAt, &asset.UpdatedAt)
	return asset, err
}

type accessoryQueryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func accessoryTrackingMode(
	ctx context.Context,
	queryer accessoryQueryer,
	productID string,
) (domain.AccessoryTrackingMode, error) {
	var mode domain.AccessoryTrackingMode
	if err := queryer.QueryRowContext(ctx, `SELECT tracking_mode FROM accessory_products WHERE id=?`, productID).
		Scan(&mode); errors.Is(err, sql.ErrNoRows) {
		return "", application.ErrAccessoryNotFound
	} else if err != nil {
		return "", fmt.Errorf("read accessory tracking mode: %w", err)
	}
	return mode, nil
}
