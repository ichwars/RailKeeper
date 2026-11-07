package application_test

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"railkeeper/backend/internal/application"
)

func legacyDecoderCaptureFixture(t *testing.T) (*sql.DB, *application.BackupService, string, string) {
	t.Helper()
	db, backup, original := decoderConfigurationFixture(t)
	if _, err := db.Exec(`DELETE FROM decoder_cv_snapshot_sources; DELETE FROM decoder_cv_snapshots;
DELETE FROM decoder_cv_workspaces; UPDATE decoder_instances SET label=' Legacy ' WHERE id='decoder-a'`); err != nil {
		t.Fatal(err)
	}
	var file string
	if err := db.QueryRow(`SELECT file FROM pragma_database_list WHERE name='main'`).Scan(&file); err != nil {
		t.Fatal(err)
	}
	return db, backup, filepath.Dir(file), original.Tables["vehicles"][0]["id"].(string)
}

func decoderSnapshots(t *testing.T, backup *application.BackupService) *application.BackupDocument {
	t.Helper()
	doc, err := backup.Export(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	return doc
}

func TestDecoderConfigurationLegacyCapturePreservesCVsAndNeverInfersMatchingHardware(t *testing.T) {
	db, backup, root, vehicle := legacyDecoderCaptureFixture(t)
	before := decoderSnapshots(t, backup)
	values, err := application.NewVehicleService(db).ListCVValues(t.Context(), vehicle)
	if err != nil {
		t.Fatal(err)
	}
	for i := range values {
		values[i].Revision = ""
	}
	sort.Slice(values, func(i, j int) bool { return values[i].ID < values[j].ID })
	if err := application.BackfillLegacyDecoderCV(t.Context(), db, root); err != nil {
		t.Fatal(err)
	}
	after := decoderSnapshots(t, backup)
	if len(after.Tables["decoder_cv_workspaces"]) != 1 || len(after.Tables["decoder_cv_snapshots"]) != 1 {
		t.Fatal("legacy group was missing or split")
	}
	workspace, snapshot := after.Tables["decoder_cv_workspaces"][0], after.Tables["decoder_cv_snapshots"][0]
	if workspace["profile_text"] != " Legacy " || workspace["origin"] != "legacy" || workspace["installation_id"] != nil ||
		snapshot["hardware_json"] != "{}" || snapshot["scope"] != "partial_working_state" {
		t.Fatal("raw profile or unknown hardware meaning changed")
	}
	var payload application.DecoderCVSnapshotPayload
	expected, err := json.Marshal(application.DecoderCVSnapshotPayload{Values: values})
	if err != nil {
		t.Fatal(err)
	}
	// Empty history lists are omitted by the established public CV representation.
	if err := json.Unmarshal([]byte(snapshot["payload_json"].(string)), &payload); err != nil || string(expected) != snapshot["payload_json"] {
		t.Fatalf("CV IDs, indices, unknown values, provenance or history changed: %v", err)
	}
	for table, rows := range before.Tables {
		if table == "decoder_cv_workspaces" || table == "decoder_cv_snapshots" || table == "decoder_cv_snapshot_sources" {
			continue
		}
		if !reflect.DeepEqual(rows, after.Tables[table]) {
			t.Fatalf("legacy capture changed existing table %s", table)
		}
	}
	if source := after.Tables["decoder_cv_snapshot_sources"][0]; source["content_state"] != "unavailable" || source["blob_id"] != nil {
		t.Fatal("missing original bytes were invented")
	}
	if err := application.BackfillLegacyDecoderCV(t.Context(), db, root); err != nil {
		t.Fatal(err)
	}
	repeated := decoderSnapshots(t, backup)
	if !reflect.DeepEqual(catalogBackupTableSnapshot(t, after), catalogBackupTableSnapshot(t, repeated)) {
		t.Fatal("repeated adoption changed timestamps, identities or evidence")
	}
	if validation, err := backup.Validate(t.Context(), after); err != nil || !validation.Compatible {
		t.Fatalf("captured unknown evidence is not restorable: %#v / %v", validation, err)
	}
}

func TestDecoderConfigurationLegacyFilesystemCaptureRollsBackBlobAndSnapshotTogether(t *testing.T) {
	db, backup, root, vehicle := legacyDecoderCaptureFixture(t)
	content := []byte("CV;Value\n1;42\n")
	if err := os.MkdirAll(filepath.Join(root, "uploads"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "uploads", "legacy.csv"), content, 0600); err != nil {
		t.Fatal(err)
	}
	file, err := application.NewVehicleService(db).CreateCVFile(t.Context(), vehicle, application.VehicleCVFileInput{
		FileName: "legacy.csv", OriginalName: "legacy.csv", StoragePath: "uploads/legacy.csv", DecoderProfile: "Legacy",
		MimeType: "text/csv", SizeBytes: int64(len(content)),
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE vehicle_cv_files SET decoder_profile=' Legacy ' WHERE id=?;
UPDATE vehicle_cv_values SET source_file_id=?`, file.ID, file.ID); err != nil {
		t.Fatal(err)
	}
	before := decoderSnapshots(t, backup)
	if _, err := db.Exec(`CREATE TRIGGER fail_decoder_source BEFORE INSERT ON decoder_cv_snapshot_sources
BEGIN SELECT RAISE(ABORT,'synthetic source failure'); END`); err != nil {
		t.Fatal(err)
	}
	if err := application.BackfillLegacyDecoderCV(t.Context(), db, root); err == nil {
		t.Fatal("injected source error was ignored")
	}
	failed := decoderSnapshots(t, backup)
	if !reflect.DeepEqual(catalogBackupTableSnapshot(t, before), catalogBackupTableSnapshot(t, failed)) ||
		!reflect.DeepEqual(before.Files, failed.Files) {
		t.Fatal("failed capture left partial workspaces, snapshots, blobs or file changes")
	}
	if _, err := db.Exec(`DROP TRIGGER fail_decoder_source`); err != nil {
		t.Fatal(err)
	}
	if err := application.BackfillLegacyDecoderCV(t.Context(), db, root); err != nil {
		t.Fatal(err)
	}
	after := decoderSnapshots(t, backup)
	if !reflect.DeepEqual(before.Tables["vehicle_cv_files"], after.Tables["vehicle_cv_files"]) ||
		len(after.Tables["file_blobs"]) != len(before.Tables["file_blobs"])+1 {
		t.Fatal("capture modified the original file record or failed to store its exact bytes")
	}
	source := after.Tables["decoder_cv_snapshot_sources"][0]
	if source["content_state"] != "stored" {
		t.Fatal("available original was marked missing")
	}
	blobs := application.NewFileBlobService(db, root)
	stored, err := blobs.Load(t.Context(), source["blob_id"].(string))
	if err != nil || !reflect.DeepEqual(stored, content) {
		t.Fatalf("source bytes changed: %v", err)
	}
	if _, err := application.NewVehicleService(db).DeleteCVFile(t.Context(), vehicle, file.ID); err != nil {
		t.Fatal(err)
	}
	if err := blobs.DeleteIfUnreferenced(t.Context(), source["blob_id"].(string)); err != nil {
		t.Fatal(err)
	}
	if stored, err := blobs.Load(t.Context(), source["blob_id"].(string)); err != nil || !reflect.DeepEqual(stored, content) {
		t.Fatalf("original deletion lost frozen bytes: %v", err)
	}
}

func TestDecoderConfigurationLegacyCaptureReusesOnlyUnambiguousArchivedSources(t *testing.T) {
	for _, scenario := range []struct {
		name                 string
		duplicate, different bool
	}{{name: "known original"}, {name: "same bytes, different blob identities", duplicate: true},
		{name: "contradictory originals", duplicate: true, different: true}} {
		t.Run(scenario.name, func(t *testing.T) {
			db, backup, doc := decoderConfigurationFixture(t)
			for _, query := range []string{
				`INSERT INTO decoder_cv_snapshots SELECT 'archive-a',workspace_id,vehicle_reference_id,vehicle_name,vehicle_inventory_number,profile_text,installation_id,'checkpoint',scope,payload_json,hardware_json,note,recorded_at FROM decoder_cv_snapshots WHERE id='snapshot-a'`,
				`INSERT INTO decoder_cv_snapshot_sources SELECT 'archive-a',source_file_id,metadata_json,content_state,blob_id FROM decoder_cv_snapshot_sources WHERE snapshot_id='snapshot-a'`,
				`DELETE FROM decoder_cv_snapshot_sources WHERE snapshot_id='snapshot-a'`,
				`DELETE FROM decoder_cv_snapshots WHERE id='snapshot-a'`,
			} {
				if _, err := db.Exec(query); err != nil {
					t.Fatal(err)
				}
			}
			if scenario.duplicate {
				blobs := application.NewFileBlobService(db, "")
				content, err := blobs.Load(t.Context(), doc.Tables["decoder_cv_snapshot_sources"][0]["blob_id"].(string))
				if err != nil {
					t.Fatal(err)
				}
				if scenario.different {
					content = []byte("Different original")
				}
				blob, err := blobs.Store(t.Context(), content)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := db.Exec(`INSERT INTO decoder_cv_snapshots SELECT 'archive-other',workspace_id,vehicle_reference_id,vehicle_name,vehicle_inventory_number,profile_text,installation_id,kind,scope,payload_json,hardware_json,note,recorded_at FROM decoder_cv_snapshots WHERE id='archive-a';
INSERT INTO decoder_cv_snapshot_sources SELECT 'archive-other',source_file_id,metadata_json,'stored',? FROM decoder_cv_snapshot_sources WHERE snapshot_id='archive-a'`, blob); err != nil {
					t.Fatal(err)
				}
			}
			if err := application.BackfillLegacyDecoderCV(t.Context(), db, ""); err != nil {
				t.Fatal(err)
			}
			after := decoderSnapshots(t, backup)
			var created string
			for _, snapshot := range after.Tables["decoder_cv_snapshots"] {
				if snapshot["kind"] == "legacy" {
					created = snapshot["id"].(string)
				}
			}
			if created == "" {
				t.Fatal("missing legacy capture")
			}
			for _, source := range after.Tables["decoder_cv_snapshot_sources"] {
				if source["snapshot_id"] != created {
					continue
				}
				if scenario.different && (source["content_state"] != "unavailable" || source["blob_id"] != nil) {
					t.Fatal("contradictory original was silently selected")
				}
				if !scenario.different && source["content_state"] != "stored" {
					t.Fatal("known immutable original was lost")
				}
			}
		})
	}
}
