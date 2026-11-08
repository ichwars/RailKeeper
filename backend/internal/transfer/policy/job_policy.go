package policy

import (
	"fmt"
	"slices"
)

func ValidateDataTransferJobFilter(filter DataTransferJobFilter) error {
	if filter.Limit < 0 || filter.Limit > 200 {
		return ErrDataTransferValidation
	}
	if filter.Direction != "" && !ValidTransferDirection(filter.Direction) {
		return ErrDataTransferValidation
	}
	for _, state := range filter.States {
		if !validDataTransferJobState(state) {
			return ErrDataTransferValidation
		}
	}
	return nil
}

func validDataTransferJobState(state TransferJobState) bool {
	return slices.Contains([]TransferJobState{
		TransferJobDraft, TransferJobReading, TransferJobReviewRequired, TransferJobReady, TransferJobRunning,
		TransferJobCompleted, TransferJobCompletedWithWarnings, TransferJobFailed, TransferJobCancelled,
	}, state)
}

func DataTransferJobOpen(state TransferJobState) bool {
	return state == TransferJobDraft || state == TransferJobReading || state == TransferJobReviewRequired ||
		state == TransferJobReady || state == TransferJobRunning
}

func DataTransferJobRetryable(state TransferJobState) bool {
	return state == TransferJobCompleted || state == TransferJobCompletedWithWarnings ||
		state == TransferJobFailed || state == TransferJobCancelled
}

func ValidateTransferSelection(format TransferFormat, areas []TransferArea) error {
	allowed := map[TransferArea]bool{
		TransferVehicles: true, TransferAccessories: true, TransferExhibitionLists: true,
	}
	if len(areas) == 0 {
		return fmt.Errorf("%w: select at least one area", ErrDataTransferValidation)
	}
	seen := map[TransferArea]bool{}
	for _, area := range areas {
		if !allowed[area] || seen[area] {
			return fmt.Errorf("%w: unsupported or repeated area %q", ErrDataTransferValidation, area)
		}
		seen[area] = true
	}
	if format == TransferCSV && len(areas) != 1 {
		return fmt.Errorf("%w: csv requires exactly one area", ErrDataTransferValidation)
	}
	if format == TransferCSV && areas[0] == TransferExhibitionLists {
		return fmt.Errorf("%w: exhibition lists require railkeeper-json", ErrDataTransferValidation)
	}
	return nil
}

func ValidTransferDirection(direction TransferDirection) bool {
	return direction == TransferImport || direction == TransferExport
}

func ValidTransferFormat(format TransferFormat) bool {
	return format == TransferCSV || format == TransferJSON
}
