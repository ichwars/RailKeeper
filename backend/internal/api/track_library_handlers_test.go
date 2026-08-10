package api

import (
	"encoding/json"
	"net/http"
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/domain"
	"railkeeper/backend/internal/infrastructure"
)

func TestTrackLibraryRoutesEnforceReviewWorkflowRolesAndCSRF(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	for _, user := range []application.CreateUserInput{
		{Username: "admin-library", Password: "admin-password", Roles: []string{"Admin"}},
		{Username: "planner-library", Password: "planner-password", Roles: []string{"Planner"}},
		{Username: "viewer-library", Password: "viewer-password", Roles: []string{"Viewer"}},
		{Username: "messe-library", Password: "messe-password", Roles: []string{"Messe"}},
	} {
		if _, err := auth.CreateUser(t.Context(), "", user); err != nil {
			t.Fatal(err)
		}
	}
	repository := infrastructure.NewTrackPlannerRepository(db)
	router := NewRouter(Config{
		AuthService: auth, TrackPlannerService: application.NewTrackPlannerService(repository),
		TrackLibraryService: application.NewTrackLibraryService(repository),
	})
	sessions := map[string]*application.LoginResult{
		"admin":   loginRouteTestUser(t, auth, "admin-library", "admin-password"),
		"planner": loginRouteTestUser(t, auth, "planner-library", "planner-password"),
		"viewer":  loginRouteTestUser(t, auth, "viewer-library", "viewer-password"),
		"messe":   loginRouteTestUser(t, auth, "messe-library", "messe-password"),
	}
	for role, session := range sessions {
		response := layoutRequest(t, router, session, http.MethodGet, "/api/v1/track-libraries", nil, true)
		want := http.StatusOK
		if role == "messe" {
			want = http.StatusForbidden
		}
		assertStatus(t, response, want)
	}
	packageBody := trackLibraryRoutePackage()
	for role, session := range sessions {
		response := layoutRequest(t, router, session, http.MethodPost,
			"/api/v1/track-libraries/import/preview", packageBody, true)
		want := http.StatusForbidden
		if role == "admin" {
			want = http.StatusOK
		}
		assertStatus(t, response, want)
	}
	withoutCSRF := layoutRequest(t, router, sessions["admin"], http.MethodPost,
		"/api/v1/track-libraries/import", map[string]any{
			"confirmed": true, "package": packageBody,
		}, false)
	assertProblem(t, withoutCSRF, http.StatusForbidden, "csrf_required")

	createdResponse := layoutRequest(t, router, sessions["admin"], http.MethodPost,
		"/api/v1/track-libraries/import", map[string]any{
			"confirmed": true, "package": packageBody,
		}, true)
	assertStatus(t, createdResponse, http.StatusCreated)
	var created domain.TrackGeometryLibrary
	if err := json.NewDecoder(createdResponse.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}
	if created.ID == "" || created.Status != domain.TrackGeometryDraft {
		t.Fatalf("unexpected imported library: %#v", created)
	}
	verified := layoutRequest(t, router, sessions["admin"], http.MethodPut,
		"/api/v1/track-libraries/"+created.ID+"/status", map[string]any{
			"confirmed": true, "status": "verified", "verificationNote": "Katalog geprüft",
		}, true)
	assertStatus(t, verified, http.StatusOK)
	exported := layoutRequest(t, router, sessions["viewer"], http.MethodGet,
		"/api/v1/track-libraries/"+created.ID+"/export", nil, true)
	assertStatus(t, exported, http.StatusOK)

	conflict := layoutRequest(t, router, sessions["admin"], http.MethodPost,
		"/api/v1/track-libraries/import", map[string]any{
			"confirmed": true, "package": packageBody,
		}, true)
	assertProblem(t, conflict, http.StatusConflict, "track_library_conflict")
	missing := layoutRequest(t, router, sessions["viewer"], http.MethodGet,
		"/api/v1/track-libraries/missing/export", nil, true)
	assertProblem(t, missing, http.StatusNotFound, "track_library_not_found")
}

func TestTrackLibraryPreviewRejectsUnknownFields(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{
		Username: "admin-library-strict", Password: "admin-password", Roles: []string{"Admin"},
	}); err != nil {
		t.Fatal(err)
	}
	repository := infrastructure.NewTrackPlannerRepository(db)
	router := NewRouter(Config{
		AuthService: auth, TrackLibraryService: application.NewTrackLibraryService(repository),
	})
	session := loginRouteTestUser(t, auth, "admin-library-strict", "admin-password")
	body := trackLibraryRoutePackage()
	body["unknown"] = true
	response := layoutRequest(t, router, session, http.MethodPost,
		"/api/v1/track-libraries/import/preview", body, true)
	assertProblem(t, response, http.StatusBadRequest, "invalid_json")
}

func trackLibraryRoutePackage() map[string]any {
	return map[string]any{
		"format": "railkeeper.track-library", "schemaVersion": 1,
		"library": map[string]any{
			"manufacturer": "Kühn", "trackSystem": "TT", "gauge": "TT", "scale": "1:120",
			"version": "2026.1", "sourceUrl": "https://example.com/catalogue.pdf", "status": "verified",
		},
		"definitions": []any{map[string]any{
			"articleNumber": "72620", "name": "Gerades Gleis", "kind": "straight", "lengthMm": 128,
			"sourceUrl": "https://example.com/72620", "status": "verified",
			"geometry": map[string]any{
				"schemaVersion": 1,
				"ports": []any{
					map[string]any{"id": "a", "xMm": 0, "yMm": 0, "directionDegrees": 180},
					map[string]any{"id": "b", "xMm": 128, "yMm": 0, "directionDegrees": 0},
				},
				"routes": []any{map[string]any{"id": "main", "points": []any{
					map[string]any{"xMm": 0, "yMm": 0}, map[string]any{"xMm": 128, "yMm": 0},
				}}},
			},
		}},
	}
}
