package policy

import (
	"encoding/json"

	"fmt"
)

type DataTransferPreview struct {
	Job            DataTransferJob                 `json:"job"`
	Records        []DataTransferPreviewRecord     `json:"records"`
	Issues         []DataTransferIssue             `json:"issues"`
	TotalRecords   int                             `json:"totalRecords"`
	ReadyRecords   int                             `json:"readyRecords"`
	WarningRecords int                             `json:"warningRecords"`
	ErrorRecords   int                             `json:"errorRecords"`
	CSVMapping     []DataTransferCSVColumnMapping  `json:"csvMapping,omitempty"`
	VehicleFields  []VehicleTransferField          `json:"vehicleFields,omitempty"`
	VehicleSets    []DataTransferVehicleSetPreview `json:"vehicleSets,omitempty"`
}

func TransferRecordHasIssue(
	issues []DataTransferIssue,
	record DataTransferPreviewRecord,
	code string,
) bool {
	for _, issue := range issues {
		if issue.Area == record.Area && issue.RecordKey == record.RecordKey && issue.Code == code &&
			transferPreviewRowsEqual(issue.RowNumber, record.RowNumber) {
			return true
		}
	}
	return false
}

func TransferIssuesForPreviewRecord(
	issues []DataTransferIssue,
	record DataTransferPreviewRecord,
) []DataTransferIssue {
	matched := make([]DataTransferIssue, 0)
	for _, issue := range issues {
		if issue.Area == record.Area && issue.RecordKey == record.RecordKey &&
			transferPreviewRowsEqual(issue.RowNumber, record.RowNumber) {
			matched = append(matched, issue)
		}
	}
	return matched
}

func transferPreviewRowsEqual(left, right *int) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func DataTransferPreviewMap(
	sourceSHA256 string,
	records []DataTransferPreviewRecord,
	vehicleSets []DataTransferVehicleSetPreview,
	csvMapping []DataTransferCSVColumnMapping,
	vehicleFields []VehicleTransferField,
) (map[string]any, error) {
	payload, err := json.Marshal(struct {
		SourceSHA256       string                          `json:"sourceSha256"`
		FingerprintVersion int                             `json:"fingerprintVersion"`
		Records            []DataTransferPreviewRecord     `json:"records"`
		VehicleSets        []DataTransferVehicleSetPreview `json:"vehicleSets,omitempty"`
		CSVMapping         []DataTransferCSVColumnMapping  `json:"csvMapping,omitempty"`
		VehicleFields      []VehicleTransferField          `json:"vehicleFields,omitempty"`
	}{SourceSHA256: sourceSHA256, FingerprintVersion: DataTransferTargetFingerprintVersion,
		Records: records, VehicleSets: vehicleSets, CSVMapping: csvMapping, VehicleFields: vehicleFields})
	if err != nil {
		return nil, fmt.Errorf("encode transfer preview: %w", err)
	}
	preview := map[string]any{}
	if err := json.Unmarshal(payload, &preview); err != nil {
		return nil, fmt.Errorf("persist transfer preview: %w", err)
	}
	return preview, nil
}

func NewDataTransferPreview(
	job DataTransferJob,
	records []DataTransferPreviewRecord,
	issues []DataTransferIssue,
	vehicleSets []DataTransferVehicleSetPreview,
	csvMapping []DataTransferCSVColumnMapping,
	vehicleFields []VehicleTransferField,
) DataTransferPreview {
	return DataTransferPreview{
		Job: job, Records: records, Issues: issues, TotalRecords: job.TotalRecords,
		ReadyRecords: job.ReadyRecords, WarningRecords: job.WarningRecords, ErrorRecords: job.ErrorRecords,
		CSVMapping: csvMapping, VehicleFields: vehicleFields, VehicleSets: vehicleSets,
	}
}

func ValidDataTransferResolution(issue DataTransferIssue, resolution string) bool {
	return resolution != "" && ValidTransferApplyResolution(issue.Code, resolution)
}
