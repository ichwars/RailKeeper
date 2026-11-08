package authdata

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"railkeeper/backend/internal/authentication"
	"railkeeper/backend/internal/credentials"
)

var _ authentication.OriginStore = (*Store)(nil)

func (s *Store) FindOrigin(ctx context.Context, username, tokenHash string) (authentication.Origin, authentication.Account, bool, error) {
	var origin authentication.Origin
	var account authentication.Account
	var expires string
	err := s.db.QueryRowContext(ctx, `SELECT o.user_id,o.token_hash,o.budget_key,o.state_hash,o.expires_at,
u.id,u.username,u.password_hash,COALESCE(u.two_factor_secret,''),u.two_factor_enabled
FROM login_origins o JOIN users u ON u.id=o.user_id WHERE o.token_hash=? AND u.username=?`, tokenHash, username).
		Scan(&origin.UserID, &origin.TokenHash, &origin.BudgetKey, &origin.StateHash, &expires,
			&account.ID, &account.Username, &account.PasswordHash, &account.TwoFactorSecret, &account.TwoFactorEnabled)
	if errors.Is(err, sql.ErrNoRows) {
		return origin, account, false, nil
	}
	if err != nil {
		return origin, account, false, fmt.Errorf("read login origin: %w", err)
	}
	origin.ExpiresAt, err = time.Parse(time.RFC3339, expires)
	if err != nil {
		return origin, account, false, fmt.Errorf("parse login origin expiry: %w", err)
	}
	return origin, account, true, nil
}

func saveOrigin(ctx context.Context, tx *sql.Tx, origin authentication.Origin, now string) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM login_origins WHERE expires_at<=?`, now); err != nil {
		return fmt.Errorf("expire login origins: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO login_origins(token_hash,user_id,budget_key,state_hash,expires_at,created_at)
VALUES(?,?,?,?,?,?) ON CONFLICT(token_hash) DO UPDATE SET state_hash=excluded.state_hash`,
		origin.TokenHash, origin.UserID, origin.BudgetKey, origin.StateHash, origin.ExpiresAt.UTC().Format(time.RFC3339), now); err != nil {
		return fmt.Errorf("save login origin: %w", err)
	}
	// Keep at most ten remembered browsers per account, including timestamp ties.
	if _, err := tx.ExecContext(ctx, `DELETE FROM login_origins WHERE user_id=? AND token_hash IN
(SELECT token_hash FROM login_origins WHERE user_id=? AND token_hash<>? ORDER BY created_at DESC,token_hash DESC LIMIT -1 OFFSET 9)`,
		origin.UserID, origin.UserID, origin.TokenHash); err != nil {
		return fmt.Errorf("bound login origins: %w", err)
	}
	return nil
}

func (s *Store) ResetLoginBudget(ctx context.Context, actorID, userID string, now time.Time) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin login recovery: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	var admin bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM user_roles ur JOIN roles r ON r.id=ur.role_id WHERE ur.user_id=? AND r.name='Admin')`, actorID).Scan(&admin); err != nil {
		return fmt.Errorf("check recovery permission: %w", err)
	}
	if !admin {
		return authentication.ErrForbidden
	}
	var username string
	if err := tx.QueryRowContext(ctx, `SELECT username FROM users WHERE id=?`, userID).Scan(&username); errors.Is(err, sql.ErrNoRows) {
		return authentication.ErrUserNotFound
	} else if err != nil {
		return fmt.Errorf("read recovery account: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM rate_limit_attempts WHERE (scope=? AND key=?) OR (scope=? AND key IN (SELECT budget_key FROM login_origins WHERE user_id=?))`,
		authentication.AccountBudgetScope, credentials.HashToken(username), authentication.OriginBudgetScope, userID); err != nil {
		return fmt.Errorf("reset login budgets: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM login_origins WHERE user_id=?`, userID); err != nil {
		return fmt.Errorf("revoke login origins: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO audit_logs(id,actor_user_id,action,target_type,target_id,created_at,details_json) VALUES(?,?,'LoginBudgetReset','user',?,?,'{}')`, credentials.RandomID(), actorID, userID, now.UTC().Format(time.RFC3339)); err != nil {
		return fmt.Errorf("audit login recovery: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit login recovery: %w", err)
	}
	return nil
}
