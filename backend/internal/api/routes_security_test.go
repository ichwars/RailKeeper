package api

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"railkeeper/backend/internal/application"
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

func loginRouteTestUser(t *testing.T, auth *application.AuthService, username, password string) *application.LoginResult {
	t.Helper()
	result, err := auth.Login(t.Context(), application.LoginInput{Username: username, Password: password})
	if err != nil {
		t.Fatal(err)
	}
	return result
}
