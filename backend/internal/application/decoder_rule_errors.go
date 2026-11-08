package application

import (
	"errors"

	"railkeeper/backend/internal/decoder"
	"railkeeper/backend/internal/decoder/workflow"
)

func cvRuleError(err error) error {
	if errors.Is(err, decoder.ErrCVValidation) {
		return ErrVehicleValidation
	}
	if errors.Is(err, workflow.ErrCVNotFound) {
		return ErrVehicleNotFound
	}
	if errors.Is(err, workflow.ErrCVFileConflict) {
		return ErrVehicleCVFileConflict
	}
	if errors.Is(err, workflow.ErrStoredCVMappingConflict) {
		return ErrDigitalCenterConflictUnresolved
	}
	return err
}
