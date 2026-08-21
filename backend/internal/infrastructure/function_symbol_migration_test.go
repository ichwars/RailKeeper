package infrastructure_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"railkeeper/backend/internal/infrastructure"
)

func TestFunctionSymbolMigrationReplacesBundledArtworkAndPreservesCustomRows(t *testing.T) {
	root := t.TempDir()
	migrationsDir := filepath.Join(root, "migrations")
	if err := os.Mkdir(migrationsDir, 0o700); err != nil {
		t.Fatal(err)
	}
	copyMigrationsThrough(t, filepath.Join("..", "..", "migrations"), migrationsDir,
		"0063_exhibition_workspace.sql")
	db, err := infrastructure.OpenSQLite(filepath.Join(root, "data"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := infrastructure.Migrate(db, migrationsDir); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE master_data_entries
SET metadata_json='{"sourceDocument":"retired.zip","imageData":"retired"}'
WHERE type='symbols' AND key='esu-f003-stirnbeleuchtung'`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`DELETE FROM master_data_entries
WHERE type='symbols' AND key='esu-f004-innenraumbeleuchtung'`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO master_data_entries(
id,type,key,label,active,sort_order,metadata_json,created_at,updated_at,origin
) VALUES('symbols:club','symbols','club','Club',1,900,
'{"imageData":"custom"}','now','now','custom')`); err != nil {
		t.Fatal(err)
	}

	applyMigrationFile(t, db, "0064_replace_bundled_function_symbols.sql")

	assertWerkstattSymbol(t, db, "esu-f003-stirnbeleuchtung", 3)
	assertWerkstattSymbol(t, db, "esu-f004-innenraumbeleuchtung", 4)
	var customMetadata string
	if err := db.QueryRow(`SELECT metadata_json FROM master_data_entries
WHERE type='symbols' AND key='club'`).Scan(&customMetadata); err != nil {
		t.Fatal(err)
	}
	if customMetadata != `{"imageData":"custom"}` {
		t.Fatalf("custom metadata changed: %s", customMetadata)
	}
}

func assertWerkstattSymbol(t *testing.T, db *sql.DB, key string, wantCode int) {
	t.Helper()
	var library string
	var code int
	var imageData string
	if err := db.QueryRow(`SELECT
json_extract(metadata_json, '$.library'),
json_extract(metadata_json, '$.ecosCode'),
json_extract(metadata_json, '$.imageData')
FROM master_data_entries WHERE type='symbols' AND key=?`, key).
		Scan(&library, &code, &imageData); err != nil {
		t.Fatal(err)
	}
	if library != "railkeeper-workshop-line" || code != wantCode {
		t.Fatalf("symbol %s library=%q code=%d", key, library, code)
	}
	if !strings.HasPrefix(imageData, "data:image/svg+xml;base64,") {
		t.Fatalf("symbol %s has invalid imageData", key)
	}
}
