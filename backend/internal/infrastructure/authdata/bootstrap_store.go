package authdata

import (
	"context"
	"database/sql"
	"errors"

	"railkeeper/backend/internal/authentication/bootstrap"
)

type setupStore struct{ db *sql.DB }
type setupMutation struct{ tx *sql.Tx }

func NewSetupStore(db *sql.DB) bootstrap.Repository { return &setupStore{db: db} }
func (s *setupStore) UserCount(ctx context.Context) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&count)
	return count, err
}
func (m *setupMutation) UserCount(ctx context.Context) (int, error) {
	var count int
	err := m.tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&count)
	return count, err
}
func (s *setupStore) PendingAuthorization(ctx context.Context) (string, bool, error) {
	var hash string
	err := s.db.QueryRowContext(ctx, `SELECT token_hash FROM setup_authorizations
WHERE id=1 AND consumed_at IS NULL`).Scan(&hash)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	return hash, err == nil, err
}
func (s *setupStore) Begin(ctx context.Context) (bootstrap.Mutation, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &setupMutation{tx: tx}, nil
}
func (m *setupMutation) ReplaceAuthorization(ctx context.Context, hash, at string) error {
	_, err := m.tx.ExecContext(ctx, `INSERT INTO setup_authorizations(id,token_hash,created_at)
VALUES(1,?,?) ON CONFLICT(id) DO UPDATE SET token_hash=excluded.token_hash,
created_at=excluded.created_at,consumed_at=NULL`, hash, at)
	return err
}
func (m *setupMutation) ConsumeAuthorization(ctx context.Context, hash, at string) (bool, error) {
	result, err := m.tx.ExecContext(ctx, `UPDATE setup_authorizations SET consumed_at=?
WHERE id=1 AND token_hash=? AND consumed_at IS NULL`, at, hash)
	if err != nil {
		return false, err
	}
	count, err := result.RowsAffected()
	return count == 1, err
}
func (m *setupMutation) Commit() error   { return m.tx.Commit() }
func (m *setupMutation) Rollback() error { return m.tx.Rollback() }
