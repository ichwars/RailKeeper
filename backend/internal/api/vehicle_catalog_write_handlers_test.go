package api

import (
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestVehicleCatalogWriteHTTPRequiresEditorAndCSRFAndRejectsRepeatedConfirmation(t *testing.T) {
	db := testRouterDB(t)
	auth, service := application.NewAuthService(db), application.NewVehicleService(db)
	router := NewRouter(Config{AuthService: auth, VehicleService: service})
	vehicle, err := service.Create(t.Context(), application.CreateVehicleInput{Manufacturer: "Piko", Name: "Owned", Gauge: "H0", Category: "Lokomotive", Gattung: "Diesellok", PurchasePrice: "42"}, "actor")
	if err != nil {
		t.Fatal(err)
	}
	path := "/api/v1/catalog/vehicle/" + vehicle.ID
	for _, suffix := range []string{"/revisions", "/adoption"} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, path+suffix, nil))
		assertProblem(t, response, http.StatusForbidden, "csrf_required")
	}
	for index, role := range []string{"Viewer", "Messe", "Editor", "Admin"} {
		if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{Username: "catalog-write-" + role, Password: "catalog-test-password", Roles: []string{role}}); err != nil {
			t.Fatal(err)
		}
		session := loginRouteTestUser(t, auth, "catalog-write-"+role, "catalog-test-password")
		before, err := service.CatalogContext(t.Context(), "vehicle", vehicle.ID)
		if err != nil {
			t.Fatal(err)
		}
		for _, suffix := range []string{"/revisions", "/adoption"} {
			assertStatus(t, authedJSONRequest(t, router, session, http.MethodPost, path+suffix, nil, false), http.StatusForbidden)
			if role == "Viewer" || role == "Messe" {
				assertStatus(t, authedJSONRequest(t, router, session, http.MethodPost, path+suffix, nil, true), http.StatusForbidden)
			}
		}
		afterRejected, err := service.CatalogContext(t.Context(), "vehicle", vehicle.ID)
		if err != nil || afterRejected.Revision != before.Revision {
			t.Fatal("rejected HTTP write changed saved data")
		}
		if role == "Viewer" || role == "Messe" {
			continue
		}
		fields := maps.Clone(before.CurrentFields)
		fields["name"] = "Confirmed " + role
		input := application.VehicleCatalogRevisionInput{ID: fmt.Sprintf("%032x", index+1), ExpectedRevision: before.Revision,
			Fields: fields, SelectedFields: []string{"name"}}
		response := authedJSONRequest(t, router, session, http.MethodPost, path+"/revisions", input, true)
		assertStatus(t, response, http.StatusCreated)
		var created application.VehicleCatalogContext
		decodeResponse(t, response, &created)
		if created.CurrentFields["name"] != fields["name"] || created.Pinned == nil || created.Pinned.ID != input.ID {
			t.Fatal("confirmed HTTP response did not describe saved values")
		}
		assertProblem(t, authedJSONRequest(t, router, session, http.MethodPost, path+"/revisions", input, true), http.StatusConflict, "catalog_conflict")
		adoption := application.VehicleCatalogAdoptionInput{ExpectedRevision: created.Revision, RevisionID: before.Pinned.ID, SelectedFields: []string{"name"}}
		response = authedJSONRequest(t, router, session, http.MethodPost, path+"/adoption", adoption, true)
		assertStatus(t, response, http.StatusOK)
		var adopted application.VehicleCatalogContext
		decodeResponse(t, response, &adopted)
		if adopted.CurrentFields["name"] != "Owned" || adopted.Pinned.ID != before.Pinned.ID {
			t.Fatal("HTTP adoption did not use the confirmed revision")
		}
		adoption.ExpectedRevision = adopted.Revision
		adoption.RevisionID = "missing"
		assertProblem(t, authedJSONRequest(t, router, session, http.MethodPost, path+"/adoption", adoption, true), http.StatusNotFound, "catalog_not_found")
	}
	stored, err := service.Get(t.Context(), vehicle.ID)
	if err != nil || stored.PurchasePrice != "42" || stored.InventoryNumber != vehicle.InventoryNumber {
		t.Fatal("catalog HTTP command changed inventory identity or acquisition")
	}
}

func TestVehicleCatalogWriteHTTPRejectsNullUnknownTrailingAndOversizedRequestsWithoutChanges(t *testing.T) {
	db := testRouterDB(t)
	auth, service := application.NewAuthService(db), application.NewVehicleService(db)
	router := NewRouter(Config{AuthService: auth, VehicleService: service})
	if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{Username: "catalog-body", Password: "catalog-test-password", Roles: []string{"Editor"}}); err != nil {
		t.Fatal(err)
	}
	session := loginRouteTestUser(t, auth, "catalog-body", "catalog-test-password")
	vehicle, err := service.Create(t.Context(), application.CreateVehicleInput{Manufacturer: "Piko", Name: "Owned", Gauge: "H0", Category: "Lokomotive", Gattung: "Diesellok"}, "actor")
	if err != nil {
		t.Fatal(err)
	}
	before, err := service.CatalogContext(t.Context(), "vehicle", vehicle.ID)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(application.VehicleCatalogRevisionInput{ID: strings.Repeat("f", 32), ExpectedRevision: before.Revision, Fields: before.CurrentFields, SelectedFields: []string{}})
	if err != nil {
		t.Fatal(err)
	}
	valid := string(raw)
	for _, body := range []string{
		strings.Replace(valid, `"description":""`, `"description":null`, 1),
		strings.Replace(valid, `"fields":{`, `"unexpected":true,"fields":{`, 1),
		strings.Replace(valid, `"selectedFields":[]`, `"selectedFields":null`, 1),
		valid + `{}`, valid + strings.Repeat(" ", 1<<20),
	} {
		request := httptest.NewRequest(http.MethodPost, "/api/v1/catalog/vehicle/"+vehicle.ID+"/revisions", strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("X-CSRF-Token", session.CSRFToken)
		request.AddCookie(&http.Cookie{Name: "rk_session", Value: session.SessionToken})
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		assertProblem(t, response, http.StatusBadRequest, "catalog_validation")
		after, err := service.CatalogContext(t.Context(), "vehicle", vehicle.ID)
		if err != nil || after.Revision != before.Revision {
			t.Fatal("invalid catalog request changed saved data")
		}
	}
}
