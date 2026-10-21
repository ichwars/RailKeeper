package application

import (
	"encoding/json"
	"fmt"
)

func legacyFunctionRowsForRestore(rows []map[string]any) []map[string]any {
	out := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		copyRow := make(map[string]any, len(row)+2)
		for key, value := range row {
			copyRow[key] = value
		}
		copyRow["legacy_direction_dependent"] = row["direction_dependent"]
		copyRow["direction_dependent"], copyRow["inverted"] = nil, nil
		out = append(out, copyRow)
	}
	return out
}

func validateBackupFunctionFlags(version int, rows []map[string]any) []string {
	var problems []string
	for i, row := range rows {
		columns := []string{"direction_dependent"}
		if version >= 24 {
			columns = []string{"direction_dependent", "inverted", "legacy_direction_dependent"}
		}
		for _, column := range columns {
			value, exists := row[column]
			if !exists && version >= 24 && column != "legacy_direction_dependent" {
				problems = append(problems, fmt.Sprintf("Funktionsangabe %s fehlt in Zeile %d.", column, i+1))
				continue
			}
			if !validBackupFunctionFlag(value) {
				problems = append(problems, fmt.Sprintf("Ungültige Funktionsangabe %s in Zeile %d.", column, i+1))
			}
		}
	}
	return problems
}

func validBackupFunctionFlag(value any) bool {
	switch v := value.(type) {
	case nil, bool:
		return true
	case int:
		return v == 0 || v == 1
	case int64:
		return v == 0 || v == 1
	case float64:
		return v == 0 || v == 1
	case json.Number:
		return v == "0" || v == "1"
	default:
		return false
	}
}
