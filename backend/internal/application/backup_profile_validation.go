package application

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	transferpolicy "railkeeper/backend/internal/transfer/policy"
)

func validateBackupDataTransferProfiles(rows []map[string]any) []string {
	allowedColumns := map[string]bool{
		"id": true, "name": true, "direction": true, "format": true, "areas_json": true,
		"options_json": true, "enabled": true, "created_by_user_id": true, "last_used_at": true,
		"created_at": true, "updated_at": true,
	}
	seen := map[string]bool{}
	errorsFound := []string{}
	for index, row := range rows {
		prefix := fmt.Sprintf("Transferprofil %d", index+1)
		for column := range row {
			if !allowedColumns[column] {
				errorsFound = append(errorsFound, prefix+" enthält die unbekannte Spalte "+column+".")
			}
		}
		id, idValid := backupNonEmptyString(row["id"])
		_, nameValid := backupNonEmptyString(row["name"])
		direction, directionValid := backupNonEmptyString(row["direction"])
		format, formatValid := backupNonEmptyString(row["format"])
		areasJSON, areasValid := backupNonEmptyString(row["areas_json"])
		optionsJSON, optionsValid := backupNonEmptyString(row["options_json"])
		createdAt, createdValid := backupNonEmptyString(row["created_at"])
		updatedAt, updatedValid := backupNonEmptyString(row["updated_at"])
		if !idValid || !nameValid || !directionValid || !formatValid || !areasValid || !optionsValid ||
			!createdValid || !updatedValid {
			errorsFound = append(errorsFound, prefix+" enthält unvollständige Pflichtfelder.")
			continue
		}
		if seen[id] {
			errorsFound = append(errorsFound, prefix+" verwendet eine doppelte ID.")
		}
		seen[id] = true
		if !transferpolicy.ValidTransferDirection(TransferDirection(direction)) || !transferpolicy.ValidTransferFormat(TransferFormat(format)) {
			errorsFound = append(errorsFound, prefix+" enthält Richtung oder Format außerhalb des Vertrags.")
		}
		areas := []TransferArea{}
		if err := json.Unmarshal([]byte(areasJSON), &areas); err != nil ||
			transferpolicy.ValidateTransferSelection(TransferFormat(format), areas) != nil {
			errorsFound = append(errorsFound, prefix+" enthält eine ungültige Bereichsauswahl.")
		}
		options := map[string]any{}
		if err := json.Unmarshal([]byte(optionsJSON), &options); err != nil || options == nil {
			errorsFound = append(errorsFound, prefix+" enthält ungültige Optionen.")
		}
		if _, valid := backupBooleanValue(row["enabled"]); !valid {
			errorsFound = append(errorsFound, prefix+" enthält einen ungültigen Aktivstatus.")
		}
		if _, err := time.Parse(time.RFC3339, createdAt); err != nil {
			errorsFound = append(errorsFound, prefix+" enthält einen ungültigen Erstellzeitpunkt.")
		}
		if _, err := time.Parse(time.RFC3339, updatedAt); err != nil {
			errorsFound = append(errorsFound, prefix+" enthält einen ungültigen Änderungszeitpunkt.")
		}
		for _, nullableColumn := range []string{"created_by_user_id", "last_used_at"} {
			value := row[nullableColumn]
			if value != nil {
				text, ok := value.(string)
				if !ok || strings.TrimSpace(text) == "" {
					errorsFound = append(errorsFound, prefix+" enthält einen ungültigen Wert für "+nullableColumn+".")
				} else if nullableColumn == "last_used_at" {
					if _, err := time.Parse(time.RFC3339, text); err != nil {
						errorsFound = append(errorsFound, prefix+" enthält einen ungültigen letzten Nutzungszeitpunkt.")
					}
				}
			}
		}
	}
	return errorsFound
}

func validateBackupProtectedArticleTypes(rows []map[string]any) []string {
	entries := make([]MasterDataEntry, 0, len(standardArticleTypeKeys))
	for _, row := range rows {
		typeName, valid := backupNonEmptyString(row["type"])
		if !valid || typeName != standardArticleType {
			continue
		}
		key, _ := backupNonEmptyString(row["key"])
		label, _ := backupNonEmptyString(row["label"])
		entries = append(entries, MasterDataEntry{Type: typeName, Key: key, Label: label})
	}
	if err := validateProtectedArticleTypes(entries, true); err != nil {
		return []string{"Tabelle master_data_entries enthält keine gültige vollständige Artikelarten-Konfiguration."}
	}
	return nil
}
