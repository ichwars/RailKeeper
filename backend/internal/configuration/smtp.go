package configuration

import (
	"context"
	"fmt"
	"net/mail"
	"strconv"
	"strings"
	"time"
)

type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
	TLSMode  string
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

type SMTPSettingsService struct {
	repository        Repository
	fallbackConfig    SMTPConfig
	fallbackPublicURL string
}

func (s *SMTPSettingsService) configured() bool {
	return s != nil && s.repository != nil && s.repository.Configured()
}
func NewSMTPSettingsService(repository Repository, fallbackConfig SMTPConfig, fallbackPublicURL string) *SMTPSettingsService {
	fallbackConfig = cleanSMTPConfig(fallbackConfig)
	return &SMTPSettingsService{
		repository:        repository,
		fallbackConfig:    fallbackConfig,
		fallbackPublicURL: strings.TrimRight(strings.TrimSpace(fallbackPublicURL), "/"),
	}
}

func (s *SMTPSettingsService) Get(ctx context.Context) (*SMTPSettings, error) {
	config, publicURL, enabled, err := s.EffectiveConfig(ctx)
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
	if !s.configured() {
		return nil, ErrSMTPSettingsValidation
	}
	currentConfig, _, _, err := s.EffectiveConfig(ctx)
	if err != nil {
		return nil, err
	}

	config := SMTPConfig{
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
	tx, err := s.repository.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin smtp settings update: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	for key, value := range settings {
		if err = tx.PutApp(ctx, key, value, now); err != nil {
			return nil, fmt.Errorf("save smtp setting %s: %w", key, err)
		}
	}
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit smtp settings update: %w", err)
	}
	return s.Get(ctx)
}

func (s *SMTPSettingsService) EffectiveConfig(ctx context.Context) (SMTPConfig, string, bool, error) {
	if s == nil {
		return SMTPConfig{}, "", false, nil
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
	if !s.configured() {
		return config, publicURL, enabled, nil
	}

	values, err := s.repository.SMTPSettings(ctx)
	if err != nil {
		return config, publicURL, enabled, err
	}

	if len(values) == 0 {
		return config, publicURL, enabled, nil
	}

	config = SMTPConfig{
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

func cleanSMTPConfig(config SMTPConfig) SMTPConfig {
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

func validateSMTPSettings(enabled bool, config SMTPConfig, publicURL string) error {
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
