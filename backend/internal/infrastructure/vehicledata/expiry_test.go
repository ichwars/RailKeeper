package vehicledata_test

import (
	"database/sql"
	"strings"
	"testing"

	"railkeeper/backend/internal/infrastructure/vehicledata"
	vehiclemodel "railkeeper/backend/internal/inventory/vehicle"
)

func TestExhibitionExpiryCanRetryLegacyFailureWithoutRestampingCompletedVehicles(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })
	for _, statement := range []string{
		`CREATE TABLE vehicles(id TEXT PRIMARY KEY,exhibition INTEGER,updated_at TEXT,name TEXT,manufacturer TEXT,digital_decoder_number TEXT)`,
		`CREATE TABLE exhibition_entries(list_id TEXT,vehicle_id TEXT,decoder_number TEXT,locomotive_name TEXT,manufacturer TEXT)`,
		`CREATE TABLE exhibition_lists(id TEXT PRIMARY KEY,list_date TEXT)`,
		`INSERT INTO vehicles VALUES('direct',1,'recorded','Direct','M','3'),('legacy',1,'recorded','Legacy','M','4'),('active',1,'recorded','Active','M','5')`,
		`INSERT INTO exhibition_lists VALUES('past',date('now','localtime','-2 day')),('future',date('now','localtime','+2 day'))`,
		`INSERT INTO exhibition_entries VALUES('past','direct','','',''),('past','','4','',''),('past','','5','',''),('future','active','','','')`,
		`CREATE TRIGGER fail_legacy_expiry BEFORE UPDATE ON vehicles WHEN NEW.id='legacy' AND NEW.exhibition=0
BEGIN SELECT RAISE(ABORT,'synthetic legacy expiry failure'); END`,
	} {
		if _, err := db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	repository := vehicledata.NewExpiryStore(db)
	first := "2026-01-01T01:00:00Z"
	second := "2026-01-01T02:00:00Z"
	if err := vehiclemodel.ResetExpiredExhibitionFlags(t.Context(), repository, first); err == nil || !strings.Contains(err.Error(), "synthetic legacy expiry failure") {
		t.Fatal("legacy failure acknowledged", err)
	}
	check := func(id string, wantFlag int, wantTime string) {
		t.Helper()
		var flag int
		var timestamp string
		if err := db.QueryRow(`SELECT exhibition,updated_at FROM vehicles WHERE id=?`, id).Scan(&flag, &timestamp); err != nil {
			t.Fatal(err)
		}
		if flag != wantFlag || timestamp != wantTime {
			t.Fatal("wrong expiry/retry state", id, flag, timestamp)
		}
	}
	check("direct", 0, first)
	check("legacy", 1, "recorded")
	check("active", 1, "recorded")
	if _, err := db.Exec(`DROP TRIGGER fail_legacy_expiry`); err != nil {
		t.Fatal(err)
	}
	if err := vehiclemodel.ResetExpiredExhibitionFlags(t.Context(), repository, second); err != nil {
		t.Fatal(err)
	}
	check("direct", 0, first)
	check("legacy", 0, second)
	check("active", 1, "recorded")
}
