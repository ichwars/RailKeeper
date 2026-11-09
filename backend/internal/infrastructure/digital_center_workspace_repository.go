package infrastructure

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"railkeeper/backend/internal/application"
)

type DigitalCenterWorkspaceRepository struct {
	db *sql.DB
}

func NewDigitalCenterWorkspaceRepository(db *sql.DB) *DigitalCenterWorkspaceRepository {
	return &DigitalCenterWorkspaceRepository{db: db}
}

func (repository *DigitalCenterWorkspaceRepository) CreateSession(
	ctx context.Context,
	session application.DigitalCenterReadSession,
) (application.DigitalCenterReadSession, error) {
	capabilities, err := encodeDigitalCenterJSON(session.Capabilities)
	if err != nil {
		return application.DigitalCenterReadSession{}, fmt.Errorf("encode digital center capabilities: %w", err)
	}
	now := timestamp()
	if session.ID == "" {
		session.ID = randomID()
	}
	session.CreatedAt = now
	session.UpdatedAt = now
	if _, err := repository.db.ExecContext(ctx, `
INSERT INTO digital_center_read_sessions(
  id, provider, state, host, port, capability_json, read_started_at, read_completed_at,
  created_by_user_id, created_at, updated_at
) VALUES(?, ?, ?, ?, ?, ?, NULLIF(?, ''), NULLIF(?, ''), NULLIF(?, ''), ?, ?)`,
		session.ID, session.Provider, session.State, session.Host, session.Port, capabilities,
		session.ReadStartedAt, session.ReadCompletedAt, session.CreatedByUserID, now, now); err != nil {
		return application.DigitalCenterReadSession{}, fmt.Errorf("create digital center read session: %w", err)
	}
	return repository.GetSession(ctx, session.ID)
}

func (repository *DigitalCenterWorkspaceRepository) UpdateSession(
	ctx context.Context,
	session application.DigitalCenterReadSession,
) error {
	capabilities, err := encodeDigitalCenterJSON(session.Capabilities)
	if err != nil {
		return fmt.Errorf("encode digital center capabilities: %w", err)
	}
	result, err := repository.db.ExecContext(ctx, `
UPDATE digital_center_read_sessions
SET provider=?, state=?, host=?, port=?, capability_json=?, read_started_at=NULLIF(?, ''),
    read_completed_at=NULLIF(?, ''), updated_at=?
WHERE id=?`, session.Provider, session.State, session.Host, session.Port, capabilities,
		session.ReadStartedAt, session.ReadCompletedAt, timestamp(), session.ID)
	if err != nil {
		return fmt.Errorf("update digital center read session: %w", err)
	}
	return requireDigitalCenterUpdate(result, "update digital center read session")
}

func (repository *DigitalCenterWorkspaceRepository) GetSession(
	ctx context.Context,
	id string,
) (application.DigitalCenterReadSession, error) {
	session, err := scanDigitalCenterReadSession(repository.db.QueryRowContext(ctx,
		digitalCenterSessionSelect+` WHERE id=?`, id))
	if err != nil {
		return application.DigitalCenterReadSession{}, fmt.Errorf("get digital center read session: %w", err)
	}
	return session, nil
}

func (repository *DigitalCenterWorkspaceRepository) ReplaceWorkItems(
	ctx context.Context,
	sessionID string,
	items []application.DigitalCenterWorkItem,
) error {
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("replace digital center work items: begin transaction: %w", err)
	}
	rollback := func(operationErr error) error {
		_ = tx.Rollback()
		return fmt.Errorf("replace digital center work items: %w", operationErr)
	}
	var sessionExists int
	if err := tx.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM digital_center_read_sessions WHERE id=?)`, sessionID).
		Scan(&sessionExists); err != nil {
		return rollback(fmt.Errorf("check read session: %w", err))
	}
	if sessionExists == 0 {
		return rollback(sql.ErrNoRows)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM digital_center_work_items WHERE session_id=?`, sessionID); err != nil {
		return rollback(fmt.Errorf("delete existing work items: %w", err))
	}
	now := timestamp()
	for _, item := range items {
		center, err := encodeDigitalCenterMap(item.Center)
		if err != nil {
			return rollback(fmt.Errorf("encode center payload: %w", err))
		}
		railKeeper, err := encodeDigitalCenterMap(item.RailKeeper)
		if err != nil {
			return rollback(fmt.Errorf("encode RailKeeper payload: %w", err))
		}
		proposed, err := encodeDigitalCenterMap(item.Proposed)
		if err != nil {
			return rollback(fmt.Errorf("encode proposed payload: %w", err))
		}
		conflicts, err := encodeDigitalCenterConflicts(item.Conflicts)
		if err != nil {
			return rollback(fmt.Errorf("encode conflict payload: %w", err))
		}
		if item.ID == "" {
			item.ID = randomID()
		}
		createdAt := item.CreatedAt
		if createdAt == "" {
			createdAt = now
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO digital_center_work_items(
  id, session_id, center_object_id, vehicle_id, name, decoder_address, protocol, compare_status,
  station_status, center_json, railkeeper_json, proposed_json, conflict_json, created_at, updated_at
) VALUES(?, ?, ?, NULLIF(?, ''), ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, item.ID, sessionID,
			item.CenterObjectID, item.VehicleID, item.Name, item.Address, item.Protocol, item.CompareStatus,
			item.StationStatus, center, railKeeper, proposed, conflicts, createdAt, now); err != nil {
			return rollback(fmt.Errorf("insert work item %q: %w", item.CenterObjectID, err))
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("replace digital center work items: commit transaction: %w", err)
	}
	return nil
}

func (repository *DigitalCenterWorkspaceRepository) ListWorkItems(
	ctx context.Context,
	sessionID string,
) ([]application.DigitalCenterWorkItem, error) {
	rows, err := repository.db.QueryContext(ctx, digitalCenterWorkItemSelect+`
WHERE session_id=?
ORDER BY name COLLATE NOCASE, center_object_id, id`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("list digital center work items: %w", err)
	}
	defer func() { _ = rows.Close() }()
	items := []application.DigitalCenterWorkItem{}
	for rows.Next() {
		item, err := scanDigitalCenterWorkItem(rows)
		if err != nil {
			return nil, fmt.Errorf("scan digital center work item: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate digital center work items: %w", err)
	}
	return items, nil
}

func (repository *DigitalCenterWorkspaceRepository) GetWorkItem(
	ctx context.Context,
	sessionID string,
	id string,
) (application.DigitalCenterWorkItem, error) {
	item, err := scanDigitalCenterWorkItem(repository.db.QueryRowContext(ctx,
		digitalCenterWorkItemSelect+` WHERE session_id=? AND id=?`, sessionID, id))
	if err != nil {
		return application.DigitalCenterWorkItem{}, fmt.Errorf("get digital center work item: %w", err)
	}
	return item, nil
}

func (repository *DigitalCenterWorkspaceRepository) UpdateWorkItem(
	ctx context.Context,
	item application.DigitalCenterWorkItem,
) (application.DigitalCenterWorkItem, error) {
	center, err := encodeDigitalCenterMap(item.Center)
	if err != nil {
		return application.DigitalCenterWorkItem{}, fmt.Errorf("update digital center work item: encode center: %w", err)
	}
	railKeeper, err := encodeDigitalCenterMap(item.RailKeeper)
	if err != nil {
		return application.DigitalCenterWorkItem{}, fmt.Errorf("update digital center work item: encode RailKeeper: %w", err)
	}
	proposed, err := encodeDigitalCenterMap(item.Proposed)
	if err != nil {
		return application.DigitalCenterWorkItem{}, fmt.Errorf("update digital center work item: encode proposed: %w", err)
	}
	conflicts, err := encodeDigitalCenterConflicts(item.Conflicts)
	if err != nil {
		return application.DigitalCenterWorkItem{}, fmt.Errorf("update digital center work item: encode conflicts: %w", err)
	}
	result, err := repository.db.ExecContext(ctx, `
UPDATE digital_center_work_items
SET center_object_id=?, vehicle_id=NULLIF(?, ''), name=?, decoder_address=?, protocol=?, compare_status=?, station_status=?,
    center_json=?, railkeeper_json=?, proposed_json=?, conflict_json=?, updated_at=?
WHERE session_id=? AND id=?`, item.CenterObjectID, item.VehicleID, item.Name, item.Address, item.Protocol, item.CompareStatus,
		item.StationStatus, center, railKeeper, proposed, conflicts, timestamp(), item.SessionID, item.ID)
	if err != nil {
		return application.DigitalCenterWorkItem{}, fmt.Errorf("update digital center work item: %w", err)
	}
	if err := requireDigitalCenterUpdate(result, "update digital center work item"); err != nil {
		return application.DigitalCenterWorkItem{}, err
	}
	return repository.GetWorkItem(ctx, item.SessionID, item.ID)
}

const digitalCenterSessionSelect = `
SELECT id, provider, state, host, port, capability_json, COALESCE(read_started_at, ''),
       COALESCE(read_completed_at, ''), COALESCE(created_by_user_id, ''), created_at, updated_at
FROM digital_center_read_sessions`

const digitalCenterWorkItemSelect = `
SELECT id, session_id, center_object_id, COALESCE(vehicle_id, ''), name, decoder_address, protocol,
       compare_status, station_status, center_json, railkeeper_json, proposed_json, conflict_json,
       created_at, updated_at
FROM digital_center_work_items`

type digitalCenterRowScanner interface {
	Scan(...any) error
}

func scanDigitalCenterReadSession(
	scanner digitalCenterRowScanner,
) (application.DigitalCenterReadSession, error) {
	session := application.DigitalCenterReadSession{}
	var capabilities string
	if err := scanner.Scan(&session.ID, &session.Provider, &session.State, &session.Host, &session.Port,
		&capabilities, &session.ReadStartedAt, &session.ReadCompletedAt, &session.CreatedByUserID,
		&session.CreatedAt, &session.UpdatedAt); err != nil {
		return application.DigitalCenterReadSession{}, err
	}
	if err := decodeDigitalCenterJSON(capabilities, &session.Capabilities); err != nil {
		return application.DigitalCenterReadSession{}, fmt.Errorf("decode digital center capabilities: %w", err)
	}
	return session, nil
}

func scanDigitalCenterWorkItem(scanner digitalCenterRowScanner) (application.DigitalCenterWorkItem, error) {
	item := application.DigitalCenterWorkItem{}
	var center, railKeeper, proposed, conflicts string
	if err := scanner.Scan(&item.ID, &item.SessionID, &item.CenterObjectID, &item.VehicleID, &item.Name,
		&item.Address, &item.Protocol, &item.CompareStatus, &item.StationStatus, &center, &railKeeper,
		&proposed, &conflicts, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return application.DigitalCenterWorkItem{}, err
	}
	if err := decodeDigitalCenterJSON(center, &item.Center); err != nil {
		return application.DigitalCenterWorkItem{}, fmt.Errorf("decode center payload: %w", err)
	}
	if err := decodeDigitalCenterJSON(railKeeper, &item.RailKeeper); err != nil {
		return application.DigitalCenterWorkItem{}, fmt.Errorf("decode RailKeeper payload: %w", err)
	}
	if err := decodeDigitalCenterJSON(proposed, &item.Proposed); err != nil {
		return application.DigitalCenterWorkItem{}, fmt.Errorf("decode proposed payload: %w", err)
	}
	if err := decodeDigitalCenterJSON(conflicts, &item.Conflicts); err != nil {
		return application.DigitalCenterWorkItem{}, fmt.Errorf("decode conflict payload: %w", err)
	}
	if item.Center == nil {
		item.Center = map[string]any{}
	}
	if item.RailKeeper == nil {
		item.RailKeeper = map[string]any{}
	}
	if item.Proposed == nil {
		item.Proposed = map[string]any{}
	}
	if item.Conflicts == nil {
		item.Conflicts = []map[string]any{}
	}
	return item, nil
}

func encodeDigitalCenterMap(value map[string]any) (string, error) {
	if value == nil {
		value = map[string]any{}
	}
	return encodeDigitalCenterJSON(value)
}

func encodeDigitalCenterConflicts(value []map[string]any) (string, error) {
	if value == nil {
		value = []map[string]any{}
	}
	return encodeDigitalCenterJSON(value)
}

func encodeDigitalCenterJSON(value any) (string, error) {
	payload, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(payload), nil
}

func decodeDigitalCenterJSON(payload string, target any) error {
	decoder := json.NewDecoder(bytes.NewBufferString(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("multiple JSON values")
		}
		return err
	}
	return nil
}

func requireDigitalCenterUpdate(result sql.Result, operation string) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: read update result: %w", operation, err)
	}
	if affected == 0 {
		return fmt.Errorf("%s: %w", operation, sql.ErrNoRows)
	}
	return nil
}

var _ application.DigitalCenterWorkspaceRepository = (*DigitalCenterWorkspaceRepository)(nil)
