package authentication

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"railkeeper/backend/internal/credentials"
)

type Service struct {
	store     Store
	passwords PasswordPolicy
	now       func() time.Time
	work      chan struct{}
}

type Option func(*Service)

func WithClock(now func() time.Time) Option {
	return func(service *Service) {
		if now != nil {
			service.now = now
		}
	}
}

func WithPasswords(policy PasswordPolicy) Option {
	return func(service *Service) {
		if policy != nil {
			service.passwords = policy
		}
	}
}

func NewService(store Store, options ...Option) *Service {
	service := &Service{store: store, passwords: passwordPolicy{}, now: time.Now, work: make(chan struct{}, 2)}
	for _, option := range options {
		option(service)
	}
	return service
}

func (s *Service) Login(ctx context.Context, input Input) (*Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case s.work <- struct{}{}:
		defer func() { <-s.work }()
	default:
		return nil, ErrLoginThrottled
	}
	username := strings.TrimSpace(input.Username)
	key := credentials.HashToken(username)
	allowed, err := s.store.AllowAttempt(ctx, Attempt{Key: key, Now: s.now().UTC(), Limit: 10, Window: 5 * time.Minute})
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrLoginThrottled
	}
	account, found, err := s.store.FindAccount(ctx, username)
	if err != nil {
		return nil, err
	}
	passwordHash := account.PasswordHash
	if !found {
		passwordHash = s.passwords.DummyHash()
	}
	validPassword := s.passwords.Verify(input.Password, passwordHash)
	if !found || !validPassword {
		s.failure(ctx, "", "LoginFailed", username)
		return nil, ErrInvalidLogin
	}
	step, err := s.verifyTwoFactor(ctx, account, input.TwoFactorCode)
	if err != nil {
		return nil, err
	}
	upgraded, err := s.passwords.Upgrade(input.Password, passwordHash)
	if err != nil {
		return nil, fmt.Errorf("prepare password upgrade: %w", err)
	}
	roles, err := s.store.Roles(ctx, account.ID)
	if err != nil {
		return nil, err
	}
	result := &Result{SessionToken: credentials.RandomToken(), CSRFToken: credentials.RandomToken(),
		Username: account.Username, ExpiresAt: s.now().UTC().Add(12 * time.Hour), Roles: roles, TwoFactorEnabled: account.TwoFactorEnabled}
	proof := Commit{Account: account, AttemptKey: key, UpgradedHash: upgraded, TwoFactorStep: step, Session: *result, Now: s.now().UTC()}
	if err := s.store.CommitLogin(ctx, proof); err != nil {
		if errors.Is(err, ErrInvalidLogin) {
			s.failure(ctx, account.ID, "LoginFailed", account.ID)
		}
		return nil, err
	}
	return result, nil
}

func (s *Service) verifyTwoFactor(ctx context.Context, account Account, code string) (int64, error) {
	if !account.TwoFactorEnabled {
		return 0, nil
	}
	if strings.TrimSpace(code) == "" {
		s.failure(ctx, account.ID, "LoginTwoFactorRequired", account.ID)
		return 0, ErrTwoFactorRequired
	}
	step, valid := credentials.MatchTOTP(account.TwoFactorSecret, code, s.now().UTC())
	if !valid {
		s.failure(ctx, account.ID, "LoginTwoFactorFailed", account.ID)
		return 0, ErrInvalidLogin
	}
	return step, nil
}

func (s *Service) failure(ctx context.Context, userID, action, target string) {
	_ = s.store.RecordFailure(ctx, Audit{UserID: userID, Action: action, TargetID: target, Now: s.now().UTC()})
}

type passwordPolicy struct{}

func (passwordPolicy) Verify(password, encoded string) bool {
	return credentials.Verify(password, encoded)
}
func (passwordPolicy) Upgrade(password, encoded string) (string, error) {
	return credentials.Upgrade(password, encoded)
}
func (passwordPolicy) DummyHash() string { return credentials.DummyHash() }
