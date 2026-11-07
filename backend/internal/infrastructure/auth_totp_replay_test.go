package infrastructure_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/infrastructure"
)

func TestTOTPNonceMigrationPreservesExistingAuthentication(t *testing.T) {
	previous := filepath.Join(t.TempDir(), "schema77")
	if err := os.Mkdir(previous, 0700); err != nil {
		t.Fatal(err)
	}
	migrations := filepath.Join("..", "..", "migrations")
	copyMigrationsThrough(t, migrations, previous, "0077_decoder_legacy_workspaces.sql")
	db, err := infrastructure.OpenSQLite(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := infrastructure.Migrate(db, previous); err != nil {
		t.Fatal(err)
	}
	if err := infrastructure.SeedRoles(db); err != nil {
		t.Fatal(err)
	}
	if err := application.NewSetupService(db).CreateAdmin(t.Context(), application.CreateAdminInput{
		Username: "existing", Email: "existing@example.test", Password: "very-secure-password",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(t.Context(), `UPDATE users SET two_factor_enabled=1,two_factor_secret='JBSWY3DPEHPK3PXP'`); err != nil {
		t.Fatal(err)
	}
	var id, hash, secret string
	if err := db.QueryRowContext(t.Context(), `SELECT id,password_hash,two_factor_secret FROM users`).Scan(&id, &hash, &secret); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(t.Context(), `INSERT INTO sessions(id,user_id,token_hash,csrf_token,expires_at,created_at)
VALUES('existing-session',?,'existing-token-hash','existing-csrf','2030-01-01T00:00:00Z','2026-11-07T00:00:00Z')`, id); err != nil {
		t.Fatal(err)
	}
	if err := infrastructure.Migrate(db, migrations); err != nil {
		t.Fatal(err)
	}
	var gotID, gotHash, gotSecret string
	var enabled bool
	var nonce sql.NullInt64
	if err := db.QueryRowContext(t.Context(), `SELECT id,password_hash,two_factor_secret,two_factor_enabled,two_factor_last_step
FROM users`).Scan(&gotID, &gotHash, &gotSecret, &enabled, &nonce); err != nil {
		t.Fatal(err)
	}
	if gotID != id || gotHash != hash || gotSecret != secret || !enabled || nonce.Valid {
		t.Fatal("migration changed existing authentication or invented a used step")
	}
	var token, csrf string
	if err := db.QueryRowContext(t.Context(), `SELECT token_hash,csrf_token FROM sessions WHERE id='existing-session'`).Scan(&token, &csrf); err != nil || token != "existing-token-hash" || csrf != "existing-csrf" {
		t.Fatalf("existing session changed: %v", err)
	}
	if _, err := db.ExecContext(t.Context(), `UPDATE users SET two_factor_last_step=-1`); err == nil {
		t.Fatal("negative replay counter bypasses the database constraint")
	}
}
