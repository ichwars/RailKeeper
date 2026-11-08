package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"strings"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/domain"
	"railkeeper/backend/internal/transfer"
	transferpolicy "railkeeper/backend/internal/transfer/policy"
)

func applyTransferAccessory(
	ctx context.Context,
	db *sql.Tx,
	record transfer.DataTransferPreviewRecord,
	action string,
) error {
	accessory := transfer.TransferAccessory{}
	if err := json.Unmarshal(record.Data, &accessory); err != nil {
		return dataTransferApplyConflict("accessory preview data is invalid")
	}
	if err := application.ValidateTransferAccessory(accessory); err != nil {
		return dataTransferApplyConflict("accessory preview data violates aggregate validation")
	}
	if accessory.ManufacturerStatus == "" {
		accessory.ManufacturerStatus = "unknown"
	}
	if accessory.Subtype == "" && accessory.ArticleType == string(domain.AccessoryArticleOther) {
		accessory.Subtype = "other:other"
	}
	referenceTargetID := record.TargetID
	if action == "copy" {
		referenceTargetID = ""
	}
	if err := validateTransferAccessoryReferenceData(ctx, db, accessory, referenceTargetID); err != nil {
		if errors.Is(err, application.ErrAccessoryValidation) {
			return dataTransferApplyConflict("accessory article type or subtype is not active")
		}
		return err
	}
	if action == "use_existing" {
		return nil
	}
	copying := action == "copy"
	if action == "copy" {
		inventoryNumber, err := uniqueTransferInventoryNumber(ctx, db, "accessory_products", accessory.InventoryNumber)
		if err != nil {
			return err
		}
		accessory.InventoryNumber = inventoryNumber
		action = "create"
	}
	now := timestamp()
	productID := record.TargetID
	if action == "create" {
		productID = randomID()
	} else if action != "replace" || productID == "" {
		return dataTransferApplyConflict("unsupported accessory resolution")
	}
	targetStrategy := domain.AccessoryInventoryStrategy(accessory.InventoryStrategy)
	if !targetStrategy.Valid() || domain.AccessoryTrackingMode(accessory.TrackingMode) != targetStrategy.TrackingMode() {
		return dataTransferApplyConflict("imported accessory inventory strategy is invalid")
	}
	if targetStrategy == domain.AccessoryInventoryQuantity && len(accessory.Assets) > 0 {
		return dataTransferApplyConflict("quantity accessory import cannot contain individual assets")
	}
	if action == "replace" {
		currentStrategy, err := accessoryInventoryStrategy(ctx, db, productID)
		if err != nil {
			return err
		}
		if err := validateAccessoryInventoryStrategyTransition(
			ctx, db, productID, currentStrategy, targetStrategy,
		); errors.Is(err, application.ErrAccessoryConflict) {
			return dataTransferApplyConflict("imported accessory inventory strategy conflicts with local allocations")
		} else if err != nil {
			return err
		}
	}
	stock, err := prepareTransferAccessoryStock(ctx, db, productID, targetStrategy, accessory.Stock, now, action == "replace")
	if err != nil {
		return err
	}
	gauges, _ := json.Marshal(transferpolicy.NonNilStrings(accessory.Gauges))
	alternatives, _ := json.Marshal(transferpolicy.NonNilStrings(accessory.AlternativeNumbers))
	keywords, _ := json.Marshal(transferpolicy.NonNilStrings(accessory.Keywords))
	arguments := []any{
		accessory.InventoryNumber, accessory.Manufacturer, accessory.ArticleNumber, accessory.Name, accessory.Category,
		accessory.TrackingMode, accessory.Description, accessory.EAN, accessory.ManufacturerStatus, accessory.ArticleType,
		accessory.Subtype, string(gauges), accessory.Scale, accessory.ListPrice, accessory.PackageQuantity,
		accessory.StockUnit, accessory.MinimumStock, accessory.InventoryStrategy, accessory.ManufacturerURL,
		accessory.ProductURL, string(alternatives), string(keywords), accessory.CompatibilityNotes,
		accessory.InternalNotes, accessory.Archived,
	}
	if action == "create" {
		_, err := db.ExecContext(ctx, `
INSERT INTO accessory_products(
  id, inventory_number, manufacturer, article_number, name, category, tracking_mode, description, ean,
  manufacturer_status, article_type, subtype, gauges_json, scale, list_price, package_quantity, stock_unit,
  minimum_stock, inventory_strategy, manufacturer_url, product_url, alternative_numbers_json, keywords_json,
  compatibility_notes, internal_notes, archived, created_at, updated_at
) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			append(append([]any{productID}, arguments...), now, now)...)
		if err != nil {
			return err
		}
	} else {
		result, err := db.ExecContext(ctx, `
UPDATE accessory_products SET inventory_number=?, manufacturer=?, article_number=?, name=?, category=?,
  tracking_mode=?, description=?, ean=?, manufacturer_status=?, article_type=?, subtype=?, gauges_json=?, scale=?,
  list_price=?, package_quantity=?, stock_unit=?, minimum_stock=?, inventory_strategy=?, manufacturer_url=?,
  product_url=?, alternative_numbers_json=?, keywords_json=?, compatibility_notes=?, internal_notes=?, archived=?,
  updated_at=? WHERE id=?`, append(arguments, now, productID)...)
		if err != nil {
			return err
		}
		if err := requireApplyUpdate(result, "replace accessory"); err != nil {
			return err
		}
	}
	for _, level := range stock {
		if _, err := db.ExecContext(ctx, `
INSERT INTO accessory_stock(product_id, location_id, quantity, updated_at) VALUES(?, ?, ?, ?)
ON CONFLICT(product_id, location_id) DO UPDATE SET quantity=excluded.quantity, updated_at=excluded.updated_at`,
			productID, level.LocationID, level.Quantity, now); err != nil {
			return err
		}
	}
	for _, asset := range accessory.Assets {
		if copying && asset.InventoryNumber != "" {
			inventoryNumber, err := uniqueTransferInventoryNumber(
				ctx, db, "accessory_assets", asset.InventoryNumber,
			)
			if err != nil {
				return err
			}
			asset.InventoryNumber = inventoryNumber
		}
		locationID := ""
		var err error
		if asset.StorageLocationID != "" || asset.StorageLocationName != "" {
			locationID, err = transferStorageLocation(ctx, db, asset.StorageLocationID, asset.StorageLocationName, now)
			if err != nil {
				return err
			}
		}
		if action == "replace" {
			matchedID, err := matchingTransferAccessoryAsset(ctx, db, productID, asset)
			if err != nil {
				return err
			}
			if matchedID != "" {
				if err := updateTransferAccessoryAsset(ctx, db, matchedID, asset, locationID, now); err != nil {
					return err
				}
				continue
			}
		}
		if asset.Lifecycle == "reserved" || asset.Lifecycle == "installed" {
			return dataTransferApplyConflict("imported accessory asset has no matching local relationship")
		}
		if _, err := db.ExecContext(ctx, `
INSERT INTO accessory_assets(
  id, product_id, inventory_number, serial_number, condition_state, lifecycle_state, storage_location_id,
  purchase_date, purchase_price, warranty_until, notes, created_at, updated_at
) VALUES(?, ?, NULLIF(?, ''), ?, ?, ?, NULLIF(?, ''), ?, ?, ?, ?, ?, ?)`, randomID(), productID,
			asset.InventoryNumber, asset.SerialNumber, asset.Condition, asset.Lifecycle, locationID, asset.PurchaseDate,
			asset.PurchasePrice, asset.WarrantyUntil, asset.Notes, now, now); err != nil {
			return err
		}
	}
	return nil
}

type preparedTransferAccessoryStock struct {
	LocationID string
	Quantity   int
}

func prepareTransferAccessoryStock(
	ctx context.Context,
	db *sql.Tx,
	productID string,
	strategy domain.AccessoryInventoryStrategy,
	stock []transfer.TransferAccessoryStock,
	now string,
	validateReservations bool,
) ([]preparedTransferAccessoryStock, error) {
	prepared := make([]preparedTransferAccessoryStock, 0, len(stock))
	seenLocations := make(map[string]bool, len(stock))
	for _, level := range stock {
		if level.Quantity < 0 || strategy == domain.AccessoryInventoryIndividual && level.Quantity != 0 {
			return nil, dataTransferApplyConflict("imported accessory stock is incompatible with its inventory strategy")
		}
		locationID, err := transferStorageLocation(ctx, db, level.LocationID, level.LocationName, now)
		if err != nil {
			return nil, err
		}
		if seenLocations[locationID] {
			return nil, dataTransferApplyConflict("imported accessory stock contains a duplicate location")
		}
		seenLocations[locationID] = true
		if validateReservations {
			var reserved int
			if err := db.QueryRowContext(ctx, `
SELECT COALESCE(SUM(quantity), 0)
FROM accessory_reservations
WHERE product_id=? AND location_id=? AND asset_id IS NULL AND status='active'`,
				productID, locationID).Scan(&reserved); err != nil {
				return nil, fmt.Errorf("read reserved accessory stock before import: %w", err)
			}
			if level.Quantity < reserved {
				return nil, dataTransferApplyConflict("imported accessory stock is below its active reserved quantity")
			}
		}
		prepared = append(prepared, preparedTransferAccessoryStock{LocationID: locationID, Quantity: level.Quantity})
	}
	return prepared, nil
}

func matchingTransferAccessoryAsset(
	ctx context.Context,
	db dataTransferApplyDB,
	productID string,
	asset transfer.TransferAccessoryAsset,
) (string, error) {
	var id string
	if asset.ID != "" {
		err := db.QueryRowContext(ctx, `SELECT id FROM accessory_assets WHERE id=? AND product_id=?`,
			asset.ID, productID).Scan(&id)
		if err == nil {
			return id, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("match imported accessory asset by id: %w", err)
		}
	}
	if asset.InventoryNumber == "" {
		return "", nil
	}
	err := db.QueryRowContext(ctx, `
SELECT id FROM accessory_assets WHERE product_id=? AND inventory_number=? COLLATE NOCASE`,
		productID, asset.InventoryNumber).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("match imported accessory asset by inventory number: %w", err)
	}
	return id, nil
}

func updateTransferAccessoryAsset(
	ctx context.Context,
	db dataTransferApplyDB,
	assetID string,
	asset transfer.TransferAccessoryAsset,
	locationID string,
	now string,
) error {
	var activeReservations, activeInstallations int
	var reservationLocation string
	if err := db.QueryRowContext(ctx, `
SELECT
  (SELECT COUNT(*) FROM accessory_reservations WHERE asset_id=? AND status='active'),
	  (SELECT COUNT(*) FROM accessory_installations WHERE asset_id=? AND removed_at IS NULL),
  COALESCE((SELECT MAX(location_id) FROM accessory_reservations WHERE asset_id=? AND status='active'), '')`,
		assetID, assetID, assetID).Scan(&activeReservations, &activeInstallations, &reservationLocation); err != nil {
		return fmt.Errorf("check imported accessory asset relationships: %w", err)
	}
	if activeReservations > 0 && (asset.Lifecycle != "reserved" || locationID != reservationLocation) ||
		activeInstallations > 0 && (asset.Lifecycle != "installed" || locationID != "") ||
		asset.Lifecycle == "reserved" && activeReservations == 0 ||
		asset.Lifecycle == "installed" && activeInstallations == 0 {
		return dataTransferApplyConflict("imported accessory asset conflicts with a local relationship")
	}
	result, err := db.ExecContext(ctx, `
UPDATE accessory_assets
SET inventory_number=NULLIF(?, ''), serial_number=?, condition_state=?, lifecycle_state=?,
    storage_location_id=NULLIF(?, ''), purchase_date=?, purchase_price=?, warranty_until=?, notes=?, updated_at=?
WHERE id=?`, asset.InventoryNumber, asset.SerialNumber, asset.Condition, asset.Lifecycle, locationID,
		asset.PurchaseDate, asset.PurchasePrice, asset.WarrantyUntil, asset.Notes, now, assetID)
	if err != nil {
		return err
	}
	return requireApplyUpdate(result, "update imported accessory asset")
}

func transferStorageLocation(
	ctx context.Context,
	db dataTransferApplyDB,
	sourceID string,
	name string,
	now string,
) (string, error) {
	var id string
	if sourceID != "" {
		err := db.QueryRowContext(ctx, `SELECT id FROM storage_locations WHERE id=?`, sourceID).Scan(&id)
		if err == nil {
			return id, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return "", err
		}
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return "", dataTransferApplyConflict("accessory storage location is missing")
	}
	err := db.QueryRowContext(ctx, `
SELECT id FROM storage_locations WHERE parent_id IS NULL AND name=? COLLATE NOCASE`, name).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	id = randomID()
	if _, err := db.ExecContext(ctx, `
INSERT INTO storage_locations(id, name, created_at, updated_at) VALUES(?, ?, ?, ?)`, id, name, now, now); err != nil {
		return "", err
	}
	return id, nil
}
