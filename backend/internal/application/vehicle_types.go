package application

import (
	"context"
	"database/sql"
	"errors"
)

var (
	ErrVehicleValidation = errors.New("vehicle validation failed")
	ErrVehicleNotFound   = errors.New("vehicle not found")
	ErrVehicleImageInUse = errors.New("vehicle image in use")
)

var allowedMaintenanceKinds = map[string]struct{}{
	"Decoder-Einbau":   {},
	"Ersatzteiltausch": {},
	"Reinigung":        {},
	"Reparatur":        {},
	"Schmierung":       {},
	"Superung":         {},
	"Umbau":            {},
	"Wartung":          {},
}

var allowedMaintenanceStatuses = map[string]struct{}{
	"erledigt": {},
	"faellig":  {},
	"geplant":  {},
}

var allowedConditionRatings = map[string]struct{}{
	"gebraucht":          {},
	"gut":                {},
	"neuwertig":          {},
	"reparaturbedürftig": {},
	"sehr gut":           {},
}

var allowedFunctionTypes = map[string]struct{}{
	"kupplung":       {},
	"licht":          {},
	"rauch":          {},
	"sonderfunktion": {},
	"sound":          {},
	"standard":       {},
}

var allowedFunctionModes = map[string]struct{}{
	"dauer":  {},
	"moment": {},
}

type VehicleService struct {
	db             *sql.DB
	imageLocalizer VehicleImageLocalizer
}

type VehicleImageLocalizer func(ctx context.Context, vehicleID string, images []VehicleImageInput) ([]VehicleImageInput, error)

type Vehicle struct {
	ID                        string               `json:"id"`
	InventoryNumber           string               `json:"inventoryNumber"`
	Manufacturer              string               `json:"manufacturer"`
	ArticleNumber             string               `json:"articleNumber,omitempty"`
	ArticleSourceURL          string               `json:"articleSourceUrl,omitempty"`
	Name                      string               `json:"name"`
	Gauge                     string               `json:"gauge"`
	Epoch                     string               `json:"epoch,omitempty"`
	RailwayCompany            string               `json:"railwayCompany,omitempty"`
	Category                  string               `json:"category,omitempty"`
	Gattung                   string               `json:"gattung,omitempty"`
	Description               string               `json:"description,omitempty"`
	Series                    string               `json:"series,omitempty"`
	VehicleNumber             string               `json:"vehicleNumber,omitempty"`
	Digital                   bool                 `json:"digital"`
	DigitalDecoderNumber      string               `json:"digitalDecoderNumber,omitempty"`
	DTDecoder                 bool                 `json:"dtDecoder"`
	DTDecoderNumber           string               `json:"dtDecoderNumber,omitempty"`
	DecoderType               string               `json:"decoderType,omitempty"`
	ExhibitionReady           bool                 `json:"exhibitionReady"`
	Exhibition                bool                 `json:"exhibition"`
	ABCBrakes                 bool                 `json:"abcBrakes"`
	EAN                       string               `json:"ean,omitempty"`
	ProductionPeriod          string               `json:"productionPeriod,omitempty"`
	ListPrice                 string               `json:"listPrice,omitempty"`
	AcquisitionType           string               `json:"acquisitionType,omitempty"`
	AcquiredFrom              string               `json:"acquiredFrom,omitempty"`
	PurchasePrice             string               `json:"purchasePrice,omitempty"`
	PurchaseDate              string               `json:"purchaseDate,omitempty"`
	StorageLocation           string               `json:"storageLocation,omitempty"`
	StorageDetails            string               `json:"storageDetails,omitempty"`
	Condition                 string               `json:"condition,omitempty"`
	ConditionDetails          string               `json:"conditionDetails,omitempty"`
	Packaging                 string               `json:"packaging,omitempty"`
	LengthMM                  string               `json:"lengthMm,omitempty"`
	WeightG                   string               `json:"weightG,omitempty"`
	Color                     string               `json:"color,omitempty"`
	Lettering                 string               `json:"lettering,omitempty"`
	Load                      string               `json:"load,omitempty"`
	Interior                  string               `json:"interior,omitempty"`
	Axles                     string               `json:"axles,omitempty"`
	AxleCount                 string               `json:"axleCount,omitempty"`
	TractionTireCount         string               `json:"tractionTireCount,omitempty"`
	Wheelset                  string               `json:"wheelset,omitempty"`
	CouplingSame              bool                 `json:"couplingSame"`
	CouplingFront             string               `json:"couplingFront,omitempty"`
	CouplingRear              string               `json:"couplingRear,omitempty"`
	PowerPickup               string               `json:"powerPickup,omitempty"`
	Adapter                   string               `json:"adapter,omitempty"`
	DriveEnabled              bool                 `json:"driveEnabled"`
	DriveDescription          string               `json:"driveDescription,omitempty"`
	HeadlightsEnabled         bool                 `json:"headlightsEnabled"`
	HeadlightsDescription     string               `json:"headlightsDescription,omitempty"`
	LightingEnabled           bool                 `json:"lightingEnabled"`
	LightingDescription       string               `json:"lightingDescription,omitempty"`
	SoundGeneratorEnabled     bool                 `json:"soundGeneratorEnabled"`
	SoundGeneratorDescription string               `json:"soundGeneratorDescription,omitempty"`
	SmokeGeneratorEnabled     bool                 `json:"smokeGeneratorEnabled"`
	SmokeGeneratorDescription string               `json:"smokeGeneratorDescription,omitempty"`
	AdditionalInfo            string               `json:"additionalInfo,omitempty"`
	QRCodeEnabled             bool                 `json:"qrCodeEnabled"`
	Images                    []VehicleImage       `json:"images,omitempty"`
	Attachments               []VehicleAttachment  `json:"attachments,omitempty"`
	Maintenance               []VehicleMaintenance `json:"maintenance,omitempty"`
	SpareParts                []VehicleSparePart   `json:"spareParts,omitempty"`
	Functions                 []VehicleFunction    `json:"functions,omitempty"`
	CVValues                  []VehicleCVValue     `json:"cvValues,omitempty"`
	CVFiles                   []VehicleCVFile      `json:"cvFiles,omitempty"`
	ExternalMappings          []VehicleExternalMap `json:"externalMappings,omitempty"`
	CreatedAt                 string               `json:"createdAt"`
	UpdatedAt                 string               `json:"updatedAt"`
}

type VehicleExternalMap struct {
	ID               string `json:"id"`
	VehicleID        string `json:"vehicleId"`
	Provider         string `json:"provider"`
	ExternalID       string `json:"externalId"`
	ExternalName     string `json:"externalName,omitempty"`
	ExternalAddress  string `json:"externalAddress,omitempty"`
	ExternalProtocol string `json:"externalProtocol,omitempty"`
	SyncStatus       string `json:"syncStatus"`
	LastSeenAt       string `json:"lastSeenAt,omitempty"`
	CreatedAt        string `json:"createdAt"`
	UpdatedAt        string `json:"updatedAt"`
}

type VehicleExternalMapInput struct {
	Provider         string `json:"provider"`
	ExternalID       string `json:"externalId"`
	ExternalName     string `json:"externalName"`
	ExternalAddress  string `json:"externalAddress"`
	ExternalProtocol string `json:"externalProtocol"`
	SyncStatus       string `json:"syncStatus"`
}

type VehicleImage struct {
	ID              string `json:"id"`
	VehicleID       string `json:"vehicleId"`
	URL             string `json:"url"`
	ThumbnailURL    string `json:"thumbnailUrl,omitempty"`
	Title           string `json:"title,omitempty"`
	SourceURL       string `json:"sourceUrl,omitempty"`
	FileName        string `json:"fileName,omitempty"`
	MimeType        string `json:"mimeType,omitempty"`
	StoragePath     string `json:"-"`
	ThumbnailPath   string `json:"-"`
	BlobID          string `json:"-"`
	ThumbnailBlobID string `json:"-"`
	MaintenanceID   string `json:"maintenanceId,omitempty"`
	IsPrimary       bool   `json:"isPrimary"`
	SortOrder       int    `json:"sortOrder"`
	CreatedAt       string `json:"createdAt"`
	UpdatedAt       string `json:"updatedAt,omitempty"`
}

type VehicleImageInput struct {
	ID              string `json:"id"`
	URL             string `json:"url"`
	Title           string `json:"title"`
	SourceURL       string `json:"sourceUrl"`
	FileName        string `json:"-"`
	MimeType        string `json:"-"`
	StoragePath     string `json:"-"`
	ThumbnailPath   string `json:"-"`
	BlobID          string `json:"-"`
	ThumbnailBlobID string `json:"-"`
	MaintenanceID   string `json:"maintenanceId"`
	IsPrimary       bool   `json:"isPrimary"`
	SortOrder       int    `json:"sortOrder"`
}

type VehicleAttachment struct {
	ID            string `json:"id"`
	VehicleID     string `json:"vehicleId"`
	FileName      string `json:"fileName"`
	OriginalName  string `json:"originalName"`
	Description   string `json:"description,omitempty"`
	Category      string `json:"category,omitempty"`
	MimeType      string `json:"mimeType,omitempty"`
	SizeBytes     int64  `json:"sizeBytes"`
	StoragePath   string `json:"-"`
	BlobID        string `json:"-"`
	MaintenanceID string `json:"maintenanceId,omitempty"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

type VehicleAttachmentInput struct {
	FileName      string
	OriginalName  string
	Description   string
	Category      string
	MimeType      string
	SizeBytes     int64
	StoragePath   string
	BlobID        string
	MaintenanceID string
}

type VehicleAttachmentUpdateInput struct {
	Description   string `json:"description"`
	Category      string `json:"category"`
	MaintenanceID string `json:"maintenanceId"`
}

type VehicleMaintenance struct {
	ID              string `json:"id"`
	VehicleID       string `json:"vehicleId"`
	Kind            string `json:"kind"`
	Status          string `json:"status"`
	ConditionRating string `json:"conditionRating,omitempty"`
	DueDate         string `json:"dueDate,omitempty"`
	CompletedAt     string `json:"completedAt,omitempty"`
	Cost            string `json:"cost,omitempty"`
	Notes           string `json:"notes,omitempty"`
	CreatedAt       string `json:"createdAt"`
	UpdatedAt       string `json:"updatedAt"`
}

type VehicleMaintenanceInput struct {
	Kind            string `json:"kind"`
	Status          string `json:"status"`
	ConditionRating string `json:"conditionRating"`
	DueDate         string `json:"dueDate"`
	CompletedAt     string `json:"completedAt"`
	Cost            string `json:"cost"`
	Notes           string `json:"notes"`
}

type VehicleSparePart struct {
	ID            string `json:"id"`
	VehicleID     string `json:"vehicleId"`
	ArticleNumber string `json:"articleNumber"`
	Description   string `json:"description"`
	Price         string `json:"price,omitempty"`
	URL           string `json:"url,omitempty"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

type VehicleSparePartInput struct {
	ArticleNumber string `json:"articleNumber"`
	Description   string `json:"description"`
	Price         string `json:"price"`
	URL           string `json:"url"`
}

type VehicleFunction struct {
	ID                 string `json:"id"`
	VehicleID          string `json:"vehicleId"`
	FunctionKey        string `json:"functionKey"`
	Name               string `json:"name,omitempty"`
	SymbolKey          string `json:"symbolKey,omitempty"`
	FunctionType       string `json:"functionType"`
	Mode               string `json:"mode"`
	DirectionDependent bool   `json:"directionDependent"`
	Notes              string `json:"notes,omitempty"`
	SortOrder          int    `json:"sortOrder"`
	CreatedAt          string `json:"createdAt"`
	UpdatedAt          string `json:"updatedAt"`
}

type VehicleFunctionInput struct {
	Name               string `json:"name"`
	SymbolKey          string `json:"symbolKey"`
	FunctionType       string `json:"functionType"`
	Mode               string `json:"mode"`
	DirectionDependent bool   `json:"directionDependent"`
	Notes              string `json:"notes"`
}

type VehicleCVValue struct {
	ID             string                  `json:"id"`
	VehicleID      string                  `json:"vehicleId"`
	CVNumber       int                     `json:"cvNumber"`
	Value          int                     `json:"value"`
	Description    string                  `json:"description,omitempty"`
	Category       string                  `json:"category,omitempty"`
	Protocol       string                  `json:"protocol,omitempty"`
	DecoderProfile string                  `json:"decoderProfile,omitempty"`
	SourceFileID   string                  `json:"sourceFileId,omitempty"`
	CreatedAt      string                  `json:"createdAt"`
	UpdatedAt      string                  `json:"updatedAt"`
	History        []VehicleCVValueHistory `json:"history,omitempty"`
}

type VehicleCVValueHistory struct {
	ID        string `json:"id"`
	CVValueID string `json:"cvValueId"`
	VehicleID string `json:"vehicleId"`
	OldValue  int    `json:"oldValue"`
	NewValue  int    `json:"newValue"`
	ChangedAt string `json:"changedAt"`
}

type VehicleCVValueInput struct {
	CVNumber       int    `json:"cvNumber"`
	Value          int    `json:"value"`
	Description    string `json:"description"`
	Category       string `json:"category"`
	Protocol       string `json:"protocol"`
	DecoderProfile string `json:"decoderProfile"`
	SourceFileID   string `json:"sourceFileId"`
}

type VehicleCVFile struct {
	ID             string `json:"id"`
	VehicleID      string `json:"vehicleId"`
	FileName       string `json:"fileName"`
	OriginalName   string `json:"originalName"`
	Description    string `json:"description,omitempty"`
	DecoderProfile string `json:"decoderProfile,omitempty"`
	MimeType       string `json:"mimeType,omitempty"`
	SizeBytes      int64  `json:"sizeBytes"`
	StoragePath    string `json:"-"`
	BlobID         string `json:"-"`
	CreatedAt      string `json:"createdAt"`
	UpdatedAt      string `json:"updatedAt"`
}

type VehicleCVFileInput struct {
	FileName       string
	OriginalName   string
	Description    string
	DecoderProfile string
	MimeType       string
	SizeBytes      int64
	StoragePath    string
	BlobID         string
}

type CreateVehicleInput struct {
	InventoryNumber           string              `json:"inventoryNumber"`
	Manufacturer              string              `json:"manufacturer"`
	ArticleNumber             string              `json:"articleNumber"`
	ArticleSourceURL          string              `json:"articleSourceUrl"`
	Name                      string              `json:"name"`
	Gauge                     string              `json:"gauge"`
	Epoch                     string              `json:"epoch"`
	RailwayCompany            string              `json:"railwayCompany"`
	Category                  string              `json:"category"`
	Gattung                   string              `json:"gattung"`
	Description               string              `json:"description"`
	Series                    string              `json:"series"`
	VehicleNumber             string              `json:"vehicleNumber"`
	Digital                   bool                `json:"digital"`
	DigitalDecoderNumber      string              `json:"digitalDecoderNumber"`
	DTDecoder                 bool                `json:"dtDecoder"`
	DTDecoderNumber           string              `json:"dtDecoderNumber"`
	DecoderType               string              `json:"decoderType"`
	ExhibitionReady           bool                `json:"exhibitionReady"`
	Exhibition                bool                `json:"exhibition"`
	ABCBrakes                 bool                `json:"abcBrakes"`
	EAN                       string              `json:"ean"`
	ProductionPeriod          string              `json:"productionPeriod"`
	ListPrice                 string              `json:"listPrice"`
	AcquisitionType           string              `json:"acquisitionType"`
	AcquiredFrom              string              `json:"acquiredFrom"`
	PurchasePrice             string              `json:"purchasePrice"`
	PurchaseDate              string              `json:"purchaseDate"`
	StorageLocation           string              `json:"storageLocation"`
	StorageDetails            string              `json:"storageDetails"`
	Condition                 string              `json:"condition"`
	ConditionDetails          string              `json:"conditionDetails"`
	Packaging                 string              `json:"packaging"`
	LengthMM                  string              `json:"lengthMm"`
	WeightG                   string              `json:"weightG"`
	Color                     string              `json:"color"`
	Lettering                 string              `json:"lettering"`
	Load                      string              `json:"load"`
	Interior                  string              `json:"interior"`
	Axles                     string              `json:"axles"`
	AxleCount                 string              `json:"axleCount"`
	TractionTireCount         string              `json:"tractionTireCount"`
	Wheelset                  string              `json:"wheelset"`
	CouplingSame              bool                `json:"couplingSame"`
	CouplingFront             string              `json:"couplingFront"`
	CouplingRear              string              `json:"couplingRear"`
	PowerPickup               string              `json:"powerPickup"`
	Adapter                   string              `json:"adapter"`
	DriveEnabled              bool                `json:"driveEnabled"`
	DriveDescription          string              `json:"driveDescription"`
	HeadlightsEnabled         bool                `json:"headlightsEnabled"`
	HeadlightsDescription     string              `json:"headlightsDescription"`
	LightingEnabled           bool                `json:"lightingEnabled"`
	LightingDescription       string              `json:"lightingDescription"`
	SoundGeneratorEnabled     bool                `json:"soundGeneratorEnabled"`
	SoundGeneratorDescription string              `json:"soundGeneratorDescription"`
	SmokeGeneratorEnabled     bool                `json:"smokeGeneratorEnabled"`
	SmokeGeneratorDescription string              `json:"smokeGeneratorDescription"`
	AdditionalInfo            string              `json:"additionalInfo"`
	QRCodeEnabled             bool                `json:"qrCodeEnabled"`
	Images                    []VehicleImageInput `json:"images"`
}
