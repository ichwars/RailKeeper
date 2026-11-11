package api

import (
	"net/http"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestVehicleReportDataRequiresViewerAndCSRF(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{
		Username: "report-viewer", Password: "viewer-report-password", Roles: []string{"Viewer"},
	}); err != nil {
		t.Fatal(err)
	}
	viewer := loginRouteTestUser(t, auth, "report-viewer", "viewer-report-password")
	router := NewRouter(Config{AuthService: auth, VehicleService: application.NewVehicleService(db)})
	const path = "/api/v1/vehicles/report-data"
	empty := map[string][]string{"vehicleIds": {}}
	assertProblem(t, authedJSONRequest(t, router, viewer, http.MethodPost, path, empty, false),
		http.StatusForbidden, "csrf_required")
	assertProblem(t, authedJSONRequest(t, router, viewer, http.MethodPost, path, struct{}{}, true),
		http.StatusBadRequest, "vehicle_ids_required")
	assertProblem(t, authedJSONRequest(t, router, viewer, http.MethodPost, path,
		map[string][]string{"vehicleIds": {"missing"}}, true), http.StatusNotFound, "vehicle_not_found")
	response := authedJSONRequest(t, router, viewer, http.MethodPost, path, empty, true)
	assertStatus(t, response, http.StatusOK)
	if response.Body.String() != "[]\n" {
		t.Fatalf("empty report data = %q", response.Body.String())
	}
}
