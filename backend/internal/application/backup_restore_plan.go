package application

import "railkeeper/backend/internal/infrastructure/backupdata"

func BackupRestorePlanner() backupdata.RestorePlanner { return backupdata.NewCompatibilityPlanner() }
