package authentication

import (
	"context"
	"encoding/json"
	"time"

	"railkeeper/backend/internal/credentials"
)

const OriginBudgetScope = "login-origin"
const OriginTTL = 30 * 24 * time.Hour

// Origin admits a browser to a separate attempt budget. It never authenticates it.
type Origin struct {
	UserID, TokenHash, BudgetKey, StateHash string
	ExpiresAt                               time.Time
}

type OriginStore interface {
	FindOrigin(context.Context, string, string) (Origin, Account, bool, error)
	ResetLoginBudget(context.Context, string, string, time.Time) error
}

// CredentialState invalidates recognition after password or second-factor changes.
// The consumed TOTP step deliberately does not change browser recognition.
func CredentialState(account Account) string {
	encoded, _ := json.Marshal([]any{account.ID, account.Username, account.PasswordHash, account.TwoFactorEnabled, account.TwoFactorSecret})
	return credentials.HashToken(string(encoded))
}

func (s *Service) recognizeOrigin(ctx context.Context, username, token string) (Origin, bool, error) {
	store, supported := s.store.(OriginStore)
	if !supported || len(token) != 43 {
		return Origin{}, false, nil
	}
	if !s.work.beginOriginCheck() {
		return Origin{}, false, ErrLoginBusy
	}
	defer s.work.endOriginCheck()
	lookupCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	origin, account, found, err := store.FindOrigin(lookupCtx, username, credentials.HashToken(token))
	if err != nil {
		if ctx.Err() == nil && lookupCtx.Err() != nil {
			return Origin{}, false, ErrLoginBusy
		}
		return Origin{}, false, err
	}
	valid := found && origin.TokenHash == credentials.HashToken(token) && account.Username == username && account.ID == origin.UserID &&
		origin.ExpiresAt.After(s.now().UTC()) && origin.StateHash == CredentialState(account)
	return origin, valid, nil
}

func (s *Service) ResetLoginBudget(ctx context.Context, actorID, userID string) error {
	store, supported := s.store.(OriginStore)
	if !supported {
		return ErrForbidden
	}
	return store.ResetLoginBudget(ctx, actorID, userID, s.now().UTC())
}
