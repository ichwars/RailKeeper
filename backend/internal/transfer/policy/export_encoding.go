package policy

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"slices"

	"strings"
)

func MarshalDataTransferExport(
	job DataTransferJob,
	snapshot DataTransferSnapshot,
	createdAt string,
) ([]byte, string, string, error) {
	switch job.Format {
	case TransferJSON:
		payload, err := MarshalDataTransferPackage(snapshot, createdAt)
		return payload, "railkeeper-transfer.json", "application/json", err
	case TransferCSV:
		if len(job.Areas) != 1 {
			return nil, "", "", ErrDataTransferValidation
		}
		payload, err := MarshalDataTransferCSV(job.Areas[0], snapshot)
		return payload, "railkeeper-" + string(job.Areas[0]) + ".csv", "text/csv; charset=utf-8", err
	default:
		return nil, "", "", ErrDataTransferValidation
	}
}

func MarshalDataTransferPackage(snapshot DataTransferSnapshot, createdAt string) ([]byte, error) {
	prepareDataTransferSnapshotForExport(&snapshot)
	document := DataTransferPackage{
		Format: DataTransferPackageFormat, Version: DataTransferPackageVersion, CreatedAt: createdAt,
		Areas: DataTransferPackageAreas(snapshot),
	}
	payload, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode transfer package: %w", err)
	}
	return append(payload, '\n'), nil
}

func MarshalDataTransferCSV(area TransferArea, snapshot DataTransferSnapshot) ([]byte, error) {
	prepareDataTransferSnapshotForExport(&snapshot)
	var builder strings.Builder
	writer := csv.NewWriter(&builder)
	writer.Comma = ';'
	switch area {
	case TransferVehicles:
		if err := writer.Write(transferVehicleCSVHeaders()); err != nil {
			return nil, err
		}
		membersByID, membersByInventory, err := transferVehicleSetCSVMemberLookups(snapshot.VehicleSets)
		if err != nil {
			return nil, err
		}
		for _, vehicle := range snapshot.Vehicles {
			member := membersByID[vehicle.ID]
			if member == nil {
				member = membersByInventory[transferIdentity(vehicle.InventoryNumber)]
			}
			if err := writer.Write(SafeDataTransferCSVRow(transferVehicleCSVValues(vehicle, member))); err != nil {
				return nil, err
			}
		}
	case TransferAccessories:
		if err := writer.Write([]string{
			"Inventarnummer", "Hersteller", "Artikelnummer", "Bezeichnung", "Kategorie", "Erfassungsart",
			"Beschreibung", "EAN", "Artikelart", "Unterart", "Spurweiten", "Maßstab", "Listenpreis",
			"Packungsmenge", "Bestandseinheit", "Mindestbestand", "Inventarstrategie", "Bestand", "Einzelobjekte",
		}); err != nil {
			return nil, err
		}
		for _, accessory := range snapshot.Accessories {
			stock, err := json.Marshal(accessory.Stock)
			if err != nil {
				return nil, err
			}
			assets, err := json.Marshal(accessory.Assets)
			if err != nil {
				return nil, err
			}
			if err := writer.Write(SafeDataTransferCSVRow([]string{
				accessory.InventoryNumber, accessory.Manufacturer, accessory.ArticleNumber, accessory.Name,
				accessory.Category, accessory.TrackingMode, accessory.Description, accessory.EAN,
				accessory.ArticleType, accessory.Subtype, strings.Join(accessory.Gauges, ","), accessory.Scale,
				accessory.ListPrice, fmt.Sprintf("%d", accessory.PackageQuantity), accessory.StockUnit,
				fmt.Sprintf("%d", accessory.MinimumStock), accessory.InventoryStrategy, string(stock), string(assets),
			})); err != nil {
				return nil, err
			}
		}
	default:
		return nil, fmt.Errorf("%w: unsupported CSV area %q", ErrDataTransferValidation, area)
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("encode transfer CSV: %w", err)
	}
	return []byte(builder.String()), nil
}

func DataTransferAreasAllowed(selected, allowed []TransferArea) bool {
	if len(allowed) == 0 {
		return true
	}
	for _, area := range selected {
		if !slices.Contains(allowed, area) {
			return false
		}
	}
	return true
}

func DataTransferRecordCount(snapshot DataTransferSnapshot, areas []TransferArea) int {
	total := 0
	for _, area := range areas {
		switch area {
		case TransferVehicles:
			total += len(snapshot.Vehicles)
		case TransferAccessories:
			total += len(snapshot.Accessories)
		case TransferExhibitionLists:
			total += len(snapshot.ExhibitionLists)
		}
	}
	return total
}

func CloneTransferOptions(options map[string]any) map[string]any {
	if options == nil {
		return map[string]any{}
	}
	clone := make(map[string]any, len(options))
	for key, value := range options {
		clone[key] = value
	}
	return clone
}
