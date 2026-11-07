package ratelimit

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"railkeeper/backend/internal/credentials"
)

type Store struct{ db *sql.DB }

func NewStore(db *sql.DB) *Store { return &Store{db: db} }

type Attempt struct {
	Scope, Key string
	Limit      int
	Window     time.Duration
	Now        time.Time
}

func (s *Store) Reset(ctx context.Context, scope, key string) error {
	if s == nil || s.db == nil {
		return nil
	}
	if _, err := s.db.ExecContext(ctx, `DELETE FROM rate_limit_attempts WHERE scope=? AND key=?`, scope, key); err != nil {
		return fmt.Errorf("reset rate limit: %w", err)
	}
	return nil
}

func (s *Store) Allow(ctx context.Context, attempt Attempt) (bool, error) {
	if s == nil || s.db == nil || attempt.Limit <= 0 {
		return true, nil
	}
	now := attempt.Now.UTC()
	cutoff := now.Add(-attempt.Window).Format(time.RFC3339Nano)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin rate limit: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `DELETE FROM rate_limit_attempts WHERE scope=? AND attempted_at<=?`, attempt.Scope, cutoff); err != nil {
		return false, fmt.Errorf("cleanup rate limits: %w", err)
	}
	var count int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM rate_limit_attempts WHERE scope=? AND key=? AND attempted_at>?`,
		attempt.Scope, attempt.Key, cutoff).Scan(&count); err != nil {
		return false, fmt.Errorf("count rate limits: %w", err)
	}
	if count >= attempt.Limit {
		if err := tx.Commit(); err != nil {
			return false, fmt.Errorf("commit rate limit block: %w", err)
		}
		return false, nil
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO rate_limit_attempts(id,scope,key,attempted_at) VALUES(?,?,?,?)`,
		credentials.RandomID(), attempt.Scope, attempt.Key, now.Format(time.RFC3339Nano)); err != nil {
		return false, fmt.Errorf("record rate limit: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit rate limit: %w", err)
	}
	return true, nil
}
