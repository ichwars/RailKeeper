package settingsdata

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"railkeeper/backend/internal/configuration"
)

type maintenanceStore struct {
	db      *sql.DB
	dataDir string
}

func NewMaintenance(db *sql.DB, dataDir string) configuration.MaintenanceStore {
	return &maintenanceStore{db: db, dataDir: dataDir}
}
func (s *maintenanceStore) Configured() bool { return s != nil && s.db != nil }
func (s *maintenanceStore) Checkpoint(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `PRAGMA wal_checkpoint(TRUNCATE)`)
	return err
}
func (s *maintenanceStore) Vacuum(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `VACUUM`)
	return err
}
func (s *maintenanceStore) OptimizeDatabase(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `PRAGMA optimize`)
	return err
}
func (s *maintenanceStore) DatabaseFileBytes() (int64, error) {
	root, err := os.Stat(s.dataDir)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return 0, fmt.Errorf("read database directory: %w", err)
	}
	if err == nil && !root.IsDir() {
		return 0, errors.New("read database directory: data path is not a directory")
	}
	var total int64
	for _, suffix := range []string{"", "-wal", "-shm"} {
		info, err := os.Stat(filepath.Join(s.dataDir, "railkeeper.db"+suffix))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return 0, fmt.Errorf("read database size: %w", err)
		}
		total += info.Size()
	}
	return total, nil
}
