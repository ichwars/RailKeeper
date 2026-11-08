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
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin session revocation: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	result, err := tx.ExecContext(
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
	if err := s.auditTx(ctx, tx, actorUserID, "SessionRevoked", "session", sessionID, "{}"); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit session revocation: %w", err)
	}
	return nil
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
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin logout: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	result, err := tx.ExecContext(
		ctx,
		`UPDATE sessions SET revoked_at=? WHERE user_id=? AND token_hash=? AND revoked_at IS NULL`,
		s.now().UTC().Format(time.RFC3339),
		userID,
		tokenHash,
	)
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read logout count: %w", err)
	}
	if count > 0 {
		if err := s.auditTx(ctx, tx, userID, "Logout", "user", userID, "{}"); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit logout: %w", err)
	}
	return nil
}
