package application_test

import (
	"testing"

	"railkeeper/backend/internal/application"
)

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
