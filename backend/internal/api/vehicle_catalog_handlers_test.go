package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestVehicleCatalogBackfillHTTPRequiresAdminCSRFAndReplaysWithoutMoreChanges(t *testing.T) {
	db := testRouterDB(t)
	auth, service := application.NewAuthService(db), application.NewVehicleService(db)
	router := NewRouter(Config{AuthService: auth, VehicleService: service})
	created, err := service.Create(t.Context(), application.CreateVehicleInput{Manufacturer: "Piko", Name: "Legacy", Gauge: "H0", Category: "Lokomotive", Gattung: "Diesellok"}, "actor")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`DELETE FROM vehicle_article_links WHERE vehicle_id=?`, created.ID); err != nil {
		t.Fatal(err)
	}
	path := "/api/v1/catalog/legacy-backfill"
	anonymous := httptest.NewRecorder()
	router.ServeHTTP(anonymous, httptest.NewRequest(http.MethodGet, path, nil))
	assertStatus(t, anonymous, http.StatusUnauthorized)
	for _, role := range []string{"Viewer", "Messe", "Editor", "Admin"} {
		if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{Username: "catalog-" + role, Password: "catalog-test-password", Roles: []string{role}}); err != nil {
			t.Fatal(err)
		}
		session := loginRouteTestUser(t, auth, "catalog-"+role, "catalog-test-password")
		if role != "Admin" {
			assertStatus(t, authedJSONRequest(t, router, session, http.MethodGet, path, nil, false), http.StatusForbidden)
			assertStatus(t, authedJSONRequest(t, router, session, http.MethodPost, path, nil, true), http.StatusForbidden)
			continue
		}
		assertUnlinked := func() {
			response := authedJSONRequest(t, router, session, http.MethodGet, path, nil, false)
			assertStatus(t, response, http.StatusOK)
			var report application.VehicleCatalogBackfillReport
			decodeResponse(t, response, &report)
			if report.LinkedVehicles != 0 || len(report.Assignments) != 1 || report.Assignments[0].Authority != "unlinked" {
				t.Fatal("reading or rejected confirmation changed catalog relationships")
			}
		}
		assertUnlinked()
		assertProblem(t, authedJSONRequest(t, router, session, http.MethodPost, path, nil, false), http.StatusForbidden, "csrf_required")
		assertUnlinked()
		for range 2 {
			response := authedJSONRequest(t, router, session, http.MethodPost, path, nil, true)
			assertStatus(t, response, http.StatusOK)
			var report application.VehicleCatalogBackfillReport
			decodeResponse(t, response, &report)
			if report.Vehicles != 1 || report.LinkedVehicles != 1 || report.DifferentRepresentations != 0 {
				t.Fatalf("report=%#v", report)
			}
		}
	}
}

func TestVehicleCatalogContextHTTPKeepsInventoryRolesAndReadsUnlinkedTargets(t *testing.T) {
	db := testRouterDB(t)
	auth, service := application.NewAuthService(db), application.NewVehicleService(db)
	router := NewRouter(Config{AuthService: auth, VehicleService: service})
	vehicle, err := service.Create(t.Context(), application.CreateVehicleInput{Manufacturer: "Piko", Name: "Owned", Gauge: "H0", Category: "Lokomotive", Gattung: "Diesellok"}, "actor")
	if err != nil {
		t.Fatal(err)
	}
	set, err := service.CreateSet(t.Context(), application.CreateVehicleSetInput{
		Set:     application.VehicleSetInput{Manufacturer: "Roco", Name: "Set", Gauge: "H0", Category: "Wagen", Gattung: "Reisezugwagen"},
		Members: []application.CreateVehicleInput{{Name: "A"}, {Name: "B"}},
	}, "actor")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`DELETE FROM vehicle_article_links WHERE vehicle_id=?`, vehicle.ID); err != nil {
		t.Fatal(err)
	}
	path := "/api/v1/catalog/vehicle/" + vehicle.ID
	anonymous := httptest.NewRecorder()
	router.ServeHTTP(anonymous, httptest.NewRequest(http.MethodGet, path, nil))
	assertStatus(t, anonymous, http.StatusUnauthorized)
	for _, role := range []string{"Messe", "Viewer", "Editor", "Admin"} {
		if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{Username: "context-" + role, Password: "catalog-test-password", Roles: []string{role}}); err != nil {
			t.Fatal(err)
		}
		session := loginRouteTestUser(t, auth, "context-"+role, "catalog-test-password")
		for _, target := range []struct{ path, mode, name string }{
			{path, "unlinked", "Owned"}, {"/api/v1/catalog/set/" + set.ID, "catalog", "Set"},
		} {
			response := authedJSONRequest(t, router, session, http.MethodGet, target.path, nil, false)
			if role == "Messe" {
				assertStatus(t, response, http.StatusForbidden)
				continue
			}
			assertStatus(t, response, http.StatusOK)
			var result application.VehicleCatalogContext
			decodeResponse(t, response, &result)
			if result.Authority != target.mode || result.CurrentFields["name"] != target.name || len(result.CurrentFields) != 13 || len(result.Revision) != 64 {
				t.Fatalf("context: %#v", result)
			}
		}
		if role != "Messe" {
			assertProblem(t, authedJSONRequest(t, router, session, http.MethodGet, "/api/v1/catalog/vehicle/missing", nil, false), http.StatusNotFound, "catalog_not_found")
			assertProblem(t, authedJSONRequest(t, router, session, http.MethodGet, "/api/v1/catalog/other/"+vehicle.ID, nil, false), http.StatusBadRequest, "catalog_validation")
		}
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM vehicle_article_links WHERE vehicle_id=?`, vehicle.ID).Scan(&count); err != nil || count != 0 {
		t.Fatal("GET changed an unlinked assignment")
	}
}
