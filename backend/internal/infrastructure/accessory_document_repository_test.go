package infrastructure_test

import (
	"errors"
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/infrastructure"
)

func TestAccessoryDocumentsManagePrimaryImageAndBlobLifecycle(t *testing.T) {
	fixture := newAllocationFixture(t)
	ctx := t.Context()
	repository := infrastructure.NewAccessoryRepository(fixture.db)
	blobs := application.NewFileBlobService(fixture.db, "")
	documents := application.NewAccessoryDocumentService(repository, blobs)

	firstBlobID, err := blobs.Store(ctx, []byte("first image"))
	if err != nil {
		t.Fatal(err)
	}
	first, err := documents.CreateDocument(ctx, application.CreateAccessoryDocumentInput{
		ProductID: fixture.quantityProduct.ID, FileBlobID: firstBlobID,
		AccessoryDocumentUploadMetadata: application.AccessoryDocumentUploadMetadata{
			FileName: "first.png", OriginalName: "Erstes Bild.png",
			Category: application.AccessoryDocumentImage, MimeType: "image/png", SizeBytes: 11,
			IsPrimary: true,
		},
		Description: "Frontansicht",
	}, 1024, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	secondBlobID, err := blobs.Store(ctx, []byte("second image"))
	if err != nil {
		t.Fatal(err)
	}
	second, err := documents.CreateDocument(ctx, application.CreateAccessoryDocumentInput{
		ProductID: fixture.quantityProduct.ID, FileBlobID: secondBlobID,
		AccessoryDocumentUploadMetadata: application.AccessoryDocumentUploadMetadata{
			FileName: "second.webp", OriginalName: "Zweites Bild.webp",
			Category: application.AccessoryDocumentImage, MimeType: "image/webp", SizeBytes: 12,
			IsPrimary: true,
		},
	}, 1024, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	listed, err := documents.ListDocuments(ctx, fixture.quantityProduct.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 2 || !listed[0].IsPrimary || listed[0].ID != second.ID || listed[1].IsPrimary {
		t.Fatalf("primary image was not switched atomically: %#v", listed)
	}
	product, err := fixture.accessories.GetProduct(ctx, fixture.quantityProduct.ID)
	if err != nil {
		t.Fatal(err)
	}
	wantURL := "/api/v1/accessory-products/" + fixture.quantityProduct.ID + "/documents/" + second.ID + "/download"
	if product.PrimaryImageURL != wantURL {
		t.Fatalf("primary image URL: got %q, want %q", product.PrimaryImageURL, wantURL)
	}

	if err := documents.DeleteDocument(ctx, first.ID, "editor-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := blobs.Load(ctx, firstBlobID); err == nil {
		t.Fatal("unreferenced document blob was retained")
	}
	if data, err := blobs.Load(ctx, secondBlobID); err != nil || string(data) != "second image" {
		t.Fatalf("referenced primary blob was removed: data=%q err=%v", data, err)
	}
}

func TestAccessoryDocumentRepositoryRequiresExistingProductAndBlob(t *testing.T) {
	fixture := newAllocationFixture(t)
	repository := infrastructure.NewAccessoryRepository(fixture.db)
	documents := application.NewAccessoryDocumentService(repository, application.NewFileBlobService(fixture.db, ""))
	input := application.CreateAccessoryDocumentInput{
		ProductID: fixture.quantityProduct.ID, FileBlobID: "missing-blob",
		AccessoryDocumentUploadMetadata: application.AccessoryDocumentUploadMetadata{
			FileName: "manual.pdf", OriginalName: "manual.pdf", Category: application.AccessoryDocumentManual,
			MimeType: "application/pdf", SizeBytes: 10,
		},
	}
	if _, err := documents.CreateDocument(t.Context(), input, 1024, "editor-1"); !errors.Is(err, application.ErrAccessoryNotFound) {
		t.Fatalf("expected missing blob rejection, got %v", err)
	}
}

func TestAccessoryDocumentImportIsIdempotentAndOnlyFillsMissingPrimary(t *testing.T) {
	fixture := newAllocationFixture(t)
	ctx := t.Context()
	repository := infrastructure.NewAccessoryRepository(fixture.db)
	blobs := application.NewFileBlobService(fixture.db, "")
	documents := application.NewAccessoryDocumentService(repository, blobs)

	primaryBlobID, err := blobs.Store(ctx, []byte("existing primary"))
	if err != nil {
		t.Fatal(err)
	}
	primary, err := documents.CreateDocument(ctx, application.CreateAccessoryDocumentInput{
		ProductID: fixture.quantityProduct.ID, FileBlobID: primaryBlobID,
		AccessoryDocumentUploadMetadata: application.AccessoryDocumentUploadMetadata{
			FileName: "primary.png", OriginalName: "primary.png",
			Category: application.AccessoryDocumentImage, MimeType: "image/png", SizeBytes: 16,
			IsPrimary: true,
		},
	}, 1024, "editor-1")
	if err != nil {
		t.Fatal(err)
	}

	importBlobID, err := blobs.Store(ctx, []byte("remote image"))
	if err != nil {
		t.Fatal(err)
	}
	input := application.CreateAccessoryDocumentInput{
		DocumentID: "remote-import-stable-id", ProductID: fixture.quantityProduct.ID, FileBlobID: importBlobID,
		PrimaryIfMissing: true,
		AccessoryDocumentUploadMetadata: application.AccessoryDocumentUploadMetadata{
			FileName: "remote.jpg", OriginalName: "remote.jpg",
			Category: application.AccessoryDocumentImage, MimeType: "image/jpeg", SizeBytes: 12,
		},
	}
	imported, err := documents.CreateDocument(ctx, input, 1024, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	if imported.IsPrimary {
		t.Fatal("remote import replaced an existing primary image")
	}

	retried, err := documents.CreateDocument(ctx, input, 1024, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	if retried.ID != imported.ID {
		t.Fatalf("retry created another document: first=%q retry=%q", imported.ID, retried.ID)
	}
	listed, err := documents.ListDocuments(ctx, fixture.quantityProduct.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 2 || listed[0].ID != primary.ID || !listed[0].IsPrimary {
		t.Fatalf("unexpected documents after retry: %#v", listed)
	}

	firstBlobID, err := blobs.Store(ctx, []byte("first remote"))
	if err != nil {
		t.Fatal(err)
	}
	first, err := documents.CreateDocument(ctx, application.CreateAccessoryDocumentInput{
		DocumentID: "remote-import-first-id", ProductID: fixture.individualProduct.ID, FileBlobID: firstBlobID,
		PrimaryIfMissing: true,
		AccessoryDocumentUploadMetadata: application.AccessoryDocumentUploadMetadata{
			FileName: "first.jpg", OriginalName: "first.jpg",
			Category: application.AccessoryDocumentImage, MimeType: "image/jpeg", SizeBytes: 12,
		},
	}, 1024, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	if !first.IsPrimary {
		t.Fatal("first imported image was not assigned as primary")
	}
}
