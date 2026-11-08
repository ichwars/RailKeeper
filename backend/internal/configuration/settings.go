package configuration

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type SettingsService struct{ repository Repository }

func NewSettingsService(repository Repository) *SettingsService {
	return &SettingsService{repository: repository}
}
func (s *SettingsService) configured() bool {
	return s != nil && s.repository != nil && s.repository.Configured()
}
func (s *SettingsService) UserSettings(ctx context.Context, userID string) (*SettingsPayload, error) {
	if !s.configured() {
		return &SettingsPayload{Settings: map[string]string{}}, nil
	}
	values, err := s.repository.UserSettings(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &SettingsPayload{Settings: values}, nil
}
func (s *SettingsService) appSettings(ctx context.Context, pattern string) (map[string]string, error) {
	if !s.configured() {
		return map[string]string{}, nil
	}
	return s.repository.AppSettings(ctx, pattern)
}
func (s *SettingsService) UpdateUserSettings(ctx context.Context, userID string, input SettingsPayload) (*SettingsPayload, error) {
	if !s.configured() {
		return nil, ErrSettingsValidation
	}
	now := time.Now().UTC().Format(time.RFC3339)
	tx, err := s.repository.Begin(ctx)
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
		if err = tx.PutUser(ctx, userID, key, value, now); err != nil {
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
	if !s.configured() {
		return nil, ErrSettingsValidation
	}
	input = normalizeDigitalSettings(input)
	if !validDigitalProvider(input.Provider) {
		return nil, ErrSettingsValidation
	}
	now := time.Now().UTC().Format(time.RFC3339)
	values := map[string]string{
		"digital.provider":            input.Provider,
		"digital.ecos.enabled":        strconv.FormatBool(input.ECoS.Enabled),
		"digital.ecos.host":           input.ECoS.Host,
		"digital.ecos.port":           input.ECoS.Port,
		"digital.z21.enabled":         strconv.FormatBool(input.Z21.Enabled),
		"digital.z21.host":            input.Z21.Host,
		"digital.z21.port":            input.Z21.Port,
		"digital.intellibox3.enabled": strconv.FormatBool(input.Intellibox3.Enabled),
		"digital.intellibox3.host":    input.Intellibox3.Host,
		"digital.intellibox3.port":    input.Intellibox3.Port,
		"digital.cs3.enabled":         strconv.FormatBool(input.CS3.Enabled),
		"digital.cs3.host":            input.CS3.Host,
		"digital.cs3.port":            input.CS3.Port,
	}
	tx, err := s.repository.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin digital settings update: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	for key, value := range values {
		if err = tx.PutApp(ctx, key, value, now); err != nil {
			return nil, fmt.Errorf("save digital setting %s: %w", key, err)
		}
	}
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit digital settings update: %w", err)
	}
	return s.DigitalSettings(ctx)
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
		Intellibox3: DigitalProviderSettings{
			Enabled: values["digital.intellibox3.enabled"] == "true",
			Host:    values["digital.intellibox3.host"],
			Port:    values["digital.intellibox3.port"],
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
	input.Intellibox3 = normalizeProviderSettings(input.Intellibox3, "21105")
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
	return provider == "ecos" || provider == "z21" || provider == "intellibox3" || provider == "cs3"
}

func valueOr(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
