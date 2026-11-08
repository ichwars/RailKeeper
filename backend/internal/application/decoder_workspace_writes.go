package application

import (
	"context"
	"database/sql"

	"railkeeper/backend/internal/decoder/workflow"
	"railkeeper/backend/internal/infrastructure/vehicledata"
)

func touchDecoderCVWorkspaceTx(ctx context.Context, tx *sql.Tx, id, profile, now string) error {
	return workflow.TouchWorkspace(ctx, vehicledata.WorkspaceWriterFor(tx), id, profile, now)
}
