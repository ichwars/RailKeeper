package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type DatabaseMaintenanceService struct {
	db      *sql.DB
	dataDir string
}

type DatabaseOptimizeResult struct {
	BeforeBytes    int64  `json:"beforeBytes"`
	AfterBytes     int64  `json:"afterBytes"`
	ReclaimedBytes int64  `json:"reclaimedBytes"`
	OptimizedAt    string `json:"optimizedAt"`
}

func NewDatabaseMaintenanceService(db *sql.DB, dataDir string) *DatabaseMaintenanceService {
	return &DatabaseMaintenanceService{db: db, dataDir: dataDir}
}

func (s *DatabaseMaintenanceService) Optimize(ctx context.Context) (*DatabaseOptimizeResult, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("database maintenance service is not configured")
	}
	before, err := s.databaseFileBytes()
	if err != nil {
		return nil, err
	}
	if _, err := s.db.ExecContext(ctx, `PRAGMA wal_checkpoint(TRUNCATE)`); err != nil {
		return nil, fmt.Errorf("checkpoint database: %w", err)
	}
	if _, err := s.db.ExecContext(ctx, `VACUUM`); err != nil {
		return nil, fmt.Errorf("vacuum database: %w", err)
	}
	if _, err := s.db.ExecContext(ctx, `PRAGMA optimize`); err != nil {
		return nil, fmt.Errorf("optimize database: %w", err)
	}
	if _, err := s.db.ExecContext(ctx, `PRAGMA wal_checkpoint(TRUNCATE)`); err != nil {
		return nil, fmt.Errorf("final checkpoint database: %w", err)
	}
	after, err := s.databaseFileBytes()
	if err != nil {
		return nil, err
	}
	reclaimed := before - after
	if reclaimed < 0 {
		reclaimed = 0
	}
	return &DatabaseOptimizeResult{
		BeforeBytes:    before,
		AfterBytes:     after,
		ReclaimedBytes: reclaimed,
		OptimizedAt:    time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func (s *DatabaseMaintenanceService) databaseFileBytes() (int64, error) {
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
