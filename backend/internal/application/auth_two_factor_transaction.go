package application

import (
	"context"
	"fmt"
	"time"
)

func (s *AuthService) commitTwoFactorEnable(ctx context.Context, userID, secret string, step int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin two factor enable: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := requireAuthMutation(tx.ExecContext(ctx, `UPDATE users SET two_factor_enabled=1,
two_factor_enabled_at=?,two_factor_last_step=? WHERE id=? AND two_factor_secret=? AND two_factor_enabled=0
AND (two_factor_last_step IS NULL OR two_factor_last_step<?)`,
		s.now().UTC().Format(time.RFC3339), step, userID, secret, step)); err != nil {
		return fmt.Errorf("enable two factor: %w", err)
	}
	if err := s.auditTx(ctx, tx, userID, "TwoFactorEnabled", "user", userID, "{}"); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit two factor enable: %w", err)
	}
	return nil
}

func (s *AuthService) commitTwoFactorDisable(ctx context.Context, userID, sessionToken, passwordHash, secret string, enabled bool, step int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin two factor disable: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := requireAuthMutation(tx.ExecContext(ctx, `UPDATE users SET two_factor_secret=NULL,
two_factor_enabled=0,two_factor_enabled_at=NULL,two_factor_last_step=NULL
WHERE id=? AND password_hash=? AND COALESCE(two_factor_secret,'')=? AND two_factor_enabled=?
AND (two_factor_enabled=0 OR two_factor_last_step IS NULL OR two_factor_last_step<?)`,
		userID, passwordHash, secret, enabled, step)); err != nil {
		return fmt.Errorf("disable two factor: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE sessions SET revoked_at=?
WHERE user_id=? AND token_hash<>? AND revoked_at IS NULL`,
		s.now().UTC().Format(time.RFC3339), userID, hashToken(sessionToken)); err != nil {
		return fmt.Errorf("revoke two factor sessions: %w", err)
	}
	if err := s.auditTx(ctx, tx, userID, "TwoFactorDisabled", "user", userID, "{}"); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit two factor disable: %w", err)
	}
	return nil
}
