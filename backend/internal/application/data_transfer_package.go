package application

import "encoding/json"

const (
	DataTransferPackageFormat  = "railkeeper-transfer"
	DataTransferPackageVersion = 1
)

type DataTransferPackage struct {
	Format    string                   `json:"format"`
	Version   int                      `json:"version"`
	CreatedAt string                   `json:"createdAt"`
	Areas     DataTransferPackageAreas `json:"areas"`
}

type DataTransferPackageAreas struct {
	Vehicles        []TransferVehicle        `json:"vehicles,omitempty"`
	Accessories     []TransferAccessory      `json:"accessories,omitempty"`
	ExhibitionLists []TransferExhibitionList `json:"exhibitionLists,omitempty"`
}

// MarshalJSON preserves a selected but empty area as [] while omitting areas that were not selected.
func (areas DataTransferPackageAreas) MarshalJSON() ([]byte, error) {
	type areaDocument struct {
		Vehicles        *[]TransferVehicle        `json:"vehicles,omitempty"`
		Accessories     *[]TransferAccessory      `json:"accessories,omitempty"`
		ExhibitionLists *[]TransferExhibitionList `json:"exhibitionLists,omitempty"`
	}
	document := areaDocument{}
	if areas.Vehicles != nil {
		document.Vehicles = &areas.Vehicles
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
	Accessories     []TransferAccessory
	ExhibitionLists []TransferExhibitionList
}

type TransferVehicle struct {
	ID                   string `json:"id,omitempty"`
	InventoryNumber      string `json:"inventoryNumber"`
	Manufacturer         string `json:"manufacturer"`
	ArticleNumber        string `json:"articleNumber,omitempty"`
	ArticleSourceURL     string `json:"articleSourceUrl,omitempty"`
	Name                 string `json:"name"`
	Gauge                string `json:"gauge"`
	Epoch                string `json:"epoch,omitempty"`
	RailwayCompany       string `json:"railwayCompany,omitempty"`
	Category             string `json:"category,omitempty"`
	Gattung              string `json:"gattung,omitempty"`
	Description          string `json:"description,omitempty"`
	Series               string `json:"series,omitempty"`
	VehicleNumber        string `json:"vehicleNumber,omitempty"`
	MaximumSpeedKmh      *int   `json:"maximumSpeedKmh,omitempty"`
	HomeBase             string `json:"homeBase,omitempty"`
	Digital              bool   `json:"digital"`
	DigitalDecoderNumber string `json:"digitalDecoderNumber,omitempty"`
	DTDecoder            bool   `json:"dtDecoder"`
	DTDecoderNumber      string `json:"dtDecoderNumber,omitempty"`
	DecoderType          string `json:"decoderType,omitempty"`
	ExhibitionReady      bool   `json:"exhibitionReady"`
	Exhibition           bool   `json:"exhibition"`
	ABCBrakes            bool   `json:"abcBrakes"`
	EAN                  string `json:"ean,omitempty"`
	ProductionPeriod     string `json:"productionPeriod,omitempty"`
	ListPrice            string `json:"listPrice,omitempty"`
	AcquisitionType      string `json:"acquisitionType,omitempty"`
	AcquiredFrom         string `json:"acquiredFrom,omitempty"`
	PurchasePrice        string `json:"purchasePrice,omitempty"`
	PurchaseDate         string `json:"purchaseDate,omitempty"`
	StorageLocation      string `json:"storageLocation,omitempty"`
	StorageDetails       string `json:"storageDetails,omitempty"`
	Condition            string `json:"condition,omitempty"`
	ConditionDetails     string `json:"conditionDetails,omitempty"`
	Packaging            string `json:"packaging,omitempty"`
	CreatedAt            string `json:"createdAt,omitempty"`
	UpdatedAt            string `json:"updatedAt,omitempty"`
}

type TransferAccessory struct {
	ID                 string                            `json:"id,omitempty"`
	InventoryNumber    string                            `json:"inventoryNumber"`
	Manufacturer       string                            `json:"manufacturer"`
	ArticleNumber      string                            `json:"articleNumber,omitempty"`
	Name               string                            `json:"name"`
	Category           string                            `json:"category"`
	TrackingMode       string                            `json:"trackingMode"`
	Description        string                            `json:"description,omitempty"`
	EAN                string                            `json:"ean,omitempty"`
	ManufacturerStatus string                            `json:"manufacturerStatus,omitempty"`
	ArticleType        string                            `json:"articleType"`
	Subtype            string                            `json:"subtype,omitempty"`
	Gauges             []string                          `json:"gauges"`
	Scale              string                            `json:"scale,omitempty"`
	ListPrice          string                            `json:"listPrice,omitempty"`
	PackageQuantity    int                               `json:"packageQuantity"`
	StockUnit          string                            `json:"stockUnit"`
	MinimumStock       int                               `json:"minimumStock"`
	InventoryStrategy  string                            `json:"inventoryStrategy"`
	ManufacturerURL    string                            `json:"manufacturerUrl,omitempty"`
	ProductURL         string                            `json:"productUrl,omitempty"`
	AlternativeNumbers []string                          `json:"alternativeNumbers"`
	Keywords           []string                          `json:"keywords"`
	CompatibilityNotes string                            `json:"compatibilityNotes,omitempty"`
	InternalNotes      string                            `json:"internalNotes,omitempty"`
	Archived           bool                              `json:"archived"`
	Stock              []TransferAccessoryStock          `json:"stock"`
	Assets             []TransferAccessoryAsset          `json:"assets"`
	CreatedAt          string                            `json:"createdAt,omitempty"`
	UpdatedAt          string                            `json:"updatedAt,omitempty"`
	FingerprintState   TransferAccessoryFingerprintState `json:"-"`
}

type TransferAccessoryFingerprintState struct {
	Reservations     []TransferAccessoryReservationFingerprint      `json:"reservations"`
	Installations    []TransferAccessoryInstallationFingerprint     `json:"installations"`
	ConditionHistory []TransferAccessoryConditionHistoryFingerprint `json:"conditionHistory"`
}

type TransferAccessoryReservationFingerprint struct {
	ID         string `json:"id"`
	AssetID    string `json:"assetId,omitempty"`
	LocationID string `json:"locationId"`
	Quantity   int    `json:"quantity"`
	Status     string `json:"status"`
	UpdatedAt  string `json:"updatedAt"`
}

type TransferAccessoryInstallationFingerprint struct {
	ID               string `json:"id"`
	AssetID          string `json:"assetId,omitempty"`
	SourceLocationID string `json:"sourceLocationId"`
	Quantity         int    `json:"quantity"`
	Condition        string `json:"condition"`
	InstalledAt      string `json:"installedAt"`
	RemovedAt        string `json:"removedAt,omitempty"`
}

type TransferAccessoryConditionHistoryFingerprint struct {
	ID             string `json:"id"`
	InstallationID string `json:"installationId"`
	Condition      string `json:"condition"`
	ChangedAt      string `json:"changedAt"`
}

type TransferAccessoryStock struct {
	LocationID   string `json:"locationId"`
	LocationName string `json:"locationName"`
	Quantity     int    `json:"quantity"`
	UpdatedAt    string `json:"updatedAt,omitempty"`
}

type TransferAccessoryAsset struct {
	ID                  string `json:"id,omitempty"`
	InventoryNumber     string `json:"inventoryNumber,omitempty"`
	SerialNumber        string `json:"serialNumber,omitempty"`
	Condition           string `json:"condition"`
	Lifecycle           string `json:"lifecycle"`
	StorageLocationID   string `json:"storageLocationId,omitempty"`
	StorageLocationName string `json:"storageLocationName,omitempty"`
	PurchaseDate        string `json:"purchaseDate,omitempty"`
	PurchasePrice       string `json:"purchasePrice,omitempty"`
	WarrantyUntil       string `json:"warrantyUntil,omitempty"`
	Notes               string `json:"notes,omitempty"`
	CreatedAt           string `json:"createdAt,omitempty"`
	UpdatedAt           string `json:"updatedAt,omitempty"`
}

type TransferExhibitionList struct {
	ID          string                    `json:"id,omitempty"`
	Designation string                    `json:"designation"`
	Date        string                    `json:"date"`
	Locked      bool                      `json:"locked"`
	Entries     []TransferExhibitionEntry `json:"entries"`
	CreatedAt   string                    `json:"createdAt,omitempty"`
	UpdatedAt   string                    `json:"updatedAt,omitempty"`
}

type TransferExhibitionEntry struct {
	ID                     string `json:"id,omitempty"`
	VehicleID              string `json:"vehicleId,omitempty"`
	VehicleInventoryNumber string `json:"vehicleInventoryNumber,omitempty"`
	Owner                  string `json:"owner"`
	ImageURL               string `json:"imageUrl,omitempty"`
	LocomotiveName         string `json:"locomotiveName"`
	Gattung                string `json:"gattung,omitempty"`
	Series                 string `json:"series,omitempty"`
	Manufacturer           string `json:"manufacturer,omitempty"`
	Epoch                  string `json:"epoch,omitempty"`
	RailwayCompany         string `json:"railwayCompany,omitempty"`
	DayScope               string `json:"dayScope"`
	DTDecoder              bool   `json:"dtDecoder"`
	DecoderNumber          string `json:"decoderNumber,omitempty"`
	DecoderType            string `json:"decoderType,omitempty"`
	Adapter                string `json:"adapter,omitempty"`
	SXAddress              string `json:"sxAddress,omitempty"`
	Analog                 bool   `json:"analog"`
	FunctionKeys           string `json:"functionKeys,omitempty"`
	Notes                  string `json:"notes,omitempty"`
	SortOrder              int    `json:"sortOrder"`
	CreatedAt              string `json:"createdAt,omitempty"`
	UpdatedAt              string `json:"updatedAt,omitempty"`
}
