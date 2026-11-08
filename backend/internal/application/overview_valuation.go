package application

import (
	"database/sql"
	"railkeeper/backend/internal/infrastructure/overviewdata"
	"railkeeper/backend/internal/inventory/overview"
)

type OverviewValuation = overview.OverviewValuation
type OverviewValuationService = overview.Service

func NewOverviewValuationService(db *sql.DB) *OverviewValuationService {
	if db == nil {
		return overview.New(nil)
	}
	return overview.New(overviewdata.New(db))
}
