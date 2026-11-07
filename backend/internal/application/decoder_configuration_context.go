package application

import (
	"context"
	"database/sql"
	"strings"

	"railkeeper/backend/internal/decoder"
	"railkeeper/backend/internal/infrastructure/vehicledata"
)

var (
	ErrDecoderValidation = decoder.ErrValidation
	ErrDecoderNotFound   = decoder.ErrNotFound
	ErrDecoderConflict   = decoder.ErrConflict
)

func (s *VehicleService) DecoderConfigurationContext(ctx context.Context, vehicleID string) (*DecoderConfigurationContext, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	return readDecoderConfigurationContextTx(ctx, tx, strings.TrimSpace(vehicleID))
}

func readDecoderConfigurationContextTx(ctx context.Context, tx *sql.Tx, vehicleID string) (*DecoderConfigurationContext, error) {
	result, err := vehicledata.ReadDecoderConfigurationContext(ctx, tx, vehicleID)
	return result, cvRuleError(err)
}

func readDecoderRows(ctx context.Context, tx *sql.Tx, statement string, args []any, scan func(*sql.Rows) error) error {
	return vehicledata.ReadDecoderRows(ctx, tx, statement, args, scan)
}

func readDecoderSnapshotSummaries(ctx context.Context, tx *sql.Tx, vehicleID string) ([]DecoderCVSnapshotSummary, error) {
	return vehicledata.ReadDecoderSnapshotSummaries(ctx, tx, vehicleID)
}
