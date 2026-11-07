package application_test

import (
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"

	"railkeeper/backend/internal/application"
)

const localSetupCode = "synthetic-local-setup-code-for-test"

func TestPublicSetupRequiresLocalCodeAndDoesNotExportIt(t *testing.T) {
	db := testDB(t)
	setup := application.NewSetupService(db)
	if _, err := setup.PrepareAuthorization(t.Context(), localSetupCode); err != nil {
		t.Fatal(err)
	}
	input := application.CreateAdminInput{Username: "admin", Email: "admin@example.test", Password: "very-secure-password"}
	for _, code := range []string{"", "short", strings.Repeat("x", 43)} {
		input.SetupCode = code
		if err := setup.CompleteSetup(t.Context(), input); !errors.Is(err, application.ErrInvalidSetupCode) {
			t.Fatalf("incorrect setup code accepted: %v", err)
		}
	}
	input.SetupCode = localSetupCode
	if err := setup.CompleteSetup(t.Context(), input); err != nil {
		t.Fatal(err)
	}
	if err := application.NewSetupService(db).CompleteSetup(t.Context(), input); !errors.Is(err, application.ErrAlreadySetup) {
		t.Fatalf("used setup code accepted after service restart: %v", err)
	}
	var stored string
	var consumed *string
	if err := db.QueryRowContext(t.Context(), `SELECT token_hash,consumed_at FROM setup_authorizations`).Scan(&stored, &consumed); err != nil || stored == localSetupCode || consumed == nil {
		t.Fatalf("setup code must be hashed and consumed: %v", err)
	}
	backup, err := application.NewBackupService(db, t.TempDir()).Export(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(backup)
	if err != nil || strings.Contains(string(encoded), localSetupCode) || strings.Contains(string(encoded), "setup_authorizations") {
		t.Fatal("application backup leaked local setup authorization")
	}
	if required, err := setup.PrepareAuthorization(t.Context(), strings.Repeat("y", 43)); err != nil || required {
		t.Fatalf("existing installation must skip setup authorization: %v", err)
	}
	var unchanged string
	if err := db.QueryRowContext(t.Context(), `SELECT token_hash FROM setup_authorizations`).Scan(&unchanged); err != nil || unchanged != stored {
		t.Fatal("existing account setup authorization changed")
	}
}

func TestConcurrentPublicSetupConsumesCodeOnce(t *testing.T) {
	db := testDB(t)
	setup := application.NewSetupService(db)
	if _, err := setup.PrepareAuthorization(t.Context(), localSetupCode); err != nil {
		t.Fatal(err)
	}
	start, results := make(chan struct{}), make(chan error, 2)
	var workers sync.WaitGroup
	for range 2 {
		workers.Go(func() {
			<-start
			results <- setup.CompleteSetup(t.Context(), application.CreateAdminInput{
				Username: "admin", Email: "admin@example.test", Password: "very-secure-password", SetupCode: localSetupCode,
			})
		})
	}
	close(start)
	workers.Wait()
	close(results)
	var successes int
	for err := range results {
		if err == nil {
			successes++
		} else if !errors.Is(err, application.ErrAlreadySetup) {
			t.Fatalf("unexpected concurrent setup result: %v", err)
		}
	}
	var users int
	if err := db.QueryRowContext(t.Context(), `SELECT COUNT(*) FROM users`).Scan(&users); err != nil || users != 1 || successes != 1 {
		t.Fatalf("one code created more than one account: users=%d successes=%d err=%v", users, successes, err)
	}
}

func TestFailedPublicSetupRollsBackCodeAccountRolesAndAudit(t *testing.T) {
	db := testDB(t)
	setup := application.NewSetupService(db)
	if _, err := setup.PrepareAuthorization(t.Context(), localSetupCode); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(t.Context(), `CREATE TRIGGER fail_setup_audit BEFORE INSERT ON audit_logs
WHEN NEW.action='SetupAdminCreated' BEGIN SELECT RAISE(ABORT,'synthetic audit failure'); END`); err != nil {
		t.Fatal(err)
	}
	input := application.CreateAdminInput{Username: "admin", Email: "admin@example.test", Password: "very-secure-password", SetupCode: localSetupCode}
	if err := setup.CompleteSetup(t.Context(), input); err == nil {
		t.Fatal("failed audit committed setup")
	}
	var users, roles int
	var consumed *string
	if err := db.QueryRowContext(t.Context(), `SELECT (SELECT COUNT(*) FROM users),
(SELECT COUNT(*) FROM user_roles),consumed_at FROM setup_authorizations`).Scan(&users, &roles, &consumed); err != nil || users != 0 || roles != 0 || consumed != nil {
		t.Fatalf("failed transaction left partial setup: users=%d roles=%d err=%v", users, roles, err)
	}
	if _, err := db.ExecContext(t.Context(), `DROP TRIGGER fail_setup_audit`); err != nil {
		t.Fatal(err)
	}
	if err := setup.CompleteSetup(t.Context(), input); err != nil {
		t.Fatalf("rolled-back code must still work: %v", err)
	}
}
