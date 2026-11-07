package transfer

type PackageAreasV2 struct {
	Vehicles        []TransferVehicle        `json:"vehicles,omitempty"`
	Accessories     []TransferAccessory      `json:"accessories,omitempty"`
	ExhibitionLists []TransferExhibitionList `json:"exhibitionLists,omitempty"`
}

type PackageAreasV1 struct {
	Vehicles        []LegacyVehicleV1        `json:"vehicles,omitempty"`
	Accessories     []TransferAccessory      `json:"accessories,omitempty"`
	ExhibitionLists []TransferExhibitionList `json:"exhibitionLists,omitempty"`
}

// LegacyVehicleV1 mirrors the vehicle JSON shape used by package version 1.
// Its field order must remain stable because legacy preview fingerprints were
// calculated from the encoded structure.
type LegacyVehicleV1 struct {
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

func LegacyTransferVehicle(vehicle TransferVehicle) LegacyVehicleV1 {
	return LegacyVehicleV1{
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

func TransferVehicleFromV1(vehicle LegacyVehicleV1) TransferVehicle {
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
