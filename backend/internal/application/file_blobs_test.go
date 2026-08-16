package application_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestMigrateFilesystemBlobsRetainsFilesForPreMigrationSafetyCopy(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()
	vehicleService := application.NewVehicleService(db)
	vehicle, err := vehicleService.Create(ctx, application.CreateVehicleInput{
		Manufacturer: "Piko", Name: "BR 118", Gauge: "TT", Category: "Lokomotive", Gattung: "Diesellok",
	}, "actor-1")
	if err != nil {
		t.Fatal(err)
	}
	relativePath := "uploads/vehicles/test/manual.pdf"
	attachment, err := vehicleService.CreateAttachment(ctx, vehicle.ID, application.VehicleAttachmentInput{
		FileName: "manual.pdf", OriginalName: "manual.pdf", MimeType: "application/pdf",
		SizeBytes: 6, StoragePath: relativePath,
	})
	if err != nil {
		t.Fatal(err)
	}
	dataDir := t.TempDir()
	fullPath := filepath.Join(dataDir, filepath.FromSlash(relativePath))
	if err = os.MkdirAll(filepath.Dir(fullPath), 0o700); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(fullPath, []byte("manual"), 0o600); err != nil {
		t.Fatal(err)
	}

	blobs := application.NewFileBlobService(db, dataDir)
	if err = blobs.MigrateFilesystemBlobs(ctx); err != nil {
		t.Fatal(err)
	}
	if data, readErr := os.ReadFile(fullPath); readErr != nil || string(data) != "manual" {
		t.Fatalf("filesystem payload needed by safety copy was removed: data=%q err=%v", data, readErr)
	}
	var blobID string
	if err = db.QueryRowContext(ctx,
		`SELECT blob_id FROM vehicle_attachments WHERE id=?`, attachment.ID,
	).Scan(&blobID); err != nil {
		t.Fatal(err)
	}
	if blobID == "" {
		t.Fatal("active database did not receive migrated blob")
	}
}

func TestFileBlobDeleteRetainsAccessoryDocumentReferences(t *testing.T) {
	db := testDB(t)
	ctx := t.Context()
	if _, err := db.ExecContext(ctx, `INSERT INTO accessory_products(
  id, inventory_number, manufacturer, name, category, tracking_mode, created_at, updated_at
) VALUES('product-1', 'RK-ART-BLOB', 'Tillig', 'Gleis', 'track', 'quantity', 'now', 'now')`); err != nil {
		t.Fatal(err)
	}
	blobs := application.NewFileBlobService(db, "")
	blobID, err := blobs.Store(ctx, []byte("manual"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO accessory_documents(
  id, product_id, file_blob_id, category, file_name, original_name, mime_type, size_bytes,
  created_by, created_at, updated_at
) VALUES('document-1', 'product-1', ?, 'manual', 'manual.pdf', 'manual.pdf',
  'application/pdf', 6, 'editor-1', 'now', 'now')`, blobID); err != nil {
		t.Fatal(err)
	}
	if err := blobs.DeleteIfUnreferenced(ctx, blobID); err != nil {
		t.Fatal(err)
	}
	if data, err := blobs.Load(ctx, blobID); err != nil || string(data) != "manual" {
		t.Fatalf("referenced blob was removed: data=%q err=%v", data, err)
	}
	if _, err := db.ExecContext(ctx, `DELETE FROM accessory_documents WHERE id='document-1'`); err != nil {
		t.Fatal(err)
	}
	if err := blobs.DeleteIfUnreferenced(ctx, blobID); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM file_blobs WHERE id=?`, blobID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("unreferenced blob remains: %d", count)
	}
}
