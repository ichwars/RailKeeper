package application_test

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestDecoderConfigurationLegacyRestoreUsesOnlyStagedOriginalsAndRollsBackAllEvidence(t *testing.T) {
	for _, scenario := range []struct {
		name, separator string
		available       bool
	}{{"original in backup", "/", true}, {"original absent from backup", "/", false},
		{"Windows original in backup", `\`, true}, {"Windows original absent from backup", `\`, false}} {
		t.Run(scenario.name, func(t *testing.T) {
			available := scenario.available
			db, backup, root, vehicle := legacyDecoderCaptureFixture(t)
			oldContent, restoredContent := []byte("Previous target content"), []byte("Restored original from backup")
			path := filepath.Join(root, "uploads", "legacy.csv")
			if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, oldContent, 0600); err != nil {
				t.Fatal(err)
			}
			file, err := application.NewVehicleService(db).CreateCVFile(t.Context(), vehicle, application.VehicleCVFileInput{
				FileName: "legacy.csv", OriginalName: "legacy.csv", DecoderProfile: "Legacy",
				StoragePath: strings.ReplaceAll("uploads/legacy.csv", "/", scenario.separator),
				MimeType:    "text/csv", SizeBytes: int64(len(oldContent)),
			})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := db.Exec(`UPDATE vehicle_cv_values SET source_file_id=?`, file.ID); err != nil {
				t.Fatal(err)
			}
			before := decoderSnapshots(t, backup)
			doc := decoderSnapshots(t, backup)
			setLegacyBackupVersion(doc, 26)
			doc.Tables["vehicle_cv_values"][0]["value"] = int64(42)
			doc.Tables["vehicle_cv_files"][0]["size_bytes"] = int64(len(restoredContent))
			doc.Files = nil
			if available {
				sum := sha256.Sum256(restoredContent)
				doc.Files = []application.BackupFile{{Path: "uploads/legacy.csv", SizeBytes: int64(len(restoredContent)),
					SHA256: hex.EncodeToString(sum[:]), ContentBase64: base64.StdEncoding.EncodeToString(restoredContent)}}
			}
			if _, err := db.Exec(`CREATE TRIGGER fail_restored_decoder_source BEFORE INSERT ON decoder_cv_snapshot_sources
BEGIN SELECT RAISE(ABORT,'synthetic restored source failure'); END`); err != nil {
				t.Fatal(err)
			}
			if _, err := backup.Import(t.Context(), doc); err == nil {
				t.Fatal("injected restored-source failure was ignored")
			}
			failed := decoderSnapshots(t, backup)
			if !reflect.DeepEqual(catalogBackupTableSnapshot(t, before), catalogBackupTableSnapshot(t, failed)) ||
				!reflect.DeepEqual(before.Files, failed.Files) {
				t.Fatal("failed adoption changed target inventory, model records, blobs or files")
			}
			if data, err := os.ReadFile(path); err != nil || !reflect.DeepEqual(data, oldContent) {
				t.Fatalf("failure replaced the old target file: %v", err)
			}
			if _, err := db.Exec(`DROP TRIGGER fail_restored_decoder_source`); err != nil {
				t.Fatal(err)
			}
			if _, err := backup.Import(t.Context(), doc); err != nil {
				t.Fatal(err)
			}
			after := decoderSnapshots(t, backup)
			for _, table := range []string{"vehicles", "vehicle_cv_values", "vehicle_cv_value_history", "vehicle_cv_files"} {
				if !reflect.DeepEqual(after.Tables[table], doc.Tables[table]) {
					t.Fatalf("restore changed original IDs, values, raw profiles or provenance in %s", table)
				}
			}
			if len(after.Tables["decoder_instances"]) != 0 || len(after.Tables["decoder_installations"]) != 0 ||
				len(after.Tables["decoder_cv_workspaces"]) != 2 || len(after.Tables["decoder_cv_snapshots"]) != 2 {
				t.Fatal("restore inferred hardware or merged different raw profile labels")
			}
			for _, source := range after.Tables["decoder_cv_snapshot_sources"] {
				if available {
					if source["content_state"] != "stored" {
						t.Fatal("staged original was not frozen")
					}
					data, err := application.NewFileBlobService(db, root).Load(t.Context(), source["blob_id"].(string))
					if err != nil || !reflect.DeepEqual(data, restoredContent) {
						t.Fatalf("old target content was frozen instead of the backup: %v", err)
					}
				} else if source["content_state"] != "unavailable" || source["blob_id"] != nil {
					t.Fatal("missing backup original was replaced with unrelated target bytes")
				}
			}
			if available {
				if data, err := os.ReadFile(path); err != nil || !reflect.DeepEqual(data, restoredContent) {
					t.Fatalf("restored original file changed: %v", err)
				}
			} else if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
				t.Fatal("old target file survived a backup without that original")
			}
			if validation, err := backup.Validate(t.Context(), after); err != nil || !validation.Compatible {
				t.Fatalf("adopted legacy restore is not exportable/restorable: %#v / %v", validation, err)
			}
			if _, err := backup.Import(t.Context(), after); err != nil {
				t.Fatal(err)
			}
			repeated := decoderSnapshots(t, backup)
			if !reflect.DeepEqual(catalogBackupTableSnapshot(t, after), catalogBackupTableSnapshot(t, repeated)) {
				t.Fatal("restoring already captured evidence changed its identities or timestamps")
			}
		})
	}
}
