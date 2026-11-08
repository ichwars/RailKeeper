package backup

import "railkeeper/backend/internal/inventory/masterdata"

func rowForCompatibilityRestore(table string, row map[string]any) map[string]any {
	if table != "master_data_entries" {
		return row
	}
	restoreRow := make(map[string]any, len(row)+1)
	for column, value := range row {
		restoreRow[column] = value
	}
	restoreRow["origin"] = string(masterdata.MasterDataOriginCustom)
	return restoreRow
}

func rowsForCompatibilityRestore(doc *BackupDocument, table string) []map[string]any {
	if table == "vehicle_functions" && doc.Version < 24 {
		return legacyFunctionRowsForRestore(doc.Tables[table])
	}

	rows := doc.Tables[table]
	legacyColumns := []string{}
	if doc.Version <= 10 {
		legacyColumn := map[string]string{
			"track_geometry_definitions": "minimum_radius_mm",
			"plan_track_objects":         "flex_path_json",
			"layouts":                    "minimum_flex_radius_mm",
		}[table]
		if legacyColumn != "" {
			legacyColumns = append(legacyColumns, legacyColumn)
		}
	}
	if doc.Version <= 11 && table == "plan_track_objects" {
		legacyColumns = append(legacyColumns, "transition_path_json")
	}
	if doc.Version <= 13 && table == "plan_track_objects" {
		legacyColumns = append(legacyColumns, "geometry_snapshot_json")
	}
	if len(legacyColumns) == 0 {
		return rows
	}
	normalized := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		copyRow := make(map[string]any, len(row)+len(legacyColumns))
		for column, value := range row {
			copyRow[column] = value
		}
		for _, legacyColumn := range legacyColumns {
			if _, exists := copyRow[legacyColumn]; !exists {
				copyRow[legacyColumn] = nil
			}
		}
		normalized = append(normalized, copyRow)
	}
	return normalized
}

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
