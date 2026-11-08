package application

import "railkeeper/backend/internal/transfer"

const (
	DataTransferPackageFormat          = transfer.DataTransferPackageFormat
	DataTransferPackageLegacyVersion   = transfer.DataTransferPackageLegacyVersion
	DataTransferPackagePreviousVersion = transfer.DataTransferPackagePreviousVersion
	DataTransferPackageVersion         = transfer.DataTransferPackageVersion
)

type DataTransferPackage = transfer.DataTransferPackage

type DataTransferPackageAreas = transfer.DataTransferPackageAreas

type DataTransferSnapshot = transfer.DataTransferSnapshot

type TransferVehicleSetMember = transfer.TransferVehicleSetMember

type TransferVehicleSetDiagnostic = transfer.TransferVehicleSetDiagnostic

type TransferVehicleSet = transfer.TransferVehicleSet

type TransferVehicle = transfer.TransferVehicle

type TransferAccessory = transfer.TransferAccessory

type TransferAccessoryFingerprintState = transfer.TransferAccessoryFingerprintState

type TransferAccessoryReservationFingerprint = transfer.TransferAccessoryReservationFingerprint

type TransferAccessoryInstallationFingerprint = transfer.TransferAccessoryInstallationFingerprint

type TransferAccessoryConditionHistoryFingerprint = transfer.TransferAccessoryConditionHistoryFingerprint

type TransferAccessoryStock = transfer.TransferAccessoryStock

type TransferAccessoryAsset = transfer.TransferAccessoryAsset

type TransferExhibitionList = transfer.TransferExhibitionList

type TransferExhibitionEntry = transfer.TransferExhibitionEntry
