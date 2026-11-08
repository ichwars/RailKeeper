package workflow

import (
	"context"

	"railkeeper/backend/internal/decoder"
)

type ConfigurationRead interface {
	Context(context.Context, string) (*decoder.DecoderConfigurationContext, error)
	Rollback() error
}
type ConfigurationMutation interface {
	ConfigurationRead
	SnapshotCapture
	WriteDecoderInstance(context.Context, decoder.DecoderInstance, bool) error
	InsertDecoderInstallation(context.Context, decoder.DecoderInstallation) error
	RemoveDecoderInstallation(context.Context, decoder.DecoderInstallation) error
	CorrectDecoderInstallationDates(context.Context, decoder.DecoderInstallation) error
	DecoderWorkspaceExists(context.Context, string, string, string) (bool, error)
	InsertDecoderWorkspace(context.Context, decoder.DecoderCVWorkspace) error
	AssignDecoderWorkspace(context.Context, decoder.DecoderCVWorkspace) error
	DecoderSnapshotExists(context.Context, string) (bool, error)
	AuditConfiguration(context.Context, ConfigurationAudit) error
	Commit() error
}
type ConfigurationAudit struct{ ID, Actor, Action, TargetID, At, Details string }
type ConfigurationRepository interface {
	BeginRead(context.Context) (ConfigurationRead, error)
	BeginChange(context.Context) (ConfigurationMutation, error)
}
type ConfigurationService struct{ repository ConfigurationRepository }

func NewConfigurationService(repository ConfigurationRepository) *ConfigurationService {
	return &ConfigurationService{repository: repository}
}

type WorkspaceTouch struct{ ID, VehicleID, Profile, At string }
type WorkspaceWriter interface {
	TouchWorkspace(context.Context, WorkspaceTouch) error
}
