package policy

import (
	"fmt"

	"strconv"
	"strings"
	"unicode"
)

func transferVehicleCSVHeaders() []string {
	fields := VehicleCSVTransferFields()
	headers := make([]string, len(fields))
	for index, field := range fields {
		headers[index] = field.LabelDE
	}
	return headers
}

type transferVehicleSetCSVMember struct {
	Set    TransferVehicleSet
	Member TransferVehicleSetMember
}

func transferVehicleSetCSVMemberLookups(
	sets []TransferVehicleSet,
) (map[string]*transferVehicleSetCSVMember, map[string]*transferVehicleSetCSVMember, error) {
	byID := map[string]*transferVehicleSetCSVMember{}
	byInventory := map[string]*transferVehicleSetCSVMember{}
	for _, set := range sets {
		for _, member := range set.Members {
			value := &transferVehicleSetCSVMember{Set: set, Member: member}
			if member.SourceVehicleID != "" {
				if byID[member.SourceVehicleID] != nil {
					return nil, nil, fmt.Errorf("%w: vehicle belongs to multiple exported sets", ErrDataTransferValidation)
				}
				byID[member.SourceVehicleID] = value
			}
			inventoryKey := transferIdentity(member.VehicleInventoryNumber)
			if inventoryKey != "" {
				if byInventory[inventoryKey] != nil {
					return nil, nil, fmt.Errorf("%w: vehicle belongs to multiple exported sets", ErrDataTransferValidation)
				}
				byInventory[inventoryKey] = value
			}
		}
	}
	return byID, byInventory, nil
}

func transferVehicleCSVValues(vehicle TransferVehicle, setMember *transferVehicleSetCSVMember) []string {
	fields := VehicleCSVTransferFields()
	values := make([]string, len(fields))
	for index, field := range fields {
		if strings.HasPrefix(field.Key, "vehicleSet") {
			values[index] = transferVehicleSetCSVValue(setMember, field.Key)
		} else {
			values[index] = transferVehicleCSVValue(vehicle, field.Key)
		}
	}
	return values
}

func transferVehicleSetCSVValue(member *transferVehicleSetCSVMember, field string) string {
	if member == nil {
		return ""
	}
	set := member.Set
	switch field {
	case "vehicleSetInventoryNumber":
		return set.InventoryNumber
	case "vehicleSetName":
		return set.Name
	case "vehicleSetManufacturer":
		return set.Manufacturer
	case "vehicleSetArticleNumber":
		return set.ArticleNumber
	case "vehicleSetArticleSourceUrl":
		return set.ArticleSourceURL
	case "vehicleSetGauge":
		return set.Gauge
	case "vehicleSetEpoch":
		return set.Epoch
	case "vehicleSetRailwayCompany":
		return set.RailwayCompany
	case "vehicleSetCategory":
		return set.Category
	case "vehicleSetGattung":
		return set.Gattung
	case "vehicleSetDescription":
		return set.Description
	case "vehicleSetEAN":
		return set.EAN
	case "vehicleSetProductionPeriod":
		return set.ProductionPeriod
	case "vehicleSetListPrice":
		return set.ListPrice
	case "vehicleSetAcquisitionType":
		return set.AcquisitionType
	case "vehicleSetAcquiredFrom":
		return set.AcquiredFrom
	case "vehicleSetPurchasePrice":
		return set.PurchasePrice
	case "vehicleSetPurchaseDate":
		return set.PurchaseDate
	case "vehicleSetStorageLocation":
		return set.StorageLocation
	case "vehicleSetStorageDetails":
		return set.StorageDetails
	case "vehicleSetCondition":
		return set.Condition
	case "vehicleSetConditionDetails":
		return set.ConditionDetails
	case "vehicleSetPackaging":
		return set.Packaging
	case "vehicleSetPosition":
		return strconv.Itoa(member.Member.Position)
	case "vehicleSetMemberCount":
		return strconv.Itoa(len(set.Members))
	case "vehicleSetMemberLabel":
		return member.Member.Label
	}
	return ""
}

func transferVehicleCSVValue(vehicle TransferVehicle, field string) string {
	switch field {
	case "inventoryNumber":
		return vehicle.InventoryNumber
	case "manufacturer":
		return vehicle.Manufacturer
	case "articleNumber":
		return vehicle.ArticleNumber
	case "articleSourceUrl":
		return vehicle.ArticleSourceURL
	case "name":
		return vehicle.Name
	case "gauge":
		return vehicle.Gauge
	case "epoch":
		return vehicle.Epoch
	case "railwayCompany":
		return vehicle.RailwayCompany
	case "category":
		return vehicle.Category
	case "gattung":
		return vehicle.Gattung
	case "description":
		return vehicle.Description
	case "series":
		return vehicle.Series
	case "vehicleNumber":
		return vehicle.VehicleNumber
	case "maximumSpeedKmh":
		if vehicle.MaximumSpeedKmh != nil {
			return strconv.Itoa(*vehicle.MaximumSpeedKmh)
		}
	case "homeBase":
		return vehicle.HomeBase
	case "digital":
		return transferCSVBoolean(vehicle.Digital)
	case "digitalDecoderNumber":
		return vehicle.DigitalDecoderNumber
	case "decoderType":
		return vehicle.DecoderType
	case "dtDecoder":
		return transferCSVBoolean(vehicle.DTDecoder)
	case "dtDecoderNumber":
		return vehicle.DTDecoderNumber
	case "exhibitionReady":
		return transferCSVBoolean(vehicle.ExhibitionReady)
	case "exhibition":
		return transferCSVBoolean(vehicle.Exhibition)
	case "abcBrakes":
		return transferCSVBoolean(vehicle.ABCBrakes)
	case "ean":
		return vehicle.EAN
	case "productionPeriod":
		return vehicle.ProductionPeriod
	case "listPrice":
		return vehicle.ListPrice
	case "acquisitionType":
		return vehicle.AcquisitionType
	case "acquiredFrom":
		return vehicle.AcquiredFrom
	case "purchasePrice":
		return vehicle.PurchasePrice
	case "purchaseDate":
		return vehicle.PurchaseDate
	case "storageLocation":
		return vehicle.StorageLocation
	case "storageDetails":
		return vehicle.StorageDetails
	case "condition":
		return vehicle.Condition
	case "conditionDetails":
		return vehicle.ConditionDetails
	case "packaging":
		return vehicle.Packaging
	case "lengthMm":
		return vehicle.LengthMM
	case "weightG":
		return vehicle.WeightG
	case "color":
		return vehicle.Color
	case "lettering":
		return vehicle.Lettering
	case "load":
		return vehicle.Load
	case "interior":
		return vehicle.Interior
	case "axles":
		return vehicle.Axles
	case "axleCount":
		return vehicle.AxleCount
	case "tractionTireCount":
		return vehicle.TractionTireCount
	case "wheelset":
		return vehicle.Wheelset
	case "couplingSame":
		return transferCSVBoolean(vehicle.CouplingSame)
	case "couplingFront":
		return vehicle.CouplingFront
	case "couplingRear":
		return vehicle.CouplingRear
	case "powerPickup":
		return vehicle.PowerPickup
	case "adapter":
		return vehicle.Adapter
	case "driveEnabled":
		return transferCSVBoolean(vehicle.DriveEnabled)
	case "driveDescription":
		return vehicle.DriveDescription
	case "headlightsEnabled":
		return transferCSVBoolean(vehicle.HeadlightsEnabled)
	case "headlightsDescription":
		return vehicle.HeadlightsDescription
	case "lightingEnabled":
		return transferCSVBoolean(vehicle.LightingEnabled)
	case "lightingDescription":
		return vehicle.LightingDescription
	case "soundGeneratorEnabled":
		return transferCSVBoolean(vehicle.SoundGeneratorEnabled)
	case "soundGeneratorDescription":
		return vehicle.SoundGeneratorDescription
	case "smokeGeneratorEnabled":
		return transferCSVBoolean(vehicle.SmokeGeneratorEnabled)
	case "smokeGeneratorDescription":
		return vehicle.SmokeGeneratorDescription
	case "additionalInfo":
		return vehicle.AdditionalInfo
	case "qrCodeEnabled":
		return transferCSVBoolean(vehicle.QRCodeEnabled)
	}
	return ""
}

func transferCSVBoolean(value bool) string {
	if value {
		return "Ja"
	}
	return "Nein"
}

func SafeDataTransferCSVRow(values []string) []string {
	protected := make([]string, len(values))
	for index, value := range values {
		candidate := strings.TrimLeftFunc(value, unicode.IsSpace)
		if candidate != "" && (candidate[0] == '\'' || strings.ContainsRune("=+-@", rune(candidate[0]))) {
			value = "'" + value
		}
		protected[index] = value
	}
	return protected
}
