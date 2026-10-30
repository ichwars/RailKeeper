package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"railkeeper/backend/internal/application"
)

func (repository *DigitalCenterWorkspaceRepository) LoadECoSBackupContexts(
	ctx context.Context, host string, port int, ids []int,
) (map[int]application.ECoSBackupContext, error) {
	result := make(map[int]application.ECoSBackupContext, len(ids))
	for _, objectID := range ids {
		item, err := scanDigitalCenterWorkItem(repository.db.QueryRowContext(ctx, digitalCenterWorkItemSelect+`
WHERE center_object_id=? AND session_id IN (
 SELECT id FROM digital_center_read_sessions WHERE provider='ecos' AND state='ready' AND host=? AND port=?)
AND CASE WHEN json_valid(center_json) THEN COALESCE(json_extract(center_json,'$.cvRead.source'),'') ELSE '' END != 'backup'
ORDER BY julianday(updated_at) DESC, rowid DESC LIMIT 1`, strconv.Itoa(objectID), host, port))
		if errors.Is(err, sql.ErrNoRows) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("read saved ECoS backup context: %w", err)
		}
		entry := application.ECoSBackupContext{Item: item}
		var mappings, ownedMappings int
		err = repository.db.QueryRowContext(ctx, `
SELECT m.vehicle_id, v.name, v.decoder_type, COALESCE(m.external_address,''), COALESCE(m.external_protocol,''),
 (SELECT COUNT(*) FROM vehicle_external_mappings duplicate WHERE duplicate.provider='ecos' AND duplicate.external_id=m.external_id),
 (SELECT COUNT(*) FROM vehicle_external_mappings owned WHERE owned.provider='ecos' AND owned.vehicle_id=m.vehicle_id)
FROM vehicle_external_mappings m JOIN vehicles v ON v.id=m.vehicle_id
WHERE m.provider='ecos' AND m.external_id=? LIMIT 1`, strconv.Itoa(objectID)).Scan(
			&entry.VehicleID, &entry.VehicleName, &entry.DecoderType, &entry.ExternalAddress, &entry.ExternalProtocol,
			&mappings, &ownedMappings)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("read existing ECoS backup mapping: %w", err)
		}
		if mappings != 1 || ownedMappings != 1 {
			entry.VehicleID = ""
		}
		result[objectID] = entry
	}
	return result, nil
}

// Persist only an offline comparison workspace. No inventory or authentication rows are changed.
func (repository *DigitalCenterWorkspaceRepository) CreateECoSBackupSession(
	ctx context.Context, session application.DigitalCenterReadSession, items []application.DigitalCenterWorkItem,
	message application.DigitalCenterSessionMessage,
) (application.DigitalCenterReadSession, error) {
	if session.Provider != "ecos" || session.State != application.DigitalCenterSessionReady ||
		session.Capabilities != (application.DigitalCenterCapabilities{}) || len(items) == 0 {
		return application.DigitalCenterReadSession{}, application.ErrDigitalCenterDeviceOutput
	}
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return application.DigitalCenterReadSession{}, fmt.Errorf("begin ECoS backup preview: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	var currentHost, currentPort string
	err = tx.QueryRowContext(ctx, `SELECT host.value, port.value FROM app_settings host, app_settings port
WHERE host.key='digital.ecos.host' AND port.key='digital.ecos.port'`).Scan(&currentHost, &currentPort)
	if err != nil {
		return application.DigitalCenterReadSession{}, fmt.Errorf("check ECoS backup configuration: %w", err)
	}
	port, err := strconv.Atoi(currentPort)
	if err != nil || currentHost != session.Host || port != session.Port {
		return application.DigitalCenterReadSession{}, application.ErrECoSBackupStationMismatch
	}
	session.ID = randomID()
	session.CreatedAt, session.UpdatedAt = timestamp(), timestamp()
	_, err = tx.ExecContext(ctx, `INSERT INTO digital_center_read_sessions(
id,provider,state,host,port,capability_json,read_started_at,read_completed_at,created_by_user_id,created_at,updated_at)
VALUES(?,?,?,?,?,'{}',?,?,NULLIF(?,''),?,?)`, session.ID, session.Provider, session.State, session.Host, session.Port,
		session.ReadStartedAt, session.ReadCompletedAt, session.CreatedByUserID, session.CreatedAt, session.UpdatedAt)
	if err != nil {
		return application.DigitalCenterReadSession{}, fmt.Errorf("create ECoS backup preview: %w", err)
	}
	for _, item := range items {
		center, err := encodeDigitalCenterMap(item.Center)
		if err != nil {
			return application.DigitalCenterReadSession{}, fmt.Errorf("encode backup CV snapshot: %w", err)
		}
		railkeeper, err := encodeDigitalCenterMap(item.RailKeeper)
		if err != nil {
			return application.DigitalCenterReadSession{}, fmt.Errorf("encode backup comparison: %w", err)
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO digital_center_work_items(
id,session_id,center_object_id,vehicle_id,name,decoder_address,protocol,compare_status,station_status,
center_json,railkeeper_json,proposed_json,conflict_json,created_at,updated_at)
VALUES(?,?,?,NULLIF(?,''),?,?,?,?,?,?,?,'{}','[]',?,?)`, randomID(), session.ID, item.CenterObjectID,
			item.VehicleID, item.Name, item.Address, item.Protocol, item.CompareStatus, item.StationStatus,
			center, railkeeper, session.CreatedAt, session.UpdatedAt)
		if err != nil {
			return application.DigitalCenterReadSession{}, fmt.Errorf("save ECoS backup CV snapshot: %w", err)
		}
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO digital_center_session_messages(
id,session_id,severity,code,message,next_action,created_at) VALUES(?,?,?,?,?,?,?)`,
		randomID(), session.ID, message.Severity, message.Code, message.Message, message.NextAction, session.CreatedAt)
	if err != nil {
		return application.DigitalCenterReadSession{}, fmt.Errorf("save ECoS backup provenance: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return application.DigitalCenterReadSession{}, fmt.Errorf("commit ECoS backup preview: %w", err)
	}
	return session, nil
}
