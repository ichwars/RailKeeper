package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestECoSFunctionReadRouteRequiresAdmin(t *testing.T) {
	auth := application.NewAuthService(testRouterDB(t))
	router := NewRouter(Config{AuthService: auth,
		DigitalCenterWorkspace: application.NewDigitalCenterWorkspaceService(nil, nil, nil, nil, nil, nil)})
	path := "/api/v1/vehicles/vehicle-1/ecos-functions"
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
	if response.Code != http.StatusUnauthorized {
		t.Fatal(response.Code)
	}
	for _, role := range []string{"Editor", "Viewer", "Messe"} {
		_, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{
			Username: "functions-" + role, Password: "test-password", Roles: []string{role}})
		if err != nil {
			t.Fatal(err)
		}
		login := loginRouteTestUser(t, auth, "functions-"+role, "test-password")
		result := authedJSONRequest(t, router, login, http.MethodGet, path, nil, true)
		if result.Code != http.StatusForbidden {
			t.Fatal(role, result.Code)
		}
	}
}
