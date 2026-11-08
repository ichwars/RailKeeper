package masterdatastore_test

import (
	"database/sql"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"railkeeper/backend/internal/infrastructure"
	"railkeeper/backend/internal/infrastructure/masterdatastore"
	"railkeeper/backend/internal/inventory/masterdata"
)

func masterDataDatabase(t *testing.T) (*sql.DB, *masterdata.Service) {
	t.Helper()
	db, err := infrastructure.OpenSQLite(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := infrastructure.Migrate(db, filepath.Join("..", "..", "..", "migrations")); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE master_data_entries SET origin='bundled'`); err != nil {
		t.Fatal(err)
	}
	service := masterdata.New(masterdatastore.New(db), nil)
	for _, key := range []string{"fixture-a", "fixture-b"} {
		if _, err := service.Create(t.Context(), "manufacturer", masterdata.MasterDataInput{
			Key: key, Label: key, Metadata: map[string]any{"recorded": "  original label  ", "empty": []any{}},
		}); err != nil {
			t.Fatal(err)
		}
	}
	return db, service
}

func masterDataRows(t *testing.T, db *sql.DB) map[string][][]any {
	t.Helper()
	state := map[string][][]any{}
	for _, table := range []string{"master_data_entries", "master_data_relations"} {
		rows, err := db.Query("SELECT * FROM " + table + " ORDER BY id")
		if err != nil {
			t.Fatal(err)
		}
		columns, err := rows.Columns()
		if err != nil {
			_ = rows.Close()
			t.Fatal(err)
		}
		for rows.Next() {
			values := make([]any, len(columns))
			pointers := make([]any, len(columns))
			for index := range values {
				pointers[index] = &values[index]
			}
			if err := rows.Scan(pointers...); err != nil {
				_ = rows.Close()
				t.Fatal(err)
			}
			for index, value := range values {
				if data, ok := value.([]byte); ok {
					values[index] = append([]byte(nil), data...)
				}
			}
			state[table] = append(state[table], values)
		}
		readErr := rows.Err()
		closeErr := rows.Close()
		if readErr != nil || closeErr != nil {
			t.Fatal(readErr, closeErr)
		}
	}
	return state
}

func TestMasterDataBatchRollsBackEarlierRowsOnLaterFailure(t *testing.T) {
	db, service := masterDataDatabase(t)
	before := masterDataRows(t, db)
	if _, err := db.Exec(`CREATE TRIGGER fail_later_active BEFORE UPDATE OF active ON master_data_entries
WHEN OLD.key='fixture-b' BEGIN SELECT RAISE(ABORT,'synthetic batch failure'); END`); err != nil {
		t.Fatal(err)
	}
	value, err := service.SetActiveMany(t.Context(), "manufacturer", []string{"fixture-a", "fixture-b"}, false)
	if err == nil || !strings.Contains(err.Error(), "synthetic batch failure") || value != nil {
		t.Fatal("failed batch was acknowledged", value, err)
	}
	if !reflect.DeepEqual(before, masterDataRows(t, db)) {
		t.Fatal("failed batch changed an earlier row, identity, timestamp or raw metadata")
	}
}

func TestMasterDataImportRollsBackDeletionAndPartialInsertion(t *testing.T) {
	db, service := masterDataDatabase(t)
	before := masterDataRows(t, db)
	if _, err := db.Exec(`CREATE TRIGGER fail_import_insert BEFORE INSERT ON master_data_entries
WHEN NEW.key='fixture-new' BEGIN SELECT RAISE(ABORT,'synthetic import failure'); END`); err != nil {
		t.Fatal(err)
	}
	value, err := service.Import(t.Context(), &masterdata.MasterDataDocument{
		Format: masterdata.ExportFormat, Version: masterdata.ExportVersion,
		Entries: map[string][]masterdata.MasterDataEntry{"manufacturer": {{Key: "fixture-new", Label: "New"}}},
	})
	if err == nil || !strings.Contains(err.Error(), "synthetic import failure") || value != nil {
		t.Fatal("failed import was acknowledged", value, err)
	}
	if !reflect.DeepEqual(before, masterDataRows(t, db)) {
		t.Fatal("failed import lost entries, relations, identities, timestamps or raw metadata")
	}
}

func TestMasterDataDeleteRestoresRelationsWhenEntryDeletionFails(t *testing.T) {
	db, service := masterDataDatabase(t)
	if _, err := db.Exec(`INSERT INTO master_data_relations(id,parent_type,parent_key,child_type,child_key,sort_order,created_at)
VALUES('recorded-relation','manufacturer','fixture-a','manufacturer','fixture-b',0,'2020-01-02T03:04:05Z');
CREATE TRIGGER fail_entry_delete BEFORE DELETE ON master_data_entries
WHEN OLD.key='fixture-a' BEGIN SELECT RAISE(ABORT,'synthetic delete failure'); END`); err != nil {
		t.Fatal(err)
	}
	before := masterDataRows(t, db)
	if err := service.Delete(t.Context(), "manufacturer", "fixture-a"); err == nil || !strings.Contains(err.Error(), "synthetic delete failure") {
		t.Fatal("failed deletion was acknowledged", err)
	}
	if !reflect.DeepEqual(before, masterDataRows(t, db)) {
		t.Fatal("failed deletion lost an already removed relation or changed existing entries")
	}
}
