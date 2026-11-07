package decoder

type VehicleCVValue struct {
	ID               string                  `json:"id"`
	Revision         string                  `json:"revision,omitempty"`
	VehicleID        string                  `json:"vehicleId"`
	CVNumber         int                     `json:"cvNumber"`
	IndexCV31        *int                    `json:"indexCV31,omitempty"`
	IndexCV32        *int                    `json:"indexCV32,omitempty"`
	Value            int                     `json:"value"`
	Description      string                  `json:"description,omitempty"`
	Category         string                  `json:"category,omitempty"`
	Protocol         string                  `json:"protocol,omitempty"`
	DecoderProfile   string                  `json:"decoderProfile,omitempty"`
	SourceFileID     string                  `json:"sourceFileId,omitempty"`
	SourceLine       *int                    `json:"sourceLine,omitempty"`
	SourceExportedAt string                  `json:"sourceExportedAt,omitempty"`
	CreatedAt        string                  `json:"createdAt"`
	UpdatedAt        string                  `json:"updatedAt"`
	History          []VehicleCVValueHistory `json:"history,omitempty"`
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
	// Original-group comparisons come from a PUT query or an outer batch item, never source files.
	SourceWorkspaceID               string  `json:"-"`
	ExpectedSourceWorkspaceRevision string  `json:"-"`
	ExpectedRevision                *string `json:"expectedRevision,omitempty"`
	WorkspaceID                     string  `json:"workspaceId,omitempty"`
	ExpectedWorkspaceRevision       string  `json:"expectedWorkspaceRevision,omitempty"`
	CVNumber                        int     `json:"cvNumber"`
	IndexCV31                       *int    `json:"indexCV31,omitempty"`
	IndexCV32                       *int    `json:"indexCV32,omitempty"`
	Value                           int     `json:"value"`
	Description                     string  `json:"description"`
	Category                        string  `json:"category"`
	Protocol                        string  `json:"protocol"`
	DecoderProfile                  string  `json:"decoderProfile"`
	SourceFileID                    string  `json:"sourceFileId"`
	SourceLine                      *int    `json:"sourceLine,omitempty"`
	SourceExportedAt                string  `json:"sourceExportedAt,omitempty"`
}

type VehicleCVFile struct {
	SourceReferenceCount int    `json:"sourceReferenceCount,omitempty"`
	Revision             string `json:"revision,omitempty"`
	ID                   string `json:"id"`
	VehicleID            string `json:"vehicleId"`
	FileName             string `json:"fileName"`
	OriginalName         string `json:"originalName"`
	Description          string `json:"description,omitempty"`
	DecoderProfile       string `json:"decoderProfile,omitempty"`
	MimeType             string `json:"mimeType,omitempty"`
	SizeBytes            int64  `json:"sizeBytes"`
	StoragePath          string `json:"-"`
	BlobID               string `json:"-"`
	CreatedAt            string `json:"createdAt"`
	UpdatedAt            string `json:"updatedAt"`
}

type VehicleCVFileInput struct {
	WorkspaceID               string
	ExpectedWorkspaceRevision string
	FileName                  string
	OriginalName              string
	Description               string
	DecoderProfile            string
	MimeType                  string
	SizeBytes                 int64
	StoragePath               string
	BlobID                    string
}
