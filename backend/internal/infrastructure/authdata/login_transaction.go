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

func (s *Store) CommitLogin(ctx context.Context, proof authentication.Commit) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin login: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	var currentHash, currentSecret string
	var currentEnabled bool
	var lastStep sql.NullInt64
	err = tx.QueryRowContext(ctx, `SELECT password_hash,COALESCE(two_factor_secret,''),two_factor_enabled,
two_factor_last_step FROM users WHERE id=?`, proof.Account.ID).Scan(&currentHash, &currentSecret, &currentEnabled, &lastStep)
	if errors.Is(err, sql.ErrNoRows) {
		return authentication.ErrInvalidLogin
	}
	if err != nil {
		return fmt.Errorf("read current login state: %w", err)
	}
	if currentHash != proof.Account.PasswordHash || currentSecret != proof.Account.TwoFactorSecret || currentEnabled != proof.Account.TwoFactorEnabled {
		return authentication.ErrInvalidLogin
	}
	if proof.OriginRecognized {
		if err := recheckOrigin(ctx, tx, proof); err != nil {
			return err
		}
	}
	if proof.Account.TwoFactorEnabled {
		if lastStep.Valid && proof.TwoFactorStep <= lastStep.Int64 {
			return authentication.ErrInvalidLogin
		}
		if err := requireOne(tx.ExecContext(ctx, `UPDATE users SET two_factor_last_step=?
WHERE id=? AND (two_factor_last_step IS NULL OR two_factor_last_step<?)`,
			proof.TwoFactorStep, proof.Account.ID, proof.TwoFactorStep)); err != nil {
			return fmt.Errorf("consume two factor code: %w", err)
		}
	}
	if proof.UpgradedHash != "" {
		if err := requireOne(tx.ExecContext(ctx, `UPDATE users SET password_hash=? WHERE id=? AND password_hash=?`,
			proof.UpgradedHash, proof.Account.ID, proof.Account.PasswordHash)); err != nil {
			return fmt.Errorf("upgrade login password hash: %w", err)
		}
	}
	now := proof.Now.UTC().Format(time.RFC3339)
	if proof.Origin != nil {
		if err := saveOrigin(ctx, tx, *proof.Origin, now); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO sessions(id,user_id,token_hash,csrf_token,expires_at,created_at)
VALUES(?,?,?,?,?,?)`, credentials.RandomID(), proof.Account.ID, credentials.HashToken(proof.Session.SessionToken), proof.Session.CSRFToken,
		proof.Session.ExpiresAt.UTC().Format(time.RFC3339), now); err != nil {
		return fmt.Errorf("create login session: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO audit_logs(id,actor_user_id,action,target_type,target_id,created_at,details_json)
VALUES(?,?,'Login','user',?,?,'{}')`, credentials.RandomID(), proof.Account.ID, proof.Account.ID, now); err != nil {
		return fmt.Errorf("write audit log: %w", err)
	}
	if proof.AttemptScope == "" {
		proof.AttemptScope = authentication.AccountBudgetScope
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM rate_limit_attempts WHERE scope=? AND key=?`,
		proof.AttemptScope, proof.AttemptKey); err != nil {
		return fmt.Errorf("reset successful login budget: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit login: %w", err)
	}
	return nil
}

// Recheck recognition before any mutation can revoke it (including hash upgrade).
// A reset or role change acknowledged before this transaction must stay effective.
func recheckOrigin(ctx context.Context, tx *sql.Tx, proof authentication.Commit) error {
	if proof.Origin == nil {
		return authentication.ErrInvalidLogin
	}
	var state, budget, expiry string
	err := tx.QueryRowContext(ctx, `SELECT state_hash,budget_key,expires_at FROM login_origins WHERE token_hash=? AND user_id=?`, proof.Origin.TokenHash, proof.Account.ID).Scan(&state, &budget, &expiry)
	if errors.Is(err, sql.ErrNoRows) {
		return authentication.ErrInvalidLogin
	}
	if err != nil {
		return fmt.Errorf("recheck login origin: %w", err)
	}
	expires, err := time.Parse(time.RFC3339, expiry)
	if err != nil {
		return fmt.Errorf("parse current login origin expiry: %w", err)
	}
	if state != authentication.CredentialState(proof.Account) || budget != proof.Origin.BudgetKey || !expires.After(proof.Now) {
		return authentication.ErrInvalidLogin
	}
	return nil
}

func requireOne(result sql.Result, err error) error {
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return authentication.ErrInvalidLogin
	}
	return nil
}
