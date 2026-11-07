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

var (
	ErrAlreadySetup     = errors.New("setup already completed")
	ErrWeakSetup        = errors.New("username or password is too short")
	ErrInvalidSetupCode = errors.New("invalid or consumed setup code")
)

type SetupService struct {
	db *sql.DB
}

type CreateAdminInput struct {
	Username  string
	Email     string
	Password  string
	SetupCode string
}

func NewSetupService(db *sql.DB) *SetupService {
	return &SetupService{db: db}
}

func (s *SetupService) SetupRequired(ctx context.Context) (bool, error) {
	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&count); err != nil {
		return false, fmt.Errorf("count users: %w", err)
	}
	return count == 0, nil
}

// CreateAdmin is a trusted in-process bootstrap. HTTP setup uses CompleteSetup,
// which additionally requires the locally provisioned one-use authorization.
func (s *SetupService) CreateAdmin(ctx context.Context, input CreateAdminInput) error {
	return s.createAdmin(ctx, input, nil)
}

func (s *SetupService) createAdmin(ctx context.Context, input CreateAdminInput, proof *setupCodeProof) error {
	username := strings.TrimSpace(input.Username)
	email := strings.TrimSpace(input.Email)
	if len(username) < 3 || len(input.Password) < 12 || !isValidEmail(email) {
		return ErrWeakSetup
	}

	hash, err := hashPassword(input.Password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin setup transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	userID := randomID()
	now := time.Now().UTC().Format(time.RFC3339)
	if proof != nil {
		var users int
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&users); err != nil {
			return fmt.Errorf("check setup completion: %w", err)
		}
		if users != 0 {
			return ErrAlreadySetup
		}
		if err := requireAuthMutation(tx.ExecContext(ctx, `UPDATE setup_authorizations SET consumed_at=?
WHERE id=1 AND token_hash=? AND consumed_at IS NULL`, now, proof.hash)); err != nil {
			if errors.Is(err, ErrInvalidLogin) {
				return ErrInvalidSetupCode
			}
			return fmt.Errorf("consume setup authorization: %w", err)
		}
	}
	result, err := tx.ExecContext(
		ctx,
		`INSERT INTO users(id, username, email, password_hash, created_at)
		 SELECT ?, ?, ?, ?, ?
		  WHERE NOT EXISTS (SELECT 1 FROM users)`,
		userID,
		username,
		email,
		hash,
		now,
	)
	if err != nil {
		return fmt.Errorf("insert admin user: %w", err)
	}
	created, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read setup insert result: %w", err)
	}
	if created == 0 {
		return ErrAlreadySetup
	}

	if _, err = tx.ExecContext(
		ctx,
		`INSERT INTO user_roles(user_id, role_id)
		 SELECT ?, id FROM roles WHERE name IN ('Admin', 'Editor', 'Viewer')`,
		userID,
	); err != nil {
		return fmt.Errorf("assign admin roles: %w", err)
	}

	if _, err = tx.ExecContext(
		ctx,
		`INSERT INTO audit_logs(id, actor_user_id, action, target_type, target_id, created_at, details_json)
		 VALUES(?, ?, 'SetupAdminCreated', 'user', ?, ?, '{}')`,
		randomID(),
		userID,
		userID,
		now,
	); err != nil {
		return fmt.Errorf("write setup audit log: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit setup transaction: %w", err)
	}

	return nil
}

func randomID() string {
	return credentials.RandomID()
}
