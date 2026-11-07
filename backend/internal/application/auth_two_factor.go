package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"time"

	"railkeeper/backend/internal/credentials"
)

func (s *AuthService) TwoFactorStatus(ctx context.Context, userID string) (*TwoFactorStatus, error) {
	var username, secret, enabledAt string
	var enabled bool
	if err := s.db.QueryRowContext(
		ctx,
		`SELECT username, COALESCE(two_factor_secret, ''), two_factor_enabled, COALESCE(two_factor_enabled_at, '')
		   FROM users
		  WHERE id=?`,
		userID,
	).Scan(&username, &secret, &enabled, &enabledAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("read two factor status: %w", err)
	}
	return &TwoFactorStatus{
		Enabled:   enabled,
		Prepared:  secret != "",
		EnabledAt: enabledAt,
		Username:  username,
	}, nil
}

func (s *AuthService) PrepareTwoFactor(ctx context.Context, userID string) (*TwoFactorStatus, error) {
	var username, existingSecret string
	var enabled bool
	if err := s.db.QueryRowContext(
		ctx,
		`SELECT username, COALESCE(two_factor_secret, ''), two_factor_enabled FROM users WHERE id=?`,
		userID,
	).Scan(&username, &existingSecret, &enabled); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("read two factor user: %w", err)
	}
	if enabled {
		return s.TwoFactorStatus(ctx, userID)
	}

	secret := existingSecret
	if secret == "" {
		var err error
		secret, err = generateTOTPSecret()
		if err != nil {
			return nil, fmt.Errorf("generate two factor secret: %w", err)
		}
		if _, err := s.db.ExecContext(
			ctx,
			`UPDATE users SET two_factor_secret=?, two_factor_enabled=0, two_factor_enabled_at=NULL WHERE id=?`,
			secret,
			userID,
		); err != nil {
			return nil, fmt.Errorf("store two factor secret: %w", err)
		}
		if err := s.audit(ctx, userID, "TwoFactorPrepared", "user", userID, "{}"); err != nil {
			return nil, err
		}
	}

	return &TwoFactorStatus{
		Enabled:    enabled,
		Prepared:   true,
		Username:   username,
		Secret:     secret,
		OtpauthURL: totpProvisioningURI("RailKeeper", username, secret),
	}, nil
}

func (s *AuthService) EnableTwoFactor(ctx context.Context, userID string, input TwoFactorEnableInput) (*TwoFactorStatus, error) {
	var secret string
	if err := s.db.QueryRowContext(ctx, `SELECT COALESCE(two_factor_secret, '') FROM users WHERE id=?`, userID).Scan(&secret); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("read two factor secret: %w", err)
	}
	if secret == "" || !verifyTOTPCode(secret, input.Code, s.now().UTC()) {
		_ = s.audit(ctx, userID, "TwoFactorEnableFailed", "user", userID, "{}")
		return nil, ErrInvalidLogin
	}
	now := s.now().UTC().Format(time.RFC3339)
	if _, err := s.db.ExecContext(
		ctx,
		`UPDATE users SET two_factor_enabled=1, two_factor_enabled_at=? WHERE id=?`,
		now,
		userID,
	); err != nil {
		return nil, fmt.Errorf("enable two factor: %w", err)
	}
	if err := s.audit(ctx, userID, "TwoFactorEnabled", "user", userID, "{}"); err != nil {
		return nil, err
	}
	return s.TwoFactorStatus(ctx, userID)
}

func (s *AuthService) DisableTwoFactor(ctx context.Context, userID, sessionToken string, input TwoFactorDisableInput) error {
	var passwordHash, secret string
	var enabled bool
	if err := s.db.QueryRowContext(
		ctx,
		`SELECT password_hash, COALESCE(two_factor_secret, ''), two_factor_enabled FROM users WHERE id=?`,
		userID,
	).Scan(&passwordHash, &secret, &enabled); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUserNotFound
		}
		return fmt.Errorf("read two factor user: %w", err)
	}
	if !verifyPassword(input.CurrentPassword, passwordHash) {
		_ = s.audit(ctx, userID, "TwoFactorDisableFailed", "user", userID, "{}")
		return ErrInvalidLogin
	}
	if enabled && !verifyTOTPCode(secret, input.Code, s.now().UTC()) {
		_ = s.audit(ctx, userID, "TwoFactorDisableFailed", "user", userID, "{}")
		return ErrInvalidLogin
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin two factor disable: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	if _, err = tx.ExecContext(
		ctx,
		`UPDATE users SET two_factor_secret=NULL, two_factor_enabled=0, two_factor_enabled_at=NULL WHERE id=?`,
		userID,
	); err != nil {
		return fmt.Errorf("disable two factor: %w", err)
	}
	if _, err = tx.ExecContext(
		ctx,
		`UPDATE sessions SET revoked_at=? WHERE user_id=? AND token_hash<>? AND revoked_at IS NULL`,
		s.now().UTC().Format(time.RFC3339),
		userID,
		hashToken(sessionToken),
	); err != nil {
		return fmt.Errorf("revoke two factor sessions: %w", err)
	}
	if err = s.auditTx(ctx, tx, userID, "TwoFactorDisabled", "user", userID, "{}"); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit two factor disable: %w", err)
	}
	return nil
}

func generateTOTPSecret() (string, error) { return credentials.GenerateTOTPSecret() }
func verifyTOTPCode(secret, code string, now time.Time) bool {
	_, ok := credentials.MatchTOTP(secret, code, now)
	return ok
}

func totpProvisioningURI(issuer, account, secret string) string {
	label := issuer + ":" + account
	query := url.Values{}
	query.Set("secret", secret)
	query.Set("issuer", issuer)
	query.Set("algorithm", "SHA1")
	query.Set("digits", "6")
	query.Set("period", "30")
	return "otpauth://totp/" + url.PathEscape(label) + "?" + query.Encode()
}
