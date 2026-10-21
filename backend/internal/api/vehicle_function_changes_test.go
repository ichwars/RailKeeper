package api

import (
	"encoding/json"
	"net/http"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestVehicleFunctionDraftRequiresEditorAndCSRF(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	for _, role := range []string{"Editor", "Viewer", "Messe"} {
		if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{Username: "function-" + role, Password: "function-test-password", Roles: []string{role}}); err != nil {
			t.Fatal(err)
		}
	}
	s := application.NewVehicleService(db)
	input := application.CreateVehicleInput{Manufacturer: "Test", Name: "Draft", Gauge: "TT", Category: "Lokomotive", Gattung: "Diesellok"}
	v, err := s.Create(t.Context(), input, "actor")
	if err != nil {
		t.Fatal(err)
	}
	router := NewRouter(Config{AuthService: auth, VehicleService: s})
	input.FunctionChanges = []application.VehicleFunctionChange{{FunctionKey: "F3", Value: &application.VehicleFunctionInput{Name: "Shunting", FlagsVersion: 2}}}
	path := "/api/v1/vehicles/" + v.ID
	for _, role := range []string{"Viewer", "Messe", "Editor"} {
		session := loginRouteTestUser(t, auth, "function-"+role, "function-test-password")
		want := http.StatusForbidden
		if role == "Editor" {
			want = http.StatusOK
		}
		response := authedJSONRequest(t, router, session, http.MethodPut, path, input, true)
		assertStatus(t, response, want)
		if role != "Editor" {
			continue
		}
		var saved application.Vehicle
		if err := json.NewDecoder(response.Body).Decode(&saved); err != nil {
			t.Fatal(err)
		}
		if len(saved.Functions) != 1 || saved.Functions[0].Revision == "" {
			t.Fatal("missing saved functions/revision")
		}
		assertProblem(t, authedJSONRequest(t, router, session, http.MethodPut, path, input, false), http.StatusForbidden, "csrf_required")
		assertProblem(t, authedJSONRequest(t, router, session, http.MethodPut, path, input, true), http.StatusConflict, "vehicle_function_conflict")
		input.FunctionChanges[0].FunctionKey = "F32"
		assertProblem(t, authedJSONRequest(t, router, session, http.MethodPut, path, input, true), http.StatusBadRequest, "vehicle_function_validation")
	}
}
