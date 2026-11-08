package application

import (
	"database/sql"
	"time"

	"railkeeper/backend/internal/authentication"
	"railkeeper/backend/internal/authentication/management"
	"railkeeper/backend/internal/infrastructure/authdata"
)

var (
	ErrInvalidLogin         = authentication.ErrInvalidLogin
	ErrUnauthorized         = management.ErrUnauthorized
	ErrInvalidCSRF          = management.ErrInvalidCSRF
	ErrForbidden            = authentication.ErrForbidden
	ErrUserValidation       = management.ErrUserValidation
	ErrUserNotFound         = authentication.ErrUserNotFound
	ErrSessionNotFound      = management.ErrSessionNotFound
	ErrDuplicateUser        = management.ErrDuplicateUser
	ErrLastAdmin            = management.ErrLastAdmin
	ErrPasswordResetInvalid = management.ErrPasswordResetInvalid
	ErrTwoFactorRequired    = authentication.ErrTwoFactorRequired
	ErrLoginThrottled       = authentication.ErrLoginThrottled
	ErrLoginBusy            = authentication.ErrLoginBusy
)

type AuthService struct {
	*management.Service
	loginStore authentication.Store
	login      *authentication.Service
	loginWork  *authentication.WorkLimiter
	now        func() time.Time
}

type LoginInput struct {
	Username      string `json:"username"`
	Password      string `json:"password"`
	TwoFactorCode string `json:"twoFactorCode"`
	OriginToken   string `json:"-"`
}

type ChangePasswordInput = management.ChangePasswordInput

type TwoFactorStatus = management.TwoFactorStatus

type TwoFactorEnableInput = management.TwoFactorEnableInput

type TwoFactorDisableInput = management.TwoFactorDisableInput

type SessionView = management.SessionView

type RoleView = management.RoleView

type UserView = management.UserView

type AuditLogEntry = management.AuditLogEntry

type SessionRecord = management.SessionRecord

type CreateUserInput = management.CreateUserInput

type UpdateUserInput = management.UpdateUserInput

type PasswordResetRequestInput = management.PasswordResetRequestInput

type PasswordResetRequestResult = management.PasswordResetRequestResult

type PasswordResetConfirmInput = management.PasswordResetConfirmInput

type LoginResult struct {
	SessionToken    string
	CSRFToken       string
	ExpiresAt       time.Time
	Session         SessionView
	OriginToken     string
	OriginExpiresAt time.Time
}

func NewAuthService(db *sql.DB, options ...AuthOption) *AuthService {
	service := &AuthService{now: time.Now}
	for _, option := range options {
		option(service)
	}
	if service.loginStore == nil {
		// Compatibility for existing in-process callers. Production supplies the port at startup.
		service.loginStore = authdata.NewStore(db)
	}
	service.Service = management.NewService(authdata.NewStoreWithClock(db, service.now), service.now)
	service.login = authentication.NewService(service.loginStore, authentication.WithClock(service.now), authentication.WithWorkLimiter(service.loginWork))
	return service
}
