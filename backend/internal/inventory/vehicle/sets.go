package vehicle

import (
	"railkeeper/backend/internal/domain"
)

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

const (
	VehicleSetMainImageModeAutomatic VehicleSetMainImageMode = "automatic"
	VehicleSetMainImageModeMember    VehicleSetMainImageMode = "member"
	VehicleSetMainImageModeDedicated VehicleSetMainImageMode = "dedicated"
)
