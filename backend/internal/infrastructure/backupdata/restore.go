package backupdata

import (
	"context"
	"errors"
	"fmt"
	"os"
)

func (s *Store) Restore(ctx context.Context, doc *BackupDocument) (*BackupImportResult, error) {
	if !s.Configured() {
		return nil, errors.New("backup service is not configured")
	}
	if doc == nil {
		return nil, ErrBackupInvalid
	}
	stagedFiles, err := s.stageRestoreFiles(doc.Files)
	if err != nil {
		return nil, err
	}
	defer func() { _ = os.RemoveAll(stagedFiles.root) }()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin backup restore: %w", err)
	}
	if _, err = tx.ExecContext(ctx, "PRAGMA defer_foreign_keys = ON"); err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("defer backup restore foreign keys: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	operation, err := s.planner.Begin(ctx, tx, doc)
	if err != nil {
		return nil, err
	}

	for i := len(backupTableOrder) - 1; i >= 0; i-- {
		table := backupTableOrder[i]
		if _, err = tx.ExecContext(ctx, "DELETE FROM "+QuoteIdentifier(table)); err != nil {
			return nil, fmt.Errorf("clear %s: %w", table, err)
		}
	}
	if doc.Profiles != nil {
		if _, err = tx.ExecContext(ctx, `DELETE FROM data_transfer_profiles`); err != nil {
			return nil, fmt.Errorf("clear data_transfer_profiles: %w", err)
		}
	}

	result := &BackupImportResult{}
	for _, table := range backupTableOrder {

		rows, err := operation.Rows(ctx, tx, table)
		if err != nil {
			return nil, err
		}
		if len(rows) == 0 {
			continue
		}
		columns, err := TableColumns(ctx, tx, table)
		if err != nil {
			return nil, err
		}
		for _, row := range rows {
			restoreRow := row
			inserted, err := InsertBackupRow(ctx, tx, table, columns, restoreRow)
			if err != nil {
				return nil, err
			}
			if inserted {
				result.RestoredRows++
			}
		}
		result.RestoredTables++
	}
	if len(doc.Profiles) > 0 {
		columns, err := TableColumns(ctx, tx, "data_transfer_profiles")
		if err != nil {
			return nil, err
		}
		for _, row := range doc.Profiles {
			restoreRow, err := operation.ProfileRow(ctx, tx, row)
			if err != nil {
				return nil, err
			}
			inserted, err := InsertBackupRow(ctx, tx, "data_transfer_profiles", columns, restoreRow)
			if err != nil {
				return nil, err
			}
			if inserted {
				result.RestoredRows++
			}
		}
		result.RestoredTables++
	}
	if err := operation.Finish(ctx, tx, stagedFiles.root); err != nil {
		return nil, err
	}

	uploadsSwap, err := s.replaceUploadsWithStaged(stagedFiles)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		uploadsSwap.rollback()
		return nil, fmt.Errorf("commit backup restore: %w", err)
	}
	committed = true
	uploadsSwap.cleanup()
	result.RestoredFiles = stagedFiles.restoredFiles

	return result, nil
}
