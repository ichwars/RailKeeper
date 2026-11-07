package application_test

import (
	"errors"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/infrastructure"
)

func TestTOTPLoginConsumptionSurvivesDatabaseRestart(t *testing.T) {
	db := testDB(t)
	if err := application.NewSetupService(db).CreateAdmin(t.Context(), application.CreateAdminInput{
		Username: "admin", Email: "admin@example.test", Password: "very-secure-password",
	}); err != nil {
		t.Fatal(err)
	}
	const secret = "JBSWY3DPEHPK3PXP"
	if _, err := db.ExecContext(t.Context(), `UPDATE users SET two_factor_enabled=1,two_factor_secret=?`, secret); err != nil {
		t.Fatal(err)
	}
	now := time.Unix(1700000000, 0)
	auth := application.NewAuthService(db, application.WithAuthClock(func() time.Time { return now }))
	input := application.LoginInput{Username: "admin", Password: "very-secure-password", TwoFactorCode: testTOTPCode(t, secret, now)}
	if _, err := auth.Login(t.Context(), input); err != nil {
		t.Fatal(err)
	}
	var sequence int
	var name, filename string
	if err := db.QueryRowContext(t.Context(), `PRAGMA database_list`).Scan(&sequence, &name, &filename); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := infrastructure.OpenSQLite(filepath.Dir(filename))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = reopened.Close() })
	auth = application.NewAuthService(reopened, application.WithAuthClock(func() time.Time { return now }))
	if _, err := auth.Login(t.Context(), input); !errors.Is(err, application.ErrInvalidLogin) {
		t.Fatalf("consumed step accepted after reopening: %v", err)
	}
	now = now.Add(90 * time.Second)
	input.TwoFactorCode = testTOTPCode(t, secret, now)
	if _, err := auth.Login(t.Context(), input); err != nil {
		t.Fatalf("fresh step must still work: %v", err)
	}
}

func TestConcurrentTOTPLoginConsumesOneStepOnce(t *testing.T) {
	db := testDB(t)
	if err := application.NewSetupService(db).CreateAdmin(t.Context(), application.CreateAdminInput{
		Username: "admin", Email: "admin@example.test", Password: "very-secure-password",
	}); err != nil {
		t.Fatal(err)
	}
	const secret = "JBSWY3DPEHPK3PXP"
	if _, err := db.ExecContext(t.Context(), `UPDATE users SET two_factor_enabled=1,two_factor_secret=?`, secret); err != nil {
		t.Fatal(err)
	}
	now := time.Unix(1700000000, 0)
	auth := application.NewAuthService(db, application.WithAuthClock(func() time.Time { return now }))
	input := application.LoginInput{Username: "admin", Password: "very-secure-password", TwoFactorCode: testTOTPCode(t, secret, now)}
	var successes atomic.Int32
	var workers sync.WaitGroup
	start := make(chan struct{})
	for range 12 {
		workers.Go(func() {
			<-start
			_, err := auth.Login(t.Context(), input)
			if err == nil {
				successes.Add(1)
			} else if !errors.Is(err, application.ErrInvalidLogin) && !errors.Is(err, application.ErrLoginThrottled) {
				t.Errorf("unexpected concurrent result: %v", err)
			}
		})
	}
	close(start)
	workers.Wait()
	var sessions int
	if err := db.QueryRowContext(t.Context(), `SELECT COUNT(*) FROM sessions`).Scan(&sessions); err != nil ||
		successes.Load() != 1 || sessions != 1 {
		t.Fatalf("one code must create exactly one session: successes=%d sessions=%d err=%v", successes.Load(), sessions, err)
	}
}

func TestFailedLoginTransactionDoesNotConsumeCodeOrCreateSession(t *testing.T) {
	db := testDB(t)
	if err := application.NewSetupService(db).CreateAdmin(t.Context(), application.CreateAdminInput{
		Username: "admin", Email: "admin@example.test", Password: "very-secure-password",
	}); err != nil {
		t.Fatal(err)
	}
	const secret = "JBSWY3DPEHPK3PXP"
	oldHash := nondefaultPasswordHashForTest("very-secure-password")
	if _, err := db.ExecContext(t.Context(), `UPDATE users SET two_factor_enabled=1,two_factor_secret=?,password_hash=?`, secret, oldHash); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(t.Context(), `CREATE TRIGGER fail_login_audit BEFORE INSERT ON audit_logs
WHEN NEW.action='Login' BEGIN SELECT RAISE(ABORT,'synthetic audit failure'); END`); err != nil {
		t.Fatal(err)
	}
	now := time.Unix(1700000000, 0)
	auth := application.NewAuthService(db, application.WithAuthClock(func() time.Time { return now }))
	input := application.LoginInput{Username: "admin", Password: "very-secure-password", TwoFactorCode: testTOTPCode(t, secret, now)}
	if _, err := auth.Login(t.Context(), input); err == nil {
		t.Fatal("audit failure must not return a valid session")
	}
	var storedHash string
	if err := db.QueryRowContext(t.Context(), `SELECT password_hash FROM users`).Scan(&storedHash); err != nil || storedHash != oldHash {
		t.Fatalf("failed transaction changed the password hash: %v", err)
	}
	var consumed *int64
	var sessions int
	if err := db.QueryRowContext(t.Context(), `SELECT two_factor_last_step FROM users`).Scan(&consumed); err != nil || consumed != nil {
		t.Fatalf("failed transaction consumed a code: %v", err)
	}
	if err := db.QueryRowContext(t.Context(), `SELECT COUNT(*) FROM sessions`).Scan(&sessions); err != nil || sessions != 0 {
		t.Fatalf("failed transaction retained a session: %v", err)
	}
	if _, err := db.ExecContext(t.Context(), `DROP TRIGGER fail_login_audit`); err != nil {
		t.Fatal(err)
	}
	if _, err := auth.Login(t.Context(), input); err != nil {
		t.Fatalf("the rolled-back code must remain usable: %v", err)
	}
}
