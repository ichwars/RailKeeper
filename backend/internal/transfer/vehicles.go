package transfer

type TransferVehicleSetMember struct {
	SourceVehicleID        string          `json:"vehicleId"`
	VehicleInventoryNumber string          `json:"vehicleInventoryNumber"`
	Position               int             `json:"position"`
	Label                  string          `json:"label,omitempty"`
	SourceRowNumber        int             `json:"-"`
	DeclaredMemberCount    int             `json:"-"`
	SourceSetInput         VehicleSetInput `json:"-"`
}

type TransferVehicleSetDiagnostic struct {
	RowNumber int    `json:"rowNumber"`
	Field     string `json:"field"`
	Code      string `json:"code"`
}

type TransferVehicleSet struct {
	ID              string `json:"id,omitempty"`
	InventoryNumber string `json:"inventoryNumber"`
	VehicleSetInput
	Members     []TransferVehicleSetMember     `json:"members"`
	Diagnostics []TransferVehicleSetDiagnostic `json:"-"`
	CreatedAt   string                         `json:"createdAt,omitempty"`
	UpdatedAt   string                         `json:"updatedAt,omitempty"`
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
