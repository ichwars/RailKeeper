package infrastructure

import (
	"context"
	"database/sql"

	"railkeeper/backend/internal/application"
)

func seedDecoderLibrary(ctx context.Context, tx *sql.Tx, now string) error {
	return application.SeedDecoderLibrary(ctx, tx, now)
}
