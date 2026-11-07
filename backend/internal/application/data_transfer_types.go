package application

import "railkeeper/backend/internal/transfer"

type TransferDirection = transfer.TransferDirection
type TransferFormat = transfer.TransferFormat
type TransferArea = transfer.TransferArea
type TransferJobState = transfer.TransferJobState
type TransferIssueSeverity = transfer.TransferIssueSeverity

const (
	TransferImport                   = transfer.TransferImport
	TransferExport                   = transfer.TransferExport
	TransferCSV                      = transfer.TransferCSV
	TransferJSON                     = transfer.TransferJSON
	TransferVehicles                 = transfer.TransferVehicles
	TransferAccessories              = transfer.TransferAccessories
	TransferExhibitionLists          = transfer.TransferExhibitionLists
	TransferJobDraft                 = transfer.TransferJobDraft
	TransferJobReading               = transfer.TransferJobReading
	TransferJobReviewRequired        = transfer.TransferJobReviewRequired
	TransferJobReady                 = transfer.TransferJobReady
	TransferJobRunning               = transfer.TransferJobRunning
	TransferJobCompleted             = transfer.TransferJobCompleted
	TransferJobCompletedWithWarnings = transfer.TransferJobCompletedWithWarnings
	TransferJobFailed                = transfer.TransferJobFailed
	TransferJobCancelled             = transfer.TransferJobCancelled
	TransferIssueWarning             = transfer.TransferIssueWarning
	TransferIssueError               = transfer.TransferIssueError
)

type DataTransferProfile = transfer.DataTransferProfile
type DataTransferJob = transfer.DataTransferJob
type DataTransferJobFilter = transfer.DataTransferJobFilter
type DataTransferIssue = transfer.DataTransferIssue
type DataTransferArtifact = transfer.DataTransferArtifact
type DataTransferImportMutation = transfer.DataTransferImportMutation
type DataTransferProfileOptionsMutation = transfer.DataTransferProfileOptionsMutation
type DataTransferImportPolicy = transfer.DataTransferImportPolicy
type DataTransferRepository = transfer.DataTransferRepository
