package application

import (
	"context"
	"database/sql"
	"errors"

	"railkeeper/backend/internal/decoder"
	"railkeeper/backend/internal/domain"
	"railkeeper/backend/internal/inventory/workshop"
)

var (
	ErrVehicleValidation              = domain.ErrVehicleValidation
	ErrVehicleOperationalValidation   = errors.New("vehicle operational validation failed")
	ErrVehicleNotFound                = domain.ErrVehicleNotFound
	ErrVehicleExternalMappingConflict = errors.New("vehicle external mapping conflict")
	ErrVehicleSetValidation           = errors.New("vehicle set validation failed")
	ErrVehicleSetNotFound             = errors.New("vehicle set not found")
	ErrVehicleSetImageValidation      = errors.New("vehicle set image validation failed")
	ErrVehicleSetImageNotFound        = errors.New("vehicle set image not found")
	ErrVehicleImageInUse              = errors.New("vehicle image in use")
)

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
	workshop       *workshop.Service
	workshopRows   workshop.Repository
}

type VehicleImageLocalizer func(ctx context.Context, vehicleID string, images []VehicleImageInput) ([]VehicleImageInput, error)

type Vehicle struct {
	Revision                  string                      `json:"revision,omitempty"`
	ImagesRevision            string                      `json:"imagesRevision,omitempty"`
	ID                        string                      `json:"id"`
	VehicleSetID              string                      `json:"vehicleSetId,omitempty"`
	VehicleSetName            string                      `json:"vehicleSetName,omitempty"`
	VehicleSetPosition        int                         `json:"vehicleSetPosition,omitempty"`
	VehicleSetSize            int                         `json:"vehicleSetSize,omitempty"`
	VehicleSet                *VehicleSetSummary          `json:"vehicleSet,omitempty"`
	InventoryNumber           string                      `json:"inventoryNumber"`
	Manufacturer              string                      `json:"manufacturer"`
	ArticleNumber             string                      `json:"articleNumber,omitempty"`
	ArticleSourceURL          string                      `json:"articleSourceUrl,omitempty"`
	Name                      string                      `json:"name"`
	Gauge                     string                      `json:"gauge"`
	Epoch                     string                      `json:"epoch,omitempty"`
	RailwayCompany            string                      `json:"railwayCompany,omitempty"`
	Category                  string                      `json:"category,omitempty"`
	Gattung                   string                      `json:"gattung,omitempty"`
	Description               string                      `json:"description,omitempty"`
	Series                    string                      `json:"series,omitempty"`
	VehicleNumber             string                      `json:"vehicleNumber,omitempty"`
	MaximumSpeedKmh           *int                        `json:"maximumSpeedKmh,omitempty"`
	HomeBase                  string                      `json:"homeBase,omitempty"`
	Digital                   bool                        `json:"digital"`
	DigitalDecoderNumber      string                      `json:"digitalDecoderNumber,omitempty"`
	DTDecoder                 bool                        `json:"dtDecoder"`
	DTDecoderNumber           string                      `json:"dtDecoderNumber,omitempty"`
	DecoderType               string                      `json:"decoderType,omitempty"`
	ExhibitionReady           bool                        `json:"exhibitionReady"`
	Exhibition                bool                        `json:"exhibition"`
	ABCBrakes                 bool                        `json:"abcBrakes"`
	EAN                       string                      `json:"ean,omitempty"`
	ProductionPeriod          string                      `json:"productionPeriod,omitempty"`
	ListPrice                 string                      `json:"listPrice,omitempty"`
	AcquisitionType           string                      `json:"acquisitionType,omitempty"`
	AcquiredFrom              string                      `json:"acquiredFrom,omitempty"`
	PurchasePrice             string                      `json:"purchasePrice,omitempty"`
	PurchaseDate              string                      `json:"purchaseDate,omitempty"`
	StorageLocation           string                      `json:"storageLocation,omitempty"`
	StorageDetails            string                      `json:"storageDetails,omitempty"`
	Condition                 string                      `json:"condition,omitempty"`
	ConditionDetails          string                      `json:"conditionDetails,omitempty"`
	Packaging                 string                      `json:"packaging,omitempty"`
	LengthMM                  string                      `json:"lengthMm,omitempty"`
	WeightG                   string                      `json:"weightG,omitempty"`
	Color                     string                      `json:"color,omitempty"`
	Lettering                 string                      `json:"lettering,omitempty"`
	Load                      string                      `json:"load,omitempty"`
	Interior                  string                      `json:"interior,omitempty"`
	Axles                     string                      `json:"axles,omitempty"`
	AxleCount                 string                      `json:"axleCount,omitempty"`
	TractionTireCount         string                      `json:"tractionTireCount,omitempty"`
	Wheelset                  string                      `json:"wheelset,omitempty"`
	CouplingSame              bool                        `json:"couplingSame"`
	CouplingFront             string                      `json:"couplingFront,omitempty"`
	CouplingRear              string                      `json:"couplingRear,omitempty"`
	PowerPickup               string                      `json:"powerPickup,omitempty"`
	Adapter                   string                      `json:"adapter,omitempty"`
	DriveEnabled              bool                        `json:"driveEnabled"`
	DriveDescription          string                      `json:"driveDescription,omitempty"`
	HeadlightsEnabled         bool                        `json:"headlightsEnabled"`
	HeadlightsDescription     string                      `json:"headlightsDescription,omitempty"`
	LightingEnabled           bool                        `json:"lightingEnabled"`
	LightingDescription       string                      `json:"lightingDescription,omitempty"`
	SoundGeneratorEnabled     bool                        `json:"soundGeneratorEnabled"`
	SoundGeneratorDescription string                      `json:"soundGeneratorDescription,omitempty"`
	SmokeGeneratorEnabled     bool                        `json:"smokeGeneratorEnabled"`
	SmokeGeneratorDescription string                      `json:"smokeGeneratorDescription,omitempty"`
	AdditionalInfo            string                      `json:"additionalInfo,omitempty"`
	QRCodeEnabled             bool                        `json:"qrCodeEnabled"`
	Images                    []VehicleImage              `json:"images,omitempty"`
	Attachments               []VehicleAttachment         `json:"attachments,omitempty"`
	Maintenance               []VehicleMaintenance        `json:"maintenance,omitempty"`
	SpareParts                []VehicleSparePart          `json:"spareParts,omitempty"`
	Functions                 []VehicleFunction           `json:"functions,omitempty"`
	CVValues                  []VehicleCVValue            `json:"cvValues,omitempty"`
	CVFiles                   []VehicleCVFile             `json:"cvFiles,omitempty"`
	ExternalMappings          []VehicleExternalMap        `json:"externalMappings,omitempty"`
	ECoSDecoderProfile        *ECoSDecoderProfileEvidence `json:"ecosDecoderProfile,omitempty"`
	CreatedAt                 string                      `json:"createdAt"`
	UpdatedAt                 string                      `json:"updatedAt"`
}

type VehicleSetSummary struct {
	ID              string               `json:"id"`
	InventoryNumber string               `json:"inventoryNumber"`
	Name            string               `json:"name"`
	Manufacturer    string               `json:"manufacturer"`
	ArticleNumber   string               `json:"articleNumber,omitempty"`
	Gauge           string               `json:"gauge"`
	Epoch           string               `json:"epoch,omitempty"`
	AcquisitionType string               `json:"acquisitionType,omitempty"`
	PurchaseDate    string               `json:"purchaseDate,omitempty"`
	PurchasePrice   string               `json:"purchasePrice,omitempty"`
	Condition       string               `json:"condition,omitempty"`
	MemberCount     int                  `json:"memberCount"`
	Position        int                  `json:"position"`
	MainImage       *VehicleSetMainImage `json:"mainImage,omitempty"`
}

type VehicleSetMainImageMode string

const (
	VehicleSetMainImageModeAutomatic VehicleSetMainImageMode = "automatic"
	VehicleSetMainImageModeMember    VehicleSetMainImageMode = "member"
	VehicleSetMainImageModeDedicated VehicleSetMainImageMode = "dedicated"
)

type VehicleSetMainImageInput struct {
	Mode          VehicleSetMainImageMode `json:"mode"`
	MemberImageID string                  `json:"memberImageId"`
}

type VehicleSetMainImage struct {
	Source       string `json:"source"`
	ImageID      string `json:"imageId,omitempty"`
	VehicleID    string `json:"vehicleId,omitempty"`
	URL          string `json:"url"`
	ThumbnailURL string `json:"thumbnailUrl,omitempty"`
	Title        string `json:"title,omitempty"`
}

type VehicleSetImage struct {
	URL             string `json:"url"`
	ThumbnailURL    string `json:"thumbnailUrl,omitempty"`
	FileName        string `json:"fileName"`
	MimeType        string `json:"mimeType"`
	BlobID          string `json:"-"`
	ThumbnailBlobID string `json:"-"`
	CreatedAt       string `json:"createdAt"`
	UpdatedAt       string `json:"updatedAt"`
}

type VehicleSetImageInput struct {
	FileName        string
	MimeType        string
	BlobID          string
	ThumbnailBlobID string
}

type VehicleSetInput = domain.VehicleSetInput

type CreateVehicleSetInput struct {
	CreationID string               `json:"creationId,omitempty"`
	Set        VehicleSetInput      `json:"set"`
	Members    []CreateVehicleInput `json:"members"`
}

type VehicleSet struct {
	ID                    string                  `json:"id"`
	InventoryNumber       string                  `json:"inventoryNumber"`
	Name                  string                  `json:"name"`
	Manufacturer          string                  `json:"manufacturer"`
	ArticleNumber         string                  `json:"articleNumber,omitempty"`
	ArticleSourceURL      string                  `json:"articleSourceUrl,omitempty"`
	Gauge                 string                  `json:"gauge"`
	Epoch                 string                  `json:"epoch,omitempty"`
	RailwayCompany        string                  `json:"railwayCompany,omitempty"`
	Category              string                  `json:"category"`
	Gattung               string                  `json:"gattung"`
	Description           string                  `json:"description,omitempty"`
	EAN                   string                  `json:"ean,omitempty"`
	ProductionPeriod      string                  `json:"productionPeriod,omitempty"`
	ListPrice             string                  `json:"listPrice,omitempty"`
	AcquisitionType       string                  `json:"acquisitionType,omitempty"`
	AcquiredFrom          string                  `json:"acquiredFrom,omitempty"`
	PurchasePrice         string                  `json:"purchasePrice,omitempty"`
	PurchaseDate          string                  `json:"purchaseDate,omitempty"`
	StorageLocation       string                  `json:"storageLocation,omitempty"`
	StorageDetails        string                  `json:"storageDetails,omitempty"`
	Condition             string                  `json:"condition,omitempty"`
	ConditionDetails      string                  `json:"conditionDetails,omitempty"`
	Packaging             string                  `json:"packaging,omitempty"`
	Members               []Vehicle               `json:"members"`
	CreatedAt             string                  `json:"createdAt"`
	UpdatedAt             string                  `json:"updatedAt"`
	MainImageMode         VehicleSetMainImageMode `json:"mainImageMode"`
	SelectedMemberImageID string                  `json:"selectedMemberImageId,omitempty"`
	DedicatedImage        *VehicleSetImage        `json:"dedicatedImage,omitempty"`
	MainImage             *VehicleSetMainImage    `json:"mainImage,omitempty"`
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
	Revision        string `json:"revision,omitempty"`
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

type VehicleImageCreation struct {
	VehicleImage
	ImagesRevision string `json:"imagesRevision"`
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
	Revision      string `json:"revision,omitempty"`
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
	ExpectedRevision *string `json:"expectedRevision,omitempty"`
	Description      string  `json:"description"`
	Category         string  `json:"category"`
	MaintenanceID    string  `json:"maintenanceId"`
}

type VehicleMaintenance = domain.VehicleMaintenance

type VehicleMaintenanceInput = domain.VehicleMaintenanceInput

type VehicleSparePart = domain.VehicleSparePart

type VehicleSparePartInput = domain.VehicleSparePartInput

type VehicleFunction struct {
	ID                       string `json:"id"`
	VehicleID                string `json:"vehicleId"`
	FunctionKey              string `json:"functionKey"`
	Name                     string `json:"name,omitempty"`
	SymbolKey                string `json:"symbolKey,omitempty"`
	FunctionType             string `json:"functionType"`
	Mode                     string `json:"mode"`
	DirectionDependent       *bool  `json:"directionDependent"`
	Inverted                 *bool  `json:"inverted"`
	LegacyDirectionDependent *bool  `json:"legacyDirectionDependent,omitempty"`
	FlagsVersion             int    `json:"flagsVersion"`
	Notes                    string `json:"notes,omitempty"`
	SortOrder                int    `json:"sortOrder"`
	CreatedAt                string `json:"createdAt"`
	UpdatedAt                string `json:"updatedAt"`
	Revision                 string `json:"revision"`
}

type VehicleFunctionInput struct {
	Name                     string `json:"name"`
	SymbolKey                string `json:"symbolKey"`
	FunctionType             string `json:"functionType"`
	Mode                     string `json:"mode"`
	DirectionDependent       *bool  `json:"directionDependent"`
	Inverted                 *bool  `json:"inverted"`
	LegacyDirectionDependent *bool  `json:"legacyDirectionDependent,omitempty"`
	FlagsVersion             int    `json:"flagsVersion"`
	Notes                    string `json:"notes"`
}

// An absent expected revision means the key must not exist yet.
type VehicleFunctionChange struct {
	FunctionKey      string                `json:"functionKey"`
	ExpectedRevision string                `json:"expectedRevision"`
	Value            *VehicleFunctionInput `json:"value,omitempty"`
	Delete           bool                  `json:"delete,omitempty"`
}

type VehicleCVValue = decoder.VehicleCVValue

type VehicleCVValueHistory = decoder.VehicleCVValueHistory

type VehicleCVValueInput = decoder.VehicleCVValueInput

type VehicleCVFile = decoder.VehicleCVFile

type VehicleCVFileInput = decoder.VehicleCVFileInput

type CreateVehicleInput struct {
	ExpectedRevision          *string             `json:"expectedRevision,omitempty"`
	ExpectedImagesRevision    *string             `json:"expectedImagesRevision,omitempty"`
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
	MaximumSpeedKmh           *int                `json:"maximumSpeedKmh,omitempty"`
	HomeBase                  string              `json:"homeBase"`
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

	FunctionChanges []VehicleFunctionChange `json:"functionChanges,omitempty"`
}
