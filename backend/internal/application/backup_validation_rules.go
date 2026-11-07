package application

import (
	"fmt"

	"railkeeper/backend/internal/backup"
)

type legacyBackupValidationRules struct{}

func BackupValidationRules() backup.Rules { return legacyBackupValidationRules{} }
func (legacyBackupValidationRules) Table(doc *BackupDocument, table string, rows []map[string]any, unknown []string) ([]string, []string) {
	result := &BackupValidationResult{Warnings: []string{}, Errors: []string{}}
	item := BackupValidationTable{UnknownColumns: unknown}
	result.Errors = append(result.Errors, validateBackupMediaPaths(table, rows)...)
	if table == "vehicle_functions" {
		result.Errors = append(result.Errors, validateBackupFunctionFlags(doc.Version, rows)...)
	}
	if table == "vehicle_cv_values" {
		result.Errors = append(result.Errors, validateBackupCVProvenance(rows)...)
	}
	if len(item.UnknownColumns) > 0 {
		if isVehicleCatalogBackupTable(table) || isDecoderConfigurationBackupTable(table) || table == "vehicle_valuations" {
			result.Errors = append(result.Errors, fmt.Sprintf("Fachtabelle %s enthält unbekannte Spalten. / Unknown model table columns.", table))
		} else {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Tabelle %s enthält unbekannte Spalten, die beim Restore ignoriert werden.", table))
		}
	}
	if table == "accessory_product_attributes" {
		result.Errors = append(result.Errors,
			validateBackupAccessoryProductAttributes(
				rows,
				doc.Tables["accessory_products"],
				doc.Tables["master_data_entries"],
				doc.Version,
			)...)
	}
	if table == "master_data_entries" && doc.Version >= 3 {
		result.Errors = append(result.Errors, validateBackupProtectedArticleTypes(rows)...)
	}
	if table == "master_data_entries" {
		result.Errors = append(result.Errors, validateBackupDecoderLibrary(rows)...)
	}
	return result.Warnings, result.Errors
}
func (legacyBackupValidationRules) Document(doc *BackupDocument) []string {
	result := &BackupValidationResult{Errors: []string{}}
	if doc.Profiles != nil {
		result.Errors = append(result.Errors, validateBackupDataTransferProfiles(doc.Profiles)...)
	}
	result.Errors = append(result.Errors, validateBackupVehicleCatalog(doc)...)
	result.Errors = append(result.Errors, validateBackupDecoderConfiguration(doc)...)
	result.Errors = append(result.Errors, validateBackupValuations(doc)...)

	return result.Errors
}
