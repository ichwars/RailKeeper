package api

import (
	"bytes"
	"database/sql"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/domain"
	"railkeeper/backend/internal/infrastructure"
)

type accessoryAPIFixture struct {
	db            *sql.DB
	router        http.Handler
	sessions      map[string]*application.LoginResult
	product       *application.AccessoryProduct
	hybridProduct *application.AccessoryProduct
	locationA     *application.StorageLocation
	locationB     *application.StorageLocation
	layout        *application.Layout
}

func TestAccessoryDocumentRoutesRoundTripMultipartAndBlobDownload(t *testing.T) {
	fixture := newAccessoryAPIFixture(t, 1024*1024)
	editor := fixture.sessions["editor"]
	path := "/api/v1/accessory-products/" + fixture.product.ID + "/documents"
	png := append([]byte("\x89PNG\r\n\x1a\n"), bytes.Repeat([]byte{0}, 64)...)

	upload := accessoryMultipartRequest(t, fixture.router, editor, http.MethodPost, path, "photo.png", png,
		map[string]string{"category": "image", "description": " Front view ", "isPrimary": "true"}, true)
	assertStatus(t, upload, http.StatusCreated)
	payload := append([]byte(nil), upload.Body.Bytes()...)
	if bytes.Contains(payload, []byte(`"fileBlobId"`)) {
		t.Fatalf("public document response exposes internal blob id: %s", payload)
	}
	var document application.AccessoryDocument
	decodeResponse(t, upload, &document)
	if document.ProductID != fixture.product.ID || document.MimeType != "image/png" ||
		document.Description != "Front view" || !document.IsPrimary {
		t.Fatalf("unexpected document: %#v", document)
	}

	list := layoutRequest(t, fixture.router, editor, http.MethodGet, path, nil, true)
	assertStatus(t, list, http.StatusOK)
	var documents []application.AccessoryDocument
	decodeResponse(t, list, &documents)
	if len(documents) != 1 || documents[0].ID != document.ID {
		t.Fatalf("unexpected document list: %#v", documents)
	}

	documentPath := path + "/" + document.ID
	get := layoutRequest(t, fixture.router, editor, http.MethodGet, documentPath, nil, true)
	assertStatus(t, get, http.StatusOK)
	update := layoutRequest(t, fixture.router, editor, http.MethodPut, documentPath,
		map[string]any{"category": "image", "description": "Updated", "isPrimary": true}, true)
	assertStatus(t, update, http.StatusOK)

	download := layoutRequest(t, fixture.router, editor, http.MethodGet, documentPath+"/download", nil, true)
	assertStatus(t, download, http.StatusOK)
	if got := download.Header().Get("Content-Type"); got != "image/png" {
		t.Fatalf("Content-Type=%q, want image/png", got)
	}
	if got := download.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("X-Content-Type-Options=%q, want nosniff", got)
	}
	if disposition := download.Header().Get("Content-Disposition"); !strings.HasPrefix(disposition, "inline;") || strings.Contains(disposition, "\\") ||
		strings.Contains(disposition, "/") {
		t.Fatalf("unsafe image content disposition: %q", disposition)
	}
	if !bytes.Equal(download.Body.Bytes(), png) {
		t.Fatalf("downloaded bytes differ: %x", download.Body.Bytes())
	}

	var blobID string
	if err := fixture.db.QueryRow(`SELECT file_blob_id FROM accessory_documents WHERE id=?`, document.ID).
		Scan(&blobID); err != nil {
		t.Fatal(err)
	}
	deleted := layoutRequest(t, fixture.router, editor, http.MethodDelete, documentPath, nil, true)
	assertStatus(t, deleted, http.StatusNoContent)
	var blobCount int
	if err := fixture.db.QueryRow(`SELECT COUNT(*) FROM file_blobs WHERE id=?`, blobID).Scan(&blobCount); err != nil {
		t.Fatal(err)
	}
	if blobCount != 0 {
		t.Fatal("unreferenced document blob was not deleted")
	}
}

func TestAccessoryDocumentUploadMapsMultipartResourceLimitTo413(t *testing.T) {
	fixture := newAccessoryAPIFixture(t, 1024*1024)
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for index := 0; index < 1001; index++ {
		if err := writer.WriteField(fmt.Sprintf("field-%d", index), "x"); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	response := rawAccessoryRequest(t, fixture.router, fixture.sessions["editor"], http.MethodPost,
		"/api/v1/accessory-products/"+fixture.product.ID+"/documents",
		writer.FormDataContentType(), body.Bytes(), true)
	assertProblem(t, response, http.StatusRequestEntityTooLarge, "accessory_document_too_large")
}

func TestAccessoryDocumentUploadNormalizesContentAwareTextFormats(t *testing.T) {
	fixture := newAccessoryAPIFixture(t, 1024*1024)
	path := "/api/v1/accessory-products/" + fixture.product.ID + "/documents"
	tests := []struct {
		name     string
		fileName string
		data     []byte
		want     int
		mimeType string
	}{
		{name: "safe JSON", fileName: "decoder.json", data: []byte(`{"address": 3}`),
			want: http.StatusCreated, mimeType: "application/json"},
		{name: "safe CSV", fileName: "stock.csv", data: []byte("name,quantity\nrail,3\n"),
			want: http.StatusCreated, mimeType: "text/csv"},
		{name: "safe XML without declaration", fileName: "decoder.xml", data: []byte("<decoder><address>3</address></decoder>"),
			want: http.StatusCreated, mimeType: "application/xml"},
		{name: "malformed JSON", fileName: "broken.json", data: []byte(`{"address":`),
			want: http.StatusUnsupportedMediaType},
		{name: "malformed CSV", fileName: "broken.csv", data: []byte("name,value\nrail,\"open\n"),
			want: http.StatusUnsupportedMediaType},
		{name: "malformed XML", fileName: "broken.xml", data: []byte("<decoder>"),
			want: http.StatusUnsupportedMediaType},
		{name: "JSON with CSV extension", fileName: "wrong.csv", data: []byte(`{"address": 3}`),
			want: http.StatusUnsupportedMediaType},
		{name: "CSV with JSON extension", fileName: "wrong.json", data: []byte("name,value\nrail,3\n"),
			want: http.StatusUnsupportedMediaType},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			response := accessoryMultipartRequest(t, fixture.router, fixture.sessions["editor"], http.MethodPost,
				path, testCase.fileName, testCase.data, map[string]string{"category": "data_sheet"}, true)
			assertStatus(t, response, testCase.want)
			if testCase.want != http.StatusCreated {
				return
			}
			var document application.AccessoryDocument
			decodeResponse(t, response, &document)
			if document.MimeType != testCase.mimeType {
				t.Fatalf("mimeType=%q, want %q", document.MimeType, testCase.mimeType)
			}
		})
	}
}

func TestAccessoryDocumentUploadRejectsInvalidUTF8JSONBeforeBlobStorage(t *testing.T) {
	fixture := newAccessoryAPIFixture(t, 1024*1024)
	var before int
	if err := fixture.db.QueryRow(`SELECT COUNT(*) FROM file_blobs`).Scan(&before); err != nil {
		t.Fatal(err)
	}
	invalidUTF8JSON := append([]byte(`{"value":"`), 0xff)
	invalidUTF8JSON = append(invalidUTF8JSON, []byte(`"}`)...)
	response := accessoryMultipartRequest(t, fixture.router, fixture.sessions["editor"], http.MethodPost,
		"/api/v1/accessory-products/"+fixture.product.ID+"/documents", "invalid.json", invalidUTF8JSON,
		map[string]string{"category": "data_sheet"}, true)
	assertProblem(t, response, http.StatusUnsupportedMediaType, "accessory_document_type_unsupported")
	var after int
	if err := fixture.db.QueryRow(`SELECT COUNT(*) FROM file_blobs`).Scan(&after); err != nil {
		t.Fatal(err)
	}
	if after != before {
		t.Fatalf("file blob count changed after invalid UTF-8 JSON: before=%d after=%d", before, after)
	}
}

func TestAccessoryDocumentUploadCleansBlobWhenProductIsMissing(t *testing.T) {
	fixture := newAccessoryAPIFixture(t, 1024*1024)
	var before int
	if err := fixture.db.QueryRow(`SELECT COUNT(*) FROM file_blobs`).Scan(&before); err != nil {
		t.Fatal(err)
	}
	response := accessoryMultipartRequest(t, fixture.router, fixture.sessions["editor"], http.MethodPost,
		"/api/v1/accessory-products/missing/documents", "manual.pdf", []byte("%PDF-1.7\nexample"),
		map[string]string{"category": "manual"}, true)
	assertProblem(t, response, http.StatusNotFound, "accessory_not_found")
	var after int
	if err := fixture.db.QueryRow(`SELECT COUNT(*) FROM file_blobs`).Scan(&after); err != nil {
		t.Fatal(err)
	}
	if after != before {
		t.Fatalf("file blob count changed after failed metadata creation: before=%d after=%d", before, after)
	}
}

func TestAccessoryDocumentUploadRejectsOversizeUnsafeAndMalformedRequests(t *testing.T) {
	fixture := newAccessoryAPIFixture(t, 32)
	editor := fixture.sessions["editor"]
	path := "/api/v1/accessory-products/" + fixture.product.ID + "/documents"

	oversize := accessoryMultipartRequest(t, fixture.router, editor, http.MethodPost, path, "manual.pdf",
		append([]byte("%PDF-1.7\n"), bytes.Repeat([]byte("x"), 64)...),
		map[string]string{"category": "manual"}, true)
	assertProblem(t, oversize, http.StatusRequestEntityTooLarge, "accessory_document_too_large")

	unsafe := accessoryMultipartRequest(t, fixture.router, editor, http.MethodPost, path, "run.exe",
		[]byte("MZ executable"), map[string]string{"category": "other"}, true)
	assertProblem(t, unsafe, http.StatusUnsupportedMediaType, "accessory_document_type_unsupported")

	badCategory := accessoryMultipartRequest(t, fixture.router, editor, http.MethodPost, path, "manual.pdf",
		[]byte("%PDF-1.7\nexample"), map[string]string{"category": "unknown"}, true)
	assertProblem(t, badCategory, http.StatusBadRequest, "accessory_document_metadata_invalid")

	badPrimary := accessoryMultipartRequest(t, fixture.router, editor, http.MethodPost, path, "manual.pdf",
		[]byte("%PDF-1.7\nexample"), map[string]string{"category": "manual", "isPrimary": "perhaps"}, true)
	assertProblem(t, badPrimary, http.StatusBadRequest, "accessory_document_metadata_invalid")

	malformedJSON := rawAccessoryRequest(t, fixture.router, editor, http.MethodPut,
		path+"/missing", "application/json", []byte(`{"category":`), true)
	assertProblem(t, malformedJSON, http.StatusBadRequest, "invalid_json")
}

func TestAccessoryDocumentDownloadUsesAttachmentForNonImagesAndChecksProduct(t *testing.T) {
	fixture := newAccessoryAPIFixture(t, 1024*1024)
	editor := fixture.sessions["editor"]
	path := "/api/v1/accessory-products/" + fixture.product.ID + "/documents"
	upload := accessoryMultipartRequest(t, fixture.router, editor, http.MethodPost, path, "manual.pdf",
		[]byte("%PDF-1.7\nexample"), map[string]string{"category": "manual"}, true)
	assertStatus(t, upload, http.StatusCreated)
	var document application.AccessoryDocument
	decodeResponse(t, upload, &document)

	downloadPath := path + "/" + document.ID + "/download"
	download := layoutRequest(t, fixture.router, editor, http.MethodGet, downloadPath, nil, true)
	assertStatus(t, download, http.StatusOK)
	if disposition := download.Header().Get("Content-Disposition"); !strings.HasPrefix(disposition, "attachment;") {
		t.Fatalf("PDF must download as attachment, got %q", disposition)
	}

	wrongProduct := layoutRequest(t, fixture.router, editor, http.MethodGet,
		"/api/v1/accessory-products/"+fixture.hybridProduct.ID+"/documents/"+document.ID+"/download", nil, true)
	assertProblem(t, wrongProduct, http.StatusNotFound, "accessory_not_found")
}

func newAccessoryAPIFixture(t *testing.T, maxAttachmentBytes int64) *accessoryAPIFixture {
	t.Helper()
	dataDir := t.TempDir()
	db := testRouterDBWithDataDir(t, dataDir)
	auth := application.NewAuthService(db)
	for _, user := range []application.CreateUserInput{
		{Username: "admin", Password: "admin-password", Roles: []string{"Admin"}},
		{Username: "editor", Password: "editor-password", Roles: []string{"Editor"}},
		{Username: "viewer", Password: "viewer-password", Roles: []string{"Viewer"}},
		{Username: "planner", Password: "planner-password", Roles: []string{"Planner"}},
		{Username: "messe", Password: "messe-password", Roles: []string{"Messe"}},
	} {
		if _, err := auth.CreateUser(t.Context(), "", user); err != nil {
			t.Fatal(err)
		}
	}
	sessions := map[string]*application.LoginResult{
		"admin":   loginRouteTestUser(t, auth, "admin", "admin-password"),
		"editor":  loginRouteTestUser(t, auth, "editor", "editor-password"),
		"viewer":  loginRouteTestUser(t, auth, "viewer", "viewer-password"),
		"planner": loginRouteTestUser(t, auth, "planner", "planner-password"),
		"messe":   loginRouteTestUser(t, auth, "messe", "messe-password"),
	}

	repository := infrastructure.NewAccessoryRepository(db)
	accessories := application.NewAccessoryService(repository)
	product, err := accessories.CreateProduct(t.Context(), typedAccessoryProductInput(
		"Tillig", "83101", "Gerades Modellgleis", domain.AccessoryInventoryQuantity,
	), "seed")
	if err != nil {
		t.Fatal(err)
	}
	hybridProduct, err := accessories.CreateProduct(t.Context(), typedAccessoryProductInput(
		"Tillig", "85110", "Flexgleis", domain.AccessoryInventoryQuantityLaterIndividual,
	), "seed")
	if err != nil {
		t.Fatal(err)
	}
	locationA, err := accessories.CreateLocation(t.Context(), application.CreateStorageLocationInput{Name: "Shelf A"}, "seed")
	if err != nil {
		t.Fatal(err)
	}
	locationB, err := accessories.CreateLocation(t.Context(), application.CreateStorageLocationInput{Name: "Shelf B"}, "seed")
	if err != nil {
		t.Fatal(err)
	}
	for _, productID := range []string{product.ID, hybridProduct.ID} {
		if _, err := accessories.AdjustStock(t.Context(), productID, application.StockAdjustmentInput{
			LocationID: locationA.ID, Delta: 20,
		}, "seed"); err != nil {
			t.Fatal(err)
		}
	}
	layouts := application.NewLayoutService(infrastructure.NewLayoutRepository(db))
	layout, err := layouts.CreateLayout(t.Context(), application.CreateLayoutInput{
		Name: "Heimanlage", Kind: domain.LayoutKindPrivate, Gauge: "TT", Scale: "1:120",
	}, "seed")
	if err != nil {
		t.Fatal(err)
	}
	blobs := application.NewFileBlobService(db, dataDir)
	documents := application.NewAccessoryDocumentService(repository, blobs)
	router := NewRouter(Config{
		AuthService: auth, FileBlobService: blobs, MaxAttachmentBytes: maxAttachmentBytes,
		AccessoryService: accessories, AccessoryAllocationService: application.NewAccessoryAllocationService(repository),
		AccessoryDocumentService: documents,
	})
	return &accessoryAPIFixture{
		db: db, router: router, sessions: sessions, product: product, hybridProduct: hybridProduct,
		locationA: locationA, locationB: locationB, layout: layout,
	}
}

func typedAccessoryProductInput(
	manufacturer, articleNumber, name string,
	strategy domain.AccessoryInventoryStrategy,
) application.CreateAccessoryProductInput {
	return application.CreateAccessoryProductInput{
		Manufacturer: manufacturer, ArticleNumber: articleNumber, Name: name,
		Category: "Gleismaterial", ArticleType: domain.AccessoryArticleTrack, Subtype: "straight",
		Gauges: []string{"TT"}, PackageQuantity: 1, StockUnit: "piece", InventoryStrategy: strategy,
	}
}

func accessoryMultipartRequest(
	t *testing.T,
	router http.Handler,
	session *application.LoginResult,
	method, path, fileName string,
	data []byte,
	fields map[string]string,
	withCSRF bool,
) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for name, value := range fields {
		if err := writer.WriteField(name, value); err != nil {
			t.Fatal(err)
		}
	}
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(data); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return rawAccessoryRequest(t, router, session, method, path, writer.FormDataContentType(), body.Bytes(), withCSRF)
}

func rawAccessoryRequest(
	t *testing.T,
	router http.Handler,
	session *application.LoginResult,
	method, path, contentType string,
	body []byte,
	withCSRF bool,
) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, path, bytes.NewReader(body))
	request.Header.Set("Content-Type", contentType)
	request.AddCookie(&http.Cookie{Name: "rk_session", Value: session.SessionToken})
	if withCSRF {
		request.Header.Set("X-CSRF-Token", session.CSRFToken)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}
