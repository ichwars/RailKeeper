package application_test

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestDecoderConfigurationMediaMigrationDoesNotCopyPrivateRootFiles(t *testing.T) {
	for _, storedPath := range []string{"private-settings.json", "uploads/../private-settings.json", `uploads\..\private-settings.json`} {
		t.Run(storedPath, func(t *testing.T) {
			db, backup, root, vehicle := legacyDecoderCaptureFixture(t)
			private := []byte("Synthetic private installation configuration")
			if err := os.WriteFile(filepath.Join(root, "private-settings.json"), private, 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := application.NewVehicleService(db).CreateCVFile(t.Context(), vehicle, application.VehicleCVFileInput{
				FileName: "source.csv", OriginalName: "source.csv", StoragePath: storedPath,
			}); err != nil {
				t.Fatal(err)
			}
			// Simulate an old writer that did not maintain workspace metadata.
			if _, err := db.Exec(`DELETE FROM decoder_cv_workspaces`); err != nil {
				t.Fatal(err)
			}
			before := decoderSnapshots(t, backup)
			if err := application.NewFileBlobService(db, root).MigrateFilesystemBlobs(t.Context()); !errors.Is(err, application.ErrBackupPath) {
				t.Fatalf("unsafe original path accepted: %v", err)
			}
			if err := application.BackfillLegacyDecoderCV(t.Context(), db, root); !errors.Is(err, application.ErrBackupPath) {
				t.Fatalf("snapshot source escaped uploads: %v", err)
			}
			after := decoderSnapshots(t, backup)
			if !reflect.DeepEqual(catalogBackupTableSnapshot(t, before), catalogBackupTableSnapshot(t, after)) {
				t.Fatal("private configuration was copied or partial evidence remained")
			}
			if validation, err := backup.Validate(t.Context(), before); err != nil || validation.Compatible {
				t.Fatalf("unsafe backup reference passed preflight: %#v / %v", validation, err)
			}
			if _, err := backup.Import(t.Context(), before); !errors.Is(err, application.ErrBackupInvalid) {
				t.Fatalf("unsafe backup entered destructive restore: %v", err)
			}
		})
	}
}

func TestDecoderConfigurationMediaMigrationPreservesPortableWindowsOriginal(t *testing.T) {
	db, _, root, vehicle := legacyDecoderCaptureFixture(t)
	if err := os.MkdirAll(filepath.Join(root, "uploads"), 0700); err != nil {
		t.Fatal(err)
	}
	content := []byte("Portable original bytes")
	if err := os.WriteFile(filepath.Join(root, "uploads", "source.csv"), content, 0600); err != nil {
		t.Fatal(err)
	}
	file, err := application.NewVehicleService(db).CreateCVFile(t.Context(), vehicle, application.VehicleCVFileInput{
		FileName: "source.csv", OriginalName: "source.csv", StoragePath: `uploads\source.csv`,
	})
	if err != nil {
		t.Fatal(err)
	}
	blobs := application.NewFileBlobService(db, root)
	if err := blobs.MigrateFilesystemBlobs(t.Context()); err != nil {
		t.Fatal(err)
	}
	after, err := application.NewVehicleService(db).GetCVFile(t.Context(), vehicle, file.ID)
	if err != nil || after.StoragePath != file.StoragePath || after.BlobID == "" {
		t.Fatalf("original path text changed or portable bytes were missed: %#v / %v", after, err)
	}
	if data, err := blobs.Load(t.Context(), after.BlobID); err != nil || !reflect.DeepEqual(data, content) {
		t.Fatalf("portable original bytes changed: %v", err)
	}
}

func TestDecoderConfigurationMediaMigrationRejectsSymlinkOutsideUploads(t *testing.T) {
	db, backup, root, vehicle := legacyDecoderCaptureFixture(t)
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "private.csv"), []byte("Private synthetic file"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "uploads"), 0700); err != nil {
		t.Fatal(err)
	}
	if _, err := application.NewVehicleService(db).CreateCVFile(t.Context(), vehicle, application.VehicleCVFileInput{
		FileName: "source.csv", OriginalName: "source.csv", StoragePath: "uploads/external/private.csv",
	}); err != nil {
		t.Fatal(err)
	}
	// This case exercises adoption of a legacy file, before workspace-aware writers.
	if _, err := db.Exec(`DELETE FROM decoder_cv_workspaces`); err != nil {
		t.Fatal(err)
	}
	before := decoderSnapshots(t, backup)
	link := filepath.Join(root, "uploads", "external")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("host cannot create symlinks: %v", err)
	}
	if err := application.NewFileBlobService(db, root).MigrateFilesystemBlobs(t.Context()); err == nil {
		t.Fatal("symlink outside upload root was followed")
	}
	if err := application.BackfillLegacyDecoderCV(t.Context(), db, root); err == nil {
		t.Fatal("snapshot followed a symlink outside uploads")
	}
	if err := os.Remove(link); err != nil {
		t.Fatal(err)
	}
	after := decoderSnapshots(t, backup)
	if !reflect.DeepEqual(catalogBackupTableSnapshot(t, before), catalogBackupTableSnapshot(t, after)) {
		t.Fatal("outside bytes were copied into private media or snapshots")
	}
}
