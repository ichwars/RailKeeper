package api

import (
	"net/http"
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
}
