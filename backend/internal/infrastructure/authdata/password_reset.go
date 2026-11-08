package authdata

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"railkeeper/backend/internal/authentication/management"
)

func (s *Store) FindResetUser(ctx context.Context, email string) (string, error) {
	var userID string
	err := s.db.QueryRowContext(ctx, `SELECT id FROM users WHERE lower(email)=lower(?)`, email).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		userID = ""
	} else if err != nil {
		return "", fmt.Errorf("read reset user: %w", err)
	}

	return userID, nil
}
func (s *Store) CreateResetRequest(ctx context.Context, request management.ResetRequest) error {
	now := s.now().UTC().Format(time.RFC3339)
	if request.UserID != "" {
		if _, err := s.db.ExecContext(ctx, `UPDATE password_reset_requests SET handled_at=? WHERE user_id=? AND handled_at IS NULL`, now, request.UserID); err != nil {
			return fmt.Errorf("close previous password reset requests: %w", err)
		}
	}
	if _, err := s.db.ExecContext(ctx, `INSERT INTO password_reset_requests(id,user_id,email,token_hash,created_at,expires_at) VALUES(?,?,?,?,?,?)`, randomID(), nullableString(request.UserID), request.Email, nullableString(request.TokenHash), now, nullableString(request.ExpiresAt)); err != nil {
		return fmt.Errorf("create password reset request: %w", err)
	}
	target := request.UserID
	if target == "" {
		target = request.Email
	}
	_ = s.audit(ctx, "", "PasswordResetRequested", "user", target, "{}")
	return nil
}
func (s *Store) ReadResetRequest(ctx context.Context, hash string) (management.ResetProof, error) {
	var proof management.ResetProof
	err := s.db.QueryRowContext(ctx, `SELECT id,COALESCE(user_id,''),COALESCE(expires_at,''),COALESCE(used_at,''),COALESCE(handled_at,'') FROM password_reset_requests WHERE token_hash=?`, hash).Scan(&proof.ID, &proof.UserID, &proof.ExpiresAt, &proof.UsedAt, &proof.HandledAt)
	if errors.Is(err, sql.ErrNoRows) {
		return proof, management.ErrPasswordResetInvalid
	}
	if err != nil {
		return proof, fmt.Errorf("read password reset token: %w", err)
	}
	return proof, nil
}
func (s *Store) CommitPasswordReset(ctx context.Context, proof management.ResetCommit) error {
	requestID, userID, nextHash := proof.RequestID, proof.UserID, proof.NextHash

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin password reset: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	now := s.now().UTC().Format(time.RFC3339)
	if _, err = tx.ExecContext(ctx, `UPDATE users SET password_hash=? WHERE id=?`, nextHash, userID); err != nil {
		return fmt.Errorf("update reset password: %w", err)
	}
	if _, err = tx.ExecContext(
		ctx,
		`UPDATE sessions SET revoked_at=? WHERE user_id=? AND revoked_at IS NULL`,
		now,
		userID,
	); err != nil {
		return fmt.Errorf("revoke reset sessions: %w", err)
	}
	if _, err = tx.ExecContext(
		ctx,
		`UPDATE password_reset_requests
		    SET used_at=?, handled_at=?
		  WHERE id=?`,
		now,
		now,
		requestID,
	); err != nil {
		return fmt.Errorf("complete password reset request: %w", err)
	}
	if _, err = tx.ExecContext(
		ctx,
		`UPDATE password_reset_requests
		    SET handled_at=?
		  WHERE user_id=?
		    AND id<>?
		    AND handled_at IS NULL`,
		now,
		userID,
		requestID,
	); err != nil {
		return fmt.Errorf("close other password reset requests: %w", err)
	}
	if err = s.auditTx(ctx, tx, userID, "PasswordResetCompleted", "user", userID, "{}"); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit password reset: %w", err)
	}
	return nil
}
