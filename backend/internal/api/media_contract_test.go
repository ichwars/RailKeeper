package api

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestMediaHTTPContractsRetainBytesAndRejectUnsafeUploads(t *testing.T) {
	f := newContractFixture(t)
	id := f.vehicle(t)
	base := "/api/v1/vehicles/" + id
	original := []byte("Original data, UTF-8: ÄÖÜ\nCV1=3\n")
	for _, contract := range []struct {
		suffix, name, download string
		data                   []byte
		fields                 map[string]string
	}{
		{"attachments", "Manual.txt", "download", original, map[string]string{"description": "Original manual", "category": "Anleitung"}},
		{"images", "Train.png", "file", testPNG(t), map[string]string{"title": "Original train"}},
		{"cv-files", "Decoder.txt", "download", original, nil},
	} {
		t.Run(contract.suffix, func(t *testing.T) {
			path := base + "/" + contract.suffix
			created := f.upload(t, path, contract.name, contract.data, contract.fields)
			assertStatus(t, created, 201)
			entryID := responseID(t, created)
			// Image creation also returns a collection revision alongside the embedded image.
			download := path + "/" + entryID + "/" + contract.download
			for _, suffix := range []string{"", "?inline=true"} {
				result := f.call(t, "GET", download+suffix, nil)
				assertStatus(t, result, 200)
				if !bytes.Equal(result.Body.Bytes(), contract.data) {
					t.Fatalf("%s bytes changed", contract.suffix)
				}
				if result.Header().Get("Content-Disposition") == "" || result.Header().Get("X-Content-Type-Options") != "nosniff" {
					t.Fatal("media response lost security headers")
				}
			}
			assertStatus(t, f.call(t, "GET", path+"/missing/"+contract.download, nil), 404)
			if contract.suffix == "images" {
				assertStatus(t, f.call(t, "GET", path+"/"+entryID+"/thumbnail", nil), 200)
			}
			// A forbidden mutation must leave the downloadable original untouched.
			assertStatus(t, authedJSONRequest(t, f.router, f.viewer, "DELETE", path+"/"+entryID, nil, true), 403)
			assertStatus(t, f.call(t, "GET", download, nil), 200)
			assertStatus(t, f.call(t, "DELETE", path+"/"+entryID, nil), 204)
			assertStatus(t, f.call(t, "GET", download, nil), 404)
		})
	}
	attachmentPath := base + "/attachments"
	for _, invalid := range []struct {
		name string
		data []byte
		code string
	}{
		{"unsafe.exe", []byte("MZ unsafe"), "attachment_type_blocked"},
		{"empty.txt", nil, "attachment_empty"},
		{"script.html", []byte("<script>alert(1)</script>"), "attachment_type_blocked"},
	} {
		assertProblem(t, f.upload(t, attachmentPath, invalid.name, invalid.data, nil), 400, invalid.code)
	}
	assertProblem(t, f.upload(t, attachmentPath, "", nil, nil), 400, "attachment_missing")
	assertProblem(t, f.raw(t, "POST", attachmentPath, "not multipart"), 400, "attachment_upload_invalid")
	assertProblem(t, f.upload(t, base+"/images", "unsafe.txt", original, nil), 400, "image_type_blocked")
	for _, endpoint := range []string{attachmentPath + "/import-url", base + "/images/import-url"} {
		assertProblem(t, f.raw(t, "POST", endpoint, "{"), 400, "invalid_json")
		status := 400
		if strings.Contains(endpoint, "/images/") {
			status = 502
		}
		assertStatus(t, f.call(t, "POST", endpoint, map[string]string{"url": "http://127.0.0.1/private"}), status)
	}
	f.config.MaxAttachmentBytes = 8
	f.router = NewRouter(f.config)
	assertProblem(t, f.upload(t, attachmentPath, "large.txt", original, nil), 400, "attachment_too_large")
	assertProblem(t, f.upload(t, "/api/v1/cv-files/preview", "large.txt", original, nil), 400, "cv_file_blocked")
}

func TestLegacyAttachmentPreviewIsSandboxedAndConfined(t *testing.T) {
	f := newContractFixture(t)
	id := f.vehicle(t)
	html := []byte("<html><script>globalThis.invalid = true</script></html>")
	rel := "uploads/manual.html"
	full := filepath.Join(f.dataDir, rel)
	if err := os.MkdirAll(filepath.Dir(full), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, html, 0600); err != nil {
		t.Fatal(err)
	}
	value, err := f.vehicles.CreateAttachment(t.Context(), id, application.VehicleAttachmentInput{FileName: "manual.html", StoragePath: rel, OriginalName: "manual.html", MimeType: "text/html", SizeBytes: int64(len(html))})
	if err != nil {
		t.Fatal(err)
	}
	path := "/api/v1/vehicles/" + id + "/attachments/" + value.ID + "/download"
	response := f.call(t, "GET", path+"?inline=true", nil)
	assertStatus(t, response, 200)
	if !strings.Contains(response.Header().Get("Content-Security-Policy"), "sandbox") || !strings.HasPrefix(response.Header().Get("Content-Disposition"), "inline") || !bytes.Equal(response.Body.Bytes(), html) {
		t.Fatal("legacy preview must preserve bytes and sandbox active content")
	}
	response = f.call(t, "GET", path, nil)
	assertStatus(t, response, 200)
	if !strings.HasPrefix(response.Header().Get("Content-Disposition"), "attachment") {
		t.Fatal("download must default to attachment")
	}
	app := &App{dataDir: f.dataDir, fileBlobs: f.config.FileBlobService}
	read, err := app.readAttachmentData(t.Context(), *value, int64(len(html)))
	if err != nil || !bytes.Equal(read, html) {
		t.Fatalf("legacy read: %v", err)
	}
	if _, err := app.readAttachmentData(t.Context(), *value, 8); err == nil {
		t.Fatal("legacy read limit ignored")
	}
	value.StoragePath = "../private.txt"
	if _, err := app.readAttachmentData(t.Context(), *value, 1024); err == nil {
		t.Fatal("escaped data directory")
	}
}

func TestCVPreviewHTTPContractsAndIndexedJSONRoundTrip(t *testing.T) {
	f := newContractFixture(t)
	values := []map[string]any{{"cvNumber": 1, "value": 3, "description": "Address"}, {"cvNumber": 257, "value": 42, "indexCV31": 16, "indexCV32": 1}}
	for _, document := range []any{values, map[string]any{"cvValues": values}} {
		data, err := json.Marshal(document)
		if err != nil {
			t.Fatal(err)
		}
		response := f.upload(t, "/api/v1/cv-files/preview", "decoder.json", data, nil)
		assertStatus(t, response, 200)
		var preview vehicleCVFilePreviewResponse
		decodeResponse(t, response, &preview)
		if len(preview.SuggestedCVValues) != 2 || preview.SuggestedCVValues[1].CVNumber != 257 || preview.SuggestedCVValues[1].Value != 42 || preview.SuggestedCVValues[1].IndexCV31 == nil || *preview.SuggestedCVValues[1].IndexCV31 != 16 || preview.SuggestedCVValues[1].IndexCV32 == nil || *preview.SuggestedCVValues[1].IndexCV32 != 1 {
			t.Fatalf("indexed CV preview changed: %s", response.Body)
		}
	}
	for _, invalid := range []struct {
		name string
		data []byte
		code string
	}{
		{"empty.txt", nil, "cv_file_empty"}, {"unsafe.exe", []byte("MZ"), "cv_file_blocked"},
		{"", nil, "cv_file_missing"},
	} {
		assertProblem(t, f.upload(t, "/api/v1/cv-files/preview", invalid.name, invalid.data, nil), 400, invalid.code)
	}
	assertProblem(t, f.raw(t, "POST", "/api/v1/cv-files/preview", "bad"), 400, "cv_file_preview_invalid")
	if _, err := previewCVJSON([]byte("[invalid")); err == nil {
		t.Fatal("invalid array accepted")
	}
	if _, err := previewCVJSON([]byte("{invalid")); err == nil {
		t.Fatal("invalid document accepted")
	}
	oversized := bytes.Repeat([]byte("{\"cvNumber\":1,\"value\":1},"), maxDecoderCVImportRows)
	oversized = append(append([]byte("["), oversized...), []byte("{\"cvNumber\":1,\"value\":1}]")...)
	if _, err := previewCVJSON(oversized); err == nil {
		t.Fatal("CV row limit ignored")
	}
	assertStatus(t, f.call(t, http.MethodGet, "/api/v1/vehicles/missing/cv-values", nil), 404)
}

func TestUploadedImageThumbnailsFitBothOrientationsWithoutChangingOriginals(t *testing.T) {
	f := newContractFixture(t)
	id := f.vehicle(t)
	for _, size := range []image.Point{{X: 900, Y: 300}, {X: 300, Y: 900}} {
		var original bytes.Buffer
		src := image.NewRGBA(image.Rect(0, 0, size.X, size.Y))
		src.SetRGBA(size.X/2, size.Y/2, color.RGBA{R: 240, G: 180, B: 20, A: 255})
		if err := png.Encode(&original, src); err != nil {
			t.Fatal(err)
		}
		response := f.upload(t, "/api/v1/vehicles/"+id+"/images", "full-size.png", original.Bytes(), nil)
		assertStatus(t, response, 201)
		path := "/api/v1/vehicles/" + id + "/images/" + responseID(t, response)
		downloaded := f.call(t, "GET", path+"/file", nil)
		assertStatus(t, downloaded, 200)
		if !bytes.Equal(downloaded.Body.Bytes(), original.Bytes()) {
			t.Fatal("thumbnail generation replaced original image")
		}
		thumb := f.call(t, "GET", path+"/thumbnail", nil)
		assertStatus(t, thumb, 200)
		config, format, err := image.DecodeConfig(bytes.NewReader(thumb.Body.Bytes()))
		if err != nil || format != "jpeg" || config.Width > 360 || config.Height > 240 || config.Width < 1 || config.Height < 1 {
			t.Fatalf("invalid bounded thumbnail: %+v %s %v", config, format, err)
		}
	}
}
