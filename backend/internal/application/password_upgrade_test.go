package application_test

import (
	"encoding/base64"
	"errors"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/argon2"
	"railkeeper/backend/internal/application"
)

func nondefaultPasswordHashForTest(password string) string {
	salt := []byte("test-legacy-salt")
	key := argon2.IDKey([]byte(password), salt, 2, 32, 1, 24)
	return "$argon2id$v=19$m=32,t=2,p=1$" + base64.RawStdEncoding.EncodeToString(salt) + "$" +
		base64.RawStdEncoding.EncodeToString(key)
}

func TestPasswordHashUpgradesOnlyAfterCompleteAuthentication(t *testing.T) {
	db := testDB(t)
	if err := application.NewSetupService(db).CreateAdmin(t.Context(), application.CreateAdminInput{
		Username: "admin", Email: "admin@example.test", Password: "very-secure-password",
	}); err != nil {
		t.Fatal(err)
	}
	old := nondefaultPasswordHashForTest("very-secure-password")
	const secret = "JBSWY3DPEHPK3PXP"
	if _, err := db.ExecContext(t.Context(), `UPDATE users SET password_hash=?,two_factor_enabled=1,two_factor_secret=?`, old, secret); err != nil {
		t.Fatal(err)
	}
	auth := application.NewAuthService(db)
	for _, input := range []application.LoginInput{
		{Username: "admin", Password: "incorrect"},
		{Username: "admin", Password: "very-secure-password"},
	} {
		_, err := auth.Login(t.Context(), input)
		if !errors.Is(err, application.ErrInvalidLogin) && !errors.Is(err, application.ErrTwoFactorRequired) {
			t.Fatalf("incomplete login result: %v", err)
		}
		var stored string
		if err := db.QueryRowContext(t.Context(), `SELECT password_hash FROM users`).Scan(&stored); err != nil || stored != old {
			t.Fatalf("incomplete authentication changed the password: %v", err)
		}
	}
	if _, err := auth.Login(t.Context(), application.LoginInput{Username: "admin", Password: "very-secure-password",
		TwoFactorCode: testTOTPCode(t, secret, time.Now().UTC())}); err != nil {
		t.Fatalf("complete legacy login must remain supported: %v", err)
	}
	var stored string
	if err := db.QueryRowContext(t.Context(), `SELECT password_hash FROM users`).Scan(&stored); err != nil ||
		!strings.HasPrefix(stored, "$argon2id$v=19$m=65536,t=2,p=4$") {
		t.Fatalf("upgrade must retain the stronger iteration count: %v", err)
	}
}
