package bootstrap

import (
	"context"
	"errors"
)

var ErrAlreadySetup = errors.New("setup already completed")
var ErrWeakSetup = errors.New("username or password is too short")
var ErrInvalidSetupCode = errors.New("invalid or consumed setup code")

type CreateAdminInput struct {
	Username  string
	Email     string
	Password  string
	SetupCode string
}
type Administrator struct{ ID, Username, Email, PasswordHash, At string }
type AdministratorAudit struct{ ID, UserID, At string }
type Reader interface {
	UserCount(context.Context) (int, error)
}
type Repository interface {
	Reader
	PendingAuthorization(context.Context) (string, bool, error)
	Begin(context.Context) (Mutation, error)
}
type Mutation interface {
	Reader
	ReplaceAuthorization(context.Context, string, string) error
	ConsumeAuthorization(context.Context, string, string) (bool, error)
	InsertFirstAdministrator(context.Context, Administrator) (int64, error)
	AssignAdministratorRoles(context.Context, string) error
	AuditAdministrator(context.Context, AdministratorAudit) error
	Commit() error
	Rollback() error
}
type Service struct{ repository Repository }

func New(repository Repository) *Service { return &Service{repository: repository} }

type setupCodeProof struct{ hash string }
