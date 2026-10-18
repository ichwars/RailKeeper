package api

import (
	"net/http"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestCVImportRequiresEditorAndCSRF(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	for _, role := range []string{"Editor", "Viewer", "Messe"} {
		if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{Username: "cv-" + role, Password: "cv-test-password", Roles: []string{role}}); err != nil {
			t.Fatal(err)
		}
	}
	service := application.NewVehicleService(db)
	vehicle, err := service.Create(t.Context(), application.CreateVehicleInput{Manufacturer: "Test", Name: "CV", Gauge: "TT", Category: "Lokomotive", Gattung: "Diesellok"}, "test")
	if err != nil {
		t.Fatal(err)
	}
	router := NewRouter(Config{AuthService: auth, VehicleService: service})
	body := application.VehicleCVImportInput{Values: []application.VehicleCVValueInput{{CVNumber: 1, Value: 3}}}
	path := "/api/v1/vehicles/" + vehicle.ID + "/cv-values/import"
	for _, role := range []string{"Editor", "Viewer", "Messe"} {
		session := loginRouteTestUser(t, auth, "cv-"+role, "cv-test-password")
		response := authedJSONRequest(t, router, session, http.MethodPost, path, body, true)
		want := http.StatusForbidden
		if role == "Editor" {
			want = http.StatusOK
		}
		assertStatus(t, response, want)
		if role == "Editor" {
			assertProblem(t, authedJSONRequest(t, router, session, http.MethodPost, path, body, false), http.StatusForbidden, "csrf_required")
			for _, row := range []map[string]any{
				{"cvNumber": 1},
				{"cvNumber": 1, "value": nil},
				{"cvNumber": 257, "value": 3, "cv31": 16, "cv32": 3},
				{"cvNumber": 1, "value": 3, "sourceLine": 0},
				{"cvNumber": 1, "value": 3, "sourceExportedAt": "2026-02-30T17:41:59"},
			} {
				invalid := map[string]any{"values": []map[string]any{row}}
				assertProblem(t, authedJSONRequest(t, router, session, http.MethodPost, path, invalid, true), http.StatusBadRequest, "cv_import_invalid")
			}
		}
	}
	entries, err := auth.ListAuditLog(t.Context(), 20)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.Action == "VehicleCVValuesImported" && entry.TargetID == vehicle.ID {
			return
		}
	}
	t.Fatal("missing CV import audit entry")
}
