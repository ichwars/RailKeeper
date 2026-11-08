package management

import (
	"context"
	"crypto/subtle"

	"railkeeper/backend/internal/authentication"
	"railkeeper/backend/internal/credentials"
)

func (s *Service) sessionUser(ctx context.Context, token string) (string, string, error) {
	if token == "" {
		return "", "", ErrUnauthorized
	}
	return s.store.SessionUser(ctx, credentials.HashToken(token))
}
func (s *Service) userRoles(ctx context.Context, id string) ([]string, error) {
	return s.store.Roles(ctx, id)
}
func (s *Service) CurrentSession(ctx context.Context, token string) (*SessionView, error) {
	id, csrf, err := s.sessionUser(ctx, token)
	if err != nil {
		return nil, err
	}
	user, err := s.store.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}
	return &SessionView{Username: user.Username, Roles: user.Roles, CSRFToken: csrf, TwoFactorEnabled: user.TwoFactorEnabled}, nil
}
func (s *Service) Logout(ctx context.Context, token string) error {
	id, _, err := s.sessionUser(ctx, token)
	if err != nil {
		return nil //nolint:nilerr // Logout remains idempotent for invalid sessions.
	}
	return s.store.RevokeOwnSession(ctx, id, credentials.HashToken(token))
}
func (s *Service) RequireRole(ctx context.Context, sessionToken, role string) (string, error) {
	return s.RequireAnyRole(ctx, sessionToken, role)
}

func (s *Service) RequireAnyRole(ctx context.Context, sessionToken string, requiredRoles ...string) (string, error) {
	userID, _, err := s.sessionUser(ctx, sessionToken)
	if err != nil {
		return "", err
	}

	if len(requiredRoles) == 0 {
		return userID, nil
	}

	roles, err := s.userRoles(ctx, userID)
	if err != nil {
		return "", err
	}

	for _, role := range requiredRoles {
		if role == "" || HasRole(roles, role) || HasRole(roles, "Admin") {
			return userID, nil
		}
	}

	return "", authentication.ErrForbidden
}

func (s *Service) ValidateCSRF(ctx context.Context, sessionToken, csrfToken string) error {
	_, storedToken, err := s.sessionUser(ctx, sessionToken)
	if err != nil {
		return err
	}
	if csrfToken == "" || subtle.ConstantTimeCompare([]byte(csrfToken), []byte(storedToken)) != 1 {
		return ErrInvalidCSRF
	}
	return nil
}
