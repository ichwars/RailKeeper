package application_test

import (
	"context"
	"testing"
	"time"

	"railkeeper2/backend/internal/application"
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
