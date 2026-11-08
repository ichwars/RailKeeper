package authdata_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/authentication"
	"railkeeper/backend/internal/authentication/management"
	"railkeeper/backend/internal/credentials"
	"railkeeper/backend/internal/infrastructure/authdata"
)

func TestUnavailableAccountStorageNeverAcknowledgesReadsOrWrites(t *testing.T) {
	db, _, id := originDatabase(t)
	store := authdata.NewStore(db)
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	for name, operation := range map[string]func(context.Context) error{
		"users":         func(ctx context.Context) error { _, err := store.ListUsers(ctx); return err },
		"roles":         func(ctx context.Context) error { _, err := store.ListRoles(ctx); return err },
		"account-roles": func(ctx context.Context) error { _, err := store.Roles(ctx, id); return err },
		"audit":         func(ctx context.Context) error { _, err := store.ListAuditLog(ctx, 50); return err },
		"sessions":      func(ctx context.Context) error { _, err := store.ListSessions(ctx, 200); return err },
		"user":          func(ctx context.Context) error { _, err := store.GetUser(ctx, id); return err },
		"credentials":   func(ctx context.Context) error { _, err := store.ReadCredential(ctx, id); return err },
		"second-factor": func(ctx context.Context) error { _, err := store.ReadTwoFactor(ctx, id); return err },
		"reset-account": func(ctx context.Context) error { _, err := store.FindResetUser(ctx, "admin@example.test"); return err },
		"reset-proof":   func(ctx context.Context) error { _, err := store.ReadResetRequest(ctx, "hash"); return err },
		"session-user":  func(ctx context.Context) error { _, _, err := store.SessionUser(ctx, "hash"); return err },
		"create": func(ctx context.Context) error {
			_, err := store.CreateUser(ctx, id, management.UserMutation{Username: "synthetic", PasswordHash: "prepared-hash", Roles: []string{"Viewer"}})
			return err
		},
		"update": func(ctx context.Context) error {
			_, err := store.UpdateUser(ctx, id, id, management.UserMutation{Username: "admin", Roles: []string{"Admin"}})
			return err
		},
		"delete": func(ctx context.Context) error { return store.DeleteUser(ctx, id, id) },
		"password": func(ctx context.Context) error {
			return store.ChangePassword(ctx, management.PasswordChange{UserID: id})
		},
		"reset-request": func(ctx context.Context) error {
			return store.CreateResetRequest(ctx, management.ResetRequest{UserID: id})
		},
		"reset-commit": func(ctx context.Context) error {
			return store.CommitPasswordReset(ctx, management.ResetCommit{UserID: id})
		},
		"revoke": func(ctx context.Context) error { return store.RevokeSession(ctx, id, "session") },
		"logout": func(ctx context.Context) error { return store.RevokeOwnSession(ctx, id, "hash") },
		"prepare": func(ctx context.Context) error {
			_, err := store.PrepareTwoFactor(ctx, id, func() (string, error) { t.Fatal("unavailable storage generated a secret"); return "", nil })
			return err
		},
		"enable": func(ctx context.Context) error { return store.CommitTwoFactorEnable(ctx, id, "secret", 1) },
		"disable": func(ctx context.Context) error {
			return store.CommitTwoFactorDisable(ctx, id, "hash", "password-hash", "secret", true, 1)
		},
		"record-audit": func(ctx context.Context) error { return store.RecordAudit(ctx, id, "Synthetic", "user", id, "{}") },
	} {
		t.Run(name, func(t *testing.T) {
			err := operation(t.Context())
			if err == nil || errors.Is(err, authentication.ErrInvalidLogin) || errors.Is(err, management.ErrPasswordResetInvalid) || errors.Is(err, management.ErrUnauthorized) {
				t.Fatal("storage failure acknowledged or disguised as rejected credentials", err)
			}
		})
	}
}

func TestRoleAssignmentFaultRollsBackAccountPasswordSessionsAndRecognition(t *testing.T) {
	for _, operation := range []string{"create", "update"} {
		t.Run(operation, func(t *testing.T) {
			db, auth, id := originDatabase(t)
			login, err := auth.Login(t.Context(), application.LoginInput{Username: "admin", Password: "very-secure-password"})
			if err != nil {
				t.Fatal(err)
			}
			snapshot := func() []any {
				t.Helper()
				var username, hash string
				var users, roles, revoked, origins int
				if err := db.QueryRow(`SELECT username,password_hash,(SELECT COUNT(*) FROM users),(SELECT COUNT(*) FROM user_roles),(SELECT COUNT(*) FROM sessions WHERE revoked_at IS NOT NULL),(SELECT COUNT(*) FROM login_origins) FROM users WHERE id=?`, id).Scan(&username, &hash, &users, &roles, &revoked, &origins); err != nil {
					t.Fatal(err)
				}
				return []any{username, hash, users, roles, revoked, origins}
			}
			before := snapshot()
			if _, err := db.Exec(`CREATE TRIGGER fail_role_assignment BEFORE INSERT ON user_roles BEGIN SELECT RAISE(ABORT,'synthetic role fault'); END`); err != nil {
				t.Fatal(err)
			}
			if operation == "create" {
				_, err = auth.CreateUser(t.Context(), id, application.CreateUserInput{Username: "synthetic-user", Password: "synthetic-password", Roles: []string{"Viewer"}})
			} else {
				_, err = auth.UpdateUser(t.Context(), id, id, application.UpdateUserInput{Username: "changed-admin", Password: "replacement-password", Roles: []string{"Admin"}})
			}
			if err == nil {
				t.Fatal("failed role assignment acknowledged")
			}
			if after := snapshot(); !reflect.DeepEqual(before, after) {
				t.Fatal("partial account effects survived role failure")
			}
			if _, err := auth.CurrentSession(t.Context(), login.SessionToken); err != nil {
				t.Fatal("role failure revoked existing session", err)
			}
		})
	}
}

func TestRoleReadFailureDoesNotReturnAnIncompleteAccount(t *testing.T) {
	db, auth, id := originDatabase(t)
	if _, err := db.Exec(`DROP TABLE roles`); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	if _, err := auth.ListUsers(ctx); err == nil {
		t.Fatal("missing role storage produced accounts without roles")
	}
	if _, err := auth.GetUser(ctx, id); err == nil {
		t.Fatal("missing role storage produced an incomplete account")
	}
}

func TestLoginCommitRejectsChangedOrUnavailableRecognizedOrigin(t *testing.T) {
	for _, mode := range []string{"state", "budget", "expired", "malformed", "missing-proof", "unavailable"} {
		t.Run(mode, func(t *testing.T) {
			db, auth, id := originDatabase(t)
			login, err := auth.Login(t.Context(), application.LoginInput{Username: "admin", Password: "very-secure-password"})
			if err != nil {
				t.Fatal(err)
			}
			store := authdata.NewStore(db)
			origin, account, found, err := store.FindOrigin(t.Context(), "admin", credentials.HashToken(login.OriginToken))
			if err != nil || !found {
				t.Fatal(err)
			}
			now := time.Now().UTC()
			proof := authentication.Commit{Account: account, Origin: &origin, OriginRecognized: true, Now: now, AttemptScope: authentication.OriginBudgetScope, AttemptKey: origin.BudgetKey, Session: authentication.Result{SessionToken: credentials.RandomToken(), CSRFToken: credentials.RandomToken(), ExpiresAt: now.Add(time.Hour)}}
			switch mode {
			case "state":
				_, err = db.Exec(`UPDATE login_origins SET state_hash='stale'`)
			case "budget":
				_, err = db.Exec(`UPDATE login_origins SET budget_key='changed'`)
			case "expired":
				_, err = db.Exec(`UPDATE login_origins SET expires_at=?`, now.Add(-time.Second).Format(time.RFC3339))
			case "malformed":
				_, err = db.Exec(`UPDATE login_origins SET expires_at='invalid'`)
			case "missing-proof":
				proof.Origin = nil
			case "unavailable":
				_, err = db.Exec(`DROP TABLE login_origins`)
			}
			if err != nil {
				t.Fatal(err)
			}
			if err := store.CommitLogin(t.Context(), proof); err == nil {
				t.Fatal("invalid recognition was committed")
			}
			var count int
			if err := db.QueryRow(`SELECT COUNT(*) FROM sessions WHERE user_id=?`, id).Scan(&count); err != nil || count != 1 {
				t.Fatal("invalid recognition created another session", count, err)
			}
		})
	}
}

func TestDefaultStoreClockKeepsExistingSessionsUsable(t *testing.T) {
	db, auth, id := originDatabase(t)
	login, err := auth.Login(t.Context(), application.LoginInput{Username: "admin", Password: "very-secure-password"})
	if err != nil {
		t.Fatal(err)
	}
	user, csrf, err := authdata.NewStoreWithClock(db, nil).SessionUser(t.Context(), credentials.HashToken(login.SessionToken))
	if err != nil || user != id || csrf != login.CSRFToken {
		t.Fatal("default store clock broke an existing session", err)
	}
}
