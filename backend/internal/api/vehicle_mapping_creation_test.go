package api

import (
	"net/http"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestVehicleMappingCreationHTTPAtomicSourceRoleCSRFAndReplay(t *testing.T) {
	db := testRouterDB(t)
	auth, service := application.NewAuthService(db), application.NewVehicleService(db)
	router := NewRouter(Config{AuthService: auth, VehicleService: service})
	body := map[string]any{"manufacturer": "Piko", "name": "Mapped source", "gauge": "H0",
		"category": "Lokomotive", "gattung": "Diesellok", "externalMapping": map[string]any{
			"provider": "cs3", "externalId": "77", "externalName": "Source"},
		"functionChanges": []any{map[string]any{"functionKey": "F0", "expectedRevision": "", "value": map[string]any{"name": "Light", "mode": "dauer", "flagsVersion": 2}}}}
	for _, role := range []string{"Viewer", "Messe", "Editor"} {
		if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{Username: "mapping-" + role,
			Password: "mapping-test-password", Roles: []string{role}}); err != nil {
			t.Fatal(err)
		}
		session := loginRouteTestUser(t, auth, "mapping-"+role, "mapping-test-password")
		if role != "Editor" {
			assertStatus(t, authedJSONRequest(t, router, session, http.MethodPost, "/api/v1/vehicles", body, true), http.StatusForbidden)
			continue
		}
		assertProblem(t, authedJSONRequest(t, router, session, http.MethodPost, "/api/v1/vehicles", body, false), http.StatusForbidden, "csrf_required")
		created := authedJSONRequest(t, router, session, http.MethodPost, "/api/v1/vehicles", body, true)
		assertStatus(t, created, http.StatusCreated)
		var vehicle application.Vehicle
		decodeResponse(t, created, &vehicle)
		if len(vehicle.ExternalMappings) != 1 || vehicle.ExternalMappings[0].VehicleID != vehicle.ID || len(vehicle.Functions) != 1 {
			t.Fatalf("creation did not acknowledge core, mapping and functions together: %#v", vehicle)
		}
		body["name"] = "Repeated source"
		assertProblem(t, authedJSONRequest(t, router, session, http.MethodPost, "/api/v1/vehicles", body, true), http.StatusConflict, "external_mapping_conflict")
		var vehicles, functions, audits int
		if err := db.QueryRow(`SELECT COUNT(*) FROM vehicles`).Scan(&vehicles); err != nil {
			t.Fatal(err)
		}
		if err := db.QueryRow(`SELECT COUNT(*) FROM vehicle_functions`).Scan(&functions); err != nil {
			t.Fatal(err)
		}
		if err := db.QueryRow(`SELECT COUNT(*) FROM audit_logs WHERE action='VehicleCreated'`).Scan(&audits); err != nil {
			t.Fatal(err)
		}
		if vehicles != 1 || functions != 1 || audits != 1 {
			t.Fatalf("partial replay: vehicles=%d functions=%d audits=%d", vehicles, functions, audits)
		}
		stored := authedJSONRequest(t, router, session, http.MethodGet, "/api/v1/vehicles/"+vehicle.ID, nil, false)
		assertStatus(t, stored, http.StatusOK)
		var detail application.Vehicle
		decodeResponse(t, stored, &detail)
		if detail.Name != "Mapped source" || detail.ExternalMappings[0].ID != vehicle.ExternalMappings[0].ID {
			t.Fatalf("replay overwrote original vehicle: %#v", detail)
		}
	}
}
