// Package management owns account, session, password and second-factor use cases.
// Persistence ports carry verified mutations; SQLite transactions belong to authdata.
package management

import (
	"context"
	"errors"

	"railkeeper/backend/internal/authentication"
)

var (
	ErrUnauthorized         = errors.New("unauthorized")
	ErrInvalidCSRF          = errors.New("invalid csrf token")
	ErrUserValidation       = errors.New("user validation failed")
	ErrSessionNotFound      = errors.New("session not found")
	ErrDuplicateUser        = errors.New("user already exists")
	ErrLastAdmin            = errors.New("last admin cannot be removed")
	ErrPasswordResetInvalid = errors.New("password reset token invalid")
)

type ChangePasswordInput struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

type TwoFactorStatus struct {
	Enabled    bool   `json:"enabled"`
	Prepared   bool   `json:"prepared"`
	EnabledAt  string `json:"enabledAt,omitempty"`
	Username   string `json:"username,omitempty"`
	OtpauthURL string `json:"otpauthUrl,omitempty"`
	Secret     string `json:"secret,omitempty"`
}

type TwoFactorEnableInput struct {
	Code string `json:"code"`
}

type TwoFactorDisableInput struct {
	CurrentPassword string `json:"currentPassword"`
	Code            string `json:"code"`
}

type SessionView struct {
	Username         string   `json:"username"`
	Roles            []string `json:"roles"`
	CSRFToken        string   `json:"csrfToken"`
	TwoFactorEnabled bool     `json:"twoFactorEnabled"`
}

type RoleView struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type UserView struct {
	ID               string   `json:"id"`
	Username         string   `json:"username"`
	Email            string   `json:"email,omitempty"`
	Roles            []string `json:"roles"`
	CreatedAt        string   `json:"createdAt"`
	TwoFactorEnabled bool     `json:"twoFactorEnabled"`
}

type AuditLogEntry struct {
	ID            string `json:"id"`
	ActorUserID   string `json:"actorUserId,omitempty"`
	ActorUsername string `json:"actorUsername,omitempty"`
	Action        string `json:"action"`
	TargetType    string `json:"targetType,omitempty"`
	TargetID      string `json:"targetId,omitempty"`
	CreatedAt     string `json:"createdAt"`
	DetailsJSON   string `json:"detailsJson"`
}

type SessionRecord struct {
	ID        string `json:"id"`
	UserID    string `json:"userId"`
	Username  string `json:"username"`
	CreatedAt string `json:"createdAt"`
	ExpiresAt string `json:"expiresAt"`
	RevokedAt string `json:"revokedAt,omitempty"`
	Active    bool   `json:"active"`
}

type CreateUserInput struct {
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Password string   `json:"password"`
	Roles    []string `json:"roles"`
}

type UpdateUserInput struct {
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Password string   `json:"password"`
	Roles    []string `json:"roles"`
}

type PasswordResetRequestInput struct {
	Email string `json:"email"`
}

type PasswordResetRequestResult struct {
	Status     string `json:"status"`
	Message    string `json:"message"`
	ResetToken string `json:"-"`
	ResetURL   string `json:"-"`
	ExpiresAt  string `json:"-"`
}

type PasswordResetConfirmInput struct {
	Token       string `json:"token"`
	NewPassword string `json:"newPassword"`
}

type UserMutation struct {
	Username, Email, PasswordHash string
	Roles                         []string
}
type PasswordChange struct{ UserID, SessionTokenHash, ExpectedHash, NextHash string }
type ResetRequest struct{ UserID, Email, TokenHash, ExpiresAt string }
type ResetProof struct{ ID, UserID, ExpiresAt, UsedAt, HandledAt string }
type ResetCommit struct{ RequestID, UserID, TokenHash, NextHash string }
type TwoFactorState struct {
	Username, Secret, EnabledAt string
	Enabled                     bool
}

type Store interface {
	Roles(context.Context, string) ([]string, error)
	ListUsers(context.Context) ([]UserView, error)
	ListRoles(context.Context) ([]RoleView, error)
	ListAuditLog(context.Context, int) ([]AuditLogEntry, error)
	RecordAudit(context.Context, string, string, string, string, string) error
	ListSessions(context.Context, int) ([]SessionRecord, error)
	RevokeSession(context.Context, string, string) error
	GetUser(context.Context, string) (*UserView, error)
	CreateUser(context.Context, string, UserMutation) (*UserView, error)
	UpdateUser(context.Context, string, string, UserMutation) (*UserView, error)
	DeleteUser(context.Context, string, string) error
	ReadCredential(context.Context, string) (authentication.Account, error)
	ChangePassword(context.Context, PasswordChange) error
	FindResetUser(context.Context, string) (string, error)
	CreateResetRequest(context.Context, ResetRequest) error
	ReadResetRequest(context.Context, string) (ResetProof, error)
	CommitPasswordReset(context.Context, ResetCommit) error
	SessionUser(context.Context, string) (string, string, error)
	RevokeOwnSession(context.Context, string, string) error
	ReadTwoFactor(context.Context, string) (TwoFactorState, error)
	PrepareTwoFactor(context.Context, string, func() (string, error)) (TwoFactorState, error)
	CommitTwoFactorEnable(context.Context, string, string, int64) error
	CommitTwoFactorDisable(context.Context, string, string, string, string, bool, int64) error
}
