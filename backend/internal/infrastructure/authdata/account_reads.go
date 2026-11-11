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

func (s *Store) ListUsers(ctx context.Context) ([]management.UserView, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT u.id, u.username, COALESCE(u.email, ''), u.created_at, u.two_factor_enabled, r.name
FROM users u
LEFT JOIN user_roles ur ON ur.user_id=u.id
LEFT JOIN roles r ON r.id=ur.role_id
ORDER BY lower(u.username), u.id, r.name`)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer func() { _ = rows.Close() }()

	users := []management.UserView{}
	for rows.Next() {
		var user management.UserView
		var role sql.NullString
		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.TwoFactorEnabled, &role); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		if len(users) == 0 || users[len(users)-1].ID != user.ID {
			user.Roles = []string{}
			users = append(users, user)
		}
		if role.Valid {
			last := &users[len(users)-1]
			last.Roles = append(last.Roles, role.String)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read users: %w", err)
	}
	return users, nil
}

func (s *Store) ListRoles(ctx context.Context) ([]management.RoleView, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name FROM roles ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list roles: %w", err)
	}
	defer func() { _ = rows.Close() }()

	roles := []management.RoleView{}
	for rows.Next() {
		var role management.RoleView
		if err := rows.Scan(&role.ID, &role.Name); err != nil {
			return nil, fmt.Errorf("scan role: %w", err)
		}
		roles = append(roles, role)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read roles: %w", err)
	}
	return roles, nil
}

func (s *Store) ListAuditLog(ctx context.Context, limit int) ([]management.AuditLogEntry, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT audit_logs.id,
		        COALESCE(audit_logs.actor_user_id, ''),
		        COALESCE(users.username, ''),
		        audit_logs.action,
		        COALESCE(audit_logs.target_type, ''),
		        COALESCE(audit_logs.target_id, ''),
		        audit_logs.created_at,
		        COALESCE(audit_logs.details_json, '{}')
		   FROM audit_logs
		   LEFT JOIN users ON users.id = audit_logs.actor_user_id
		  ORDER BY audit_logs.created_at DESC, audit_logs.id DESC
		  LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list audit log: %w", err)
	}
	defer func() { _ = rows.Close() }()

	entries := []management.AuditLogEntry{}
	for rows.Next() {
		var entry management.AuditLogEntry
		if err := rows.Scan(
			&entry.ID,
			&entry.ActorUserID,
			&entry.ActorUsername,
			&entry.Action,
			&entry.TargetType,
			&entry.TargetID,
			&entry.CreatedAt,
			&entry.DetailsJSON,
		); err != nil {
			return nil, fmt.Errorf("scan audit log: %w", err)
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read audit log: %w", err)
	}
	return entries, nil
}

func (s *Store) RecordAudit(ctx context.Context, actorUserID, action, targetType, targetID, detailsJSON string) error {
	return s.audit(ctx, actorUserID, action, targetType, targetID, detailsJSON)
}

func (s *Store) ListSessions(ctx context.Context, limit int) ([]management.SessionRecord, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT sessions.id,
		        sessions.user_id,
		        users.username,
		        sessions.created_at,
		        sessions.expires_at,
		        COALESCE(sessions.revoked_at, ''),
		        CASE WHEN sessions.revoked_at IS NULL AND sessions.expires_at > ? THEN 1 ELSE 0 END AS active
		   FROM sessions
		   JOIN users ON users.id = sessions.user_id
		  ORDER BY active DESC, sessions.created_at DESC
		  LIMIT ?`,
		s.now().UTC().Format(time.RFC3339),
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	sessions := []management.SessionRecord{}
	for rows.Next() {
		var session management.SessionRecord
		var active int
		if err := rows.Scan(&session.ID, &session.UserID, &session.Username, &session.CreatedAt, &session.ExpiresAt, &session.RevokedAt, &active); err != nil {
			return nil, fmt.Errorf("scan session: %w", err)
		}
		session.Active = active == 1
		sessions = append(sessions, session)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read sessions: %w", err)
	}
	return sessions, nil
}

func (s *Store) GetUser(ctx context.Context, userID string) (*management.UserView, error) {
	var user management.UserView
	if err := s.db.QueryRowContext(
		ctx,
		`SELECT id, username, COALESCE(email, ''), created_at, two_factor_enabled FROM users WHERE id=?`,
		userID,
	).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.TwoFactorEnabled); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, authentication.ErrUserNotFound
		}
		return nil, fmt.Errorf("read user: %w", err)
	}
	var err error
	user.Roles, err = s.roles(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Store) roles(ctx context.Context, userID string) ([]string, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT roles.name
		 FROM roles
		 JOIN user_roles ON user_roles.role_id = roles.id
		 WHERE user_roles.user_id=?
		 ORDER BY roles.name`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("read roles: %w", err)
	}
	defer func() { _ = rows.Close() }()

	roles := []string{}
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, fmt.Errorf("scan role: %w", err)
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func (s *Store) rolesTx(ctx context.Context, tx *sql.Tx, userID string) ([]string, error) {
	rows, err := tx.QueryContext(
		ctx,
		`SELECT roles.name
		 FROM roles
		 JOIN user_roles ON user_roles.role_id = roles.id
		 WHERE user_roles.user_id=?
		 ORDER BY roles.name`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("read roles: %w", err)
	}
	defer func() { _ = rows.Close() }()

	roles := []string{}
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, fmt.Errorf("scan role: %w", err)
		}
		roles = append(roles, role)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read roles: %w", err)
	}
	return roles, nil
}
