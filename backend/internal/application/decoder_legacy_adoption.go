package application

import (
	"context"
	"database/sql"

	"railkeeper/backend/internal/decoder/workflow"
	"railkeeper/backend/internal/infrastructure/vehicledata"
)

func BackfillLegacyDecoderCV(ctx context.Context, db *sql.DB, sourceRoot string) error {
	return workflow.BackfillLegacySnapshots(ctx, vehicledata.NewLegacySnapshotStore(db, sourceRoot))
}
