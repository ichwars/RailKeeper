package api

import (
	"net/http"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestVehicleCreationIdentityHTTPKeepsRolesCSRFAndOneCreation(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	router := NewRouter(Config{AuthService: auth, VehicleService: application.NewVehicleService(db)})
	inputs := []struct {
		path string
		body map[string]any
	}{
		{"/api/v1/vehicles", map[string]any{"creationId": strings.Repeat("a", 32), "manufacturer": "Piko", "name": "Single", "gauge": "H0", "category": "Lokomotive", "gattung": "Diesellok"}},
		{"/api/v1/vehicle-sets", map[string]any{"creationId": strings.Repeat("b", 32), "set": map[string]any{"manufacturer": "Roco", "name": "Set", "gauge": "H0", "category": "Wagen", "gattung": "Reisezugwagen"}, "members": []any{map[string]any{"name": "A"}, map[string]any{"name": "B"}}}},
	}
	for _, role := range []string{"Viewer", "Messe", "Editor"} {
		if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{Username: "creation-" + role, Password: "creation-test-password", Roles: []string{role}}); err != nil {
			t.Fatal(err)
		}
		session := loginRouteTestUser(t, auth, "creation-"+role, "creation-test-password")
		for _, input := range inputs {
			if role != "Editor" {
				assertStatus(t, authedJSONRequest(t, router, session, http.MethodPost, input.path, input.body, true), http.StatusForbidden)
				continue
			}
			assertProblem(t, authedJSONRequest(t, router, session, http.MethodPost, input.path, input.body, false), http.StatusForbidden, "csrf_required")
			assertStatus(t, authedJSONRequest(t, router, session, http.MethodPost, input.path, input.body, true), http.StatusCreated)
			assertProblem(t, authedJSONRequest(t, router, session, http.MethodPost, input.path, input.body, true), http.StatusConflict, "vehicle_creation_conflict")
			input.body["creationId"] = "invalid"
			assertProblem(t, authedJSONRequest(t, router, session, http.MethodPost, input.path, input.body, true), http.StatusBadRequest, "vehicle_creation_identity")
		}
	}
	var vehicles, sets, members int
	if err := db.QueryRow(`SELECT COUNT(*) FROM vehicles`).Scan(&vehicles); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM vehicle_sets`).Scan(&sets); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM vehicle_set_members`).Scan(&members); err != nil {
		t.Fatal(err)
	}
	if vehicles != 3 || sets != 1 || members != 2 {
		t.Fatalf("replayed creation: vehicles=%d sets=%d members=%d", vehicles, sets, members)
	}
}
