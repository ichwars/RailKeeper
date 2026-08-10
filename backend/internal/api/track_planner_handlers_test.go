package api

import (
	"fmt"
	"math"
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
			"/api/v1/plan-revisions/"+draft.ID+"/track-reservations", map[string]any{
				"confirmed": false, "items": []any{},
			}, true)
		want := http.StatusForbidden
		if role == "admin" || role == "planner" {
			want = http.StatusBadRequest
		}
		assertStatus(t, response, want)
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

	for role, session := range sessions {
		response := layoutRequest(t, router, session, http.MethodPost,
			"/api/v1/plan-track-objects/missing/flex-preview", map[string]any{
				"endXMm": 500, "endYMm": 100, "endDirectionDegrees": 20, "expectedVersion": 1,
			}, true)
		want := http.StatusForbidden
		if role == "admin" || role == "planner" {
			want = http.StatusNotFound
		}
		assertStatus(t, response, want)
	}
	previewWithoutCSRF := layoutRequest(t, router, sessions["planner"], http.MethodPost,
		"/api/v1/plan-track-objects/missing/flex-preview", map[string]any{
			"endXMm": 500, "endYMm": 100, "expectedVersion": 1,
		}, false)
	assertProblem(t, previewWithoutCSRF, http.StatusForbidden, "csrf_required")
}

func TestTrackPlannerFlexPreviewAndUpdateWorkflow(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{
		Username: "planner-flex", Password: "planner-password", Roles: []string{"Planner"},
	}); err != nil {
		t.Fatal(err)
	}
	session := loginRouteTestUser(t, auth, "planner-flex", "planner-password")
	layouts := application.NewLayoutService(infrastructure.NewLayoutRepository(db))
	planner := application.NewTrackPlannerService(infrastructure.NewTrackPlannerRepository(db))
	router := NewRouter(Config{AuthService: auth, LayoutService: layouts, TrackPlannerService: planner})
	_, draft := trackPlannerRouteRevision(t, layouts)
	path := map[string]any{
		"schemaVersion": 1, "endXMm": 664, "endYMm": 0, "endDirectionDegrees": 0,
		"startHandleMm": 664.0 / 3, "endHandleMm": 664.0 / 3,
	}
	createdResponse := layoutRequest(t, router, session, http.MethodPost,
		"/api/v1/plan-revisions/"+draft.ID+"/track-objects", map[string]any{
			"geometryId": "tillig-tt-modellgleis-83125-v1", "flexPath": path,
		}, true)
	assertStatus(t, createdResponse, http.StatusCreated)
	var created domain.PlanTrackObject
	decodeResponse(t, createdResponse, &created)

	previewResponse := layoutRequest(t, router, session, http.MethodPost,
		"/api/v1/plan-track-objects/"+created.ID+"/flex-preview", map[string]any{
			"endXMm": 500, "endYMm": 100, "endDirectionDegrees": 20,
			"expectedVersion": created.Version,
		}, true)
	assertStatus(t, previewResponse, http.StatusOK)
	var preview application.FlexTrackPreview
	decodeResponse(t, previewResponse, &preview)
	if !preview.Applicable || preview.Path.EndXMM != 500 || preview.EffectiveLengthMM <= 0 {
		t.Fatalf("unexpected flex preview: %#v", preview)
	}

	updatedResponse := layoutRequest(t, router, session, http.MethodPut,
		"/api/v1/plan-track-objects/"+created.ID, map[string]any{
			"flexPath": preview.Path, "expectedVersion": created.Version,
		}, true)
	assertStatus(t, updatedResponse, http.StatusOK)
	var updated domain.PlanTrackObject
	decodeResponse(t, updatedResponse, &updated)
	if updated.FlexPath == nil || updated.FlexPath.EndXMM != 500 || updated.Version != 2 {
		t.Fatalf("unexpected flex update: %#v", updated)
	}

	staleResponse := layoutRequest(t, router, session, http.MethodPost,
		"/api/v1/plan-track-objects/"+created.ID+"/flex-preview", map[string]any{
			"endXMm": 400, "endYMm": 80, "endDirectionDegrees": 10,
			"expectedVersion": created.Version,
		}, true)
	assertProblem(t, staleResponse, http.StatusConflict, "track_plan_conflict")
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
	if len(geometries) != 2 || geometries[0].ArticleNumber != "83101" ||
		geometries[1].ArticleNumber != "83125" || geometries[1].Kind != domain.TrackGeometryFlex ||
		geometries[1].MinimumRadiusMM == nil || *geometries[1].MinimumRadiusMM != 543 {
		t.Fatalf("unexpected track geometries: %#v", geometries)
	}

	createdResponse := layoutRequest(t, router, session, http.MethodPost,
		"/api/v1/plan-revisions/"+draft.ID+"/track-objects", map[string]any{
			"geometryId": geometries[0].ID, "positionXMm": 100, "positionYMm": 50,
			"rotationDegrees": -15, "elevationStartMm": -2, "elevationEndMm": 2.15,
		}, true)
	assertStatus(t, createdResponse, http.StatusCreated)
	var created domain.PlanTrackObject
	decodeResponse(t, createdResponse, &created)
	if created.Version != 1 || created.RotationDegrees != 345 ||
		created.ElevationStartMM != -2 || created.ElevationEndMM != 2.15 {
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
			"elevationStartMm": 0, "elevationEndMm": 4.15, "expectedVersion": created.Version,
		}, true)
	assertStatus(t, updatedResponse, http.StatusOK)
	var updated domain.PlanTrackObject
	decodeResponse(t, updatedResponse, &updated)
	if updated.Version != 2 || updated.PositionXMM != 110 || updated.ElevationEndMM != 4.15 {
		t.Fatalf("unexpected updated track object: %#v", updated)
	}
	secondResponse := layoutRequest(t, router, session, http.MethodPost,
		"/api/v1/plan-revisions/"+draft.ID+"/track-objects", map[string]any{
			"geometryId": geometries[0].ID, "positionXMm": 276, "positionYMm": 55,
			"rotationDegrees": 0, "elevationStartMm": 6.15, "elevationEndMm": 6.15,
		}, true)
	assertStatus(t, secondResponse, http.StatusCreated)
	var second domain.PlanTrackObject
	decodeResponse(t, secondResponse, &second)

	analysisResponse := layoutRequest(t, router, session, http.MethodGet,
		"/api/v1/plan-revisions/"+draft.ID+"/track-analysis", nil, true)
	assertStatus(t, analysisResponse, http.StatusOK)
	var analysis application.TrackPlanAnalysis
	decodeResponse(t, analysisResponse, &analysis)
	if analysis.RevisionID != draft.ID || len(analysis.Connections) != 1 ||
		len(analysis.BOM) != 1 || analysis.BOM[0].Quantity != 2 || len(analysis.Grades) != 2 {
		t.Fatalf("unexpected track plan analysis: %#v", analysis)
	}
	updatedGradeFound := false
	for _, grade := range analysis.Grades {
		if grade.ObjectID == updated.ID && grade.GradePercent == 2.5 {
			updatedGradeFound = true
		}
	}
	if !updatedGradeFound {
		t.Fatalf("updated track grade missing from analysis: %#v", analysis.Grades)
	}
	elevationIssues := 0
	gradeLimitIssues := 0
	for _, issue := range analysis.Issues {
		switch issue.Code {
		case domain.TrackPlanIssueElevationMismatch:
			elevationIssues++
			if issue.ElevationDifferenceMM == nil || *issue.ElevationDifferenceMM != 2 {
				t.Fatalf("unexpected elevation mismatch detail: %#v", issue)
			}
		case domain.TrackPlanIssueGradeLimitExceeded:
			gradeLimitIssues++
			if issue.GradePercent == nil || *issue.GradePercent != 2.5 ||
				issue.GradeLimitPercent == nil || *issue.GradeLimitPercent != 2 {
				t.Fatalf("unexpected grade limit detail: %#v", issue)
			}
		}
	}
	if elevationIssues != 1 {
		t.Fatalf("expected one elevation mismatch, got %#v", analysis.Issues)
	}
	if gradeLimitIssues != 1 {
		t.Fatalf("expected one grade limit warning, got %#v", analysis.Issues)
	}
	if _, err := db.Exec(`
INSERT INTO storage_locations(id, name, created_at, updated_at)
VALUES('route-track-location', 'Gleislager', 'now', 'now');
INSERT INTO accessory_products(
  id, inventory_number, manufacturer, article_number, name, category, tracking_mode,
  created_at, updated_at
) VALUES(
  'route-track-product', 'RK-ART-0083101', 'Tillig', '83101', 'Gleisstück G1', 'track', 'quantity',
  'now', 'now'
);
INSERT INTO accessory_stock(product_id, location_id, quantity, updated_at)
VALUES('route-track-product', 'route-track-location', 2, 'now');`); err != nil {
		t.Fatal(err)
	}
	reservationResponse := layoutRequest(t, router, session, http.MethodPost,
		"/api/v1/plan-revisions/"+draft.ID+"/track-reservations", map[string]any{
			"confirmed": true,
			"items": []map[string]any{
				{"trackObjectId": created.ID, "productId": "route-track-product",
					"locationId": "route-track-location", "expectedObjectVersion": updated.Version},
				{"trackObjectId": second.ID, "productId": "route-track-product",
					"locationId": "route-track-location", "expectedObjectVersion": second.Version},
			},
		}, true)
	assertStatus(t, reservationResponse, http.StatusCreated)
	var reservationBatch application.TrackPlanReservationBatch
	decodeResponse(t, reservationResponse, &reservationBatch)
	if len(reservationBatch.Reservations) != 2 {
		t.Fatalf("unexpected route reservation batch: %#v", reservationBatch)
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
	assertProblem(t, deleteResponse, http.StatusConflict, "track_plan_conflict")

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

func TestTrackPlannerRoutesExposeInsufficientClearanceDetails(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{
		Username: "planner-clearance", Password: "planner-password", Roles: []string{"Planner"},
	}); err != nil {
		t.Fatal(err)
	}
	session := loginRouteTestUser(t, auth, "planner-clearance", "planner-password")
	layouts := application.NewLayoutService(infrastructure.NewLayoutRepository(db))
	planner := application.NewTrackPlannerService(infrastructure.NewTrackPlannerRepository(db))
	router := NewRouter(Config{AuthService: auth, LayoutService: layouts, TrackPlannerService: planner})
	_, draft := trackPlannerRouteRevision(t, layouts)

	for _, input := range []application.CreatePlanTrackObjectInput{
		{GeometryID: "tillig-tt-modellgleis-83101-v1", ElevationStartMM: 0, ElevationEndMM: 0},
		{GeometryID: "tillig-tt-modellgleis-83101-v1", PositionXMM: 83, PositionYMM: -83,
			RotationDegrees: 90, ElevationStartMM: 25, ElevationEndMM: 25},
	} {
		if _, err := planner.CreateObject(t.Context(), draft.ID, input, "planner-clearance"); err != nil {
			t.Fatal(err)
		}
	}

	response := layoutRequest(t, router, session, http.MethodGet,
		"/api/v1/plan-revisions/"+draft.ID+"/track-analysis", nil, true)
	assertStatus(t, response, http.StatusOK)
	var analysis application.TrackPlanAnalysis
	decodeResponse(t, response, &analysis)
	issues := make([]domain.TrackPlanIssue, 0)
	for _, issue := range analysis.Issues {
		if issue.Code == domain.TrackPlanIssueInsufficientClearance {
			issues = append(issues, issue)
		}
	}
	if len(issues) != 1 || issues[0].ClearanceMM == nil ||
		math.Abs(*issues[0].ClearanceMM-25) > 1e-9 ||
		issues[0].ClearanceLimitMM == nil || *issues[0].ClearanceLimitMM != 40 ||
		issues[0].IntersectionXMM == nil || math.Abs(*issues[0].IntersectionXMM-83) > 1e-9 ||
		issues[0].IntersectionYMM == nil || math.Abs(*issues[0].IntersectionYMM) > 1e-9 {
		t.Fatalf("unexpected clearance response: %#v", issues)
	}
}

func trackPlannerRouteRevision(
	t *testing.T,
	layouts *application.LayoutService,
) (*application.PlanVariant, *application.PlanRevision) {
	t.Helper()
	maxGradePercent := 2.0
	minimumTrackClearanceMM := 40.0
	layout, err := layouts.CreateLayout(t.Context(), application.CreateLayoutInput{
		Name: "Track API", Kind: domain.LayoutKindPrivate, Gauge: "TT", Scale: "1:120",
		MaxGradePercent: &maxGradePercent, MinimumTrackClearanceMM: &minimumTrackClearanceMM,
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
