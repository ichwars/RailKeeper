package application_test

import (
	"database/sql"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"railkeeper/backend/internal/application"
)

func decoderConfigurationFixture(t *testing.T) (*sql.DB, *application.BackupService, *application.BackupDocument) {
	t.Helper()
	dir := t.TempDir()
	db := backupTestDB(t, dir)
	tx, err := db.BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := application.SeedDecoderLibrary(t.Context(), tx, "2026-11-07T01:00:00Z"); err != nil {
		_ = tx.Rollback()
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	vehicles := application.NewVehicleService(db)
	vehicle, err := vehicles.Create(t.Context(), application.CreateVehicleInput{
		Name: "Two decoder vehicle", Manufacturer: "Test", Gauge: "H0", Category: "Lokomotive", Gattung: "Diesellok",
	}, "actor")
	if err != nil {
		t.Fatal(err)
	}
	blobs := application.NewFileBlobService(db, dir)
	blob, err := blobs.Store(t.Context(), []byte("CV;Value\n1;3\n"))
	if err != nil {
		t.Fatal(err)
	}
	file, err := vehicles.CreateCVFile(t.Context(), vehicle.ID, application.VehicleCVFileInput{
		FileName: "original.csv", OriginalName: "original.csv", DecoderProfile: "Legacy", BlobID: blob,
		MimeType: "text/csv", SizeBytes: 13,
	})
	if err != nil {
		t.Fatal(err)
	}
	zero, line := 0, 2
	input := application.VehicleCVValueInput{CVNumber: 1, Value: 3, DecoderProfile: "Legacy",
		SourceFileID: file.ID, SourceLine: &line, SourceExportedAt: "2026-10-20T08:30:00"}
	first, err := vehicles.CreateCVValue(t.Context(), vehicle.ID, input)
	if err != nil {
		t.Fatal(err)
	}
	input.Value = 5
	if _, err := vehicles.UpdateCVValue(t.Context(), vehicle.ID, first.ID, input); err != nil {
		t.Fatal(err)
	}
	for _, indexed := range []bool{false, true} {
		input.CVNumber, input.Value = 257, 11
		if indexed {
			input.IndexCV31, input.IndexCV32 = &zero, &zero
		}
		if _, err := vehicles.CreateCVValue(t.Context(), vehicle.ID, input); err != nil {
			t.Fatal(err)
		}
	}
	// Preserve an old raw profile and an unknown historic value without normalizing them.
	if _, err := db.Exec(`UPDATE vehicle_cv_values SET decoder_profile=' Legacy ';
UPDATE vehicle_cv_values SET value=-1 WHERE cv_number=257 AND index_cv31 IS NULL`); err != nil {
		t.Fatal(err)
	}
	values, err := vehicles.ListCVValues(t.Context(), vehicle.ID)
	if err != nil {
		t.Fatal(err)
	}
	payload, _ := json.Marshal(application.DecoderCVSnapshotPayload{Values: values})
	metadata, _ := json.Marshal(file)
	hardware, _ := json.Marshal(application.DecoderSnapshotHardware{DecoderID: "decoder-b", DecoderLabel: "Sound",
		Manufacturer: "Test", Model: "Same model", InstallationID: "installation-b", InstallationState: "installed"})
	now := "2026-11-07T01:00:00Z"
	for _, query := range []struct {
		sql  string
		args []any
	}{
		{`INSERT INTO decoder_instances VALUES('decoder-a','Drive','Test','Same model','','',?,?),('decoder-b','Sound','Test','Same model','','',?,?)`, []any{now, now, now, now}},
		{`INSERT INTO decoder_installations VALUES('installation-a','decoder-a',?,?,?,'RK-1','installed',NULL,NULL,'',?,?),('installation-b','decoder-b',?,?,?,'RK-1','installed',NULL,NULL,'',?,?)`, []any{vehicle.ID, vehicle.ID, vehicle.Name, now, now, vehicle.ID, vehicle.ID, vehicle.Name, now, now}},
		{`INSERT INTO decoder_cv_workspaces VALUES('workspace-a',?,' Legacy ','legacy',NULL,?,?),('workspace-b',?,'Other','manual','installation-b',?,?)`, []any{vehicle.ID, now, now, vehicle.ID, now, now}},
		{`INSERT INTO decoder_cv_snapshots VALUES('snapshot-a','workspace-a',?,?,'RK-1',' Legacy ',NULL,'legacy','partial_working_state',?,'{}','Adopted, not measured',?)`, []any{vehicle.ID, vehicle.Name, string(payload), now}},
		{`INSERT INTO decoder_cv_snapshots VALUES('snapshot-b','workspace-b',?,?,'RK-1','Other','installation-b','checkpoint','partial_working_state','{"values":[]}',?,'Explicit hardware identity',?)`, []any{vehicle.ID, vehicle.Name, string(hardware), now}},
		{`INSERT INTO decoder_cv_snapshot_sources VALUES('snapshot-a',?,?,'stored',?)`, []any{file.ID, string(metadata), blob}},
	} {
		if _, err := db.Exec(query.sql, query.args...); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := vehicles.DeleteCVFile(t.Context(), vehicle.ID, file.ID); err != nil {
		t.Fatal(err)
	}
	if err := blobs.DeleteIfUnreferenced(t.Context(), blob); err != nil {
		t.Fatal(err)
	}
	if content, err := blobs.Load(t.Context(), blob); err != nil || string(content) != "CV;Value\n1;3\n" {
		t.Fatalf("frozen original was garbage collected: %q / %v", content, err)
	}
	backup := application.NewBackupService(db, dir)
	doc, err := backup.Export(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	return db, backup, doc
}

func TestDecoderConfigurationBackupRoundTripPreservesUnknownHardwareIndicesSourcesAndHistory(t *testing.T) {
	db, service, doc := decoderConfigurationFixture(t)
	if doc.Version != 27 {
		t.Fatalf("version=%d", doc.Version)
	}
	for _, query := range []string{
		`UPDATE decoder_cv_snapshots SET note='Changed'`,
		`UPDATE decoder_cv_snapshot_sources SET metadata_json='{}'`,
		`INSERT INTO decoder_installations SELECT 'duplicate',decoder_id,vehicle_id,vehicle_reference_id,vehicle_name,vehicle_inventory_number,state,installed_at,removed_at,notes,recorded_at,updated_at FROM decoder_installations WHERE id='installation-a'`,
	} {
		if _, err := db.Exec(query); err == nil {
			t.Fatalf("unsafe mutation accepted: %s", query)
		}
	}
	validation, err := service.Validate(t.Context(), doc)
	if err != nil || !validation.Compatible {
		t.Fatalf("validation: %#v / %v", validation, err)
	}
	if _, err := service.Import(t.Context(), doc); err != nil {
		t.Fatal(err)
	}
	after, err := service.Export(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(catalogBackupTableSnapshot(t, doc), catalogBackupTableSnapshot(t, after)) || !reflect.DeepEqual(doc.Files, after.Files) {
		t.Fatal("decoder or existing inventory data changed across restore")
	}
	var payload application.DecoderCVSnapshotPayload
	if err := json.Unmarshal([]byte(after.Tables["decoder_cv_snapshots"][0]["payload_json"].(string)), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Values) != 3 || len(payload.Values[0].History) != 1 || payload.Values[1].IndexCV31 != nil ||
		payload.Values[2].IndexCV31 == nil || *payload.Values[2].IndexCV31 != 0 || payload.Values[1].Value != -1 {
		t.Fatalf("CV evidence lost: %#v", payload)
	}
}

func TestDecoderConfigurationHistorySurvivesVehicleDeletionAndRestore(t *testing.T) {
	db, service, doc := decoderConfigurationFixture(t)
	vehicle := doc.Tables["vehicles"][0]["id"].(string)
	if err := application.NewVehicleService(db).Delete(t.Context(), vehicle, "actor"); err != nil {
		t.Fatal(err)
	}
	after, err := service.Export(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(after.Tables["decoder_cv_workspaces"]) != 0 || len(after.Tables["decoder_cv_snapshots"]) != 2 ||
		!reflect.DeepEqual(after.Tables["decoder_cv_snapshots"], doc.Tables["decoder_cv_snapshots"]) {
		t.Fatal("deleting current inventory changed historical snapshots")
	}
	for _, row := range after.Tables["decoder_installations"] {
		if row["vehicle_id"] != nil || row["vehicle_reference_id"] != vehicle || row["state"] != "unresolved" || row["removed_at"] != nil {
			t.Fatalf("deletion invented a removal date or lost ownership history: %#v", row)
		}
	}
	if validation, err := service.Validate(t.Context(), after); err != nil || !validation.Compatible {
		t.Fatalf("historical backup: %#v / %v", validation, err)
	}
	if _, err := service.Import(t.Context(), after); err != nil {
		t.Fatal(err)
	}
}

func TestDecoderConfigurationFrozenHardwareSurvivesRemovalAndModelCorrection(t *testing.T) {
	db, service, before := decoderConfigurationFixture(t)
	if _, err := db.Exec(`UPDATE decoder_instances SET model='Corrected model', label='New label' WHERE id='decoder-b';
UPDATE decoder_installations SET state='removed', removed_at=NULL WHERE id='installation-b'`); err != nil {
		t.Fatal(err)
	}
	after, err := service.Export(t.Context())
	if err != nil || !reflect.DeepEqual(before.Tables["decoder_cv_snapshots"], after.Tables["decoder_cv_snapshots"]) {
		t.Fatalf("hardware change mutated a recorded snapshot: %v", err)
	}
	if validation, err := service.Validate(t.Context(), after); err != nil || !validation.Compatible {
		t.Fatalf("captured historical hardware: %#v / %v", validation, err)
	}
	if _, err := service.Import(t.Context(), after); err != nil {
		t.Fatal(err)
	}
}

func TestDecoderConfigurationUnavailableOriginalRetainsUnknownOrDifferentSourceOwnership(t *testing.T) {
	_, service, doc := decoderConfigurationFixture(t)
	source := doc.Tables["decoder_cv_snapshot_sources"][0]
	source["content_state"], source["blob_id"] = "unavailable", nil
	for _, owner := range []string{"", "historical-other-vehicle"} {
		metadata, _ := json.Marshal(map[string]string{"id": source["source_file_id"].(string), "vehicleId": owner})
		source["metadata_json"] = string(metadata)
		validation, err := service.Validate(t.Context(), doc)
		if err != nil || !validation.Compatible {
			t.Fatalf("unavailable original reference was discarded: %#v / %v", validation, err)
		}
		if _, err := service.Import(t.Context(), doc); err != nil {
			t.Fatal(err)
		}
	}
}

func TestDecoderConfigurationPreflightRejectsMalformedEvidenceBeforeChangingAnything(t *testing.T) {
	_, service, original := decoderConfigurationFixture(t)
	for name, mutate := range map[string]func(*application.BackupDocument){
		"old version with new data": func(d *application.BackupDocument) { d.Version = 26 },
		"missing table":             func(d *application.BackupDocument) { delete(d.Tables, "decoder_instances") },
		"unknown column":            func(d *application.BackupDocument) { d.Tables["decoder_instances"][0]["future"] = "keep" },
		"missing decoder":           func(d *application.BackupDocument) { d.Tables["decoder_installations"][0]["decoder_id"] = "missing" },
		"duplicate active decoder": func(d *application.BackupDocument) {
			d.Tables["decoder_installations"][1]["decoder_id"] = d.Tables["decoder_installations"][0]["decoder_id"]
		},
		"invalid calendar date": func(d *application.BackupDocument) {
			d.Tables["decoder_installations"][0]["installed_at"] = "2026-02-30"
		},
		"missing unknown date": func(d *application.BackupDocument) { delete(d.Tables["decoder_installations"][0], "installed_at") },
		"invented full measurement": func(d *application.BackupDocument) {
			d.Tables["decoder_cv_snapshots"][0]["scope"] = "complete_measurement"
		},
		"null payload": func(d *application.BackupDocument) { d.Tables["decoder_cv_snapshots"][0]["payload_json"] = "null" },
		"unknown payload field": func(d *application.BackupDocument) {
			d.Tables["decoder_cv_snapshots"][0]["payload_json"] = `{"values":[],"unrecognizedEvidence":true}`
		},
		"missing value identity": func(d *application.BackupDocument) {
			d.Tables["decoder_cv_snapshots"][0]["payload_json"] = `{"values":[{}]}`
		},
		"unconfirmed hardware": func(d *application.BackupDocument) {
			d.Tables["decoder_cv_snapshots"][0]["hardware_json"] = `{"decoderId":"guess"}`
		},
		"missing frozen blob": func(d *application.BackupDocument) { d.Tables["decoder_cv_snapshot_sources"][0]["blob_id"] = "missing" },
		"wrong source metadata": func(d *application.BackupDocument) {
			d.Tables["decoder_cv_snapshot_sources"][0]["metadata_json"] = `{"id":"wrong"}`
		},
	} {
		t.Run(name, func(t *testing.T) {
			raw, _ := json.Marshal(original)
			var doc application.BackupDocument
			if err := json.Unmarshal(raw, &doc); err != nil {
				t.Fatal(err)
			}
			mutate(&doc)
			validation, err := service.Validate(t.Context(), &doc)
			if err != nil || validation.Compatible {
				t.Fatalf("invalid evidence accepted: %#v / %v", validation, err)
			}
			if _, err := service.Import(t.Context(), &doc); !errors.Is(err, application.ErrBackupInvalid) {
				t.Fatalf("restore returned %v", err)
			}
			after, err := service.Export(t.Context())
			if err != nil || !reflect.DeepEqual(catalogBackupTableSnapshot(t, after), catalogBackupTableSnapshot(t, original)) {
				t.Fatalf("preflight changed current data: %v", err)
			}
		})
	}
}
