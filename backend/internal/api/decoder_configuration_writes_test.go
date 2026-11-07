package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
)

func decoderHTTPWritePaths(root string) []struct{ method, path string } {
	return []struct{ method, path string }{
		{http.MethodPost, root + "/instances"}, {http.MethodPut, root + "/instances/instance"},
		{http.MethodPost, root + "/installations"}, {http.MethodPost, root + "/installations/episode/removal"},
		{http.MethodPut, root + "/installations/episode/dates"}, {http.MethodPost, root + "/workspaces"},
		{http.MethodPut, root + "/workspaces/group/assignment"}, {http.MethodPost, root + "/workspaces/group/checkpoints"},
	}
}

func TestDecoderConfigurationWriteHTTPRequiresEditorCSRFAndCurrentConfirmation(t *testing.T) {
	db, root := testRouterDB(t), t.TempDir()
	auth, service := application.NewAuthService(db), application.NewVehicleService(db)
	router := NewRouter(Config{AuthService: auth, VehicleService: service, DataDir: root})
	vehicle, err := service.Create(t.Context(), application.CreateVehicleInput{Name: "HTTP decoder", Manufacturer: "Test", Gauge: "H0", Category: "Lokomotive", Gattung: "Diesellok"}, "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.CreateCVValue(t.Context(), vehicle.ID, application.VehicleCVValueInput{CVNumber: 1, Value: 3, DecoderProfile: "Existing"}); err != nil {
		t.Fatal(err)
	}
	oldCVs, err := service.ListCVValues(t.Context(), vehicle.ID)
	if err != nil {
		t.Fatal(err)
	}
	path := "/api/v1/vehicles/" + vehicle.ID + "/decoder-configurations"
	for _, endpoint := range decoderHTTPWritePaths(path) {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(endpoint.method, endpoint.path, nil))
		assertProblem(t, response, http.StatusForbidden, "csrf_required")
	}
	for index, role := range []string{"Viewer", "Messe", "Editor", "Admin"} {
		user, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{Username: "decoder-write-" + role, Password: "decoder-write-password", Roles: []string{role}})
		if err != nil {
			t.Fatal(err)
		}
		session := loginRouteTestUser(t, auth, "decoder-write-"+role, "decoder-write-password")
		before, err := service.DecoderConfigurationContext(t.Context(), vehicle.ID)
		if err != nil {
			t.Fatal(err)
		}
		for _, endpoint := range decoderHTTPWritePaths(path) {
			assertStatus(t, authedJSONRequest(t, router, session, endpoint.method, endpoint.path, nil, false), http.StatusForbidden)
			if role == "Viewer" || role == "Messe" {
				assertStatus(t, authedJSONRequest(t, router, session, endpoint.method, endpoint.path, nil, true), http.StatusForbidden)
			}
		}
		afterRejected, err := service.DecoderConfigurationContext(t.Context(), vehicle.ID)
		if err != nil || afterRejected.Revision != before.Revision {
			t.Fatal("rejected request changed decoder state")
		}
		if role == "Viewer" || role == "Messe" {
			continue
		}
		current := before
		confirm := func(method, endpoint string, input any, status int) {
			response := authedJSONRequest(t, router, session, method, endpoint, input, true)
			assertStatus(t, response, status)
			decodeResponse(t, response, &current)
			assertProblem(t, authedJSONRequest(t, router, session, method, endpoint, input, true), http.StatusConflict, "decoder_conflict")
		}
		identity := func(offset int) string { return fmt.Sprintf("%032x", index*16+offset) }
		instance := application.DecoderInstanceInput{ID: identity(1), ExpectedRevision: current.Revision, Label: "Physical " + role, Model: "Original model"}
		confirm(http.MethodPost, path+"/instances", instance, http.StatusCreated)
		installation := application.DecoderInstallationInput{ID: identity(2), ExpectedRevision: current.Revision, DecoderID: instance.ID, Notes: "Original episode notes"}
		confirm(http.MethodPost, path+"/installations", installation, http.StatusCreated)
		workspace := application.DecoderCVWorkspaceInput{ID: identity(3), ExpectedRevision: current.Revision, Profile: "Working " + role}
		confirm(http.MethodPost, path+"/workspaces", workspace, http.StatusCreated)
		confirm(http.MethodPut, path+"/workspaces/"+workspace.ID+"/assignment", application.DecoderCVAssignmentInput{ExpectedRevision: current.Revision, InstallationID: &installation.ID, SnapshotID: identity(4)}, http.StatusOK)
		confirm(http.MethodPost, path+"/workspaces/"+workspace.ID+"/checkpoints", application.DecoderCVCheckpointInput{ID: identity(5), ExpectedRevision: current.Revision, Note: "Saved partial state"}, http.StatusCreated)
		instance.ExpectedRevision, instance.Model = current.Revision, "Corrected model"
		confirm(http.MethodPut, path+"/instances/"+instance.ID, instance, http.StatusOK)
		installed, removed := "2026-10-20", "2026-10-21"
		confirm(http.MethodPut, path+"/installations/"+installation.ID+"/dates", application.DecoderInstallationDatesInput{ExpectedRevision: current.Revision, InstalledAt: &installed}, http.StatusOK)
		confirm(http.MethodPost, path+"/installations/"+installation.ID+"/removal", application.DecoderRemovalInput{ExpectedRevision: current.Revision, RemovedAt: &removed, Note: "Confirmed removal"}, http.StatusOK)
		confirm(http.MethodPut, path+"/workspaces/"+workspace.ID+"/assignment", application.DecoderCVAssignmentInput{ExpectedRevision: current.Revision, SnapshotID: identity(6)}, http.StatusOK)
		var audits int
		if err := db.QueryRow(`SELECT COUNT(*) FROM audit_logs WHERE action LIKE 'Decoder%' AND actor_user_id=?`, user.ID).Scan(&audits); err != nil || audits != 9 {
			t.Fatalf("missing or repeated confirmation audit: %d / %v", audits, err)
		}
	}
	values, err := service.ListCVValues(t.Context(), vehicle.ID)
	if err != nil || !reflect.DeepEqual(oldCVs, values) {
		t.Fatal("decoder HTTP commands changed unselected legacy CVs")
	}
}

func TestDecoderConfigurationWriteHTTPRejectsIncompleteNullUnknownTrailingAndOversizedEdits(t *testing.T) {
	db := testRouterDB(t)
	auth, service := application.NewAuthService(db), application.NewVehicleService(db)
	router := NewRouter(Config{AuthService: auth, VehicleService: service, DataDir: t.TempDir()})
	if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{Username: "decoder-json", Password: "decoder-json-password", Roles: []string{"Editor"}}); err != nil {
		t.Fatal(err)
	}
	session := loginRouteTestUser(t, auth, "decoder-json", "decoder-json-password")
	vehicle, err := service.Create(t.Context(), application.CreateVehicleInput{Name: "JSON decoder", Manufacturer: "Test", Gauge: "H0", Category: "Lokomotive", Gattung: "Diesellok"}, "")
	if err != nil {
		t.Fatal(err)
	}
	before, err := service.DecoderConfigurationContext(t.Context(), vehicle.ID)
	if err != nil {
		t.Fatal(err)
	}
	root := "/api/v1/vehicles/" + vehicle.ID + "/decoder-configurations"
	for _, endpoint := range decoderHTTPWritePaths(root) {
		assertProblem(t, authedJSONRequest(t, router, session, endpoint.method, endpoint.path, nil, true), http.StatusBadRequest, "decoder_validation")
	}
	raw, err := json.Marshal(application.DecoderInstanceInput{ID: strings.Repeat("f", 32), ExpectedRevision: before.Revision, Label: "Preserved"})
	if err != nil {
		t.Fatal(err)
	}
	valid := string(raw)
	reject := func(method, path, body string) {
		request := httptest.NewRequest(method, path, strings.NewReader(body))
		request.AddCookie(&http.Cookie{Name: "rk_session", Value: session.SessionToken})
		request.Header.Set("X-CSRF-Token", session.CSRFToken)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		assertProblem(t, response, http.StatusBadRequest, "decoder_validation")
	}
	for _, body := range []string{
		"null", "[]", strings.Replace(valid, `"notes":""`, `"notes":null`, 1),
		strings.Replace(valid, `,"notes":""`, ``, 1), strings.Replace(valid, `"notes":""`, `"unknown":""`, 1),
		valid + `{}`, valid + strings.Repeat(" ", maxJSONBodyBytes),
	} {
		reject(http.MethodPost, root+"/instances", body)
	}
	reject(http.MethodPut, root+"/installations/episode/dates", `{"expectedRevision":"`+before.Revision+`","installedAt":null}`)
	reject(http.MethodPut, root+"/workspaces/group/assignment", `{"expectedRevision":"`+before.Revision+`","snapshotId":"`+strings.Repeat("e", 32)+`"}`)
	after, err := service.DecoderConfigurationContext(t.Context(), vehicle.ID)
	if err != nil || before.Revision != after.Revision {
		t.Fatal("invalid decoder edit changed saved data")
	}
}
