package vehicledata

import (
	"context"
	"database/sql"
	"fmt"

	"railkeeper/backend/internal/decoder/workflow"
)

type workspaceWriter struct{ tx *sql.Tx }

func WorkspaceWriterFor(tx *sql.Tx) workflow.WorkspaceWriter { return &workspaceWriter{tx: tx} }
func (m *workspaceWriter) TouchWorkspace(ctx context.Context, record workflow.WorkspaceTouch) error {
	if _, err := m.tx.ExecContext(ctx, `INSERT INTO decoder_cv_workspaces
(id, vehicle_id, profile_text, origin, installation_id, created_at, updated_at)
VALUES(?,?,?,'manual',NULL,?,?)
ON CONFLICT(vehicle_id, profile_text) DO UPDATE SET updated_at=excluded.updated_at`,
		record.ID, record.VehicleID, record.Profile, record.At, record.At); err != nil {
		return fmt.Errorf("update decoder CV working group: %w", err)
	}
	return nil
}
