package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// Source context is derived from the saved comparison workspace. It never
// replaces decoder_type or inserts a manufacturer/firmware CV reading.
type ECoSDecoderProfileEvidence struct {
	Name     string `json:"name"`
	ObjectID string `json:"objectId"`
	Host     string `json:"host"`
	ReadAt   string `json:"readAt"`
}

func (s *VehicleService) loadECoSDecoderProfile(ctx context.Context, vehicleID string) (*ECoSDecoderProfileEvidence, error) {
	var data, status, protocol, completed, mappedAddress, mappedProtocol string
	var address int
	var source ECoSDecoderProfileEvidence
	// Separate backup workspaces must not supersede the last device observation.
	err := s.db.QueryRowContext(ctx, `
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
LIMIT 1`, vehicleID).Scan(&data, &status, &address, &protocol, &source.Host, &completed,
		&source.ObjectID, &mappedAddress, &mappedProtocol)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("load cached ECoS decoder profile: %w", err)
	}
	return validatedECoSDecoderProfile(data, status, protocol, completed, mappedAddress, mappedProtocol, address, source), nil
}

// Invalid optional source evidence is omitted without failing the vehicle read.
func validatedECoSDecoderProfile(data, status, protocol, completed, mappedAddress, mappedProtocol string,
	address int, source ECoSDecoderProfileEvidence,
) *ECoSDecoderProfileEvidence {
	// A newer missing/incomplete or conflicting snapshot invalidates older context.
	if (status != "read" && status != "incomplete") || len(data) > 8<<20 {
		return nil
	}
	if mappedAddress != "" {
		mapped, err := strconv.Atoi(mappedAddress)
		if err != nil || mapped != address {
			return nil
		}
	}
	stationProtocol, err := normalizeDigitalCenterProtocol(protocol)
	if err != nil || stationProtocol != "DCC" {
		return nil
	}
	if mappedProtocol != "" {
		mapped, err := normalizeDigitalCenterProtocol(mappedProtocol)
		if err != nil || mapped != stationProtocol {
			return nil
		}
	}
	var snapshot struct {
		Profile string            `json:"profile"`
		Read    *ECoSCVReadState  `json:"cvRead"`
		Sources *ECoSCVSourceInfo `json:"cvSources"`
	}
	if json.Unmarshal([]byte(data), &snapshot) != nil || !utf8.ValidString(snapshot.Profile) || len(snapshot.Profile) > 160 {
		return nil
	}
	if snapshot.Sources != nil && len(snapshot.Sources.Conflicts) != 0 {
		return nil // The latest contradictory observation must not revive older model context.
	}
	source.Name = strings.Join(strings.Fields(snapshot.Profile), " ")
	source.ReadAt = completed
	if status == "incomplete" && (snapshot.Read == nil || snapshot.Read.Source != "database") {
		return nil
	}
	if snapshot.Read != nil && snapshot.Read.Source == "database" {
		if snapshot.Read.SourceHost != source.Host || snapshot.Read.NextCV != 1025 || snapshot.Read.Unavailable != 0 {
			return nil
		}
		source.ReadAt = snapshot.Read.ReadAt
	}
	if _, err := time.Parse(time.RFC3339Nano, source.ReadAt); err != nil || source.Name == "" {
		return nil
	}
	return &source
}
