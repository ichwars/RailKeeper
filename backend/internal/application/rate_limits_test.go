package application_test

import (
	"context"
	"testing"
	"time"

	"railkeeper/backend/internal/application"
)

func TestRateLimitPersistsAttemptsInDatabase(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()

	first := application.NewRateLimitService(db)
	second := application.NewRateLimitService(db)

	allowed, err := first.Allow(ctx, "login", "127.0.0.1", 2, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if !allowed {
		t.Fatal("first attempt should be allowed")
	}

	allowed, err = second.Allow(ctx, "login", "127.0.0.1", 2, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if !allowed {
		t.Fatal("second attempt should be allowed")
	}

	allowed, err = first.Allow(ctx, "login", "127.0.0.1", 2, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if allowed {
		t.Fatal("third attempt should be blocked across service instances")
	}
}

func TestRateLimitCleanupPreservesLongerWindowsInOtherScopes(t *testing.T) {
	db := testDB(t)
	ctx := t.Context()
	oldAttempt := time.Now().UTC().Add(-10 * time.Minute).Format(time.RFC3339Nano)
	for _, scope := range []string{"login", "bug_report_attempt", "bug_report_new"} {
		if _, err := db.ExecContext(ctx, `
INSERT INTO rate_limit_attempts(id, scope, key, attempted_at) VALUES(?, ?, ?, ?)`,
			scope, scope, "user-1", oldAttempt); err != nil {
			t.Fatal(err)
		}
	}
	service := application.NewRateLimitService(db)
	allowed, err := service.Allow(ctx, "login", "user-1", 1, 5*time.Minute)
	if err != nil || !allowed {
		t.Fatalf("expired login attempt should be pruned: allowed=%v err=%v", allowed, err)
	}
	for _, scope := range []string{"bug_report_attempt", "bug_report_new"} {
		allowed, err := service.Allow(ctx, scope, "user-1", 1, time.Hour)
		if err != nil || allowed {
			t.Fatalf("hourly %s limit must survive login cleanup: allowed=%v err=%v", scope, allowed, err)
		}
	}
}
