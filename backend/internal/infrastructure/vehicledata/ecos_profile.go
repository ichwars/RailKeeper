package vehicledata

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"railkeeper/backend/internal/ecosmodel"
)

func ReadCachedECoSDecoderProfile(ctx context.Context, db *sql.DB, vehicleID string) (*ecosmodel.CachedDecoderProfile, error) {
	var observation ecosmodel.CachedDecoderProfile
	// Separate backup workspaces must not supersede the last device observation.
	err := db.QueryRowContext(ctx, `
SELECT w.center_json, w.station_status, w.decoder_address, w.protocol,
       r.host, COALESCE(r.read_completed_at, ''), m.external_id,
       COALESCE(m.external_address, ''), COALESCE(m.external_protocol, '')
FROM vehicle_external_mappings m
JOIN digital_center_work_items w ON w.center_object_id=m.external_id
JOIN digital_center_read_sessions r ON r.id=w.session_id AND r.provider=m.provider
JOIN app_settings host ON host.key='digital.ecos.host' AND host.value=r.host
JOIN app_settings port ON port.key='digital.ecos.port' AND CAST(port.value AS INTEGER)=r.port
WHERE m.vehicle_id=? AND m.provider='ecos' AND r.state='ready'
  AND CASE WHEN json_valid(w.center_json)
      THEN COALESCE(json_extract(w.center_json, '$.cvRead.source'), '') ELSE '' END != 'backup'
  AND (SELECT COUNT(*) FROM vehicle_external_mappings own
       WHERE own.vehicle_id=m.vehicle_id AND own.provider='ecos')=1
ORDER BY julianday(w.updated_at) DESC, w.rowid DESC
LIMIT 1`, vehicleID).Scan(&observation.Data, &observation.Status, &observation.Address, &observation.Protocol, &observation.Source.Host, &observation.Completed,
		&observation.Source.ObjectID, &observation.MappedAddress, &observation.MappedProtocol)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("load cached ECoS decoder profile: %w", err)
	}
	return &observation, nil
}
