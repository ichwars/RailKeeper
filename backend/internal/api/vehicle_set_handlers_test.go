package api

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestVehicleSetRoutesEnforceReadWriteAndCSRFRules(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	for _, user := range []application.CreateUserInput{
		{Username: "set-editor", Password: "editor-password", Roles: []string{"Editor"}},
		{Username: "set-viewer", Password: "viewer-password", Roles: []string{"Viewer"}},
	} {
		if _, err := auth.CreateUser(t.Context(), "", user); err != nil {
			t.Fatal(err)
		}
	}
	service := application.NewVehicleService(db)
	created, err := service.CreateSet(t.Context(), application.CreateVehicleSetInput{
		Set: application.VehicleSetInput{
			Name: "Rheingold", Manufacturer: "Roco", Gauge: "H0", Category: "Wagen",
			Gattung: "Reisezugwagen",
		},
		Members: []application.CreateVehicleInput{{Name: "Wagen 1"}, {Name: "Wagen 2"}},
	}, "test-actor")
	if err != nil {
		t.Fatal(err)
	}
	router := NewRouter(Config{AuthService: auth, VehicleService: service})
	editor := loginRouteTestUser(t, auth, "set-editor", "editor-password")
	viewer := loginRouteTestUser(t, auth, "set-viewer", "viewer-password")
	path := "/api/v1/vehicle-sets/" + created.ID
	image, err := service.CreateImage(t.Context(), created.Members[0].ID, application.VehicleImageInput{
		FileName: "wagen.jpg", MimeType: "image/jpeg", StoragePath: "images/wagen.jpg", IsPrimary: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	update := map[string]any{
		"name": "Rheingold neu", "manufacturer": "Roco", "gauge": "H0", "category": "Wagen",
		"gattung": "Reisezugwagen",
	}

	assertStatus(t, layoutRequest(t, router, viewer, http.MethodGet, path, nil, false), http.StatusOK)
	assertProblem(t, layoutRequest(t, router, viewer, http.MethodPatch, path, update, true),
		http.StatusForbidden, "forbidden")
	assertProblem(t, layoutRequest(t, router, editor, http.MethodPatch, path, update, false),
		http.StatusForbidden, "csrf_required")
	updatedResponse := layoutRequest(t, router, editor, http.MethodPatch, path, update, true)
	assertStatus(t, updatedResponse, http.StatusOK)
	var updated application.VehicleSet
	decodeResponse(t, updatedResponse, &updated)
	if updated.Name != "Rheingold neu" || updated.InventoryNumber != created.InventoryNumber {
		t.Fatalf("unexpected updated set: %#v", updated)
	}
	assertProblem(t, layoutRequest(t, router, viewer, http.MethodGet,
		"/api/v1/vehicle-sets/missing", nil, false), http.StatusNotFound, "vehicle_set_not_found")
	mainImagePath := path + "/main-image"
	selection := map[string]any{"mode": "member", "memberImageId": image.ID}
	assertProblem(t, layoutRequest(t, router, viewer, http.MethodPut, mainImagePath, selection, true),
		http.StatusForbidden, "forbidden")
	assertProblem(t, layoutRequest(t, router, editor, http.MethodPut, mainImagePath, selection, false),
		http.StatusForbidden, "csrf_required")
	selectedResponse := layoutRequest(t, router, editor, http.MethodPut, mainImagePath, selection, true)
	assertStatus(t, selectedResponse, http.StatusOK)
	var selected application.VehicleSet
	decodeResponse(t, selectedResponse, &selected)
	if selected.MainImage == nil || selected.MainImage.ImageID != image.ID || selected.MainImage.Source != "member" {
		t.Fatalf("unexpected selected set image: %#v", selected.MainImage)
	}
}

func TestVehicleSetDedicatedImageUploadDownloadAndDelete(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{
		Username: "set-image-editor", Password: "editor-password", Roles: []string{"Editor"},
	}); err != nil {
		t.Fatal(err)
	}
	service := application.NewVehicleService(db)
	created, err := service.CreateSet(t.Context(), application.CreateVehicleSetInput{
		Set: application.VehicleSetInput{
			Name: "Rheingold", Manufacturer: "Roco", Gauge: "H0", Category: "Wagen",
			Gattung: "Reisezugwagen",
		},
		Members: []application.CreateVehicleInput{{Name: "Wagen 1"}, {Name: "Wagen 2"}},
	}, "test-actor")
	if err != nil {
		t.Fatal(err)
	}
	blobs := application.NewFileBlobService(db, "")
	router := NewRouter(Config{AuthService: auth, VehicleService: service, FileBlobService: blobs})
	editor := loginRouteTestUser(t, auth, "set-image-editor", "editor-password")

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "rheingold.png")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(testPNG(t)); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	imagePath := "/api/v1/vehicle-sets/" + created.ID + "/image"
	request := httptest.NewRequest(http.MethodPost, imagePath, &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set("X-CSRF-Token", editor.CSRFToken)
	request.AddCookie(&http.Cookie{Name: "rk_session", Value: editor.SessionToken})
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	assertStatus(t, response, http.StatusCreated)
	var uploaded application.VehicleSet
	decodeResponse(t, response, &uploaded)
	if uploaded.DedicatedImage == nil || uploaded.MainImage == nil || uploaded.MainImage.Source != "dedicated" {
		t.Fatalf("unexpected uploaded set image: %#v", uploaded)
	}

	download := layoutRequest(t, router, editor, http.MethodGet, imagePath+"/thumbnail", nil, false)
	assertStatus(t, download, http.StatusOK)
	if contentType := download.Header().Get("Content-Type"); contentType != "image/jpeg" {
		t.Fatalf("unexpected thumbnail content type %q", contentType)
	}
	assertStatus(t, layoutRequest(t, router, editor, http.MethodDelete, imagePath, nil, true), http.StatusNoContent)
	loaded, err := service.GetSet(t.Context(), created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.DedicatedImage != nil || loaded.MainImage != nil {
		t.Fatalf("dedicated image was not removed: %#v", loaded)
	}
}
