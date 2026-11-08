package vehicledata_test

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"railkeeper/backend/internal/decoder"
	"railkeeper/backend/internal/decoder/workflow"
	"railkeeper/backend/internal/infrastructure/vehicledata"
)

func historicalSnapshotFixture(t *testing.T) (*sql.DB, *workflow.SnapshotReader) {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })
	for _, statement := range []string{
		`CREATE TABLE decoder_cv_snapshots(id TEXT PRIMARY KEY,workspace_id TEXT,vehicle_reference_id TEXT,
vehicle_name TEXT,vehicle_inventory_number TEXT,profile_text TEXT,installation_id TEXT,kind TEXT,scope TEXT,
note TEXT,recorded_at TEXT,payload_json TEXT,hardware_json TEXT)`,
		`CREATE TABLE decoder_cv_snapshot_sources(snapshot_id TEXT,source_file_id TEXT,metadata_json TEXT,content_state TEXT,blob_id TEXT)`,
		`CREATE TABLE file_blobs(id TEXT PRIMARY KEY,compression TEXT,data BLOB)`,
		`INSERT INTO decoder_cv_snapshots VALUES('snapshot','original-workspace','removed-vehicle','Original vehicle','KEEP-1',
'  original profile  ',NULL,'legacy','partial_working_state','recorded note','2001-02-03T04:05:06Z',
'{"values":[{"id":"absent-page","vehicleId":"removed-vehicle","cvNumber":257,"value":0,"decoderProfile":"  original profile  "},
{"id":"zero-page","vehicleId":"removed-vehicle","cvNumber":257,"indexCV31":0,"indexCV32":0,"value":255,
"history":[{"id":"original-history","cvValueId":"zero-page","vehicleId":"removed-vehicle","oldValue":1,"newValue":255,"changedAt":"original time"}]}]}','{}')`,
		`INSERT INTO decoder_cv_snapshot_sources VALUES('snapshot','source',
'{"id":"source","vehicleId":"removed-vehicle","fileName":"original.csv","originalName":"  original name  ","decoderProfile":"  original profile  ","sizeBytes":4}',
'stored','blob')`,
	} {
		if _, err := db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec(`INSERT INTO file_blobs VALUES('blob','',?)`, []byte{0, 255, 13, 10}); err != nil {
		t.Fatal(err)
	}
	return db, workflow.NewSnapshotReader(vehicledata.NewSnapshotStore(db))
}
func TestHistoricalDecoderSnapshotReadsKeepExactEvidenceWithoutCurrentVehicleTables(t *testing.T) {
	_, reader := historicalSnapshotFixture(t)
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	snapshot, err := reader.Snapshot(ctx, "snapshot")
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Profile != "  original profile  " || snapshot.VehicleReferenceID != "removed-vehicle" ||
		snapshot.RecordedAt != "2001-02-03T04:05:06Z" || snapshot.Hardware != nil || snapshot.InstallationID != nil || snapshot.ValueCount != 2 {
		t.Fatal("historical identity, unknown hardware or recording time changed", snapshot)
	}
	first, second := snapshot.Payload.Values[0], snapshot.Payload.Values[1]
	if first.IndexCV31 != nil || first.IndexCV32 != nil || second.IndexCV31 == nil || second.IndexCV32 == nil ||
		*second.IndexCV31 != 0 || *second.IndexCV32 != 0 || first.DecoderProfile != "  original profile  " ||
		len(second.History) != 1 || second.History[0].ChangedAt != "original time" {
		t.Fatal("absent/zero index page or historical profile/history changed")
	}
	if len(snapshot.Sources) != 1 || snapshot.Sources[0].Metadata.OriginalName != "  original name  " {
		t.Fatal("source metadata normalized")
	}
	file, content, err := reader.SourceContent(ctx, "snapshot", "source")
	if err != nil || file.DecoderProfile != "  original profile  " || !bytes.Equal(content, []byte{0, 255, 13, 10}) {
		t.Fatal("original source bytes or labels changed", err)
	}
	for _, id := range []string{"", "removed-vehicle"} {
		list, err := reader.List(ctx, id)
		if err != nil || len(list) != 1 || list[0].ValueCount != 2 || list[0].Profile != snapshot.Profile {
			t.Fatal("historical list depends on current inventory", id, list, err)
		}
	}
	if list, err := reader.List(ctx, " removed-vehicle "); err != nil || list == nil || len(list) != 0 {
		t.Fatal("list silently changed caller identity", list, err)
	}
}
func TestHistoricalDecoderSnapshotReadFailuresRetainMissingVersusUnavailable(t *testing.T) {
	for _, test := range []struct {
		name, sql string
		missing   bool
	}{
		{"missing source", `DELETE FROM decoder_cv_snapshot_sources`, true},
		{"missing blob", `DELETE FROM file_blobs`, false},
		{"unavailable content before malformed metadata", `UPDATE decoder_cv_snapshot_sources SET content_state='unavailable',metadata_json='{'`, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			db, reader := historicalSnapshotFixture(t)
			if _, err := db.Exec(test.sql); err != nil {
				t.Fatal(err)
			}
			file, data, err := reader.SourceContent(t.Context(), "snapshot", "source")
			want := workflow.ErrDecoderSourceUnavailable
			if test.missing {
				want = decoder.ErrNotFound
			}
			if file != nil || data != nil || !errors.Is(err, want) {
				t.Fatal("wrong historical source result", file, data, err)
			}
		})
	}
}
func TestHistoricalDecoderSnapshotRejectsMalformedEvidenceAndReleasesReadTransaction(t *testing.T) {
	for _, test := range []struct{ name, sql, context string }{
		{"payload", `UPDATE decoder_cv_snapshots SET payload_json='{'`, ""},
		{"hardware", `UPDATE decoder_cv_snapshots SET hardware_json='['`, ""},
		{"source metadata", `UPDATE decoder_cv_snapshot_sources SET metadata_json='{'`, "scan decoder configuration"},
	} {
		t.Run(test.name, func(t *testing.T) {
			db, reader := historicalSnapshotFixture(t)
			if _, err := db.Exec(test.sql); err != nil {
				t.Fatal(err)
			}
			result, err := reader.Snapshot(t.Context(), "snapshot")
			if result != nil || err == nil || (test.context != "" && !strings.Contains(err.Error(), test.context)) {
				t.Fatal("malformed historical evidence acknowledged", result, err)
			}
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()
			var rows int
			if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM decoder_cv_snapshots`).Scan(&rows); err != nil || rows != 1 {
				t.Fatal("failed read retained transaction or altered stored evidence", rows, err)
			}
		})
	}
}
