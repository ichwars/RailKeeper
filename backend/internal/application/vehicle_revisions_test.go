package application_test

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestVehicleRevisionGuardsAllCoreWritesInOneTransaction(t *testing.T) {
	db := testDB(t)
	s := application.NewVehicleService(db)
	input := application.CreateVehicleInput{InventoryNumber: "REV-1", Manufacturer: "Reference", Name: "Original",
		Gauge: "TT", Category: "Lokomotive", Gattung: "Diesellok",
		Images: []application.VehicleImageInput{{URL: "https://example.test/original.png"}}}
	v, err := s.Create(t.Context(), input, "actor")
	if err != nil {
		t.Fatal(err)
	}
	if len(v.Revision) != 64 || len(v.ImagesRevision) != 64 {
		t.Fatal("missing revisions")
	}
	// A second writer changes content without advancing the second-resolution clock.
	if _, err := db.Exec(`UPDATE vehicles SET name='Newer' WHERE id=?`, v.ID); err != nil {
		t.Fatal(err)
	}
	before, err := s.Get(t.Context(), v.ID)
	if err != nil {
		t.Fatal(err)
	}
	if before.UpdatedAt != v.UpdatedAt || before.Revision == v.Revision {
		t.Fatal("revision depends only on the clock")
	}
	var audits int
	if err := db.QueryRow(`SELECT COUNT(*) FROM audit_logs`).Scan(&audits); err != nil {
		t.Fatal(err)
	}
	input.Name, input.InventoryNumber, input.Images = "Stale", "REV-2", []application.VehicleImageInput{}
	input.ExpectedRevision, input.ExpectedImagesRevision = &v.Revision, &v.ImagesRevision
	input.FunctionChanges = []application.VehicleFunctionChange{{FunctionKey: "F4", Value: &application.VehicleFunctionInput{Name: "Must roll back", FlagsVersion: 2}}}
	if _, err := s.Update(t.Context(), v.ID, input, "actor"); !errors.Is(err, application.ErrVehicleConflict) {
		t.Fatalf("expected core conflict: %v", err)
	}
	after, err := s.Get(t.Context(), v.ID)
	if err != nil {
		t.Fatal(err)
	}
	var afterAudits int
	if err := db.QueryRow(`SELECT COUNT(*) FROM audit_logs`).Scan(&afterAudits); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, after) || audits != afterAudits {
		t.Fatal("conflict changed core, images, functions or audit")
	}
	input.ExpectedRevision = &before.Revision
	saved, err := s.Update(t.Context(), v.ID, input, "actor")
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := s.Get(t.Context(), v.ID)
	if err != nil {
		t.Fatal(err)
	}
	if saved.Revision != loaded.Revision || saved.ImagesRevision != loaded.ImagesRevision || saved.Name != "Stale" || len(saved.Images) != 0 || len(saved.Functions) != 1 {
		t.Fatal("acknowledgment is not the stored revision/data")
	}
}

func TestVehicleImagesRevisionProtectsNewUploadsAndMetadata(t *testing.T) {
	db := testDB(t)
	s := application.NewVehicleService(db)
	input := application.CreateVehicleInput{Manufacturer: "Reference", Name: "Original", Gauge: "TT", Category: "Lokomotive", Gattung: "Diesellok"}
	v, err := s.Create(t.Context(), input, "actor")
	if err != nil {
		t.Fatal(err)
	}
	uploaded, err := s.CreateImage(t.Context(), v.ID, application.VehicleImageInput{FileName: "local.png", MimeType: "image/png", StoragePath: "uploads/local.png"})
	if err != nil {
		t.Fatal(err)
	}
	current, err := s.Get(t.Context(), v.ID)
	if err != nil {
		t.Fatal(err)
	}
	if current.Revision != v.Revision || current.ImagesRevision == v.ImagesRevision {
		t.Fatal("independent image changed core revision or failed to change media revision")
	}
	if uploaded.Revision != current.Images[0].Revision {
		t.Fatal("image acknowledgment revision differs from loaded image")
	}
	input.Images, input.ExpectedRevision, input.ExpectedImagesRevision = []application.VehicleImageInput{}, &v.Revision, &v.ImagesRevision
	if _, err := s.Update(t.Context(), v.ID, input, "actor"); !errors.Is(err, application.ErrVehicleImageConflict) {
		t.Fatalf("expected image conflict: %v", err)
	}
	// Omitting images never replaces them, including legacy requests with no guards.
	input.Images = nil
	saved, err := s.Update(t.Context(), v.ID, input, "actor")
	if err != nil {
		t.Fatal(err)
	}
	if len(saved.Images) != 1 || saved.Images[0].ID != uploaded.ID {
		t.Fatal("omitted images were removed")
	}
	old := saved.Images[0].Revision
	if _, err := db.Exec(`UPDATE vehicle_images SET title='New title' WHERE id=?`, uploaded.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.DeleteImage(t.Context(), v.ID, uploaded.ID, old); !errors.Is(err, application.ErrVehicleImageConflict) {
		t.Fatalf("expected image delete conflict: %v", err)
	}
	if _, err := s.DeleteImage(t.Context(), v.ID, uploaded.ID, ""); !errors.Is(err, application.ErrVehicleValidation) {
		t.Fatalf("expected invalid empty guard: %v", err)
	}
	if _, err := s.DeleteImage(t.Context(), v.ID, uploaded.ID); err != nil {
		t.Fatalf("legacy delete: %v", err)
	}
}

func TestAttachmentRevisionGuardsUpdateAndDelete(t *testing.T) {
	s := application.NewVehicleService(testDB(t))
	v, err := s.Create(t.Context(), application.CreateVehicleInput{Manufacturer: "Reference", Name: "Original", Gauge: "TT", Category: "Lokomotive", Gattung: "Diesellok"}, "actor")
	if err != nil {
		t.Fatal(err)
	}
	a, err := s.CreateAttachment(t.Context(), v.ID, application.VehicleAttachmentInput{FileName: "manual.txt", OriginalName: "manual.txt", StoragePath: "uploads/manual.txt"})
	if err != nil {
		t.Fatal(err)
	}
	updated, err := s.UpdateAttachment(t.Context(), v.ID, a.ID, application.VehicleAttachmentUpdateInput{Description: "Newer", ExpectedRevision: &a.Revision})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Revision == a.Revision {
		t.Fatal("metadata update did not change revision")
	}
	if _, err := s.UpdateAttachment(t.Context(), v.ID, a.ID, application.VehicleAttachmentUpdateInput{Description: "Stale", ExpectedRevision: &a.Revision}); !errors.Is(err, application.ErrVehicleAttachmentConflict) {
		t.Fatalf("expected attachment update conflict: %v", err)
	}
	if _, err := s.DeleteAttachment(t.Context(), v.ID, a.ID, a.Revision); !errors.Is(err, application.ErrVehicleAttachmentConflict) {
		t.Fatalf("expected attachment delete conflict: %v", err)
	}
	bad := strings.Repeat("Z", 64)
	if _, err := s.DeleteAttachment(t.Context(), v.ID, a.ID, bad); !errors.Is(err, application.ErrVehicleValidation) {
		t.Fatalf("expected invalid guard: %v", err)
	}
	loaded, err := s.Get(t.Context(), v.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Attachments) != 1 || loaded.Attachments[0].Revision != updated.Revision || loaded.Revision != v.Revision {
		t.Fatal("attachment conflicts changed data or core revision")
	}
	if _, err := s.DeleteAttachment(t.Context(), v.ID, a.ID, updated.Revision); err != nil {
		t.Fatal(err)
	}
}

func TestImageCreationReceiptProtectsAConcurrentPrimarySelection(t *testing.T) {
	db := testDB(t)
	s := application.NewVehicleService(db)
	v, err := s.Create(t.Context(), application.CreateVehicleInput{Manufacturer: "Reference", Name: "Upload scope",
		Gauge: "TT", Category: "Lokomotive", Gattung: "Diesellok"}, "actor")
	if err != nil {
		t.Fatal(err)
	}
	input := application.VehicleImageInput{FileName: "reference.png", MimeType: "image/png", StoragePath: "uploads/reference.png", IsPrimary: true}
	if _, err := s.CreateImage(t.Context(), v.ID, input); err != nil {
		t.Fatal(err)
	}
	before, err := s.Get(t.Context(), v.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateImageWithRevision(t.Context(), v.ID, input, &v.ImagesRevision); !errors.Is(err, application.ErrVehicleImageConflict) {
		t.Fatalf("expected stale upload conflict: %v", err)
	}
	after, err := s.Get(t.Context(), v.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, after) {
		t.Fatal("rejected upload changed the current primary or collection")
	}
	created, err := s.CreateImageWithRevision(t.Context(), v.ID, input, &before.ImagesRevision)
	if err != nil {
		t.Fatal(err)
	}
	current, err := s.Get(t.Context(), v.ID)
	if err != nil {
		t.Fatal(err)
	}
	if created.ImagesRevision != current.ImagesRevision || len(current.Images) != 2 ||
		current.Images[0].ID != created.ID || !current.Images[0].IsPrimary || current.Images[1].IsPrimary {
		t.Fatal("creation receipt does not match its committed collection")
	}
	if _, err := db.Exec(`UPDATE vehicle_images SET title='Newer metadata' WHERE id=?`, created.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateImageWithRevision(t.Context(), v.ID, input, &created.ImagesRevision); !errors.Is(err, application.ErrVehicleImageConflict) {
		t.Fatalf("same-clock metadata change did not invalidate the upload scope: %v", err)
	}
}

func TestVehicleConflictCleansUnreferencedLocalizedBlobs(t *testing.T) {
	db := testDB(t)
	s := application.NewVehicleService(db)
	input := application.CreateVehicleInput{Manufacturer: "Reference", Name: "Original", Gauge: "TT", Category: "Lokomotive", Gattung: "Diesellok"}
	v, err := s.Create(t.Context(), input, "actor")
	if err != nil {
		t.Fatal(err)
	}
	blobs := application.NewFileBlobService(db, "")
	shared, err := blobs.Store(t.Context(), []byte("referenced bytes"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateAttachment(t.Context(), v.ID, application.VehicleAttachmentInput{FileName: "keep.txt", OriginalName: "keep.txt", BlobID: shared}); err != nil {
		t.Fatal(err)
	}
	var localizedID string
	s.SetImageLocalizer(func(ctx context.Context, id string, images []application.VehicleImageInput) ([]application.VehicleImageInput, error) {
		localizedID, err = blobs.Store(ctx, []byte("new image bytes"))
		images[0].BlobID, images[0].ThumbnailBlobID = localizedID, shared
		return images, err
	})
	input.Images = []application.VehicleImageInput{{URL: "https://example.test/new.png"}}
	bad := strings.Repeat("0", 64)
	input.ExpectedRevision = &bad
	if _, err := s.Update(t.Context(), v.ID, input, "actor"); !errors.Is(err, application.ErrVehicleConflict) {
		t.Fatalf("expected conflict: %v", err)
	}
	if _, err := blobs.Load(t.Context(), localizedID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("orphaned localized image remains: %v", err)
	}
	if data, err := blobs.Load(t.Context(), shared); err != nil || string(data) != "referenced bytes" {
		t.Fatal("cleanup deleted a referenced blob")
	}
}
