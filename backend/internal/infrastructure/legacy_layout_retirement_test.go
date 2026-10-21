package infrastructure_test

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/infrastructure"
	"railkeeper/backend/migrations"
)

func TestRetirementPreservesStockVehiclesOriginalRecordsAndReadAccess(t *testing.T) {
	root := t.TempDir()
	partial := filepath.Join(root, "before-retirement")
	if err := os.Mkdir(partial, 0700); err != nil {
		t.Fatal(err)
	}
	copyMigrationsThrough(t, filepath.Join("..", "..", "migrations"), partial,
		"0068_vehicle_cv_provenance.sql")
	db, err := infrastructure.OpenSQLite(filepath.Join(root, "data"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := infrastructure.Migrate(db, partial); err != nil {
		t.Fatal(err)
	}
	seedLegacyRetirement(t, db)
	if err := infrastructure.Migrate(db, filepath.Join("..", "..", "migrations")); err != nil {
		t.Fatal(err)
	}
	assertRetiredInventory(t, db)
	var original string
	if err := db.QueryRow(`SELECT row_json FROM retired_allocation_records
WHERE source_table='accessory_reservations' AND source_id='old-reservation'`).Scan(&original); err != nil {
		t.Fatal(err)
	}
	var row map[string]any
	if err := json.Unmarshal([]byte(original), &row); err != nil {
		t.Fatal(err)
	}
	if row["status"] != "active" || row["layout_id"] != "old-layout" || row["note"] != "Original note" {
		t.Fatalf("lost original reservation: %s", original)
	}
	assertRetirementInt(t, db, `SELECT COUNT(*) FROM user_roles ur JOIN roles r ON r.id=ur.role_id
WHERE ur.user_id='former-planner' AND r.name='Viewer'`, 1)
	assertRetirementInt(t, db, `SELECT COUNT(*) FROM roles WHERE name='Planner'`, 0)
	if _, err := db.Exec(migrations.LegacyLayoutRetirementSQL); err != nil {
		t.Fatal(err)
	}
	assertRetiredInventory(t, db)
}

func TestLegacyBackupRestoreRetiresAllocationsOnceAndKeepsArchive(t *testing.T) {
	dataDir := t.TempDir()
	db, err := infrastructure.OpenSQLite(dataDir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := infrastructure.Migrate(db, filepath.Join("..", "..", "migrations")); err != nil {
		t.Fatal(err)
	}
	seedLegacyRetirement(t, db)
	service := application.NewBackupService(db, dataDir)
	legacy, err := service.Export(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	legacy.Version = 22
	delete(legacy.Tables, "retired_allocation_records")
	validation, err := service.Validate(t.Context(), legacy)
	if err != nil || !validation.Compatible || len(validation.Warnings) == 0 {
		t.Fatalf("legacy preflight: %#v, %v", validation, err)
	}
	if _, err := service.Import(t.Context(), legacy); err != nil {
		t.Fatal(err)
	}
	assertRetiredInventory(t, db)
	current, err := service.Export(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if current.Version != 24 || len(current.Tables["retired_allocation_records"]) != 4 {
		t.Fatalf("archive absent from current backup: version %d", current.Version)
	}
	for range 2 {
		if _, err := service.Import(t.Context(), current); err != nil {
			t.Fatal(err)
		}
		assertRetiredInventory(t, db)
	}
	// Restore must not alter local authentication, including unrelated old role assignments.
	assertRetirementInt(t, db, `SELECT COUNT(*) FROM users WHERE id='former-planner'`, 1)
	assertRetirementInt(t, db, `SELECT COUNT(*) FROM user_roles ur JOIN roles r ON r.id=ur.role_id
WHERE ur.user_id='former-planner' AND r.name='Planner'`, 1)
	// A failure after data conversion must roll back stock, archive rows and local auth together.
	current.Tables["accessory_installations"][0]["source_location_id"] = "missing-source"
	if _, err := service.Import(t.Context(), current); err == nil {
		t.Fatal("invalid restore unexpectedly committed")
	}
	assertRetiredInventory(t, db)
}

func seedLegacyRetirement(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.Exec(`
INSERT INTO users(id,username,password_hash,created_at) VALUES('former-planner','historical-reader','unused','now');
INSERT OR IGNORE INTO roles(id,name) VALUES('historical-planner','Planner');
INSERT INTO user_roles(user_id,role_id) SELECT 'former-planner',id FROM roles WHERE name='Planner';
INSERT INTO storage_locations(id,name,created_at,updated_at) VALUES('source','Source','now','now');
INSERT INTO layouts(id,name,kind,gauge,scale,created_at,updated_at) VALUES('old-layout','Historical plan','private','TT','1:120','now','now');
INSERT INTO vehicles(id,inventory_number,manufacturer,name,gauge,created_at,updated_at)
VALUES('vehicle','RK-RETIRE-1','ESU','Vehicle','TT','now','now');
INSERT INTO accessory_products(id,inventory_number,manufacturer,name,category,tracking_mode,inventory_strategy,list_price,created_at,updated_at)
VALUES('counted','RK-RETIRE-COUNTED','Tillig','Track','track','quantity','quantity','10.00','now','now'),
('individual','RK-RETIRE-ITEM','ESU','Decoder','decoder','individual','individual','20.00','now','now');
INSERT INTO accessory_stock(product_id,location_id,quantity,updated_at) VALUES('counted','source',7,'now');
INSERT INTO accessory_assets(id,product_id,condition_state,lifecycle_state,storage_location_id,created_at,updated_at)
VALUES('reserved-item','individual','ready','reserved','source','now','now'),
('installed-item','individual','maintenance_due','installed',NULL,'now','now'),
('vehicle-item','individual','ready','installed',NULL,'now','now');
INSERT INTO accessory_reservations(id,product_id,asset_id,location_id,quantity,vehicle_id,layout_id,status,note,created_by,created_at,updated_at)
VALUES('old-reservation','counted',NULL,'source',4,NULL,'old-layout','active','Original note','historical','now','now'),
('old-asset-reservation','individual','reserved-item','source',1,NULL,'old-layout','active','','historical','now','now'),
('vehicle-reservation','counted',NULL,'source',1,'vehicle',NULL,'active','','editor','now','now');
INSERT INTO accessory_installations(id,product_id,asset_id,source_location_id,quantity,vehicle_id,layout_id,condition_state,installed_by,installed_at)
VALUES('old-installation','counted',NULL,'source',3,NULL,'old-layout','ready','historical','now'),
('old-asset-installation','individual','installed-item','source',1,NULL,'old-layout','maintenance_due','historical','now'),
('vehicle-installation','counted',NULL,'source',2,'vehicle',NULL,'ready','editor','now'),
('vehicle-asset-installation','individual','vehicle-item','source',1,'vehicle',NULL,'ready','editor','now');
`)
	if err != nil {
		t.Fatal(err)
	}
}

func assertRetiredInventory(t *testing.T, db *sql.DB) {
	t.Helper()
	assertRetirementInt(t, db, `SELECT quantity FROM accessory_stock WHERE product_id='counted' AND location_id='source'`, 10)
	assertRetirementInt(t, db, `SELECT COUNT(*) FROM retired_allocation_records`, 4)
	assertRetirementInt(t, db, `SELECT COUNT(*) FROM accessory_reservations WHERE vehicle_id IS NULL AND status='active'`, 0)
	assertRetirementInt(t, db, `SELECT COUNT(*) FROM accessory_installations WHERE vehicle_id IS NULL AND removed_at IS NULL`, 0)
	assertRetirementInt(t, db, `SELECT COUNT(*) FROM accessory_reservations WHERE id='vehicle-reservation' AND status='active'`, 1)
	assertRetirementInt(t, db, `SELECT COUNT(*) FROM accessory_installations WHERE vehicle_id='vehicle' AND removed_at IS NULL`, 2)
	assertRetirementInt(t, db, `SELECT COUNT(*) FROM accessory_assets WHERE id='reserved-item' AND lifecycle_state='stored' AND storage_location_id='source'`, 1)
	assertRetirementInt(t, db, `SELECT COUNT(*) FROM accessory_assets WHERE id='installed-item' AND lifecycle_state='maintenance' AND storage_location_id='source'`, 1)
	assertRetirementInt(t, db, `SELECT COUNT(*) FROM accessory_assets WHERE id='vehicle-item' AND lifecycle_state='installed' AND storage_location_id IS NULL`, 1)
	assertRetirementInt(t, db, `SELECT COUNT(*) FROM accessory_stock_movements WHERE source_type='retired_layout'`, 1)
	var foreignKeyErrors int
	if err := db.QueryRow(`SELECT COUNT(*) FROM pragma_foreign_key_check`).Scan(&foreignKeyErrors); err != nil || foreignKeyErrors != 0 {
		t.Fatalf("foreign key check: %d, %v", foreignKeyErrors, err)
	}
	allocations := application.NewAccessoryAllocationService(infrastructure.NewAccessoryRepository(db))
	summary, err := allocations.GetAllocationSummary(t.Context(), "counted")
	if err != nil || summary.Owned != 12 || summary.Reserved != 1 || summary.Installed != 2 || summary.Available != 9 {
		t.Fatalf("inventory conservation: %#v, %v", summary, err)
	}
	valuation, err := application.NewOverviewValuationService(db).Get(t.Context())
	if err != nil || valuation.AccessoryListValue != "180.00" {
		t.Fatalf("valuation conservation: %#v, %v", valuation, err)
	}
	reservations, err := allocations.ListReservations(t.Context(), "counted")
	if err != nil || len(reservations) != 1 || reservations[0].VehicleID != "vehicle" {
		t.Fatalf("vehicle reservations: %#v, %v", reservations, err)
	}
	installations, err := allocations.ListInstallations(t.Context(), "counted")
	if err != nil || len(installations) != 1 || installations[0].VehicleID != "vehicle" {
		t.Fatalf("vehicle installations: %#v, %v", installations, err)
	}
}

func assertRetirementInt(t *testing.T, db *sql.DB, query string, want int) {
	t.Helper()
	var got int
	if err := db.QueryRow(query).Scan(&got); err != nil || got != want {
		t.Fatalf("%s: got %d, want %d, err %v", query, got, want, err)
	}
}
