package authdata

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"railkeeper/backend/internal/credentials"
)

func isUniqueConstraint(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "unique")
}

func nullableString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func (s *Store) audit(ctx context.Context, actorUserID, action, targetType, targetID, detailsJSON string) error {
	if _, err := s.db.ExecContext(
		ctx,
		`INSERT INTO audit_logs(id, actor_user_id, action, target_type, target_id, created_at, details_json)
		 VALUES(?, ?, ?, ?, ?, ?, ?)`,
		randomID(),
		actorUserID,
		action,
		targetType,
		targetID,
		s.now().UTC().Format(time.RFC3339),
		detailsJSON,
	); err != nil {
		return fmt.Errorf("write audit log: %w", err)
	}
	return nil
}

func (s *Store) auditTx(ctx context.Context, tx *sql.Tx, actorUserID, action, targetType, targetID, detailsJSON string) error {
	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO audit_logs(id, actor_user_id, action, target_type, target_id, created_at, details_json)
		 VALUES(?, ?, ?, ?, ?, ?, ?)`,
		randomID(),
		actorUserID,
		action,
		targetType,
		targetID,
		s.now().UTC().Format(time.RFC3339),
		detailsJSON,
	); err != nil {
		return fmt.Errorf("write audit log: %w", err)
	}
	return nil
}

func randomID() string { return credentials.RandomID() }
