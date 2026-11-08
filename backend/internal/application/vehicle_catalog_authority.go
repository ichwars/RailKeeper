package application

import (
	"context"
	"database/sql"

	"railkeeper/backend/internal/infrastructure/catalogdata"
)

func suspendVehicleCatalogProjectionTx(ctx context.Context, tx *sql.Tx) error {
	return catalogdata.SuspendProjection(ctx, tx)
}
func restoreVehicleCatalogProjectionTx(ctx context.Context, tx *sql.Tx, legacy bool) error {
	return catalogdata.RestoreProjection(ctx, tx, legacy)
}
