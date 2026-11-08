package vehicledata

import (
	"context"
	"database/sql"
	"fmt"

	"railkeeper/backend/internal/decoder"
	"railkeeper/backend/internal/decoder/workflow"
)

type configurationStore struct {
	db         *sql.DB
	sourceRoot string
}
type configurationMutation struct{ *snapshotCaptureMutation }

func NewConfigurationStore(db *sql.DB, sourceRoot string) workflow.ConfigurationRepository {
	return &configurationStore{db: db, sourceRoot: sourceRoot}
}
func (s *configurationStore) begin(ctx context.Context) (*configurationMutation, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &configurationMutation{snapshotCaptureMutation: &snapshotCaptureMutation{tx: tx, sourceRoot: s.sourceRoot}}, nil
}
func (s *configurationStore) BeginRead(ctx context.Context) (workflow.ConfigurationRead, error) {
	return s.begin(ctx)
}
func (s *configurationStore) BeginChange(ctx context.Context) (workflow.ConfigurationMutation, error) {
	return s.begin(ctx)
}
func (m *configurationMutation) Context(ctx context.Context, id string) (*decoder.DecoderConfigurationContext, error) {
	return ReadDecoderConfigurationContext(ctx, m.tx, id)
}
func (m *configurationMutation) DecoderWorkspaceExists(ctx context.Context, id, vehicleID, profile string) (bool, error) {
	return DecoderWorkspaceExists(ctx, m.tx, id, vehicleID, profile)
}
func (m *configurationMutation) DecoderSnapshotExists(ctx context.Context, id string) (bool, error) {
	return DecoderSnapshotExists(ctx, m.tx, id)
}
func (m *configurationMutation) WriteDecoderInstance(ctx context.Context, item decoder.DecoderInstance, create bool) error {
	return WriteDecoderInstance(ctx, m.tx, item, create)
}
func (m *configurationMutation) InsertDecoderInstallation(ctx context.Context, item decoder.DecoderInstallation) error {
	return InsertDecoderInstallation(ctx, m.tx, item)
}
func (m *configurationMutation) RemoveDecoderInstallation(ctx context.Context, item decoder.DecoderInstallation) error {
	return RemoveDecoderInstallation(ctx, m.tx, item)
}
func (m *configurationMutation) CorrectDecoderInstallationDates(ctx context.Context, item decoder.DecoderInstallation) error {
	return CorrectDecoderInstallationDates(ctx, m.tx, item)
}
func (m *configurationMutation) InsertDecoderWorkspace(ctx context.Context, item decoder.DecoderCVWorkspace) error {
	return InsertDecoderWorkspace(ctx, m.tx, item)
}
func (m *configurationMutation) AssignDecoderWorkspace(ctx context.Context, item decoder.DecoderCVWorkspace) error {
	return AssignDecoderWorkspace(ctx, m.tx, item)
}
func (m *configurationMutation) AuditConfiguration(ctx context.Context, record workflow.ConfigurationAudit) error {
	if _, err := m.tx.ExecContext(ctx, `INSERT INTO audit_logs
(id,actor_user_id,action,target_type,target_id,created_at,details_json)
VALUES(?,NULLIF(?,''),?,'decoder_configuration',?,?,?)`, record.ID, record.Actor, record.Action, record.TargetID, record.At,
		record.Details); err != nil {
		return fmt.Errorf("audit decoder configuration: %w", err)
	}
	return nil
}
