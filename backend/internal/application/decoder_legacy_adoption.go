package application

import (
	"context"
	"database/sql"

	"railkeeper/backend/internal/decoder/workflow"
	"railkeeper/backend/internal/infrastructure/vehicledata"
)

type decoderWorkspaceRecord struct {
	ID, VehicleID, Profile string
	InstallationID         sql.NullString
}

func BackfillLegacyDecoderCV(ctx context.Context, db *sql.DB, sourceRoot string) error {
	return workflow.BackfillLegacySnapshots(ctx, vehicledata.NewLegacySnapshotStore(db, sourceRoot))
}
func backfillLegacyDecoderCVTx(ctx context.Context, tx *sql.Tx, sourceRoot string) error {
	return workflow.AdoptLegacySnapshotsWithin(ctx, vehicledata.LegacySnapshotMutationFor(tx, sourceRoot))
}
