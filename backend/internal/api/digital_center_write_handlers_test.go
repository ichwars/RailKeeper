package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestDigitalCenterAddressConflictReturnsSafeStructuredDetails(t *testing.T) {
	response := httptest.NewRecorder()
	(&App{logger: slog.Default()}).respondDigitalCenterWorkspaceError(response,
		&application.DigitalCenterAddressConflictError{ObjectID: 2002, Name: "Other", Address: 18})

	if response.Code != http.StatusConflict {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	var body struct {
		Error   string `json:"error"`
		Details struct {
			ObjectID       int    `json:"objectId"`
			Name           string `json:"name"`
			DecoderAddress int    `json:"decoderAddress"`
		} `json:"details"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Error != "digital_center_address_conflict" || body.Details.ObjectID != 2002 ||
		body.Details.Name != "Other" || body.Details.DecoderAddress != 18 {
		t.Fatalf("body=%#v", body)
	}
}

func TestDigitalCenterWriteRoutesAreAdminOnly(t *testing.T) {
	want := map[string]bool{
		"/api/v1/digital-centers/read-sessions/{id}/items/{itemID}/write-preview": false,
		"/api/v1/digital-centers/read-sessions/{id}/items/{itemID}/write-confirm": false,
	}
	for _, route := range apiRouteSpecs() {
		if _, found := want[route.Path]; !found || route.Method != http.MethodPost {
			continue
		}
		if route.Access != routeAccessAdmin {
			t.Fatalf("route %s access = %q, want Admin", route.Path, route.Access)
		}
		want[route.Path] = true
	}
	for path, found := range want {
		if !found {
			t.Fatalf("missing Admin write route %s", path)
		}
	}
}

func TestDigitalCenterWriteRoutesRequireCSRFAndExplicitConfirmation(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{
		Username: "write-admin", Password: "admin-password", Roles: []string{"Admin"},
	}); err != nil {
		t.Fatal(err)
	}
	router := NewRouter(Config{
		AuthService: auth,
		DigitalCenterWorkspace: application.NewDigitalCenterWorkspaceService(
			nil, nil, nil, nil, nil, nil,
		),
	})
	admin := loginRouteTestUser(t, auth, "write-admin", "admin-password")
	path := "/api/v1/digital-centers/read-sessions/session-1/items/item-1/write-confirm"

	withoutCSRF := layoutRequest(t, router, admin, http.MethodPost, path,
		map[string]any{"token": "grant", "confirm": true}, false)
	assertProblem(t, withoutCSRF, http.StatusForbidden, "csrf_required")

	withoutConfirmation := layoutRequest(t, router, admin, http.MethodPost, path,
		map[string]any{"token": "grant", "confirm": false}, true)
	assertProblem(t, withoutConfirmation, http.StatusBadRequest, "digital_center_confirmation_required")
}

func TestDigitalCenterWriteGrantAndFreshnessErrorsAreConflicts(t *testing.T) {
	app := &App{logger: slog.Default()}
	tests := []struct {
		name string
		err  error
		code string
	}{
		{name: "read stale", err: application.ErrDigitalCenterReadNotFresh,
			code: "digital_center_read_not_fresh"},
		{name: "conflict", err: application.ErrDigitalCenterConflictUnresolved,
			code: "digital_center_conflict_unresolved"},
		{name: "preview stale", err: application.ErrDigitalCenterPreviewStale,
			code: "digital_center_write_preview_stale"},
		{name: "expired", err: application.ErrDigitalCenterGrantExpired,
			code: "digital_center_write_grant_conflict"},
		{name: "consumed", err: application.ErrDigitalCenterGrantConsumed,
			code: "digital_center_write_grant_conflict"},
		{name: "actor mismatch", err: application.ErrDigitalCenterGrantActorMismatch,
			code: "digital_center_write_grant_conflict"},
		{name: "request mismatch", err: application.ErrDigitalCenterGrantMismatch,
			code: "digital_center_write_grant_conflict"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			app.respondDigitalCenterWorkspaceError(response, test.err)
			assertProblem(t, response, http.StatusConflict, test.code)
		})
	}
}

func TestLegacyECoSConfirmedWriteRequiresWorkspaceGrantBeforeDependencies(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/api/v1/digital-centers/ecos/locomotives/sync",
		strings.NewReader(`{"host":"device.local","port":15471,"vehicleId":"vehicle-1","objectId":3,"confirm":true}`))
	response := httptest.NewRecorder()

	(&App{}).syncECoSLocomotive(response, request)

	assertProblem(t, response, http.StatusConflict, "digital_center_write_grant_required")
}

func TestDigitalCenterWriteInputErrorsStayClientVisible(t *testing.T) {
	app := &App{logger: slog.Default()}
	tests := []struct {
		err    error
		status int
		code   string
	}{
		{application.ErrDigitalCenterWriteFieldUnsupported, http.StatusBadRequest,
			"digital_center_write_field_unsupported"},
		{application.ErrDigitalCenterConfirmationRequired, http.StatusBadRequest,
			"digital_center_confirmation_required"},
		{application.ErrDigitalCenterWriteNoChanges, http.StatusConflict,
			"digital_center_write_no_changes"},
		{errors.New("wrapped: " + application.ErrDigitalCenterDeviceWrite.Error()), http.StatusInternalServerError,
			"digital_center_workspace_failed"},
		{application.ErrDigitalCenterDeviceWrite, http.StatusBadGateway, "ecos_sync_failed"},
		{application.ErrDigitalCenterLivePauseFailed, http.StatusBadGateway, "ecos_live_pause_failed"},
	}
	for _, test := range tests {
		response := httptest.NewRecorder()
		app.respondDigitalCenterWorkspaceError(response, test.err)
		assertProblem(t, response, test.status, test.code)
	}
}
