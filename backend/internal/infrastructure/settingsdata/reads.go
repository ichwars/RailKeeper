package settingsdata

import (
	"context"
	"database/sql"
	"fmt"

	"railkeeper/backend/internal/configuration"
)

type Store struct{ db *sql.DB }

func New(db *sql.DB) *Store       { return &Store{db: db} }
func (s *Store) Configured() bool { return s != nil && s.db != nil }

var _ configuration.Repository = (*Store)(nil)

func (s *Store) UserSettings(ctx context.Context, userID string) (map[string]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT key, value FROM user_settings WHERE user_id=?`, userID)
	if err != nil {
		return nil, fmt.Errorf("load user settings: %w", err)
	}
	defer func() { _ = rows.Close() }()
	settings := map[string]string{}
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf("scan user settings: %w", err)
		}
		settings[key] = value
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate user settings: %w", err)
	}
	return settings, nil
}

func (s *Store) AppSettings(ctx context.Context, pattern string) (map[string]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT key, value FROM app_settings WHERE key LIKE ?`, pattern)
	if err != nil {
		return nil, fmt.Errorf("load app settings: %w", err)
	}
	defer func() { _ = rows.Close() }()
	values := map[string]string{}
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf("scan app settings: %w", err)
		}
		values[key] = value
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate app settings: %w", err)
	}
	return values, nil
}

func (s *Store) SMTPSettings(ctx context.Context) (map[string]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT key, value FROM app_settings WHERE key LIKE 'smtp.%'`)
	if err != nil {
		return nil, fmt.Errorf("load smtp settings: %w", err)
	}
	defer func() { _ = rows.Close() }()

	values := map[string]string{}
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf("scan smtp settings: %w", err)
		}
		values[key] = value
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate smtp settings: %w", err)
	}
	return values, nil
}
