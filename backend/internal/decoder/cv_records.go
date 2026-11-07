package decoder

import (
	"railkeeper/backend/internal/revision"
)

func CVRecordValue(item CVRecord) VehicleCVValue {
	return VehicleCVValue{ID: item.ID, VehicleID: item.VehicleID, CVNumber: item.CVNumber,
		IndexCV31: item.IndexCV31, IndexCV32: item.IndexCV32, Value: item.Value, Description: item.Description,
		Category: item.Category, Protocol: item.Protocol, DecoderProfile: item.DecoderProfile,
		SourceFileID: item.SourceFileID, SourceLine: item.SourceLine, SourceExportedAt: item.SourceExportedAt,
		CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt, Revision: CVRecordRevision(item)}
}

func CVRecordInput(vehicleID string, input VehicleCVValueInput) CVRecord {
	return CVRecord{VehicleID: vehicleID, CVNumber: input.CVNumber, IndexCV31: input.IndexCV31,
		IndexCV32: input.IndexCV32, Value: input.Value, Description: input.Description, Category: input.Category,
		Protocol: input.Protocol, DecoderProfile: input.DecoderProfile, SourceFileID: input.SourceFileID,
		SourceLine: input.SourceLine, SourceExportedAt: input.SourceExportedAt}
}

func ValidCVRevision(value string) bool { return revision.Valid(value) }

func CVRecordRevision(item CVRecord) string { return revision.Hash(item) }
