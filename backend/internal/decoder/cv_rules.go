package decoder

import (
	"fmt"
	"strings"
	"time"
)

func CleanCVValueInput(input VehicleCVValueInput) VehicleCVValueInput {
	input.Description = strings.TrimSpace(input.Description)
	input.Category = strings.TrimSpace(input.Category)
	input.Protocol = strings.TrimSpace(input.Protocol)
	input.DecoderProfile = strings.TrimSpace(input.DecoderProfile)
	input.SourceFileID = strings.TrimSpace(input.SourceFileID)
	input.SourceExportedAt = strings.TrimSpace(input.SourceExportedAt)
	return input
}

func ValidCVValueInput(input VehicleCVValueInput) bool {
	return ValidCVNumber(input.CVNumber) && ValidCVIndex(input) &&
		(input.SourceLine == nil || (*input.SourceLine >= 1 && *input.SourceLine <= 10000000)) &&
		ValidCVExportTimestamp(input.SourceExportedAt) &&
		ValidCVValue(input.Value) &&
		len(input.Description) <= 1000 &&
		len(input.Category) <= 80 &&
		len(input.Protocol) <= 80 &&
		len(input.DecoderProfile) <= 160 &&
		len(input.SourceFileID) <= 80
}

// LokProgrammer exports wall-clock time without a timezone. Do not invent an offset.
func ValidCVExportTimestamp(value string) bool {
	if value == "" {
		return true
	}
	parsed, err := time.Parse("2006-01-02T15:04:05", value)
	return err == nil && parsed.Year() >= 1 && parsed.Format("2006-01-02T15:04:05") == value
}

func CleanCVFileInput(input VehicleCVFileInput) VehicleCVFileInput {
	input.FileName = strings.TrimSpace(input.FileName)
	input.OriginalName = strings.TrimSpace(input.OriginalName)
	input.Description = strings.TrimSpace(input.Description)
	input.DecoderProfile = strings.TrimSpace(input.DecoderProfile)
	input.MimeType = strings.TrimSpace(input.MimeType)
	input.StoragePath = strings.TrimSpace(input.StoragePath)
	input.BlobID = strings.TrimSpace(input.BlobID)
	return input
}

func ValidCVIndex(input VehicleCVValueInput) bool {
	if input.IndexCV31 == nil && input.IndexCV32 == nil {
		return true
	}
	return input.IndexCV31 != nil && input.IndexCV32 != nil &&
		ValidCVValue(*input.IndexCV31) && ValidCVValue(*input.IndexCV32) && input.CVNumber >= 257 && input.CVNumber <= 512
}

func CVIndexKey(value *int) int {
	if value == nil {
		return -1
	}
	return *value
}

// Scoped inputs are resolved exactly inside the transaction before checking mixed scoped/legacy targets.
func CVRequestIdentity(value VehicleCVValueInput) string {
	if value.WorkspaceID != "" {
		return fmt.Sprintf("workspace:%q:%d:%d:%d", value.WorkspaceID, value.CVNumber, CVIndexKey(value.IndexCV31), CVIndexKey(value.IndexCV32))
	}
	return "profile:" + CVIdentity(value)
}

func CVIdentity(value VehicleCVValueInput) string {
	return fmt.Sprintf("%s:%d:%d:%d", value.DecoderProfile, value.CVNumber,
		CVIndexKey(value.IndexCV31), CVIndexKey(value.IndexCV32))
}
