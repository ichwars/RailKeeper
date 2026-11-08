package settingsdata

import (
	"context"
	"database/sql"

	"railkeeper/backend/internal/configuration"
)

type mutation struct{ tx *sql.Tx }

func (s *Store) Begin(ctx context.Context) (configuration.Mutation, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &mutation{tx: tx}, nil
}
func (m *mutation) PutUser(ctx context.Context, userID, key, value, at string) error {
	_, err := m.tx.ExecContext(ctx, `
INSERT INTO user_settings(user_id, key, value, updated_at)
VALUES(?, ?, ?, ?)
ON CONFLICT(user_id, key) DO UPDATE SET value=excluded.value, updated_at=excluded.updated_at
`, userID, key, value, at)
	return err
}
func (m *mutation) PutApp(ctx context.Context, key, value, at string) error {
	_, err := m.tx.ExecContext(ctx, `
INSERT INTO app_settings(key, value, updated_at)
VALUES(?, ?, ?)
ON CONFLICT(key) DO UPDATE SET value=excluded.value, updated_at=excluded.updated_at
`, key, value, at)
	return err
}
func (m *mutation) Commit() error   { return m.tx.Commit() }
func (m *mutation) Rollback() error { return m.tx.Rollback() }
