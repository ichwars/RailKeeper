package authdata

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"railkeeper/backend/internal/authentication"
	"railkeeper/backend/internal/authentication/management"
)

func (s *Store) CreateUser(ctx context.Context, actorUserID string, input management.UserMutation) (*management.UserView, error) {
	username, email, roleNames, hash := input.Username, input.Email, input.Roles, input.PasswordHash

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin user create: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if err = s.ensureRolesExist(ctx, tx, roleNames); err != nil {
		return nil, err
	}

	userID := randomID()
	now := s.now().UTC().Format(time.RFC3339)
	if _, err = tx.ExecContext(
		ctx,
		`INSERT INTO users(id, username, email, password_hash, created_at) VALUES(?, ?, ?, ?, ?)`,
		userID,
		username,
		nullableString(email),
		hash,
		now,
	); err != nil {
		if isUniqueConstraint(err) {
			return nil, management.ErrDuplicateUser
		}
		return nil, fmt.Errorf("insert user: %w", err)
	}

	if err = s.assignRoles(ctx, tx, userID, roleNames); err != nil {
		return nil, err
	}
	if err = s.auditTx(ctx, tx, actorUserID, "UserCreated", "user", userID, "{}"); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit user create: %w", err)
	}

	return s.GetUser(ctx, userID)
}

func (s *Store) UpdateUser(ctx context.Context, actorUserID, userID string, input management.UserMutation) (*management.UserView, error) {
	username, email, roleNames := input.Username, input.Email, input.Roles

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin user update: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var existingID string
	if err = tx.QueryRowContext(ctx, `SELECT id FROM users WHERE id=?`, userID).Scan(&existingID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, authentication.ErrUserNotFound
		}
		return nil, fmt.Errorf("read user for update: %w", err)
	}
	if err = s.ensureRolesExist(ctx, tx, roleNames); err != nil {
		return nil, err
	}
	if !management.HasRole(roleNames, "Admin") {
		if err = s.ensureAnotherAdmin(ctx, tx, userID); err != nil {
			return nil, err
		}
	}

	if input.PasswordHash != "" {
		hash := input.PasswordHash
		_, err = tx.ExecContext(ctx, `UPDATE users SET username=?, email=?, password_hash=? WHERE id=?`, username, nullableString(email), hash, userID)
		if err == nil {
			_, err = tx.ExecContext(
				ctx,
				`UPDATE sessions SET revoked_at=? WHERE user_id=? AND revoked_at IS NULL`,
				s.now().UTC().Format(time.RFC3339),
				userID,
			)
		}
	} else {
		_, err = tx.ExecContext(ctx, `UPDATE users SET username=?, email=? WHERE id=?`, username, nullableString(email), userID)
	}
	if err != nil {
		if isUniqueConstraint(err) {
			return nil, management.ErrDuplicateUser
		}
		return nil, fmt.Errorf("update user: %w", err)
	}

	if err = s.assignRoles(ctx, tx, userID, roleNames); err != nil {
		return nil, err
	}
	if err = s.auditTx(ctx, tx, actorUserID, "UserUpdated", "user", userID, "{}"); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit user update: %w", err)
	}

	return s.GetUser(ctx, userID)
}

func (s *Store) DeleteUser(ctx context.Context, actorUserID, userID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin user delete: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var count int
	if err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE id=?`, userID).Scan(&count); err != nil {
		return fmt.Errorf("read user for delete: %w", err)
	}
	if count == 0 {
		return authentication.ErrUserNotFound
	}
	roles, err := s.rolesTx(ctx, tx, userID)
	if err != nil {
		return err
	}
	if management.HasRole(roles, "Admin") {
		if err = s.ensureAnotherAdmin(ctx, tx, userID); err != nil {
			return err
		}
	}
	if err = s.auditTx(ctx, tx, actorUserID, "UserDeleted", "user", userID, "{}"); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM users WHERE id=?`, userID); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit user delete: %w", err)
	}
	return nil
}

func (s *Store) ensureRolesExist(ctx context.Context, tx *sql.Tx, roleNames []string) error {
	if len(roleNames) == 0 {
		return management.ErrUserValidation
	}
	for _, role := range roleNames {
		var count int
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM roles WHERE name=?`, role).Scan(&count); err != nil {
			return fmt.Errorf("check role %s: %w", role, err)
		}
		if count == 0 {
			return management.ErrUserValidation
		}
	}
	return nil
}

func (s *Store) assignRoles(ctx context.Context, tx *sql.Tx, userID string, roleNames []string) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM user_roles WHERE user_id=?`, userID); err != nil {
		return fmt.Errorf("clear user roles: %w", err)
	}
	for _, role := range roleNames {
		if _, err := tx.ExecContext(
			ctx,
			`INSERT INTO user_roles(user_id, role_id)
			 SELECT ?, id FROM roles WHERE name=?`,
			userID,
			role,
		); err != nil {
			return fmt.Errorf("assign role %s: %w", role, err)
		}
	}
	return nil
}

func (s *Store) ensureAnotherAdmin(ctx context.Context, tx *sql.Tx, userID string) error {
	var count int
	if err := tx.QueryRowContext(
		ctx,
		`SELECT COUNT(DISTINCT users.id)
		 FROM users
		 JOIN user_roles ON user_roles.user_id = users.id
		 JOIN roles ON roles.id = user_roles.role_id
		 WHERE roles.name='Admin' AND users.id<>?`,
		userID,
	).Scan(&count); err != nil {
		return fmt.Errorf("count other admins: %w", err)
	}
	return management.RequireAnotherAdmin(count)
}
