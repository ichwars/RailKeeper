package application

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func (s *VehicleService) UpsertExternalMapping(ctx context.Context, vehicleID string, input VehicleExternalMapInput, _ string) (*VehicleExternalMap, error) {
	vehicleID = strings.TrimSpace(vehicleID)
	input = cleanVehicleExternalMapInput(input)
	if vehicleID == "" || input.Provider == "" || input.ExternalID == "" {
		return nil, ErrVehicleValidation
	}
	if _, err := s.get(ctx, vehicleID); err != nil {
		return nil, err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	lastSeenAt := now
	if input.SyncStatus == "" {
		input.SyncStatus = "linked"
	}
	_, err := s.db.ExecContext(ctx, `
INSERT INTO vehicle_external_mappings(id, vehicle_id, provider, external_id, external_name, external_address, external_protocol, sync_status, last_seen_at, created_at, updated_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(provider, external_id) DO UPDATE SET
  vehicle_id=excluded.vehicle_id,
  external_name=excluded.external_name,
  external_address=excluded.external_address,
  external_protocol=excluded.external_protocol,
  sync_status=excluded.sync_status,
  last_seen_at=excluded.last_seen_at,
  updated_at=excluded.updated_at
`, randomID(), vehicleID, input.Provider, input.ExternalID, nullableString(input.ExternalName), nullableString(input.ExternalAddress), nullableString(input.ExternalProtocol), input.SyncStatus, lastSeenAt, now, now)
	if err != nil {
		return nil, fmt.Errorf("upsert vehicle external mapping: %w", err)
	}

	return s.getVehicleExternalMapping(ctx, input.Provider, input.ExternalID)
}

func cleanVehicleExternalMapInput(input VehicleExternalMapInput) VehicleExternalMapInput {
	input.Provider = strings.ToLower(strings.TrimSpace(input.Provider))
	input.ExternalID = strings.TrimSpace(input.ExternalID)
	input.ExternalName = strings.TrimSpace(input.ExternalName)
	input.ExternalAddress = strings.TrimSpace(input.ExternalAddress)
	input.ExternalProtocol = strings.TrimSpace(input.ExternalProtocol)
	input.SyncStatus = strings.TrimSpace(input.SyncStatus)
	if input.SyncStatus == "" {
		input.SyncStatus = "linked"
	}
	return input
}

func (s *VehicleService) getVehicleExternalMapping(ctx context.Context, provider, externalID string) (*VehicleExternalMap, error) {
	var mapping VehicleExternalMap
	if err := s.db.QueryRowContext(ctx, `
SELECT id, vehicle_id, provider, external_id, COALESCE(external_name, ''), COALESCE(external_address, ''), COALESCE(external_protocol, ''), sync_status, COALESCE(last_seen_at, ''), created_at, updated_at
FROM vehicle_external_mappings
WHERE provider=? AND external_id=?
`, provider, externalID).Scan(&mapping.ID, &mapping.VehicleID, &mapping.Provider, &mapping.ExternalID, &mapping.ExternalName, &mapping.ExternalAddress, &mapping.ExternalProtocol, &mapping.SyncStatus, &mapping.LastSeenAt, &mapping.CreatedAt, &mapping.UpdatedAt); err != nil {
		return nil, fmt.Errorf("get vehicle external mapping: %w", err)
	}
	return &mapping, nil
}

func (s *VehicleService) attachExternalMappings(ctx context.Context, vehicles []Vehicle) error {
	for index := range vehicles {
		mappings, err := s.loadVehicleExternalMappings(ctx, vehicles[index].ID)
		if err != nil {
			return err
		}
		vehicles[index].ExternalMappings = mappings
	}
	return nil
}

func (s *VehicleService) loadVehicleExternalMappings(ctx context.Context, vehicleID string) ([]VehicleExternalMap, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, vehicle_id, provider, external_id, COALESCE(external_name, ''), COALESCE(external_address, ''), COALESCE(external_protocol, ''), sync_status, COALESCE(last_seen_at, ''), created_at, updated_at
FROM vehicle_external_mappings
WHERE vehicle_id=?
ORDER BY provider ASC, external_id ASC
`, vehicleID)
	if err != nil {
		return nil, fmt.Errorf("list vehicle external mappings: %w", err)
	}
	defer func() { _ = rows.Close() }()

	mappings := []VehicleExternalMap{}
	for rows.Next() {
		var mapping VehicleExternalMap
		if err := rows.Scan(&mapping.ID, &mapping.VehicleID, &mapping.Provider, &mapping.ExternalID, &mapping.ExternalName, &mapping.ExternalAddress, &mapping.ExternalProtocol, &mapping.SyncStatus, &mapping.LastSeenAt, &mapping.CreatedAt, &mapping.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan vehicle external mapping: %w", err)
		}
		mappings = append(mappings, mapping)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate vehicle external mappings: %w", err)
	}
	return mappings, nil
}
