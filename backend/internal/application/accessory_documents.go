package application

import (
	"context"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

type AccessoryDocumentCategory string

const (
	AccessoryDocumentInvoice      AccessoryDocumentCategory = "invoice"
	AccessoryDocumentDeliveryNote AccessoryDocumentCategory = "delivery_note"
	AccessoryDocumentManual       AccessoryDocumentCategory = "manual"
	AccessoryDocumentDataSheet    AccessoryDocumentCategory = "data_sheet"
	AccessoryDocumentFloorPlan    AccessoryDocumentCategory = "floor_plan"
	AccessoryDocumentImage        AccessoryDocumentCategory = "image"
	AccessoryDocumentOther        AccessoryDocumentCategory = "other"
)

func (category AccessoryDocumentCategory) Valid() bool {
	switch category {
	case AccessoryDocumentInvoice, AccessoryDocumentDeliveryNote, AccessoryDocumentManual,
		AccessoryDocumentDataSheet, AccessoryDocumentFloorPlan, AccessoryDocumentImage,
		AccessoryDocumentOther:
		return true
	default:
		return false
	}
}

type AccessoryDocument struct {
	ID           string                    `json:"id"`
	ProductID    string                    `json:"productId"`
	FileBlobID   string                    `json:"fileBlobId"`
	FileName     string                    `json:"fileName"`
	OriginalName string                    `json:"originalName"`
	Description  string                    `json:"description,omitempty"`
	Category     AccessoryDocumentCategory `json:"category"`
	MimeType     string                    `json:"mimeType"`
	SizeBytes    int64                     `json:"sizeBytes"`
	IsPrimary    bool                      `json:"isPrimary"`
	CreatedBy    string                    `json:"createdBy"`
	CreatedAt    string                    `json:"createdAt"`
	UpdatedAt    string                    `json:"updatedAt"`
}

type AccessoryDocumentUploadMetadata struct {
	FileName     string                    `json:"fileName"`
	OriginalName string                    `json:"originalName"`
	Category     AccessoryDocumentCategory `json:"category"`
	MimeType     string                    `json:"mimeType"`
	SizeBytes    int64                     `json:"sizeBytes"`
	IsPrimary    bool                      `json:"isPrimary"`
}

type CreateAccessoryDocumentInput struct {
	ProductID  string `json:"productId"`
	FileBlobID string `json:"fileBlobId"`
	AccessoryDocumentUploadMetadata
	Description string `json:"description,omitempty"`
}

type UpdateAccessoryDocumentInput struct {
	Description string                    `json:"description,omitempty"`
	Category    AccessoryDocumentCategory `json:"category"`
	IsPrimary   bool                      `json:"isPrimary"`
}

type AccessoryDocumentRepository interface {
	ListDocuments(context.Context, string) ([]AccessoryDocument, error)
	GetDocument(context.Context, string) (*AccessoryDocument, error)
	CreateDocument(context.Context, CreateAccessoryDocumentInput, string) (*AccessoryDocument, error)
	UpdateDocument(context.Context, string, UpdateAccessoryDocumentInput, string) (*AccessoryDocument, error)
	DeleteDocument(context.Context, string, string) (string, error)
}

type FileBlobReferenceCleaner interface {
	DeleteIfUnreferenced(context.Context, string) error
}

type AccessoryDocumentService struct {
	repository AccessoryDocumentRepository
	blobs      FileBlobReferenceCleaner
}

func NewAccessoryDocumentService(
	repository AccessoryDocumentRepository,
	blobs FileBlobReferenceCleaner,
) *AccessoryDocumentService {
	return &AccessoryDocumentService{repository: repository, blobs: blobs}
}

func (s *AccessoryDocumentService) ListDocuments(
	ctx context.Context,
	productID string,
) ([]AccessoryDocument, error) {
	productID = strings.TrimSpace(productID)
	if productID == "" {
		return nil, ErrAccessoryValidation
	}
	return s.repository.ListDocuments(ctx, productID)
}

func (s *AccessoryDocumentService) GetDocument(ctx context.Context, id string) (*AccessoryDocument, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrAccessoryValidation
	}
	return s.repository.GetDocument(ctx, id)
}

func (s *AccessoryDocumentService) CreateDocument(
	ctx context.Context,
	input CreateAccessoryDocumentInput,
	maxSize int64,
	actor string,
) (*AccessoryDocument, error) {
	input = cleanAccessoryDocumentInput(input)
	if input.ProductID == "" || input.FileBlobID == "" ||
		ValidateAccessoryDocumentUpload(input.AccessoryDocumentUploadMetadata, maxSize) != nil {
		return nil, ErrAccessoryValidation
	}
	return s.repository.CreateDocument(ctx, input, actor)
}

func (s *AccessoryDocumentService) UpdateDocument(
	ctx context.Context,
	id string,
	input UpdateAccessoryDocumentInput,
	actor string,
) (*AccessoryDocument, error) {
	id = strings.TrimSpace(id)
	input.Description = strings.TrimSpace(input.Description)
	input.Category = AccessoryDocumentCategory(strings.TrimSpace(string(input.Category)))
	if id == "" || !input.Category.Valid() || (input.IsPrimary && input.Category != AccessoryDocumentImage) {
		return nil, ErrAccessoryValidation
	}
	return s.repository.UpdateDocument(ctx, id, input, actor)
}

func (s *AccessoryDocumentService) DeleteDocument(ctx context.Context, id, actor string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return ErrAccessoryValidation
	}
	blobID, err := s.repository.DeleteDocument(ctx, id, actor)
	if err != nil {
		return err
	}
	if s.blobs == nil {
		return nil
	}
	return s.blobs.DeleteIfUnreferenced(ctx, blobID)
}

var safeAccessoryDocumentFileName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]*$`)

func ValidateAccessoryDocumentUpload(metadata AccessoryDocumentUploadMetadata, maxSize int64) error {
	fileName := strings.TrimSpace(metadata.FileName)
	originalName := strings.TrimSpace(metadata.OriginalName)
	mimeType := normalizeAccessoryDocumentMime(metadata.MimeType)
	if metadata.SizeBytes <= 0 || maxSize <= 0 || metadata.SizeBytes > maxSize ||
		!metadata.Category.Valid() || (metadata.IsPrimary && metadata.Category != AccessoryDocumentImage) ||
		!validAccessoryDocumentBaseName(fileName, true) || !validAccessoryDocumentBaseName(originalName, false) ||
		blockedAccessoryDocumentExtension(fileName) || blockedAccessoryDocumentExtension(originalName) ||
		blockedAccessoryDocumentMime(mimeType) || !accessoryDocumentExtensionMatchesMime(fileName, mimeType) ||
		!accessoryDocumentExtensionMatchesMime(originalName, mimeType) {
		return ErrAccessoryValidation
	}
	return nil
}

func cleanAccessoryDocumentInput(input CreateAccessoryDocumentInput) CreateAccessoryDocumentInput {
	input.ProductID = strings.TrimSpace(input.ProductID)
	input.FileBlobID = strings.TrimSpace(input.FileBlobID)
	input.FileName = strings.TrimSpace(input.FileName)
	input.OriginalName = strings.TrimSpace(input.OriginalName)
	input.Description = strings.TrimSpace(input.Description)
	input.Category = AccessoryDocumentCategory(strings.TrimSpace(string(input.Category)))
	input.MimeType = normalizeAccessoryDocumentMime(input.MimeType)
	return input
}

func validAccessoryDocumentBaseName(value string, restricted bool) bool {
	if value == "" || value == "." || value == ".." || strings.ContainsAny(value, `/\`) ||
		filepath.Base(value) != value || strings.IndexFunc(value, unicode.IsControl) >= 0 {
		return false
	}
	return !restricted || safeAccessoryDocumentFileName.MatchString(value)
}

func blockedAccessoryDocumentExtension(value string) bool {
	switch strings.ToLower(filepath.Ext(value)) {
	case ".exe", ".bat", ".cmd", ".com", ".scr", ".msi", ".dll", ".ps1", ".vbs", ".js", ".jar", ".sh":
		return true
	default:
		return false
	}
}

func blockedAccessoryDocumentMime(value string) bool {
	return strings.Contains(value, "x-msdownload") || strings.Contains(value, "x-dosexec") ||
		strings.Contains(value, "x-sh") || strings.Contains(value, "javascript") ||
		strings.Contains(value, "ecmascript") || strings.Contains(value, "x-msdos-program")
}

func normalizeAccessoryDocumentMime(value string) string {
	return strings.ToLower(strings.TrimSpace(strings.Split(value, ";")[0]))
}

func accessoryDocumentExtensionMatchesMime(fileName, mimeType string) bool {
	switch strings.ToLower(filepath.Ext(fileName)) {
	case ".pdf":
		return mimeType == "application/pdf"
	case ".jpg", ".jpeg":
		return mimeType == "image/jpeg"
	case ".png":
		return mimeType == "image/png"
	case ".webp":
		return mimeType == "image/webp"
	case ".zip":
		return mimeType == "application/zip" || mimeType == "application/x-zip-compressed" ||
			mimeType == "application/octet-stream"
	case ".txt", ".csv", ".json", ".xml":
		return strings.HasPrefix(mimeType, "text/") || mimeType == "application/json" ||
			mimeType == "application/xml"
	default:
		return false
	}
}
