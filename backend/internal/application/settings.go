package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var ErrSettingsValidation = errors.New("settings validation failed")

type SettingsService struct {
	db *sql.DB
}

type SettingsPayload struct {
	Settings map[string]string `json:"settings"`
}

type DigitalProviderSettings struct {
	Enabled bool   `json:"enabled"`
	Host    string `json:"host"`
	Port    string `json:"port"`
}

type DigitalCenterSettings struct {
	Provider string                  `json:"provider"`
	ECoS     DigitalProviderSettings `json:"ecos"`
	Z21      DigitalProviderSettings `json:"z21"`
	CS3      DigitalProviderSettings `json:"cs3"`
}

func NewSettingsService(db *sql.DB) *SettingsService {
	return &SettingsService{db: db}
}

func (s *SettingsService) UserSettings(ctx context.Context, userID string) (*SettingsPayload, error) {
	if s == nil || s.db == nil {
		return &SettingsPayload{Settings: map[string]string{}}, nil
	}
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
	return &SettingsPayload{Settings: settings}, nil
}

func (s *SettingsService) UpdateUserSettings(ctx context.Context, userID string, input SettingsPayload) (*SettingsPayload, error) {
	if s == nil || s.db == nil {
		return nil, ErrSettingsValidation
	}
	now := time.Now().UTC().Format(time.RFC3339)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin user settings update: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	for key, value := range input.Settings {
		key = strings.TrimSpace(key)
		if key == "" || len(key) > 160 || len(value) > 8000 {
			err = ErrSettingsValidation
			return nil, err
		}
		if _, err = tx.ExecContext(ctx, `
INSERT INTO user_settings(user_id, key, value, updated_at)
VALUES(?, ?, ?, ?)
ON CONFLICT(user_id, key) DO UPDATE SET value=excluded.value, updated_at=excluded.updated_at
`, userID, key, value, now); err != nil {
			return nil, fmt.Errorf("save user setting %s: %w", key, err)
		}
	}
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit user settings update: %w", err)
	}
	return s.UserSettings(ctx, userID)
}

func (s *SettingsService) DigitalSettings(ctx context.Context) (*DigitalCenterSettings, error) {
	values, err := s.appSettings(ctx, "digital.%")
	if err != nil {
		return nil, err
	}
	return digitalSettingsFromValues(values), nil
}

func (s *SettingsService) UpdateDigitalSettings(ctx context.Context, input DigitalCenterSettings) (*DigitalCenterSettings, error) {
	if s == nil || s.db == nil {
		return nil, ErrSettingsValidation
	}
	input = normalizeDigitalSettings(input)
	if !validDigitalProvider(input.Provider) {
		return nil, ErrSettingsValidation
	}
	now := time.Now().UTC().Format(time.RFC3339)
	values := map[string]string{
		"digital.provider":     input.Provider,
		"digital.ecos.enabled": strconv.FormatBool(input.ECoS.Enabled),
		"digital.ecos.host":    input.ECoS.Host,
		"digital.ecos.port":    input.ECoS.Port,
		"digital.z21.enabled":  strconv.FormatBool(input.Z21.Enabled),
		"digital.z21.host":     input.Z21.Host,
		"digital.z21.port":     input.Z21.Port,
		"digital.cs3.enabled":  strconv.FormatBool(input.CS3.Enabled),
		"digital.cs3.host":     input.CS3.Host,
		"digital.cs3.port":     input.CS3.Port,
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin digital settings update: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	for key, value := range values {
		if _, err = tx.ExecContext(ctx, `
INSERT INTO app_settings(key, value, updated_at)
VALUES(?, ?, ?)
ON CONFLICT(key) DO UPDATE SET value=excluded.value, updated_at=excluded.updated_at
`, key, value, now); err != nil {
			return nil, fmt.Errorf("save digital setting %s: %w", key, err)
		}
	}
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit digital settings update: %w", err)
	}
	return s.DigitalSettings(ctx)
}

func (s *SettingsService) appSettings(ctx context.Context, pattern string) (map[string]string, error) {
	if s == nil || s.db == nil {
		return map[string]string{}, nil
	}
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

func digitalSettingsFromValues(values map[string]string) *DigitalCenterSettings {
	settings := normalizeDigitalSettings(DigitalCenterSettings{
		Provider: valueOr(values["digital.provider"], "ecos"),
		ECoS: DigitalProviderSettings{
			Enabled: values["digital.ecos.enabled"] == "true",
			Host:    values["digital.ecos.host"],
			Port:    values["digital.ecos.port"],
		},
		Z21: DigitalProviderSettings{
			Enabled: values["digital.z21.enabled"] == "true",
			Host:    values["digital.z21.host"],
			Port:    values["digital.z21.port"],
		},
		CS3: DigitalProviderSettings{
			Enabled: values["digital.cs3.enabled"] == "true",
			Host:    values["digital.cs3.host"],
			Port:    values["digital.cs3.port"],
		},
	})
	return &settings
}

func normalizeDigitalSettings(input DigitalCenterSettings) DigitalCenterSettings {
	input.Provider = strings.ToLower(strings.TrimSpace(input.Provider))
	if input.Provider == "" {
		input.Provider = "ecos"
	}
	input.ECoS = normalizeProviderSettings(input.ECoS, "15471")
	input.Z21 = normalizeProviderSettings(input.Z21, "21105")
	input.CS3 = normalizeProviderSettings(input.CS3, "80")
	return input
}

func normalizeProviderSettings(input DigitalProviderSettings, defaultPort string) DigitalProviderSettings {
	input.Host = strings.TrimSpace(input.Host)
	input.Port = strings.TrimSpace(input.Port)
	if input.Port == "" {
		input.Port = defaultPort
	}
	return input
}

func validDigitalProvider(provider string) bool {
	return provider == "ecos" || provider == "z21" || provider == "cs3"
}

func valueOr(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
