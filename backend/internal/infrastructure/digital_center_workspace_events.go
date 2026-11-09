package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"railkeeper/backend/internal/application"
	"strings"
	"unicode"
	"unicode/utf8"
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
	application.DigitalCenterMessageWriteUnknown:          {},
	application.DigitalCenterMessageWriteVerified:         {},
	application.DigitalCenterMessageWriteVerifyFailed:     {},
	application.DigitalCenterMessageLiveRestartFailed:     {},
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
