// Package policy owns transfer parsing, field mapping, conflict decisions and encoding.
// AggregateRules reuses the same aggregate validation as ordinary inventory writes.
package policy

import (
	"railkeeper/backend/internal/domain"
	"railkeeper/backend/internal/transfer"
)

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

const (
	DataTransferPackageFormat          = transfer.DataTransferPackageFormat
	DataTransferPackageLegacyVersion   = transfer.DataTransferPackageLegacyVersion
	DataTransferPackagePreviousVersion = transfer.DataTransferPackagePreviousVersion
	DataTransferPackageVersion         = transfer.DataTransferPackageVersion
)

type DataTransferPackage = transfer.DataTransferPackage

type DataTransferPackageAreas = transfer.DataTransferPackageAreas

type dataTransferPackageAreasV2 = transfer.PackageAreasV2

type dataTransferPackageAreasV1 = transfer.PackageAreasV1

type DataTransferSnapshot = transfer.DataTransferSnapshot

type TransferVehicleSetMember = transfer.TransferVehicleSetMember

type TransferVehicleSetDiagnostic = transfer.TransferVehicleSetDiagnostic

type TransferVehicleSet = transfer.TransferVehicleSet

type TransferVehicle = transfer.TransferVehicle

type transferVehicleV1 = transfer.LegacyVehicleV1

type TransferAccessory = transfer.TransferAccessory

type TransferAccessoryFingerprintState = transfer.TransferAccessoryFingerprintState

type TransferAccessoryReservationFingerprint = transfer.TransferAccessoryReservationFingerprint

type TransferAccessoryInstallationFingerprint = transfer.TransferAccessoryInstallationFingerprint

type TransferAccessoryConditionHistoryFingerprint = transfer.TransferAccessoryConditionHistoryFingerprint

type TransferAccessoryStock = transfer.TransferAccessoryStock

type TransferAccessoryAsset = transfer.TransferAccessoryAsset

type TransferExhibitionList = transfer.TransferExhibitionList

type TransferExhibitionEntry = transfer.TransferExhibitionEntry

type VehicleTransferValueKind = transfer.VehicleTransferValueKind

type VehicleTransferField = transfer.VehicleTransferField

type DataTransferCSVMappingOrigin = transfer.DataTransferCSVMappingOrigin

type DataTransferCSVColumnMapping = transfer.DataTransferCSVColumnMapping

type DataTransferCSVMappingInput = transfer.DataTransferCSVMappingInput

const (
	VehicleTransferString  = transfer.VehicleTransferString
	VehicleTransferInteger = transfer.VehicleTransferInteger
	VehicleTransferBoolean = transfer.VehicleTransferBoolean
	CSVMappingAlias        = transfer.CSVMappingAlias
	CSVMappingProfile      = transfer.CSVMappingProfile
	CSVMappingManual       = transfer.CSVMappingManual
	CSVMappingIgnored      = transfer.CSVMappingIgnored
	CSVMappingUnmapped     = transfer.CSVMappingUnmapped
)

type DataTransferVehicleSetPreview = transfer.DataTransferVehicleSetPreview

const DataTransferMaxUploadBytes int64 = 25 * 1024 * 1024

const DataTransferTargetFingerprintVersion = transfer.DataTransferTargetFingerprintVersion

var (
	ErrDataTransferUploadTooLarge = transfer.ErrDataTransferUploadTooLarge
	ErrDataTransferUploadPath     = transfer.ErrDataTransferUploadPath
)

type DataTransferPreviewRecord = transfer.DataTransferPreviewRecord

var (
	ErrDataTransferArtifactPath      = transfer.ErrDataTransferArtifactPath
	ErrDataTransferForbidden         = transfer.ErrDataTransferForbidden
	ErrDataTransferArtifactDeleted   = transfer.ErrDataTransferArtifactDeleted
	ErrDataTransferConflict          = transfer.ErrDataTransferConflict
	ErrDataTransferOpenUnavailable   = transfer.ErrDataTransferOpenUnavailable
	ErrDataTransferExportUnavailable = transfer.ErrDataTransferExportUnavailable
)

var (
	ErrDataTransferValidation = transfer.ErrDataTransferValidation
	ErrDataTransferNotFound   = transfer.ErrDataTransferNotFound
)

func DataTransferTargetFingerprint(value any) string {
	return transfer.DataTransferTargetFingerprint(value)
}

func DataTransferTargetFingerprintForVersion(value any, version int) string {
	return transfer.DataTransferTargetFingerprintForVersion(value, version)
}

type AggregateRules struct {
	Vehicle   func(TransferVehicle) error
	Accessory func(TransferAccessory) error
}

func (r AggregateRules) validateVehicle(v TransferVehicle) error {
	if r.Vehicle == nil {
		return ErrDataTransferValidation
	}
	return r.Vehicle(v)
}
func (r AggregateRules) validateAccessory(v TransferAccessory) error {
	if r.Accessory == nil {
		return ErrDataTransferValidation
	}
	return r.Accessory(v)
}

type VehicleSetInput = domain.VehicleSetInput
