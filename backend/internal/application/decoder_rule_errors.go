package application

import (
	"errors"

	"railkeeper/backend/internal/decoder"
)

func cvRuleError(err error) error {
	if errors.Is(err, decoder.ErrCVValidation) {
		return ErrVehicleValidation
	}
	return err
}
