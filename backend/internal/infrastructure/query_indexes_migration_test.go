package infrastructure_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"railkeeper/backend/internal/infrastructure"
)

func TestQueryIndexMigrationsPreserveDataAndImprovePlans(t *testing.T) {
	ctx := t.Context()
	root := t.TempDir()
	originalMigrations := filepath.Join(root, "original-migrations")
	if err := os.Mkdir(originalMigrations, 0700); err != nil {
		t.Fatal(err)
	}
	migrations := filepath.Join("..", "..", "migrations")
	copyMigrationsThrough(t, migrations, originalMigrations, "0081_login_origins.sql")
	db, err := infrastructure.OpenSQLite(filepath.Join(root, "data"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := infrastructure.Migrate(db, originalMigrations); err != nil {
		t.Fatal(err)
	}

	if _, err := db.ExecContext(ctx, `
PRAGMA foreign_keys=ON;
INSERT INTO vehicles(id, inventory_number, manufacturer, name, gauge, created_at, updated_at)
VALUES('v','RK-001','Tillig','BR 118','TT','2026-01-01','2026-01-01');
INSERT INTO vehicle_cv_values(id,vehicle_id,cv_number,value,source_file_id,created_at,updated_at)
VALUES('cv','v',1,3,'source','2026-01-01','2026-01-01');
INSERT INTO vehicle_cv_value_history(id,cv_value_id,vehicle_id,old_value,new_value,changed_at)
VALUES('old','cv','v',1,2,'2026-01-01'),('new','cv','v',2,3,'2026-01-02');
INSERT INTO audit_logs(id,action,created_at)
VALUES('audit-1','test','2026-01-01'),('audit-2','test','2026-01-01');
`); err != nil {
		t.Fatal(err)
	}

	const history = `SELECT id FROM vehicle_cv_value_history WHERE vehicle_id=? ORDER BY changed_at DESC,id DESC`
	const source = `SELECT id FROM vehicle_cv_values WHERE source_file_id=? ORDER BY id`
	const audit = `SELECT id FROM audit_logs ORDER BY created_at DESC,id DESC LIMIT ?`
	cases := []struct {
		query, arg, wantIndex, wantRows string
	}{
		{history, "v", "idx_vehicle_cv_history_vehicle_order", "new,old"},
		{source, "source", "idx_vehicle_cv_values_source", "cv"},
		{audit, "1", "idx_audit_logs_created_order", "audit-2"},
	}
	for _, test := range cases {
		if plan := queryIndexPlan(t, db, test.query, test.arg); strings.Contains(plan, test.wantIndex) {
			t.Fatalf("index %s existed before the upgrade", test.wantIndex)
		}
	}

	if err := infrastructure.Migrate(db, migrations); err != nil {
		t.Fatal(err)
	}
	if err := infrastructure.Migrate(db, migrations); err != nil {
		t.Fatalf("restart after upgrade: %v", err)
	}
	for _, test := range cases {
		plan := queryIndexPlan(t, db, test.query, test.arg)
		if !strings.Contains(plan, test.wantIndex) || strings.Contains(plan, "TEMP B-TREE") {
			t.Fatalf("expected %s without a separate sort, plan %s", test.wantIndex, plan)
		}
		if rows := queryIndexIDs(t, db, test.query, test.arg); rows != test.wantRows {
			t.Fatalf("rows after upgrade: got %q, want %q", rows, test.wantRows)
		}
	}
	if result := queryIndexIDs(t, db, "PRAGMA integrity_check"); result != "ok" {
		t.Fatalf("integrity_check: %q", result)
	}
	fkRows, err := db.QueryContext(ctx, "PRAGMA foreign_key_check")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = fkRows.Close() }()
	if fkRows.Next() {
		t.Fatal("foreign_key_check reported a violation")
	}
	if err := fkRows.Err(); err != nil {
		t.Fatal(err)
	}
}

func queryIndexPlan(t *testing.T, db *sql.DB, query string, arg any) string {
	t.Helper()
	rows, err := db.QueryContext(t.Context(), "EXPLAIN QUERY PLAN "+query, arg)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	var steps []string
	for rows.Next() {
		var id, parent, aux int
		var detail string
		if err := rows.Scan(&id, &parent, &aux, &detail); err != nil {
			t.Fatal(err)
		}
		steps = append(steps, detail)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return strings.Join(steps, " | ")
}

func queryIndexIDs(t *testing.T, db *sql.DB, query string, args ...any) string {
	t.Helper()
	rows, err := db.QueryContext(t.Context(), query, args...)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			t.Fatal(err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return strings.Join(ids, ",")
}
