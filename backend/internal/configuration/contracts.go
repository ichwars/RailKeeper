package configuration

import (
	"context"
	"errors"
)

var ErrSettingsValidation = errors.New("settings validation failed")
var ErrSMTPSettingsValidation = errors.New("smtp settings validation failed")

type Repository interface {
	Configured() bool
	UserSettings(context.Context, string) (map[string]string, error)
	AppSettings(context.Context, string) (map[string]string, error)
	SMTPSettings(context.Context) (map[string]string, error)
	Begin(context.Context) (Mutation, error)
}
type Mutation interface {
	PutUser(context.Context, string, string, string, string) error
	PutApp(context.Context, string, string, string) error
	Commit() error
	Rollback() error
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
	Provider    string                  `json:"provider"`
	ECoS        DigitalProviderSettings `json:"ecos"`
	Z21         DigitalProviderSettings `json:"z21"`
	Intellibox3 DigitalProviderSettings `json:"intellibox3"`
	CS3         DigitalProviderSettings `json:"cs3"`
}
