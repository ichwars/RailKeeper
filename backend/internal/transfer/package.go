package transfer

import "encoding/json"

const (
	DataTransferPackageFormat          = "railkeeper-transfer"
	DataTransferPackageLegacyVersion   = 1
	DataTransferPackagePreviousVersion = 2
	DataTransferPackageVersion         = 3
)

type DataTransferPackage struct {
	Format    string                   `json:"format"`
	Version   int                      `json:"version"`
	CreatedAt string                   `json:"createdAt"`
	Areas     DataTransferPackageAreas `json:"areas"`
}

type DataTransferPackageAreas struct {
	Vehicles        []TransferVehicle        `json:"vehicles,omitempty"`
	VehicleSets     []TransferVehicleSet     `json:"vehicleSets,omitempty"`
	Accessories     []TransferAccessory      `json:"accessories,omitempty"`
	ExhibitionLists []TransferExhibitionList `json:"exhibitionLists,omitempty"`
}

// MarshalJSON preserves a selected but empty area as [] while omitting areas that were not selected.
func (areas DataTransferPackageAreas) MarshalJSON() ([]byte, error) {
	type areaDocument struct {
		Vehicles        *[]TransferVehicle        `json:"vehicles,omitempty"`
		VehicleSets     *[]TransferVehicleSet     `json:"vehicleSets,omitempty"`
		Accessories     *[]TransferAccessory      `json:"accessories,omitempty"`
		ExhibitionLists *[]TransferExhibitionList `json:"exhibitionLists,omitempty"`
	}
	document := areaDocument{}
	if areas.Vehicles != nil {
		document.Vehicles = &areas.Vehicles
	}
	if areas.VehicleSets != nil {
		document.VehicleSets = &areas.VehicleSets
	}
	if areas.Accessories != nil {
		document.Accessories = &areas.Accessories
	}
	if areas.ExhibitionLists != nil {
		document.ExhibitionLists = &areas.ExhibitionLists
	}
	return json.Marshal(document)
}

type DataTransferSnapshot struct {
	Vehicles        []TransferVehicle
	VehicleSets     []TransferVehicleSet
	Accessories     []TransferAccessory
	ExhibitionLists []TransferExhibitionList
}
