// Package authentication owns login decisions without HTTP or persistence dependencies.
package authentication

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidLogin      = errors.New("invalid credentials")
	ErrTwoFactorRequired = errors.New("two factor code required")
	ErrLoginThrottled    = errors.New("login temporarily limited")
)

const AccountBudgetScope = "login-account"

type Input struct{ Username, Password, TwoFactorCode string }

type Account struct {
	ID, Username, PasswordHash, TwoFactorSecret string
	TwoFactorEnabled                            bool
}

type Result struct {
	SessionToken, CSRFToken, Username string
	ExpiresAt                         time.Time
	Roles                             []string
	TwoFactorEnabled                  bool
}

type Attempt struct {
	Key    string
	Now    time.Time
	Limit  int
	Window time.Duration
}

// Commit is a proof against the account state originally verified by Login.
// The store must recheck it and atomically consume the code, upgrade the hash,
// create the session, audit success and reset this account's attempt budget.
type Commit struct {
	Account                  Account
	AttemptKey, UpgradedHash string
	TwoFactorStep            int64
	Session                  Result
	Now                      time.Time
}

type Audit struct {
	UserID, Action, TargetID string
	Now                      time.Time
}

type Store interface {
	AllowAttempt(context.Context, Attempt) (bool, error)
	FindAccount(context.Context, string) (Account, bool, error)
	Roles(context.Context, string) ([]string, error)
	CommitLogin(context.Context, Commit) error
	RecordFailure(context.Context, Audit) error
}

type PasswordPolicy interface {
	Verify(password, encoded string) bool
	Upgrade(password, encoded string) (string, error)
	DummyHash() string
}
