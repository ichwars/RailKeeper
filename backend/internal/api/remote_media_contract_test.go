package api

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"railkeeper/backend/internal/application"
)

// Keep public-URL validation and the production transport wrapper active. The
// cloned transport's TLS dialer terminates on an in-memory HTTP peer, so this test
// cannot contact the public address or any installation/hardware service.
func syntheticRemoteMedia(t *testing.T, respond func(*http.Request) *http.Response) {
	t.Helper()
	original := http.DefaultTransport
	transport := original.(*http.Transport).Clone()
	transport.DialTLSContext = func(_ context.Context, _, address string) (net.Conn, error) {
		if address != "93.184.216.34:443" {
			return nil, fmt.Errorf("unexpected synthetic target %q", address)
		}
		client, server := net.Pipe()
		go func() {
			defer func() { _ = server.Close() }()
			request, err := http.ReadRequest(bufio.NewReader(server))
			if err != nil {
				return
			}
			response := respond(request)
			defer func() { _ = response.Body.Close() }()
			response.ProtoMajor, response.ProtoMinor = 1, 1
			response.Close = true
			_ = response.Write(server)
		}()
		return client, nil
	}
	http.DefaultTransport = transport
	t.Cleanup(func() { transport.CloseIdleConnections(); http.DefaultTransport = original })
}
func mediaResponse(status int, mime string, data []byte) *http.Response {
	return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": []string{mime}}, Body: io.NopCloser(bytes.NewReader(data)), ContentLength: int64(len(data))}
}

func TestRemoteDocumentHTTPContractRetainsOriginalAndCleansRejectedWrites(t *testing.T) {
	f := newContractFixture(t)
	id := f.vehicle(t)
	original := []byte("%PDF-1.4\nSynthetic original manual\n%%EOF")
	content, status, mime := original, 200, "application/pdf; charset=binary"
	syntheticRemoteMedia(t, func(r *http.Request) *http.Response {
		if r.Method != "GET" || r.Header.Get("User-Agent") == "" {
			t.Error("unexpected remote document request")
		}
		return mediaResponse(status, mime, content)
	})
	path := "/api/v1/vehicles/" + id + "/attachments/import-url"
	input := map[string]any{"url": "https://93.184.216.34/manual", "title": "Bedienungsanleitung"}
	response := f.call(t, "POST", path, input)
	assertStatus(t, response, 201)
	var attachment application.VehicleAttachment
	decodeResponse(t, response, &attachment)
	if attachment.OriginalName != "Bedienungsanleitung.pdf" || attachment.Category != "Anleitung" || attachment.MimeType != "application/pdf" {
		t.Fatalf("document metadata changed: %#v", attachment)
	}
	download := f.call(t, "GET", "/api/v1/vehicles/"+id+"/attachments/"+attachment.ID+"/download", nil)
	assertStatus(t, download, 200)
	if !bytes.Equal(download.Body.Bytes(), original) {
		t.Fatal("download changed imported document bytes")
	}
	var before int
	if err := f.db.QueryRow(`SELECT COUNT(*) FROM file_blobs`).Scan(&before); err != nil {
		t.Fatal(err)
	}
	assertProblem(t, f.call(t, "POST", "/api/v1/vehicles/missing/attachments/import-url", input), 404, "vehicle_not_found")
	for _, c := range []struct {
		status int
		mime   string
		data   []byte
		code   string
		want   int
	}{
		{404, "text/plain", []byte("missing"), "attachment_fetch_failed", 502},
		{200, "text/plain", nil, "attachment_empty", 400},
		{200, "application/octet-stream", []byte("MZ\x00\x00 executable"), "attachment_type_blocked", 400},
	} {
		status, mime, content = c.status, c.mime, c.data
		assertProblem(t, f.call(t, "POST", path, input), c.want, c.code)
	}
	var after int
	if err := f.db.QueryRow(`SELECT COUNT(*) FROM file_blobs`).Scan(&after); err != nil || after != before {
		t.Fatalf("rejected import left blobs: %d -> %d (%v)", before, after, err)
	}
	content, status, mime = bytes.Repeat([]byte("X"), 256), 200, "text/plain"
	f.config.MaxAttachmentBytes = 128
	f.router = NewRouter(f.config)
	assertProblem(t, f.call(t, "POST", path, input), 400, "attachment_too_large")
}

func TestRemoteImageHTTPContractLocalizesOnceAndPreservesSource(t *testing.T) {
	f := newContractFixture(t)
	id := f.vehicle(t)
	image := testPNG(t)
	content, status, mime := image, 200, "image/png"
	syntheticRemoteMedia(t, func(*http.Request) *http.Response { return mediaResponse(status, mime, content) })
	path := "/api/v1/vehicles/" + id + "/images/import-url"
	input := map[string]any{"url": "https://93.184.216.34/train.png", "title": "Original train"}
	response := f.call(t, "POST", path, input)
	assertStatus(t, response, 201)
	var stored application.VehicleImageCreation
	decodeResponse(t, response, &stored)
	download := f.call(t, "GET", "/api/v1/vehicles/"+id+"/images/"+stored.ID+"/file", nil)
	assertStatus(t, download, 200)
	if !bytes.Equal(download.Body.Bytes(), image) {
		t.Fatal("remote image bytes changed")
	}
	assertStatus(t, f.call(t, "GET", "/api/v1/vehicles/"+id+"/images/"+stored.ID+"/thumbnail", nil), 200)
	assertProblem(t, f.call(t, "POST", "/api/v1/vehicles/missing/images/import-url", input), 404, "vehicle_not_found")
	// The normal vehicle creation path also localizes source images, while leaving
	// existing blob/file references and failed optional suggestions intact.
	body := map[string]any{"name": "Suggested image vehicle", "manufacturer": "Test", "gauge": "H0", "category": "Lokomotive", "gattung": "Diesellok", "images": []any{
		map[string]any{"url": "https://93.184.216.34/train.png", "title": "Source"},
		map[string]any{"url": "http://127.0.0.1/private", "title": "Retained suggestion"},
	}}
	response = f.call(t, "POST", "/api/v1/vehicles", body)
	assertStatus(t, response, 201)
	var vehicle application.Vehicle
	decodeResponse(t, response, &vehicle)
	if len(vehicle.Images) != 2 || !strings.HasPrefix(vehicle.Images[0].URL, "/api/v1/vehicles/") || vehicle.Images[0].SourceURL != "https://93.184.216.34/train.png" || vehicle.Images[1].URL != "http://127.0.0.1/private" {
		t.Fatalf("optional localization lost original suggestions: %#v", vehicle.Images)
	}
	for _, invalid := range []struct {
		status int
		data   []byte
	}{{404, image}, {200, nil}, {200, []byte("plain text")}, {200, pngHeader(20000, 20000)}} {
		status, content = invalid.status, invalid.data
		assertProblem(t, f.call(t, "POST", path, input), 502, "image_import_failed")
	}
	content, status = image, 200
	f.config.MaxImageBytes = 8
	f.router = NewRouter(f.config)
	assertProblem(t, f.call(t, "POST", path, input), 502, "image_import_failed")
}

func TestRemoteFileNamesKeepSafeExtensionsAndDocumentCategories(t *testing.T) {
	for _, c := range []struct{ mime, extension string }{{"application/pdf", ".pdf"}, {"application/json", ".json"}, {"text/xml", ".xml"}, {"application/zip", ".zip"}, {"image/jpeg", ".jpg"}, {"image/png", ".png"}, {"image/webp", ".webp"}, {"text/csv", ".csv"}, {"text/plain", ".txt"}} {
		if got := attachmentExtensionForMime(c.mime); got != c.extension {
			t.Fatalf("%s extension: %s", c.mime, got)
		}
		name := remoteAttachmentFileName(importVehicleAttachmentInput{Title: "Manual"}, "https://example.test/", c.mime)
		if name != "Manual"+c.extension {
			t.Fatalf("unsafe file name: %s", name)
		}
	}
	for _, c := range []struct{ url, title, mime, want string }{
		{"https://example.test/photo.jpg", "", "image/png", "photo.png"},
		{"https://example.test/", "", "image/webp", "artikelbild.webp"},
		{"invalid-url", "Our locomotive", "image/jpeg", "Our-locomotive.jpg"},
	} {
		if got := remoteImageFileName(application.VehicleImageInput{URL: c.url, Title: c.title}, c.mime); got != c.want {
			t.Fatalf("image name %q != %q", got, c.want)
		}
	}
	for _, c := range []struct{ name, category string }{{"spare parts.pdf", "Ersatzteilliste"}, {"manual.pdf", "Anleitung"}, {"other.pdf", "Dokumentation"}} {
		if got := attachmentCategoryForRemoteDocument(c.name, ""); got != c.category {
			t.Fatalf("category %s", got)
		}
	}
	if got := remoteAttachmentFileName(importVehicleAttachmentInput{}, "https://example.test/", "application/pdf"); !strings.HasSuffix(got, ".pdf") {
		t.Fatalf("missing fallback extension: %s", got)
	}

}

func TestAccessoryRemoteImageHTTPContractRetainsBytesAndRetryIdentity(t *testing.T) {
	fixture := newAccessoryAPIFixture(t, 1024)
	image := testPNG(t)
	content, status := image, 200
	var requests atomic.Int32
	syntheticRemoteMedia(t, func(*http.Request) *http.Response {
		requests.Add(1)
		return mediaResponse(status, "image/png", content)
	})
	path := "/api/v1/accessory-products/" + fixture.product.ID + "/documents/import-url"
	input := map[string]any{"url": "https://93.184.216.34/image.png", "title": "Original accessory", "idempotencyKey": "original-remote-image", "isPrimary": true}
	call := func(body any) *httptest.ResponseRecorder {
		return authedJSONRequest(t, fixture.router, fixture.sessions["editor"], "POST", path, body, true)
	}
	response := call(input)
	assertStatus(t, response, 201)
	id := responseID(t, response)
	assertStatus(t, call(input), 200)
	if requests.Load() != 1 {
		t.Fatal("retry downloaded remote image again")
	}
	response = authedJSONRequest(t, fixture.router, fixture.sessions["editor"], "GET", "/api/v1/accessory-products/"+fixture.product.ID+"/documents/"+id+"/download", nil, true)
	assertStatus(t, response, 200)
	if !bytes.Equal(response.Body.Bytes(), image) {
		t.Fatal("accessory import changed original bytes")
	}
	for index, c := range []struct {
		status int
		data   []byte
		want   int
		code   string
	}{
		{404, []byte("missing"), 502, "accessory_image_import_failed"},
		{200, []byte("not an image"), 415, "accessory_image_type_unsupported"},
		{200, bytes.Repeat([]byte("x"), 1025), 413, "accessory_document_too_large"},
	} {
		status, content = c.status, c.data
		input["idempotencyKey"] = fmt.Sprintf("failed-%d", index)
		assertProblem(t, call(input), c.want, c.code)
	}
}

func TestRemoteImageConflictsAndUnavailableBlobStorageRetainOriginalCollection(t *testing.T) {
	f := newContractFixture(t)
	id := f.vehicle(t)
	png := testPNG(t)
	syntheticRemoteMedia(t, func(*http.Request) *http.Response { return mediaResponse(200, "image/png", png) })
	base := "/api/v1/vehicles/" + id
	input := map[string]any{"url": "https://93.184.216.34/photo.png", "expectedImagesRevision": strings.Repeat("0", 64)}
	assertProblem(t, f.call(t, "POST", base+"/images/import-url", input), 409, "vehicle_image_conflict")
	assertProblem(t, f.call(t, "POST", base+"/images/import-url", map[string]any{}), 400, "image_url_missing")
	// Missing maintenance links are rejected after download and cannot leave uploaded blobs behind.
	input = map[string]any{"url": "https://93.184.216.34/photo.png", "maintenanceId": "missing-maintenance"}
	assertStatus(t, f.call(t, "POST", base+"/images/import-url", input), 400)
	f.config.FileBlobService = nil
	f.router = NewRouter(f.config)
	assertProblem(t, f.call(t, "POST", base+"/images/import-url", map[string]any{"url": "https://93.184.216.34/photo.png"}), 502, "image_import_failed")
	assertProblem(t, f.call(t, "POST", base+"/attachments/import-url", map[string]any{"url": "https://93.184.216.34/photo.png"}), 500, "attachment_import_failed")
	var count int
	if err := f.db.QueryRow(`SELECT COUNT(*) FROM file_blobs`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("rejected imports left blobs: %d %v", count, err)
	}
	vehicle, err := f.vehicles.Get(t.Context(), id)
	if err != nil || len(vehicle.Images) != 0 || len(vehicle.Attachments) != 0 {
		t.Fatalf("failed imports changed original vehicle: %v", err)
	}
}
