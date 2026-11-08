package application

import (
	"context"
	"database/sql"

	"railkeeper/backend/internal/inventory/masterdata"

	"railkeeper/backend/internal/infrastructure/masterdatastore"
)

// SeedDecoderLibrary participates in the caller's seed or restore transaction.
func SeedDecoderLibrary(ctx context.Context, tx *sql.Tx, now string) error {
	return masterdata.SeedDecoderLibrary(ctx, masterdatastore.SeedStoreFor(tx), now)
}
