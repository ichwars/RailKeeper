package application

import (
	"database/sql"

	"railkeeper/backend/internal/backup"
	"railkeeper/backend/internal/infrastructure/backupdata"
)

type BackupService = backup.Service
type BackupDocument = backup.BackupDocument
type BackupFile = backup.BackupFile
type BackupImportResult = backup.BackupImportResult
type BackupValidationResult = backup.BackupValidationResult
type BackupValidationTable = backup.BackupValidationTable

var ErrBackupInvalid = backup.ErrBackupInvalid
var ErrBackupPath = backup.ErrBackupPath

func NewBackupService(db *sql.DB, dataDir string) *BackupService {
	return backup.New(backupdata.NewStore(db, dataDir, BackupRestorePlanner()), BackupValidationRules())
}
func DecodeBackup(data []byte) (*BackupDocument, error) { return backup.DecodeBackup(data) }
func backupNumberValue(value any) (float64, bool)       { return backup.BackupNumberValue(value) }
func backupBooleanValue(value any) (bool, bool)         { return backup.BackupBooleanValue(value) }
func backupNonEmptyString(value any) (string, bool)     { return backup.BackupNonEmptyString(value) }
func validBackupNumber(value any) bool                  { return backup.ValidBackupNumber(value) }
func validBackupBoolean(value any) bool                 { return backup.ValidBackupBoolean(value) }

func validateBackupFilePath(value string) error { return backup.ValidateBackupFilePath(value) }
