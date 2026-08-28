package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/domain"
	"railkeeper/backend/internal/infrastructure"
)

func TestLayoutRoutesEnforceRoleAndCSRFBoundaries(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	layouts := application.NewLayoutService(infrastructure.NewLayoutRepository(db))
	for _, user := range []application.CreateUserInput{
		{Username: "admin", Password: "admin-password", Roles: []string{"Admin"}},
		{Username: "editor", Password: "editor-password", Roles: []string{"Editor"}},
		{Username: "viewer", Password: "viewer-password", Roles: []string{"Viewer"}},
		{Username: "planner", Password: "planner-password", Roles: []string{"Planner"}},
		{Username: "messe", Password: "messe-password", Roles: []string{"Messe"}},
	} {
		if _, err := auth.CreateUser(t.Context(), "", user); err != nil {
			t.Fatal(err)
		}
	}
	router := NewRouter(Config{AuthService: auth, LayoutService: layouts})

	sessions := map[string]*application.LoginResult{
		"admin":   loginRouteTestUser(t, auth, "admin", "admin-password"),
		"editor":  loginRouteTestUser(t, auth, "editor", "editor-password"),
		"viewer":  loginRouteTestUser(t, auth, "viewer", "viewer-password"),
		"planner": loginRouteTestUser(t, auth, "planner", "planner-password"),
		"messe":   loginRouteTestUser(t, auth, "messe", "messe-password"),
	}
	for role, session := range sessions {
		response := layoutRequest(t, router, session, http.MethodGet, "/api/v1/layouts", nil, true)
		want := http.StatusOK
		if role == "messe" {
			want = http.StatusForbidden
		}
		if response.Code != want {
			t.Fatalf("%s read: got %d, want %d: %s", role, response.Code, want, response.Body.String())
		}
	}
	layout, err := layouts.CreateLayout(t.Context(), application.CreateLayoutInput{
		Name: "Twin", Kind: domain.LayoutKindPrivate, Gauge: "TT", Scale: "1:120",
	}, "")
	if err != nil {
		t.Fatal(err)
	}
	unit, err := layouts.CreateUnit(t.Context(), layout.ID, application.CreateLayoutUnitInput{
		Name: "Module", Kind: domain.LayoutUnitKindModule, WidthMM: 1000, HeightMM: 500,
	}, "")
	if err != nil {
		t.Fatal(err)
	}
	for role, session := range sessions {
		response := layoutRequest(t, router, session, http.MethodGet,
			"/api/v1/layouts/"+layout.ID+"/twin", nil, true)
		want := http.StatusOK
		if role == "messe" {
			want = http.StatusForbidden
		}
		if response.Code != want {
			t.Fatalf("%s twin read: got %d, want %d: %s", role, response.Code, want, response.Body.String())
		}
	}
	for role, session := range sessions {
		response := layoutRequest(t, router, session, http.MethodGet,
			"/api/v1/layout-units/"+unit.ID+"/ports", nil, true)
		want := http.StatusOK
		if role == "messe" {
			want = http.StatusForbidden
		}
		if response.Code != want {
			t.Fatalf("%s port read: got %d, want %d: %s", role, response.Code, want, response.Body.String())
		}
	}
	for role, session := range sessions {
		response := layoutRequest(t, router, session, http.MethodPost,
			"/api/v1/layout-units/"+unit.ID+"/ports", map[string]any{
				"name": "Port " + role, "kind": "track", "interfaceKey": "track:tillig-tt-modellgleis",
				"xMm": 0, "yMm": 250, "directionDegrees": 180,
			}, true)
		want := http.StatusForbidden
		if role == "admin" || role == "planner" {
			want = http.StatusCreated
		}
		if response.Code != want {
			t.Fatalf("%s port write: got %d, want %d: %s", role, response.Code, want, response.Body.String())
		}
	}
	withoutPortCSRF := layoutRequest(t, router, sessions["planner"], http.MethodPost,
		"/api/v1/layout-units/"+unit.ID+"/ports", map[string]any{
			"name": "No CSRF port", "kind": "track", "interfaceKey": "track:tillig-tt-modellgleis",
		}, false)
	assertProblem(t, withoutPortCSRF, http.StatusForbidden, "csrf_required")
	configuration, err := layouts.SaveConfiguration(t.Context(), layout.ID,
		application.SaveLayoutConfigurationInput{Name: "Port analysis", Units: []application.ConfigurationUnitInput{
			{UnitID: unit.ID},
		}}, "")
	if err != nil {
		t.Fatal(err)
	}
	for role, session := range sessions {
		response := layoutRequest(t, router, session, http.MethodGet,
			"/api/v1/layout-configurations/"+configuration.ID+"/port-analysis", nil, true)
		want := http.StatusOK
		if role == "messe" {
			want = http.StatusForbidden
		}
		if response.Code != want {
			t.Fatalf("%s configuration port analysis: got %d, want %d: %s",
				role, response.Code, want, response.Body.String())
		}
	}
	for role, session := range sessions {
		response := layoutRequest(t, router, session, http.MethodPost,
			"/api/v1/layout-configurations/"+configuration.ID+"/unit-snap-preview", map[string]any{
				"unitId": unit.ID, "positionXMm": 0, "positionYMm": 0, "rotationDegrees": 0,
				"units": []map[string]any{{
					"unitId": unit.ID, "positionXMm": 0, "positionYMm": 0, "rotationDegrees": 0,
				}},
			}, true)
		want := http.StatusForbidden
		if role == "admin" || role == "planner" {
			want = http.StatusOK
		}
		if response.Code != want {
			t.Fatalf("%s configuration port preview: got %d, want %d: %s",
				role, response.Code, want, response.Body.String())
		}
	}
	withoutPreviewCSRF := layoutRequest(t, router, sessions["planner"], http.MethodPost,
		"/api/v1/layout-configurations/"+configuration.ID+"/unit-snap-preview", map[string]any{
			"unitId": unit.ID,
		}, false)
	assertProblem(t, withoutPreviewCSRF, http.StatusForbidden, "csrf_required")
	for role, session := range sessions {
		response := layoutRequest(t, router, session, http.MethodPost, "/api/v1/layouts", map[string]any{
			"name": "Layout " + role, "kind": "private", "gauge": "TT", "scale": "1:120",
		}, true)
		want := http.StatusForbidden
		if role == "admin" || role == "planner" {
			want = http.StatusCreated
		}
		if response.Code != want {
			t.Fatalf("%s write: got %d, want %d: %s", role, response.Code, want, response.Body.String())
		}
	}
	withoutCSRF := layoutRequest(t, router, sessions["planner"], http.MethodPost, "/api/v1/layouts", map[string]any{
		"name": "No CSRF", "kind": "private", "gauge": "TT", "scale": "1:120",
	}, false)
	assertProblem(t, withoutCSRF, http.StatusForbidden, "csrf_required")
}

func TestLayoutRoutesCoverStructureAndRevisionWorkflow(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{
		Username: "planner", Password: "planner-password", Roles: []string{"Planner"},
	}); err != nil {
		t.Fatal(err)
	}
	session := loginRouteTestUser(t, auth, "planner", "planner-password")
	router := NewRouter(Config{
		AuthService:   auth,
		LayoutService: application.NewLayoutService(infrastructure.NewLayoutRepository(db)),
	})

	layoutResponse := layoutRequest(t, router, session, http.MethodPost, "/api/v1/layouts", map[string]any{
		"name": "Heimanlage", "kind": "private", "gauge": "TT", "scale": "1:120",
		"maxGradePercent": 3.5, "minimumTrackClearanceMm": 40,
	}, true)
	if layoutResponse.Code != http.StatusCreated {
		t.Fatalf("create layout: %d: %s", layoutResponse.Code, layoutResponse.Body.String())
	}
	var layout application.Layout
	decodeResponse(t, layoutResponse, &layout)
	if layout.MaxGradePercent == nil || *layout.MaxGradePercent != 3.5 {
		t.Fatalf("unexpected layout grade limit: %#v", layout)
	}
	if layout.MinimumTrackClearanceMM == nil || *layout.MinimumTrackClearanceMM != 40 {
		t.Fatalf("unexpected layout clearance limit: %#v", layout)
	}

	assertStatus(t, layoutRequest(t, router, session, http.MethodGet, "/api/v1/layouts", nil, true), http.StatusOK)
	assertStatus(t, layoutRequest(t, router, session, http.MethodGet, "/api/v1/layouts/"+layout.ID, nil, true), http.StatusOK)
	updatedLayoutResponse := layoutRequest(t, router, session, http.MethodPut, "/api/v1/layouts/"+layout.ID, map[string]any{
		"name": "Heimanlage erweitert", "kind": "private", "gauge": "TT", "scale": "1:120",
		"maxGradePercent": 2.5, "minimumTrackClearanceMm": 25.5, "expectedVersion": layout.Version,
	}, true)
	assertStatus(t, updatedLayoutResponse, http.StatusOK)
	var updatedLayout application.Layout
	decodeResponse(t, updatedLayoutResponse, &updatedLayout)
	if updatedLayout.MaxGradePercent == nil || *updatedLayout.MaxGradePercent != 2.5 {
		t.Fatalf("unexpected updated layout grade limit: %#v", updatedLayout)
	}
	if updatedLayout.MinimumTrackClearanceMM == nil || *updatedLayout.MinimumTrackClearanceMM != 25.5 {
		t.Fatalf("unexpected updated layout clearance limit: %#v", updatedLayout)
	}
	staleLayout := layoutRequest(t, router, session, http.MethodPut, "/api/v1/layouts/"+layout.ID, map[string]any{
		"name": "Stale", "kind": "private", "gauge": "TT", "scale": "1:120", "expectedVersion": 1,
	}, true)
	assertProblem(t, staleLayout, http.StatusConflict, "layout_version_conflict")

	unitResponse := layoutRequest(t, router, session, http.MethodPost, "/api/v1/layouts/"+layout.ID+"/units", map[string]any{
		"name": "Grundplatte", "kind": "baseboard", "widthMm": 2000, "heightMm": 1000,
	}, true)
	assertStatus(t, unitResponse, http.StatusCreated)
	var unit application.LayoutUnit
	decodeResponse(t, unitResponse, &unit)
	updatedUnitResponse := layoutRequest(t, router, session, http.MethodPut, "/api/v1/layout-units/"+unit.ID, map[string]any{
		"name": "Grundplatte erweitert", "kind": "baseboard", "widthMm": 2200, "heightMm": 1000,
		"expectedVersion": unit.Version,
	}, true)
	assertStatus(t, updatedUnitResponse, http.StatusOK)
	assertStatus(t, layoutRequest(t, router, session, http.MethodGet, "/api/v1/layouts/"+layout.ID+"/units", nil, true), http.StatusOK)

	portResponse := layoutRequest(t, router, session, http.MethodPost,
		"/api/v1/layout-units/"+unit.ID+"/ports", map[string]any{
			"name": "West", "kind": "track", "interfaceKey": "TRACK:Tillig-TT-Modellgleis",
			"xMm": 0, "yMm": 500, "directionDegrees": -180,
		}, true)
	assertStatus(t, portResponse, http.StatusCreated)
	var port application.LayoutUnitPort
	decodeResponse(t, portResponse, &port)
	if port.InterfaceKey != "track:tillig-tt-modellgleis" || port.DirectionDegrees != 180 {
		t.Fatalf("unexpected normalized module port: %#v", port)
	}
	assertStatus(t, layoutRequest(t, router, session, http.MethodGet,
		"/api/v1/layout-units/"+unit.ID+"/ports", nil, true), http.StatusOK)
	portUpdate := layoutRequest(t, router, session, http.MethodPut,
		"/api/v1/layout-unit-ports/"+port.ID, map[string]any{
			"name": "West main", "kind": "track", "interfaceKey": "track:tillig-tt-modellgleis",
			"xMm": 0, "yMm": 500, "directionDegrees": 180, "expectedVersion": port.Version,
		}, true)
	assertStatus(t, portUpdate, http.StatusOK)
	stalePort := layoutRequest(t, router, session, http.MethodPut,
		"/api/v1/layout-unit-ports/"+port.ID, map[string]any{
			"name": "Stale", "kind": "track", "interfaceKey": "track:tillig-tt-modellgleis",
			"expectedVersion": port.Version,
		}, true)
	assertProblem(t, stalePort, http.StatusConflict, "layout_version_conflict")

	configurationResponse := layoutRequest(
		t, router, session, http.MethodPost, "/api/v1/layouts/"+layout.ID+"/configurations", map[string]any{
			"name":  "Standardaufbau",
			"units": []map[string]any{{"unitId": unit.ID, "positionXMm": 10, "positionYMm": 20, "rotationDegrees": -15}},
		}, true,
	)
	assertStatus(t, configurationResponse, http.StatusCreated)
	var configuration application.LayoutConfiguration
	decodeResponse(t, configurationResponse, &configuration)
	assertStatus(t, layoutRequest(t, router, session, http.MethodGet,
		"/api/v1/layouts/"+layout.ID+"/configurations", nil, true), http.StatusOK)
	assertStatus(t, layoutRequest(t, router, session, http.MethodGet,
		"/api/v1/layouts/"+layout.ID+"/twin?configurationId="+configuration.ID, nil, true), http.StatusOK)
	assertProblem(t, layoutRequest(t, router, session, http.MethodGet,
		"/api/v1/layouts/"+layout.ID+"/twin?configurationId="+configuration.ID+"&unitId="+unit.ID,
		nil, true), http.StatusBadRequest, "layout_validation")
	configurationUpdate := layoutRequest(
		t, router, session, http.MethodPut, "/api/v1/layout-configurations/"+configuration.ID, map[string]any{
			"name": "Standardaufbau erweitert", "expectedVersion": configuration.Version,
			"units": configuration.Units,
		}, true,
	)
	assertStatus(t, configurationUpdate, http.StatusOK)

	variantResponse := layoutRequest(
		t, router, session, http.MethodPost, "/api/v1/layout-units/"+unit.ID+"/plan-variants",
		map[string]any{"name": "Standard"}, true,
	)
	assertStatus(t, variantResponse, http.StatusCreated)
	var variant application.PlanVariant
	decodeResponse(t, variantResponse, &variant)
	draftResponse := layoutRequest(
		t, router, session, http.MethodPost, "/api/v1/plan-variants/"+variant.ID+"/revisions",
		map[string]any{}, true,
	)
	assertStatus(t, draftResponse, http.StatusCreated)
	var revision application.PlanRevision
	decodeResponse(t, draftResponse, &revision)
	submitResponse := layoutRequest(
		t, router, session, http.MethodPost, "/api/v1/plan-revisions/"+revision.ID+"/submit",
		map[string]any{"expectedVersion": revision.Version}, true,
	)
	assertStatus(t, submitResponse, http.StatusOK)
	decodeResponse(t, submitResponse, &revision)
	publishResponse := layoutRequest(
		t, router, session, http.MethodPost, "/api/v1/plan-revisions/"+revision.ID+"/publish",
		map[string]any{"expectedVersion": revision.Version}, true,
	)
	assertStatus(t, publishResponse, http.StatusOK)
	decodeResponse(t, publishResponse, &revision)
	if revision.Status != domain.PlanRevisionPublished {
		t.Fatalf("expected published revision, got %#v", revision)
	}
	assertStatus(t, layoutRequest(t, router, session, http.MethodGet,
		"/api/v1/layout-units/"+unit.ID+"/plan-variants", nil, true), http.StatusOK)
	immutable := layoutRequest(
		t, router, session, http.MethodPost, "/api/v1/plan-revisions/"+revision.ID+"/submit",
		map[string]any{"expectedVersion": revision.Version}, true,
	)
	assertProblem(t, immutable, http.StatusConflict, "plan_revision_immutable")
	assertProblem(t, layoutRequest(t, router, session, http.MethodGet,
		"/api/v1/layouts/missing", nil, true), http.StatusNotFound, "layout_not_found")
	assertProblem(t, layoutRequest(t, router, session, http.MethodPost, "/api/v1/layouts",
		map[string]any{"name": "Invalid"}, true), http.StatusBadRequest, "layout_validation")
}

func layoutRequest(
	t *testing.T,
	router http.Handler,
	session *application.LoginResult,
	method, path string,
	body any,
	withCSRF bool,
) *httptest.ResponseRecorder {
	t.Helper()
	var encodedBody bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&encodedBody).Encode(body); err != nil {
			t.Fatal(err)
		}
	}
	request := httptest.NewRequest(method, path, &encodedBody)
	request.Header.Set("Content-Type", "application/json")
	request.AddCookie(&http.Cookie{Name: "rk_session", Value: session.SessionToken})
	if withCSRF {
		request.Header.Set("X-CSRF-Token", session.CSRFToken)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func decodeResponse(t *testing.T, response *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		t.Fatal(err)
	}
}

func assertStatus(t *testing.T, response *httptest.ResponseRecorder, want int) {
	t.Helper()
	if response.Code != want {
		t.Fatalf("got status %d, want %d: %s", response.Code, want, response.Body.String())
	}
}

func assertProblem(t *testing.T, response *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	if response.Code != status {
		t.Fatalf("got status %d, want %d: %s", response.Code, status, response.Body.String())
	}
	var problem map[string]string
	decodeResponse(t, response, &problem)
	if problem["error"] != code {
		t.Fatalf("got problem code %q, want %q: %#v", problem["error"], code, problem)
	}
}
