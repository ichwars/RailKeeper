package masterdatastore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"railkeeper/backend/internal/inventory/masterdata"
)

func (s *Store) SetActive(ctx context.Context, typeName, key string, active bool) error {
	result, err := s.db.ExecContext(ctx, `
UPDATE master_data_entries SET active=?, updated_at=? WHERE type=? AND key=?
`, boolToInt(active), time.Now().UTC().Format(time.RFC3339), typeName, key)
	if err != nil {
		return fmt.Errorf("set master data active state: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read master data active-state result: %w", err)
	}
	if affected == 0 {
		return masterdata.ErrMasterDataNotFound
	}
	return nil
}
func (s *Store) SetActiveMany(ctx context.Context, typeName string, keys []string, active bool) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin master data active-state batch: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := reserveMasterDataWriteTransaction(ctx, tx); err != nil {
		return err
	}

	updatedAt := time.Now().UTC().Format(time.RFC3339)
	for _, key := range keys {
		result, err := tx.ExecContext(ctx, `
UPDATE master_data_entries SET active=?, updated_at=? WHERE type=? AND key=?
`, boolToInt(active), updatedAt, typeName, key)
		if err != nil {
			return fmt.Errorf("set master data active-state batch: %w", err)
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("read master data active-state batch result: %w", err)
		}
		if affected == 0 {
			return fmt.Errorf("%w: %s/%s", masterdata.ErrMasterDataNotFound, typeName, key)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit master data active-state batch: %w", err)
	}

	return nil
}
func (s *Store) Delete(ctx context.Context, typeName, key string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin master data delete: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := reserveMasterDataWriteTransaction(ctx, tx); err != nil {
		return err
	}

	var entry masterdata.MasterDataEntry
	if err := tx.QueryRowContext(ctx, `
SELECT type, key, label, origin FROM master_data_entries WHERE type=? AND key=?
`, typeName, key).Scan(&entry.Type, &entry.Key, &entry.Label, &entry.Origin); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return masterdata.ErrMasterDataNotFound
		}
		return fmt.Errorf("load master data for deletion: %w", err)
	}
	if err := masterdata.ValidateDeletion(entry, false); err != nil {
		return err
	}

	used, err := masterDataIsUsed(ctx, tx, entry)
	if err != nil {
		return fmt.Errorf("check master data usage for %s/%s: %w", typeName, key, err)
	}
	if err := masterdata.ValidateDeletion(entry, used); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
DELETE FROM master_data_relations
WHERE (parent_type=? AND parent_key=?) OR (child_type=? AND child_key=?)
`, typeName, key, typeName, key); err != nil {
		return fmt.Errorf("delete master data relations: %w", err)
	}
	result, err := tx.ExecContext(ctx, `DELETE FROM master_data_entries WHERE type=? AND key=?`, typeName, key)
	if err != nil {
		return fmt.Errorf("delete master data: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read master data delete result: %w", err)
	}
	if affected == 0 {
		return masterdata.ErrMasterDataNotFound
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit master data delete: %w", err)
	}
	return nil
}
