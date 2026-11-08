package authdata

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"railkeeper/backend/internal/authentication"
	"railkeeper/backend/internal/authentication/management"
)

func (s *Store) PrepareTwoFactor(ctx context.Context, userID string, generateSecret func() (string, error)) (management.TwoFactorState, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return management.TwoFactorState{}, fmt.Errorf("begin two factor preparation: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	var username, existingSecret, enabledAt string
	var enabled bool
	if err := tx.QueryRowContext(
		ctx,
		`SELECT username, COALESCE(two_factor_secret, ''), two_factor_enabled,
COALESCE(two_factor_enabled_at,'') FROM users WHERE id=?`,
		userID,
	).Scan(&username, &existingSecret, &enabled, &enabledAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return management.TwoFactorState{}, authentication.ErrUserNotFound
		}
		return management.TwoFactorState{}, fmt.Errorf("read two factor user: %w", err)
	}
	if enabled {
		return management.TwoFactorState{Enabled: true, Secret: existingSecret, EnabledAt: enabledAt, Username: username}, nil
	}

	secret := existingSecret
	if secret == "" {
		var err error
		secret, err = generateSecret()
		if err != nil {
			return management.TwoFactorState{}, fmt.Errorf("generate two factor secret: %w", err)
		}
		if _, err := tx.ExecContext(
			ctx,
			`UPDATE users SET two_factor_secret=?, two_factor_enabled=0, two_factor_enabled_at=NULL,
two_factor_last_step=NULL WHERE id=? AND two_factor_enabled=0 AND COALESCE(two_factor_secret,'')=''`,
			secret,
			userID,
		); err != nil {
			return management.TwoFactorState{}, fmt.Errorf("store two factor secret: %w", err)
		}
		if err := s.auditTx(ctx, tx, userID, "TwoFactorPrepared", "user", userID, "{}"); err != nil {
			return management.TwoFactorState{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return management.TwoFactorState{}, fmt.Errorf("commit two factor preparation: %w", err)
	}

	return management.TwoFactorState{Enabled: enabled, Secret: secret, EnabledAt: enabledAt, Username: username}, nil
}
func (s *Store) ReadTwoFactor(ctx context.Context, id string) (management.TwoFactorState, error) {
	var state management.TwoFactorState
	err := s.db.QueryRowContext(ctx, `SELECT username,COALESCE(two_factor_secret,''),two_factor_enabled,COALESCE(two_factor_enabled_at,'') FROM users WHERE id=?`, id).Scan(&state.Username, &state.Secret, &state.Enabled, &state.EnabledAt)
	if errors.Is(err, sql.ErrNoRows) {
		return state, authentication.ErrUserNotFound
	}
	if err != nil {
		return state, fmt.Errorf("read two factor status: %w", err)
	}
	return state, nil
}
func (s *Store) ReadCredential(ctx context.Context, id string) (authentication.Account, error) {
	var account authentication.Account
	err := s.db.QueryRowContext(ctx, `SELECT id,username,password_hash,COALESCE(two_factor_secret,''),two_factor_enabled FROM users WHERE id=?`, id).Scan(&account.ID, &account.Username, &account.PasswordHash, &account.TwoFactorSecret, &account.TwoFactorEnabled)
	if errors.Is(err, sql.ErrNoRows) {
		return account, authentication.ErrUserNotFound
	}
	if err != nil {
		return account, fmt.Errorf("read account credential state: %w", err)
	}
	return account, nil
}
