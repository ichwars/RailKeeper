package bootstrap

import (
	"context"
	"fmt"
	"strings"
	"time"

	"railkeeper/backend/internal/authentication/management"
	"railkeeper/backend/internal/credentials"
)

func (s *Service) SetupRequired(ctx context.Context) (bool, error) {
	count, err := s.repository.UserCount(ctx)
	if err != nil {
		return false, fmt.Errorf("count users: %w", err)
	}
	return count == 0, nil
}

func (s *Service) CreateAdmin(ctx context.Context, input CreateAdminInput) error {
	return s.createAdmin(ctx, input, nil)
}

func (s *Service) createAdmin(ctx context.Context, input CreateAdminInput, proof *setupCodeProof) error {
	username := strings.TrimSpace(input.Username)
	email := strings.TrimSpace(input.Email)
	if len(username) < 3 || len(input.Password) < 12 || !management.ValidEmail(email) {
		return ErrWeakSetup
	}

	hash, err := credentials.Hash(input.Password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	tx, err := s.repository.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin setup transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	userID := credentials.RandomID()
	now := time.Now().UTC().Format(time.RFC3339)
	if proof != nil {
		users, err := tx.UserCount(ctx)
		if err != nil {
			return fmt.Errorf("check setup completion: %w", err)
		}
		if users != 0 {
			return ErrAlreadySetup
		}
		consumed, err := tx.ConsumeAuthorization(ctx, proof.hash, now)
		if err != nil {
			return fmt.Errorf("consume setup authorization: %w", err)
		}
		if !consumed {
			return ErrInvalidSetupCode
		}
	}
	created, err := tx.InsertFirstAdministrator(ctx, Administrator{ID: userID, Username: username, Email: email, PasswordHash: hash, At: now})
	if err != nil {
		return err
	}
	if created == 0 {
		return ErrAlreadySetup
	}

	if err = tx.AssignAdministratorRoles(ctx, userID); err != nil {
		return fmt.Errorf("assign admin roles: %w", err)
	}
	if err = tx.AuditAdministrator(ctx, AdministratorAudit{ID: credentials.RandomID(), UserID: userID, At: now}); err != nil {
		return fmt.Errorf("write setup audit log: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit setup transaction: %w", err)
	}

	return nil
}
