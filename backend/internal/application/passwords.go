package application

import (
	"context"
	"fmt"

	"railkeeper/backend/internal/credentials"
)

func hashPassword(password string) (string, error) { return credentials.Hash(password) }
func verifyPassword(password, encoded string) bool { return credentials.Verify(password, encoded) }

func (s *AuthService) upgradePassword(ctx context.Context, userID, password, original string) error {
	next, err := credentials.Upgrade(password, original)
	if err != nil {
		return fmt.Errorf("prepare password hash upgrade: %w", err)
	}
	if next == "" {
		return nil
	}
	result, err := s.db.ExecContext(ctx, `UPDATE users SET password_hash=? WHERE id=? AND password_hash=?`, next, userID, original)
	if err != nil {
		return fmt.Errorf("upgrade password hash: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read password hash upgrade: %w", err)
	}
	if count != 1 {
		return ErrInvalidLogin
	}
	return nil
}
