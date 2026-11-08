package authdata

import (
	"context"
	"fmt"

	"railkeeper/backend/internal/authentication/bootstrap"
)

func (m *setupMutation) InsertFirstAdministrator(ctx context.Context, record bootstrap.Administrator) (int64, error) {
	result, err := m.tx.ExecContext(ctx, `INSERT INTO users(id, username, email, password_hash, created_at)
		 SELECT ?, ?, ?, ?, ?
		  WHERE NOT EXISTS (SELECT 1 FROM users)`, record.ID, record.Username, record.Email, record.PasswordHash, record.At)
	if err != nil {
		return 0, fmt.Errorf("insert admin user: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("read setup insert result: %w", err)
	}
	return count, nil
}
func (m *setupMutation) AssignAdministratorRoles(ctx context.Context, userID string) error {
	_, err := m.tx.ExecContext(ctx, `INSERT INTO user_roles(user_id, role_id)
		 SELECT ?, id FROM roles WHERE name IN ('Admin', 'Editor', 'Viewer')`, userID)
	return err
}
func (m *setupMutation) AuditAdministrator(ctx context.Context, record bootstrap.AdministratorAudit) error {
	_, err := m.tx.ExecContext(ctx, `INSERT INTO audit_logs(id, actor_user_id, action, target_type, target_id, created_at, details_json)
		 VALUES(?, ?, 'SetupAdminCreated', 'user', ?, ?, '{}')`, record.ID, record.UserID, record.UserID, record.At)
	return err
}
