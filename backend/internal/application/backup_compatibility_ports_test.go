package application_test

import (
	"bytes"
	"database/sql"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestLegacyBackupRestoreRetainsRawNullableArticleRows(t *testing.T) {
	for _, version := range []int{1, 2} {
		t.Run(backupVersionName(version), func(t *testing.T) {
			db := backupTestDB(t, t.TempDir())
			db.SetMaxOpenConns(1)
			for _, setting := range []struct {
				key    string
				source any
			}{{"track", nil}, {"signal", ""}} {
				if _, err := db.Exec(`UPDATE master_data_entries SET label='Recorded local label',active=0,
sort_order=77,source_url=?,metadata_json=' { "recorded": " original metadata " } ',
created_at='2025-01-01T01:02:03Z',updated_at='2025-01-02T04:05:06.123456789Z',origin='custom'
WHERE type='article_type' AND key=?`, setting.source, setting.key); err != nil {
					t.Fatal(err)
				}
			}
			read := func() string {
				t.Helper()
				var state string
				if err := db.QueryRow(`SELECT json_group_array(json_array(id,type,key,label,active,sort_order,
source_url,typeof(source_url),metadata_json,created_at,updated_at,origin)) FROM
(SELECT * FROM master_data_entries WHERE type IN ('article_type','accessory_subtype') ORDER BY type,key)`).Scan(&state); err != nil {
					t.Fatal(err)
				}
				return state
			}
			before := read()
			doc := &application.BackupDocument{
				Format: "railkeeper-backup", Version: version, Tables: backupDocumentTablesThroughVersion(version),
			}
			if _, err := application.NewBackupService(db, t.TempDir()).Import(t.Context(), doc); err != nil {
				t.Fatal(err)
			}
			if read() != before {
				t.Fatal("legacy restore changed raw article metadata, SQL NULL/empty, storage type, IDs, origin or timestamps")
			}
		})
	}
}

func TestLegacyBackupRestoreFinalizationFailuresPreserveRowsTriggersAndOriginalFiles(t *testing.T) {
	for _, stage := range []struct{ name, event, predicate, context string }{
		{"articles", "INSERT", "NEW.type='article_type'", "restore current article master data"},
		{"symbols", "INSERT", "NEW.type='symbols'", "restore bundled function symbol"},
		{"decoder-docs", "INSERT", "NEW.type='decoder_profile'", "seed decoder documentation"},
		{"origins", "UPDATE OF origin", "NEW.origin='bundled'", "reconcile bundled master data"},
	} {
		t.Run(stage.name, func(t *testing.T) {
			root := t.TempDir()
			db := backupTestDB(t, root)
			db.SetMaxOpenConns(1)
			path := filepath.Join(root, "uploads", "original.txt")
			original := []byte("original private upload bytes")
			if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, original, 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := db.Exec(`CREATE TRIGGER fail_compatibility BEFORE ` + stage.event + ` ON master_data_entries
WHEN ` + stage.predicate + ` BEGIN SELECT RAISE(FAIL,'synthetic compatibility failure'); END`); err != nil {
				t.Fatal(err)
			}
			triggers := func() string {
				t.Helper()
				var value string
				if err := db.QueryRow(`SELECT COALESCE(group_concat(name||'='||sql,char(10)),'') FROM
(SELECT name,sql FROM sqlite_master WHERE type='trigger' ORDER BY name)`).Scan(&value); err != nil {
					t.Fatal(err)
				}
				return value
			}
			before, beforeTriggers := databaseRowFingerprints(t, db), triggers()
			doc := &application.BackupDocument{Format: "railkeeper-backup", Version: 1, Tables: backupDocumentTablesThroughVersion(1)}
			service := application.NewBackupService(db, root)
			result, err := service.Import(t.Context(), doc)
			if result != nil || err == nil || !strings.Contains(err.Error(), "synthetic compatibility failure") || !strings.Contains(err.Error(), stage.context) {
				t.Fatal("intended finalization failure was not observed", result, err)
			}
			if !reflect.DeepEqual(before, databaseRowFingerprints(t, db)) || beforeTriggers != triggers() {
				t.Fatal("failed finalization changed database rows, authentication or projection triggers")
			}
			if data, err := os.ReadFile(path); err != nil || !bytes.Equal(data, original) {
				t.Fatal("failed finalization replaced the original upload", err)
			}
			if db.Stats().InUse != 0 {
				t.Fatal("failed finalization retained its transaction")
			}
			if _, err := db.Exec(`DROP TRIGGER fail_compatibility`); err != nil {
				t.Fatal(err)
			}
			if _, err := service.Import(t.Context(), doc); err != nil {
				t.Fatal("retry after complete rollback failed", err)
			}
		})
	}
}

func TestLegacyBackupRestoreProfileCreatorsRespectLocalAccountsWithoutRestoringAuthentication(t *testing.T) {
	db := backupTestDB(t, t.TempDir())
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`INSERT INTO users(id,username,email,password_hash,created_at)
VALUES('local-owner','LocalOwner','local@example.test','synthetic-unused-hash','2026-01-01T00:00:00Z')`); err != nil {
		t.Fatal(err)
	}
	service := application.NewBackupService(db, t.TempDir())
	doc, err := service.Export(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	doc.Profiles = nil
	for _, owner := range []string{"local-owner", "unknown-owner"} {
		doc.Profiles = append(doc.Profiles, map[string]any{
			"id": owner, "name": owner, "direction": "export", "format": "csv", "areas_json": `["vehicles"]`,
			"options_json": `{"delimiter":";"}`, "enabled": int64(1), "created_by_user_id": owner,
			"created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z",
		})
	}
	before := databaseRowFingerprints(t, db)
	if _, err := service.Import(t.Context(), doc); err != nil {
		t.Fatal(err)
	}
	for _, owner := range []string{"local-owner", "unknown-owner"} {
		var actual sql.NullString
		if err := db.QueryRow(`SELECT created_by_user_id FROM data_transfer_profiles WHERE id=?`, owner).Scan(&actual); err != nil {
			t.Fatal(err)
		}
		if (owner == "local-owner" && (!actual.Valid || actual.String != owner)) || (owner == "unknown-owner" && actual.Valid) {
			t.Fatal("restored profile did not respect existing local ownership", owner)
		}
	}
	after := databaseRowFingerprints(t, db)
	for table, state := range before {
		if _, restored := doc.Tables[table]; !restored && table != "data_transfer_profiles" && after[table] != state {
			t.Fatal("restore changed local authentication or another table outside the application backup", table)
		}
	}
}
