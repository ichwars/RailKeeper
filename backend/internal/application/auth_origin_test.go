package application_test

import (
	"errors"
	"testing"
	"time"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/authentication"
	"railkeeper/backend/internal/credentials"
)

func TestForeignFloodCannotSpendRecognizedBrowserBudget(t *testing.T) {
	db := testDB(t)
	auth := application.NewAuthService(db)
	if err := application.NewSetupService(db).CreateAdmin(t.Context(), application.CreateAdminInput{Username: "admin", Email: "admin@example.test", Password: "very-secure-password"}); err != nil {
		t.Fatal(err)
	}
	input := application.LoginInput{Username: "admin", Password: "very-secure-password"}
	first, err := auth.Login(t.Context(), input)
	if err != nil {
		t.Fatal(err)
	}
	if len(first.OriginToken) != 43 {
		t.Fatal("full login did not establish a browser proof")
	}
	for range 10 {
		if _, err := auth.Login(t.Context(), application.LoginInput{Username: "admin", Password: "wrong"}); !errors.Is(err, application.ErrInvalidLogin) {
			t.Fatal(err)
		}
	}
	if _, err := auth.Login(t.Context(), input); !errors.Is(err, application.ErrLoginThrottled) {
		t.Fatal("new browser bypassed account budget", err)
	}
	input.OriginToken = first.OriginToken
	if _, err := auth.Login(t.Context(), input); err != nil {
		t.Fatal("foreign attempts locked recognized browser out", err)
	}
	var attempts int
	if err := db.QueryRow(`SELECT COUNT(*) FROM rate_limit_attempts WHERE scope='login-account'`).Scan(&attempts); err != nil || attempts != 10 {
		t.Fatal("recognized success cleared foreign budget", attempts, err)
	}
	input.Password = "wrong"
	for range 10 {
		if _, err := auth.Login(t.Context(), input); !errors.Is(err, application.ErrInvalidLogin) {
			t.Fatal(err)
		}
	}
	input.Password = "very-secure-password"
	if _, err := auth.Login(t.Context(), input); !errors.Is(err, application.ErrLoginThrottled) {
		t.Fatal("stolen proof bypassed recognized password-attempt limit", err)
	}
	var userID string
	if err := db.QueryRow(`SELECT id FROM users`).Scan(&userID); err != nil {
		t.Fatal(err)
	}
	if err := auth.ResetLoginBudget(t.Context(), "unknown-actor", userID); !errors.Is(err, application.ErrForbidden) {
		t.Fatal("unauthorized recovery", err)
	}
	if err := auth.ResetLoginBudget(t.Context(), userID, "missing"); !errors.Is(err, application.ErrUserNotFound) {
		t.Fatal(err)
	}
	if err := auth.ResetLoginBudget(t.Context(), userID, userID); err != nil {
		t.Fatal(err)
	}
	if next, err := auth.Login(t.Context(), input); err != nil || next.OriginToken == first.OriginToken {
		t.Fatal("admin recovery did not revoke old proof and restore access", err)
	}
	var audited int
	if err := db.QueryRow(`SELECT COUNT(*) FROM audit_logs WHERE action='LoginBudgetReset' AND actor_user_id=? AND target_id=?`, userID, userID).Scan(&audited); err != nil || audited != 1 {
		t.Fatal("missing recovery audit", err)
	}
}

func TestCredentialAndExpiryChangesInvalidateOriginRecognition(t *testing.T) {
	for _, change := range []string{"password_hash", "two_factor_secret", "two_factor_enabled", "username", "expiry"} {
		t.Run(change, func(t *testing.T) {
			db := testDB(t)
			auth := application.NewAuthService(db)
			user, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{Username: "operator", Password: "synthetic-password", Roles: []string{"Viewer"}})
			if err != nil {
				t.Fatal(err)
			}
			login, err := auth.Login(t.Context(), application.LoginInput{Username: user.Username, Password: "synthetic-password"})
			if err != nil {
				t.Fatal(err)
			}
			switch change {
			case "expiry":
				_, err = db.Exec(`UPDATE login_origins SET expires_at=?`, time.Now().Add(-time.Minute).UTC().Format(time.RFC3339))
			case "two_factor_enabled":
				_, err = db.Exec(`UPDATE users SET two_factor_enabled=1`)
			default:
				_, err = db.Exec(`UPDATE users SET ` + change + `='changed'`)
			}
			if err != nil {
				t.Fatal(err)
			}
			// Exhaust the untrusted budget without expensive password work. Old
			// browser proof must no longer select a different budget after mutation.
			username := "operator"
			if change == "username" {
				username = "changed"
			}
			for i := range 10 {
				if _, err := db.Exec(`INSERT INTO rate_limit_attempts(id,scope,key,attempted_at) VALUES(?,?,?,?)`, credentials.RandomID(), authentication.AccountBudgetScope, credentials.HashToken(username), time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
					t.Fatal(i, err)
				}
			}
			if _, err := auth.Login(t.Context(), application.LoginInput{Username: username, Password: "synthetic-password", OriginToken: login.OriginToken}); !errors.Is(err, application.ErrLoginThrottled) {
				t.Fatal("stale proof retained protected access", err)
			}
		})
	}
}
