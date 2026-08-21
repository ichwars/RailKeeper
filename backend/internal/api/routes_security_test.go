package api

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/infrastructure"
)

var routeParameterPattern = regexp.MustCompile(`\{[^}]+\}`)

func TestAPIRoutesDeclareAccess(t *testing.T) {
	publicWrites := map[string]bool{
		"POST /api/v1/setup/admin":                 true,
		"POST /api/v1/auth/login":                  true,
		"POST /api/v1/auth/password-reset":         true,
		"POST /api/v1/auth/password-reset/confirm": true,
		"POST /api/v1/auth/logout":                 true,
	}

	for _, route := range apiRouteSpecs() {
		if route.Handler == nil {
			t.Fatalf("route %s %s has no handler", route.Method, route.Path)
		}
		switch route.Access {
		case routeAccessPublic, routeAccessAdmin, routeAccessEditor, routeAccessViewer, routeAccessMesse,
			routeAccessPlanner, routeAccessEditorOrPlanner:
		default:
			t.Fatalf("route %s %s has invalid access %q", route.Method, route.Path, route.Access)
		}
		if route.Access == routeAccessPublic && route.Method != http.MethodGet && !publicWrites[route.Method+" "+route.Path] {
			t.Fatalf("public write route %s %s is not explicitly allowed", route.Method, route.Path)
		}
	}
}

func TestProtectedRoutesRejectUnauthorizedAndInsufficientRoles(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	for _, user := range []application.CreateUserInput{
		{Username: "viewer", Password: "viewer-password", Roles: []string{"Viewer"}},
		{Username: "messe", Password: "messe-password", Roles: []string{"Messe"}},
		{Username: "planner", Password: "planner-password", Roles: []string{"Planner"}},
	} {
		if _, err := auth.CreateUser(t.Context(), "", user); err != nil {
			t.Fatal(err)
		}
	}
	viewer := loginRouteTestUser(t, auth, "viewer", "viewer-password")
	messe := loginRouteTestUser(t, auth, "messe", "messe-password")
	planner := loginRouteTestUser(t, auth, "planner", "planner-password")
	insufficient := map[routeAccess]*application.LoginResult{
		routeAccessAdmin:           viewer,
		routeAccessEditor:          planner,
		routeAccessViewer:          messe,
		routeAccessMesse:           viewer,
		routeAccessPlanner:         viewer,
		routeAccessEditorOrPlanner: viewer,
	}

	app := &App{authService: auth, logger: slog.Default()}
	mux := http.NewServeMux()
	app.registerRoutes(mux)

	for _, route := range apiRouteSpecs() {
		if route.Access == routeAccessPublic {
			continue
		}
		path := routeParameterPattern.ReplaceAllString(route.Path, "test")
		t.Run(route.Method+" "+route.Path+" unauthenticated", func(t *testing.T) {
			response := httptest.NewRecorder()
			mux.ServeHTTP(response, httptest.NewRequest(route.Method, path, nil))
			if response.Code != http.StatusUnauthorized {
				t.Fatalf("expected 401, got %d: %s", response.Code, response.Body.String())
			}
		})

		t.Run(route.Method+" "+route.Path+" insufficient role", func(t *testing.T) {
			if route.Authorize != nil && strings.HasPrefix(route.Path, "/api/v1/data-transfer/") {
				return // Dedicated data-transfer tests cover the intentional Viewer-or-Messe policy.
			}
			response := httptest.NewRecorder()
			request := httptest.NewRequest(route.Method, path, nil)
			request.AddCookie(&http.Cookie{Name: "rk_session", Value: insufficient[route.Access].SessionToken})
			mux.ServeHTTP(response, request)
			if response.Code != http.StatusForbidden {
				t.Fatalf("expected 403, got %d: %s", response.Code, response.Body.String())
			}
		})
	}
}

func TestDataTransferJobCreationEnforcesViewerAndMesseScopes(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	for _, user := range []application.CreateUserInput{
		{Username: "viewer-transfer", Password: "viewer-password", Roles: []string{"Viewer"}},
		{Username: "messe-transfer", Password: "messe-password", Roles: []string{"Messe"}},
	} {
		if _, err := auth.CreateUser(t.Context(), "", user); err != nil {
			t.Fatal(err)
		}
	}
	service := application.NewDataTransferService(infrastructure.NewDataTransferRepository(db), t.TempDir())
	vehicleImport := createRouteSecurityProfile(t, service, "Vehicle import", application.TransferImport,
		application.TransferVehicles)
	exhibitionImport := createRouteSecurityProfile(t, service, "Exhibition import", application.TransferImport,
		application.TransferExhibitionLists)
	vehicleExport := createRouteSecurityProfile(t, service, "Vehicle export", application.TransferExport,
		application.TransferVehicles)
	exhibitionExport := createRouteSecurityProfile(t, service, "Exhibition export", application.TransferExport,
		application.TransferExhibitionLists)
	router := NewRouter(Config{AuthService: auth, DataTransferService: service})
	viewer := loginRouteTestUser(t, auth, "viewer-transfer", "viewer-password")
	messe := loginRouteTestUser(t, auth, "messe-transfer", "messe-password")

	viewerImport := layoutRequest(t, router, viewer, http.MethodPost, "/api/v1/data-transfer/jobs/import",
		map[string]any{"profileId": exhibitionImport.ID}, true)
	assertProblem(t, viewerImport, http.StatusForbidden, "forbidden")

	for _, test := range []struct {
		name      string
		path      string
		profileID string
		status    int
		code      string
	}{
		{"vehicle import", "/api/v1/data-transfer/jobs/import", vehicleImport.ID, http.StatusForbidden, "data_transfer_forbidden"},
		{"exhibition import", "/api/v1/data-transfer/jobs/import", exhibitionImport.ID, http.StatusCreated, ""},
		{"vehicle export", "/api/v1/data-transfer/jobs/export", vehicleExport.ID, http.StatusForbidden, "data_transfer_forbidden"},
		{"exhibition export", "/api/v1/data-transfer/jobs/export", exhibitionExport.ID, http.StatusCreated, ""},
	} {
		t.Run("Messe "+test.name, func(t *testing.T) {
			response := layoutRequest(t, router, messe, http.MethodPost, test.path,
				map[string]any{"profileId": test.profileID}, true)
			if test.code != "" {
				assertProblem(t, response, test.status, test.code)
				return
			}
			assertStatus(t, response, test.status)
		})
	}
}

func TestMasterDataRouteSecurityContractRemainsIndependentFromDataTransfer(t *testing.T) {
	expected := []struct {
		method        string
		path          string
		access        routeAccess
		hasAuthorizer bool
	}{
		{http.MethodGet, "/api/v1/master-data-all", routeAccessViewer, false},
		{http.MethodGet, "/api/v1/master-data/export", routeAccessAdmin, false},
		{http.MethodPost, "/api/v1/master-data/import", routeAccessAdmin, false},
		{http.MethodGet, "/api/v1/master-data/{type}", routeAccessViewer, true},
		{http.MethodPost, "/api/v1/master-data/{type}", routeAccessEditor, false},
		{http.MethodPut, "/api/v1/master-data/{type}/{key}", routeAccessEditor, false},
		{http.MethodPatch, "/api/v1/master-data/{type}/{key}/active", routeAccessEditor, false},
		{http.MethodDelete, "/api/v1/master-data/{type}/{key}", routeAccessEditor, false},
		{http.MethodGet, "/api/v1/master-data-relations", routeAccessViewer, false},
	}

	routes := apiRouteSpecs()
	for _, want := range expected {
		t.Run(want.method+" "+want.path, func(t *testing.T) {
			var matches []routeSpec
			for _, route := range routes {
				if route.Method == want.method && route.Path == want.path {
					matches = append(matches, route)
				}
			}
			if len(matches) != 1 {
				t.Fatalf("route matches = %d, want exactly 1", len(matches))
			}
			if matches[0].Access != want.access {
				t.Fatalf("access = %q, want %q", matches[0].Access, want.access)
			}
			if (matches[0].Authorize != nil) != want.hasAuthorizer {
				t.Fatalf("custom authorizer present = %t, want %t", matches[0].Authorize != nil, want.hasAuthorizer)
			}
		})
	}
}

func createRouteSecurityProfile(
	t *testing.T,
	service *application.DataTransferService,
	name string,
	direction application.TransferDirection,
	area application.TransferArea,
) application.DataTransferProfile {
	t.Helper()
	profile, err := service.CreateProfile(t.Context(), application.CreateDataTransferProfileInput{
		Name: name, Direction: direction, Format: application.TransferJSON,
		Areas: []application.TransferArea{area},
	}, "admin-1")
	if err != nil {
		t.Fatal(err)
	}
	return profile
}

func loginRouteTestUser(t *testing.T, auth *application.AuthService, username, password string) *application.LoginResult {
	t.Helper()
	result, err := auth.Login(t.Context(), application.LoginInput{Username: username, Password: password})
	if err != nil {
		t.Fatal(err)
	}
	return result
}
