package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/domain"
)

type AccessoryRepository struct {
	db *sql.DB
}

func NewAccessoryRepository(db *sql.DB) *AccessoryRepository {
	return &AccessoryRepository{db: db}
}

func (r *AccessoryRepository) ListProducts(ctx context.Context, query string) ([]application.AccessoryProduct, error) {
	rows, err := r.db.QueryContext(ctx, accessoryProductSelect+`
WHERE ?='' OR manufacturer LIKE '%' || ? || '%' COLLATE NOCASE
   OR article_number LIKE '%' || ? || '%' COLLATE NOCASE
   OR name LIKE '%' || ? || '%' COLLATE NOCASE
   OR category LIKE '%' || ? || '%' COLLATE NOCASE
ORDER BY manufacturer COLLATE NOCASE, name COLLATE NOCASE, id`, query, query, query, query, query)
	if err != nil {
		return nil, fmt.Errorf("list accessory products: %w", err)
	}
	defer func() { _ = rows.Close() }()
	products := []application.AccessoryProduct{}
	for rows.Next() {
		product, err := scanAccessoryProduct(rows)
		if err != nil {
			return nil, fmt.Errorf("scan accessory product: %w", err)
		}
		products = append(products, *product)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate accessory products: %w", err)
	}
	productIDs := make([]string, len(products))
	for index := range products {
		productIDs[index] = products[index].ID
	}
	attributes, err := loadAccessoryAttributes(ctx, r.db, productIDs)
	if err != nil {
		return nil, err
	}
	for index := range products {
		products[index].Attributes = attributes[products[index].ID]
	}
	return products, nil
}

func (r *AccessoryRepository) GetProduct(ctx context.Context, id string) (*application.AccessoryProduct, error) {
	product, err := scanAccessoryProduct(r.db.QueryRowContext(ctx, accessoryProductSelect+` WHERE id=?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, application.ErrAccessoryNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get accessory product: %w", err)
	}
	attributes, err := loadAccessoryAttributes(ctx, r.db, []string{id})
	if err != nil {
		return nil, err
	}
	product.Attributes = attributes[id]
	var primaryDocumentID string
	err = r.db.QueryRowContext(ctx, `
SELECT id FROM accessory_documents
WHERE product_id=? AND category='image' AND is_primary=1`, id).Scan(&primaryDocumentID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("get accessory product primary image: %w", err)
	}
	if primaryDocumentID != "" {
		product.PrimaryImageURL = fmt.Sprintf(
			"/api/v1/accessory-products/%s/documents/%s/download",
			product.ID,
			primaryDocumentID,
		)
	}
	return product, nil
}

func (r *AccessoryRepository) AccessorySubtypeActive(ctx context.Context, key string) (bool, error) {
	var active bool
	if err := r.db.QueryRowContext(ctx, `
SELECT EXISTS(
  SELECT 1 FROM master_data_entries
  WHERE type='accessory_subtype' AND key=? AND active=1
)
`, key).Scan(&active); err != nil {
		return false, fmt.Errorf("check active accessory subtype: %w", err)
	}
	return active, nil
}

func (r *AccessoryRepository) CreateProduct(
	ctx context.Context,
	input application.CreateAccessoryProductInput,
	actor string,
) (*application.AccessoryProduct, error) {
	now := timestamp()
	productID := randomID()
	err := r.withTx(ctx, func(tx *sql.Tx) error {
		gauges, alternativeNumbers, keywords, err := accessoryProductJSON(input)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO accessory_products(
	  id, manufacturer, article_number, name, category, tracking_mode, description,
	  ean, manufacturer_status, article_type, subtype, gauges_json, scale, package_quantity,
	  stock_unit, minimum_stock, inventory_strategy, manufacturer_url, product_url,
	  alternative_numbers_json, keywords_json, compatibility_notes, internal_notes, archived,
	  created_at, updated_at
	) VALUES(?, ?, ?, ?, ?, ?, ?, ?, COALESCE(NULLIF(?, ''), 'unknown'), ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			productID, input.Manufacturer, input.ArticleNumber, input.Name, input.Category, input.TrackingMode,
			input.Description, input.EAN, input.ManufacturerStatus, input.ArticleType, input.Subtype, gauges,
			input.Scale, input.PackageQuantity, input.StockUnit, input.MinimumStock, input.InventoryStrategy,
			input.ManufacturerURL, input.ProductURL, alternativeNumbers, keywords, input.CompatibilityNotes,
			input.InternalNotes, boolToInt(input.Archived), now, now); err != nil {
			if isSQLiteConstraint(err) {
				return application.ErrAccessoryConflict
			}
			return fmt.Errorf("insert accessory product: %w", err)
		}
		if err := replaceAccessoryAttributes(ctx, tx, productID, input.Attributes, now); err != nil {
			return err
		}
		return writeAccessoryAudit(ctx, tx, "AccessoryProductCreated", "accessory_product", productID, actor, now, "{}")
	})
	if err != nil {
		return nil, err
	}
	return r.GetProduct(ctx, productID)
}

func (r *AccessoryRepository) UpdateProduct(
	ctx context.Context,
	id string,
	input application.UpdateAccessoryProductInput,
	actor string,
) (*application.AccessoryProduct, error) {
	now := timestamp()
	err := r.withTx(ctx, func(tx *sql.Tx) error {
		currentMode, err := accessoryTrackingMode(ctx, tx, id)
		if err != nil {
			return err
		}
		if currentMode != input.TrackingMode {
			var dependentCount int
			query := `
SELECT
  COALESCE((SELECT SUM(quantity) FROM accessory_stock WHERE product_id=?), 0) +
  COALESCE((SELECT SUM(quantity) FROM accessory_reservations WHERE product_id=? AND status='active'), 0) +
  COALESCE((SELECT SUM(quantity) FROM accessory_installations WHERE product_id=? AND removed_at IS NULL), 0)`
			args := []any{id, id, id}
			if currentMode == domain.AccessoryTrackingModeIndividual {
				query = `SELECT COUNT(*) FROM accessory_assets WHERE product_id=?`
				args = []any{id}
			}
			if err := tx.QueryRowContext(ctx, query, args...).Scan(&dependentCount); err != nil {
				return fmt.Errorf("check accessory tracking mode change: %w", err)
			}
			if dependentCount > 0 {
				return application.ErrAccessoryConflict
			}
		}
		gauges, alternativeNumbers, keywords, err := accessoryProductJSON(input.CreateAccessoryProductInput)
		if err != nil {
			return err
		}
		result, err := tx.ExecContext(ctx, `
UPDATE accessory_products
SET manufacturer=?, article_number=?, name=?, category=?, tracking_mode=?, description=?, ean=?,
    manufacturer_status=COALESCE(NULLIF(?, ''), 'unknown'), article_type=?, subtype=?, gauges_json=?, scale=?,
    package_quantity=?, stock_unit=?, minimum_stock=?, inventory_strategy=?, manufacturer_url=?, product_url=?,
    alternative_numbers_json=?, keywords_json=?, compatibility_notes=?, internal_notes=?, archived=?, updated_at=?
WHERE id=?`, input.Manufacturer, input.ArticleNumber, input.Name, input.Category, input.TrackingMode,
			input.Description, input.EAN, input.ManufacturerStatus, input.ArticleType, input.Subtype, gauges, input.Scale,
			input.PackageQuantity, input.StockUnit, input.MinimumStock, input.InventoryStrategy, input.ManufacturerURL,
			input.ProductURL, alternativeNumbers, keywords, input.CompatibilityNotes, input.InternalNotes,
			boolToInt(input.Archived), now, id)
		if err != nil {
			if isSQLiteConstraint(err) {
				return application.ErrAccessoryConflict
			}
			return fmt.Errorf("update accessory product: %w", err)
		}
		if err := requireAccessoryUpdated(result); err != nil {
			return err
		}
		if err := replaceAccessoryAttributes(ctx, tx, id, input.Attributes, now); err != nil {
			return err
		}
		return writeAccessoryAudit(ctx, tx, "AccessoryProductUpdated", "accessory_product", id, actor, now, "{}")
	})
	if err != nil {
		return nil, err
	}
	return r.GetProduct(ctx, id)
}

func (r *AccessoryRepository) SetProductArchived(
	ctx context.Context,
	id string,
	archived bool,
	actor string,
) (*application.AccessoryProduct, error) {
	now := timestamp()
	action := "AccessoryProductRestored"
	if archived {
		action = "AccessoryProductArchived"
	}
	err := r.withTx(ctx, func(tx *sql.Tx) error {
		result, err := tx.ExecContext(ctx,
			`UPDATE accessory_products SET archived=?, updated_at=? WHERE id=?`,
			boolToInt(archived), now, id,
		)
		if err != nil {
			return fmt.Errorf("set accessory product archived: %w", err)
		}
		if err := requireAccessoryUpdated(result); err != nil {
			return err
		}
		return writeAccessoryAudit(ctx, tx, action, "accessory_product", id, actor, now, "{}")
	})
	if err != nil {
		return nil, err
	}
	return r.GetProduct(ctx, id)
}

func (r *AccessoryRepository) ListLocations(ctx context.Context) ([]application.StorageLocation, error) {
	rows, err := r.db.QueryContext(ctx, storageLocationSelect+` ORDER BY name COLLATE NOCASE, id`)
	if err != nil {
		return nil, fmt.Errorf("list storage locations: %w", err)
	}
	defer func() { _ = rows.Close() }()
	locations := []application.StorageLocation{}
	for rows.Next() {
		location, err := scanStorageLocation(rows)
		if err != nil {
			return nil, fmt.Errorf("scan storage location: %w", err)
		}
		locations = append(locations, *location)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate storage locations: %w", err)
	}
	return locations, nil
}

func (r *AccessoryRepository) CreateLocation(
	ctx context.Context,
	input application.CreateStorageLocationInput,
	actor string,
) (*application.StorageLocation, error) {
	now := timestamp()
	location := &application.StorageLocation{
		ID: randomID(), ParentID: input.ParentID, Name: input.Name, Description: input.Description,
		Archived: input.Archived, CreatedAt: now, UpdatedAt: now,
	}
	err := r.withTx(ctx, func(tx *sql.Tx) error {
		if err := requireStorageLocation(ctx, tx, input.ParentID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO storage_locations(id, parent_id, name, description, archived, created_at, updated_at)
VALUES(?, NULLIF(?, ''), ?, ?, ?, ?, ?)`, location.ID, location.ParentID, location.Name,
			location.Description, boolToInt(location.Archived), now, now); err != nil {
			if isSQLiteConstraint(err) {
				return application.ErrAccessoryConflict
			}
			return fmt.Errorf("insert storage location: %w", err)
		}
		return writeAccessoryAudit(ctx, tx, "StorageLocationCreated", "storage_location", location.ID, actor, now, "{}")
	})
	if err != nil {
		return nil, err
	}
	return location, nil
}

func (r *AccessoryRepository) UpdateLocation(
	ctx context.Context,
	id string,
	input application.UpdateStorageLocationInput,
	actor string,
) (*application.StorageLocation, error) {
	now := timestamp()
	err := r.withTx(ctx, func(tx *sql.Tx) error {
		exists, err := accessoryRecordExists(
			ctx,
			tx,
			`SELECT COUNT(*) FROM storage_locations WHERE id=?`,
			id,
		)
		if err != nil {
			return err
		}
		if !exists {
			return application.ErrAccessoryNotFound
		}
		if err := requireStorageLocation(ctx, tx, input.ParentID); err != nil {
			return err
		}
		if input.ParentID != "" {
			var cycleCount int
			if err := tx.QueryRowContext(ctx, `
WITH RECURSIVE descendants(id) AS (
  SELECT id FROM storage_locations WHERE parent_id=?
  UNION ALL
  SELECT location.id FROM storage_locations location JOIN descendants ON location.parent_id=descendants.id
)
SELECT COUNT(*) FROM descendants WHERE id=?`, id, input.ParentID).Scan(&cycleCount); err != nil {
				return fmt.Errorf("check storage location cycle: %w", err)
			}
			if cycleCount > 0 {
				return application.ErrAccessoryValidation
			}
		}
		result, err := tx.ExecContext(ctx, `
UPDATE storage_locations SET parent_id=NULLIF(?, ''), name=?, description=?, archived=?, updated_at=? WHERE id=?`,
			input.ParentID, input.Name, input.Description, boolToInt(input.Archived), now, id)
		if err != nil {
			if isSQLiteConstraint(err) {
				return application.ErrAccessoryConflict
			}
			return fmt.Errorf("update storage location: %w", err)
		}
		if err := requireAccessoryUpdated(result); err != nil {
			return err
		}
		return writeAccessoryAudit(ctx, tx, "StorageLocationUpdated", "storage_location", id, actor, now, "{}")
	})
	if err != nil {
		return nil, err
	}
	location, err := scanStorageLocation(r.db.QueryRowContext(ctx, storageLocationSelect+` WHERE id=?`, id))
	if err != nil {
		return nil, fmt.Errorf("get updated storage location: %w", err)
	}
	return location, nil
}

const accessoryProductSelect = `SELECT
  id, manufacturer, article_number, name, category, tracking_mode, description,
  ean, manufacturer_status, article_type, subtype, gauges_json, scale, package_quantity,
  stock_unit, minimum_stock, inventory_strategy, manufacturer_url, product_url,
  alternative_numbers_json, keywords_json, compatibility_notes, internal_notes, archived,
  created_at, updated_at
FROM accessory_products`

const storageLocationSelect = `SELECT id, COALESCE(parent_id, ''), name, description, archived, created_at, updated_at FROM storage_locations`

func scanAccessoryProduct(scanner rowScanner) (*application.AccessoryProduct, error) {
	product := &application.AccessoryProduct{}
	var gauges, alternativeNumbers, keywords string
	var archived int
	err := scanner.Scan(&product.ID, &product.Manufacturer, &product.ArticleNumber, &product.Name,
		&product.Category, &product.TrackingMode, &product.Description, &product.EAN, &product.ManufacturerStatus,
		&product.ArticleType, &product.Subtype, &gauges, &product.Scale, &product.PackageQuantity,
		&product.StockUnit, &product.MinimumStock, &product.InventoryStrategy, &product.ManufacturerURL,
		&product.ProductURL, &alternativeNumbers, &keywords, &product.CompatibilityNotes, &product.InternalNotes,
		&archived, &product.CreatedAt, &product.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if err := decodeAccessoryStringArray(gauges, &product.Gauges); err != nil {
		return nil, err
	}
	if err := decodeAccessoryStringArray(alternativeNumbers, &product.AlternativeNumbers); err != nil {
		return nil, err
	}
	if err := decodeAccessoryStringArray(keywords, &product.Keywords); err != nil {
		return nil, err
	}
	product.Archived = archived != 0
	return product, err
}

func scanStorageLocation(scanner rowScanner) (*application.StorageLocation, error) {
	location := &application.StorageLocation{}
	var archived int
	err := scanner.Scan(&location.ID, &location.ParentID, &location.Name, &location.Description,
		&archived, &location.CreatedAt, &location.UpdatedAt)
	location.Archived = archived != 0
	return location, err
}

func requireStorageLocation(ctx context.Context, tx *sql.Tx, id string) error {
	if id == "" {
		return nil
	}
	exists, err := accessoryRecordExists(
		ctx,
		tx,
		`SELECT COUNT(*) FROM storage_locations WHERE id=?`,
		id,
	)
	if err != nil {
		return err
	}
	if !exists {
		return application.ErrAccessoryNotFound
	}
	return nil
}

func accessoryRecordExists(ctx context.Context, tx *sql.Tx, query, id string) (bool, error) {
	var count int
	if err := tx.QueryRowContext(ctx, query, id).Scan(&count); err != nil {
		return false, fmt.Errorf("check accessory record: %w", err)
	}
	return count > 0, nil
}

func requireAccessoryUpdated(result sql.Result) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read accessory update result: %w", err)
	}
	if affected == 0 {
		return application.ErrAccessoryNotFound
	}
	return nil
}

func isSQLiteConstraint(err error) bool {
	var sqliteErr *sqlite.Error
	return errors.As(err, &sqliteErr) && sqliteErr.Code()&0xff == sqlite3.SQLITE_CONSTRAINT
}

func (r *AccessoryRepository) withTx(ctx context.Context, work func(*sql.Tx) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin accessory transaction: %w", err)
	}
	if err := work(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit accessory transaction: %w", err)
	}
	return nil
}

func writeAccessoryAudit(
	ctx context.Context,
	tx *sql.Tx,
	action, targetType, targetID, actor, createdAt, details string,
) error {
	if _, err := tx.ExecContext(ctx, `
INSERT INTO audit_logs(id, actor_user_id, action, target_type, target_id, created_at, details_json)
VALUES(?, NULLIF(?, ''), ?, ?, ?, ?, ?)`, randomID(), actor, action, targetType, targetID, createdAt, details); err != nil {
		return fmt.Errorf("write accessory audit log: %w", err)
	}
	return nil
}

var _ application.AccessoryRepository = (*AccessoryRepository)(nil)
