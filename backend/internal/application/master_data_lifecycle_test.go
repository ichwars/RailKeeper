package application_test

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestMasterDataOriginAndNormalLists(t *testing.T) {
	db := testDB(t)
	service := application.NewMasterDataService(db)
	created, err := service.Create(t.Context(), "manufacturer", application.MasterDataInput{
		Key: "club", Label: "Club",
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.Origin != application.MasterDataOriginCustom {
		t.Fatalf("created origin=%q", created.Origin)
	}
	if created.Capabilities != nil {
		t.Fatalf("normal get exposed management capabilities: %#v", created.Capabilities)
	}

	insertLifecycleEntry(t, db, "gauge", "club-gauge", "Club Gauge", "bundled", true)
	got, err := service.Get(t.Context(), "gauge", "club-gauge")
	if err != nil {
		t.Fatal(err)
	}
	if got.Origin != application.MasterDataOriginBundled {
		t.Fatalf("bundled origin=%q", got.Origin)
	}
	listed, err := service.List(t.Context(), "gauge", false)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range listed {
		if entry.Capabilities != nil {
			t.Fatalf("normal list exposed management capabilities for %s", entry.Key)
		}
	}
}

func TestMasterDataDeleteRejectsBundledEntry(t *testing.T) {
	db := testDB(t)
	insertLifecycleEntry(t, db, "manufacturer", "bundled-club", "Bundled Club", "bundled", true)
	service := application.NewMasterDataService(db)
	if err := service.Delete(t.Context(), "manufacturer", "bundled-club"); !errors.Is(
		err, application.ErrMasterDataBundled,
	) {
		t.Fatalf("delete error=%v", err)
	}
	if _, err := service.Get(t.Context(), "manufacturer", "bundled-club"); err != nil {
		t.Fatalf("bundled entry changed: %v", err)
	}
}

func TestMasterDataDeleteRejectsEveryKnownReference(t *testing.T) {
	tests := []struct {
		name     string
		typeName string
		key      string
		label    string
		use      func(*testing.T, *sql.DB, string, string)
	}{
		{"vehicle manufacturer", "manufacturer", "club", "Club", useVehicleColumn("manufacturer")},
		{"vehicle gauge", "gauge", "club", "Club", useVehicleColumn("gauge")},
		{"vehicle epoch", "epoch", "club", "Club", useVehicleColumn("epoch")},
		{"vehicle railway company", "railway_company", "club", "Club", useVehicleColumn("railway_company")},
		{"vehicle category", "vehicle_category", "club", "Club", useVehicleColumn("category")},
		{"vehicle gattung", "vehicle_gattung", "club", "Club", useVehicleColumn("gattung")},
		{"exhibition manufacturer", "manufacturer", "club", "Club", useExhibitionColumn("manufacturer")},
		{"exhibition epoch", "epoch", "club", "Club", useExhibitionColumn("epoch")},
		{"exhibition railway company", "railway_company", "club", "Club", useExhibitionColumn("railway_company")},
		{"exhibition gattung", "vehicle_gattung", "club", "Club", useExhibitionColumn("gattung")},
		{"accessory manufacturer", "manufacturer", "club", "Club", useAccessoryColumn("manufacturer")},
		{"accessory gauge json", "gauge", "club", "Club", useAccessoryColumn("gauges_json")},
		{"accessory stock unit", "stock_unit", "club", "Club", useAccessoryColumn("stock_unit")},
		{"accessory article type", "article_type", "club", "Club", useAccessoryColumn("article_type")},
		{"accessory subtype", "accessory_subtype", "club", "Club", useAccessoryColumn("subtype")},
		{"function symbol", "symbols", "club-light", "Club Light", useFunctionSymbol},
		{"accessory custom field", "accessory_custom_field", "club-field", "Club Field", useCustomField},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := testDB(t)
			insertLifecycleEntry(t, db, test.typeName, test.key, test.label, "custom", true)
			test.use(t, db, test.key, test.label)
			service := application.NewMasterDataService(db)
			if err := service.Delete(t.Context(), test.typeName, test.key); !errors.Is(
				err, application.ErrMasterDataInUse,
			) {
				t.Fatalf("delete error=%v", err)
			}
			if _, err := service.Get(t.Context(), test.typeName, test.key); err != nil {
				t.Fatalf("used entry changed: %v", err)
			}
		})
	}
}

func TestMasterDataDeleteAllowsKnownUnusedCustomAndCleansRelations(t *testing.T) {
	db := testDB(t)
	insertLifecycleEntry(t, db, "cv8_manufacturer", "club", "Club", "custom", true)
	insertLifecycleEntry(t, db, "vehicle_category", "parent", "Parent", "custom", true)
	insertLifecycleEntry(t, db, "vehicle_gattung", "child", "Child", "custom", true)
	if _, err := db.Exec(`
INSERT INTO master_data_relations(
  id, parent_type, parent_key, child_type, child_key, sort_order, created_at
) VALUES
  ('r1', 'vehicle_category', 'parent', 'vehicle_gattung', 'child', 0, 'now'),
  ('r2', 'vehicle_category', 'other', 'vehicle_gattung', 'child', 1, 'now')`); err != nil {
		t.Fatal(err)
	}

	service := application.NewMasterDataService(db)
	if err := service.Delete(t.Context(), "cv8_manufacturer", "club"); err != nil {
		t.Fatalf("delete known unused entry: %v", err)
	}
	if err := service.Delete(t.Context(), "vehicle_gattung", "child"); err != nil {
		t.Fatalf("delete related unused entry: %v", err)
	}
	var relations int
	if err := db.QueryRow(`SELECT COUNT(*) FROM master_data_relations WHERE child_key='child'`).Scan(&relations); err != nil {
		t.Fatal(err)
	}
	if relations != 0 {
		t.Fatalf("remaining relations=%d", relations)
	}
}

func TestMasterDataDeleteUnknownUsageFailsClosed(t *testing.T) {
	db := testDB(t)
	insertLifecycleEntry(t, db, "future_type", "club", "Club", "custom", true)
	service := application.NewMasterDataService(db)
	if err := service.Delete(t.Context(), "future_type", "club"); !errors.Is(
		err, application.ErrMasterDataUsageUnknown,
	) {
		t.Fatalf("delete error=%v", err)
	}
	if _, err := service.Get(t.Context(), "future_type", "club"); err != nil {
		t.Fatalf("unknown entry changed: %v", err)
	}
}

func TestMasterDataManagementCapabilities(t *testing.T) {
	db := testDB(t)
	insertLifecycleEntry(t, db, "manufacturer", "bundled", "Bundled", "bundled", true)
	insertLifecycleEntry(t, db, "manufacturer", "unused", "Unused", "custom", true)
	insertLifecycleEntry(t, db, "manufacturer", "inactive", "Inactive", "custom", false)
	insertLifecycleEntry(t, db, "manufacturer", "used", "Used", "custom", true)
	useVehicleColumn("manufacturer")(t, db, "used", "Used")

	service := application.NewMasterDataService(db)
	entries, err := service.ListForManagement(t.Context(), "manufacturer")
	if err != nil {
		t.Fatal(err)
	}
	byKey := make(map[string]application.MasterDataCapabilities, len(entries))
	for _, entry := range entries {
		if entry.Capabilities == nil {
			t.Fatalf("missing capabilities for %s", entry.Key)
		}
		byKey[entry.Key] = *entry.Capabilities
	}
	assertCapabilities(t, byKey["bundled"], true, false, false)
	assertCapabilities(t, byKey["unused"], true, false, true)
	assertCapabilities(t, byKey["inactive"], false, true, true)
	assertCapabilities(t, byKey["used"], true, false, false)

	all, err := service.ListAllForManagement(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(all["manufacturer"]) != len(entries) {
		t.Fatalf("all management manufacturers=%d, list=%d", len(all["manufacturer"]), len(entries))
	}
}

func TestMasterDataSetActiveInvalidatesCache(t *testing.T) {
	db := testDB(t)
	insertLifecycleEntry(t, db, "manufacturer", "club", "Club", "custom", true)
	service := application.NewMasterDataService(db)
	if _, err := service.ListAll(t.Context(), true); err != nil {
		t.Fatal(err)
	}

	updated, err := service.SetActive(t.Context(), "manufacturer", "club", false)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Active {
		t.Fatal("entry remained active")
	}
	active, err := service.ListAll(t.Context(), true)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range active["manufacturer"] {
		if entry.Key == "club" {
			t.Fatal("inactive entry remained in cached active list")
		}
	}
}

func TestMasterDataSetActiveManyNormalizesDeduplicatesAndInvalidatesCache(t *testing.T) {
	db := testDB(t)
	insertLifecycleEntry(t, db, "manufacturer", "first", "First", "custom", true)
	insertLifecycleEntry(t, db, "manufacturer", "second", "Second", "custom", true)
	service := application.NewMasterDataService(db)
	if _, err := service.ListAll(t.Context(), true); err != nil {
		t.Fatal(err)
	}

	updated, err := service.SetActiveMany(t.Context(), " manufacturer ",
		[]string{" first ", "second", "first"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(updated) != 2 || updated[0].Key != "first" || updated[1].Key != "second" {
		t.Fatalf("unexpected result order: %#v", updated)
	}
	for _, entry := range updated {
		if entry.Active {
			t.Fatalf("entry %q remained active", entry.Key)
		}
		if entry.Capabilities == nil || !entry.Capabilities.CanReactivate || entry.Capabilities.CanDeactivate {
			t.Fatalf("entry %q capabilities=%#v", entry.Key, entry.Capabilities)
		}
	}
	active, err := service.ListAll(t.Context(), true)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range active["manufacturer"] {
		if entry.Key == "first" || entry.Key == "second" {
			t.Fatalf("inactive entry %q remained in cached active list", entry.Key)
		}
	}
}

func TestMasterDataSetActiveManyRollsBackMissingEntry(t *testing.T) {
	db := testDB(t)
	insertLifecycleEntry(t, db, "manufacturer", "first", "First", "custom", true)
	service := application.NewMasterDataService(db)

	_, err := service.SetActiveMany(t.Context(), "manufacturer", []string{"first", "missing"}, false)
	if !errors.Is(err, application.ErrMasterDataNotFound) {
		t.Fatalf("error=%v", err)
	}
	entry, err := service.Get(t.Context(), "manufacturer", "first")
	if err != nil {
		t.Fatal(err)
	}
	if !entry.Active {
		t.Fatal("valid entry changed despite batch rollback")
	}
}

func TestMasterDataSetActiveManyValidatesInput(t *testing.T) {
	service := application.NewMasterDataService(testDB(t))
	tests := []struct {
		name     string
		typeName string
		keys     []string
	}{
		{name: "empty type", keys: []string{"first"}},
		{name: "empty keys", typeName: "manufacturer"},
		{name: "blank key", typeName: "manufacturer", keys: []string{"first", " "}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := service.SetActiveMany(t.Context(), test.typeName, test.keys, false); !errors.Is(
				err, application.ErrMasterDataValidation,
			) {
				t.Fatalf("error=%v", err)
			}
		})
	}

	tooMany := make([]string, application.MaxMasterDataActiveBatchSize+1)
	for index := range tooMany {
		tooMany[index] = fmt.Sprintf("key-%d", index)
	}
	if _, err := service.SetActiveMany(t.Context(), "manufacturer", tooMany, false); !errors.Is(
		err, application.ErrMasterDataValidation,
	) {
		t.Fatalf("over-limit error=%v", err)
	}
}

func insertLifecycleEntry(
	t *testing.T,
	db *sql.DB,
	typeName, key, label, origin string,
	active bool,
) {
	t.Helper()
	if _, err := db.Exec(`
INSERT INTO master_data_entries(
  id, type, key, label, active, sort_order, source_url, metadata_json,
  created_at, updated_at, origin
) VALUES(?, ?, ?, ?, ?, 0, '', '{}', 'now', 'now', ?)`,
		typeName+":"+key, typeName, key, label, boolInt(active), origin); err != nil {
		t.Fatal(err)
	}
}

func useVehicleColumn(column string) func(*testing.T, *sql.DB, string, string) {
	return func(t *testing.T, db *sql.DB, _ string, label string) {
		t.Helper()
		query := `INSERT INTO vehicles(
  id, inventory_number, manufacturer, name, gauge, epoch, railway_company, category,
  gattung, created_at, updated_at
) VALUES('v1', 'RK-1', 'M', 'Test', 'G', 'E', 'R', 'C', 'T', 'now', 'now')`
		if _, err := db.Exec(query); err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(`UPDATE vehicles SET `+column+`=? WHERE id='v1'`, label); err != nil {
			t.Fatal(err)
		}
	}
}

func useExhibitionColumn(column string) func(*testing.T, *sql.DB, string, string) {
	return func(t *testing.T, db *sql.DB, _ string, label string) {
		t.Helper()
		if _, err := db.Exec(`
INSERT INTO exhibition_lists(id, designation, list_date, created_at, updated_at)
VALUES('l1', 'List', '2026-01-01', 'now', 'now')`); err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(`
INSERT INTO exhibition_entries(id, list_id, owner, locomotive_name, created_at, updated_at)
VALUES('e1', 'l1', 'Club', 'Test', 'now', 'now')`); err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(`UPDATE exhibition_entries SET `+column+`=? WHERE id='e1'`, label); err != nil {
			t.Fatal(err)
		}
	}
}

func useAccessoryColumn(column string) func(*testing.T, *sql.DB, string, string) {
	return func(t *testing.T, db *sql.DB, key, label string) {
		t.Helper()
		if _, err := db.Exec(`
INSERT INTO accessory_products(
  id, inventory_number, manufacturer, name, category, tracking_mode, created_at, updated_at
) VALUES('p1', 'RK-ART-1', 'M', 'Test', 'C', 'quantity', 'now', 'now')`); err != nil {
			t.Fatal(err)
		}
		value := any(key)
		if column == "manufacturer" {
			value = label
		}
		if column == "gauges_json" {
			value = `["` + label + `"]`
		}
		if _, err := db.Exec(`UPDATE accessory_products SET `+column+`=? WHERE id='p1'`, value); err != nil {
			t.Fatal(err)
		}
	}
}

func useFunctionSymbol(t *testing.T, db *sql.DB, key, _ string) {
	t.Helper()
	useVehicleColumn("manufacturer")(t, db, "", "M")
	if _, err := db.Exec(`
INSERT INTO vehicle_functions(
  id, vehicle_id, function_key, symbol_key, created_at, updated_at
) VALUES('f1', 'v1', 'F0', ?, 'now', 'now')`, key); err != nil {
		t.Fatal(err)
	}
}

func useCustomField(t *testing.T, db *sql.DB, key, _ string) {
	t.Helper()
	useAccessoryColumn("article_type")(t, db, "other", "")
	if _, err := db.Exec(`
INSERT INTO accessory_product_attributes(
  id, product_id, attribute_key, value_type, text_value, created_at, updated_at
) VALUES('a1', 'p1', ?, 'text', 'value', 'now', 'now')`, key); err != nil {
		t.Fatal(err)
	}
}

func assertCapabilities(
	t *testing.T,
	got application.MasterDataCapabilities,
	deactivate, reactivate, deleteEntry bool,
) {
	t.Helper()
	if got.CanDeactivate != deactivate || got.CanReactivate != reactivate || got.CanDelete != deleteEntry {
		t.Fatalf("capabilities=%#v, want deactivate=%t reactivate=%t delete=%t",
			got, deactivate, reactivate, deleteEntry)
	}
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
