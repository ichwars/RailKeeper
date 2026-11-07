package transfer

import "encoding/json"

type DataTransferPreviewRecord struct {
	Area              TransferArea    `json:"area"`
	RecordKey         string          `json:"recordKey"`
	RowNumber         *int            `json:"rowNumber,omitempty"`
	Classification    string          `json:"classification"`
	ProposedAction    string          `json:"proposedAction"`
	TargetID          string          `json:"targetId,omitempty"`
	TargetUpdatedAt   string          `json:"targetUpdatedAt,omitempty"`
	TargetFingerprint string          `json:"targetFingerprint,omitempty"`
	Data              json.RawMessage `json:"data"`
}

type DataTransferVehicleSetPreview struct {
	RecordKey         string                         `json:"recordKey"`
	Classification    string                         `json:"classification"`
	ProposedAction    string                         `json:"proposedAction"`
	TargetID          string                         `json:"targetId,omitempty"`
	TargetUpdatedAt   string                         `json:"targetUpdatedAt,omitempty"`
	TargetFingerprint string                         `json:"targetFingerprint,omitempty"`
	MemberRecordKeys  []string                       `json:"memberRecordKeys"`
	RowNumbers        []int                          `json:"rowNumbers,omitempty"`
	Diagnostics       []TransferVehicleSetDiagnostic `json:"diagnostics,omitempty"`
	Data              TransferVehicleSet             `json:"data"`
}
