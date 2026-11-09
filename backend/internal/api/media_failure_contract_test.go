package api

import (
	"bytes"
	stdimage "image"
	"image/jpeg"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestMediaHTTPFailureContractsNeverLoseExistingOriginals(t *testing.T) {
	f := newContractFixture(t)
	id := f.vehicle(t)
	base := "/api/v1/vehicles/" + id
	original := []byte("Original decoder CV1=3\n")
	image := f.upload(t, base+"/images", "original.png", testPNG(t), nil)
	assertStatus(t, image, 201)
	imageID := responseID(t, image)
	attachment := f.upload(t, base+"/attachments", "original.txt", original, nil)
	assertStatus(t, attachment, 201)
	attachmentID := responseID(t, attachment)
	cvFile := f.upload(t, base+"/cv-files", "original.txt", original, nil)
	assertStatus(t, cvFile, 201)
	cvFileID := responseID(t, cvFile)
	assertStatus(t, f.call(t, "GET", base+"/cv-values", nil), 200)
	for _, c := range []struct{ suffix, missing, invalid string }{
		{"images", "image_missing", "image_upload_invalid"}, {"attachments", "attachment_missing", "attachment_upload_invalid"}, {"cv-files", "cv_file_missing", "cv_file_upload_invalid"},
	} {
		assertProblem(t, f.upload(t, base+"/"+c.suffix, "", nil, nil), 400, c.missing)
		assertProblem(t, f.raw(t, "POST", base+"/"+c.suffix, "invalid multipart"), 400, c.invalid)
	}
	assertProblem(t, f.upload(t, base+"/images", "dimensions.png", pngHeader(20000, 20000), nil), 400, "image_dimensions_invalid")
	assertProblem(t, f.upload(t, "/api/v1/vehicles/missing/images", "original.png", testPNG(t), nil), 404, "vehicle_not_found")
	assertProblem(t, f.upload(t, "/api/v1/vehicles/missing/attachments", "original.txt", original, nil), 404, "vehicle_not_found")
	assertProblem(t, f.upload(t, "/api/v1/vehicles/missing/cv-files", "original.txt", original, nil), 404, "vehicle_not_found")
	f.config.FileBlobService = nil
	f.router = NewRouter(f.config)
	for _, c := range []struct{ path, code string }{
		{base + "/images/" + imageID + "/file", "image_download_failed"},
		{base + "/images/" + imageID + "/thumbnail", "image_thumbnail_failed"},
		{base + "/attachments/" + attachmentID + "/download", "attachment_download_failed"},
		{base + "/cv-files/" + cvFileID + "/download", "cv_file_download_failed"},
	} {
		assertProblem(t, f.call(t, "GET", c.path, nil), 500, c.code)
	}
	for _, c := range []struct {
		suffix, name string
		data         []byte
		code         string
	}{
		{"images", "new.png", testPNG(t), "image_upload_failed"},
		{"attachments", "new.txt", original, "attachment_upload_failed"},
		{"cv-files", "new.txt", original, "cv_file_upload_failed"},
	} {
		assertProblem(t, f.upload(t, base+"/"+c.suffix, c.name, c.data, nil), 500, c.code)
	}
	f.config.FileBlobService = application.NewFileBlobService(f.db, f.dataDir)
	f.config.MaxImageBytes = 8
	f.router = NewRouter(f.config)
	assertProblem(t, f.upload(t, base+"/images", "too-large.png", testPNG(t), nil), 400, "image_too_large")
	// A failed writer and loader must not have damaged the existing metadata or bytes.
	for _, path := range []string{base + "/attachments/" + attachmentID + "/download", base + "/cv-files/" + cvFileID + "/download"} {
		response := f.call(t, "GET", path, nil)
		assertStatus(t, response, 200)
		if !bytes.Equal(response.Body.Bytes(), original) {
			t.Fatal("original changed after storage failure")
		}
	}
	value, err := f.vehicles.GetAttachment(t.Context(), id, attachmentID)
	if err != nil {
		t.Fatal(err)
	}
	app := &App{dataDir: f.dataDir, fileBlobs: f.config.FileBlobService}
	if _, err := app.readAttachmentData(t.Context(), *value, 1); err == nil {
		t.Fatal("blob read limit ignored")
	}
	if data, err := app.readAttachmentData(t.Context(), *value, 1024); err != nil || !bytes.Equal(data, original) {
		t.Fatalf("blob read changed original: %v", err)
	}
	app.fileBlobs = nil
	if _, err := app.readAttachmentData(t.Context(), *value, 1024); err == nil {
		t.Fatal("unavailable blob must fail")
	}
}

func TestLegacyMediaHTTPFilesRemainConfinedAndKeepTheirOriginalBytes(t *testing.T) {
	f := newContractFixture(t)
	id := f.vehicle(t)
	base := "/api/v1/vehicles/" + id
	image, err := f.vehicles.CreateImage(t.Context(), id, application.VehicleImageInput{URL: "https://example.test/original.png", Title: "Legacy", FileName: "original.png", MimeType: "image/png", StoragePath: "uploads/legacy.png"})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(f.dataDir, "uploads"), 0700); err != nil {
		t.Fatal(err)
	}
	data := testPNG(t)
	if err := os.WriteFile(filepath.Join(f.dataDir, "uploads/legacy.png"), data, 0600); err != nil {
		t.Fatal(err)
	}
	for _, suffix := range []string{"file", "thumbnail"} {
		response := f.call(t, "GET", base+"/images/"+image.ID+"/"+suffix, nil)
		assertStatus(t, response, 200)
		if !bytes.Equal(response.Body.Bytes(), data) {
			t.Fatal("legacy image changed")
		}
	}
	var thumbnail bytes.Buffer
	if err := jpeg.Encode(&thumbnail, stdimage.NewRGBA(stdimage.Rect(0, 0, 8, 8)), nil); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(f.dataDir, "uploads/legacy-thumb.jpg"), thumbnail.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := f.db.Exec(`UPDATE vehicle_images SET thumbnail_path='uploads/legacy-thumb.jpg' WHERE id=?`, image.ID); err != nil {
		t.Fatal(err)
	}
	preview := f.call(t, "GET", base+"/images/"+image.ID+"/thumbnail", nil)
	assertStatus(t, preview, 200)
	if !bytes.Equal(preview.Body.Bytes(), thumbnail.Bytes()) || preview.Header().Get("Content-Type") != "image/jpeg" {
		t.Fatal("legacy thumbnail changed")
	}
	// Simulate a damaged historical path. Neither the original nor any outside file may be served.
	if _, err := f.db.Exec(`UPDATE vehicle_images SET storage_path='../outside.png',thumbnail_path='../outside.jpg' WHERE id=?`, image.ID); err != nil {
		t.Fatal(err)
	}
	assertProblem(t, f.call(t, "GET", base+"/images/"+image.ID+"/file", nil), 500, "image_path_invalid")
	assertProblem(t, f.call(t, "GET", base+"/images/"+image.ID+"/thumbnail", nil), 500, "image_path_invalid")
	if _, err := f.db.Exec(`UPDATE vehicle_images SET storage_path='',thumbnail_path='' WHERE id=?`, image.ID); err != nil {
		t.Fatal(err)
	}
	assertProblem(t, f.call(t, "GET", base+"/images/"+image.ID+"/file", nil), 404, "image_file_missing")
	assertStatus(t, f.call(t, http.MethodGet, base+"/images/missing/thumbnail", nil), 404)
	legacy, err := f.vehicles.CreateCVFile(t.Context(), id, application.VehicleCVFileInput{FileName: "original.txt", OriginalName: "original.txt", StoragePath: "uploads/original.txt", MimeType: "text/plain", SizeBytes: 8})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(f.dataDir, "uploads/original.txt"), []byte("CV1=3\n"), 0600); err != nil {
		t.Fatal(err)
	}
	response := f.call(t, "GET", base+"/cv-files/"+legacy.ID+"/download", nil)
	assertStatus(t, response, 200)
	if response.Body.String() != "CV1=3\n" {
		t.Fatal("legacy CV file changed")
	}
	if _, err := f.db.Exec(`UPDATE vehicle_cv_files SET storage_path='../outside.txt' WHERE id=?`, legacy.ID); err != nil {
		t.Fatal(err)
	}
	assertProblem(t, f.call(t, "GET", base+"/cv-files/"+legacy.ID+"/download", nil), 500, "cv_file_path_invalid")
}
