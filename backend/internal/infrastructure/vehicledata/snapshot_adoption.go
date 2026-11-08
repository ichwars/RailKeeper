package vehicledata

import (
	"context"
	"database/sql"
	"fmt"

	"railkeeper/backend/internal/decoder/workflow"
)

type legacySnapshotStore struct {
	db         *sql.DB
	sourceRoot string
}

func NewLegacySnapshotStore(db *sql.DB, sourceRoot string) workflow.LegacySnapshotRepository {
	return &legacySnapshotStore{db: db, sourceRoot: sourceRoot}
}
func (s *legacySnapshotStore) BeginLegacy(ctx context.Context) (workflow.LegacySnapshotMutation, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return LegacySnapshotMutationFor(tx, s.sourceRoot), nil
}
func (m *snapshotCaptureMutation) Commit() error   { return m.tx.Commit() }
func (m *snapshotCaptureMutation) Rollback() error { return m.tx.Rollback() }
func (m *snapshotCaptureMutation) AdoptLegacyGroups(ctx context.Context, now string) error {
	if _, err := m.tx.ExecContext(ctx, `INSERT OR IGNORE INTO decoder_cv_workspaces
(id, vehicle_id, profile_text, origin, installation_id, created_at, updated_at)
SELECT lower(hex(randomblob(16))), vehicle_id, profile_text, 'legacy', NULL, ?, ? FROM (
 SELECT vehicle_id, COALESCE(decoder_profile,'') AS profile_text FROM vehicle_cv_values
 UNION SELECT vehicle_id, COALESCE(decoder_profile,'') AS profile_text FROM vehicle_cv_files
)`, now, now); err != nil {
		return fmt.Errorf("adopt legacy decoder working groups: %w", err)
	}
	return nil
}
func (m *snapshotCaptureMutation) PendingLegacyWorkspaces(ctx context.Context) ([]workflow.SnapshotWorkspace, error) {
	rows, err := m.tx.QueryContext(ctx, `SELECT id, vehicle_id, profile_text, installation_id
FROM decoder_cv_workspaces w WHERE origin='legacy'
AND NOT EXISTS(SELECT 1 FROM decoder_cv_snapshots s WHERE s.workspace_id=w.id AND s.kind='legacy')
ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("list unadopted legacy decoder groups: %w", err)
	}
	defer func() { _ = rows.Close() }()
	workspaces := []workflow.SnapshotWorkspace{}
	for rows.Next() {
		var item workflow.SnapshotWorkspace
		var installation sql.NullString
		if err := rows.Scan(&item.ID, &item.VehicleID, &item.Profile, &installation); err != nil {
			return nil, fmt.Errorf("scan legacy decoder group: %w", err)
		}
		item.InstallationID = workflow.OptionalText{Text: installation.String, Valid: installation.Valid}
		workspaces = append(workspaces, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	return workspaces, nil
}
