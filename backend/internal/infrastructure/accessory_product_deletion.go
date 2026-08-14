package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"railkeeper/backend/internal/application"
)

func (r *AccessoryRepository) DeleteProduct(
	ctx context.Context,
	id string,
	actor string,
) ([]string, error) {
	var blobIDs []string
	err := r.withTx(ctx, func(tx *sql.Tx) error {
		if err := reserveAccessoryWriteTransaction(ctx, tx); err != nil {
			return err
		}
		exists, err := accessoryRecordExists(ctx, tx,
			`SELECT COUNT(*) FROM accessory_products WHERE id = ?`, id)
		if err != nil {
			return fmt.Errorf("check accessory product: %w", err)
		}
		if !exists {
			return application.ErrAccessoryNotFound
		}

		var blocked bool
		err = tx.QueryRowContext(ctx, `
SELECT
  EXISTS(SELECT 1 FROM accessory_stock WHERE product_id = ? AND quantity <> 0)
  OR EXISTS(SELECT 1 FROM accessory_assets WHERE product_id = ?)
  OR EXISTS(SELECT 1 FROM accessory_stock_movements WHERE product_id = ?)
  OR EXISTS(SELECT 1 FROM accessory_purchases WHERE product_id = ?)
  OR EXISTS(SELECT 1 FROM accessory_reservations WHERE product_id = ?)
  OR EXISTS(SELECT 1 FROM accessory_installations WHERE product_id = ?)
  OR EXISTS(SELECT 1 FROM layout_technical_positions WHERE product_id = ?)
`, id, id, id, id, id, id, id).Scan(&blocked)
		if err != nil {
			return fmt.Errorf("check accessory product deletion references: %w", err)
		}
		if blocked {
			return application.ErrAccessoryDeleteBlocked
		}

		rows, err := tx.QueryContext(ctx, `
SELECT file_blob_id
FROM accessory_documents
WHERE product_id = ?
ORDER BY id
`, id)
		if err != nil {
			return fmt.Errorf("list accessory product blobs: %w", err)
		}
		for rows.Next() {
			var blobID string
			if err := rows.Scan(&blobID); err != nil {
				_ = rows.Close()
				return fmt.Errorf("scan accessory product blob: %w", err)
			}
			blobIDs = append(blobIDs, blobID)
		}
		if err := rows.Err(); err != nil {
			_ = rows.Close()
			return fmt.Errorf("iterate accessory product blobs: %w", err)
		}
		if err := rows.Close(); err != nil {
			return fmt.Errorf("close accessory product blobs: %w", err)
		}

		for _, statement := range []string{
			`DELETE FROM accessory_documents WHERE product_id = ?`,
			`DELETE FROM accessory_stock WHERE product_id = ?`,
			`DELETE FROM accessory_products WHERE id = ?`,
		} {
			if _, err := tx.ExecContext(ctx, statement, id); err != nil {
				return fmt.Errorf("delete accessory product data: %w", err)
			}
		}
		return writeAccessoryAudit(ctx, tx, "AccessoryProductDeleted", "accessory_product", id,
			actor, time.Now().UTC().Format(time.RFC3339Nano), "{}")
	})
	if err != nil {
		return nil, err
	}
	return blobIDs, nil
}
