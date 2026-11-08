package management

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"railkeeper/backend/internal/authentication"
	"railkeeper/backend/internal/credentials"
)

type memoryPort struct {
	Store
	err             error
	user            UserView
	account         authentication.Account
	roles           []string
	mutation        UserMutation
	change          PasswordChange
	request         ResetRequest
	proof           ResetProof
	reset           ResetCommit
	calls           int
	sessionID, csrf string
	limit           int
	state           TwoFactorState
	step            int64
	audit           string
}

func (p *memoryPort) GetUser(context.Context, string) (*UserView, error) { return &p.user, p.err }
func (p *memoryPort) Roles(context.Context, string) ([]string, error)    { return p.roles, p.err }
func (p *memoryPort) SessionUser(context.Context, string) (string, string, error) {
	return p.sessionID, p.csrf, p.err
}
func (p *memoryPort) CreateUser(_ context.Context, _ string, m UserMutation) (*UserView, error) {
	p.calls++
	p.mutation = m
	return &p.user, p.err
}
func (p *memoryPort) UpdateUser(_ context.Context, _, _ string, m UserMutation) (*UserView, error) {
	p.calls++
	p.mutation = m
	return &p.user, p.err
}
func (p *memoryPort) ReadCredential(context.Context, string) (authentication.Account, error) {
	return p.account, p.err
}
func (p *memoryPort) ChangePassword(_ context.Context, change PasswordChange) error {
	p.calls++
	p.change = change
	return p.err
}
func (p *memoryPort) RecordAudit(_ context.Context, _, action, _, _, _ string) error {
	p.audit = action
	return p.err
}
func (p *memoryPort) FindResetUser(context.Context, string) (string, error) {
	return p.account.ID, p.err
}
func (p *memoryPort) CreateResetRequest(_ context.Context, r ResetRequest) error {
	p.calls++
	p.request = r
	return p.err
}
func (p *memoryPort) ReadResetRequest(context.Context, string) (ResetProof, error) {
	return p.proof, p.err
}
func (p *memoryPort) CommitPasswordReset(_ context.Context, r ResetCommit) error {
	p.calls++
	p.reset = r
	return p.err
}
func (p *memoryPort) ListAuditLog(_ context.Context, n int) ([]AuditLogEntry, error) {
	p.limit = n
	return nil, p.err
}
func (p *memoryPort) ListSessions(_ context.Context, n int) ([]SessionRecord, error) {
	p.limit = n
	return nil, p.err
}
func (p *memoryPort) ReadTwoFactor(context.Context, string) (TwoFactorState, error) {
	return p.state, p.err
}
func (p *memoryPort) PrepareTwoFactor(context.Context, string, func() (string, error)) (TwoFactorState, error) {
	return p.state, p.err
}
func (p *memoryPort) CommitTwoFactorEnable(_ context.Context, _, _ string, step int64) error {
	p.calls++
	p.step = step
	return p.err
}
func (p *memoryPort) CommitTwoFactorDisable(_ context.Context, _, _, _, _ string, _ bool, step int64) error {
	p.calls++
	p.step = step
	return p.err
}
func (p *memoryPort) RevokeOwnSession(context.Context, string, string) error { p.calls++; return p.err }

func TestAccountValidationPreparesHashesAndRetainsOmittedPassword(t *testing.T) {
	p := &memoryPort{}
	service := NewService(p, nil)
	for _, input := range []CreateUserInput{
		{Username: "a", Password: "long-password", Roles: []string{"Viewer"}},
		{Username: "valid", Password: "short", Roles: []string{"Viewer"}},
		{Username: "valid", Password: "long-password"},
		{Username: "valid", Password: "long-password", Email: "invalid", Roles: []string{"Viewer"}},
	} {
		if _, err := service.CreateUser(t.Context(), "actor", input); !errors.Is(err, ErrUserValidation) {
			t.Fatal(err)
		}
	}
	if p.calls != 0 {
		t.Fatal("invalid account reached persistence")
	}
	if _, err := service.CreateUser(t.Context(), "actor", CreateUserInput{Username: " valid ", Email: " valid@example.test ", Password: "long-password", Roles: []string{" Viewer ", "Viewer", "", "Editor"}}); err != nil {
		t.Fatal(err)
	}
	if p.mutation.Username != "valid" || p.mutation.Email != "valid@example.test" || !reflect.DeepEqual(p.mutation.Roles, []string{"Viewer", "Editor"}) || !credentials.Verify("long-password", p.mutation.PasswordHash) {
		t.Fatal("lost normalized account policy")
	}
	if _, err := service.UpdateUser(t.Context(), "actor", "id", UpdateUserInput{Username: "valid", Roles: []string{"Viewer"}}); err != nil || p.mutation.PasswordHash != "" {
		t.Fatal("omitted password was replaced", err)
	}
	if _, err := service.UpdateUser(t.Context(), "actor", "id", UpdateUserInput{Username: "valid", Password: "short", Roles: []string{"Viewer"}}); !errors.Is(err, ErrUserValidation) {
		t.Fatal(err)
	}
}

func TestPasswordChangeProducesVerifiedMutationAndRejectsIncorrectPassword(t *testing.T) {
	hash, err := credentials.Hash("current-password")
	if err != nil {
		t.Fatal(err)
	}
	p := &memoryPort{account: authentication.Account{PasswordHash: hash}}
	service := NewService(p, nil)
	for _, input := range []ChangePasswordInput{{NewPassword: "short"}, {CurrentPassword: "wrong", NewPassword: "replacement-password"}} {
		if err := service.ChangeOwnPassword(t.Context(), "user", "session", input); err == nil {
			t.Fatal("invalid password change accepted")
		}
	}
	if p.calls != 0 || p.audit != "PasswordChangeFailed" {
		t.Fatal("unverified password reached mutation")
	}
	if err := service.ChangeOwnPassword(t.Context(), "user", "session", ChangePasswordInput{CurrentPassword: "current-password", NewPassword: "replacement-password"}); err != nil {
		t.Fatal(err)
	}
	if p.change.UserID != "user" || p.change.ExpectedHash != hash || p.change.SessionTokenHash != credentials.HashToken("session") || !credentials.Verify("replacement-password", p.change.NextHash) {
		t.Fatal("password proof lost")
	}
	p.err = errors.New("synthetic credential read failure")
	if err := service.ChangeOwnPassword(t.Context(), "user", "session", ChangePasswordInput{NewPassword: "replacement-password"}); !errors.Is(err, p.err) {
		t.Fatal("read failure masked")
	}
}

func TestPasswordResetSeparatesPublicAcknowledgementFromPrivateToken(t *testing.T) {
	now := time.Unix(120, 0)
	p := &memoryPort{}
	service := NewService(p, func() time.Time { return now })
	if _, err := service.RequestPasswordReset(t.Context(), PasswordResetRequestInput{Email: "invalid"}); !errors.Is(err, ErrUserValidation) {
		t.Fatal(err)
	}
	unknown, err := service.RequestPasswordReset(t.Context(), PasswordResetRequestInput{Email: " Unknown@Example.Test "})
	if err != nil || unknown.ResetToken != "" || p.request.TokenHash != "" || p.request.Email != "unknown@example.test" {
		t.Fatal("unknown reset leaked a token", err)
	}
	p.account.ID = "user"
	known, err := service.RequestPasswordReset(t.Context(), PasswordResetRequestInput{Email: " Known@Example.Test "})
	if err != nil || known.Status != unknown.Status || known.Message != unknown.Message || credentials.HashToken(known.ResetToken) != p.request.TokenHash || p.request.ExpiresAt != now.UTC().Add(passwordResetTTL).Format(time.RFC3339) {
		t.Fatal("known reset lost privacy or expiry", err)
	}
	p.proof = ResetProof{ID: "request", UserID: "user", ExpiresAt: now.Add(time.Minute).Format(time.RFC3339)}
	p.calls = 0
	if err := service.ResetPassword(t.Context(), PasswordResetConfirmInput{Token: known.ResetToken, NewPassword: "replacement-password"}); err != nil {
		t.Fatal(err)
	}
	if p.reset.RequestID != "request" || p.reset.UserID != "user" || !credentials.Verify("replacement-password", p.reset.NextHash) {
		t.Fatal("reset proof lost")
	}
	p.proof.UsedAt = "already used"
	if err := service.ResetPassword(t.Context(), PasswordResetConfirmInput{Token: known.ResetToken, NewPassword: "replacement-password"}); !errors.Is(err, ErrPasswordResetInvalid) || p.calls != 1 {
		t.Fatal("consumed reset reached persistence")
	}
	if err := service.ResetPassword(t.Context(), PasswordResetConfirmInput{NewPassword: "short"}); !errors.Is(err, ErrUserValidation) {
		t.Fatal(err)
	}
}

func TestResetExpiryAndLastAdministratorPolicies(t *testing.T) {
	now := time.Unix(120, 0)
	valid := ResetProof{UserID: "user", ExpiresAt: now.Add(time.Minute).Format(time.RFC3339)}
	if !ValidResetProof(valid, now) {
		t.Fatal("valid proof rejected")
	}
	for _, proof := range []ResetProof{{}, {UserID: "user", ExpiresAt: "bad"}, {UserID: "user", ExpiresAt: now.Format(time.RFC3339)}, {UserID: "user", ExpiresAt: valid.ExpiresAt, HandledAt: "handled"}} {
		if ValidResetProof(proof, now) {
			t.Fatal("invalid/expired reset accepted")
		}
	}
	if !errors.Is(RequireAnotherAdmin(0), ErrLastAdmin) || RequireAnotherAdmin(1) != nil {
		t.Fatal("last admin policy")
	}
	for _, role := range []string{"Admin", "Editor", "Viewer"} {
		if !HasRole([]string{role}, "Viewer") {
			t.Fatal("read inheritance lost", role)
		}
	}
	if HasRole([]string{"Viewer"}, "Admin") {
		t.Fatal("viewer escalated")
	}
}

func TestSessionPolicyKeepsAdminOverrideCSRFAndActualRoles(t *testing.T) {
	p := &memoryPort{sessionID: "user", csrf: "secret", roles: []string{"Editor"}, user: UserView{Username: "operator", Roles: []string{"Editor"}}}
	service := NewService(p, nil)
	if _, err := service.RequireAnyRole(t.Context(), "session", "Viewer"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.RequireRole(t.Context(), "session", "Messe"); !errors.Is(err, authentication.ErrForbidden) {
		t.Fatal("ungranted role accepted")
	}
	p.roles = []string{"Admin"}
	if _, err := service.RequireAnyRole(t.Context(), "session", "Messe"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.RequireAnyRole(t.Context(), "session"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.RequireRole(t.Context(), "", "Viewer"); !errors.Is(err, ErrUnauthorized) {
		t.Fatal(err)
	}
	for _, csrf := range []string{"", "incorrect", "secret"} {
		err := service.ValidateCSRF(t.Context(), "session", csrf)
		if (err == nil) != (csrf == "secret") {
			t.Fatal("csrf policy", csrf, err)
		}
	}
	session, err := service.CurrentSession(t.Context(), "session")
	if err != nil || session.Username != "operator" || session.CSRFToken != "secret" || session.Roles[0] != "Editor" {
		t.Fatal("session projection changed", err)
	}
	p.err = ErrUnauthorized
	if err := service.Logout(t.Context(), "missing"); err != nil {
		t.Fatal("idempotent logout", err)
	}
}

func TestSecondFactorPreparationHidesSecretOnStatusAndConsumesVerifiedStep(t *testing.T) {
	now := time.Unix(59, 0)
	secret := "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"
	p := &memoryPort{state: TwoFactorState{Username: "operator", Secret: secret}}
	service := NewService(p, func() time.Time { return now })
	status, err := service.TwoFactorStatus(t.Context(), "user")
	if err != nil || status.Secret != "" || status.OtpauthURL != "" || !status.Prepared {
		t.Fatal("status leaked secret", err)
	}
	prepared, err := service.PrepareTwoFactor(t.Context(), "user")
	if err != nil || prepared.Secret != secret || prepared.OtpauthURL == "" {
		t.Fatal("preparation lost provisioning", err)
	}
	p.state.Enabled = true
	prepared, err = service.PrepareTwoFactor(t.Context(), "user")
	if err != nil || prepared.Secret != "" || prepared.OtpauthURL != "" {
		t.Fatal("enabled preparation leaked secret", err)
	}
	for _, code := range []string{"", "000000", "287082"} {
		_, err := service.EnableTwoFactor(t.Context(), "user", TwoFactorEnableInput{Code: code})
		if (err == nil) != (code == "287082") {
			t.Fatal(code, err)
		}
	}
	if p.calls != 1 || p.step != 1 || p.audit != "TwoFactorEnableFailed" {
		t.Fatal("wrong second factor reached commit")
	}
	hash, err := credentials.Hash("current-password")
	if err != nil {
		t.Fatal(err)
	}
	p.account = authentication.Account{PasswordHash: hash, TwoFactorSecret: secret, TwoFactorEnabled: true}
	for _, input := range []TwoFactorDisableInput{{CurrentPassword: "wrong", Code: "287082"}, {CurrentPassword: "current-password", Code: "000000"}, {CurrentPassword: "current-password", Code: "287082"}} {
		err := service.DisableTwoFactor(t.Context(), "user", "session", input)
		if (err == nil) != (input.CurrentPassword == "current-password" && input.Code == "287082") {
			t.Fatal("disable proof", err)
		}
	}
	if p.calls != 2 || p.step != 1 {
		t.Fatal("second factor disable did not pass exactly one verified proof")
	}
}

func TestPublicListLimitsAreBoundedBeforePersistence(t *testing.T) {
	p := &memoryPort{}
	service := NewService(p, nil)
	for _, requested := range []int{-1, 10, 999} {
		if _, err := service.ListAuditLog(t.Context(), requested); err != nil {
			t.Fatal(err)
		}
		want := requested
		if requested < 0 {
			want = 50
		}
		if requested > 200 {
			want = 200
		}
		if p.limit != want {
			t.Fatal("audit limit", p.limit, want)
		}
		if _, err := service.ListSessions(t.Context(), requested); err != nil {
			t.Fatal(err)
		}
		if requested < 0 {
			want = 200
		}
		if p.limit != want {
			t.Fatal("session limit", p.limit, want)
		}
	}
}
