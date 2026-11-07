package api

import (
	"net/http"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestCVDraftHTTPRoleCSRFRequiredFieldsAndAtomicConflict(t *testing.T) {
	db := testRouterDB(t)
	auth, service := application.NewAuthService(db), application.NewVehicleService(db)
	vehicle, err := service.Create(t.Context(), application.CreateVehicleInput{Manufacturer: "Test", Name: "CV HTTP draft", Gauge: "TT", Category: "Lokomotive", Gattung: "Diesellok"}, "test")
	if err != nil {
		t.Fatal(err)
	}
	router := NewRouter(Config{AuthService: auth, VehicleService: service})
	path := "/api/v1/vehicles/" + vehicle.ID + "/cv-values/changes"
	valid := map[string]any{"changes": []any{map[string]any{"expectedRevision": "", "value": map[string]any{"cvNumber": 1, "value": 0}}}}
	for _, role := range []string{"Editor", "Viewer", "Messe"} {
		if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{Username: "draft-" + role, Password: "cv-draft-test-password", Roles: []string{role}}); err != nil {
			t.Fatal(err)
		}
		session := loginRouteTestUser(t, auth, "draft-"+role, "cv-draft-test-password")
		if role != "Editor" {
			assertStatus(t, authedJSONRequest(t, router, session, http.MethodPost, path, valid, true), http.StatusForbidden)
			continue
		}
		assertProblem(t, authedJSONRequest(t, router, session, http.MethodPost, path, valid, false), http.StatusForbidden, "csrf_required")
		assertStatus(t, authedJSONRequest(t, router, session, http.MethodPost, path, valid, true), http.StatusOK)
		assertProblem(t, authedJSONRequest(t, router, session, http.MethodPost, path, valid, true), http.StatusConflict, "cv_value_conflict")
		for _, value := range []map[string]any{
			{"value": map[string]any{"cvNumber": 2, "value": 3}},
			{"expectedRevision": nil, "value": map[string]any{"cvNumber": 2, "value": 3}},
			{"expectedRevision": "", "value": map[string]any{"cvNumber": 2}},
			{"expectedRevision": "", "value": map[string]any{"cvNumber": 2, "value": nil}},
			{"expectedRevision": "", "value": map[string]any{"cvNumber": 257, "value": 3, "indexCV31": 0}},
			{"expectedRevision": "", "value": map[string]any{"cvNumber": 2, "value": 3, "cv31": 0}},
			{"expectedRevision": "", "value": map[string]any{"cvNumber": 2, "value": 3}, "extra": true},
		} {
			assertProblem(t, authedJSONRequest(t, router, session, http.MethodPost, path,
				map[string]any{"changes": []any{value}}, true), http.StatusBadRequest, "cv_changes_invalid")
		}
		stored, err := service.ListCVValues(t.Context(), vehicle.ID)
		if err != nil || len(stored) != 1 || stored[0].Value != 0 {
			t.Fatalf("invalid requests changed rows: %+v %v", stored, err)
		}
		conflict := application.VehicleCVChangesInput{Changes: []application.VehicleCVChange{
			{ExpectedRevision: "", Value: &application.VehicleCVValueInput{CVNumber: 2, Value: 9}},
			{ID: stored[0].ID, ExpectedRevision: strings.Repeat("0", 64), Value: &application.VehicleCVValueInput{CVNumber: 1, Value: 7}},
		}}
		assertProblem(t, authedJSONRequest(t, router, session, http.MethodPost, path, conflict, true), http.StatusConflict, "cv_value_conflict")
		stored, err = service.ListCVValues(t.Context(), vehicle.ID)
		if err != nil || len(stored) != 1 || stored[0].Value != 0 || len(stored[0].History) != 0 {
			t.Fatalf("partial HTTP draft write: %+v %v", stored, err)
		}
		legacy := application.VehicleCVValueInput{CVNumber: 1, Value: 7, ExpectedRevision: &conflict.Changes[1].ExpectedRevision}
		assertProblem(t, authedJSONRequest(t, router, session, http.MethodPut, "/api/v1/vehicles/"+vehicle.ID+"/cv-values/"+stored[0].ID, legacy, true), http.StatusConflict, "cv_value_conflict")
		tooMany := make([]application.VehicleCVChange, application.MaxVehicleCVChanges+1)
		assertProblem(t, authedJSONRequest(t, router, session, http.MethodPost, path, application.VehicleCVChangesInput{Changes: tooMany}, true), http.StatusBadRequest, "cv_changes_invalid")
	}
	entries, err := auth.ListAuditLog(t.Context(), 50)
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, entry := range entries {
		if entry.Action == "VehicleCVChangesApplied" && entry.TargetID == vehicle.ID {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("successful draft audit entries = %d, want 1", count)
	}
}
