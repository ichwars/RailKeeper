package policy

import (
	"strings"

	"railkeeper/backend/internal/domain"
	"railkeeper/backend/internal/transfer"
)

func ValidTransferApplyResolution(code, resolution string) bool {
	allowed := map[string]map[string]bool{
		"missing_inventory_number":               {"skip": true},
		"missing_manufacturer":                   {"skip": true},
		"missing_name":                           {"skip": true},
		"missing_gauge":                          {"skip": true},
		"missing_category":                       {"skip": true},
		"missing_gattung":                        {"skip": true},
		"invalid_vehicle":                        {"skip": true},
		"invalid_accessory":                      {"skip": true},
		"duplicate_inventory_number":             {"replace": true, "copy": true, "skip": true},
		"matching_manufacturer_article_number":   {"use_existing": true, "create": true, "skip": true},
		"duplicate_exhibition_list":              {"replace": true, "merge": true, "copy": true, "skip": true},
		"locked_exhibition_list":                 {"copy": true, "skip": true},
		"exhibition_vehicle_reference":           {"link": true, "skip": true},
		"missing_vehicle_reference":              {"skip": true},
		"duplicate_input_inventory_number":       {"skip": true},
		"duplicate_vehicle_set_inventory_number": {"replace": true, "copy": true, "skip": true},
		"vehicle_set_member_external_conflict":   {"copy": true, "skip": true},
	}
	return allowed[code][resolution]
}

func TransferRecordIssues(
	issues []transfer.DataTransferIssue,
	record transfer.DataTransferPreviewRecord,
) []transfer.DataTransferIssue {
	matched := []transfer.DataTransferIssue{}
	for _, issue := range issues {
		if issue.Area == record.Area && issue.RecordKey == record.RecordKey &&
			transferRowNumbersEqual(issue.RowNumber, record.RowNumber) {
			matched = append(matched, issue)
		}
	}
	return matched
}

func transferRowNumbersEqual(left, right *int) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func TransferRecordAction(
	record transfer.DataTransferPreviewRecord,
	issues []transfer.DataTransferIssue,
) (string, bool) {
	action := record.ProposedAction
	for _, issue := range issues {
		switch issue.Code {
		case "missing_inventory_number", "missing_manufacturer", "missing_name", "missing_gauge",
			"missing_category", "missing_gattung", "invalid_vehicle", "invalid_accessory",
			"duplicate_input_inventory_number":
			return action, true
		case "duplicate_inventory_number", "matching_manufacturer_article_number", "duplicate_exhibition_list",
			"locked_exhibition_list":
			if issue.SelectedResolution == "skip" {
				return action, true
			}
			action = issue.SelectedResolution
		}
	}
	return action, false
}

func NormalizeTransferExhibitionList(list *transfer.TransferExhibitionList) {
	if strings.TrimSpace(list.EndDate) == "" {
		list.EndDate = list.Date
	}
	if list.Status == "" {
		list.Status = domain.ExhibitionStatusOpen
		if list.Locked {
			list.Status = domain.ExhibitionStatusLocked
		}
	}
	list.Locked = list.Status == domain.ExhibitionStatusLocked ||
		list.Status == domain.ExhibitionStatusRunning
}

func NormalizeTransferExhibitionEntry(entry *transfer.TransferExhibitionEntry) {
	if strings.TrimSpace(entry.InterfaceName) == "" {
		entry.InterfaceName = entry.Adapter
	}
	if strings.TrimSpace(entry.Availability) == "" {
		entry.Availability = "available"
	}
}

func NonNilStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}
