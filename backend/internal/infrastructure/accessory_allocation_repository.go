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

func (r *AccessoryRepository) ListReservations(
	ctx context.Context,
	productID string,
) ([]application.AccessoryReservation, error) {
	rows, err := r.db.QueryContext(ctx, accessoryReservationSelect+`
WHERE ?='' OR product_id=?
ORDER BY created_at DESC, id`, productID, productID)
	if err != nil {
		return nil, fmt.Errorf("list accessory reservations: %w", err)
	}
	defer func() { _ = rows.Close() }()
	reservations := []application.AccessoryReservation{}
	for rows.Next() {
		reservation, err := scanAccessoryReservation(rows)
		if err != nil {
			return nil, fmt.Errorf("scan accessory reservation: %w", err)
		}
		reservations = append(reservations, *reservation)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate accessory reservations: %w", err)
	}
	return reservations, nil
}

func (r *AccessoryRepository) CreateReservation(
	ctx context.Context,
	input application.CreateAccessoryReservationInput,
	actor string,
) (*application.AccessoryReservation, error) {
	now := timestamp()
	reservationID := randomID()
	err := r.withTx(ctx, func(tx *sql.Tx) error {
		strategy, err := accessoryInventoryStrategy(ctx, tx, input.ProductID)
		if err != nil {
			return err
		}
		if err := requireStorageLocation(ctx, tx, input.LocationID); err != nil {
			return err
		}
		if err := requireAllocationTarget(ctx, tx, input.AllocationTargetInput); err != nil {
			return err
		}
		usesAsset, err := accessoryAllocationUsesAsset(strategy, input.AssetID)
		if err != nil {
			return err
		}
		if usesAsset {
			if err := requireReservableAsset(ctx, tx, input); err != nil {
				return err
			}
		} else {
			available, err := availableAccessoryQuantity(ctx, tx, input.ProductID, input.LocationID)
			if err != nil {
				return err
			}
			if available < input.Quantity {
				return application.ErrAccessoryInsufficientStock
			}
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO accessory_reservations(
  id, product_id, asset_id, location_id, quantity, vehicle_id, layout_id, layout_unit_id,
  status, note, created_by, created_at, updated_at
) VALUES(?, ?, NULLIF(?, ''), ?, ?, NULLIF(?, ''), NULLIF(?, ''), NULLIF(?, ''), ?, ?, ?, ?, ?)`,
			reservationID, input.ProductID, input.AssetID, input.LocationID, input.Quantity,
			input.VehicleID, input.LayoutID, input.LayoutUnitID, domain.AccessoryReservationActive,
			input.Note, actor, now, now); err != nil {
			if isSQLiteConstraint(err) {
				return application.ErrAccessoryConflict
			}
			return fmt.Errorf("insert accessory reservation: %w", err)
		}
		if input.AssetID != "" {
			result, err := tx.ExecContext(ctx, `
UPDATE accessory_assets SET lifecycle_state=?, updated_at=?
WHERE id=? AND lifecycle_state=?`, domain.AccessoryLifecycleReserved, now, input.AssetID,
				domain.AccessoryLifecycleStored)
			if err != nil {
				return fmt.Errorf("reserve accessory asset: %w", err)
			}
			if err := requireAccessoryConflictFreeUpdate(result); err != nil {
				return err
			}
		}
		details, err := allocationAuditDetails(input.Quantity, input.AllocationTargetInput, "")
		if err != nil {
			return err
		}
		return writeAccessoryAudit(ctx, tx, "AccessoryReservationCreated", "accessory_reservation",
			reservationID, actor, now, details)
	})
	if err != nil {
		return nil, err
	}
	return r.getReservation(ctx, reservationID)
}

func (r *AccessoryRepository) CancelReservation(
	ctx context.Context,
	id string,
	actor string,
) (*application.AccessoryReservation, error) {
	now := timestamp()
	err := r.withTx(ctx, func(tx *sql.Tx) error {
		reservation, err := getReservationWith(ctx, tx, id)
		if err != nil {
			return err
		}
		if reservation.Status != domain.AccessoryReservationActive {
			return application.ErrAccessoryConflict
		}
		result, err := tx.ExecContext(ctx, `
UPDATE accessory_reservations SET status=?, updated_at=? WHERE id=? AND status=?`,
			domain.AccessoryReservationCancelled, now, id, domain.AccessoryReservationActive)
		if err != nil {
			return fmt.Errorf("cancel accessory reservation: %w", err)
		}
		if err := requireAccessoryConflictFreeUpdate(result); err != nil {
			return err
		}
		if reservation.AssetID != "" {
			result, err = tx.ExecContext(ctx, `
UPDATE accessory_assets SET lifecycle_state=?, updated_at=? WHERE id=? AND lifecycle_state=?`,
				domain.AccessoryLifecycleStored, now, reservation.AssetID, domain.AccessoryLifecycleReserved)
			if err != nil {
				return fmt.Errorf("release reserved accessory asset: %w", err)
			}
			if err := requireAccessoryConflictFreeUpdate(result); err != nil {
				return err
			}
		}
		return writeAccessoryAudit(ctx, tx, "AccessoryReservationCancelled", "accessory_reservation",
			id, actor, now, "{}")
	})
	if err != nil {
		return nil, err
	}
	return r.getReservation(ctx, id)
}

func (r *AccessoryRepository) GetAllocationSummary(
	ctx context.Context,
	productID string,
) (*application.AccessoryAllocationSummary, error) {
	strategy, err := accessoryInventoryStrategy(ctx, r.db, productID)
	if err != nil {
		return nil, err
	}
	summary := &application.AccessoryAllocationSummary{ProductID: productID}
	switch strategy {
	case domain.AccessoryInventoryQuantity:
		if err := r.db.QueryRowContext(ctx, `
SELECT
  COALESCE((SELECT SUM(quantity) FROM accessory_stock WHERE product_id=?), 0),
  COALESCE((SELECT SUM(quantity) FROM accessory_reservations WHERE product_id=? AND status='active'), 0),
  COALESCE((SELECT SUM(quantity) FROM accessory_installations WHERE product_id=? AND removed_at IS NULL), 0)`,
			productID, productID, productID).Scan(&summary.Stored, &summary.Reserved, &summary.Installed); err != nil {
			return nil, fmt.Errorf("summarize quantity allocations: %w", err)
		}
		summary.Owned = summary.Stored + summary.Installed
	case domain.AccessoryInventoryIndividual:
		if err := r.db.QueryRowContext(ctx, `
SELECT
  COALESCE((SELECT COUNT(*) FROM accessory_assets WHERE product_id=?), 0),
  COALESCE((SELECT COUNT(*) FROM accessory_assets
            WHERE product_id=? AND storage_location_id IS NOT NULL
              AND lifecycle_state IN ('stored', 'reserved')), 0),
  COALESCE((SELECT SUM(quantity) FROM accessory_reservations WHERE product_id=? AND status='active'), 0),
  COALESCE((SELECT SUM(quantity) FROM accessory_installations WHERE product_id=? AND removed_at IS NULL), 0)`,
			productID, productID, productID, productID).
			Scan(&summary.Owned, &summary.Stored, &summary.Reserved, &summary.Installed); err != nil {
			return nil, fmt.Errorf("summarize individual allocations: %w", err)
		}
	case domain.AccessoryInventoryQuantityLaterIndividual:
		if err := r.db.QueryRowContext(ctx, `
SELECT
  COALESCE((SELECT SUM(quantity) FROM accessory_stock WHERE product_id=?), 0) +
  COALESCE((SELECT COUNT(*) FROM accessory_assets WHERE product_id=?), 0) +
  COALESCE((SELECT SUM(quantity) FROM accessory_installations
            WHERE product_id=? AND asset_id IS NULL AND removed_at IS NULL), 0),
  COALESCE((SELECT SUM(quantity) FROM accessory_stock WHERE product_id=?), 0) +
  COALESCE((SELECT COUNT(*) FROM accessory_assets
            WHERE product_id=? AND storage_location_id IS NOT NULL
              AND lifecycle_state IN ('stored', 'reserved')), 0),
  COALESCE((SELECT SUM(quantity) FROM accessory_reservations
            WHERE product_id=? AND status='active'), 0),
  COALESCE((SELECT SUM(quantity) FROM accessory_installations
            WHERE product_id=? AND removed_at IS NULL), 0)`,
			productID, productID, productID, productID, productID, productID, productID).
			Scan(&summary.Owned, &summary.Stored, &summary.Reserved, &summary.Installed); err != nil {
			return nil, fmt.Errorf("summarize hybrid allocations: %w", err)
		}
	default:
		return nil, application.ErrAccessoryTrackingMode
	}
	summary.Available = summary.Stored - summary.Reserved
	if summary.Available < 0 {
		summary.Missing = -summary.Available
		summary.Available = 0
	}
	return summary, nil
}

func accessoryAllocationUsesAsset(
	strategy domain.AccessoryInventoryStrategy,
	assetID string,
) (bool, error) {
	switch strategy {
	case domain.AccessoryInventoryQuantity:
		if assetID != "" {
			return false, application.ErrAccessoryTrackingMode
		}
		return false, nil
	case domain.AccessoryInventoryIndividual:
		if assetID == "" {
			return false, application.ErrAccessoryTrackingMode
		}
		return true, nil
	case domain.AccessoryInventoryQuantityLaterIndividual:
		return assetID != "", nil
	default:
		return false, application.ErrAccessoryTrackingMode
	}
}

func requireReservableAsset(
	ctx context.Context,
	tx *sql.Tx,
	input application.CreateAccessoryReservationInput,
) error {
	var productID, locationID string
	var lifecycle domain.AccessoryLifecycle
	err := tx.QueryRowContext(ctx, `
SELECT product_id, COALESCE(storage_location_id, ''), lifecycle_state
FROM accessory_assets WHERE id=?`, input.AssetID).Scan(&productID, &locationID, &lifecycle)
	if errors.Is(err, sql.ErrNoRows) {
		return application.ErrAccessoryNotFound
	}
	if err != nil {
		return fmt.Errorf("read reservable accessory asset: %w", err)
	}
	if productID != input.ProductID || locationID != input.LocationID || lifecycle != domain.AccessoryLifecycleStored {
		return application.ErrAccessoryConflict
	}
	var active int
	if err := tx.QueryRowContext(ctx, `
SELECT COUNT(*) FROM accessory_reservations WHERE asset_id=? AND status='active'`, input.AssetID).Scan(&active); err != nil {
		return fmt.Errorf("check active asset reservation: %w", err)
	}
	if active > 0 {
		return application.ErrAccessoryConflict
	}
	return nil
}

func availableAccessoryQuantity(ctx context.Context, tx *sql.Tx, productID, locationID string) (int, error) {
	var available int
	if err := tx.QueryRowContext(ctx, `
SELECT
  COALESCE((SELECT quantity FROM accessory_stock WHERE product_id=? AND location_id=?), 0) -
  COALESCE((SELECT SUM(quantity) FROM accessory_reservations
            WHERE product_id=? AND location_id=? AND status='active'), 0)`,
		productID, locationID, productID, locationID).Scan(&available); err != nil {
		return 0, fmt.Errorf("calculate available accessory quantity: %w", err)
	}
	return available, nil
}

func requireAllocationTarget(
	ctx context.Context,
	tx *sql.Tx,
	target application.AllocationTargetInput,
) error {
	var query, id string
	switch {
	case target.VehicleID != "" && target.LayoutID == "" && target.LayoutUnitID == "":
		query, id = `SELECT COUNT(*) FROM vehicles WHERE id=?`, target.VehicleID
	case target.VehicleID == "" && target.LayoutID != "" && target.LayoutUnitID == "":
		query, id = `SELECT COUNT(*) FROM layouts WHERE id=?`, target.LayoutID
	case target.VehicleID == "" && target.LayoutID == "" && target.LayoutUnitID != "":
		query, id = `SELECT COUNT(*) FROM layout_units WHERE id=?`, target.LayoutUnitID
	default:
		return application.ErrAccessoryValidation
	}
	var count int
	if err := tx.QueryRowContext(ctx, query, id).Scan(&count); err != nil {
		return fmt.Errorf("check allocation target: %w", err)
	}
	if count == 0 {
		return application.ErrAccessoryNotFound
	}
	return nil
}

const accessoryReservationSelect = `SELECT id, product_id, COALESCE(asset_id, ''), location_id, quantity,
COALESCE(vehicle_id, ''), COALESCE(layout_id, ''), COALESCE(layout_unit_id, ''),
status, note, created_by, created_at, updated_at FROM accessory_reservations`

func scanAccessoryReservation(scanner rowScanner) (*application.AccessoryReservation, error) {
	reservation := &application.AccessoryReservation{}
	err := scanner.Scan(&reservation.ID, &reservation.ProductID, &reservation.AssetID,
		&reservation.LocationID, &reservation.Quantity, &reservation.VehicleID, &reservation.LayoutID,
		&reservation.LayoutUnitID, &reservation.Status, &reservation.Note, &reservation.CreatedBy,
		&reservation.CreatedAt, &reservation.UpdatedAt)
	return reservation, err
}

func (r *AccessoryRepository) getReservation(
	ctx context.Context,
	id string,
) (*application.AccessoryReservation, error) {
	return getReservationWith(ctx, r.db, id)
}

func getReservationWith(
	ctx context.Context,
	queryer accessoryQueryer,
	id string,
) (*application.AccessoryReservation, error) {
	reservation, err := scanAccessoryReservation(queryer.QueryRowContext(
		ctx, accessoryReservationSelect+` WHERE id=?`, id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, application.ErrAccessoryNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get accessory reservation: %w", err)
	}
	return reservation, nil
}

func allocationAuditDetails(
	quantity int,
	target application.AllocationTargetInput,
	disposition domain.AccessoryRemovalDisposition,
) (string, error) {
	details := map[string]any{"quantity": quantity}
	switch {
	case target.VehicleID != "":
		details["targetType"], details["targetId"] = "vehicle", target.VehicleID
	case target.LayoutID != "":
		details["targetType"], details["targetId"] = "layout", target.LayoutID
	case target.LayoutUnitID != "":
		details["targetType"], details["targetId"] = "layout_unit", target.LayoutUnitID
	}
	if disposition != "" {
		details["disposition"] = disposition
	}
	encoded, err := json.Marshal(details)
	if err != nil {
		return "", fmt.Errorf("encode accessory allocation audit details: %w", err)
	}
	return string(encoded), nil
}

func requireAccessoryConflictFreeUpdate(result sql.Result) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read accessory allocation update result: %w", err)
	}
	if affected == 0 {
		return application.ErrAccessoryConflict
	}
	return nil
}

var _ application.AccessoryAllocationRepository = (*AccessoryRepository)(nil)
