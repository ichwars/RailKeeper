package application

import (
	"context"
	"database/sql"

	"railkeeper/backend/internal/configuration"
	"railkeeper/backend/internal/infrastructure/settingsdata"
)

var ErrSMTPSettingsValidation = configuration.ErrSMTPSettingsValidation

type SMTPSettings = configuration.SMTPSettings
type SMTPSettingsInput = configuration.SMTPSettingsInput
type SMTPSettingsService struct {
	core *configuration.SMTPSettingsService
}

func NewSMTPSettingsService(db *sql.DB, fallbackConfig SMTPPasswordResetMailConfig, fallbackPublicURL string) *SMTPSettingsService {
	return &SMTPSettingsService{core: configuration.NewSMTPSettingsService(settingsdata.New(db), fallbackConfig, fallbackPublicURL)}
}
func (s *SMTPSettingsService) configuration() *configuration.SMTPSettingsService {
	if s == nil {
		return nil
	}
	return s.core
}
func (s *SMTPSettingsService) Get(ctx context.Context) (*SMTPSettings, error) {
	return s.configuration().Get(ctx)
}
func (s *SMTPSettingsService) Update(ctx context.Context, input SMTPSettingsInput) (*SMTPSettings, error) {
	return s.configuration().Update(ctx, input)
}
func (s *SMTPSettingsService) EffectiveMailer(ctx context.Context) (*SMTPPasswordResetMailer, string, error) {
	if s == nil {
		return nil, "", nil
	}
	config, publicURL, enabled, err := s.configuration().EffectiveConfig(ctx)
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
