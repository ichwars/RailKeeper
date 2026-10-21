package infrastructure_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"railkeeper/backend/internal/infrastructure"
)

func TestFunctionFlagMigrationRetainsAmbiguousValues(t *testing.T) {
	root := t.TempDir()
	partial := filepath.Join(root, "before-flags")
	if err := os.Mkdir(partial, 0700); err != nil {
		t.Fatal(err)
	}
	copyMigrationsThrough(t, filepath.Join("..", "..", "migrations"), partial, "0070_retire_planner_role.sql")
	db, err := infrastructure.OpenSQLite(filepath.Join(root, "data"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err = infrastructure.Migrate(db, partial); err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO vehicles(id,inventory_number,manufacturer,name,gauge,category,gattung,created_at,updated_at) VALUES('flags-lok','FLAGS-1','ESU','Flags','TT','Lokomotive','Diesellok','before','before')`)
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []int{0, 1} {
		_, err = db.Exec(`INSERT INTO vehicle_functions(id,vehicle_id,function_key,name,mode,direction_dependent,notes,created_at,updated_at) VALUES(?, 'flags-lok', ?, 'Eigener Name','moment',?,'Eigene Notiz','before','before')`, value, string(rune('0'+value)), value)
		if err != nil {
			t.Fatal(err)
		}
	}
	if err = infrastructure.Migrate(db, filepath.Join("..", "..", "migrations")); err != nil {
		t.Fatal(err)
	}
	rows, err := db.Query(`SELECT direction_dependent,inverted,legacy_direction_dependent,name,mode,notes,updated_at FROM vehicle_functions ORDER BY id`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		var direction, inverted sql.NullBool
		var legacy int
		var name, mode, notes, updated string
		if err = rows.Scan(&direction, &inverted, &legacy, &name, &mode, &notes, &updated); err != nil {
			t.Fatal(err)
		}
		if direction.Valid || inverted.Valid || legacy != count || name != "Eigener Name" || mode != "moment" || notes != "Eigene Notiz" || updated != "before" {
			t.Fatalf("migration lost or guessed data: %v %v %d", direction, inverted, legacy)
		}
		count++
	}
	if count != 2 {
		t.Fatal(count)
	}
}
