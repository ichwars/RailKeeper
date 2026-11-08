package application

import (
	"railkeeper/backend/internal/transfer"
	transferpolicy "railkeeper/backend/internal/transfer/policy"
)

const DataTransferMaxUploadBytes int64 = 25 * 1024 * 1024

const DataTransferTargetFingerprintVersion = transfer.DataTransferTargetFingerprintVersion

var (
	ErrDataTransferUploadTooLarge = transfer.ErrDataTransferUploadTooLarge
	ErrDataTransferUploadPath     = transfer.ErrDataTransferUploadPath
)

type DataTransferPreview = transferpolicy.DataTransferPreview

type DataTransferPreviewRecord = transfer.DataTransferPreviewRecord

func DataTransferTargetFingerprint(value any) string {
	return transfer.DataTransferTargetFingerprint(value)
}

func DataTransferTargetFingerprintForVersion(value any, version int) string {
	return transfer.DataTransferTargetFingerprintForVersion(value, version)
}
