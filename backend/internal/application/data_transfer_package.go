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

func transferVehicleFromV1(vehicle transferVehicleV1) TransferVehicle {
	return transfer.TransferVehicleFromV1(vehicle)
}
