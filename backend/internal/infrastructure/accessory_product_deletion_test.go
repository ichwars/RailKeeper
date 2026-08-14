package infrastructure_test

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/domain"
	"railkeeper/backend/internal/infrastructure"
)

func TestAccessoryProductDeleteRemovesUnusedProductMetadataAndAudits(t *testing.T) {
	_, db := testAccessoryService(t)
	blobs := application.NewFileBlobService(db, t.TempDir())
	blobID, err := blobs.Store(t.Context(), []byte("accessory image"))
	if err != nil {
		t.Fatal(err)
	}
	service := application.NewAccessoryService(infrastructure.NewAccessoryRepository(db), blobs)
	product := createDeletableAccessoryProduct(t, service)
	insertDeletableProductMetadata(t, db, product.ID, blobID)

	if err := service.DeleteProduct(t.Context(), product.ID, "admin-1"); err != nil {
		t.Fatal(err)
	}
	assertAccessoryDeleteRowCount(t, db, "accessory_products", "id", product.ID, 0)
	assertAccessoryDeleteRowCount(t, db, "accessory_stock", "product_id", product.ID, 0)
	assertAccessoryDeleteRowCount(t, db, "accessory_product_attributes", "product_id", product.ID, 0)
	assertAccessoryDeleteRowCount(t, db, "accessory_documents", "product_id", product.ID, 0)
	assertAccessoryDeleteRowCount(t, db, "file_blobs", "id", blobID, 0)
	assertAccessoryAuditCount(t, db, "AccessoryProductDeleted", 1)
}

func TestAccessoryProductDeleteRejectsUsageReferencesWithoutChangingData(t *testing.T) {
	tests := []struct {
		name   string
		insert func(*testing.T, *sql.DB, string)
	}{
		{name: "positive stock", insert: insertPositiveAccessoryStock},
		{name: "asset", insert: insertAccessoryAssetReference},
		{name: "stock movement", insert: insertAccessoryMovementReference},
		{name: "purchase", insert: insertAccessoryPurchaseReference},
		{name: "reservation", insert: insertAccessoryReservationReference},
		{name: "installation", insert: insertAccessoryInstallationReference},
		{name: "layout technical position", insert: insertLayoutTechnicalPositionReference},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, db := testAccessoryService(t)
			blobs := application.NewFileBlobService(db, t.TempDir())
			blobID, err := blobs.Store(t.Context(), []byte("blocked accessory image"))
			if err != nil {
				t.Fatal(err)
			}
			service := application.NewAccessoryService(infrastructure.NewAccessoryRepository(db), blobs)
			product := createDeletableAccessoryProduct(t, service)
			insertDeletableProductMetadata(t, db, product.ID, blobID)
			insertAccessoryDeleteReferencePrerequisites(t, db)
			test.insert(t, db, product.ID)

			if err := service.DeleteProduct(t.Context(), product.ID, "admin-1"); !errors.Is(err, application.ErrAccessoryDeleteBlocked) {
				t.Fatalf("delete error = %v", err)
			}
			assertAccessoryDeleteRowCount(t, db, "accessory_products", "id", product.ID, 1)
			assertAccessoryDeleteRowCount(t, db, "accessory_documents", "product_id", product.ID, 1)
			assertAccessoryDeleteRowCount(t, db, "file_blobs", "id", blobID, 1)
			assertAccessoryAuditCount(t, db, "AccessoryProductDeleted", 0)
		})
	}
}

func TestAccessoryProductDeleteReturnsNotFound(t *testing.T) {
	service, _ := testAccessoryService(t)
	if err := service.DeleteProduct(t.Context(), "missing", "admin-1"); !errors.Is(err, application.ErrAccessoryNotFound) {
		t.Fatalf("delete missing product error = %v", err)
	}
}

func createDeletableAccessoryProduct(
	t *testing.T,
	service *application.AccessoryService,
) *application.AccessoryProduct {
	t.Helper()
	product, err := service.CreateProduct(t.Context(), application.CreateAccessoryProductInput{
		Manufacturer: "Tillig", ArticleNumber: "DELETE-1", Name: "Löschtest",
		Category: "Gleismaterial", TrackingMode: domain.AccessoryTrackingModeQuantity,
	}, "admin-1")
	if err != nil {
		t.Fatal(err)
	}
	return product
}

func insertDeletableProductMetadata(t *testing.T, db *sql.DB, productID, blobID string) {
	t.Helper()
	const timestamp = "2026-08-14T10:00:00Z"
	statements := []struct {
		query string
		args  []any
	}{
		{`INSERT INTO storage_locations(id, name, description, archived, created_at, updated_at)
VALUES('location-delete', 'Löschlager', '', 0, ?, ?)`, []any{timestamp, timestamp}},
		{`INSERT INTO accessory_stock(product_id, location_id, quantity, updated_at)
VALUES(?, 'location-delete', 0, ?)`, []any{productID, timestamp}},
		{`INSERT INTO accessory_product_attributes(
  id, product_id, attribute_key, value_type, text_value, created_at, updated_at
) VALUES('attribute-delete', ?, 'test-note', 'text', 'Metadatum', ?, ?)`,
			[]any{productID, timestamp, timestamp}},
		{`INSERT INTO accessory_documents(
  id, product_id, file_blob_id, file_name, original_name, description, category,
  mime_type, size_bytes, is_primary, created_by, created_at, updated_at
) VALUES('document-delete', ?, ?, 'bild.png', 'bild.png', '', 'image',
  'image/png', 15, 1, 'admin-1', ?, ?)`, []any{productID, blobID, timestamp, timestamp}},
	}
	for _, statement := range statements {
		mustExecuteAccessoryDeleteFixture(t, db, statement.query, statement.args...)
	}
}

func insertAccessoryDeleteReferencePrerequisites(t *testing.T, db *sql.DB) {
	t.Helper()
	for _, statement := range []string{
		`INSERT INTO storage_locations(id, name, description, archived, created_at, updated_at)
VALUES('location-block', 'Sperrlager', '', 0, '2026-08-14T10:00:00Z', '2026-08-14T10:00:00Z')`,
		`INSERT INTO layouts(id, name, kind, gauge, scale, description, version, archived, created_at, updated_at)
VALUES('layout-block', 'Sperranlage', 'club', 'TT', '1:120', '', 1, 0,
  '2026-08-14T10:00:00Z', '2026-08-14T10:00:00Z')`,
		`INSERT INTO layout_units(
  id, layout_id, name, kind, owner_label, width_mm, height_mm, version, archived, created_at, updated_at
) VALUES('unit-block', 'layout-block', 'Sperrmodul', 'module', '', 1000, 500, 1, 0,
  '2026-08-14T10:00:00Z', '2026-08-14T10:00:00Z')`,
	} {
		mustExecuteAccessoryDeleteFixture(t, db, statement)
	}
}

func insertPositiveAccessoryStock(t *testing.T, db *sql.DB, productID string) {
	mustExecuteAccessoryDeleteFixture(t, db, `
INSERT INTO accessory_stock(product_id, location_id, quantity, updated_at)
VALUES(?, 'location-block', 1, '2026-08-14T10:00:00Z')`, productID)
}

func insertAccessoryAssetReference(t *testing.T, db *sql.DB, productID string) {
	mustExecuteAccessoryDeleteFixture(t, db, `
INSERT INTO accessory_assets(id, product_id, condition_state, lifecycle_state, created_at, updated_at)
VALUES('asset-block', ?, 'ready', 'stored', '2026-08-14T10:00:00Z', '2026-08-14T10:00:00Z')`, productID)
}

func insertAccessoryMovementReference(t *testing.T, db *sql.DB, productID string) {
	mustExecuteAccessoryDeleteFixture(t, db, `
INSERT INTO accessory_stock_movements(id, product_id, location_id, movement_type, quantity, created_at)
VALUES('movement-block', ?, 'location-block', 'adjustment', 1, '2026-08-14T10:00:00Z')`, productID)
}

func insertAccessoryPurchaseReference(t *testing.T, db *sql.DB, productID string) {
	mustExecuteAccessoryDeleteFixture(t, db, `
INSERT INTO accessory_purchases(id, product_id, quantity, purchased_at, created_at, updated_at)
VALUES('purchase-block', ?, 1, '2026-08-14', '2026-08-14T10:00:00Z', '2026-08-14T10:00:00Z')`, productID)
}

func insertAccessoryReservationReference(t *testing.T, db *sql.DB, productID string) {
	mustExecuteAccessoryDeleteFixture(t, db, `
INSERT INTO accessory_reservations(
  id, product_id, location_id, quantity, layout_id, status, created_by, created_at, updated_at
) VALUES('reservation-block', ?, 'location-block', 1, 'layout-block', 'cancelled', 'admin-1',
  '2026-08-14T10:00:00Z', '2026-08-14T10:00:00Z')`, productID)
}

func insertAccessoryInstallationReference(t *testing.T, db *sql.DB, productID string) {
	mustExecuteAccessoryDeleteFixture(t, db, `
INSERT INTO accessory_installations(
  id, product_id, source_location_id, quantity, layout_id, condition_state,
  installed_by, installed_at, removed_by, removed_at, removal_disposition, notes, removal_notes
) VALUES('installation-block', ?, 'location-block', 1, 'layout-block', 'ready',
  'admin-1', '2026-08-14T10:00:00Z', 'admin-1', '2026-08-14T11:00:00Z', 'stored', '', '')`, productID)
}

func insertLayoutTechnicalPositionReference(t *testing.T, db *sql.DB, productID string) {
	mustExecuteAccessoryDeleteFixture(t, db, `
INSERT INTO layout_technical_positions(
  id, layout_unit_id, label, kind, position_x_mm, position_y_mm, rotation_degrees,
  product_id, description, version, archived, created_at, updated_at
) VALUES('position-block', 'unit-block', 'Sperrposition', 'turnout', 10, 20, 0,
  ?, '', 1, 0, '2026-08-14T10:00:00Z', '2026-08-14T10:00:00Z')`, productID)
}

func mustExecuteAccessoryDeleteFixture(t *testing.T, db *sql.DB, query string, args ...any) {
	t.Helper()
	if _, err := db.ExecContext(t.Context(), query, args...); err != nil {
		t.Fatal(err)
	}
}

func assertAccessoryDeleteRowCount(t *testing.T, db *sql.DB, table, column, value string, want int) {
	t.Helper()
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s = ?", table, column)
	var got int
	if err := db.QueryRowContext(t.Context(), query, value).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("%s row count = %d, want %d", table, got, want)
	}
}
