package vehicledata

import (
	"context"
	"database/sql"

	"railkeeper/backend/internal/decoder"
)

func ReadCVWorkspaceBinding(ctx context.Context, tx *sql.Tx, vehicleID, workspaceID string) (decoder.CVBinding, error) {
	result := decoder.CVBinding{WorkspaceID: workspaceID, VehicleID: vehicleID}
	err := tx.QueryRowContext(ctx, `SELECT w.profile_text,COALESCE(w.installation_id,''),COALESCE(i.decoder_id,''),COALESCE(i.state,'')
FROM decoder_cv_workspaces w LEFT JOIN decoder_installations i ON i.id=w.installation_id
WHERE w.id=? AND w.vehicle_id=?`, workspaceID, vehicleID).Scan(&result.Profile, &result.InstallationID, &result.DecoderID, &result.State)
	return result, err
}
