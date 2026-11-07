package authdata_test

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/authentication"
	"railkeeper/backend/internal/credentials"
	"railkeeper/backend/internal/infrastructure"
	"railkeeper/backend/internal/infrastructure/authdata"
)

func TestLoginCommitRejectsConcurrentCredentialChangesWithoutPartialEffects(t *testing.T) {
	for _, column := range []string{"password_hash", "two_factor_secret", "two_factor_enabled"} {
		t.Run(column, func(t *testing.T) {
			db, err := infrastructure.OpenSQLite(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = db.Close() })
			if err := infrastructure.Migrate(db, filepath.Join("..", "..", "..", "migrations")); err != nil {
				t.Fatal(err)
			}
			if err := infrastructure.SeedRoles(db); err != nil {
				t.Fatal(err)
			}
			if err := application.NewSetupService(db).CreateAdmin(t.Context(), application.CreateAdminInput{
				Username: "admin", Email: "admin@example.test", Password: "very-secure-password",
			}); err != nil {
				t.Fatal(err)
			}
			store := authdata.NewStore(db)
			account, found, err := store.FindAccount(t.Context(), "admin")
			if err != nil || !found {
				t.Fatalf("find account: %v", err)
			}
			now := time.Now().UTC()
			attempt := authentication.Attempt{Key: credentials.HashToken("admin"), Limit: 10, Window: 5 * time.Minute, Now: now}
			if allowed, err := store.AllowAttempt(t.Context(), attempt); err != nil || !allowed {
				t.Fatalf("reserve: %v", err)
			}
			value := any("changed")
			if column == "two_factor_enabled" {
				value = true
			}
			if _, err := db.ExecContext(t.Context(), `UPDATE users SET `+column+`=? WHERE id=?`, value, account.ID); err != nil {
				t.Fatal(err)
			}
			proof := authentication.Commit{Account: account, AttemptKey: attempt.Key, UpgradedHash: "must-not-be-written", Now: now,
				Session: authentication.Result{SessionToken: "synthetic-session", CSRFToken: "synthetic-csrf", ExpiresAt: now.Add(time.Hour)}}
			if err := store.CommitLogin(t.Context(), proof); !errors.Is(err, authentication.ErrInvalidLogin) {
				t.Fatalf("stale authentication proof committed: %v", err)
			}
			var sessions, attempts, successes int
			var currentHash string
			if err := db.QueryRowContext(t.Context(), `SELECT (SELECT COUNT(*) FROM sessions),
(SELECT COUNT(*) FROM rate_limit_attempts WHERE scope='login-account'),
(SELECT COUNT(*) FROM audit_logs WHERE action='Login'),password_hash FROM users`).Scan(&sessions, &attempts, &successes, &currentHash); err != nil || sessions != 0 || attempts != 1 || successes != 0 {
				t.Fatalf("stale proof left partial effects: sessions=%d attempts=%d successes=%d err=%v", sessions, attempts, successes, err)
			}
			wantHash := account.PasswordHash
			if column == "password_hash" {
				wantHash = "changed"
			}
			if currentHash != wantHash {
				t.Fatal("stale login overwrote a concurrent credential change")
			}
		})
	}
}
