package configuration

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type MaintenanceStore interface {
	Configured() bool
	DatabaseFileBytes() (int64, error)
	Checkpoint(context.Context) error
	Vacuum(context.Context) error
	OptimizeDatabase(context.Context) error
}
type DatabaseMaintenanceService struct{ store MaintenanceStore }

func NewDatabaseMaintenanceService(store MaintenanceStore) *DatabaseMaintenanceService {
	return &DatabaseMaintenanceService{store: store}
}
func (s *DatabaseMaintenanceService) configured() bool {
	return s != nil && s.store != nil && s.store.Configured()
}

type DatabaseOptimizeResult struct {
	BeforeBytes    int64  `json:"beforeBytes"`
	AfterBytes     int64  `json:"afterBytes"`
	ReclaimedBytes int64  `json:"reclaimedBytes"`
	OptimizedAt    string `json:"optimizedAt"`
}

func (s *DatabaseMaintenanceService) Optimize(ctx context.Context) (*DatabaseOptimizeResult, error) {
	if !s.configured() {
		return nil, errors.New("database maintenance service is not configured")
	}
	before, err := s.store.DatabaseFileBytes()
	if err != nil {
		return nil, err
	}
	if err := s.store.Checkpoint(ctx); err != nil {
		return nil, fmt.Errorf("checkpoint database: %w", err)
	}
	if err := s.store.Vacuum(ctx); err != nil {
		return nil, fmt.Errorf("vacuum database: %w", err)
	}
	if err := s.store.OptimizeDatabase(ctx); err != nil {
		return nil, fmt.Errorf("optimize database: %w", err)
	}
	if err := s.store.Checkpoint(ctx); err != nil {
		return nil, fmt.Errorf("final checkpoint database: %w", err)
	}
	after, err := s.store.DatabaseFileBytes()
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
