package backupdata_test

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"railkeeper/backend/internal/backup"
	"railkeeper/backend/internal/infrastructure/backupdata"

	"modernc.org/sqlite"
)

type faultPlanner struct{ stage string }
type faultOperation struct {
	doc   *backup.BackupDocument
	stage string
}

func (p faultPlanner) Begin(_ context.Context, _ *sql.Tx, doc *backup.BackupDocument) (backupdata.RestoreOperation, error) {
	return faultOperation{doc: doc, stage: p.stage}, nil
}
func (op faultOperation) Rows(_ context.Context, _ *sql.Tx, table string) ([]map[string]any, error) {
	if op.stage == "rows" && table == "vehicles" {
		return nil, errors.New("synthetic row policy failure")
	}
	return op.doc.Tables[table], nil
}
func (faultOperation) ProfileRow(_ context.Context, _ *sql.Tx, row map[string]any) (map[string]any, error) {
	return row, nil
}
func (op faultOperation) Finish(ctx context.Context, tx *sql.Tx, _ string) error {
	if op.stage == "finish" {
		return errors.New("synthetic finalization failure")
	}
	if op.stage == "commit" {
		_, err := tx.ExecContext(ctx, `INSERT INTO commit_guard(id,vehicle_id) VALUES('deferred','missing-vehicle')`)
		return err
	}
	return nil
}

func TestRestoreFaultsPreserveAllRowsUploadsAndAuthentication(t *testing.T) {
	for _, stage := range []string{"rows", "finish", "commit", "sqlite-full", "upload-stage"} {
		t.Run(stage, func(t *testing.T) {
			root := t.TempDir()
			db, err := sql.Open("sqlite", filepath.Join(root, "test.db"))
			if err != nil {
				t.Fatal(err)
			}
			db.SetMaxOpenConns(1)
			t.Cleanup(func() { _ = db.Close() })
			if _, err := db.ExecContext(t.Context(), `PRAGMA foreign_keys=ON`); err != nil {
				t.Fatal(err)
			}
			doc := &backup.BackupDocument{Format: backup.Format, Version: backup.Version, Tables: map[string][]map[string]any{}, Profiles: []map[string]any{},
				Files: []backup.BackupFile{{Path: "uploads/restored.txt", ContentBase64: base64.StdEncoding.EncodeToString([]byte("new upload"))}}}
			for _, table := range backup.TableOrder() {
				if _, err := db.ExecContext(t.Context(), `CREATE TABLE `+backupdata.QuoteIdentifier(table)+` (id TEXT PRIMARY KEY)`); err != nil {
					t.Fatal(err)
				}
				if _, err := db.ExecContext(t.Context(), `INSERT INTO `+backupdata.QuoteIdentifier(table)+`(id) VALUES(?)`, "original-"+table); err != nil {
					t.Fatal(err)
				}
				doc.Tables[table] = []map[string]any{{"id": "restored-" + table}}
			}
			for _, statement := range []string{
				`CREATE TABLE users(id TEXT PRIMARY KEY,password_hash TEXT)`,
				`INSERT INTO users VALUES('local-owner','synthetic-private-hash')`,
				`CREATE TABLE data_transfer_profiles(id TEXT PRIMARY KEY,name_key TEXT,created_by_user_id TEXT)`,
				`INSERT INTO data_transfer_profiles VALUES('local-profile','private-name-key','local-owner')`,
				`CREATE TABLE commit_guard(id TEXT PRIMARY KEY,vehicle_id TEXT REFERENCES vehicles(id) DEFERRABLE INITIALLY DEFERRED)`,
			} {
				if _, err := db.ExecContext(t.Context(), statement); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.MkdirAll(filepath.Join(root, "uploads"), 0o700); err != nil {
				t.Fatal(err)
			}
			original := []byte("original upload bytes")
			file := filepath.Join(root, "uploads", "original.txt")
			if err := os.WriteFile(file, original, 0o600); err != nil {
				t.Fatal(err)
			}
			store := backupdata.NewStore(db, root, faultPlanner{stage: stage})
			before, err := store.Snapshot(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			if _, exists := before.Tables["users"]; exists {
				t.Fatal("authentication included in snapshot")
			}
			if before.Profiles[0]["created_by_user_id"] != "local-owner" {
				t.Fatal("existing profile creator lost")
			}
			if _, exists := before.Profiles[0]["name_key"]; exists {
				t.Fatal("internal profile key exported")
			}
			if stage == "sqlite-full" {
				var pages int
				if err := db.QueryRowContext(t.Context(), `PRAGMA page_count`).Scan(&pages); err != nil {
					t.Fatal(err)
				}
				if _, err := db.ExecContext(t.Context(), fmt.Sprintf("PRAGMA max_page_count=%d", pages)); err != nil {
					t.Fatal(err)
				}
				doc.Tables["vehicles"] = []map[string]any{{"id": strings.Repeat("synthetic-full-", 32768)}}
			}
			if stage == "upload-stage" {
				doc.Files = append(doc.Files,
					backup.BackupFile{Path: "uploads/conflict", ContentBase64: base64.StdEncoding.EncodeToString([]byte("file"))},
					backup.BackupFile{Path: "uploads/conflict/child.txt", ContentBase64: base64.StdEncoding.EncodeToString([]byte("blocked child"))})
			}
			_, restoreErr := store.Restore(t.Context(), doc)
			if restoreErr == nil {
				t.Fatal("injected restore failure succeeded")
			}
			if stage == "sqlite-full" {
				var sqliteErr *sqlite.Error
				if !errors.As(restoreErr, &sqliteErr) || sqliteErr.Code() != 13 {
					t.Fatal("expected actual SQLite FULL failure", restoreErr)
				}
			}
			if stage == "upload-stage" {
				var pathErr *os.PathError
				if !errors.As(restoreErr, &pathErr) {
					t.Fatal("expected actual staged-filesystem failure", restoreErr)
				}
			}
			after, err := store.Snapshot(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(before, after) {
				t.Fatal("failed restore changed application rows")
			}
			actual, err := os.ReadFile(file)
			if err != nil || !reflect.DeepEqual(actual, original) {
				t.Fatal("failed restore changed original upload")
			}
			if _, err := os.Stat(filepath.Join(root, "uploads", "restored.txt")); !errors.Is(err, os.ErrNotExist) {
				t.Fatal("failed restore activated a new upload")
			}
			var password string
			if err := db.QueryRowContext(t.Context(), `SELECT password_hash FROM users WHERE id='local-owner'`).Scan(&password); err != nil || password != "synthetic-private-hash" {
				t.Fatal("local authentication changed")
			}
			leftovers, err := filepath.Glob(filepath.Join(root, ".restore-*"))
			if err != nil || len(leftovers) != 0 {
				t.Fatal("restore staging remained after failure")
			}
		})
	}
}
