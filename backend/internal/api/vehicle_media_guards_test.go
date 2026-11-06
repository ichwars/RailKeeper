package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
)

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
	for _, storagePath := range []string{"uploads/shared.txt", "uploads/./shared.txt"} {
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
