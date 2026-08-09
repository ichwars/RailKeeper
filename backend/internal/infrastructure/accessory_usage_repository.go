package infrastructure

import (
	"context"
	"fmt"

	"railkeeper/backend/internal/application"
)

func (r *AccessoryRepository) GetUsageHistory(
	ctx context.Context,
	productID string,
) (*application.AccessoryUsageHistory, error) {
	var exists int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM accessory_products WHERE id=?`, productID).
		Scan(&exists); err != nil {
		return nil, fmt.Errorf("check accessory product usage history: %w", err)
	}
	if exists == 0 {
		return nil, application.ErrAccessoryNotFound
	}
	rows, err := r.db.QueryContext(ctx, accessoryUsageHistoryQuery, productID, productID, productID, productID)
	if err != nil {
		return nil, fmt.Errorf("list accessory usage history: %w", err)
	}
	defer func() { _ = rows.Close() }()
	history := &application.AccessoryUsageHistory{
		ProductID: productID,
		Events:    []application.AccessoryUsageEvent{},
	}
	for rows.Next() {
		event := application.AccessoryUsageEvent{}
		if err := rows.Scan(&event.ID, &event.Type, &event.ProductID, &event.ReservationID,
			&event.InstallationID, &event.AssetID, &event.LocationID, &event.Quantity,
			&event.VehicleID, &event.LayoutID, &event.LayoutUnitID, &event.Placement,
			&event.DigitalAddress, &event.DecoderOutput, &event.Connection, &event.WiringNotes,
			&event.Status, &event.PreviousCondition, &event.Condition, &event.RemovalDisposition,
			&event.Actor, &event.OccurredAt); err != nil {
			return nil, fmt.Errorf("scan accessory usage history: %w", err)
		}
		history.Events = append(history.Events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate accessory usage history: %w", err)
	}
	return history, nil
}

const accessoryUsageHistoryQuery = `
SELECT * FROM (
  SELECT
    'reservation:' || id AS event_id, 'reservation' AS event_type, product_id,
    id AS reservation_id, '' AS installation_id, COALESCE(asset_id, '') AS asset_id,
    location_id, quantity, COALESCE(vehicle_id, ''), COALESCE(layout_id, ''),
    COALESCE(layout_unit_id, ''), placement, digital_address, decoder_output, connection,
    wiring_notes, status, '' AS previous_condition, '' AS condition_state,
    '' AS removal_disposition, created_by AS actor, created_at AS occurred_at
  FROM accessory_reservations WHERE product_id=?

  UNION ALL

  SELECT
    'installation:' || id, 'installation', product_id, '', id, COALESCE(asset_id, ''),
    source_location_id, quantity, COALESCE(vehicle_id, ''), COALESCE(layout_id, ''),
    COALESCE(layout_unit_id, ''), placement, digital_address, decoder_output, connection,
    wiring_notes, '', '', condition_state, '', installed_by, installed_at
  FROM accessory_installations WHERE product_id=?

  UNION ALL

  SELECT
    'condition:' || history.id, 'condition_changed', installation.product_id, '', installation.id,
    COALESCE(installation.asset_id, ''), installation.source_location_id, installation.quantity,
    COALESCE(installation.vehicle_id, ''), COALESCE(installation.layout_id, ''),
    COALESCE(installation.layout_unit_id, ''), installation.placement, installation.digital_address,
    installation.decoder_output, installation.connection, installation.wiring_notes, '',
    history.previous_condition, history.condition_state, '', history.changed_by, history.changed_at
  FROM accessory_installation_condition_history history
  JOIN accessory_installations installation ON installation.id=history.installation_id
  WHERE installation.product_id=?

  UNION ALL

  SELECT
    'removal:' || id, 'removal', product_id, '', id, COALESCE(asset_id, ''),
    source_location_id, quantity, COALESCE(vehicle_id, ''), COALESCE(layout_id, ''),
    COALESCE(layout_unit_id, ''), placement, digital_address, decoder_output, connection,
    wiring_notes, '', '', condition_state, COALESCE(removal_disposition, ''),
    COALESCE(removed_by, ''), removed_at
  FROM accessory_installations WHERE product_id=? AND removed_at IS NOT NULL
)
ORDER BY occurred_at DESC, event_id DESC`
