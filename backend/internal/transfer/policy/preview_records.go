package policy

import (
	"crypto/sha256"

	"encoding/hex"
	"encoding/json"

	"fmt"

	"strings"
)

func appendRequiredTransferIssues(
	issues []DataTransferIssue,
	jobID string,
	area TransferArea,
	recordKey string,
	rowNumber *int,
	inventoryNumber string,
	manufacturer string,
	name string,
) []DataTransferIssue {
	if strings.TrimSpace(inventoryNumber) == "" {
		issues = append(issues, NewTransferIssue(jobID, area, recordKey, rowNumber, "inventoryNumber",
			TransferIssueError, "missing_inventory_number", "Inventory number is required.", "skip"))
	}
	if strings.TrimSpace(manufacturer) == "" {
		issues = append(issues, NewTransferIssue(jobID, area, recordKey, rowNumber, "manufacturer",
			TransferIssueError, "missing_manufacturer", "Manufacturer is required.", "skip"))
	}
	if strings.TrimSpace(name) == "" {
		issues = append(issues, NewTransferIssue(jobID, area, recordKey, rowNumber, "name",
			TransferIssueError, "missing_name", "Name is required.", "skip"))
	}
	return issues
}

func NewTransferIssue(
	jobID string,
	area TransferArea,
	recordKey string,
	rowNumber *int,
	field string,
	severity TransferIssueSeverity,
	code string,
	message string,
	proposedResolution string,
) DataTransferIssue {
	row := 0
	if rowNumber != nil {
		row = *rowNumber
	}
	identity := fmt.Sprintf("%s\x00%s\x00%s\x00%d\x00%s\x00%s", jobID, area, recordKey, row, field, code)
	digest := sha256.Sum256([]byte(identity))
	return DataTransferIssue{
		ID: hex.EncodeToString(digest[:]), JobID: jobID, Area: area, RecordKey: recordKey,
		RowNumber: rowNumber, Field: field,
		Severity: severity, Code: code, Message: message, ProposedResolution: proposedResolution,
	}
}

func newDataTransferPreviewRecord(
	area TransferArea,
	recordKey string,
	rowNumber *int,
	data any,
) DataTransferPreviewRecord {
	payload, _ := json.Marshal(data)
	return DataTransferPreviewRecord{
		Area: area, RecordKey: recordKey, RowNumber: rowNumber, Classification: "ready",
		ProposedAction: "create", Data: payload,
	}
}

func FinalizeDataTransferRecord(record *DataTransferPreviewRecord, issues []DataTransferIssue) {
	record.Classification = "ready"
	for _, issue := range issues {
		if issue.Severity == TransferIssueError {
			record.Classification = "error"
			return
		}
		if issue.Severity == TransferIssueWarning {
			record.Classification = "warning"
		}
	}
}

func CountDataTransferPreviewRecords(records []DataTransferPreviewRecord) (ready, warnings, failures int) {
	for _, record := range records {
		switch record.Classification {
		case "ready":
			ready++
		case "warning":
			warnings++
		case "error":
			failures++
		}
	}
	return ready, warnings, failures
}

func transferIdentity(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func transferArticleIdentity(manufacturer, articleNumber string) string {
	manufacturer = transferIdentity(manufacturer)
	articleNumber = transferIdentity(articleNumber)
	if manufacturer == "" || articleNumber == "" {
		return ""
	}
	return manufacturer + "\x00" + articleNumber
}

func transferExhibitionIdentity(designation, date string) string {
	return transferIdentity(designation) + "\x00" + strings.TrimSpace(date)
}

func transferRowNumber(rows []int, index int) *int {
	if index < 0 || index >= len(rows) {
		value := index + 1
		return &value
	}
	value := rows[index]
	return &value
}

func transferExhibitionEntryIssueField(index int) string {
	return fmt.Sprintf("entries[%d].vehicleReference", index)
}

func transferRowValue(rowNumber *int, fallback int) int {
	if rowNumber == nil {
		return fallback
	}
	return *rowNumber
}
