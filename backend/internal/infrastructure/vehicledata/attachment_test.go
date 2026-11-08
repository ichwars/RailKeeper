package vehicledata_test

import (
	"database/sql"
	"errors"
	"reflect"
	"testing"

	"railkeeper/backend/internal/infrastructure/vehicledata"
	vehiclemodel "railkeeper/backend/internal/inventory/vehicle"

	_ "modernc.org/sqlite"
)

func attachmentFixture(t *testing.T) (*sql.DB, vehiclemodel.AttachmentRepository, vehiclemodel.VehicleAttachment) {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })
	if _, err := db.Exec(`CREATE TABLE vehicle_attachments(
id TEXT PRIMARY KEY,vehicle_id TEXT,file_name TEXT,original_name TEXT,description TEXT,category TEXT,
mime_type TEXT,size_bytes INTEGER,storage_path TEXT,blob_id TEXT,maintenance_id TEXT,created_at TEXT,updated_at TEXT)`); err != nil {
		t.Fatal(err)
	}
	repository := vehicledata.NewAttachmentStore(db)
	value := vehiclemodel.VehicleAttachment{ID: "document", VehicleID: "owner", FileName: "manual.pdf", OriginalName: "Original manual.pdf",
		Description: "recorded description", Category: "manual", MimeType: "application/pdf", SizeBytes: 123,
		StoragePath: "private/original.pdf", BlobID: "original-blob", CreatedAt: "recorded", UpdatedAt: "recorded"}
	if err := repository.InsertAttachment(t.Context(), value); err != nil {
		t.Fatal(err)
	}
	value.Revision = vehiclemodel.AttachmentRevision(value)
	return db, repository, value
}

func TestAttachmentReadbackFailureRollsBackMetadataAndDeletedRow(t *testing.T) {
	db, repository, before := attachmentFixture(t)
	if _, err := db.Exec(`CREATE TRIGGER fail_attachment_readback AFTER UPDATE ON vehicle_attachments
BEGIN DELETE FROM vehicle_attachments WHERE id=NEW.id; END`); err != nil {
		t.Fatal(err)
	}
	service := vehiclemodel.NewAttachmentService(repository, nil)
	value, err := service.UpdateAttachment(t.Context(), before.VehicleID, before.ID, vehiclemodel.VehicleAttachmentUpdateInput{
		ExpectedRevision: &before.Revision, Description: "must roll back", Category: "replacement",
	})
	if value != nil || !errors.Is(err, vehiclemodel.ErrVehicleNotFound) {
		t.Fatal("failed readback acknowledged", value, err)
	}
	after, err := repository.Attachment(t.Context(), before.VehicleID, before.ID)
	if err != nil || !reflect.DeepEqual(after, &before) {
		t.Fatal("readback failure lost row, metadata or original file reference", after, err)
	}
}

func TestAttachmentDeletionCannotAffectAnotherVehiclesDocument(t *testing.T) {
	_, repository, before := attachmentFixture(t)
	service := vehiclemodel.NewAttachmentService(repository, nil)
	value, err := service.DeleteAttachment(t.Context(), "other-vehicle", before.ID, before.Revision)
	if value != nil || !errors.Is(err, vehiclemodel.ErrVehicleNotFound) {
		t.Fatal("cross-vehicle deletion accepted", value, err)
	}
	after, err := repository.Attachment(t.Context(), before.VehicleID, before.ID)
	if err != nil || !reflect.DeepEqual(after, &before) {
		t.Fatal("foreign document changed", after, err)
	}
}
