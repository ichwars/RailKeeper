package infrastructure

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCVIndexMigrationPreservesExistingValuesAndHistory(t *testing.T) {
	db, err := OpenSQLite(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close test database: %v", err)
		}
	})
	if _, err := db.Exec(`PRAGMA foreign_keys=ON; CREATE TABLE vehicles(id TEXT PRIMARY KEY); INSERT INTO vehicles VALUES('v')`); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"0014_vehicle_cv_values.sql", "0030_vehicle_cv_protocol.sql"} {
		body, err := os.ReadFile(filepath.Join("..", "..", "migrations", name))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(string(body)); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec(`INSERT INTO vehicle_cv_values(id,vehicle_id,cv_number,value,created_at,updated_at) VALUES('cv','v',1,3,'old','old');
INSERT INTO vehicle_cv_value_history VALUES('history','cv','v',2,3,'old')`); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	name := "0067_vehicle_cv_indexes.sql"
	body, err := os.ReadFile(filepath.Join("..", "..", "migrations", name))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), body, 0600); err != nil {
		t.Fatal(err)
	}
	if err := Migrate(db, dir); err != nil {
		t.Fatal(err)
	}
	var value, history int
	if err := db.QueryRow(`SELECT value FROM vehicle_cv_values WHERE id='cv' AND index_cv31 IS NULL AND index_cv32 IS NULL`).Scan(&value); err != nil || value != 3 {
		t.Fatalf("value lost: %v %d", err, value)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM vehicle_cv_value_history WHERE cv_value_id='cv'`).Scan(&history); err != nil || history != 1 {
		t.Fatalf("history lost: %v %d", err, history)
	}
	if _, err := db.Exec(`INSERT INTO vehicle_cv_values(id,vehicle_id,cv_number,value,created_at,updated_at) VALUES('duplicate','v',1,4,'new','new')`); err == nil {
		t.Fatal("unindexed duplicate accepted")
	}
	if _, err := db.Exec(`DELETE FROM vehicle_cv_values WHERE id='cv'`); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM vehicle_cv_value_history`).Scan(&history); err != nil || history != 0 {
		t.Fatalf("history foreign key lost: %v %d", err, history)
	}
}
