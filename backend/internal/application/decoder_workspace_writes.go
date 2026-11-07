package application

import (
	"context"
	"database/sql"
	"fmt"
)

// Working CV rows remain canonical. Recording a checkpoint is a separate, explicit action.
func touchDecoderCVWorkspaceTx(ctx context.Context, tx *sql.Tx, vehicleID, profile, now string) error {
	if _, err := tx.ExecContext(ctx, `INSERT INTO decoder_cv_workspaces
(id, vehicle_id, profile_text, origin, installation_id, created_at, updated_at)
VALUES(?,?,?,'manual',NULL,?,?)
ON CONFLICT(vehicle_id, profile_text) DO UPDATE SET updated_at=excluded.updated_at`,
		randomID(), vehicleID, profile, now, now); err != nil {
		return fmt.Errorf("update decoder CV working group: %w", err)
	}
	return nil
}
