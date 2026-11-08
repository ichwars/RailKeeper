package masterdatastore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"railkeeper/backend/internal/inventory/masterdata"
)

type seedStore struct{ tx *sql.Tx }

// SeedStoreFor participates in the caller's seed/restore transaction without owning its commit.
func SeedStoreFor(tx *sql.Tx) masterdata.SeedStore { return &seedStore{tx: tx} }

func (s *seedStore) DecoderSeedState(ctx context.Context, typeName, key string) (masterdata.SeedState, bool, error) {
	var state masterdata.SeedState
	err := s.tx.QueryRowContext(ctx, `SELECT metadata_json,origin FROM master_data_entries WHERE type=? AND key=?`, typeName, key).
		Scan(&state.MetadataJSON, &state.Origin)
	if errors.Is(err, sql.ErrNoRows) {
		return state, false, nil
	}
	if err != nil {
		return state, false, fmt.Errorf("read decoder library seed: %w", err)
	}
	return state, true, nil
}
func (s *seedStore) UpgradeDecoderSeed(ctx context.Context, change masterdata.SeedMetadataChange) error {
	_, err := s.tx.ExecContext(ctx, `UPDATE master_data_entries SET metadata_json=?,updated_at=?
 WHERE type=? AND key=? AND metadata_json=? AND origin='bundled'`, change.MetadataJSON, change.At, change.Type, change.Key, change.PreviousJSON)
	if err != nil {
		return fmt.Errorf("upgrade bundled decoder documentation: %w", err)
	}
	return nil
}
func (s *seedStore) InsertDecoderSeed(ctx context.Context, entry masterdata.DecoderSeedEntry) error {
	_, err := s.tx.ExecContext(ctx, `INSERT INTO master_data_entries
 (id,type,key,label,active,sort_order,source_url,metadata_json,created_at,updated_at,origin)
 VALUES(?,?,?,?,1,?,?,?,?,?,'bundled')`, entry.Type+":"+entry.Key, entry.Type, entry.Key, entry.Label, entry.SortOrder, entry.SourceURL, entry.MetadataJSON, entry.At, entry.At)
	if err != nil {
		return fmt.Errorf("seed decoder documentation: %w", err)
	}
	return nil
}
func (s *seedStore) FunctionSymbolSeedState(ctx context.Context, key string) (masterdata.SeedState, bool, error) {
	var state masterdata.SeedState
	err := s.tx.QueryRowContext(ctx, `SELECT metadata_json,origin FROM master_data_entries
 WHERE type='symbols' AND key=?`, key).Scan(&state.MetadataJSON, &state.Origin)
	if errors.Is(err, sql.ErrNoRows) {
		return state, false, nil
	}
	if err != nil {
		return state, false, fmt.Errorf("read bundled function symbol code: %w", err)
	}
	return state, true, nil
}
func (s *seedStore) CorrectFunctionSymbolSeed(ctx context.Context, change masterdata.SeedMetadataChange) error {
	if _, err := s.tx.ExecContext(ctx, `UPDATE master_data_entries SET metadata_json=?,updated_at=?
 WHERE type='symbols' AND key=? AND origin='bundled' AND metadata_json=?`, change.MetadataJSON, change.At, change.Key, change.PreviousJSON); err != nil {
		return fmt.Errorf("correct bundled function symbol code: %w", err)
	}
	return nil
}
