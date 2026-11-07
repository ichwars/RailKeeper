package transfer

import "errors"

var (
	ErrDataTransferUploadTooLarge = errors.New("data transfer upload exceeds the size limit")
	ErrDataTransferUploadPath     = errors.New("data transfer upload path is not confined")
)

var (
	ErrDataTransferValidation = errors.New("data transfer validation failed")
	ErrDataTransferNotFound   = errors.New("data transfer profile not found")
)

var (
	ErrDataTransferArtifactPath      = errors.New("data transfer artifact path is not confined")
	ErrDataTransferForbidden         = errors.New("data transfer area is forbidden")
	ErrDataTransferArtifactDeleted   = errors.New("data transfer artifact was deleted")
	ErrDataTransferConflict          = errors.New("data transfer state conflict")
	ErrDataTransferOpenUnavailable   = errors.New("opening the data transfer export folder is unavailable")
	ErrDataTransferExportUnavailable = errors.New("data transfer export repository is unavailable")
)
