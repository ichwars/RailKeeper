package application

import (
	"context"
	"database/sql"
	"time"

	"railkeeper/backend/internal/infrastructure/ratelimit"
)

type RateLimitService struct {
	db  *sql.DB
	now func() time.Time
}

func NewRateLimitService(db *sql.DB) *RateLimitService {
	return &RateLimitService{db: db, now: time.Now}
}

func (s *RateLimitService) Reset(ctx context.Context, scope, key string) error {
	if s == nil {
		return nil
	}
	return ratelimit.NewStore(s.db).Reset(ctx, scope, key)
}

func (s *RateLimitService) Allow(ctx context.Context, scope, key string, limit int, window time.Duration) (bool, error) {
	if s == nil || s.db == nil || limit <= 0 {
		return true, nil
	}
	return ratelimit.NewStore(s.db).Allow(ctx, ratelimit.Attempt{Scope: scope, Key: key, Limit: limit, Window: window, Now: s.now()})
}
