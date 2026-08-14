package api

import (
	"net/http"
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/domain"
	"railkeeper/backend/internal/infrastructure"
)

func TestFreePlanObjectRoutesEnforceRolesCSRFAndWorkflow(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	users := []application.CreateUserInput{
		{Username: "admin-free", Password: "admin-password", Roles: []string{"Admin"}},
		{Username: "planner-free", Password: "planner-password", Roles: []string{"Planner"}},
		{Username: "editor-free", Password: "editor-password", Roles: []string{"Editor"}},
		{Username: "viewer-free", Password: "viewer-password", Roles: []string{"Viewer"}},
		{Username: "messe-free", Password: "messe-password", Roles: []string{"Messe"}},
	}
	for _, user := range users {
		if _, err := auth.CreateUser(t.Context(), "", user); err != nil {
			t.Fatal(err)
		}
	}
	sessions := map[string]*application.LoginResult{
		"admin":   loginRouteTestUser(t, auth, "admin-free", "admin-password"),
		"planner": loginRouteTestUser(t, auth, "planner-free", "planner-password"),
		"editor":  loginRouteTestUser(t, auth, "editor-free", "editor-password"),
		"viewer":  loginRouteTestUser(t, auth, "viewer-free", "viewer-password"),
		"messe":   loginRouteTestUser(t, auth, "messe-free", "messe-password"),
	}
	layouts := application.NewLayoutService(infrastructure.NewLayoutRepository(db))
	planner := application.NewTrackPlannerService(infrastructure.NewTrackPlannerRepository(db))
	router := NewRouter(Config{AuthService: auth, LayoutService: layouts, TrackPlannerService: planner})
	_, draft := trackPlannerRouteRevision(t, layouts)
	createInput := map[string]any{
		"name": "Bahnsteig", "category": "platform", "positionXMm": 100,
		"positionYMm": 50, "rotationDegrees": 0,
		"shape": map[string]any{
			"schemaVersion": 1, "kind": "rectangle", "widthMm": 500, "heightMm": 70,
		},
	}
	for role, session := range sessions {
		response := layoutRequest(t, router, session, http.MethodPost,
			"/api/v1/plan-revisions/"+draft.ID+"/free-objects", createInput, true)
		want := http.StatusForbidden
		if role == "admin" || role == "planner" {
			want = http.StatusCreated
		}
		assertStatus(t, response, want)
	}
	withoutCSRF := layoutRequest(t, router, sessions["planner"], http.MethodPost,
		"/api/v1/plan-revisions/"+draft.ID+"/free-objects", createInput, false)
	assertProblem(t, withoutCSRF, http.StatusForbidden, "csrf_required")

	createdResponse := layoutRequest(t, router, sessions["planner"], http.MethodPost,
		"/api/v1/plan-revisions/"+draft.ID+"/free-objects", createInput, true)
	assertStatus(t, createdResponse, http.StatusCreated)
	var created domain.PlanFreeObject
	decodeResponse(t, createdResponse, &created)
	updateInput := createInput
	updateInput["name"] = "Bahnsteig 1"
	updateInput["expectedVersion"] = created.Version
	forbiddenUpdate := layoutRequest(t, router, sessions["editor"], http.MethodPut,
		"/api/v1/plan-free-objects/"+created.ID, updateInput, true)
	assertStatus(t, forbiddenUpdate, http.StatusForbidden)
	updateWithoutCSRF := layoutRequest(t, router, sessions["planner"], http.MethodPut,
		"/api/v1/plan-free-objects/"+created.ID, updateInput, false)
	assertProblem(t, updateWithoutCSRF, http.StatusForbidden, "csrf_required")
	updatedResponse := layoutRequest(t, router, sessions["planner"], http.MethodPut,
		"/api/v1/plan-free-objects/"+created.ID, updateInput, true)
	assertStatus(t, updatedResponse, http.StatusOK)
	var updated domain.PlanFreeObject
	decodeResponse(t, updatedResponse, &updated)
	if updated.Name != "Bahnsteig 1" || updated.Version != created.Version+1 {
		t.Fatalf("unexpected updated free object: %#v", updated)
	}
	stale := layoutRequest(t, router, sessions["planner"], http.MethodPut,
		"/api/v1/plan-free-objects/"+created.ID, updateInput, true)
	assertProblem(t, stale, http.StatusConflict, "track_plan_conflict")
	forbiddenDelete := layoutRequest(t, router, sessions["viewer"], http.MethodDelete,
		"/api/v1/plan-free-objects/"+created.ID+"?expectedVersion=2", nil, true)
	assertStatus(t, forbiddenDelete, http.StatusForbidden)
	deleteWithoutCSRF := layoutRequest(t, router, sessions["planner"], http.MethodDelete,
		"/api/v1/plan-free-objects/"+created.ID+"?expectedVersion=2", nil, false)
	assertProblem(t, deleteWithoutCSRF, http.StatusForbidden, "csrf_required")
	deleted := layoutRequest(t, router, sessions["planner"], http.MethodDelete,
		"/api/v1/plan-free-objects/"+created.ID+"?expectedVersion=2", nil, true)
	assertStatus(t, deleted, http.StatusNoContent)
	notFound := layoutRequest(t, router, sessions["planner"], http.MethodDelete,
		"/api/v1/plan-free-objects/missing?expectedVersion=1", nil, true)
	assertProblem(t, notFound, http.StatusNotFound, "track_plan_not_found")

	immutableResponse := layoutRequest(t, router, sessions["planner"], http.MethodPost,
		"/api/v1/plan-revisions/"+draft.ID+"/free-objects", createInput, true)
	assertStatus(t, immutableResponse, http.StatusCreated)
	var immutable domain.PlanFreeObject
	decodeResponse(t, immutableResponse, &immutable)
	submitted, err := layouts.SubmitRevision(t.Context(), draft.ID, draft.Version, "planner")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := layouts.PublishRevision(t.Context(), submitted.ID, submitted.Version, "planner"); err != nil {
		t.Fatal(err)
	}
	updateInput["expectedVersion"] = immutable.Version
	immutableUpdate := layoutRequest(t, router, sessions["planner"], http.MethodPut,
		"/api/v1/plan-free-objects/"+immutable.ID, updateInput, true)
	assertProblem(t, immutableUpdate, http.StatusConflict, "track_plan_immutable")
}
