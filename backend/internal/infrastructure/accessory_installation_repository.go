package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/domain"
)

func (r *AccessoryRepository) ListInstallations(
	ctx context.Context,
	productID string,
) ([]application.AccessoryInstallation, error) {
	rows, err := r.db.QueryContext(ctx, accessoryInstallationSelect+`
WHERE ?='' OR product_id=?
ORDER BY installed_at DESC, id`, productID, productID)
	if err != nil {
		return nil, fmt.Errorf("list accessory installations: %w", err)
	}
	defer func() { _ = rows.Close() }()
	installations := []application.AccessoryInstallation{}
	for rows.Next() {
		installation, err := scanAccessoryInstallation(rows)
		if err != nil {
			return nil, fmt.Errorf("scan accessory installation: %w", err)
		}
		installations = append(installations, *installation)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate accessory installations: %w", err)
	}
	return installations, nil
}

func (r *AccessoryRepository) Install(
	ctx context.Context,
	input application.CreateAccessoryInstallationInput,
	actor string,
) (*application.AccessoryInstallation, error) {
	now := timestamp()
	installationID := randomID()
	err := r.withTx(ctx, func(tx *sql.Tx) error {
		strategy, err := accessoryInventoryStrategy(ctx, tx, input.ProductID)
		if err != nil {
			return err
		}
		if err := requireActiveStorageLocation(ctx, tx, input.SourceLocationID); err != nil {
			return err
		}
		if err := requireAllocationTarget(ctx, tx, input.AllocationTargetInput); err != nil {
			return err
		}
		var reservation *application.AccessoryReservation
		if input.ReservationID != "" {
			reservation, err = getReservationWith(ctx, tx, input.ReservationID)
			if err != nil {
				return err
			}
			if !reservationMatchesInstallation(reservation, input) {
				return application.ErrAccessoryConflict
			}
			inheritReservationTechnicalData(&input, reservation)
		}
		usesAsset, err := accessoryAllocationUsesAsset(strategy, input.AssetID)
		if err != nil {
			return err
		}
		if usesAsset {
			if err := installAccessoryAsset(ctx, tx, input, reservation != nil, now); err != nil {
				return err
			}
		} else {
			if reservation == nil {
				available, err := availableAccessoryQuantity(ctx, tx, input.ProductID, input.SourceLocationID)
				if err != nil {
					return err
				}
				if available < input.Quantity {
					return application.ErrAccessoryInsufficientStock
				}
			}
			result, err := tx.ExecContext(ctx, `
UPDATE accessory_stock SET quantity=quantity-?, updated_at=?
WHERE product_id=? AND location_id=? AND quantity>=?`, input.Quantity, now, input.ProductID,
				input.SourceLocationID, input.Quantity)
			if err != nil {
				return fmt.Errorf("decrement installed accessory stock: %w", err)
			}
			if affected, err := result.RowsAffected(); err != nil {
				return fmt.Errorf("read installed stock result: %w", err)
			} else if affected == 0 {
				return application.ErrAccessoryInsufficientStock
			}
		}
		if reservation != nil {
			result, err := tx.ExecContext(ctx, `
UPDATE accessory_reservations SET status=?, updated_at=? WHERE id=? AND status=?`,
				domain.AccessoryReservationFulfilled, now, reservation.ID, domain.AccessoryReservationActive)
			if err != nil {
				return fmt.Errorf("fulfill accessory reservation: %w", err)
			}
			if err := requireAccessoryConflictFreeUpdate(result); err != nil {
				return err
			}
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO accessory_installations(
  id, product_id, asset_id, source_location_id, quantity, vehicle_id, layout_id, layout_unit_id,
  condition_state, installed_by, installed_at, notes, placement, digital_address, decoder_output,
  connection, wiring_notes
) VALUES(?, ?, NULLIF(?, ''), ?, ?, NULLIF(?, ''), NULLIF(?, ''), NULLIF(?, ''), ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			installationID, input.ProductID, input.AssetID, input.SourceLocationID, input.Quantity,
			input.VehicleID, input.LayoutID, input.LayoutUnitID, input.Condition, actor, now, input.Notes,
			input.Placement, input.DigitalAddress, input.DecoderOutput, input.Connection, input.WiringNotes); err != nil {
			if isSQLiteConstraint(err) {
				return application.ErrAccessoryConflict
			}
			return fmt.Errorf("insert accessory installation: %w", err)
		}
		if !usesAsset {
			if err := insertAccessoryStockMovement(ctx, tx, input.ProductID, input.SourceLocationID,
				"installation", -input.Quantity, "installation", installationID, actor, "", now); err != nil {
				return err
			}
		}
		details, err := allocationAuditDetails(input.Quantity, input.AllocationTargetInput, "")
		if err != nil {
			return err
		}
		return writeAccessoryAudit(ctx, tx, "AccessoryInstalled", "accessory_installation",
			installationID, actor, now, details)
	})
	if err != nil {
		return nil, err
	}
	return r.getInstallation(ctx, installationID)
}

func (r *AccessoryRepository) RemoveInstallation(
	ctx context.Context,
	id string,
	input application.RemoveAccessoryInstallationInput,
	actor string,
) (*application.AccessoryInstallation, error) {
	now := timestamp()
	err := r.withTx(ctx, func(tx *sql.Tx) error {
		installation, err := getInstallationWith(ctx, tx, id)
		if err != nil {
			return err
		}
		if installation.RemovedAt != "" {
			return application.ErrAccessoryConflict
		}
		strategy, err := accessoryInventoryStrategy(ctx, tx, installation.ProductID)
		if err != nil {
			return err
		}
		usesAsset, err := accessoryAllocationUsesAsset(strategy, installation.AssetID)
		if err != nil {
			return err
		}
		if input.Disposition == domain.AccessoryRemovalStored {
			if err := requireActiveStorageLocation(ctx, tx, input.StorageLocationID); err != nil {
				return err
			}
		}
		if !usesAsset {
			if input.Disposition == domain.AccessoryRemovalStored {
				if _, err := tx.ExecContext(ctx, `
INSERT INTO accessory_stock(product_id, location_id, quantity, updated_at)
VALUES(?, ?, ?, ?)
ON CONFLICT(product_id, location_id) DO UPDATE
SET quantity=quantity+excluded.quantity, updated_at=excluded.updated_at`, installation.ProductID,
					input.StorageLocationID, installation.Quantity, now); err != nil {
					return fmt.Errorf("restore removed accessory stock: %w", err)
				}
			}
		} else if err := removeAccessoryAsset(ctx, tx, installation, input, now); err != nil {
			return err
		}
		result, err := tx.ExecContext(ctx, `
UPDATE accessory_installations
SET removed_by=?, removed_at=?, removal_disposition=?, removal_notes=?
WHERE id=? AND removed_at IS NULL`, actor, now, input.Disposition, input.Notes, id)
		if err != nil {
			return fmt.Errorf("close accessory installation: %w", err)
		}
		if err := requireAccessoryConflictFreeUpdate(result); err != nil {
			return err
		}
		if !usesAsset && input.Disposition == domain.AccessoryRemovalStored {
			if err := insertAccessoryStockMovement(ctx, tx, installation.ProductID, input.StorageLocationID,
				"removal", installation.Quantity, "installation", id, actor, "", now); err != nil {
				return err
			}
		}
		details, err := allocationAuditDetails(installation.Quantity,
			installation.AllocationTargetInput, input.Disposition)
		if err != nil {
			return err
		}
		return writeAccessoryAudit(ctx, tx, "AccessoryRemoved", "accessory_installation",
			id, actor, now, details)
	})
	if err != nil {
		return nil, err
	}
	return r.getInstallation(ctx, id)
}

func (r *AccessoryRepository) UpdateInstallationCondition(
	ctx context.Context,
	id string,
	input application.UpdateAccessoryInstallationConditionInput,
	actor string,
) (*application.AccessoryInstallation, error) {
	now := timestamp()
	err := r.withTx(ctx, func(tx *sql.Tx) error {
		installation, err := getInstallationWith(ctx, tx, id)
		if err != nil {
			return err
		}
		if installation.RemovedAt != "" {
			return application.ErrAccessoryConflict
		}
		result, err := tx.ExecContext(ctx, `
UPDATE accessory_installations SET condition_state=? WHERE id=? AND removed_at IS NULL`, input.Condition, id)
		if err != nil {
			return fmt.Errorf("update installation condition: %w", err)
		}
		if err := requireAccessoryConflictFreeUpdate(result); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO accessory_installation_condition_history(
  id, installation_id, previous_condition, condition_state, changed_by, changed_at
) VALUES(?, ?, ?, ?, ?, ?)`, randomID(), id, installation.Condition, input.Condition, actor, now); err != nil {
			return fmt.Errorf("record installation condition history: %w", err)
		}
		if installation.AssetID != "" {
			result, err = tx.ExecContext(ctx, `
UPDATE accessory_assets SET condition_state=?, updated_at=? WHERE id=?`, input.Condition, now,
				installation.AssetID)
			if err != nil {
				return fmt.Errorf("update installed asset condition: %w", err)
			}
			if err := requireAccessoryUpdated(result); err != nil {
				return err
			}
		}
		return writeAccessoryAudit(ctx, tx, "AccessoryInstallationConditionUpdated",
			"accessory_installation", id, actor, now, "{}")
	})
	if err != nil {
		return nil, err
	}
	return r.getInstallation(ctx, id)
}

func installAccessoryAsset(
	ctx context.Context,
	tx *sql.Tx,
	input application.CreateAccessoryInstallationInput,
	reserved bool,
	now string,
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
		return fmt.Errorf("read installable accessory asset: %w", err)
	}
	wantLifecycle := domain.AccessoryLifecycleStored
	if reserved {
		wantLifecycle = domain.AccessoryLifecycleReserved
	}
	if productID != input.ProductID || locationID != input.SourceLocationID || lifecycle != wantLifecycle {
		return application.ErrAccessoryConflict
	}
	if !reserved {
		var active int
		if err := tx.QueryRowContext(ctx, `
SELECT COUNT(*) FROM accessory_reservations WHERE asset_id=? AND status='active'`, input.AssetID).Scan(&active); err != nil {
			return fmt.Errorf("check reserved asset before installation: %w", err)
		}
		if active > 0 {
			return application.ErrAccessoryConflict
		}
	}
	result, err := tx.ExecContext(ctx, `
UPDATE accessory_assets
SET storage_location_id=NULL, lifecycle_state=?, condition_state=?, updated_at=?
WHERE id=? AND lifecycle_state=?`, domain.AccessoryLifecycleInstalled, input.Condition, now,
		input.AssetID, wantLifecycle)
	if err != nil {
		return fmt.Errorf("mark accessory asset installed: %w", err)
	}
	return requireAccessoryConflictFreeUpdate(result)
}

func removeAccessoryAsset(
	ctx context.Context,
	tx *sql.Tx,
	installation *application.AccessoryInstallation,
	input application.RemoveAccessoryInstallationInput,
	now string,
) error {
	var lifecycle domain.AccessoryLifecycle
	condition := installation.Condition
	locationID := ""
	switch input.Disposition {
	case domain.AccessoryRemovalStored:
		lifecycle = domain.AccessoryLifecycleStored
		locationID = input.StorageLocationID
	case domain.AccessoryRemovalMaintenance:
		lifecycle = domain.AccessoryLifecycleMaintenance
	case domain.AccessoryRemovalDefective:
		lifecycle = domain.AccessoryLifecycleMaintenance
		condition = domain.AccessoryConditionDefective
	case domain.AccessoryRemovalRetired:
		lifecycle = domain.AccessoryLifecycleRetired
	default:
		return application.ErrAccessoryValidation
	}
	result, err := tx.ExecContext(ctx, `
UPDATE accessory_assets
SET storage_location_id=NULLIF(?, ''), lifecycle_state=?, condition_state=?, updated_at=?
WHERE id=? AND lifecycle_state=?`, locationID, lifecycle, condition, now, installation.AssetID,
		domain.AccessoryLifecycleInstalled)
	if err != nil {
		return fmt.Errorf("update removed accessory asset: %w", err)
	}
	return requireAccessoryConflictFreeUpdate(result)
}

func reservationMatchesInstallation(
	reservation *application.AccessoryReservation,
	input application.CreateAccessoryInstallationInput,
) bool {
	return reservation.Status == domain.AccessoryReservationActive &&
		reservation.ProductID == input.ProductID && reservation.AssetID == input.AssetID &&
		reservation.LocationID == input.SourceLocationID && reservation.Quantity == input.Quantity &&
		reservation.VehicleID == input.VehicleID && reservation.LayoutID == input.LayoutID &&
		reservation.LayoutUnitID == input.LayoutUnitID
}

func inheritReservationTechnicalData(
	input *application.CreateAccessoryInstallationInput,
	reservation *application.AccessoryReservation,
) {
	if input.Placement == "" {
		input.Placement = reservation.Placement
	}
	if input.DigitalAddress == "" {
		input.DigitalAddress = reservation.DigitalAddress
	}
	if input.DecoderOutput == "" {
		input.DecoderOutput = reservation.DecoderOutput
	}
	if input.Connection == "" {
		input.Connection = reservation.Connection
	}
	if input.WiringNotes == "" {
		input.WiringNotes = reservation.WiringNotes
	}
}

const accessoryInstallationSelect = `SELECT id, product_id, COALESCE(asset_id, ''), source_location_id, quantity,
COALESCE(vehicle_id, ''), COALESCE(layout_id, ''), COALESCE(layout_unit_id, ''), condition_state,
installed_by, installed_at, COALESCE(removed_by, ''), COALESCE(removed_at, ''),
COALESCE(removal_disposition, ''), notes, removal_notes, placement, digital_address, decoder_output,
connection, wiring_notes FROM accessory_installations`

func scanAccessoryInstallation(scanner rowScanner) (*application.AccessoryInstallation, error) {
	installation := &application.AccessoryInstallation{}
	err := scanner.Scan(&installation.ID, &installation.ProductID, &installation.AssetID,
		&installation.SourceLocationID, &installation.Quantity, &installation.VehicleID,
		&installation.LayoutID, &installation.LayoutUnitID, &installation.Condition,
		&installation.InstalledBy, &installation.InstalledAt, &installation.RemovedBy,
		&installation.RemovedAt, &installation.RemovalDisposition, &installation.Notes,
		&installation.RemovalNotes, &installation.Placement, &installation.DigitalAddress,
		&installation.DecoderOutput, &installation.Connection, &installation.WiringNotes)
	return installation, err
}

func (r *AccessoryRepository) getInstallation(
	ctx context.Context,
	id string,
) (*application.AccessoryInstallation, error) {
	return getInstallationWith(ctx, r.db, id)
}

func getInstallationWith(
	ctx context.Context,
	queryer accessoryQueryer,
	id string,
) (*application.AccessoryInstallation, error) {
	installation, err := scanAccessoryInstallation(queryer.QueryRowContext(
		ctx, accessoryInstallationSelect+` WHERE id=?`, id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, application.ErrAccessoryNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get accessory installation: %w", err)
	}
	return installation, nil
}
