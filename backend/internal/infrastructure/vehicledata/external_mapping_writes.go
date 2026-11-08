package vehicledata

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	vehiclemodel "railkeeper/backend/internal/inventory/vehicle"
)

func (s *externalMappingStore) WriteMapping(ctx context.Context, write vehiclemodel.ExternalMappingWrite) (*vehiclemodel.VehicleExternalMap, error) {
	return WriteExternalMapping(ctx, s.db, write)
}
func (s *externalMappingStore) ConsolidateMappings(ctx context.Context, vehicleID, previousExternalID string, input vehiclemodel.VehicleExternalMapInput, now string) error {
	tx, txErr := s.db.BeginTx(ctx, nil)
	if txErr != nil {
		return fmt.Errorf("consolidate external mappings: begin transaction: %w", txErr)
	}
	rollback := func(err error) error {
		_ = tx.Rollback()
		return err
	}
	deleted, txErr := tx.ExecContext(ctx, `DELETE FROM vehicle_external_mappings
WHERE vehicle_id=? AND provider=? AND external_id=?`, vehicleID, input.Provider, previousExternalID)
	if txErr != nil {
		return rollback(fmt.Errorf("consolidate external mappings: delete previous mapping: %w", txErr))
	}
	updated, txErr := tx.ExecContext(ctx, `UPDATE vehicle_external_mappings
SET external_name=?, external_address=?, external_protocol=?, sync_status=?, last_seen_at=?, updated_at=?
WHERE vehicle_id=? AND provider=? AND external_id=?`, nullableMappingString(input.ExternalName),
		nullableMappingString(input.ExternalAddress), nullableMappingString(input.ExternalProtocol), input.SyncStatus, now, now,
		vehicleID, input.Provider, input.ExternalID)
	if txErr != nil {
		return rollback(fmt.Errorf("consolidate external mappings: update current mapping: %w", txErr))
	}
	deletedCount, deleteCountErr := deleted.RowsAffected()
	updatedCount, updateCountErr := updated.RowsAffected()
	if deleteCountErr != nil || updateCountErr != nil || deletedCount != 1 || updatedCount != 1 {
		return rollback(vehiclemodel.ErrVehicleExternalMappingConflict)
	}
	if txErr := tx.Commit(); txErr != nil {
		return fmt.Errorf("consolidate external mappings: commit: %w", txErr)
	}
	return nil
}
func (s *externalMappingStore) RebindMapping(ctx context.Context, vehicleID, previousExternalID string, input vehiclemodel.VehicleExternalMapInput, now string) error {
	result, err := s.db.ExecContext(ctx, `
UPDATE vehicle_external_mappings
SET external_id=?, external_name=?, external_address=?, external_protocol=?, sync_status=?, last_seen_at=?, updated_at=?
WHERE vehicle_id=? AND provider=? AND external_id=?
`, input.ExternalID, nullableMappingString(input.ExternalName), nullableMappingString(input.ExternalAddress),
		nullableMappingString(input.ExternalProtocol), input.SyncStatus, now, now, vehicleID, input.Provider, previousExternalID)
	if err != nil {
		return fmt.Errorf("rebind vehicle external mapping: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil || affected != 1 {
		return vehiclemodel.ErrVehicleExternalMappingConflict
	}
	return nil
}
func nullableMappingString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}
func WriteExternalMapping(
	ctx context.Context, writer queryRower, write vehiclemodel.ExternalMappingWrite,
) (*vehiclemodel.VehicleExternalMap, error) {
	vehicleID, input, now := write.VehicleID, write.Input, write.At
	var mapping vehiclemodel.VehicleExternalMap
	err := writer.QueryRowContext(ctx, `
INSERT INTO vehicle_external_mappings(id, vehicle_id, provider, external_id, external_name,
 external_address, external_protocol, sync_status, last_seen_at, created_at, updated_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(provider, external_id) DO UPDATE SET
 external_name=excluded.external_name, external_address=excluded.external_address,
 external_protocol=excluded.external_protocol, sync_status=excluded.sync_status,
 last_seen_at=excluded.last_seen_at, updated_at=excluded.updated_at
WHERE vehicle_external_mappings.vehicle_id=excluded.vehicle_id
RETURNING id, vehicle_id, provider, external_id, COALESCE(external_name,''),
 COALESCE(external_address,''), COALESCE(external_protocol,''), sync_status,
 COALESCE(last_seen_at,''), created_at, updated_at`, write.ID, vehicleID, input.Provider,
		input.ExternalID, nullableMappingString(input.ExternalName), nullableMappingString(input.ExternalAddress),
		nullableMappingString(input.ExternalProtocol), input.SyncStatus, now, now, now).Scan(
		&mapping.ID, &mapping.VehicleID, &mapping.Provider, &mapping.ExternalID, &mapping.ExternalName,
		&mapping.ExternalAddress, &mapping.ExternalProtocol, &mapping.SyncStatus, &mapping.LastSeenAt,
		&mapping.CreatedAt, &mapping.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, vehiclemodel.ErrVehicleExternalMappingConflict
	}
	if err != nil {
		return nil, fmt.Errorf("write vehicle external mapping: %w", err)
	}
	return &mapping, nil
}
