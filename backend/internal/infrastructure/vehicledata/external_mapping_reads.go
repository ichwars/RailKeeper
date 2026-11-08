package vehicledata

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	vehiclemodel "railkeeper/backend/internal/inventory/vehicle"
)

type externalMappingStore struct{ db *sql.DB }

func NewExternalMappingStore(db *sql.DB) vehiclemodel.ExternalMappingRepository {
	return &externalMappingStore{db: db}
}
func (s *externalMappingStore) FindMapping(ctx context.Context, provider, id string) (*vehiclemodel.VehicleExternalMap, bool, error) {
	mapping, err := s.Mapping(ctx, provider, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	}
	return mapping, err == nil, err
}
func (s *externalMappingStore) Mapping(ctx context.Context, provider, externalID string) (*vehiclemodel.VehicleExternalMap, error) {
	var mapping vehiclemodel.VehicleExternalMap
	if err := s.db.QueryRowContext(ctx, `
SELECT id, vehicle_id, provider, external_id, COALESCE(external_name, ''), COALESCE(external_address, ''), COALESCE(external_protocol, ''), sync_status, COALESCE(last_seen_at, ''), created_at, updated_at
FROM vehicle_external_mappings
WHERE provider=? AND external_id=?
`, provider, externalID).Scan(&mapping.ID, &mapping.VehicleID, &mapping.Provider, &mapping.ExternalID, &mapping.ExternalName, &mapping.ExternalAddress, &mapping.ExternalProtocol, &mapping.SyncStatus, &mapping.LastSeenAt, &mapping.CreatedAt, &mapping.UpdatedAt); err != nil {
		return nil, fmt.Errorf("get vehicle external mapping: %w", err)
	}
	return &mapping, nil
}

func (s *externalMappingStore) Mappings(ctx context.Context, vehicleID string) ([]vehiclemodel.VehicleExternalMap, error) {
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

	mappings := []vehiclemodel.VehicleExternalMap{}
	for rows.Next() {
		var mapping vehiclemodel.VehicleExternalMap
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
