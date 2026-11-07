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
	var twoFactorStep int64
	if twoFactorEnabled {
		if strings.TrimSpace(input.TwoFactorCode) == "" {
			_ = s.audit(ctx, userID, "LoginTwoFactorRequired", "user", userID, "{}")
			return nil, ErrTwoFactorRequired
		}
		var validCode bool
		twoFactorStep, validCode = credentials.MatchTOTP(twoFactorSecret, input.TwoFactorCode, s.now().UTC())
		if !validCode {
			_ = s.audit(ctx, userID, "LoginTwoFactorFailed", "user", userID, "{}")
			return nil, ErrInvalidLogin
		}
	}

	upgradedHash, err := credentials.Upgrade(input.Password, passwordHash)
	if err != nil {
		return nil, fmt.Errorf("prepare password upgrade: %w", err)
	}
	roles, err := s.roles(ctx, userID)
	if err != nil {
		return nil, err
	}

	sessionToken := randomToken()
	csrfToken := randomToken()
	expiresAt := s.now().UTC().Add(12 * time.Hour)
	result := &LoginResult{
		SessionToken: sessionToken,
		CSRFToken:    csrfToken,
		ExpiresAt:    expiresAt,
		Session: SessionView{
			Username:         storedUsername,
			Roles:            roles,
			CSRFToken:        csrfToken,
			TwoFactorEnabled: twoFactorEnabled,
		},
	}
	proof := loginProof{userID: userID, passwordHash: passwordHash, twoFactorSecret: twoFactorSecret,
		accountBudgetKey: key, upgradedHash: upgradedHash, twoFactorEnabled: twoFactorEnabled, twoFactorStep: twoFactorStep}
	if err := s.completeLogin(ctx, proof, result); err != nil {
		if errors.Is(err, ErrInvalidLogin) {
			_ = s.audit(ctx, userID, "LoginFailed", "user", userID, "{}")
		}
		return nil, err
	}
	return result, nil
}
