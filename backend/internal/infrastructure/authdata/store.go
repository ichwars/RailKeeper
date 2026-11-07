// Package authdata implements login persistence and its atomic SQLite commit.
package authdata

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"railkeeper/backend/internal/authentication"
	"railkeeper/backend/internal/credentials"
	"railkeeper/backend/internal/infrastructure/ratelimit"
)

type Store struct{ db *sql.DB }

func NewStore(db *sql.DB) *Store { return &Store{db: db} }

var _ authentication.Store = (*Store)(nil)

func (s *Store) AllowAttempt(ctx context.Context, attempt authentication.Attempt) (bool, error) {
	return ratelimit.NewStore(s.db).Allow(ctx, ratelimit.Attempt{Scope: authentication.AccountBudgetScope,
		Key: attempt.Key, Limit: attempt.Limit, Window: attempt.Window, Now: attempt.Now})
}

func (s *Store) FindAccount(ctx context.Context, username string) (authentication.Account, bool, error) {
	var account authentication.Account
	err := s.db.QueryRowContext(ctx, `SELECT id,username,password_hash,COALESCE(two_factor_secret,''),two_factor_enabled
FROM users WHERE username=?`, username).Scan(&account.ID, &account.Username, &account.PasswordHash, &account.TwoFactorSecret, &account.TwoFactorEnabled)
	if errors.Is(err, sql.ErrNoRows) {
		return account, false, nil
	}
	if err != nil {
		return account, false, fmt.Errorf("read login account: %w", err)
	}
	return account, true, nil
}

func (s *Store) Roles(ctx context.Context, userID string) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT r.name FROM roles r JOIN user_roles ur ON ur.role_id=r.id
WHERE ur.user_id=? ORDER BY r.name`, userID)
	if err != nil {
		return nil, fmt.Errorf("query roles: %w", err)
	}
	defer func() { _ = rows.Close() }()
	roles := []string{}
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, fmt.Errorf("scan role: %w", err)
		}
		roles = append(roles, role)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate roles: %w", err)
	}
	return roles, nil
}

func (s *Store) RecordFailure(ctx context.Context, event authentication.Audit) error {
	if _, err := s.db.ExecContext(ctx, `INSERT INTO audit_logs(id,actor_user_id,action,target_type,target_id,created_at,details_json)
VALUES(?,?,?,'user',?,?,'{}')`, credentials.RandomID(), event.UserID, event.Action, event.TargetID,
		event.Now.UTC().Format(time.RFC3339)); err != nil {
		return fmt.Errorf("write audit log: %w", err)
	}
	return nil
}
