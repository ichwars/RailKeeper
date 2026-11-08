package authdata_test

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/authentication"
	"railkeeper/backend/internal/authentication/management"
	"railkeeper/backend/internal/credentials"
	"railkeeper/backend/internal/infrastructure/authdata"
)

func TestAccountMutationsRollBackPasswordSessionsOriginsAndAudit(t *testing.T) {
	for _, action := range []string{"PasswordChanged", "PasswordResetCompleted", "SessionRevoked", "Logout"} {
		t.Run(action, func(t *testing.T) {
			db, auth, id := originDatabase(t)
			input := application.LoginInput{Username: "admin", Password: "very-secure-password"}
			session, err := auth.Login(t.Context(), input)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := auth.Login(t.Context(), input); err != nil {
				t.Fatal(err)
			}
			var resetToken string
			if action == "PasswordResetCompleted" {
				result, err := auth.RequestPasswordReset(t.Context(), application.PasswordResetRequestInput{Email: "admin@example.test"})
				if err != nil {
					t.Fatal(err)
				}
				resetToken = result.ResetToken
			}
			var sessionID string
			if err := db.QueryRow(`SELECT id FROM sessions ORDER BY id LIMIT 1`).Scan(&sessionID); err != nil {
				t.Fatal(err)
			}
			snapshot := func() []any {
				t.Helper()
				var hash string
				var revoked, origins, resets, audits int
				ctx, cancel := context.WithTimeout(t.Context(), time.Second)
				defer cancel()
				if err := db.QueryRowContext(ctx, `SELECT password_hash,(SELECT COUNT(*) FROM sessions WHERE revoked_at IS NOT NULL),(SELECT COUNT(*) FROM login_origins),(SELECT COUNT(*) FROM password_reset_requests WHERE used_at IS NOT NULL),(SELECT COUNT(*) FROM audit_logs WHERE action=?) FROM users WHERE id=?`, action, id).Scan(&hash, &revoked, &origins, &resets, &audits); err != nil {
					t.Fatal(err)
				}
				return []any{hash, revoked, origins, resets, audits}
			}
			before := snapshot()
			if _, err := db.Exec(`CREATE TRIGGER fail_account_audit BEFORE INSERT ON audit_logs WHEN NEW.action='` + action + `' BEGIN SELECT RAISE(ABORT,'synthetic audit fault'); END`); err != nil {
				t.Fatal(err)
			}
			switch action {
			case "PasswordChanged":
				err = auth.ChangeOwnPassword(t.Context(), id, session.SessionToken, application.ChangePasswordInput{CurrentPassword: input.Password, NewPassword: "replacement-password"})
			case "PasswordResetCompleted":
				err = auth.ResetPassword(t.Context(), application.PasswordResetConfirmInput{Token: resetToken, NewPassword: "replacement-password"})
			case "SessionRevoked":
				err = auth.RevokeSession(t.Context(), id, sessionID)
			case "Logout":
				err = auth.Logout(t.Context(), session.SessionToken)
			}
			if err == nil {
				t.Fatal("failed account mutation acknowledged")
			}
			if after := snapshot(); !reflect.DeepEqual(before, after) {
				t.Fatal("partial account mutation survived audit failure")
			}
			if _, err := auth.CurrentSession(t.Context(), session.SessionToken); err != nil {
				t.Fatal("audit failure revoked prior session", err)
			}
		})
	}
}

func TestFailedResetRequestRetainsPreviouslyIssuedLink(t *testing.T) {
	db, auth, _ := originDatabase(t)
	original, err := auth.RequestPasswordReset(t.Context(), application.PasswordResetRequestInput{Email: "admin@example.test"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TRIGGER fail_reset_request BEFORE INSERT ON password_reset_requests BEGIN SELECT RAISE(ABORT,'synthetic insert fault'); END`); err != nil {
		t.Fatal(err)
	}
	if result, err := auth.RequestPasswordReset(t.Context(), application.PasswordResetRequestInput{Email: "admin@example.test"}); err == nil || result != nil {
		t.Fatal("failed request returned a token")
	}
	var active int
	if err := db.QueryRow(`SELECT COUNT(*) FROM password_reset_requests WHERE handled_at IS NULL AND token_hash=?`, credentials.HashToken(original.ResetToken)).Scan(&active); err != nil || active != 1 {
		t.Fatal("failed request invalidated prior link", active, err)
	}
}

func TestPasswordResetCannotBeConsumedTwiceConcurrently(t *testing.T) {
	db, auth, _ := originDatabase(t)
	request, err := auth.RequestPasswordReset(t.Context(), application.PasswordResetRequestInput{Email: "admin@example.test"})
	if err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	results := make(chan error, 2)
	var workers sync.WaitGroup
	for range 2 {
		workers.Go(func() {
			<-start
			results <- auth.ResetPassword(t.Context(), application.PasswordResetConfirmInput{Token: request.ResetToken, NewPassword: "replacement-password"})
		})
	}
	close(start)
	workers.Wait()
	close(results)
	success := 0
	for err := range results {
		if err == nil {
			success++
		} else if !errors.Is(err, application.ErrPasswordResetInvalid) {
			t.Fatal(err)
		}
	}
	if success != 1 {
		t.Fatal("reset token was not single-use", success)
	}
	var audits, used int
	if err := db.QueryRow(`SELECT (SELECT COUNT(*) FROM audit_logs WHERE action='PasswordResetCompleted'),(SELECT COUNT(*) FROM password_reset_requests WHERE used_at IS NOT NULL)`).Scan(&audits, &used); err != nil || audits != 1 || used != 1 {
		t.Fatal("duplicate reset effects", audits, used, err)
	}
}

func TestStalePasswordProofCannotOverwriteConcurrentPasswordChange(t *testing.T) {
	db, auth, id := originDatabase(t)
	session, err := auth.Login(t.Context(), application.LoginInput{Username: "admin", Password: "very-secure-password"})
	if err != nil {
		t.Fatal(err)
	}
	store := authdata.NewStore(db)
	old, err := store.ReadCredential(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}
	if err := auth.ChangeOwnPassword(t.Context(), id, session.SessionToken, application.ChangePasswordInput{CurrentPassword: "very-secure-password", NewPassword: "first-replacement-password"}); err != nil {
		t.Fatal(err)
	}
	if err := store.ChangePassword(t.Context(), management.PasswordChange{UserID: id, SessionTokenHash: credentials.HashToken(session.SessionToken), ExpectedHash: old.PasswordHash, NextHash: "stale-derived-hash"}); !errors.Is(err, authentication.ErrInvalidLogin) {
		t.Fatal("stale password proof accepted", err)
	}
	current, err := store.ReadCredential(t.Context(), id)
	if err != nil || !credentials.Verify("first-replacement-password", current.PasswordHash) {
		t.Fatal("concurrent password change was overwritten", err)
	}
}

func TestResetCommitRechecksExpiryTokenAndAdministrativeReplacement(t *testing.T) {
	for _, mode := range []string{"expired", "replaced", "wrong-token"} {
		t.Run(mode, func(t *testing.T) {
			db, _, id := originDatabase(t)
			now := time.Now().UTC()
			clock := func() time.Time { return now }
			auth := application.NewAuthService(db, application.WithAuthClock(clock))
			store := authdata.NewStoreWithClock(db, clock)
			request, err := auth.RequestPasswordReset(t.Context(), application.PasswordResetRequestInput{Email: "admin@example.test"})
			if err != nil {
				t.Fatal(err)
			}
			hash := credentials.HashToken(request.ResetToken)
			proof, err := store.ReadResetRequest(t.Context(), hash)
			if err != nil {
				t.Fatal(err)
			}
			switch mode {
			case "expired":
				now = now.Add(31 * time.Minute)
			case "replaced":
				if _, err := auth.RequestPasswordReset(t.Context(), application.PasswordResetRequestInput{Email: "admin@example.test"}); err != nil {
					t.Fatal(err)
				}
			case "wrong-token":
				hash = "incorrect-hash"
			}
			if err := store.CommitPasswordReset(t.Context(), management.ResetCommit{RequestID: proof.ID, UserID: id, TokenHash: hash, NextHash: "must-not-be-written"}); !errors.Is(err, management.ErrPasswordResetInvalid) {
				t.Fatal("stale reset proof accepted", err)
			}
			account, err := store.ReadCredential(t.Context(), id)
			if err != nil || !credentials.Verify("very-secure-password", account.PasswordHash) {
				t.Fatal("invalid reset changed password", err)
			}
		})
	}
}

func TestMissingUserAndLastAdminRejectionsReleaseTheirTransaction(t *testing.T) {
	db, auth, id := originDatabase(t)
	if err := auth.DeleteUser(t.Context(), id, "missing"); !errors.Is(err, application.ErrUserNotFound) {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	if _, err := auth.GetUser(ctx, id); err != nil {
		t.Fatal("missing user leaked a transaction", err)
	}
	if err := auth.DeleteUser(ctx, id, id); !errors.Is(err, application.ErrLastAdmin) {
		t.Fatal(err)
	}
	if _, err := auth.GetUser(ctx, id); err != nil {
		t.Fatal("last-admin rejection leaked a transaction", err)
	}
	if err := db.PingContext(ctx); err != nil {
		t.Fatal(err)
	}
}
