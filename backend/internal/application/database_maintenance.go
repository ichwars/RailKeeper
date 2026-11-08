package application

import (
	"database/sql"

	"railkeeper/backend/internal/configuration"
	"railkeeper/backend/internal/infrastructure/settingsdata"
)

type DatabaseMaintenanceService = configuration.DatabaseMaintenanceService
type DatabaseOptimizeResult = configuration.DatabaseOptimizeResult

func NewDatabaseMaintenanceService(db *sql.DB, dataDir string) *DatabaseMaintenanceService {
	return configuration.NewDatabaseMaintenanceService(settingsdata.NewMaintenance(db, dataDir))
}
