package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"railkeeper/backend/internal/credentials"
)

var dummyLoginHash = credentials.DummyHash

func (s *AuthService) Login(ctx context.Context, input LoginInput) (*LoginResult, error) {
	username := strings.TrimSpace(input.Username)
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case s.loginWork <- struct{}{}:
		defer func() { <-s.loginWork }()
	default:
		return nil, ErrLoginThrottled
	}
	key := hashToken(username)
	allowed, err := s.loginLimits.Allow(ctx, "login-account", key, 10, 5*time.Minute)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrLoginThrottled
	}

	var userID, storedUsername, passwordHash, twoFactorSecret string
	var twoFactorEnabled bool
	err = s.db.QueryRowContext(
		ctx,
		`SELECT id, username, password_hash, COALESCE(two_factor_secret, ''), two_factor_enabled FROM users WHERE username=?`,
		username,
	).Scan(&userID, &storedUsername, &passwordHash, &twoFactorSecret, &twoFactorEnabled)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("read login account: %w", err)
	}
	found := err == nil
	if !found {
		passwordHash = dummyLoginHash()
	}
	validPassword := verifyPassword(input.Password, passwordHash)
	if !found || !validPassword {
		_ = s.audit(ctx, "", "LoginFailed", "user", username, "{}")
		return nil, ErrInvalidLogin
	}
	if twoFactorEnabled {
		if strings.TrimSpace(input.TwoFactorCode) == "" {
			_ = s.audit(ctx, userID, "LoginTwoFactorRequired", "user", userID, "{}")
			return nil, ErrTwoFactorRequired
		}
		if !verifyTOTPCode(twoFactorSecret, input.TwoFactorCode, time.Now().UTC()) {
			_ = s.audit(ctx, userID, "LoginTwoFactorFailed", "user", userID, "{}")
			return nil, ErrInvalidLogin
		}
	}

	if err := s.upgradePassword(ctx, userID, input.Password, passwordHash); err != nil {
		return nil, err
	}
	if err := s.loginLimits.Reset(ctx, "login-account", key); err != nil {
		return nil, err
	}
	roles, err := s.roles(ctx, userID)
	if err != nil {
		return nil, err
	}

	sessionToken := randomToken()
	csrfToken := randomToken()
	expiresAt := time.Now().UTC().Add(12 * time.Hour)
	now := time.Now().UTC().Format(time.RFC3339)

	if _, err := s.db.ExecContext(
		ctx,
		`INSERT INTO sessions(id, user_id, token_hash, csrf_token, expires_at, created_at)
		 VALUES(?, ?, ?, ?, ?, ?)`,
		randomID(),
		userID,
		hashToken(sessionToken),
		csrfToken,
		expiresAt.Format(time.RFC3339),
		now,
	); err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	if err := s.audit(ctx, userID, "Login", "user", userID, "{}"); err != nil {
		return nil, err
	}

	return &LoginResult{
		SessionToken: sessionToken,
		CSRFToken:    csrfToken,
		ExpiresAt:    expiresAt,
		Session: SessionView{
			Username:         storedUsername,
			Roles:            roles,
			CSRFToken:        csrfToken,
			TwoFactorEnabled: twoFactorEnabled,
		},
	}, nil
}
