package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"railkeeper/backend/internal/application"
)

func (repository *DigitalCenterWorkspaceRepository) LoadECoSBackupContexts(
	ctx context.Context, host string, port int, ids []int,
) (map[int]application.ECoSBackupContext, error) {
	result := make(map[int]application.ECoSBackupContext, len(ids))
	if len(ids) == 0 {
		return result, nil
	}
	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = strconv.Itoa(id)
	}
	payload, err := json.Marshal(keys)
	if err != nil {
		return nil, err
	}
	rows, err := repository.db.QueryContext(ctx, `
WITH ranked AS (
  SELECT item.rowid AS item_row, ROW_NUMBER() OVER (
    PARTITION BY item.center_object_id ORDER BY julianday(item.updated_at) DESC, item.rowid DESC
  ) AS rank
  FROM digital_center_work_items item
  JOIN digital_center_read_sessions session ON session.id=item.session_id
  WHERE session.provider='ecos' AND session.state='ready' AND session.host=? AND session.port=?
    AND item.center_object_id IN (SELECT value FROM json_each(?))
    AND CASE WHEN json_valid(item.center_json)
             THEN COALESCE(json_extract(item.center_json,'$.cvRead.source'),'') ELSE '' END != 'backup'
)
`+digitalCenterWorkItemSelect+` WHERE rowid IN (SELECT item_row FROM ranked WHERE rank=1)`, host, port, string(payload))
	if err != nil {
		return nil, fmt.Errorf("read saved ECoS backup contexts: %w", err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		item, err := scanDigitalCenterWorkItem(rows)
		if err != nil {
			return nil, fmt.Errorf("scan saved ECoS backup context: %w", err)
		}
		objectID, err := strconv.Atoi(item.CenterObjectID)
		if err != nil {
			return nil, fmt.Errorf("read ECoS backup context ID: %w", err)
		}
		result[objectID] = application.ECoSBackupContext{Item: item}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return result, nil
	}
	if err := repository.loadECoSBackupMappings(ctx, string(payload), result); err != nil {
		return nil, err
	}
	return result, nil
}

func (repository *DigitalCenterWorkspaceRepository) loadECoSBackupMappings(
	ctx context.Context, idsJSON string, contexts map[int]application.ECoSBackupContext,
) error {
	rows, err := repository.db.QueryContext(ctx, `
SELECT m.external_id, m.vehicle_id, v.name, COALESCE(v.decoder_type,''), COALESCE(m.external_address,''),
       COALESCE(m.external_protocol,''),
       (SELECT COUNT(*) FROM vehicle_external_mappings owned WHERE owned.provider='ecos' AND owned.vehicle_id=m.vehicle_id)
FROM vehicle_external_mappings m
JOIN vehicles v ON v.id=m.vehicle_id
WHERE m.provider='ecos' AND m.external_id IN (SELECT value FROM json_each(?))`, idsJSON)
	if err != nil {
		return fmt.Errorf("read existing ECoS backup mappings: %w", err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var objectID, ownedMappings int
		var mapping application.ECoSBackupContext
		if err := rows.Scan(&objectID, &mapping.VehicleID, &mapping.VehicleName, &mapping.DecoderType,
			&mapping.ExternalAddress, &mapping.ExternalProtocol, &ownedMappings); err != nil {
			return fmt.Errorf("scan existing ECoS backup mapping: %w", err)
		}
		entry, exists := contexts[objectID]
		if !exists {
			continue
		}
		mapping.Item = entry.Item
		if ownedMappings != 1 {
			mapping.VehicleID = ""
		}
		contexts[objectID] = mapping
	}
	return rows.Err()
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
