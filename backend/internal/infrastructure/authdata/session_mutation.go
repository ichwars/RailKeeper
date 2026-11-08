package authdata

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"railkeeper/backend/internal/authentication/management"
)

func (s *Store) RevokeSession(ctx context.Context, actorUserID, sessionID string) error {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return management.ErrSessionNotFound
	}
	result, err := s.db.ExecContext(
		ctx,
		`UPDATE sessions SET revoked_at=? WHERE id=? AND revoked_at IS NULL`,
		s.now().UTC().Format(time.RFC3339),
		sessionID,
	)
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read revoked session count: %w", err)
	}
	if affected == 0 {
		return management.ErrSessionNotFound
	}
	return s.audit(ctx, actorUserID, "SessionRevoked", "session", sessionID, "{}")
}

func (s *Store) SessionUser(ctx context.Context, tokenHash string) (string, string, error) {
	var userID, csrfToken string
	if err := s.db.QueryRowContext(
		ctx,
		`SELECT user_id, csrf_token
		 FROM sessions
		 WHERE token_hash=? AND revoked_at IS NULL AND expires_at > ?`,
		tokenHash,
		s.now().UTC().Format(time.RFC3339),
	).Scan(&userID, &csrfToken); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", management.ErrUnauthorized
		}
		return "", "", fmt.Errorf("read session: %w", err)
	}

	return userID, csrfToken, nil
}

func (s *Store) RevokeOwnSession(ctx context.Context, userID, tokenHash string) error {
	if _, err := s.db.ExecContext(
		ctx,
		`UPDATE sessions SET revoked_at=? WHERE token_hash=? AND revoked_at IS NULL`,
		s.now().UTC().Format(time.RFC3339),
		tokenHash,
	); err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}

	return s.audit(ctx, userID, "Logout", "user", userID, "{}")
}
