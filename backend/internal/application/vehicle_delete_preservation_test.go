package application_test

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"railkeeper/backend/internal/application"
)

func databaseRowFingerprints(t *testing.T, db *sql.DB) map[string]string {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	tables, err := db.QueryContext(ctx, `SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name`)
	if err != nil {
		t.Fatal(err)
	}
	names := []string{}
	for tables.Next() {
		var name string
		if err := tables.Scan(&name); err != nil {
			t.Fatal(err)
		}
		names = append(names, name)
	}
	if err := tables.Err(); err != nil {
		t.Fatal(err)
	}
	if err := tables.Close(); err != nil {
		t.Fatal(err)
	}
	result := map[string]string{}
	for _, name := range names {
		rows, err := db.QueryContext(ctx, `SELECT * FROM "`+strings.ReplaceAll(name, `"`, `""`)+`"`)
		if err != nil {
			t.Fatal(name, err)
		}
		columns, err := rows.Columns()
		if err != nil {
			t.Fatal(name, err)
		}
		records := []string{}
		for rows.Next() {
			values := make([]any, len(columns))
			targets := make([]any, len(columns))
			for i := range values {
				targets[i] = &values[i]
			}
			if err := rows.Scan(targets...); err != nil {
				t.Fatal(name, err)
			}
			encoded, err := json.Marshal(values)
			if err != nil {
				t.Fatal(name, err)
			}
			records = append(records, string(encoded))
		}
		if err := rows.Err(); err != nil {
			t.Fatal(name, err)
		}
		if err := rows.Close(); err != nil {
			t.Fatal(name, err)
		}
		sort.Strings(records)
		sum := sha256.Sum256([]byte(strings.Join(records, "\n")))
		result[name] = hex.EncodeToString(sum[:])
	}
	return result
}
func TestVehicleDeletionAuditFailureRestoresEveryTableAndLastSetMembership(t *testing.T) {
	db := testDB(t)
	db.SetMaxOpenConns(1)
	service := application.NewVehicleService(db)
	created, err := service.CreateSet(t.Context(), setMutationInput(), "actor")
	if err != nil {
		t.Fatal(err)
	}
	if err := service.Delete(t.Context(), created.Members[0].ID, "actor"); err != nil {
		t.Fatal(err)
	}
	lastID := created.Members[1].ID
	file, err := service.CreateCVFile(t.Context(), lastID, application.VehicleCVFileInput{
		FileName: "decoder.csv", OriginalName: "Original decoder.csv", DecoderProfile: "recorded profile", MimeType: "text/csv",
		StoragePath: "uploads/vehicles/original/cv/decoder.csv", SizeBytes: 12,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.CreateCVValue(t.Context(), lastID, application.VehicleCVValueInput{
		CVNumber: 29, Value: 38, DecoderProfile: "recorded profile", SourceFileID: file.ID, Description: "recorded evidence",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := service.CreateAttachment(t.Context(), lastID, application.VehicleAttachmentInput{
		FileName: "manual.pdf", OriginalName: "Original manual.pdf", StoragePath: "uploads/vehicles/original/manual.pdf", MimeType: "application/pdf", SizeBytes: 6,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TRIGGER fail_vehicle_delete_audit BEFORE INSERT ON audit_logs
WHEN NEW.action='VehicleDeleted' BEGIN SELECT RAISE(ABORT,'synthetic deletion audit failure'); END`); err != nil {
		t.Fatal(err)
	}
	before := databaseRowFingerprints(t, db)
	if err := service.Delete(t.Context(), lastID, "actor"); err == nil || !strings.Contains(err.Error(), "synthetic deletion audit failure") {
		t.Fatal("expected failure after cascade and empty-set deletion", err)
	}
	after := databaseRowFingerprints(t, db)
	if !reflect.DeepEqual(before, after) {
		changed := []string{}
		for table, value := range before {
			if after[table] != value {
				changed = append(changed, table)
			}
		}
		sort.Strings(changed)
		t.Fatal("rollback changed database tables", changed)
	}
	if _, err := db.Exec(`DROP TRIGGER fail_vehicle_delete_audit`); err != nil {
		t.Fatal(err)
	}
	if err := service.Delete(t.Context(), lastID, "actor"); err != nil {
		t.Fatal("retry after rollback failed", err)
	}
	if _, err := service.GetSet(t.Context(), created.ID); !errors.Is(err, application.ErrVehicleSetNotFound) {
		t.Fatal("final member deletion retained empty set", err)
	}
}
