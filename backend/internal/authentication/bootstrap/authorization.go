package bootstrap

import (
	"context"
	"crypto/subtle"
	"fmt"
	"strings"
	"time"

	"railkeeper/backend/internal/credentials"
)

// PrepareAuthorization is called only by local startup; existing accounts retain their authorization.
func (s *Service) PrepareAuthorization(ctx context.Context, code string) (bool, error) {
	if len(code) < 32 || len(code) > 128 || strings.TrimSpace(code) != code {
		return false, ErrInvalidSetupCode
	}
	tx, err := s.repository.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("begin setup authorization: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	users, err := tx.UserCount(ctx)
	if err != nil {
		return false, fmt.Errorf("check setup authorization need: %w", err)
	}
	if users != 0 {
		return false, nil
	}
	if err := tx.ReplaceAuthorization(ctx, credentials.HashToken(code), time.Now().UTC().Format(time.RFC3339)); err != nil {
		return false, fmt.Errorf("prepare setup authorization: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit setup authorization: %w", err)
	}
	return true, nil
}

// CompleteSetup is the public HTTP setup use case. A code is never returned here.

// CompleteSetup requires the locally provisioned one-use authorization.
func (s *Service) CompleteSetup(ctx context.Context, input CreateAdminInput) error {
	required, err := s.SetupRequired(ctx)
	if err != nil {
		return err
	}
	if !required {
		return ErrAlreadySetup
	}
	code := strings.TrimSpace(input.SetupCode)
	if len(code) < 32 || len(code) > 128 {
		return ErrInvalidSetupCode
	}
	expected, found, err := s.repository.PendingAuthorization(ctx)
	if err != nil {
		return fmt.Errorf("read setup authorization: %w", err)
	}
	if !found {
		return ErrInvalidSetupCode
	}

	hash := credentials.HashToken(code)
	if subtle.ConstantTimeCompare([]byte(hash), []byte(expected)) != 1 {
		return ErrInvalidSetupCode
	}
	return s.createAdmin(ctx, input, &setupCodeProof{hash: hash})
}
