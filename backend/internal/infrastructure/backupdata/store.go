package backupdata

import (
	"context"
	"database/sql"
	"fmt"

	"railkeeper/backend/internal/backup"
)

type Store struct {
	db      *sql.DB
	dataDir string
	planner RestorePlanner
}

func NewStore(db *sql.DB, dataDir string, planner RestorePlanner) *Store {
	return &Store{db: db, dataDir: dataDir, planner: planner}
}
func (s *Store) Configured() bool { return s != nil && s.db != nil && s.planner != nil }

var _ backup.Store = (*Store)(nil)

type BackupDocument = backup.BackupDocument
type BackupFile = backup.BackupFile
type BackupImportResult = backup.BackupImportResult

var ErrBackupInvalid = backup.ErrBackupInvalid
var ErrBackupPath = backup.ErrBackupPath
var backupTableOrder = backup.TableOrder()

type RestorePlanner interface {
	Begin(context.Context, *sql.Tx, *BackupDocument) (RestoreOperation, error)
}
type RestoreOperation interface {
	Rows(context.Context, *sql.Tx, string) ([]map[string]any, error)
	ProfileRow(context.Context, *sql.Tx, map[string]any) (map[string]any, error)
	Finish(context.Context, *sql.Tx, string) error
}

func (s *Store) Schema(ctx context.Context, tables []string) (map[string]map[string]bool, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, fmt.Errorf("begin backup validation: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	result := map[string]map[string]bool{}
	for _, table := range tables {
		columns, err := TableColumns(ctx, tx, table)
		if err != nil {
			return nil, err
		}
		result[table] = columns
	}
	return result, nil
}

func (s *Store) Snapshot(ctx context.Context) (backup.Snapshot, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return backup.Snapshot{}, fmt.Errorf("begin backup export: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	snapshot := backup.Snapshot{Tables: map[string][]map[string]any{}, Profiles: []map[string]any{}}
	for _, table := range backupTableOrder {
		rows, err := s.exportTable(ctx, tx, table)
		if err != nil {
			return backup.Snapshot{}, err
		}
		snapshot.Tables[table] = rows
	}
	snapshot.Profiles, err = s.exportDataTransferProfiles(ctx, tx)
	if err != nil {
		return backup.Snapshot{}, err
	}
	if err := tx.Commit(); err != nil {
		return backup.Snapshot{}, fmt.Errorf("commit backup export: %w", err)
	}
	return snapshot, nil
}
