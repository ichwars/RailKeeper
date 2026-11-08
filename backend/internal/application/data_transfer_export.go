package application

import (
	"railkeeper/backend/internal/transfer"
	"railkeeper/backend/internal/transfer/workflow"
)

var (
	ErrDataTransferArtifactPath      = transfer.ErrDataTransferArtifactPath
	ErrDataTransferForbidden         = transfer.ErrDataTransferForbidden
	ErrDataTransferArtifactDeleted   = transfer.ErrDataTransferArtifactDeleted
	ErrDataTransferConflict          = transfer.ErrDataTransferConflict
	ErrDataTransferOpenUnavailable   = transfer.ErrDataTransferOpenUnavailable
	ErrDataTransferExportUnavailable = transfer.ErrDataTransferExportUnavailable
)

type DataTransferExportResult = workflow.DataTransferExportResult
type DataTransferFolderOpener = workflow.DataTransferFolderOpener
type DataTransferServiceOption = workflow.Option

func WithDataTransferFolderOpener(available bool, opener DataTransferFolderOpener) DataTransferServiceOption {
	return workflow.WithDataTransferFolderOpener(available, opener)
}
