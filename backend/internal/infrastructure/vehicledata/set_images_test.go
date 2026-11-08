package vehicledata_test

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"testing"

	"railkeeper/backend/internal/infrastructure/vehicledata"
	vehiclemodel "railkeeper/backend/internal/inventory/vehicle"
)

func setImageFixture(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })
	for _, statement := range []string{
		`CREATE TABLE vehicle_sets(id TEXT PRIMARY KEY,main_image_mode TEXT,main_member_image_id TEXT,
set_image_file_name TEXT,set_image_mime_type TEXT,set_image_blob_id TEXT,set_image_thumbnail_blob_id TEXT,
set_image_created_at TEXT,set_image_updated_at TEXT,updated_at TEXT)`,
		`CREATE TABLE audit_logs(id TEXT,actor_user_id TEXT,action TEXT,target_type TEXT,target_id TEXT,created_at TEXT,details_json TEXT)`,
		`INSERT INTO vehicle_sets VALUES('set','member','selected-member','original.png','image/png',
'original-blob','original-blob','recorded creation','recorded update','recorded row update')`,
	} {
		if _, err := db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	return db
}

func setImageRow(t *testing.T, db *sql.DB) []any {
	t.Helper()
	values := make([]any, 10)
	targets := make([]any, 10)
	for i := range values {
		targets[i] = &values[i]
	}
	if err := db.QueryRow(`SELECT * FROM vehicle_sets WHERE id='set'`).Scan(targets...); err != nil {
		t.Fatal(err)
	}
	return values
}

func TestSetImageAuditFailureRollsBackUploadAndDeletionBeforeReturningBlobIDs(t *testing.T) {
	for _, operation := range []string{"upload", "delete"} {
		t.Run(operation, func(t *testing.T) {
			db := setImageFixture(t)
			before := setImageRow(t, db)
			if _, err := db.Exec(`CREATE TRIGGER reject_set_image_audit BEFORE INSERT ON audit_logs
BEGIN SELECT RAISE(ABORT,'synthetic audit failure'); END`); err != nil {
				t.Fatal(err)
			}
			readbacks := 0
			service := vehiclemodel.NewSetImageService(vehicledata.NewSetImageStore(db), func(context.Context, string) (*vehiclemodel.VehicleSet, error) {
				readbacks++
				return nil, errors.New("unexpected readback")
			})
			var retired []string
			var err error
			if operation == "upload" {
				_, retired, err = service.UpsertSetImage(t.Context(), "set", vehiclemodel.VehicleSetImageInput{FileName: "new.png", MimeType: "image/png", BlobID: "new-blob"}, "actor")
			} else {
				retired, err = service.DeleteSetImage(t.Context(), "set", "actor")
			}
			if err == nil || retired != nil || readbacks != 0 || !reflect.DeepEqual(before, setImageRow(t, db)) {
				t.Fatal("failed audit changed metadata or exposed blobs for cleanup", retired, readbacks, err)
			}
		})
	}
}

func TestSetImageCommittedUploadRetainsOriginalCreationAndRetiredIDsAfterReadbackFailure(t *testing.T) {
	db := setImageFixture(t)
	failure := errors.New("postcommit set read unavailable")
	service := vehiclemodel.NewSetImageService(vehicledata.NewSetImageStore(db), func(context.Context, string) (*vehiclemodel.VehicleSet, error) { return nil, failure })
	set, retired, err := service.UpsertSetImage(t.Context(), "set", vehiclemodel.VehicleSetImageInput{
		FileName: "new.png", MimeType: "image/png", BlobID: "new-blob", ThumbnailBlobID: "new-thumbnail",
	}, "actor")
	if set != nil || !errors.Is(err, failure) || !reflect.DeepEqual(retired, []string{"original-blob"}) {
		t.Fatal("committed receipt lost old blob identities", set, retired, err)
	}
	row := setImageRow(t, db)
	if row[1] != "dedicated" || row[2] != nil || row[5] != "new-blob" || row[6] != "new-thumbnail" || row[7] != "recorded creation" {
		t.Fatal("postcommit failure lost new files or original capture time", row)
	}
	var audits int
	if err := db.QueryRow(`SELECT COUNT(*) FROM audit_logs WHERE action='VehicleSetImageUploaded'`).Scan(&audits); err != nil || audits != 1 {
		t.Fatal("committed upload lost audit", audits, err)
	}
}
