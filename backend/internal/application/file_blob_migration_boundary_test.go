package application_test

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestBlobMigrationReferenceFailureKeepsWindowsPathOriginalBytesAndRetryState(t *testing.T) {
	db := testDB(t)
	db.SetMaxOpenConns(1)
	vehicles := application.NewVehicleService(db)
	vehicle, err := vehicles.Create(t.Context(), mutationInput(), "actor")
	if err != nil {
		t.Fatal(err)
	}
	storedPath := `uploads\vehicles\legacy\manual.pdf`
	original, err := vehicles.CreateAttachment(t.Context(), vehicle.ID, application.VehicleAttachmentInput{
		FileName: "manual.pdf", OriginalName: "Original manual.pdf", StoragePath: storedPath, MimeType: "application/pdf", SizeBytes: 8,
	})
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	path := filepath.Join(root, "uploads", "vehicles", "legacy", "manual.pdf")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	contents := []byte{0, 255, 13, 10, 32, 1, 2, 3}
	if err := os.WriteFile(path, contents, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TRIGGER fail_blob_reference BEFORE UPDATE ON vehicle_attachments
WHEN NEW.blob_id<>'' BEGIN SELECT RAISE(ABORT,'synthetic blob-reference failure'); END`); err != nil {
		t.Fatal(err)
	}
	blobs := application.NewFileBlobService(db, root)
	if err := blobs.MigrateFilesystemBlobs(t.Context()); err == nil || !strings.Contains(err.Error(), "synthetic blob-reference failure") {
		t.Fatal("reference failure acknowledged", err)
	}
	afterFailure, err := vehicles.GetAttachment(t.Context(), vehicle.ID, original.ID)
	if err != nil || !reflect.DeepEqual(original, afterFailure) {
		t.Fatal("failed reference write changed original metadata", err)
	}
	if bytesOnDisk, err := os.ReadFile(path); err != nil || !bytes.Equal(bytesOnDisk, contents) {
		t.Fatal("failed migration removed or changed source bytes", err)
	}
	if _, err := db.Exec(`DROP TRIGGER fail_blob_reference`); err != nil {
		t.Fatal(err)
	}
	if err := blobs.MigrateFilesystemBlobs(t.Context()); err != nil {
		t.Fatal("retry after failed binding", err)
	}
	migrated, err := vehicles.GetAttachment(t.Context(), vehicle.ID, original.ID)
	if err != nil {
		t.Fatal(err)
	}
	if migrated.ID != original.ID || migrated.CreatedAt != original.CreatedAt || migrated.StoragePath != storedPath || migrated.BlobID == "" {
		t.Fatal("migration changed identity, timestamp or stored path")
	}
	content, err := blobs.Load(t.Context(), migrated.BlobID)
	if err != nil || !bytes.Equal(content, contents) {
		t.Fatal("stored original bytes differ", err)
	}
	beforeRepeat := databaseRowFingerprints(t, db)
	if err := blobs.MigrateFilesystemBlobs(t.Context()); err != nil {
		t.Fatal(err)
	}
	if afterRepeat := databaseRowFingerprints(t, db); !reflect.DeepEqual(beforeRepeat, afterRepeat) {
		t.Fatal("completed migration changed database again")
	}
	if bytesOnDisk, err := os.ReadFile(path); err != nil || !bytes.Equal(bytesOnDisk, contents) {
		t.Fatal("successful migration removed or changed source bytes", err)
	}
}

func TestFileBlobInsertFailureAcknowledgesNoIDAndKeepsLegacySource(t *testing.T) {
	db := testDB(t)
	db.SetMaxOpenConns(1)
	vehicles := application.NewVehicleService(db)
	vehicle, err := vehicles.Create(t.Context(), mutationInput(), "actor")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := vehicles.CreateAttachment(t.Context(), vehicle.ID, application.VehicleAttachmentInput{
		FileName: "original.pdf", OriginalName: "original.pdf", StoragePath: "uploads/original.pdf", MimeType: "application/pdf", SizeBytes: 8,
	}); err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	path := filepath.Join(root, "uploads", "original.pdf")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	contents := []byte{0, 255, 13, 10, 32, 1, 2, 3}
	if err := os.WriteFile(path, contents, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TRIGGER fail_blob_insert BEFORE INSERT ON file_blobs
BEGIN SELECT RAISE(ABORT,'synthetic blob insert failure'); END`); err != nil {
		t.Fatal(err)
	}
	before := databaseRowFingerprints(t, db)
	blobs := application.NewFileBlobService(db, root)
	if id, err := blobs.Store(t.Context(), contents); id != "" || err == nil || !strings.Contains(err.Error(), "synthetic blob insert failure") {
		t.Fatal("failed insertion exposed an unpersisted blob ID", id, err)
	}
	if err := blobs.MigrateFilesystemBlobs(t.Context()); err == nil || !strings.Contains(err.Error(), "synthetic blob insert failure") {
		t.Fatal("failed legacy capture acknowledged", err)
	}
	if after := databaseRowFingerprints(t, db); !reflect.DeepEqual(before, after) {
		t.Fatal("failed capture changed rows or references")
	}
	if disk, err := os.ReadFile(path); err != nil || !bytes.Equal(disk, contents) {
		t.Fatal("failed capture lost source bytes", err)
	}
}
func TestFileBlobLoadsLegacyEncodingsAndRejectsCorruptOriginalWithoutPartialData(t *testing.T) {
	db := testDB(t)
	blobs := application.NewFileBlobService(db, "")
	contents := []byte{0, 255, 13, 10, 32, 1, 2, 3}
	id, err := blobs.Store(t.Context(), contents)
	if err != nil {
		t.Fatal(err)
	}
	var compressed []byte
	if err := db.QueryRow(`SELECT data FROM file_blobs WHERE id=?`, id).Scan(&compressed); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name, compression string
		data              []byte
		valid             bool
	}{
		{"legacy none", "none", contents, true},
		{"legacy empty marker", "", contents, true},
		{"case and whitespace marker", " ZLIB ", compressed, true},
		{"invalid header", "zlib", []byte("invalid compressed original"), false},
		{"truncated trailer", "zlib", compressed[:len(compressed)-2], false},
		{"unknown encoding", "unsupported", contents, false},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if _, err := db.Exec(`UPDATE file_blobs SET compression=?,data=? WHERE id=?`, test.compression, test.data, id); err != nil {
				t.Fatal(err)
			}
			data, err := blobs.Load(t.Context(), id)
			if test.valid {
				if err != nil || !bytes.Equal(data, contents) {
					t.Fatal("legacy original changed", err)
				}
			} else if err == nil || data != nil {
				t.Fatal("corrupt original returned partial or guessed data", err)
			}
			var stored []byte
			var compression string
			if err := db.QueryRow(`SELECT compression,data FROM file_blobs WHERE id=?`, id).Scan(&compression, &stored); err != nil {
				t.Fatal(err)
			}
			if compression != test.compression || !bytes.Equal(stored, test.data) {
				t.Fatal("reading silently rewrote original encoding")
			}
		})
	}
}
