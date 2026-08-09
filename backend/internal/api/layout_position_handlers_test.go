package api

import (
	"net/http"
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/infrastructure"
)

func TestLayoutTechnicalPositionRoutesEnforceRolesAndCSRF(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	for _, user := range []application.CreateUserInput{
		{Username: "admin-position", Password: "admin-password", Roles: []string{"Admin"}},
		{Username: "planner-position", Password: "planner-password", Roles: []string{"Planner"}},
		{Username: "editor-position", Password: "editor-password", Roles: []string{"Editor"}},
		{Username: "viewer-position", Password: "viewer-password", Roles: []string{"Viewer"}},
		{Username: "messe-position", Password: "messe-password", Roles: []string{"Messe"}},
	} {
		if _, err := auth.CreateUser(t.Context(), "", user); err != nil {
			t.Fatal(err)
		}
	}
	layouts := application.NewLayoutService(infrastructure.NewLayoutRepository(db))
	router := NewRouter(Config{AuthService: auth, LayoutService: layouts})
	sessions := map[string]*application.LoginResult{
		"admin":   loginRouteTestUser(t, auth, "admin-position", "admin-password"),
		"planner": loginRouteTestUser(t, auth, "planner-position", "planner-password"),
		"editor":  loginRouteTestUser(t, auth, "editor-position", "editor-password"),
		"viewer":  loginRouteTestUser(t, auth, "viewer-position", "viewer-password"),
		"messe":   loginRouteTestUser(t, auth, "messe-position", "messe-password"),
	}
	layout, err := layouts.CreateLayout(t.Context(), application.CreateLayoutInput{
		Name: "Positionsanlage", Kind: "private", Gauge: "TT", Scale: "1:120",
	}, "")
	if err != nil {
		t.Fatal(err)
	}
	unit, err := layouts.CreateUnit(t.Context(), layout.ID, application.CreateLayoutUnitInput{
		Name: "Segment A", Kind: "module", WidthMM: 1000, HeightMM: 500,
	}, "")
	if err != nil {
		t.Fatal(err)
	}

	for role, session := range sessions {
		response := layoutRequest(t, router, session, http.MethodGet,
			"/api/v1/layout-units/"+unit.ID+"/technical-positions", nil, true)
		want := http.StatusOK
		if role == "messe" {
			want = http.StatusForbidden
		}
		assertStatus(t, response, want)
	}
	for role, session := range sessions {
		response := layoutRequest(t, router, session, http.MethodPost,
			"/api/v1/layout-units/"+unit.ID+"/technical-positions", map[string]any{
				"label": "Signal " + role, "kind": "signal", "positionXMm": 100, "positionYMm": 50,
			}, true)
		want := http.StatusForbidden
		if role == "admin" || role == "planner" {
			want = http.StatusCreated
		}
		assertStatus(t, response, want)
	}
	for role, session := range sessions {
		outlineUnit, err := layouts.CreateUnit(t.Context(), layout.ID, application.CreateLayoutUnitInput{
			Name: "Outline " + role, Kind: "module", WidthMM: 100, HeightMM: 50,
		}, "")
		if err != nil {
			t.Fatal(err)
		}
		response := layoutRequest(t, router, session, http.MethodPut,
			"/api/v1/layout-units/"+outlineUnit.ID+"/outline", map[string]any{
				"expectedVersion": outlineUnit.Version,
				"points": []map[string]any{{"xMm": 0, "yMm": 0}, {"xMm": 100, "yMm": 0},
					{"xMm": 100, "yMm": 50}, {"xMm": 0, "yMm": 50}},
			}, true)
		want := http.StatusForbidden
		if role == "admin" || role == "planner" {
			want = http.StatusOK
		}
		assertStatus(t, response, want)
	}
	withoutCSRF := layoutRequest(t, router, sessions["planner"], http.MethodPost,
		"/api/v1/layout-units/"+unit.ID+"/technical-positions", map[string]any{
			"label": "Ohne CSRF", "kind": "signal", "positionXMm": 10, "positionYMm": 10,
		}, false)
	assertProblem(t, withoutCSRF, http.StatusForbidden, "csrf_required")
	outlineWithoutCSRF := layoutRequest(t, router, sessions["planner"], http.MethodPut,
		"/api/v1/layout-units/"+unit.ID+"/outline", map[string]any{
			"expectedVersion": unit.Version,
			"points": []map[string]any{{"xMm": 0, "yMm": 0}, {"xMm": 100, "yMm": 0},
				{"xMm": 0, "yMm": 50}},
		}, false)
	assertProblem(t, outlineWithoutCSRF, http.StatusForbidden, "csrf_required")
}

func TestLayoutTechnicalPositionRoutesCoverWorkflowAndProblems(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{
		Username: "position-planner", Password: "planner-password", Roles: []string{"Planner"},
	}); err != nil {
		t.Fatal(err)
	}
	session := loginRouteTestUser(t, auth, "position-planner", "planner-password")
	layouts := application.NewLayoutService(infrastructure.NewLayoutRepository(db))
	router := NewRouter(Config{AuthService: auth, LayoutService: layouts})
	layout, err := layouts.CreateLayout(t.Context(), application.CreateLayoutInput{
		Name: "Testanlage", Kind: "private", Gauge: "TT", Scale: "1:120",
	}, "")
	if err != nil {
		t.Fatal(err)
	}
	unit, err := layouts.CreateUnit(t.Context(), layout.ID, application.CreateLayoutUnitInput{
		Name: "Bahnhof", Kind: "module", WidthMM: 1200, HeightMM: 400,
	}, "")
	if err != nil {
		t.Fatal(err)
	}

	createdResponse := layoutRequest(t, router, session, http.MethodPost,
		"/api/v1/layout-units/"+unit.ID+"/technical-positions", map[string]any{
			"label": "Einfahrsignal A", "kind": "signal", "positionXMm": 120.5,
			"positionYMm": 30, "rotationDegrees": -90, "description": "Gleis 1",
		}, true)
	assertStatus(t, createdResponse, http.StatusCreated)
	var created application.LayoutTechnicalPosition
	decodeResponse(t, createdResponse, &created)
	if created.RotationDegrees != 270 || created.Version != 1 {
		t.Fatalf("unexpected created position: %#v", created)
	}

	listResponse := layoutRequest(t, router, session, http.MethodGet,
		"/api/v1/layout-units/"+unit.ID+"/technical-positions", nil, true)
	assertStatus(t, listResponse, http.StatusOK)
	var positions []application.LayoutTechnicalPosition
	decodeResponse(t, listResponse, &positions)
	if len(positions) != 1 || positions[0].ID != created.ID {
		t.Fatalf("unexpected positions: %#v", positions)
	}

	updatedResponse := layoutRequest(t, router, session, http.MethodPut,
		"/api/v1/layout-technical-positions/"+created.ID, map[string]any{
			"label": "Einfahrsignal A neu", "kind": "signal", "positionXMm": 125,
			"positionYMm": 35, "rotationDegrees": 90, "expectedVersion": created.Version,
		}, true)
	assertStatus(t, updatedResponse, http.StatusOK)
	var updated application.LayoutTechnicalPosition
	decodeResponse(t, updatedResponse, &updated)
	if updated.Version != 2 || updated.Label != "Einfahrsignal A neu" {
		t.Fatalf("unexpected updated position: %#v", updated)
	}

	staleResponse := layoutRequest(t, router, session, http.MethodPut,
		"/api/v1/layout-technical-positions/"+created.ID, map[string]any{
			"label": "Veraltet", "kind": "signal", "positionXMm": 125,
			"positionYMm": 35, "rotationDegrees": 90, "expectedVersion": created.Version,
		}, true)
	assertProblem(t, staleResponse, http.StatusConflict, "layout_position_version_conflict")
	assertProblem(t, layoutRequest(t, router, session, http.MethodGet,
		"/api/v1/layout-units/missing/technical-positions", nil, true),
		http.StatusNotFound, "layout_not_found")
	assertProblem(t, layoutRequest(t, router, session, http.MethodPost,
		"/api/v1/layout-units/"+unit.ID+"/technical-positions", map[string]any{
			"label": "Invalid", "kind": "unknown",
		}, true), http.StatusBadRequest, "layout_validation")
	assertProblem(t, layoutRequest(t, router, session, http.MethodPut,
		"/api/v1/layout-technical-positions/missing", map[string]any{
			"label": "Missing", "kind": "signal", "expectedVersion": 1,
		}, true), http.StatusNotFound, "layout_position_not_found")
}
