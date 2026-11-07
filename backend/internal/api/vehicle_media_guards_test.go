package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestGuardedImageUploadCleansRejectedPayloadsAndReturnsCommittedRevision(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{Username: "image-scope-editor", Password: "media-test-password", Roles: []string{"Editor"}}); err != nil {
		t.Fatal(err)
	}
	s := application.NewVehicleService(db)
	v, err := s.Create(t.Context(), application.CreateVehicleInput{Manufacturer: "Test", Name: "Image scope", Gauge: "TT", Category: "Lokomotive", Gattung: "Diesellok"}, "actor")
	if err != nil {
		t.Fatal(err)
	}
	blobs := application.NewFileBlobService(db, "")
	router := NewRouter(Config{AuthService: auth, VehicleService: s, FileBlobService: blobs})
	session := loginRouteTestUser(t, auth, "image-scope-editor", "media-test-password")
	upload := func(revision string) *httptest.ResponseRecorder {
		t.Helper()
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		if err := writer.WriteField("expectedImagesRevision", revision); err != nil {
			t.Fatal(err)
		}
		part, err := writer.CreateFormFile("file", "scope.png")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := part.Write(testPNG(t)); err != nil {
			t.Fatal(err)
		}
		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}
		request := httptest.NewRequest(http.MethodPost, "/api/v1/vehicles/"+v.ID+"/images", &body)
		request.Header.Set("Content-Type", writer.FormDataContentType())
		request.Header.Set("X-CSRF-Token", session.CSRFToken)
		request.AddCookie(&http.Cookie{Name: "rk_session", Value: session.SessionToken})
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		return response
	}
	assertProblem(t, upload(strings.Repeat("0", 64)), http.StatusConflict, "vehicle_image_conflict")
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM file_blobs`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("rejected upload left payloads: %d, %v", count, err)
	}
	response := upload(v.ImagesRevision)
	assertStatus(t, response, http.StatusCreated)
	var created application.VehicleImageCreation
	if err := json.NewDecoder(response.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}
	current, err := s.Get(t.Context(), v.ID)
	if err != nil || created.ImagesRevision != current.ImagesRevision || len(current.Images) != 1 {
		t.Fatalf("invalid upload receipt: %v", err)
	}
}

func TestFileBlobCleanupSurvivesCancelledUploadRequest(t *testing.T) {
	db := testRouterDB(t)
	blobs := application.NewFileBlobService(db, "")
	id, err := blobs.Store(t.Context(), []byte("cancelled synthetic upload"))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	(&App{fileBlobs: blobs, logger: slog.Default()}).deleteFileBlobIfUnreferenced(ctx, id)
	if _, err := blobs.Load(t.Context(), id); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("cancelled request left an orphaned blob: %v", err)
	}
}

func TestCVFileRevisionHTTPGuard(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{Username: "cv-file-editor", Password: "media-test-password", Roles: []string{"Editor"}}); err != nil {
		t.Fatal(err)
	}
	s := application.NewVehicleService(db)
	v, err := s.Create(t.Context(), application.CreateVehicleInput{Manufacturer: "Test", Name: "CV file", Gauge: "TT", Category: "Lokomotive", Gattung: "Diesellok"}, "actor")
	if err != nil {
		t.Fatal(err)
	}
	file, err := s.CreateCVFile(t.Context(), v.ID, application.VehicleCVFileInput{FileName: "decoder.csv", OriginalName: "decoder.csv", StoragePath: "uploads/decoder.csv"})
	if err != nil {
		t.Fatal(err)
	}
	router := NewRouter(Config{AuthService: auth, VehicleService: s})
	session := loginRouteTestUser(t, auth, "cv-file-editor", "media-test-password")
	path := "/api/v1/vehicles/" + v.ID + "/cv-files/" + file.ID
	assertProblem(t, authedJSONRequest(t, router, session, http.MethodDelete, path+"?expectedRevision="+strings.Repeat("0", 64), nil, true), http.StatusConflict, "cv_file_conflict")
	assertProblem(t, authedJSONRequest(t, router, session, http.MethodDelete, path+"?expectedRevision=", nil, true), http.StatusBadRequest, "cv_file_invalid")
	assertProblem(t, authedJSONRequest(t, router, session, http.MethodDelete, path+"?expectedRevision="+file.Revision, nil, false), http.StatusForbidden, "csrf_required")
	assertStatus(t, authedJSONRequest(t, router, session, http.MethodDelete, path+"?expectedRevision="+file.Revision, nil, true), http.StatusNoContent)
}

func TestVehicleAndMediaRevisionHTTPGuards(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{Username: "media-editor", Password: "media-test-password", Roles: []string{"Editor"}}); err != nil {
		t.Fatal(err)
	}
	s := application.NewVehicleService(db)
	input := application.CreateVehicleInput{Manufacturer: "Test", Name: "Media", Gauge: "TT", Category: "Lokomotive", Gattung: "Diesellok"}
	v, err := s.Create(t.Context(), input, "actor")
	if err != nil {
		t.Fatal(err)
	}
	a, err := s.CreateAttachment(t.Context(), v.ID, application.VehicleAttachmentInput{FileName: "manual.txt", OriginalName: "manual.txt", StoragePath: "uploads/manual.txt"})
	if err != nil {
		t.Fatal(err)
	}
	router := NewRouter(Config{AuthService: auth, VehicleService: s})
	session := loginRouteTestUser(t, auth, "media-editor", "media-test-password")
	path := "/api/v1/vehicles/" + v.ID
	bad := strings.Repeat("0", 64)
	input.ExpectedRevision = &bad
	assertProblem(t, authedJSONRequest(t, router, session, http.MethodPut, path, input, true), http.StatusConflict, "vehicle_conflict")
	input.ExpectedRevision = &v.Revision
	input.ExpectedImagesRevision, input.Images = &bad, []application.VehicleImageInput{}
	assertProblem(t, authedJSONRequest(t, router, session, http.MethodPut, path, input, true), http.StatusConflict, "vehicle_image_conflict")
	attachmentPath := path + "/attachments/" + a.ID
	metadata := application.VehicleAttachmentUpdateInput{Description: "Stale", ExpectedRevision: &bad}
	assertProblem(t, authedJSONRequest(t, router, session, http.MethodPut, attachmentPath, metadata, true), http.StatusConflict, "attachment_conflict")
	assertProblem(t, authedJSONRequest(t, router, session, http.MethodDelete, attachmentPath+"?expectedRevision="+bad, nil, true), http.StatusConflict, "attachment_conflict")
	assertProblem(t, authedJSONRequest(t, router, session, http.MethodDelete, attachmentPath+"?expectedRevision=", nil, true), http.StatusBadRequest, "attachment_invalid")
	assertProblem(t, authedJSONRequest(t, router, session, http.MethodDelete, attachmentPath+"?expectedRevision="+a.Revision, nil, false), http.StatusForbidden, "csrf_required")
	assertStatus(t, authedJSONRequest(t, router, session, http.MethodDelete, attachmentPath+"?expectedRevision="+a.Revision, nil, true), http.StatusNoContent)
}

func TestMediaDeletionPreservesSharedLegacyFileUntilLastReference(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{Username: "legacy-editor", Password: "media-test-password", Roles: []string{"Editor"}}); err != nil {
		t.Fatal(err)
	}
	s := application.NewVehicleService(db)
	v, err := s.Create(t.Context(), application.CreateVehicleInput{Manufacturer: "Test", Name: "Legacy", Gauge: "TT", Category: "Lokomotive", Gattung: "Diesellok"}, "actor")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "uploads"), 0o700); err != nil {
		t.Fatal(err)
	}
	fullPath := filepath.Join(dir, "uploads", "shared.txt")
	payload := []byte("shared legacy payload")
	if err := os.WriteFile(fullPath, payload, 0o600); err != nil {
		t.Fatal(err)
	}
	attachments := []*application.VehicleAttachment{}
	paths := []string{"uploads/shared.txt", "uploads/./shared.txt"}
	if runtime.GOOS == "windows" {
		paths = append(paths, "uploads/SHARED.txt")
	}
	for _, storagePath := range paths {
		a, err := s.CreateAttachment(t.Context(), v.ID, application.VehicleAttachmentInput{FileName: "shared.txt", OriginalName: "shared.txt", StoragePath: storagePath})
		if err != nil {
			t.Fatal(err)
		}
		attachments = append(attachments, a)
	}
	image, err := s.CreateImage(t.Context(), v.ID, application.VehicleImageInput{FileName: "legacy.png", MimeType: "image/png", StoragePath: "uploads/shared.txt"})
	if err != nil {
		t.Fatal(err)
	}
	router := NewRouter(Config{AuthService: auth, VehicleService: s, DataDir: dir})
	session := loginRouteTestUser(t, auth, "legacy-editor", "media-test-password")
	root := "/api/v1/vehicles/" + v.ID
	for _, a := range attachments {
		assertStatus(t, authedJSONRequest(t, router, session, http.MethodDelete, root+"/attachments/"+a.ID, nil, true), http.StatusNoContent)
		if data, err := os.ReadFile(fullPath); err != nil || string(data) != string(payload) {
			t.Fatalf("referenced bytes were removed: %v", err)
		}
	}
	assertStatus(t, authedJSONRequest(t, router, session, http.MethodDelete, root+"/images/"+image.ID+"?expectedRevision="+image.Revision, nil, true), http.StatusNoContent)
	if _, err := os.Stat(fullPath); !os.IsNotExist(err) {
		t.Fatalf("unreferenced legacy file was not removed: %v", err)
	}
}
