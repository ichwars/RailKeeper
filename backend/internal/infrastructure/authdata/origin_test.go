package authdata_test

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/credentials"
	"railkeeper/backend/internal/infrastructure"
	"railkeeper/backend/internal/infrastructure/authdata"
)

func originDatabase(t *testing.T) (*sql.DB, *application.AuthService, string) {
	t.Helper()
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
	if err := application.NewSetupService(db).CreateAdmin(t.Context(), application.CreateAdminInput{Username: "admin", Email: "admin@example.test", Password: "very-secure-password"}); err != nil {
		t.Fatal(err)
	}
	var id string
	if err := db.QueryRow(`SELECT id FROM users`).Scan(&id); err != nil {
		t.Fatal(err)
	}
	return db, application.NewAuthService(db), id
}

func TestOriginCommitFailuresRollBackSessionProofAuditAndBudgetReset(t *testing.T) {
	for _, fault := range []struct{ table, operation string }{{"login_origins", "INSERT"}, {"sessions", "INSERT"}, {"audit_logs", "INSERT"}, {"rate_limit_attempts", "DELETE"}} {
		t.Run(fault.table, func(t *testing.T) {
			db, auth, _ := originDatabase(t)
			if _, err := db.Exec(`CREATE TRIGGER fail_origin BEFORE ` + fault.operation + ` ON ` + fault.table + ` BEGIN SELECT RAISE(ABORT,'synthetic fault'); END`); err != nil {
				t.Fatal(err)
			}
			result, err := auth.Login(t.Context(), application.LoginInput{Username: "admin", Password: "very-secure-password"})
			if err == nil || result != nil {
				t.Fatal("partial login acknowledged")
			}
			var origins, sessions, successes, attempts int
			if err := db.QueryRow(`SELECT (SELECT COUNT(*) FROM login_origins),(SELECT COUNT(*) FROM sessions),
(SELECT COUNT(*) FROM audit_logs WHERE action='Login'),(SELECT COUNT(*) FROM rate_limit_attempts WHERE scope='login-account')`).Scan(&origins, &sessions, &successes, &attempts); err != nil || origins != 0 || sessions != 0 || successes != 0 || attempts != 1 {
				t.Fatal("partial login survived rollback", origins, sessions, successes, attempts, err)
			}
		})
	}
}

func TestAdminRecoveryRollsBackBudgetsAndProofRevocationWhenAuditFails(t *testing.T) {
	db, auth, id := originDatabase(t)
	login, err := auth.Login(t.Context(), application.LoginInput{Username: "admin", Password: "very-secure-password"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := auth.Login(t.Context(), application.LoginInput{Username: "admin", Password: "wrong", OriginToken: login.OriginToken}); !errors.Is(err, application.ErrInvalidLogin) {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TRIGGER fail_recovery BEFORE INSERT ON audit_logs WHEN NEW.action='LoginBudgetReset' BEGIN SELECT RAISE(ABORT,'synthetic fault'); END`); err != nil {
		t.Fatal(err)
	}
	if err := auth.ResetLoginBudget(t.Context(), id, id); err == nil {
		t.Fatal("failed recovery acknowledged")
	}
	var origins, attempts int
	if err := db.QueryRow(`SELECT (SELECT COUNT(*) FROM login_origins),(SELECT COUNT(*) FROM rate_limit_attempts WHERE scope='login-origin')`).Scan(&origins, &attempts); err != nil || origins != 1 || attempts != 1 {
		t.Fatal("recovery left partial effects", origins, attempts, err)
	}
}

func TestOriginStorageIsBoundedPrivateAndExcludedFromApplicationRestore(t *testing.T) {
	db, auth, id := originDatabase(t)
	var last *application.LoginResult
	for range 12 {
		var err error
		last, err = auth.Login(t.Context(), application.LoginInput{Username: "admin", Password: "very-secure-password"})
		if err != nil {
			t.Fatal(err)
		}
	}
	var count int
	var hash string
	if err := db.QueryRow(`SELECT COUNT(*),MAX(token_hash) FROM login_origins`).Scan(&count, &hash); err != nil || count != 10 || hash == last.OriginToken {
		t.Fatal("origin storage grew or retained raw token", count, err)
	}
	store := authdata.NewStore(db)
	if _, _, found, err := store.FindOrigin(t.Context(), "admin", credentials.HashToken(last.OriginToken)); err != nil || !found {
		t.Fatal("newest proof was evicted", err)
	}
	if _, _, found, err := store.FindOrigin(t.Context(), "different-user", credentials.HashToken(last.OriginToken)); err != nil || found {
		t.Fatal("proof granted a different account", err)
	}
	backup := application.NewBackupService(db, t.TempDir())
	doc, err := backup.Export(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if _, found := doc.Tables["login_origins"]; found {
		t.Fatal("proofs exported")
	}
	if _, err := backup.Import(t.Context(), doc); err != nil {
		t.Fatal("app restore failed", err)
	}
	if _, _, found, err := store.FindOrigin(t.Context(), "admin", credentials.HashToken(last.OriginToken)); err != nil || !found {
		t.Fatal("app restore overwrote target authentication", err)
	}
	if _, err := db.Exec(`DELETE FROM user_roles WHERE user_id=?`, id); err != nil {
		t.Fatal(err)
	}
	if _, _, found, err := store.FindOrigin(t.Context(), "admin", credentials.HashToken(last.OriginToken)); err != nil || found {
		t.Fatal("role change did not revoke recognition", err)
	}
}

func TestOriginLookupAndRecoveryStorageFailuresRemainErrors(t *testing.T) {
	db, _, id := originDatabase(t)
	store := authdata.NewStore(db)
	if _, _, found, err := store.FindOrigin(t.Context(), "admin", "missing"); err != nil || found {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO login_origins VALUES('synthetic',?,'budget','state','invalid','2026-01-01T00:00:00Z')`, id); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := store.FindOrigin(t.Context(), "admin", "synthetic"); err == nil {
		t.Fatal("invalid stored expiry accepted")
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := store.FindOrigin(t.Context(), "admin", "synthetic"); err == nil {
		t.Fatal("closed lookup succeeded")
	}
	if err := store.ResetLoginBudget(t.Context(), id, id, time.Now()); err == nil {
		t.Fatal("closed recovery succeeded")
	}
}
