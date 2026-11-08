package exhibitiondata

import (
	"context"
	"database/sql"

	"railkeeper/backend/internal/exhibition"
)

type Store struct{ db *sql.DB }
type mutation struct{ tx *sql.Tx }

func New(db *sql.DB) *Store { return &Store{db: db} }
func (s *Store) Begin(ctx context.Context) (exhibition.Mutation, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &mutation{tx: tx}, nil
}
func (m *mutation) Commit() error    { return m.tx.Commit() }
func (m *mutation) Rollback() error  { return m.tx.Rollback() }
func changed(result sql.Result) bool { count, _ := result.RowsAffected(); return count != 0 }
func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

var _ exhibition.Repository = (*Store)(nil)
var _ exhibition.Mutation = (*mutation)(nil)
