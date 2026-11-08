package application

import (
	transferpolicy "railkeeper/backend/internal/transfer/policy"
)

func classifyDataTransferImport(
	jobID string,
	incoming DataTransferSnapshot,
	current DataTransferSnapshot,
	rowNumbers map[TransferArea][]int,
) ([]DataTransferPreviewRecord, []DataTransferIssue) {
	return transferpolicy.ClassifyDataTransferImport(jobID, incoming, current, rowNumbers, transferpolicy.AggregateRules{Vehicle: ValidateTransferVehicle, Accessory: ValidateTransferAccessory})
}
