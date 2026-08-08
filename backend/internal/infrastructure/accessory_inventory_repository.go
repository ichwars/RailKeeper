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
		strategy, err := accessoryInventoryStrategy(ctx, tx, productID)
		if err != nil {
			return err
		}
		if strategy != domain.AccessoryInventoryQuantity &&
			strategy != domain.AccessoryInventoryQuantityLaterIndividual {
			return application.ErrAccessoryTrackingMode
		}
		if err := requireActiveStorageLocation(ctx, tx, input.LocationID); err != nil {
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
		if err := insertAccessoryStockMovement(ctx, tx, productID, input.LocationID,
			"adjustment", input.Delta, "", "", actor, "", now); err != nil {
			return err
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
	strategy, err := accessoryInventoryStrategy(ctx, r.db, productID)
	if err != nil {
		return nil, err
	}
	if strategy != domain.AccessoryInventoryIndividual &&
		strategy != domain.AccessoryInventoryQuantityLaterIndividual {
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
		if err := requireActiveStorageLocation(ctx, tx, input.StorageLocationID); err != nil {
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
		var currentLifecycle domain.AccessoryLifecycle
		if err := tx.QueryRowContext(ctx, `SELECT lifecycle_state FROM accessory_assets WHERE id=?`, id).
			Scan(&currentLifecycle); errors.Is(err, sql.ErrNoRows) {
			return application.ErrAccessoryNotFound
		} else if err != nil {
			return fmt.Errorf("read accessory asset lifecycle: %w", err)
		}
		if currentLifecycle == domain.AccessoryLifecycleReserved ||
			currentLifecycle == domain.AccessoryLifecycleInstalled {
			return application.ErrAccessoryConflict
		}
		if err := requireActiveStorageLocation(ctx, tx, input.StorageLocationID); err != nil {
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

const accessoryAssetSelect = `SELECT id, product_id, COALESCE(purchase_id, ''), COALESCE(inventory_number, ''), serial_number, condition_state, lifecycle_state, COALESCE(storage_location_id, ''), purchase_date, purchase_price, warranty_until, notes, created_at, updated_at FROM accessory_assets`

func scanAccessoryAsset(scanner rowScanner) (*application.AccessoryAsset, error) {
	asset := &application.AccessoryAsset{}
	err := scanner.Scan(&asset.ID, &asset.ProductID, &asset.PurchaseID, &asset.InventoryNumber, &asset.SerialNumber,
		&asset.Condition, &asset.Lifecycle, &asset.StorageLocationID, &asset.PurchaseDate,
		&asset.PurchasePrice, &asset.WarrantyUntil, &asset.Notes, &asset.CreatedAt, &asset.UpdatedAt)
	return asset, err
}

func (r *AccessoryRepository) ListStockMovements(
	ctx context.Context,
	productID string,
) ([]application.AccessoryStockMovement, error) {
	if _, err := accessoryInventoryStrategy(ctx, r.db, productID); err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(ctx, `
SELECT id, product_id, location_id, movement_type, quantity, source_type, source_id, actor, note, created_at
FROM accessory_stock_movements
WHERE product_id=?
ORDER BY created_at DESC, rowid DESC`, productID)
	if err != nil {
		return nil, fmt.Errorf("list accessory stock movements: %w", err)
	}
	defer func() { _ = rows.Close() }()
	movements := []application.AccessoryStockMovement{}
	for rows.Next() {
		movement := application.AccessoryStockMovement{}
		if err := rows.Scan(&movement.ID, &movement.ProductID, &movement.LocationID,
			&movement.MovementType, &movement.Quantity, &movement.SourceType, &movement.SourceID,
			&movement.Actor, &movement.Note, &movement.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan accessory stock movement: %w", err)
		}
		movements = append(movements, movement)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate accessory stock movements: %w", err)
	}
	return movements, nil
}

func (r *AccessoryRepository) TransferStock(
	ctx context.Context,
	productID string,
	input application.TransferAccessoryStockInput,
	actor string,
) (*application.AccessoryStockSummary, error) {
	now := timestamp()
	transferID := randomID()
	err := r.withTx(ctx, func(tx *sql.Tx) error {
		strategy, err := accessoryInventoryStrategy(ctx, tx, productID)
		if err != nil {
			return err
		}
		if strategy != domain.AccessoryInventoryQuantity &&
			strategy != domain.AccessoryInventoryQuantityLaterIndividual {
			return application.ErrAccessoryTrackingMode
		}
		if err := requireActiveStorageLocation(ctx, tx, input.FromLocationID); err != nil {
			return err
		}
		if err := requireActiveStorageLocation(ctx, tx, input.ToLocationID); err != nil {
			return err
		}
		result, err := tx.ExecContext(ctx, `
UPDATE accessory_stock SET quantity=quantity-?, updated_at=?
WHERE product_id=? AND location_id=? AND quantity>=?`, input.Quantity, now, productID,
			input.FromLocationID, input.Quantity)
		if err != nil {
			return fmt.Errorf("decrement transferred accessory stock: %w", err)
		}
		if affected, err := result.RowsAffected(); err != nil {
			return fmt.Errorf("read transferred stock decrement: %w", err)
		} else if affected == 0 {
			return application.ErrAccessoryInsufficientStock
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO accessory_stock(product_id, location_id, quantity, updated_at)
VALUES(?, ?, ?, ?)
ON CONFLICT(product_id, location_id) DO UPDATE
SET quantity=quantity+excluded.quantity, updated_at=excluded.updated_at`, productID,
			input.ToLocationID, input.Quantity, now); err != nil {
			return fmt.Errorf("increment transferred accessory stock: %w", err)
		}
		if err := insertAccessoryStockMovement(ctx, tx, productID, input.FromLocationID,
			"transfer_out", -input.Quantity, "transfer", transferID, actor, input.Note, now); err != nil {
			return err
		}
		if err := insertAccessoryStockMovement(ctx, tx, productID, input.ToLocationID,
			"transfer_in", input.Quantity, "transfer", transferID, actor, input.Note, now); err != nil {
			return err
		}
		details, err := json.Marshal(map[string]any{
			"fromLocationId": input.FromLocationID,
			"toLocationId":   input.ToLocationID,
			"quantity":       input.Quantity,
		})
		if err != nil {
			return fmt.Errorf("encode accessory transfer audit details: %w", err)
		}
		return writeAccessoryAudit(ctx, tx, "AccessoryStockTransferred", "accessory_product",
			productID, actor, now, string(details))
	})
	if err != nil {
		return nil, err
	}
	return r.GetStock(ctx, productID)
}

func (r *AccessoryRepository) Individualize(
	ctx context.Context,
	productID string,
	input application.IndividualizeAccessoryInput,
	actor string,
) (*application.AccessoryAsset, error) {
	now := timestamp()
	assetID := randomID()
	err := r.withTx(ctx, func(tx *sql.Tx) error {
		strategy, err := accessoryInventoryStrategy(ctx, tx, productID)
		if err != nil {
			return err
		}
		if strategy != domain.AccessoryInventoryQuantityLaterIndividual {
			return application.ErrAccessoryTrackingMode
		}
		if err := requireActiveStorageLocation(ctx, tx, input.LocationID); err != nil {
			return err
		}
		result, err := tx.ExecContext(ctx, `
UPDATE accessory_stock SET quantity=quantity-1, updated_at=?
WHERE product_id=? AND location_id=? AND quantity>=1`, now, productID, input.LocationID)
		if err != nil {
			return fmt.Errorf("decrement individualized accessory stock: %w", err)
		}
		if affected, err := result.RowsAffected(); err != nil {
			return fmt.Errorf("read individualized stock decrement: %w", err)
		} else if affected == 0 {
			return application.ErrAccessoryInsufficientStock
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO accessory_assets(
  id, product_id, inventory_number, serial_number, condition_state, lifecycle_state,
  storage_location_id, purchase_date, purchase_price, warranty_until, notes, created_at, updated_at
) VALUES(?, ?, NULLIF(?, ''), ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, assetID, productID,
			input.Asset.InventoryNumber, input.Asset.SerialNumber, input.Asset.Condition,
			input.Asset.Lifecycle, input.LocationID, input.Asset.PurchaseDate, input.Asset.PurchasePrice,
			input.Asset.WarrantyUntil, input.Asset.Notes, now, now); err != nil {
			if isSQLiteConstraint(err) {
				return application.ErrAccessoryConflict
			}
			return fmt.Errorf("insert individualized accessory asset: %w", err)
		}
		if err := insertAccessoryStockMovement(ctx, tx, productID, input.LocationID,
			"individualization", -1, "asset", assetID, actor, "", now); err != nil {
			return err
		}
		details, err := json.Marshal(map[string]any{
			"locationId": input.LocationID,
			"productId":  productID,
		})
		if err != nil {
			return fmt.Errorf("encode accessory individualization audit details: %w", err)
		}
		return writeAccessoryAudit(ctx, tx, "AccessoryAssetIndividualized", "accessory_asset",
			assetID, actor, now, string(details))
	})
	if err != nil {
		return nil, err
	}
	asset, err := scanAccessoryAsset(r.db.QueryRowContext(ctx, accessoryAssetSelect+` WHERE id=?`, assetID))
	if err != nil {
		return nil, fmt.Errorf("get individualized accessory asset: %w", err)
	}
	return asset, nil
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

func accessoryInventoryStrategy(
	ctx context.Context,
	queryer accessoryQueryer,
	productID string,
) (domain.AccessoryInventoryStrategy, error) {
	var strategy domain.AccessoryInventoryStrategy
	if err := queryer.QueryRowContext(ctx,
		`SELECT inventory_strategy FROM accessory_products WHERE id=?`, productID).
		Scan(&strategy); errors.Is(err, sql.ErrNoRows) {
		return "", application.ErrAccessoryNotFound
	} else if err != nil {
		return "", fmt.Errorf("read accessory inventory strategy: %w", err)
	}
	return strategy, nil
}

func requireActiveStorageLocation(ctx context.Context, queryer accessoryQueryer, id string) error {
	if id == "" {
		return nil
	}
	var count, archived int
	if err := queryer.QueryRowContext(ctx, `
WITH RECURSIVE location_chain(id, parent_id, archived, path) AS (
  SELECT id, parent_id, archived, ',' || id || ',' FROM storage_locations WHERE id=?
  UNION ALL
  SELECT parent.id, parent.parent_id, parent.archived, chain.path || parent.id || ','
  FROM storage_locations parent
  JOIN location_chain chain ON parent.id=chain.parent_id
  WHERE instr(chain.path, ',' || parent.id || ',')=0
)
SELECT COUNT(*), COALESCE(MAX(archived), 0) FROM location_chain`, id).Scan(&count, &archived); err != nil {
		return fmt.Errorf("read accessory storage location: %w", err)
	}
	if count == 0 {
		return application.ErrAccessoryNotFound
	}
	if archived != 0 {
		return application.ErrAccessoryConflict
	}
	return nil
}

func insertAccessoryStockMovement(
	ctx context.Context,
	tx *sql.Tx,
	productID, locationID, movementType string,
	quantity int,
	sourceType, sourceID, actor, note, createdAt string,
) error {
	if _, err := tx.ExecContext(ctx, `
INSERT INTO accessory_stock_movements(
  id, product_id, location_id, movement_type, quantity, source_type, source_id, actor, note, created_at
) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, randomID(), productID, locationID, movementType,
		quantity, sourceType, sourceID, actor, note, createdAt); err != nil {
		return fmt.Errorf("insert accessory stock movement: %w", err)
	}
	return nil
}
