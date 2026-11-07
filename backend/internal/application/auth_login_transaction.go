package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type loginProof struct {
	userID, passwordHash, twoFactorSecret string
	accountBudgetKey, upgradedHash        string
	twoFactorEnabled                      bool
	twoFactorStep                         int64
}

func (s *AuthService) completeLogin(ctx context.Context, proof loginProof, result *LoginResult) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin login: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	var currentHash, currentSecret string
	var currentEnabled bool
	var lastStep sql.NullInt64
	err = tx.QueryRowContext(ctx, `SELECT password_hash,COALESCE(two_factor_secret,''),two_factor_enabled,
two_factor_last_step FROM users WHERE id=?`, proof.userID).Scan(&currentHash, &currentSecret, &currentEnabled, &lastStep)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrInvalidLogin
	}
	if err != nil {
		return fmt.Errorf("read current login state: %w", err)
	}
	if currentHash != proof.passwordHash || currentSecret != proof.twoFactorSecret || currentEnabled != proof.twoFactorEnabled {
		return ErrInvalidLogin
	}
	if proof.twoFactorEnabled {
		if lastStep.Valid && proof.twoFactorStep <= lastStep.Int64 {
			return ErrInvalidLogin
		}
		if err := requireAuthMutation(tx.ExecContext(ctx, `UPDATE users SET two_factor_last_step=?
WHERE id=? AND (two_factor_last_step IS NULL OR two_factor_last_step<?)`,
			proof.twoFactorStep, proof.userID, proof.twoFactorStep)); err != nil {
			return fmt.Errorf("consume two factor code: %w", err)
		}
	}
	if proof.upgradedHash != "" {
		if err := requireAuthMutation(tx.ExecContext(ctx, `UPDATE users SET password_hash=? WHERE id=? AND password_hash=?`,
			proof.upgradedHash, proof.userID, proof.passwordHash)); err != nil {
			return fmt.Errorf("upgrade login password hash: %w", err)
		}
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO sessions(id,user_id,token_hash,csrf_token,expires_at,created_at)
VALUES(?,?,?,?,?,?)`, randomID(), proof.userID, hashToken(result.SessionToken), result.CSRFToken,
		result.ExpiresAt.Format(time.RFC3339), s.now().UTC().Format(time.RFC3339)); err != nil {
		return fmt.Errorf("create login session: %w", err)
	}
	if err := s.auditTx(ctx, tx, proof.userID, "Login", "user", proof.userID, "{}"); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM rate_limit_attempts WHERE scope='login-account' AND key=?`,
		proof.accountBudgetKey); err != nil {
		return fmt.Errorf("reset successful login budget: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit login: %w", err)
	}
	return nil
}

func requireAuthMutation(result sql.Result, err error) error {
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return ErrInvalidLogin
	}
	return nil
}
