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
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin password reset request: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	now := s.now().UTC().Format(time.RFC3339)
	if request.UserID != "" {
		if _, err := tx.ExecContext(ctx, `UPDATE password_reset_requests SET handled_at=? WHERE user_id=? AND handled_at IS NULL`, now, request.UserID); err != nil {
			return fmt.Errorf("close previous password reset requests: %w", err)
		}
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO password_reset_requests(id,user_id,email,token_hash,created_at,expires_at) VALUES(?,?,?,?,?,?)`, randomID(), nullableString(request.UserID), request.Email, nullableString(request.TokenHash), now, nullableString(request.ExpiresAt)); err != nil {
		return fmt.Errorf("create password reset request: %w", err)
	}
	target := request.UserID
	if target == "" {
		target = request.Email
	}
	if err := s.auditTx(ctx, tx, "", "PasswordResetRequested", "user", target, "{}"); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit password reset request: %w", err)
	}
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
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin password reset: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	var current management.ResetProof
	err = tx.QueryRowContext(ctx, `SELECT id,COALESCE(user_id,''),COALESCE(expires_at,''),COALESCE(used_at,''),COALESCE(handled_at,'') FROM password_reset_requests WHERE id=? AND user_id=? AND token_hash=?`, proof.RequestID, proof.UserID, proof.TokenHash).Scan(&current.ID, &current.UserID, &current.ExpiresAt, &current.UsedAt, &current.HandledAt)
	if errors.Is(err, sql.ErrNoRows) {
		return management.ErrPasswordResetInvalid
	}
	if err != nil {
		return fmt.Errorf("recheck password reset: %w", err)
	}
	now := s.now().UTC()
	if !management.ValidResetProof(current, now) {
		return management.ErrPasswordResetInvalid
	}
	timestamp := now.Format(time.RFC3339)
	result, err := tx.ExecContext(ctx, `UPDATE password_reset_requests SET used_at=?,handled_at=? WHERE id=? AND used_at IS NULL AND handled_at IS NULL`, timestamp, timestamp, proof.RequestID)
	if err != nil {
		return fmt.Errorf("consume password reset: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read reset consumption: %w", err)
	}
	if count != 1 {
		return management.ErrPasswordResetInvalid
	}
	if err := requireOne(tx.ExecContext(ctx, `UPDATE users SET password_hash=? WHERE id=?`, proof.NextHash, proof.UserID)); err != nil {
		return fmt.Errorf("update reset password: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE sessions SET revoked_at=? WHERE user_id=? AND revoked_at IS NULL`, timestamp, proof.UserID); err != nil {
		return fmt.Errorf("revoke reset sessions: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE password_reset_requests SET handled_at=? WHERE user_id=? AND id<>? AND handled_at IS NULL`, timestamp, proof.UserID, proof.RequestID); err != nil {
		return fmt.Errorf("close other password reset requests: %w", err)
	}
	if err := s.auditTx(ctx, tx, proof.UserID, "PasswordResetCompleted", "user", proof.UserID, "{}"); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit password reset: %w", err)
	}
	return nil
}
