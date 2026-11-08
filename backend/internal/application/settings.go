package application

import (
	"database/sql"

	"railkeeper/backend/internal/configuration"
	"railkeeper/backend/internal/infrastructure/settingsdata"
)

var ErrSettingsValidation = configuration.ErrSettingsValidation

type SettingsService = configuration.SettingsService
type SettingsPayload = configuration.SettingsPayload
type DigitalProviderSettings = configuration.DigitalProviderSettings
type DigitalCenterSettings = configuration.DigitalCenterSettings

func NewSettingsService(db *sql.DB) *SettingsService {
	return configuration.NewSettingsService(settingsdata.New(db))
}
