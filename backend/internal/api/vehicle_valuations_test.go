package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/infrastructure/valuationdata"
	"railkeeper/backend/internal/valuation"
)

func TestValuationHTTPPermissionsCorrectionsAndExport(t *testing.T) {
	db := testRouterDB(t)
	auth, vehicles := application.NewAuthService(db), application.NewVehicleService(db)
	values := valuation.New(valuationdata.NewStore(db), func() time.Time { return time.Date(2026, 11, 7, 0, 0, 0, 0, time.UTC) })
	router := NewRouter(Config{AuthService: auth, VehicleService: vehicles, ValuationService: values})
	vehicle, err := vehicles.Create(t.Context(), application.CreateVehicleInput{
		Name: "Manual valuation", Manufacturer: "Test", Gauge: "H0", Category: "Lokomotive", Gattung: "Diesellok",
	}, "")
	if err != nil {
		t.Fatal(err)
	}
	path := "/api/v1/vehicles/" + vehicle.ID + "/valuations"
	for index, role := range []string{"Viewer", "Messe", "Editor", "Admin"} {
		_, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{Username: "valuation-" + role, Password: "synthetic-valuation-password", Roles: []string{role}})
		if err != nil {
			t.Fatal(err)
		}
		session := loginRouteTestUser(t, auth, "valuation-"+role, "synthetic-valuation-password")
		readStatus := http.StatusOK
		if role == "Messe" {
			readStatus = http.StatusForbidden
		}
		for _, endpoint := range []string{path, path + "/export", "/api/v1/valuations/summary?asOf=2026-11-07"} {
			assertStatus(t, authedJSONRequest(t, router, session, http.MethodGet, endpoint, nil, true), readStatus)
		}
		input := valuation.Input{ID: fmt.Sprintf("%032x", index+1), Amount: "123,456789", Currency: " eur ", ValuedOn: "2026-01-01", Source: "Manual estimate"}
		assertStatus(t, authedJSONRequest(t, router, session, http.MethodPost, path, input, false), http.StatusForbidden)
		if role == "Viewer" || role == "Messe" {
			assertStatus(t, authedJSONRequest(t, router, session, http.MethodPost, path, input, true), http.StatusForbidden)
			continue
		}
		created := authedJSONRequest(t, router, session, http.MethodPost, path, input, true)
		assertStatus(t, created, http.StatusCreated)
		var first valuation.Record
		if err := json.Unmarshal(created.Body.Bytes(), &first); err != nil || first.Amount != "123.456789" || first.Currency != "EUR" {
			t.Fatalf("exact normalized record: %+v %v", first, err)
		}
		input.Amount = "0"
		assertProblem(t, authedJSONRequest(t, router, session, http.MethodPost, path, input, true), http.StatusConflict, "valuation_conflict")
		input.ID, input.ReplacesID = fmt.Sprintf("%032x", index+100), first.ID
		assertStatus(t, authedJSONRequest(t, router, session, http.MethodPost, path, input, true), http.StatusCreated)
		assertStatus(t, authedJSONRequest(t, router, session, http.MethodPost, path, input, true), http.StatusCreated)
		input.ID = fmt.Sprintf("%032x", index+200)
		assertProblem(t, authedJSONRequest(t, router, session, http.MethodPost, path, input, true), http.StatusConflict, "valuation_conflict")
		for _, invalid := range []string{"-1", "1.1234567", "1e3"} {
			input.Amount = invalid
			assertProblem(t, authedJSONRequest(t, router, session, http.MethodPost, path, input, true), http.StatusBadRequest, "valuation_invalid")
		}
		assertStatus(t, authedJSONRequest(t, router, session, http.MethodPost, path, map[string]any{"unexpected": true}, true), http.StatusBadRequest)
		for _, invalid := range []string{path + "?offset=-1", path + "?offset=invalid", "/api/v1/valuations/summary?asOf=2026-02-29"} {
			assertStatus(t, authedJSONRequest(t, router, session, http.MethodGet, invalid, nil, true), http.StatusBadRequest)
		}
		exported := authedJSONRequest(t, router, session, http.MethodGet, path+"/export", nil, true)
		assertStatus(t, exported, http.StatusOK)
		var document valuation.ExportDocument
		if err := json.Unmarshal(exported.Body.Bytes(), &document); err != nil || document.Version != 1 || document.Format != "railkeeper-valuations" || len(document.Entries) != (index-1)*2 || exported.Header().Get("Content-Disposition") == "" {
			t.Fatalf("history export: %+v %v", document, err)
		}
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM vehicle_valuations`).Scan(&count); err != nil || count != 4 {
		t.Fatalf("unexpected writes: %d %v", count, err)
	}
}
