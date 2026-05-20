package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/mail"
	"strconv"
	"strings"
	"time"
)

var ErrSMTPSettingsValidation = errors.New("smtp settings validation failed")

type SMTPSettingsService struct {
	db                *sql.DB
	fallbackConfig    SMTPPasswordResetMailConfig
	fallbackPublicURL string
}

type SMTPSettings struct {
	Enabled            bool   `json:"enabled"`
	PublicURL          string `json:"publicUrl"`
	Host               string `json:"host"`
	Port               string `json:"port"`
	Username           string `json:"username"`
	From               string `json:"from"`
	TLSMode            string `json:"tlsMode"`
	PasswordConfigured bool   `json:"passwordConfigured"`
}

type SMTPSettingsInput struct {
	Enabled       bool   `json:"enabled"`
	PublicURL     string `json:"publicUrl"`
	Host          string `json:"host"`
	Port          string `json:"port"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	From          string `json:"from"`
	TLSMode       string `json:"tlsMode"`
	ClearPassword bool   `json:"clearPassword"`
}

func NewSMTPSettingsService(db *sql.DB, fallbackConfig SMTPPasswordResetMailConfig, fallbackPublicURL string) *SMTPSettingsService {
	fallbackConfig = cleanSMTPConfig(fallbackConfig)
	return &SMTPSettingsService{
		db:                db,
		fallbackConfig:    fallbackConfig,
		fallbackPublicURL: strings.TrimRight(strings.TrimSpace(fallbackPublicURL), "/"),
	}
}

func (s *SMTPSettingsService) Get(ctx context.Context) (*SMTPSettings, error) {
	config, publicURL, enabled, err := s.effectiveConfig(ctx)
	if err != nil {
		return nil, err
	}
	return &SMTPSettings{
		Enabled:            enabled,
		PublicURL:          publicURL,
		Host:               config.Host,
		Port:               config.Port,
		Username:           config.Username,
		From:               config.From,
		TLSMode:            config.TLSMode,
		PasswordConfigured: config.Password != "",
	}, nil
}

func (s *SMTPSettingsService) Update(ctx context.Context, input SMTPSettingsInput) (*SMTPSettings, error) {
	if s == nil || s.db == nil {
		return nil, ErrSMTPSettingsValidation
	}
	currentConfig, _, _, err := s.effectiveConfig(ctx)
	if err != nil {
		return nil, err
	}

	config := SMTPPasswordResetMailConfig{
		Host:     input.Host,
		Port:     input.Port,
		Username: input.Username,
		Password: input.Password,
		From:     input.From,
		TLSMode:  input.TLSMode,
	}
	config = cleanSMTPConfig(config)
	publicURL := strings.TrimRight(strings.TrimSpace(input.PublicURL), "/")
	if config.Password == "" && !input.ClearPassword {
		config.Password = currentConfig.Password
	}
	if input.ClearPassword {
		config.Password = ""
	}
	if err := validateSMTPSettings(input.Enabled, config, publicURL); err != nil {
		return nil, err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	settings := map[string]string{
		"smtp.enabled":    strconv.FormatBool(input.Enabled),
		"smtp.public_url": publicURL,
		"smtp.host":       config.Host,
		"smtp.port":       config.Port,
		"smtp.username":   config.Username,
		"smtp.password":   config.Password,
		"smtp.from":       config.From,
		"smtp.tls_mode":   config.TLSMode,
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin smtp settings update: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	for key, value := range settings {
		if _, err = tx.ExecContext(ctx, `
INSERT INTO app_settings(key, value, updated_at)
VALUES(?, ?, ?)
ON CONFLICT(key) DO UPDATE SET value=excluded.value, updated_at=excluded.updated_at
`, key, value, now); err != nil {
			return nil, fmt.Errorf("save smtp setting %s: %w", key, err)
		}
	}
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit smtp settings update: %w", err)
	}
	return s.Get(ctx)
}

func (s *SMTPSettingsService) EffectiveMailer(ctx context.Context) (*SMTPPasswordResetMailer, string, error) {
	if s == nil {
		return nil, "", nil
	}
	config, publicURL, enabled, err := s.effectiveConfig(ctx)
	if err != nil {
		return nil, publicURL, err
	}
	if !enabled {
		return nil, publicURL, nil
	}
	mailer, err := NewSMTPPasswordResetMailer(config)
	if err != nil {
		return nil, publicURL, err
	}
	return mailer, publicURL, nil
}

func (s *SMTPSettingsService) effectiveConfig(ctx context.Context) (SMTPPasswordResetMailConfig, string, bool, error) {
	if s == nil {
		return SMTPPasswordResetMailConfig{}, "", false, nil
	}
	config := s.fallbackConfig
	publicURL := s.fallbackPublicURL
	enabled := config.Host != "" && config.From != ""
	if config.Port == "" {
		config.Port = "587"
	}
	if config.TLSMode == "" {
		config.TLSMode = "starttls"
	}
	if s == nil || s.db == nil {
		return config, publicURL, enabled, nil
	}

	rows, err := s.db.QueryContext(ctx, `SELECT key, value FROM app_settings WHERE key LIKE 'smtp.%'`)
	if err != nil {
		return config, publicURL, enabled, fmt.Errorf("load smtp settings: %w", err)
	}
	defer func() { _ = rows.Close() }()

	values := map[string]string{}
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return config, publicURL, enabled, fmt.Errorf("scan smtp settings: %w", err)
		}
		values[key] = value
	}
	if err := rows.Err(); err != nil {
		return config, publicURL, enabled, fmt.Errorf("iterate smtp settings: %w", err)
	}
	if len(values) == 0 {
		return config, publicURL, enabled, nil
	}

	config = SMTPPasswordResetMailConfig{
		Host:     values["smtp.host"],
		Port:     values["smtp.port"],
		Username: values["smtp.username"],
		Password: values["smtp.password"],
		From:     values["smtp.from"],
		TLSMode:  values["smtp.tls_mode"],
	}
	config = cleanSMTPConfig(config)
	publicURL = strings.TrimRight(strings.TrimSpace(values["smtp.public_url"]), "/")
	enabled = values["smtp.enabled"] == "true"
	return config, publicURL, enabled, nil
}

func cleanSMTPConfig(config SMTPPasswordResetMailConfig) SMTPPasswordResetMailConfig {
	config.Host = strings.TrimSpace(config.Host)
	config.Port = strings.TrimSpace(config.Port)
	config.Username = strings.TrimSpace(config.Username)
	config.From = strings.TrimSpace(config.From)
	config.TLSMode = strings.ToLower(strings.TrimSpace(config.TLSMode))
	if config.Port == "" {
		config.Port = "587"
	}
	if config.TLSMode == "" {
		config.TLSMode = "starttls"
	}
	return config
}

func validateSMTPSettings(enabled bool, config SMTPPasswordResetMailConfig, publicURL string) error {
	switch config.TLSMode {
	case "starttls", "implicit", "none":
	default:
		return fmt.Errorf("%w: unsupported TLS mode", ErrSMTPSettingsValidation)
	}
	if _, err := strconv.Atoi(config.Port); err != nil {
		return fmt.Errorf("%w: invalid SMTP port", ErrSMTPSettingsValidation)
	}
	if !enabled {
		return nil
	}
	if config.Host == "" || config.From == "" || publicURL == "" {
		return fmt.Errorf("%w: host, sender and public URL are required", ErrSMTPSettingsValidation)
	}
	if _, err := mail.ParseAddress(config.From); err != nil {
		return fmt.Errorf("%w: invalid sender address", ErrSMTPSettingsValidation)
	}
	return nil
}
