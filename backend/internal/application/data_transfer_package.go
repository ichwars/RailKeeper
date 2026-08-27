package application

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

type dataTransferPackageAreasV2 struct {
	Vehicles        []TransferVehicle        `json:"vehicles,omitempty"`
	Accessories     []TransferAccessory      `json:"accessories,omitempty"`
	ExhibitionLists []TransferExhibitionList `json:"exhibitionLists,omitempty"`
}

type dataTransferPackageAreasV1 struct {
	Vehicles        []transferVehicleV1      `json:"vehicles,omitempty"`
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

type TransferVehicleSetMember struct {
	SourceVehicleID        string `json:"vehicleId"`
	VehicleInventoryNumber string `json:"vehicleInventoryNumber"`
	Position               int    `json:"position"`
	Label                  string `json:"label,omitempty"`
}

type TransferVehicleSet struct {
	ID              string `json:"id,omitempty"`
	InventoryNumber string `json:"inventoryNumber"`
	VehicleSetInput
	Members   []TransferVehicleSetMember `json:"members"`
	CreatedAt string                     `json:"createdAt,omitempty"`
	UpdatedAt string                     `json:"updatedAt,omitempty"`
}

type TransferVehicle struct {
	ID                        string `json:"id,omitempty"`
	InventoryNumber           string `json:"inventoryNumber"`
	Manufacturer              string `json:"manufacturer"`
	ArticleNumber             string `json:"articleNumber,omitempty"`
	ArticleSourceURL          string `json:"articleSourceUrl,omitempty"`
	Name                      string `json:"name"`
	Gauge                     string `json:"gauge"`
	Epoch                     string `json:"epoch,omitempty"`
	RailwayCompany            string `json:"railwayCompany,omitempty"`
	Category                  string `json:"category,omitempty"`
	Gattung                   string `json:"gattung,omitempty"`
	Description               string `json:"description,omitempty"`
	Series                    string `json:"series,omitempty"`
	VehicleNumber             string `json:"vehicleNumber,omitempty"`
	MaximumSpeedKmh           *int   `json:"maximumSpeedKmh,omitempty"`
	HomeBase                  string `json:"homeBase,omitempty"`
	Digital                   bool   `json:"digital"`
	DigitalDecoderNumber      string `json:"digitalDecoderNumber,omitempty"`
	DTDecoder                 bool   `json:"dtDecoder"`
	DTDecoderNumber           string `json:"dtDecoderNumber,omitempty"`
	DecoderType               string `json:"decoderType,omitempty"`
	ExhibitionReady           bool   `json:"exhibitionReady"`
	Exhibition                bool   `json:"exhibition"`
	ABCBrakes                 bool   `json:"abcBrakes"`
	EAN                       string `json:"ean,omitempty"`
	ProductionPeriod          string `json:"productionPeriod,omitempty"`
	ListPrice                 string `json:"listPrice,omitempty"`
	AcquisitionType           string `json:"acquisitionType,omitempty"`
	AcquiredFrom              string `json:"acquiredFrom,omitempty"`
	PurchasePrice             string `json:"purchasePrice,omitempty"`
	PurchaseDate              string `json:"purchaseDate,omitempty"`
	StorageLocation           string `json:"storageLocation,omitempty"`
	StorageDetails            string `json:"storageDetails,omitempty"`
	Condition                 string `json:"condition,omitempty"`
	ConditionDetails          string `json:"conditionDetails,omitempty"`
	Packaging                 string `json:"packaging,omitempty"`
	LengthMM                  string `json:"lengthMm,omitempty"`
	WeightG                   string `json:"weightG,omitempty"`
	Color                     string `json:"color,omitempty"`
	Lettering                 string `json:"lettering,omitempty"`
	Load                      string `json:"load,omitempty"`
	Interior                  string `json:"interior,omitempty"`
	Axles                     string `json:"axles,omitempty"`
	AxleCount                 string `json:"axleCount,omitempty"`
	TractionTireCount         string `json:"tractionTireCount,omitempty"`
	Wheelset                  string `json:"wheelset,omitempty"`
	CouplingSame              bool   `json:"couplingSame"`
	CouplingFront             string `json:"couplingFront,omitempty"`
	CouplingRear              string `json:"couplingRear,omitempty"`
	PowerPickup               string `json:"powerPickup,omitempty"`
	Adapter                   string `json:"adapter,omitempty"`
	DriveEnabled              bool   `json:"driveEnabled"`
	DriveDescription          string `json:"driveDescription,omitempty"`
	HeadlightsEnabled         bool   `json:"headlightsEnabled"`
	HeadlightsDescription     string `json:"headlightsDescription,omitempty"`
	LightingEnabled           bool   `json:"lightingEnabled"`
	LightingDescription       string `json:"lightingDescription,omitempty"`
	SoundGeneratorEnabled     bool   `json:"soundGeneratorEnabled"`
	SoundGeneratorDescription string `json:"soundGeneratorDescription,omitempty"`
	SmokeGeneratorEnabled     bool   `json:"smokeGeneratorEnabled"`
	SmokeGeneratorDescription string `json:"smokeGeneratorDescription,omitempty"`
	AdditionalInfo            string `json:"additionalInfo,omitempty"`
	QRCodeEnabled             bool   `json:"qrCodeEnabled"`
	CreatedAt                 string `json:"createdAt,omitempty"`
	UpdatedAt                 string `json:"updatedAt,omitempty"`
}

// transferVehicleV1 mirrors the vehicle JSON shape used by package version 1.
// Its field order must remain stable because legacy preview fingerprints were
// calculated from the encoded structure.
type transferVehicleV1 struct {
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

func legacyTransferVehicle(vehicle TransferVehicle) transferVehicleV1 {
	return transferVehicleV1{
		ID: vehicle.ID, InventoryNumber: vehicle.InventoryNumber, Manufacturer: vehicle.Manufacturer,
		ArticleNumber: vehicle.ArticleNumber, ArticleSourceURL: vehicle.ArticleSourceURL, Name: vehicle.Name,
		Gauge: vehicle.Gauge, Epoch: vehicle.Epoch, RailwayCompany: vehicle.RailwayCompany,
		Category: vehicle.Category, Gattung: vehicle.Gattung, Description: vehicle.Description,
		Series: vehicle.Series, VehicleNumber: vehicle.VehicleNumber, MaximumSpeedKmh: vehicle.MaximumSpeedKmh,
		HomeBase: vehicle.HomeBase, Digital: vehicle.Digital, DigitalDecoderNumber: vehicle.DigitalDecoderNumber,
		DTDecoder: vehicle.DTDecoder, DTDecoderNumber: vehicle.DTDecoderNumber, DecoderType: vehicle.DecoderType,
		ExhibitionReady: vehicle.ExhibitionReady, Exhibition: vehicle.Exhibition, ABCBrakes: vehicle.ABCBrakes,
		EAN: vehicle.EAN, ProductionPeriod: vehicle.ProductionPeriod, ListPrice: vehicle.ListPrice,
		AcquisitionType: vehicle.AcquisitionType, AcquiredFrom: vehicle.AcquiredFrom,
		PurchasePrice: vehicle.PurchasePrice, PurchaseDate: vehicle.PurchaseDate,
		StorageLocation: vehicle.StorageLocation, StorageDetails: vehicle.StorageDetails,
		Condition: vehicle.Condition, ConditionDetails: vehicle.ConditionDetails, Packaging: vehicle.Packaging,
		CreatedAt: vehicle.CreatedAt, UpdatedAt: vehicle.UpdatedAt,
	}
}

func transferVehicleFromV1(vehicle transferVehicleV1) TransferVehicle {
	return TransferVehicle{
		ID: vehicle.ID, InventoryNumber: vehicle.InventoryNumber, Manufacturer: vehicle.Manufacturer,
		ArticleNumber: vehicle.ArticleNumber, ArticleSourceURL: vehicle.ArticleSourceURL, Name: vehicle.Name,
		Gauge: vehicle.Gauge, Epoch: vehicle.Epoch, RailwayCompany: vehicle.RailwayCompany,
		Category: vehicle.Category, Gattung: vehicle.Gattung, Description: vehicle.Description,
		Series: vehicle.Series, VehicleNumber: vehicle.VehicleNumber, MaximumSpeedKmh: vehicle.MaximumSpeedKmh,
		HomeBase: vehicle.HomeBase, Digital: vehicle.Digital, DigitalDecoderNumber: vehicle.DigitalDecoderNumber,
		DTDecoder: vehicle.DTDecoder, DTDecoderNumber: vehicle.DTDecoderNumber, DecoderType: vehicle.DecoderType,
		ExhibitionReady: vehicle.ExhibitionReady, Exhibition: vehicle.Exhibition, ABCBrakes: vehicle.ABCBrakes,
		EAN: vehicle.EAN, ProductionPeriod: vehicle.ProductionPeriod, ListPrice: vehicle.ListPrice,
		AcquisitionType: vehicle.AcquisitionType, AcquiredFrom: vehicle.AcquiredFrom,
		PurchasePrice: vehicle.PurchasePrice, PurchaseDate: vehicle.PurchaseDate,
		StorageLocation: vehicle.StorageLocation, StorageDetails: vehicle.StorageDetails,
		Condition: vehicle.Condition, ConditionDetails: vehicle.ConditionDetails, Packaging: vehicle.Packaging,
		CreatedAt: vehicle.CreatedAt, UpdatedAt: vehicle.UpdatedAt,
	}
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
	ID                string                    `json:"id,omitempty"`
	Designation       string                    `json:"designation"`
	Date              string                    `json:"date"`
	EndDate           string                    `json:"endDate,omitempty"`
	Location          string                    `json:"location,omitempty"`
	Description       string                    `json:"description,omitempty"`
	OrganizationNotes string                    `json:"organizationNotes,omitempty"`
	Status            ExhibitionStatus          `json:"status,omitempty"`
	Locked            bool                      `json:"locked"`
	LockReason        string                    `json:"lockReason,omitempty"`
	LockedAt          string                    `json:"lockedAt,omitempty"`
	CompletedAt       string                    `json:"completedAt,omitempty"`
	ArchivedAt        string                    `json:"archivedAt,omitempty"`
	Entries           []TransferExhibitionEntry `json:"entries"`
	CreatedAt         string                    `json:"createdAt,omitempty"`
	UpdatedAt         string                    `json:"updatedAt,omitempty"`
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
	InterfaceName          string `json:"interfaceName,omitempty"`
	SXAddress              string `json:"sxAddress,omitempty"`
	Analog                 bool   `json:"analog"`
	Availability           string `json:"availability,omitempty"`
	FunctionKeys           string `json:"functionKeys,omitempty"`
	Notes                  string `json:"notes,omitempty"`
	SortOrder              int    `json:"sortOrder"`
	CreatedAt              string `json:"createdAt,omitempty"`
	UpdatedAt              string `json:"updatedAt,omitempty"`
}
