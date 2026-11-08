package application_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestSetupPortFailuresPreserveNonceAndAllTablesForSameCodeRetry(t *testing.T) {
	for _, stage := range []struct{ name, sql, context string }{
		{"roles", `CREATE TRIGGER fail_setup_completion BEFORE INSERT ON user_roles
BEGIN SELECT RAISE(FAIL,'synthetic setup roles failure'); END`, "assign admin roles"},
		{"commit", `CREATE TABLE setup_commit_guard(id TEXT PRIMARY KEY,user_id TEXT REFERENCES users(id) DEFERRABLE INITIALLY DEFERRED);
CREATE TRIGGER fail_setup_completion AFTER INSERT ON audit_logs WHEN NEW.action='SetupAdminCreated'
BEGIN INSERT INTO setup_commit_guard(id,user_id)VALUES('guard','missing-user'); END`, "commit setup transaction"},
	} {
		t.Run(stage.name, func(t *testing.T) {
			db := testDB(t)
			db.SetMaxOpenConns(1)
			service := application.NewSetupService(db)
			if required, err := service.PrepareAuthorization(t.Context(), localSetupCode); err != nil || !required {
				t.Fatal("authorization not prepared", err)
			}
			if _, err := db.Exec(stage.sql); err != nil {
				t.Fatal(err)
			}
			before := databaseRowFingerprints(t, db)
			input := application.CreateAdminInput{Username: "admin", Email: "admin@example.test", Password: "very-secure-password", SetupCode: localSetupCode}
			if err := service.CompleteSetup(t.Context(), input); err == nil || !strings.Contains(err.Error(), stage.context) {
				t.Fatal("intended setup failure was not observed", err)
			}
			if !reflect.DeepEqual(before, databaseRowFingerprints(t, db)) || db.Stats().InUse != 0 {
				t.Fatal("failed setup changed authorization, users, roles, audit or other tables")
			}
			if _, err := db.Exec(`DROP TRIGGER fail_setup_completion`); err != nil {
				t.Fatal(err)
			}
			if err := service.CompleteSetup(t.Context(), input); err != nil {
				t.Fatal("same-code retry failed", err)
			}
			var users, roles, audits int
			var consumed *string
			if err := db.QueryRow(`SELECT (SELECT COUNT(*) FROM users),(SELECT COUNT(*) FROM user_roles),
(SELECT COUNT(*) FROM audit_logs WHERE action='SetupAdminCreated'),consumed_at FROM setup_authorizations`).Scan(&users, &roles, &audits, &consumed); err != nil || users != 1 || roles != 3 || audits != 1 || consumed == nil {
				t.Fatal("retry did not commit exactly one complete setup", err)
			}
			if err := service.CompleteSetup(t.Context(), input); !errors.Is(err, application.ErrAlreadySetup) {
				t.Fatal("setup replay accepted", err)
			}
		})
	}
}

func TestSetupAuthorizationRefreshFailureKeepsPreviousProofAndExistingAccounts(t *testing.T) {
	db := testDB(t)
	db.SetMaxOpenConns(1)
	service := application.NewSetupService(db)
	input := application.CreateAdminInput{Username: "admin", Email: "admin@example.test", Password: "very-secure-password", SetupCode: localSetupCode}
	if err := service.CompleteSetup(t.Context(), input); !errors.Is(err, application.ErrInvalidSetupCode) {
		t.Fatal("unprovisioned setup authorization accepted", err)
	}
	before := databaseRowFingerprints(t, db)
	for _, invalid := range []string{"", strings.Repeat("x", 31), strings.Repeat("x", 129), " " + localSetupCode} {
		if required, err := service.PrepareAuthorization(t.Context(), invalid); required || !errors.Is(err, application.ErrInvalidSetupCode) {
			t.Fatal("invalid local code accepted", err)
		}
	}
	if !reflect.DeepEqual(before, databaseRowFingerprints(t, db)) {
		t.Fatal("invalid code provision changed rows")
	}
	if _, err := service.PrepareAuthorization(t.Context(), localSetupCode); err != nil {
		t.Fatal(err)
	}
	before = databaseRowFingerprints(t, db)
	if _, err := db.Exec(`CREATE TRIGGER fail_authorization_refresh BEFORE UPDATE ON setup_authorizations
BEGIN SELECT RAISE(FAIL,'synthetic authorization refresh failure'); END`); err != nil {
		t.Fatal(err)
	}
	const replacement = "synthetic-replacement-local-code-for-test"
	if required, err := service.PrepareAuthorization(t.Context(), replacement); required || err == nil || !strings.Contains(err.Error(), "authorization refresh failure") {
		t.Fatal("authorization refresh failure not observed", err)
	}
	if !reflect.DeepEqual(before, databaseRowFingerprints(t, db)) {
		t.Fatal("failed refresh lost previous authorization or changed other rows")
	}
	if _, err := db.Exec(`DROP TRIGGER fail_authorization_refresh`); err != nil {
		t.Fatal(err)
	}
	if err := service.CompleteSetup(t.Context(), input); err != nil {
		t.Fatal("previous local proof no longer works", err)
	}
	before = databaseRowFingerprints(t, db)
	if required, err := service.PrepareAuthorization(t.Context(), replacement); err != nil || required {
		t.Fatal("existing account was reopened for setup", err)
	}
	if !reflect.DeepEqual(before, databaseRowFingerprints(t, db)) {
		t.Fatal("existing account or consumed authorization was changed")
	}
}
