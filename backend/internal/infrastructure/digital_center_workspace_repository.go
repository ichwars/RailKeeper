package infrastructure

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"railkeeper/backend/internal/application"
)

const digitalCenterSessionMessageLimit = 100

var digitalCenterPersistedMessageCodes = map[application.DigitalCenterMessageCode]struct{}{
	application.DigitalCenterMessageConnectionSucceeded:   {},
	application.DigitalCenterMessageConnectionFailed:      {},
	application.DigitalCenterMessageConnectionInterrupted: {},
	application.DigitalCenterMessageReadStarted:           {},
	application.DigitalCenterMessageReadCompleted:         {},
	application.DigitalCenterMessageReadFailed:            {},
	application.DigitalCenterMessageParseFailed:           {},
	application.DigitalCenterMessageCapabilityUnavailable: {},
	application.DigitalCenterMessageLiveStarted:           {},
	application.DigitalCenterMessageLiveStopped:           {},
	application.DigitalCenterMessageLiveInterrupted:       {},
	application.DigitalCenterMessageWritePreviewFailed:    {},
	application.DigitalCenterMessageWriteFailed:           {},
	application.DigitalCenterMessageWriteVerified:         {},
	application.DigitalCenterMessageWriteVerifyFailed:     {},
}

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

func (repository *DigitalCenterWorkspaceRepository) AddMessage(
	ctx context.Context,
	message application.DigitalCenterSessionMessage,
) error {
	message.Code = application.DigitalCenterMessageCode(strings.TrimSpace(string(message.Code)))
	message.Message = normalizeDigitalCenterMessageText(message.Message)
	message.NextAction = normalizeDigitalCenterMessageText(message.NextAction)
	if _, allowed := digitalCenterPersistedMessageCodes[message.Code]; !allowed ||
		!isSafeDigitalCenterMessageText(message.Message, false, 512) ||
		!isSafeDigitalCenterMessageText(message.NextAction, true, 256) {
		return errors.New("add digital center session message: raw or private protocol content is not allowed")
	}
	if message.ID == "" {
		message.ID = randomID()
	}
	if message.CreatedAt == "" {
		message.CreatedAt = timestamp()
	}
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("add digital center session message: begin transaction: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO digital_center_session_messages(id, session_id, severity, code, message, next_action, created_at)
VALUES(?, ?, ?, ?, ?, ?, ?)`, message.ID, message.SessionID, message.Severity, message.Code,
		message.Message, message.NextAction, message.CreatedAt); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("add digital center session message: insert: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
DELETE FROM digital_center_session_messages
WHERE session_id=? AND id NOT IN (
  SELECT id FROM digital_center_session_messages
  WHERE session_id=? ORDER BY created_at DESC, rowid DESC LIMIT ?
)`, message.SessionID, message.SessionID, digitalCenterSessionMessageLimit); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("add digital center session message: enforce bound: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("add digital center session message: commit transaction: %w", err)
	}
	return nil
}

func (repository *DigitalCenterWorkspaceRepository) ListMessages(
	ctx context.Context,
	sessionID string,
) ([]application.DigitalCenterSessionMessage, error) {
	rows, err := repository.db.QueryContext(ctx, `
SELECT id, session_id, severity, code, message, next_action, created_at
FROM digital_center_session_messages
WHERE session_id=?
ORDER BY created_at DESC, rowid DESC`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("list digital center session messages: %w", err)
	}
	defer func() { _ = rows.Close() }()
	messages := []application.DigitalCenterSessionMessage{}
	for rows.Next() {
		message := application.DigitalCenterSessionMessage{}
		if err := rows.Scan(&message.ID, &message.SessionID, &message.Severity, &message.Code,
			&message.Message, &message.NextAction, &message.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan digital center session message: %w", err)
		}
		messages = append(messages, message)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate digital center session messages: %w", err)
	}
	return messages, nil
}

func (repository *DigitalCenterWorkspaceRepository) CreateWriteGrant(
	ctx context.Context,
	grant application.DigitalCenterWriteGrant,
) error {
	if grant.ID == "" {
		grant.ID = randomID()
	}
	if grant.CreatedAt == "" {
		grant.CreatedAt = timestamp()
	}
	result, err := repository.db.ExecContext(ctx, `
INSERT INTO digital_center_write_grants(
  id, token_hash, session_id, work_item_id, preview_hash, actor_user_id, expires_at, consumed_at, created_at
) SELECT ?, ?, ?, ?, ?, ?, ?, NULLIF(?, ''), ?
WHERE EXISTS(
  SELECT 1 FROM digital_center_work_items WHERE id=? AND session_id=?
)`, grant.ID, grant.TokenHash, grant.SessionID, grant.WorkItemID, grant.PreviewHash, grant.ActorUserID,
		grant.ExpiresAt, grant.ConsumedAt, grant.CreatedAt, grant.WorkItemID, grant.SessionID)
	if err != nil {
		return fmt.Errorf("create digital center write grant: %w", err)
	}
	if err := requireDigitalCenterUpdate(result, "create digital center write grant"); err != nil {
		return err
	}
	return nil
}

func (repository *DigitalCenterWorkspaceRepository) ConsumeWriteGrant(
	ctx context.Context,
	tokenHash string,
	actorUserID string,
) (application.DigitalCenterWriteGrant, error) {
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return application.DigitalCenterWriteGrant{}, fmt.Errorf("consume digital center write grant: begin: %w", err)
	}
	rollback := func(operationErr error) (application.DigitalCenterWriteGrant, error) {
		_ = tx.Rollback()
		return application.DigitalCenterWriteGrant{}, operationErr
	}
	consumedAt := timestamp()
	result, err := tx.ExecContext(ctx, `
UPDATE digital_center_write_grants
SET consumed_at=?
WHERE token_hash=? AND consumed_at IS NULL AND actor_user_id=?
  AND julianday(expires_at) > julianday(?)`, consumedAt, tokenHash, actorUserID, consumedAt)
	if err != nil {
		return rollback(fmt.Errorf("consume digital center write grant: update: %w", err))
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return rollback(fmt.Errorf("consume digital center write grant: read update result: %w", err))
	}
	grant, loadErr := scanDigitalCenterWriteGrant(tx.QueryRowContext(ctx, digitalCenterWriteGrantSelect+`
WHERE token_hash=?`, tokenHash))
	if loadErr != nil {
		return rollback(fmt.Errorf("consume digital center write grant: load: %w", loadErr))
	}
	if affected != 1 {
		switch {
		case grant.ActorUserID != actorUserID:
			return rollback(fmt.Errorf("consume digital center write grant: %w",
				application.ErrDigitalCenterGrantActorMismatch))
		case grant.ConsumedAt != "":
			return rollback(fmt.Errorf("consume digital center write grant: %w",
				application.ErrDigitalCenterGrantConsumed))
		}
		expiresAt, parseErr := time.Parse(time.RFC3339, grant.ExpiresAt)
		if parseErr != nil {
			return rollback(fmt.Errorf("consume digital center write grant: parse expiry: %w", parseErr))
		}
		if !time.Now().UTC().Before(expiresAt) {
			return rollback(fmt.Errorf("consume digital center write grant: %w",
				application.ErrDigitalCenterGrantExpired))
		}
		return rollback(fmt.Errorf("consume digital center write grant: %w",
			application.ErrDigitalCenterGrantConsumed))
	}
	if err := tx.Commit(); err != nil {
		return application.DigitalCenterWriteGrant{}, fmt.Errorf("consume digital center write grant: commit: %w", err)
	}
	grant.ConsumedAt = consumedAt
	return grant, nil
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

const digitalCenterWriteGrantSelect = `
SELECT id, token_hash, session_id, work_item_id, preview_hash, actor_user_id, expires_at,
       COALESCE(consumed_at, ''), created_at
FROM digital_center_write_grants`

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

func scanDigitalCenterWriteGrant(scanner digitalCenterRowScanner) (application.DigitalCenterWriteGrant, error) {
	grant := application.DigitalCenterWriteGrant{}
	if err := scanner.Scan(&grant.ID, &grant.TokenHash, &grant.SessionID, &grant.WorkItemID,
		&grant.PreviewHash, &grant.ActorUserID, &grant.ExpiresAt, &grant.ConsumedAt,
		&grant.CreatedAt); err != nil {
		return application.DigitalCenterWriteGrant{}, err
	}
	return grant, nil
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

func isSafeDigitalCenterMessageText(value string, allowEmpty bool, maximumRunes int) bool {
	if value == "" {
		return allowEmpty
	}
	if utf8.RuneCountInString(value) > maximumRunes {
		return false
	}
	for _, character := range value {
		if unicode.IsLetter(character) || unicode.IsDigit(character) || unicode.IsSpace(character) {
			continue
		}
		if !strings.ContainsRune(".,:;!?-/+%'", character) {
			return false
		}
	}
	compact := strings.Map(func(character rune) rune {
		if unicode.IsLetter(character) || unicode.IsDigit(character) {
			return unicode.ToLower(character)
		}
		return -1
	}, value)
	for _, restricted := range []string{
		"password", "passwort", "passwd", "pwd", "secret", "token", "apikey",
		"authorization", "bearer", "basic", "queryobjects", "release", "subscribe", "unsubscribe",
	} {
		if strings.Contains(compact, restricted) {
			return false
		}
	}
	return true
}

func normalizeDigitalCenterMessageText(value string) string {
	withoutControls := strings.Map(func(character rune) rune {
		if unicode.IsControl(character) || unicode.In(character, unicode.Cf, unicode.Co, unicode.Cs) {
			return ' '
		}
		return character
	}, value)
	return strings.Join(strings.Fields(withoutControls), " ")
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
