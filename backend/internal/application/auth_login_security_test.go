package application_test

import (
	"errors"
	"sync"
	"testing"
	"time"

	"railkeeper/backend/internal/application"
)

func TestLoginAccountBudgetSurvivesServiceRestartAndExpires(t *testing.T) {
	db := testDB(t)
	auth := application.NewAuthService(db)
	input := application.LoginInput{Username: "unknown", Password: "RailKeeper-unknown-user"}
	for i := range 10 {
		if _, err := auth.Login(t.Context(), input); !errors.Is(err, application.ErrInvalidLogin) {
			t.Fatalf("attempt %d: %v", i, err)
		}
	}
	auth = application.NewAuthService(db)
	if _, err := auth.Login(t.Context(), input); !errors.Is(err, application.ErrLoginThrottled) {
		t.Fatalf("same account budget must survive restart: %v", err)
	}
	if _, err := auth.Login(t.Context(), application.LoginInput{Username: "another", Password: "wrong"}); !errors.Is(err, application.ErrInvalidLogin) {
		t.Fatalf("another account must retain its own budget: %v", err)
	}
	_, err := db.ExecContext(t.Context(), `UPDATE rate_limit_attempts SET attempted_at=? WHERE scope='login-account'`,
		time.Now().UTC().Add(-6*time.Minute).Format(time.RFC3339Nano))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := auth.Login(t.Context(), input); !errors.Is(err, application.ErrInvalidLogin) {
		t.Fatalf("expired budget must recover automatically: %v", err)
	}
	var accounts int
	if err := db.QueryRowContext(t.Context(), `SELECT COUNT(*) FROM users`).Scan(&accounts); err != nil || accounts != 0 {
		t.Fatalf("dummy password must never create a user: count=%d err=%v", accounts, err)
	}
}

func TestSuccessfulLoginClearsOnlyItsOwnAccountBudget(t *testing.T) {
	db := testDB(t)
	if err := application.NewSetupService(db).CreateAdmin(t.Context(), application.CreateAdminInput{
		Username: "admin", Email: "admin@example.test", Password: "very-secure-password",
	}); err != nil {
		t.Fatal(err)
	}
	auth := application.NewAuthService(db)
	for range 3 {
		_, _ = auth.Login(t.Context(), application.LoginInput{Username: "admin", Password: "wrong"})
	}
	_, _ = auth.Login(t.Context(), application.LoginInput{Username: "other", Password: "wrong"})
	for range 12 {
		if _, err := auth.Login(t.Context(), application.LoginInput{Username: " admin ", Password: "very-secure-password"}); err != nil {
			t.Fatalf("successful logins must not accumulate failed-attempt debt: %v", err)
		}
	}
	var count int
	if err := db.QueryRowContext(t.Context(), `SELECT COUNT(*) FROM rate_limit_attempts WHERE scope='login-account'`).Scan(&count); err != nil || count != 1 {
		t.Fatalf("only the other account's failed attempt should remain: count=%d err=%v", count, err)
	}
}

func TestConcurrentLoginAttemptsDoNotCreateSessionsForUnknownUser(t *testing.T) {
	db := testDB(t)
	auth := application.NewAuthService(db)
	var workers sync.WaitGroup
	for range 12 {
		workers.Go(func() {
			_, err := auth.Login(t.Context(), application.LoginInput{Username: "missing", Password: "RailKeeper-unknown-user"})
			if !errors.Is(err, application.ErrInvalidLogin) && !errors.Is(err, application.ErrLoginThrottled) {
				t.Errorf("unexpected concurrent result: %v", err)
			}
		})
	}
	workers.Wait()
	var count int
	if err := db.QueryRowContext(t.Context(), `SELECT COUNT(*) FROM sessions`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("unknown login created a session: count=%d err=%v", count, err)
	}
}
