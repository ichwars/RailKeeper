package infrastructure

import (
	"context"
	"database/sql"
	"fmt"

	"railkeeper/backend/internal/transfer"
)

func loadTransferAccessoryFingerprints(
	ctx context.Context,
	tx *sql.Tx,
	accessories map[string]*transfer.TransferAccessory,
) error {
	rows, err := tx.QueryContext(ctx, `
SELECT r.product_id, r.id, COALESCE(r.asset_id, ''), r.location_id, r.quantity, r.status, r.updated_at
FROM accessory_reservations r JOIN accessory_products p ON p.id=r.product_id
ORDER BY r.id`)
	if err != nil {
		return fmt.Errorf("query transfer accessory reservation fingerprints: %w", err)
	}
	for rows.Next() {
		var productID string
		reservation := transfer.TransferAccessoryReservationFingerprint{}
		if err := rows.Scan(
			&productID, &reservation.ID, &reservation.AssetID, &reservation.LocationID, &reservation.Quantity,
			&reservation.Status, &reservation.UpdatedAt,
		); err != nil {
			_ = rows.Close()
			return fmt.Errorf("scan transfer accessory reservation fingerprint: %w", err)
		}
		state := &accessories[productID].FingerprintState
		state.Reservations = append(state.Reservations, reservation)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return fmt.Errorf("iterate transfer accessory reservation fingerprints: %w", err)
	}
	if err := rows.Close(); err != nil {
		return fmt.Errorf("close transfer accessory reservation fingerprints: %w", err)
	}

	rows, err = tx.QueryContext(ctx, `
SELECT i.product_id, i.id, COALESCE(i.asset_id, ''), i.source_location_id, i.quantity, i.condition_state,
       i.installed_at, COALESCE(i.removed_at, '')
FROM accessory_installations i JOIN accessory_products p ON p.id=i.product_id
ORDER BY i.id`)
	if err != nil {
		return fmt.Errorf("query transfer accessory installation fingerprints: %w", err)
	}
	for rows.Next() {
		var productID string
		installation := transfer.TransferAccessoryInstallationFingerprint{}
		if err := rows.Scan(
			&productID, &installation.ID, &installation.AssetID, &installation.SourceLocationID, &installation.Quantity,
			&installation.Condition, &installation.InstalledAt, &installation.RemovedAt,
		); err != nil {
			_ = rows.Close()
			return fmt.Errorf("scan transfer accessory installation fingerprint: %w", err)
		}
		state := &accessories[productID].FingerprintState
		state.Installations = append(state.Installations, installation)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return fmt.Errorf("iterate transfer accessory installation fingerprints: %w", err)
	}
	if err := rows.Close(); err != nil {
		return fmt.Errorf("close transfer accessory installation fingerprints: %w", err)
	}

	rows, err = tx.QueryContext(ctx, `
SELECT installation.product_id, history.id, history.installation_id, history.condition_state, history.changed_at
FROM accessory_installation_condition_history history
JOIN accessory_installations installation ON installation.id=history.installation_id
JOIN accessory_products product ON product.id=installation.product_id
ORDER BY history.id`)
	if err != nil {
		return fmt.Errorf("query transfer accessory condition history fingerprints: %w", err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var productID string
		history := transfer.TransferAccessoryConditionHistoryFingerprint{}
		if err := rows.Scan(&productID, &history.ID, &history.InstallationID, &history.Condition, &history.ChangedAt); err != nil {
			return fmt.Errorf("scan transfer accessory condition history fingerprint: %w", err)
		}
		state := &accessories[productID].FingerprintState
		state.ConditionHistory = append(state.ConditionHistory, history)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate transfer accessory condition history fingerprints: %w", err)
	}
	return nil
}
