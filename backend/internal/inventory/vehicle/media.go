package vehicle

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
