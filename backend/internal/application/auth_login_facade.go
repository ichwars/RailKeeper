package application

import (
	"context"

	"railkeeper/backend/internal/authentication"
)

// Login preserves the transport contract while the independent module owns the use case.
func (s *AuthService) Login(ctx context.Context, input LoginInput) (*LoginResult, error) {
	result, err := s.login.Login(ctx, authentication.Input{Username: input.Username, Password: input.Password, TwoFactorCode: input.TwoFactorCode})
	if err != nil {
		return nil, err
	}
	return &LoginResult{SessionToken: result.SessionToken, CSRFToken: result.CSRFToken, ExpiresAt: result.ExpiresAt,
		Session: SessionView{Username: result.Username, Roles: result.Roles, CSRFToken: result.CSRFToken,
			TwoFactorEnabled: result.TwoFactorEnabled}}, nil
}
