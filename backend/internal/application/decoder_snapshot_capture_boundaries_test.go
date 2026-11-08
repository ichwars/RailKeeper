package application_test

import (
	"encoding/json"
	"errors"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestDecoderConfigurationLegacyCaptureKeepsUnreadableArchivedOriginalUnresolved(t *testing.T) {
	db, _, doc := decoderConfigurationFixture(t)
	db.SetMaxOpenConns(1)
	originalBlob := doc.Tables["decoder_cv_snapshot_sources"][0]["blob_id"].(string)
	sourceID := doc.Tables["decoder_cv_snapshot_sources"][0]["source_file_id"].(string)
	blobs := application.NewFileBlobService(db, "")
	content, err := blobs.Load(t.Context(), originalBlob)
	if err != nil {
		t.Fatal(err)
	}
	secondBlob, err := blobs.Store(t.Context(), content)
	if err != nil {
		t.Fatal(err)
	}
	for _, statement := range []string{
		`INSERT INTO decoder_cv_snapshots SELECT 'archive-a',workspace_id,vehicle_reference_id,vehicle_name,vehicle_inventory_number,profile_text,installation_id,'checkpoint',scope,payload_json,hardware_json,note,recorded_at FROM decoder_cv_snapshots WHERE id='snapshot-a'`,
		`INSERT INTO decoder_cv_snapshot_sources SELECT 'archive-a',source_file_id,metadata_json,content_state,blob_id FROM decoder_cv_snapshot_sources WHERE snapshot_id='snapshot-a'`,
		`DELETE FROM decoder_cv_snapshot_sources WHERE snapshot_id='snapshot-a'`,
		`DELETE FROM decoder_cv_snapshots WHERE id='snapshot-a'`,
		`INSERT INTO decoder_cv_snapshots SELECT 'archive-other',workspace_id,vehicle_reference_id,vehicle_name,vehicle_inventory_number,profile_text,installation_id,kind,scope,payload_json,hardware_json,note,recorded_at FROM decoder_cv_snapshots WHERE id='archive-a'`,
	} {
		if _, err := db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec(`INSERT INTO decoder_cv_snapshot_sources SELECT 'archive-other',source_file_id,metadata_json,'stored',? FROM decoder_cv_snapshot_sources WHERE snapshot_id='archive-a'`, secondBlob); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE file_blobs SET compression='unsupported' WHERE id=?`, secondBlob); err != nil {
		t.Fatal(err)
	}
	before := databaseRowFingerprints(t, db)
	if err := application.BackfillLegacyDecoderCV(t.Context(), db, ""); err != nil {
		t.Fatal(err)
	}
	var state string
	var blob any
	if err := db.QueryRow(`SELECT s.content_state,s.blob_id FROM decoder_cv_snapshot_sources s
JOIN decoder_cv_snapshots p ON p.id=s.snapshot_id WHERE p.kind='legacy' AND s.source_file_id=?`, sourceID).Scan(&state, &blob); err != nil {
		t.Fatal(err)
	}
	if state != "unavailable" || blob != nil {
		t.Fatal("unreadable archived variant was silently selected", state)
	}
	after := databaseRowFingerprints(t, db)
	for _, table := range []string{"vehicle_cv_values", "vehicle_cv_value_history", "vehicle_cv_files", "file_blobs", "decoder_instances", "decoder_installations"} {
		if before[table] != after[table] {
			t.Fatal("capture rewrote original evidence", table)
		}
	}
}
func TestDecoderConfigurationLegacyCapturePreservesForeignReferenceWithoutAdoptingItsBytes(t *testing.T) {
	db, _, _, vehicleID := legacyDecoderCaptureFixture(t)
	db.SetMaxOpenConns(1)
	service := application.NewVehicleService(db)
	other, err := service.Create(t.Context(), mutationInput(), "actor")
	if err != nil {
		t.Fatal(err)
	}
	blob, err := application.NewFileBlobService(db, "").Store(t.Context(), []byte("another vehicles original"))
	if err != nil {
		t.Fatal(err)
	}
	foreign, err := service.CreateCVFile(t.Context(), other.ID, application.VehicleCVFileInput{
		FileName: "foreign.csv", OriginalName: "foreign.csv", DecoderProfile: "other profile", BlobID: blob, MimeType: "text/csv", SizeBytes: 25,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE vehicle_cv_values SET source_file_id=? WHERE vehicle_id=?`, foreign.ID, vehicleID); err != nil {
		t.Fatal(err)
	}
	before := databaseRowFingerprints(t, db)
	if err := application.BackfillLegacyDecoderCV(t.Context(), db, ""); err != nil {
		t.Fatal(err)
	}
	var snapshotID, state, metadata string
	var adoptedBlob any
	if err := db.QueryRow(`SELECT p.id,s.content_state,s.metadata_json,s.blob_id FROM decoder_cv_snapshot_sources s
JOIN decoder_cv_snapshots p ON p.id=s.snapshot_id WHERE p.vehicle_reference_id=? AND s.source_file_id=?`, vehicleID, foreign.ID).Scan(&snapshotID, &state, &metadata, &adoptedBlob); err != nil {
		t.Fatal(err)
	}
	var recorded application.VehicleCVFile
	if err := json.Unmarshal([]byte(metadata), &recorded); err != nil {
		t.Fatal(err)
	}
	if state != "unavailable" || adoptedBlob != nil || recorded.VehicleID != other.ID || recorded.ID != foreign.ID {
		t.Fatal("foreign legacy source was relabeled or adopted", state, recorded.ID, recorded.VehicleID)
	}
	if file, data, err := service.DecoderCVSnapshotSourceContent(t.Context(), snapshotID, foreign.ID); file != nil || data != nil || !errors.Is(err, application.ErrDecoderSourceUnavailable) {
		t.Fatal("foreign source bytes exposed as local evidence", err)
	}
	after := databaseRowFingerprints(t, db)
	for _, table := range []string{"vehicle_cv_values", "vehicle_cv_value_history", "vehicle_cv_files", "file_blobs"} {
		if before[table] != after[table] {
			t.Fatal("foreign source preservation changed evidence", table)
		}
	}
}
