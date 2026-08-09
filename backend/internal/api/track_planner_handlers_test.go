package api

import (
	"fmt"
	"net/http"
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/domain"
	"railkeeper/backend/internal/infrastructure"
)

func TestTrackPlannerRoutesEnforceRolesAndCSRF(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	for _, user := range []application.CreateUserInput{
		{Username: "admin-track", Password: "admin-password", Roles: []string{"Admin"}},
		{Username: "planner-track", Password: "planner-password", Roles: []string{"Planner"}},
		{Username: "editor-track", Password: "editor-password", Roles: []string{"Editor"}},
		{Username: "viewer-track", Password: "viewer-password", Roles: []string{"Viewer"}},
		{Username: "messe-track", Password: "messe-password", Roles: []string{"Messe"}},
	} {
		if _, err := auth.CreateUser(t.Context(), "", user); err != nil {
			t.Fatal(err)
		}
	}
	layouts := application.NewLayoutService(infrastructure.NewLayoutRepository(db))
	planner := application.NewTrackPlannerService(infrastructure.NewTrackPlannerRepository(db))
	router := NewRouter(Config{AuthService: auth, LayoutService: layouts, TrackPlannerService: planner})
	sessions := map[string]*application.LoginResult{
		"admin":   loginRouteTestUser(t, auth, "admin-track", "admin-password"),
		"planner": loginRouteTestUser(t, auth, "planner-track", "planner-password"),
		"editor":  loginRouteTestUser(t, auth, "editor-track", "editor-password"),
		"viewer":  loginRouteTestUser(t, auth, "viewer-track", "viewer-password"),
		"messe":   loginRouteTestUser(t, auth, "messe-track", "messe-password"),
	}
	_, draft := trackPlannerRouteRevision(t, layouts)

	for role, session := range sessions {
		for _, path := range []string{
			"/api/v1/track-geometries?gauge=TT",
			"/api/v1/plan-revisions/" + draft.ID + "/track-plan",
			"/api/v1/plan-revisions/" + draft.ID + "/track-analysis",
			"/api/v1/plan-revisions/" + draft.ID + "/track-change-preview",
		} {
			response := layoutRequest(t, router, session, http.MethodGet, path, nil, true)
			want := http.StatusOK
			if role == "messe" {
				want = http.StatusForbidden
			}
			assertStatus(t, response, want)
		}
	}

	for role, session := range sessions {
		response := layoutRequest(t, router, session, http.MethodPost,
			"/api/v1/plan-revisions/"+draft.ID+"/track-objects", map[string]any{
				"geometryId":  "tillig-tt-modellgleis-83101-v1",
				"positionXMm": 100, "positionYMm": 50, "rotationDegrees": 0,
			}, true)
		want := http.StatusForbidden
		if role == "admin" || role == "planner" {
			want = http.StatusCreated
		}
		assertStatus(t, response, want)
	}

	withoutCSRF := layoutRequest(t, router, sessions["planner"], http.MethodPost,
		"/api/v1/plan-revisions/"+draft.ID+"/track-objects", map[string]any{
			"geometryId": "tillig-tt-modellgleis-83101-v1",
		}, false)
	assertProblem(t, withoutCSRF, http.StatusForbidden, "csrf_required")
}

func TestTrackPlannerRoutesCoverWorkflowAndProblems(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{
		Username: "planner-track-workflow", Password: "planner-password", Roles: []string{"Planner"},
	}); err != nil {
		t.Fatal(err)
	}
	session := loginRouteTestUser(t, auth, "planner-track-workflow", "planner-password")
	layouts := application.NewLayoutService(infrastructure.NewLayoutRepository(db))
	planner := application.NewTrackPlannerService(infrastructure.NewTrackPlannerRepository(db))
	router := NewRouter(Config{AuthService: auth, LayoutService: layouts, TrackPlannerService: planner})
	_, draft := trackPlannerRouteRevision(t, layouts)

	geometryResponse := layoutRequest(t, router, session, http.MethodGet,
		"/api/v1/track-geometries?gauge=TT", nil, true)
	assertStatus(t, geometryResponse, http.StatusOK)
	var geometries []domain.TrackGeometryDefinition
	decodeResponse(t, geometryResponse, &geometries)
	if len(geometries) != 1 || geometries[0].ArticleNumber != "83101" {
		t.Fatalf("unexpected track geometries: %#v", geometries)
	}

	createdResponse := layoutRequest(t, router, session, http.MethodPost,
		"/api/v1/plan-revisions/"+draft.ID+"/track-objects", map[string]any{
			"geometryId": geometries[0].ID, "positionXMm": 100, "positionYMm": 50,
			"rotationDegrees": -15,
		}, true)
	assertStatus(t, createdResponse, http.StatusCreated)
	var created domain.PlanTrackObject
	decodeResponse(t, createdResponse, &created)
	if created.Version != 1 || created.RotationDegrees != 345 {
		t.Fatalf("unexpected created track object: %#v", created)
	}

	planResponse := layoutRequest(t, router, session, http.MethodGet,
		"/api/v1/plan-revisions/"+draft.ID+"/track-plan", nil, true)
	assertStatus(t, planResponse, http.StatusOK)
	var plan application.TrackPlan
	decodeResponse(t, planResponse, &plan)
	if len(plan.Objects) != 1 || plan.Objects[0].ID != created.ID {
		t.Fatalf("unexpected track plan: %#v", plan)
	}

	updatedResponse := layoutRequest(t, router, session, http.MethodPut,
		"/api/v1/plan-track-objects/"+created.ID, map[string]any{
			"positionXMm": 110, "positionYMm": 55, "rotationDegrees": 0,
			"expectedVersion": created.Version,
		}, true)
	assertStatus(t, updatedResponse, http.StatusOK)
	var updated domain.PlanTrackObject
	decodeResponse(t, updatedResponse, &updated)
	if updated.Version != 2 || updated.PositionXMM != 110 {
		t.Fatalf("unexpected updated track object: %#v", updated)
	}
	secondResponse := layoutRequest(t, router, session, http.MethodPost,
		"/api/v1/plan-revisions/"+draft.ID+"/track-objects", map[string]any{
			"geometryId": geometries[0].ID, "positionXMm": 276, "positionYMm": 55,
			"rotationDegrees": 0,
		}, true)
	assertStatus(t, secondResponse, http.StatusCreated)

	analysisResponse := layoutRequest(t, router, session, http.MethodGet,
		"/api/v1/plan-revisions/"+draft.ID+"/track-analysis", nil, true)
	assertStatus(t, analysisResponse, http.StatusOK)
	var analysis application.TrackPlanAnalysis
	decodeResponse(t, analysisResponse, &analysis)
	if analysis.RevisionID != draft.ID || len(analysis.Connections) != 1 ||
		len(analysis.BOM) != 1 || analysis.BOM[0].Quantity != 2 {
		t.Fatalf("unexpected track plan analysis: %#v", analysis)
	}
	previewResponse := layoutRequest(t, router, session, http.MethodGet,
		"/api/v1/plan-revisions/"+draft.ID+"/track-change-preview", nil, true)
	assertStatus(t, previewResponse, http.StatusOK)
	var preview application.TrackPlanChangePreview
	decodeResponse(t, previewResponse, &preview)
	if preview.RevisionID != draft.ID || preview.BaseRevisionID != "" ||
		len(preview.ObjectChanges) != 2 || len(preview.MaterialDeltas) != 1 ||
		preview.MaterialDeltas[0].Delta != 2 {
		t.Fatalf("unexpected track plan change preview: %#v", preview)
	}

	staleResponse := layoutRequest(t, router, session, http.MethodPut,
		"/api/v1/plan-track-objects/"+created.ID, map[string]any{
			"positionXMm": 120, "positionYMm": 60, "rotationDegrees": 30,
			"expectedVersion": created.Version,
		}, true)
	assertProblem(t, staleResponse, http.StatusConflict, "track_plan_conflict")

	deleteResponse := layoutRequest(t, router, session, http.MethodDelete,
		fmt.Sprintf("/api/v1/plan-track-objects/%s?expectedVersion=%d", created.ID, updated.Version), nil, true)
	assertStatus(t, deleteResponse, http.StatusNoContent)

	assertProblem(t, layoutRequest(t, router, session, http.MethodGet,
		"/api/v1/plan-revisions/missing/track-plan", nil, true),
		http.StatusNotFound, "track_plan_not_found")
	assertProblem(t, layoutRequest(t, router, session, http.MethodGet,
		"/api/v1/plan-revisions/missing/track-analysis", nil, true),
		http.StatusNotFound, "track_plan_not_found")
	assertProblem(t, layoutRequest(t, router, session, http.MethodGet,
		"/api/v1/plan-revisions/missing/track-change-preview", nil, true),
		http.StatusNotFound, "track_plan_not_found")
	assertProblem(t, layoutRequest(t, router, session, http.MethodPost,
		"/api/v1/plan-revisions/"+draft.ID+"/track-objects", map[string]any{
			"geometryId": "missing",
		}, true), http.StatusBadRequest, "track_plan_validation")
	assertProblem(t, layoutRequest(t, router, session, http.MethodDelete,
		"/api/v1/plan-track-objects/missing?expectedVersion=1", nil, true),
		http.StatusNotFound, "track_plan_not_found")
}

func trackPlannerRouteRevision(
	t *testing.T,
	layouts *application.LayoutService,
) (*application.PlanVariant, *application.PlanRevision) {
	t.Helper()
	layout, err := layouts.CreateLayout(t.Context(), application.CreateLayoutInput{
		Name: "Track API", Kind: domain.LayoutKindPrivate, Gauge: "TT", Scale: "1:120",
	}, "")
	if err != nil {
		t.Fatal(err)
	}
	unit, err := layouts.CreateUnit(t.Context(), layout.ID, application.CreateLayoutUnitInput{
		Name: "Segment", Kind: domain.LayoutUnitKindSegment, WidthMM: 1000, HeightMM: 400,
	}, "")
	if err != nil {
		t.Fatal(err)
	}
	variant, err := layouts.CreateVariant(t.Context(), unit.ID,
		application.CreatePlanVariantInput{Name: "Standard"}, "")
	if err != nil {
		t.Fatal(err)
	}
	draft, err := layouts.CreateDraft(t.Context(), variant.ID, application.CreatePlanRevisionInput{}, "")
	if err != nil {
		t.Fatal(err)
	}
	return variant, draft
}
