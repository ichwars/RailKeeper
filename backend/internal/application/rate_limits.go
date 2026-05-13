package application

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type RateLimitService struct {
	db *sql.DB
}

func NewRateLimitService(db *sql.DB) *RateLimitService {
	return &RateLimitService{db: db}
}

func (s *RateLimitService) Allow(ctx context.Context, scope, key string, limit int, window time.Duration) (allowed bool, err error) {
	if s == nil || s.db == nil || limit <= 0 {
		return true, nil
	}

	now := time.Now().UTC()
	cutoff := now.Add(-window).Format(time.RFC3339Nano)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin rate limit: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(ctx, `DELETE FROM rate_limit_attempts WHERE attempted_at <= ?`, cutoff); err != nil {
		return false, fmt.Errorf("cleanup rate limits: %w", err)
	}

	var count int
	if err = tx.QueryRowContext(
		ctx,
		`SELECT COUNT(*) FROM rate_limit_attempts WHERE scope=? AND key=? AND attempted_at > ?`,
		scope,
		key,
		cutoff,
	).Scan(&count); err != nil {
		return false, fmt.Errorf("count rate limits: %w", err)
	}

	if count >= limit {
		if err = tx.Commit(); err != nil {
			return false, fmt.Errorf("commit rate limit block: %w", err)
		}
		return false, nil
	}

	if _, err = tx.ExecContext(
		ctx,
		`INSERT INTO rate_limit_attempts(id, scope, key, attempted_at) VALUES(?, ?, ?, ?)`,
		randomID(),
		scope,
		key,
		now.Format(time.RFC3339Nano),
	); err != nil {
		return false, fmt.Errorf("record rate limit: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return false, fmt.Errorf("commit rate limit: %w", err)
	}
	return true, nil
}
