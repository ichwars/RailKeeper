package application

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type setupCodeProof struct{ hash string }

// PrepareAuthorization is called only by the local startup composition root.
// Existing accounts leave the setup authorization unchanged.
func (s *SetupService) PrepareAuthorization(ctx context.Context, code string) (bool, error) {
	if len(code) < 32 || len(code) > 128 || strings.TrimSpace(code) != code {
		return false, ErrInvalidSetupCode
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin setup authorization: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	var users int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&users); err != nil {
		return false, fmt.Errorf("check setup authorization need: %w", err)
	}
	if users != 0 {
		return false, nil
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO setup_authorizations(id,token_hash,created_at)
VALUES(1,?,?) ON CONFLICT(id) DO UPDATE SET token_hash=excluded.token_hash,
created_at=excluded.created_at,consumed_at=NULL`, hashToken(code), time.Now().UTC().Format(time.RFC3339)); err != nil {
		return false, fmt.Errorf("prepare setup authorization: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit setup authorization: %w", err)
	}
	return true, nil
}

// CompleteSetup is the public HTTP setup use case. A code is never returned here.
func (s *SetupService) CompleteSetup(ctx context.Context, input CreateAdminInput) error {
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
	var expected string
	if err := s.db.QueryRowContext(ctx, `SELECT token_hash FROM setup_authorizations
WHERE id=1 AND consumed_at IS NULL`).Scan(&expected); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInvalidSetupCode
		}
		return fmt.Errorf("read setup authorization: %w", err)
	}
	hash := hashToken(code)
	if subtle.ConstantTimeCompare([]byte(hash), []byte(expected)) != 1 {
		return ErrInvalidSetupCode
	}
	return s.createAdmin(ctx, input, &setupCodeProof{hash: hash})
}
