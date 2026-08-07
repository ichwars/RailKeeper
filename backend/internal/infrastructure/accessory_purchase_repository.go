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

func (r *AccessoryRepository) ListPurchases(
	ctx context.Context,
	productID string,
) ([]application.AccessoryPurchase, error) {
	if _, err := accessoryInventoryStrategy(ctx, r.db, productID); err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(ctx, accessoryPurchaseSelect+`
WHERE product_id=?
ORDER BY purchased_at DESC, created_at DESC, id`, productID)
	if err != nil {
		return nil, fmt.Errorf("list accessory purchases: %w", err)
	}
	defer func() { _ = rows.Close() }()
	purchases := []application.AccessoryPurchase{}
	for rows.Next() {
		purchase, err := scanAccessoryPurchase(rows)
		if err != nil {
			return nil, fmt.Errorf("scan accessory purchase: %w", err)
		}
		purchases = append(purchases, *purchase)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate accessory purchases: %w", err)
	}
	return purchases, nil
}

func (r *AccessoryRepository) CreatePurchase(
	ctx context.Context,
	productID string,
	input application.CreateAccessoryPurchaseInput,
	actor string,
) (*application.AccessoryPurchase, error) {
	now := timestamp()
	purchaseID := randomID()
	err := r.withTx(ctx, func(tx *sql.Tx) error {
		strategy, err := accessoryInventoryStrategy(ctx, tx, productID)
		if err != nil {
			return err
		}
		if input.StorageLocationID != "" {
			if input.BookToStock {
				err = requireActiveStorageLocation(ctx, tx, input.StorageLocationID)
			} else {
				err = requireStorageLocation(ctx, tx, input.StorageLocationID)
			}
			if err != nil {
				return err
			}
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO accessory_purchases(
  id, product_id, destination_location_id, quantity, purchased_at, supplier, unit_price,
  currency, invoice_number, warranty_until, booked_to_stock, note, created_by, created_at, updated_at
) VALUES(?, ?, NULLIF(?, ''), ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, purchaseID, productID,
			input.StorageLocationID, input.Quantity, input.PurchasedAt, input.Supplier, input.UnitPrice,
			input.Currency, input.InvoiceNumber, input.WarrantyUntil, boolToInt(input.BookToStock),
			input.Notes, actor, now, now); err != nil {
			if isSQLiteConstraint(err) {
				return application.ErrAccessoryConflict
			}
			return fmt.Errorf("insert accessory purchase: %w", err)
		}
		if input.BookToStock {
			switch strategy {
			case domain.AccessoryInventoryQuantity,
				domain.AccessoryInventoryQuantityLaterIndividual:
				if err := bookAccessoryQuantityPurchase(ctx, tx, productID, purchaseID, input, actor, now); err != nil {
					return err
				}
			case domain.AccessoryInventoryIndividual:
				if err := bookAccessoryIndividualPurchase(ctx, tx, productID, purchaseID, input, now); err != nil {
					return err
				}
			default:
				return application.ErrAccessoryTrackingMode
			}
		}
		details, err := json.Marshal(map[string]any{
			"bookToStock": input.BookToStock,
			"locationId":  input.StorageLocationID,
			"quantity":    input.Quantity,
		})
		if err != nil {
			return fmt.Errorf("encode accessory purchase audit details: %w", err)
		}
		return writeAccessoryAudit(ctx, tx, "AccessoryPurchaseCreated", "accessory_purchase",
			purchaseID, actor, now, string(details))
	})
	if err != nil {
		return nil, err
	}
	return getAccessoryPurchase(ctx, r.db, purchaseID)
}

func bookAccessoryQuantityPurchase(
	ctx context.Context,
	tx *sql.Tx,
	productID, purchaseID string,
	input application.CreateAccessoryPurchaseInput,
	actor, now string,
) error {
	if _, err := tx.ExecContext(ctx, `
INSERT INTO accessory_stock(product_id, location_id, quantity, updated_at)
VALUES(?, ?, ?, ?)
ON CONFLICT(product_id, location_id) DO UPDATE
SET quantity=quantity+excluded.quantity, updated_at=excluded.updated_at`, productID,
		input.StorageLocationID, input.Quantity, now); err != nil {
		return fmt.Errorf("book purchased accessory stock: %w", err)
	}
	return insertAccessoryStockMovement(ctx, tx, productID, input.StorageLocationID,
		"purchase", input.Quantity, "purchase", purchaseID, actor, input.Notes, now)
}

func bookAccessoryIndividualPurchase(
	ctx context.Context,
	tx *sql.Tx,
	productID, purchaseID string,
	input application.CreateAccessoryPurchaseInput,
	now string,
) error {
	for index := 0; index < input.Quantity; index++ {
		if _, err := tx.ExecContext(ctx, `
INSERT INTO accessory_assets(
  id, product_id, purchase_id, condition_state, lifecycle_state, storage_location_id,
  purchase_date, purchase_price, warranty_until, created_at, updated_at
) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, randomID(), productID, purchaseID,
			domain.AccessoryConditionUnknown, domain.AccessoryLifecycleStored, input.StorageLocationID,
			input.PurchasedAt, input.UnitPrice, input.WarrantyUntil, now, now); err != nil {
			if isSQLiteConstraint(err) {
				return application.ErrAccessoryConflict
			}
			return fmt.Errorf("insert purchased accessory asset: %w", err)
		}
	}
	return nil
}

const accessoryPurchaseSelect = `SELECT id, product_id, COALESCE(destination_location_id, ''), quantity,
purchased_at, supplier, unit_price, currency, invoice_number, warranty_until, booked_to_stock,
note, created_by, created_at, updated_at FROM accessory_purchases`

func getAccessoryPurchase(
	ctx context.Context,
	queryer accessoryQueryer,
	id string,
) (*application.AccessoryPurchase, error) {
	purchase, err := scanAccessoryPurchase(queryer.QueryRowContext(ctx, accessoryPurchaseSelect+` WHERE id=?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, application.ErrAccessoryNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get accessory purchase: %w", err)
	}
	return purchase, nil
}

func scanAccessoryPurchase(scanner rowScanner) (*application.AccessoryPurchase, error) {
	purchase := &application.AccessoryPurchase{}
	var bookedToStock int
	err := scanner.Scan(&purchase.ID, &purchase.ProductID, &purchase.StorageLocationID,
		&purchase.Quantity, &purchase.PurchasedAt, &purchase.Supplier, &purchase.UnitPrice,
		&purchase.Currency, &purchase.InvoiceNumber, &purchase.WarrantyUntil, &bookedToStock,
		&purchase.Notes, &purchase.CreatedBy, &purchase.CreatedAt, &purchase.UpdatedAt)
	purchase.BookToStock = bookedToStock != 0
	return purchase, err
}
