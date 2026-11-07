package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type externalMappingWriter interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func writeVehicleExternalMapping(
	ctx context.Context, writer externalMappingWriter, vehicleID string, input VehicleExternalMapInput, now string,
) (*VehicleExternalMap, error) {
	var mapping VehicleExternalMap
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
 COALESCE(last_seen_at,''), created_at, updated_at`, randomID(), vehicleID, input.Provider,
		input.ExternalID, nullableString(input.ExternalName), nullableString(input.ExternalAddress),
		nullableString(input.ExternalProtocol), input.SyncStatus, now, now, now).Scan(
		&mapping.ID, &mapping.VehicleID, &mapping.Provider, &mapping.ExternalID, &mapping.ExternalName,
		&mapping.ExternalAddress, &mapping.ExternalProtocol, &mapping.SyncStatus, &mapping.LastSeenAt,
		&mapping.CreatedAt, &mapping.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrVehicleExternalMappingConflict
	}
	if err != nil {
		return nil, fmt.Errorf("write vehicle external mapping: %w", err)
	}
	return &mapping, nil
}
