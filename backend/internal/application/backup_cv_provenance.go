package application

import (
	"fmt"
	"strconv"
)

func validateBackupCVProvenance(rows []map[string]any) []string {
	for i, row := range rows {
		if value := row["source_line"]; value != nil {
			line, err := strconv.ParseInt(fmt.Sprint(value), 10, 64)
			if err != nil || line < 1 || line > 10000000 {
				return []string{fmt.Sprintf("CV-Zeile %d: ungültige Originalzeile.", i+1)}
			}
		}
		if value, exists := row["source_exported_at"]; exists {
			date, ok := value.(string)
			if !ok || !validCVExportTimestamp(date) {
				return []string{fmt.Sprintf("CV-Zeile %d: ungültiges Exportdatum.", i+1)}
			}
		}
	}
	return nil
}
