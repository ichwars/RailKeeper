package application_test

import (
	"errors"
	"testing"
	"time"

	"railkeeper/backend/internal/application"
)

func TestTOTPCodeCannotBeReusedAcrossActivationLoginAndDisable(t *testing.T) {
	db := testDB(t)
	if err := application.NewSetupService(db).CreateAdmin(t.Context(), application.CreateAdminInput{
		Username: "admin", Email: "admin@example.test", Password: "very-secure-password",
	}); err != nil {
		t.Fatal(err)
	}
	var userID string
	if err := db.QueryRowContext(t.Context(), `SELECT id FROM users`).Scan(&userID); err != nil {
		t.Fatal(err)
	}
	now := time.Unix(1700000000, 0)
	auth := application.NewAuthService(db, application.WithAuthClock(func() time.Time { return now }))
	prepared, err := auth.PrepareTwoFactor(t.Context(), userID)
	if err != nil {
		t.Fatal(err)
	}
	code := testTOTPCode(t, prepared.Secret, now)
	if _, err := auth.EnableTwoFactor(t.Context(), userID, application.TwoFactorEnableInput{Code: code}); err != nil {
		t.Fatal(err)
	}
	if _, err := auth.EnableTwoFactor(t.Context(), userID, application.TwoFactorEnableInput{Code: code}); !errors.Is(err, application.ErrInvalidLogin) {
		t.Fatalf("activation code reused: %v", err)
	}
	input := application.LoginInput{Username: "admin", Password: "very-secure-password", TwoFactorCode: code}
	if _, err := auth.Login(t.Context(), input); !errors.Is(err, application.ErrInvalidLogin) {
		t.Fatalf("activation code accepted for login: %v", err)
	}
	now = now.Add(90 * time.Second)
	input.TwoFactorCode = testTOTPCode(t, prepared.Secret, now)
	login, err := auth.Login(t.Context(), input)
	if err != nil {
		t.Fatal(err)
	}
	disable := application.TwoFactorDisableInput{CurrentPassword: input.Password, Code: input.TwoFactorCode}
	if err := auth.DisableTwoFactor(t.Context(), userID, login.SessionToken, disable); !errors.Is(err, application.ErrInvalidLogin) {
		t.Fatalf("login code reused to disable two factor: %v", err)
	}
	now = now.Add(90 * time.Second)
	disable.Code = testTOTPCode(t, prepared.Secret, now)
	if err := auth.DisableTwoFactor(t.Context(), userID, login.SessionToken, disable); err != nil {
		t.Fatal(err)
	}
	if _, err := auth.CurrentSession(t.Context(), login.SessionToken); err != nil {
		t.Fatalf("current session was lost when disabling two factor: %v", err)
	}
}

func TestTOTPActionAuditFailureRollsBackActivationAndDisable(t *testing.T) {
	db := testDB(t)
	if err := application.NewSetupService(db).CreateAdmin(t.Context(), application.CreateAdminInput{
		Username: "admin", Email: "admin@example.test", Password: "very-secure-password",
	}); err != nil {
		t.Fatal(err)
	}
	var userID string
	if err := db.QueryRowContext(t.Context(), `SELECT id FROM users`).Scan(&userID); err != nil {
		t.Fatal(err)
	}
	now := time.Unix(1700000000, 0)
	auth := application.NewAuthService(db, application.WithAuthClock(func() time.Time { return now }))
	prepared, err := auth.PrepareTwoFactor(t.Context(), userID)
	if err != nil {
		t.Fatal(err)
	}
	code := testTOTPCode(t, prepared.Secret, now)
	for _, action := range []string{"TwoFactorEnabled", "TwoFactorDisabled"} {
		if _, err := db.ExecContext(t.Context(), `CREATE TRIGGER fail_action_audit BEFORE INSERT ON audit_logs
WHEN NEW.action='`+action+`' BEGIN SELECT RAISE(ABORT,'synthetic audit failure'); END`); err != nil {
			t.Fatal(err)
		}
		if action == "TwoFactorEnabled" {
			if _, err := auth.EnableTwoFactor(t.Context(), userID, application.TwoFactorEnableInput{Code: code}); err == nil {
				t.Fatal("activation should fail when its audit cannot commit")
			}
		} else {
			now = now.Add(90 * time.Second)
			code = testTOTPCode(t, prepared.Secret, now)
			if err := auth.DisableTwoFactor(t.Context(), userID, "current-session", application.TwoFactorDisableInput{
				CurrentPassword: "very-secure-password", Code: code,
			}); err == nil {
				t.Fatal("disable should fail when its audit cannot commit")
			}
		}
		var enabled bool
		var secret string
		var consumed *int64
		if err := db.QueryRowContext(t.Context(), `SELECT two_factor_enabled,two_factor_secret,two_factor_last_step FROM users`).Scan(&enabled, &secret, &consumed); err != nil {
			t.Fatal(err)
		}
		if secret != prepared.Secret || enabled != (action == "TwoFactorDisabled") || (action == "TwoFactorEnabled" && consumed != nil) {
			t.Fatal("failed action changed persistent two factor state")
		}
		if _, err := db.ExecContext(t.Context(), `DROP TRIGGER fail_action_audit`); err != nil {
			t.Fatal(err)
		}
		if action == "TwoFactorEnabled" {
			if _, err := auth.EnableTwoFactor(t.Context(), userID, application.TwoFactorEnableInput{Code: code}); err != nil {
				t.Fatalf("rolled-back activation code must still work: %v", err)
			}
		} else if err := auth.DisableTwoFactor(t.Context(), userID, "current-session", application.TwoFactorDisableInput{
			CurrentPassword: "very-secure-password", Code: code,
		}); err != nil {
			t.Fatalf("rolled-back disable code must still work: %v", err)
		}
	}
}
