package policy

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
)

func ParseDataTransferUpload(
	job DataTransferJob,
	reader io.Reader,
	mapping *DataTransferCSVMappingInput,
) (DataTransferSnapshot, int, map[TransferArea][]int, []DataTransferCSVColumnMapping, error) {
	switch job.Format {
	case TransferJSON:
		packageDocument, err := DecodeDataTransferPackage(reader)
		if err != nil {
			return DataTransferSnapshot{}, 0, nil, nil, err
		}
		if err := validateDataTransferPackageSelection(packageDocument.Areas, job.Areas); err != nil {
			return DataTransferSnapshot{}, 0, nil, nil, err
		}
		return DataTransferSnapshot{
			Vehicles: packageDocument.Areas.Vehicles, VehicleSets: packageDocument.Areas.VehicleSets,
			Accessories:     packageDocument.Areas.Accessories,
			ExhibitionLists: packageDocument.Areas.ExhibitionLists,
		}, packageDocument.Version, nil, nil, nil
	case TransferCSV:
		if len(job.Areas) != 1 {
			return DataTransferSnapshot{}, 0, nil, nil, ErrDataTransferValidation
		}
		snapshot, rows, effectiveMapping, err := parseDataTransferCSVWithMapping(
			job.Areas[0], reader, mapping, dataTransferCSVMappingDefaults(job.Options),
		)
		return snapshot, DataTransferPackageVersion, map[TransferArea][]int{job.Areas[0]: rows}, effectiveMapping, err
	default:
		return DataTransferSnapshot{}, 0, nil, nil, ErrDataTransferValidation
	}
}

func DecodeDataTransferPackage(reader io.Reader) (DataTransferPackage, error) {
	envelope := struct {
		Format    string          `json:"format"`
		Version   int             `json:"version"`
		CreatedAt string          `json:"createdAt"`
		Areas     json.RawMessage `json:"areas"`
	}{}
	if err := decodeStrictTransferJSON(reader, &envelope); err != nil {
		return DataTransferPackage{}, fmt.Errorf("%w: invalid transfer package: %v", ErrDataTransferValidation, err)
	}
	if envelope.Format != DataTransferPackageFormat ||
		(envelope.Version != DataTransferPackageLegacyVersion &&
			envelope.Version != DataTransferPackagePreviousVersion &&
			envelope.Version != DataTransferPackageVersion) {
		return DataTransferPackage{}, fmt.Errorf("%w: unsupported transfer package format or version", ErrDataTransferValidation)
	}
	document := DataTransferPackage{
		Format: envelope.Format, Version: envelope.Version, CreatedAt: envelope.CreatedAt,
	}
	if envelope.Version == DataTransferPackageVersion {
		if err := decodeStrictTransferJSON(bytes.NewReader(envelope.Areas), &document.Areas); err != nil {
			return DataTransferPackage{}, fmt.Errorf("%w: invalid transfer package: %v", ErrDataTransferValidation, err)
		}
		return document, nil
	}
	if envelope.Version == DataTransferPackagePreviousVersion {
		previousAreas := dataTransferPackageAreasV2{}
		if err := decodeStrictTransferJSON(bytes.NewReader(envelope.Areas), &previousAreas); err != nil {
			return DataTransferPackage{}, fmt.Errorf("%w: invalid transfer package: %v", ErrDataTransferValidation, err)
		}
		document.Areas = DataTransferPackageAreas{
			Vehicles: previousAreas.Vehicles, Accessories: previousAreas.Accessories,
			ExhibitionLists: previousAreas.ExhibitionLists,
		}
		return document, nil
	}
	legacyAreas := dataTransferPackageAreasV1{}
	if err := decodeStrictTransferJSON(bytes.NewReader(envelope.Areas), &legacyAreas); err != nil {
		return DataTransferPackage{}, fmt.Errorf("%w: invalid transfer package: %v", ErrDataTransferValidation, err)
	}
	document.Areas = DataTransferPackageAreas{
		Accessories: legacyAreas.Accessories, ExhibitionLists: legacyAreas.ExhibitionLists,
	}
	if legacyAreas.Vehicles != nil {
		document.Areas.Vehicles = make([]TransferVehicle, len(legacyAreas.Vehicles))
		for index, vehicle := range legacyAreas.Vehicles {
			document.Areas.Vehicles[index] = transferVehicleFromV1(vehicle)
		}
	}
	return document, nil
}

func decodeStrictTransferJSON(reader io.Reader, target any) error {
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("trailing JSON data")
	}
	return nil
}

func validateDataTransferPackageSelection(areas DataTransferPackageAreas, selected []TransferArea) error {
	if areas.VehicleSets != nil && (areas.Vehicles == nil || !slices.Contains(selected, TransferVehicles)) {
		return fmt.Errorf("%w: vehicle sets require the vehicles area", ErrDataTransferValidation)
	}
	present := []TransferArea{}
	if areas.Vehicles != nil {
		present = append(present, TransferVehicles)
	}
	if areas.Accessories != nil {
		present = append(present, TransferAccessories)
	}
	if areas.ExhibitionLists != nil {
		present = append(present, TransferExhibitionLists)
	}
	if len(present) == 0 || len(present) != len(selected) {
		return fmt.Errorf("%w: package areas do not match the import profile", ErrDataTransferValidation)
	}
	for _, area := range selected {
		if !slices.Contains(present, area) {
			return fmt.Errorf("%w: package areas do not match the import profile", ErrDataTransferValidation)
		}
	}
	return nil
}
