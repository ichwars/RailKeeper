package application_test

import (
	"context"
	"errors"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestAccessoryDocumentUploadMetadataValidation(t *testing.T) {
	valid := application.AccessoryDocumentUploadMetadata{
		FileName: "manual.pdf", OriginalName: "Bedienungsanleitung.pdf",
		Category: application.AccessoryDocumentManual, MimeType: "application/pdf", SizeBytes: 8,
	}
	if err := application.ValidateAccessoryDocumentUpload(valid, 8); err != nil {
		t.Fatalf("valid document metadata rejected: %v", err)
	}

	tests := []struct {
		name     string
		metadata application.AccessoryDocumentUploadMetadata
		maxSize  int64
	}{
		{"empty", application.AccessoryDocumentUploadMetadata{}, 8},
		{"too large", withDocumentSize(valid, 9), 8},
		{"path", withDocumentFileName(valid, `..\manual.pdf`), 8},
		{"executable extension", withDocumentFileName(valid, "setup.exe"), 8},
		{"executable original extension", withDocumentOriginalName(valid, "setup.exe"), 8},
		{"original mime mismatch", withDocumentOriginalName(valid, "manual.png"), 8},
		{"original control character", withDocumentOriginalName(valid, "manual\n.pdf"), 8},
		{"script mime", withDocumentMime(valid, "application/javascript"), 8},
		{"mime mismatch", withDocumentMime(valid, "image/png"), 8},
		{"unknown category", withDocumentCategory(valid, "brochure"), 8},
		{"non-image primary", withDocumentPrimary(valid, true), 8},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := application.ValidateAccessoryDocumentUpload(test.metadata, test.maxSize); !errors.Is(err, application.ErrAccessoryValidation) {
				t.Fatalf("expected validation error, got %v", err)
			}
		})
	}

	image := application.AccessoryDocumentUploadMetadata{
		FileName: "front.webp", OriginalName: "front.webp", Category: application.AccessoryDocumentImage,
		MimeType: "image/webp; charset=binary", SizeBytes: 8, IsPrimary: true,
	}
	if err := application.ValidateAccessoryDocumentUpload(image, 8); err != nil {
		t.Fatalf("valid primary image rejected: %v", err)
	}
}

func TestAccessoryDocumentDeleteRemovesMetadataBeforeUnreferencedBlob(t *testing.T) {
	repository := &accessoryDocumentRepositorySpy{blobID: "blob-1"}
	cleaner := &accessoryBlobCleanerSpy{repository: repository}
	service := application.NewAccessoryDocumentService(repository, cleaner)

	if err := service.DeleteDocument(t.Context(), " document-1 ", "editor-1"); err != nil {
		t.Fatal(err)
	}
	if repository.deletedID != "document-1" || cleaner.deletedBlobID != "blob-1" {
		t.Fatalf("unexpected delete calls: document=%q blob=%q", repository.deletedID, cleaner.deletedBlobID)
	}
	if !cleaner.metadataWasDeleted {
		t.Fatal("blob cleanup ran before document metadata deletion")
	}
}

type accessoryDocumentRepositorySpy struct {
	blobID    string
	deletedID string
}

func (s *accessoryDocumentRepositorySpy) ListDocuments(context.Context, string) ([]application.AccessoryDocument, error) {
	return []application.AccessoryDocument{}, nil
}

func (s *accessoryDocumentRepositorySpy) GetDocument(context.Context, string) (*application.AccessoryDocument, error) {
	return &application.AccessoryDocument{}, nil
}

func (s *accessoryDocumentRepositorySpy) CreateDocument(
	context.Context, application.CreateAccessoryDocumentInput, string,
) (*application.AccessoryDocument, error) {
	return &application.AccessoryDocument{}, nil
}

func (s *accessoryDocumentRepositorySpy) UpdateDocument(
	context.Context, string, application.UpdateAccessoryDocumentInput, string,
) (*application.AccessoryDocument, error) {
	return &application.AccessoryDocument{}, nil
}

func (s *accessoryDocumentRepositorySpy) DeleteDocument(_ context.Context, id, _ string) (string, error) {
	s.deletedID = id
	return s.blobID, nil
}

type accessoryBlobCleanerSpy struct {
	repository         *accessoryDocumentRepositorySpy
	deletedBlobID      string
	metadataWasDeleted bool
}

func (s *accessoryBlobCleanerSpy) DeleteIfUnreferenced(_ context.Context, id string) error {
	s.deletedBlobID = id
	s.metadataWasDeleted = s.repository.deletedID != ""
	return nil
}

func withDocumentSize(input application.AccessoryDocumentUploadMetadata, value int64) application.AccessoryDocumentUploadMetadata {
	input.SizeBytes = value
	return input
}

func withDocumentFileName(input application.AccessoryDocumentUploadMetadata, value string) application.AccessoryDocumentUploadMetadata {
	input.FileName = value
	return input
}

func withDocumentOriginalName(input application.AccessoryDocumentUploadMetadata, value string) application.AccessoryDocumentUploadMetadata {
	input.OriginalName = value
	return input
}

func withDocumentMime(input application.AccessoryDocumentUploadMetadata, value string) application.AccessoryDocumentUploadMetadata {
	input.MimeType = value
	return input
}

func withDocumentCategory(input application.AccessoryDocumentUploadMetadata, value application.AccessoryDocumentCategory) application.AccessoryDocumentUploadMetadata {
	input.Category = value
	return input
}

func withDocumentPrimary(input application.AccessoryDocumentUploadMetadata, value bool) application.AccessoryDocumentUploadMetadata {
	input.IsPrimary = value
	return input
}
