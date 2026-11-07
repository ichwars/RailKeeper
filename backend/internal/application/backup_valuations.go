package application

import (
	"encoding/json"

	"railkeeper/backend/internal/valuation"
)

func validateBackupValuations(doc *BackupDocument) []string {
	const invalid = "Bewertungsverlauf ungültig. / Invalid valuation history."
	rows := doc.Tables["vehicle_valuations"]
	if len(rows) == 0 {
		return nil
	}
	if doc.Version < 28 {
		return []string{invalid}
	}
	vehicles := map[string]bool{}
	for _, row := range doc.Tables["vehicles"] {
		if id, ok := backupNonEmptyString(row["id"]); ok {
			vehicles[id] = true
		}
	}
	records := map[string]valuation.Record{}
	successors := map[string]string{}
	for _, row := range rows {
		r, ok := valuationBackupRecord(row)
		if !ok || valuation.ValidateRecord(r) != nil || !vehicles[r.VehicleID] || records[r.ID].ID != "" {
			return []string{invalid}
		}
		if r.ReplacesID != "" {
			if successors[r.ReplacesID] != "" {
				return []string{invalid}
			}
			successors[r.ReplacesID] = r.ID
		}
		records[r.ID] = r
	}
	for _, r := range records {
		if r.ReplacesID != "" && records[r.ReplacesID].VehicleID != r.VehicleID {
			return []string{invalid}
		}
	}
	// Iterative colouring keeps long correction chains linear and stack bounded.
	visited := map[string]uint8{}
	for id := range records {
		var chain []string
		for next := id; next != "" && visited[next] != 2; next = records[next].ReplacesID {
			if visited[next] == 1 {
				return []string{invalid}
			}
			visited[next] = 1
			chain = append(chain, next)
		}
		for _, next := range chain {
			visited[next] = 2
		}
	}
	return nil
}

func valuationBackupRecord(row map[string]any) (valuation.Record, bool) {
	var r valuation.Record
	fields := map[string]*string{"id": &r.ID, "vehicle_id": &r.VehicleID, "amount": &r.Amount,
		"currency": &r.Currency, "valued_on": &r.ValuedOn, "source": &r.Source, "notes": &r.Notes, "recorded_at": &r.RecordedAt}
	for key, target := range fields {
		text, ok := row[key].(string)
		if !ok {
			return valuation.Record{}, false
		}
		*target = text
	}
	predecessor, exists := row["replaces_id"]
	if !exists {
		return valuation.Record{}, false
	}
	if predecessor != nil {
		text, ok := predecessor.(string)
		if !ok || text == "" {
			return valuation.Record{}, false
		}
		r.ReplacesID = text
	}
	switch flag := row["withdrawn"].(type) {
	case int64:
		if flag != 0 && flag != 1 {
			return valuation.Record{}, false
		}
		r.Withdrawn = flag == 1
	case int:
		if flag != 0 && flag != 1 {
			return valuation.Record{}, false
		}
		r.Withdrawn = flag == 1
	case float64:
		if flag != 0 && flag != 1 {
			return valuation.Record{}, false
		}
		r.Withdrawn = flag == 1
	case json.Number:
		if flag != "0" && flag != "1" {
			return valuation.Record{}, false
		}
		r.Withdrawn = flag == "1"
	case bool:
		r.Withdrawn = flag
	default:
		return valuation.Record{}, false
	}
	return r, true
}
