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

func TestAccessoryDocumentUploadRejectsUnsafeBasenamesForBothFields(t *testing.T) {
	valid := application.AccessoryDocumentUploadMetadata{
		FileName: "manual.pdf", OriginalName: "Anleitung.pdf",
		Category: application.AccessoryDocumentManual, MimeType: "application/pdf", SizeBytes: 8,
	}
	invalidNames := []struct {
		name  string
		value string
	}{
		{"empty", ""},
		{"dot", "."},
		{"dot dot", ".."},
		{"forward slash", "folder/manual.pdf"},
		{"backslash", `folder\manual.pdf`},
		{"control character", "manual\n.pdf"},
		{"less than", "manual<copy>.pdf"},
		{"greater than", "manual>copy.pdf"},
		{"colon", "manual:copy.pdf"},
		{"quote", `manual"copy.pdf`},
		{"pipe", "manual|copy.pdf"},
		{"question mark", "manual?.pdf"},
		{"asterisk", "manual*.pdf"},
		{"trailing dot", "manual.pdf."},
		{"trailing space", "manual.pdf "},
		{"reserved con", "CON.pdf"},
		{"reserved prn case insensitive", "prn.PDF"},
		{"reserved aux", "AUX.pdf"},
		{"reserved nul", "NUL.pdf"},
		{"reserved com lower bound", "COM1.pdf"},
		{"reserved com upper bound", "com9.pdf"},
		{"reserved lpt lower bound", "LPT1.pdf"},
		{"reserved lpt upper bound", "lpt9.pdf"},
	}
	fields := []struct {
		name string
		set  func(application.AccessoryDocumentUploadMetadata, string) application.AccessoryDocumentUploadMetadata
	}{
		{"file name", withDocumentFileName},
		{"original name", withDocumentOriginalName},
	}
	for _, field := range fields {
		for _, invalid := range invalidNames {
			t.Run(field.name+"/"+invalid.name, func(t *testing.T) {
				metadata := field.set(valid, invalid.value)
				if err := application.ValidateAccessoryDocumentUpload(metadata, 8); !errors.Is(err, application.ErrAccessoryValidation) {
					t.Fatalf("unsafe basename %q was accepted", invalid.value)
				}
			})
		}
	}

	unicodeNames := valid
	unicodeNames.FileName = "Pläne-äöü.pdf"
	unicodeNames.OriginalName = "Änderungsplan 東京.pdf"
	if err := application.ValidateAccessoryDocumentUpload(unicodeNames, 8); err != nil {
		t.Fatalf("safe Unicode basenames rejected: %v", err)
	}
}

func TestAccessoryDocumentUploadUsesExtensionSpecificMIMETypes(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		mimeType string
		valid    bool
	}{
		{"txt", "notes.txt", "text/plain", true},
		{"csv text", "parts.csv", "text/csv", true},
		{"csv application", "parts.csv", "application/csv", true},
		{"json", "decoder.json", "application/json", true},
		{"xml application", "decoder.xml", "application/xml", true},
		{"xml text", "decoder.xml", "text/xml", true},
		{"csv with json", "parts.csv", "application/json", false},
		{"json with text plain", "decoder.json", "text/plain", false},
		{"txt with html", "notes.txt", "text/html", false},
		{"xml with json", "decoder.xml", "application/json", false},
		{"csv with html", "parts.csv", "text/html", false},
		{"json with html", "decoder.json", "text/html", false},
		{"xml with html", "decoder.xml", "text/html", false},
		{"txt with script", "notes.txt", "application/javascript", false},
		{"csv with script", "parts.csv", "text/javascript", false},
		{"json with script", "decoder.json", "application/ecmascript", false},
		{"xml with script", "decoder.xml", "application/javascript", false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			metadata := application.AccessoryDocumentUploadMetadata{
				FileName: test.fileName, OriginalName: test.fileName,
				Category: application.AccessoryDocumentOther, MimeType: test.mimeType, SizeBytes: 8,
			}
			err := application.ValidateAccessoryDocumentUpload(metadata, 8)
			if test.valid && err != nil {
				t.Fatalf("valid extension/MIME pair rejected: %v", err)
			}
			if !test.valid && !errors.Is(err, application.ErrAccessoryValidation) {
				t.Fatalf("invalid extension/MIME pair accepted: %s + %s", test.fileName, test.mimeType)
			}
		})
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
