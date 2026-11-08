package management

import (
	"context"
	"fmt"
	"strings"

	"railkeeper/backend/internal/credentials"
)

func prepareUser(username, email, password string, roles []string, create bool) (UserMutation, error) {
	mutation := UserMutation{Username: strings.TrimSpace(username), Email: strings.TrimSpace(email), Roles: CleanRoleNames(roles)}
	if len(mutation.Username) < 3 || len(mutation.Roles) == 0 || (mutation.Email != "" && !ValidEmail(mutation.Email)) ||
		((create || password != "") && len(password) < 12) {
		return UserMutation{}, ErrUserValidation
	}
	if password != "" {
		hash, err := credentials.Hash(password)
		if err != nil {
			return UserMutation{}, fmt.Errorf("hash password: %w", err)
		}
		mutation.PasswordHash = hash
	}
	return mutation, nil
}
func (s *Service) CreateUser(ctx context.Context, actor string, input CreateUserInput) (*UserView, error) {
	mutation, err := prepareUser(input.Username, input.Email, input.Password, input.Roles, true)
	if err != nil {
		return nil, err
	}
	return s.store.CreateUser(ctx, actor, mutation)
}
func (s *Service) UpdateUser(ctx context.Context, actor, id string, input UpdateUserInput) (*UserView, error) {
	mutation, err := prepareUser(input.Username, input.Email, input.Password, input.Roles, false)
	if err != nil {
		return nil, err
	}
	return s.store.UpdateUser(ctx, actor, id, mutation)
}
