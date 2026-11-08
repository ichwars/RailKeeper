package vehicle

import (
	"errors"

	"railkeeper/backend/internal/domain"
)

var (
	ErrVehicleValidation              = domain.ErrVehicleValidation
	ErrVehicleOperationalValidation   = errors.New("vehicle operational validation failed")
	ErrVehicleNotFound                = domain.ErrVehicleNotFound
	ErrVehicleExternalMappingConflict = errors.New("vehicle external mapping conflict")
	ErrVehicleSetValidation           = errors.New("vehicle set validation failed")
	ErrVehicleSetNotFound             = errors.New("vehicle set not found")
	ErrVehicleSetImageValidation      = errors.New("vehicle set image validation failed")
	ErrVehicleSetImageNotFound        = errors.New("vehicle set image not found")
	ErrVehicleImageInUse              = errors.New("vehicle image in use")
)
var (
	ErrVehicleCVFileConflict     = errors.New("vehicle CV file or its source references changed")
	ErrVehicleConflict           = errors.New("vehicle changed since the draft was opened")
	ErrVehicleImageConflict      = errors.New("vehicle images changed since the draft was opened")
	ErrVehicleAttachmentConflict = errors.New("vehicle attachment changed since the draft was opened")
)
var (
	ErrVehicleFunctionValidation = errors.New("invalid vehicle function changes")
	ErrVehicleFunctionConflict   = errors.New("vehicle function changed concurrently")
)
