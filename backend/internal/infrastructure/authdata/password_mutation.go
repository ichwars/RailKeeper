package authdata

import (
	"context"
	"fmt"
	"time"

	"railkeeper/backend/internal/authentication/management"
)

func (s *Store) ChangePassword(ctx context.Context, change management.PasswordChange) error {
	userID, nextHash, tokenHash := change.UserID, change.NextHash, change.SessionTokenHash

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin password change: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(ctx, `UPDATE users SET password_hash=? WHERE id=?`, nextHash, userID); err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	if _, err = tx.ExecContext(
		ctx,
		`UPDATE sessions SET revoked_at=? WHERE user_id=? AND token_hash<>? AND revoked_at IS NULL`,
		s.now().UTC().Format(time.RFC3339),
		userID,
		tokenHash,
	); err != nil {
		return fmt.Errorf("revoke other sessions: %w", err)
	}
	if err = s.auditTx(ctx, tx, userID, "PasswordChanged", "user", userID, "{}"); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit password change: %w", err)
	}
	return nil
}
