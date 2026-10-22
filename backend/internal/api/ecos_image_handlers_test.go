package api

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
	ecospkg "railkeeper/backend/internal/ecos"
)

func TestECoSImagePreparationErrorsRetainTheirSpecificPhase(t *testing.T) {
	app := &App{}
	for _, test := range []struct {
		err  error
		code string
	}{
		{ecospkg.ErrImageRead, "ecos_image_read"},
		{ecospkg.ErrImageInventory, "ecos_image_inventory"},
		{ecospkg.ErrImageUploadForm, "ecos_image_form"},
	} {
		response := httptest.NewRecorder()
		app.respondECoSImageError(response, fmt.Errorf("%w: %w", test.err, ecospkg.ErrImageWeb))
		assertProblem(t, response, http.StatusBadGateway, test.code)
	}
}

func TestECoSImageRoutesRequireAdminCSRFAndConfirmation(t *testing.T) {
	auth := application.NewAuthService(testRouterDB(t))
	for _, role := range []string{"Admin", "Editor", "Viewer"} {
		if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{Username: "images-" + role, Password: "test-password", Roles: []string{role}}); err != nil {
			t.Fatal(err)
		}
	}
	router := NewRouter(Config{AuthService: auth, DigitalCenterWorkspace: application.NewDigitalCenterWorkspaceService(nil, nil, nil, nil, nil, nil)})
	for _, role := range []string{"Editor", "Viewer"} {
		login := loginRouteTestUser(t, auth, "images-"+role, "test-password")
		for _, operation := range []struct{ method, path string }{{http.MethodGet, "options"}, {http.MethodPost, "preview"}, {http.MethodPost, "confirm"}} {
			response := authedJSONRequest(t, router, login, operation.method, "/api/v1/vehicles/vehicle-1/ecos-image/"+operation.path, map[string]any{"confirm": true}, true)
			if response.Code != http.StatusForbidden {
				t.Fatalf("%s %s: %d", role, operation.path, response.Code)
			}
		}
	}
	admin := loginRouteTestUser(t, auth, "images-Admin", "test-password")
	for _, path := range []string{"preview", "confirm"} {
		response := authedJSONRequest(t, router, admin, http.MethodPost, "/api/v1/vehicles/vehicle-1/ecos-image/"+path, map[string]any{"confirm": true}, false)
		assertProblem(t, response, http.StatusForbidden, "csrf_required")
	}
	response := authedJSONRequest(t, router, admin, http.MethodPost, "/api/v1/vehicles/vehicle-1/ecos-image/confirm", map[string]any{"token": "token", "confirm": false}, true)
	assertProblem(t, response, http.StatusBadRequest, "digital_center_confirmation_required")
	for _, input := range []map[string]any{{"token": "token", "confirm": true, "host": "8.8.8.8"}, {"token": "token", "confirm": true, "objectId": 1001}} {
		response := authedJSONRequest(t, router, admin, http.MethodPost, "/api/v1/vehicles/vehicle-1/ecos-image/confirm", input, true)
		assertProblem(t, response, http.StatusBadRequest, "ecos_image_input")
	}
}

func TestECoSImagePreviewRejectsMalformedExistingSlotBeforeDeviceAccess(t *testing.T) {
	app := &App{digitalCenterWorkspace: application.NewDigitalCenterWorkspaceService(nil, nil, nil, nil, nil, nil)}
	for _, values := range [][]string{{"-1"}, {"500"}, {""}, {"1", "2"}, {"1.5"}} {
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		_ = writer.WriteField("objectId", "1001")
		_ = writer.WriteField("description", "Test")
		for _, value := range values {
			_ = writer.WriteField("existingSlot", value)
		}
		file, _ := writer.CreateFormFile("image", "railkeeper.bmp")
		_, _ = file.Write([]byte("invalid"))
		_ = writer.Close()
		request := httptest.NewRequest(http.MethodPost, "/", &body)
		request.Header.Set("Content-Type", writer.FormDataContentType())
		response := httptest.NewRecorder()
		app.previewECoSImage(response, request)
		assertProblem(t, response, http.StatusBadRequest, "ecos_image_input")
	}
}

func TestECoSImagePreviewRejectsNonMultipartWithoutPanic(t *testing.T) {
	app := &App{digitalCenterWorkspace: application.NewDigitalCenterWorkspaceService(nil, nil, nil, nil, nil, nil)}
	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("objectId=1001&description=test"))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()
	app.previewECoSImage(response, request)
	assertProblem(t, response, http.StatusBadRequest, "ecos_image_input")
}
