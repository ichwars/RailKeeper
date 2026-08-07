package infrastructure_test

import (
	"path/filepath"
	"slices"
	"testing"

	"railkeeper/backend/internal/infrastructure"
)

func TestSeedRolesIncludesPlannerAndIsIdempotent(t *testing.T) {
	db, err := infrastructure.OpenSQLite(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := infrastructure.Migrate(db, filepath.Join("..", "..", "migrations")); err != nil {
		t.Fatal(err)
	}
	if err := infrastructure.SeedRoles(db); err != nil {
		t.Fatal(err)
	}
	if err := infrastructure.SeedRoles(db); err != nil {
		t.Fatal(err)
	}

	rows, err := db.Query(`SELECT name FROM roles ORDER BY name`)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()

	var roles []string
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			t.Fatal(err)
		}
		roles = append(roles, role)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}

	want := []string{"Admin", "Editor", "Messe", "Planner", "Viewer"}
	if !slices.Equal(roles, want) {
		t.Fatalf("unexpected roles: got %v, want %v", roles, want)
	}
}
