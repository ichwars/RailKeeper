package backup

import (
	"context"
	"errors"
	"fmt"
)

func (s *Service) Validate(ctx context.Context, doc *BackupDocument) (*BackupValidationResult, error) {
	if !s.configured() {
		return nil, errors.New("backup service is not configured")
	}

	result := &BackupValidationResult{
		Tables:   []BackupValidationTable{},
		Warnings: []string{},
		Errors:   []string{},
	}
	if doc == nil {
		result.Errors = append(result.Errors, "Backup-Dokument fehlt.")
		return FinishBackupValidation(result), nil
	}

	result.Format = doc.Format
	result.Version = doc.Version
	result.CreatedAt = doc.CreatedAt
	result.FileCount = len(doc.Files)
	for _, file := range doc.Files {
		result.FileBytes += file.SizeBytes
	}
	if doc.Format != backupFormat {
		result.Errors = append(result.Errors, "Backup-Format wird nicht unterstützt.")
	}
	if doc.Version < 1 || doc.Version > backupVersion {
		result.Errors = append(result.Errors, "Backup-Version wird nicht unterstützt.")
	}
	if err := ValidateBackupFiles(doc.Files); err != nil {
		result.Errors = append(result.Errors, "Backup-Dateien sind unvollständig oder beschädigt.")
	}

	requested := []string{}
	for _, table := range backupTableOrder {
		if _, present := doc.Tables[table]; present {
			requested = append(requested, table)
		}
	}
	schemas, err := s.store.Schema(ctx, requested)
	if err != nil {
		return nil, err
	}

	if len(doc.Tables["layouts"]) > 0 {
		result.Warnings = append(result.Warnings, "Anlagenplanung ist entfernt. Planungsdaten bleiben archiviert; Anlagenreservierungen werden freigegeben und Anlageneinbauten ins Ursprungslager zurückgeführt. / Layout planning is retired; legacy data is archived, reservations released and installations returned to their source storage.")
	}
	knownTables := map[string]struct{}{}
	for _, table := range backupTableOrder {
		knownTables[table] = struct{}{}
		rows, exists := doc.Tables[table]
		item := BackupValidationTable{Name: table, Rows: len(rows), Missing: !exists}
		if !exists {
			if backupTableOptional(doc.Version, table) {
				result.Warnings = append(result.Warnings, fmt.Sprintf("Optionale Tabelle %s fehlt im Backup und wird leer wiederhergestellt.", table))
			} else {
				result.Errors = append(result.Errors, fmt.Sprintf("Tabelle %s fehlt im Backup.", table))
			}
			result.Tables = append(result.Tables, item)
			continue
		}
		result.TableCount++
		result.RowCount += len(rows)

		columns := schemas[table]

		unknownColumns := map[string]struct{}{}
		for _, row := range rows {
			for column := range row {
				if !columns[column] {
					unknownColumns[column] = struct{}{}
				}
			}
		}
		item.UnknownColumns = SortedKeys(unknownColumns)
		warnings, problems := s.rules.Table(doc, table, rows, item.UnknownColumns)
		result.Warnings = append(result.Warnings, warnings...)
		result.Errors = append(result.Errors, problems...)

		result.Tables = append(result.Tables, item)
	}
	for table := range doc.Tables {
		if _, ok := knownTables[table]; !ok {
			if _, operational := backupOperationalTableExclusions[table]; operational {
				result.Warnings = append(result.Warnings,
					fmt.Sprintf("Operative Tabelle %s wird beim Restore ignoriert.", table))
				continue
			}
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("Unbekannte Tabelle %s wird beim Restore ignoriert.", table))
		}
	}
	result.Errors = append(result.Errors, s.rules.Document(doc)...)

	return FinishBackupValidation(result), nil
}
