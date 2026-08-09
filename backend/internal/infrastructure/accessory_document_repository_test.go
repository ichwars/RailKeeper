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
