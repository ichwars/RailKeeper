package api

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/infrastructure"
)

func TestDigitalCenterWorkspaceRoutesUseConfiguredTargetAndBoundPagination(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{
		Username: "center-admin", Password: "admin-password", Roles: []string{"Admin"},
	}); err != nil {
		t.Fatal(err)
	}
	settings := application.NewSettingsService(db)
	if _, err := settings.UpdateDigitalSettings(t.Context(), application.DigitalCenterSettings{
		Provider: "ecos",
		ECoS: application.DigitalProviderSettings{
			Enabled: true, Host: "configured-center.local", Port: "15471",
		},
	}); err != nil {
		t.Fatal(err)
	}
	ecos := &apiWorkspaceECoSReader{}
	for index := 1; index <= 125; index++ {
		ecos.probe.Locomotives = append(ecos.probe.Locomotives, application.ECoSRawLocomotive{
			ObjectID: index, Name: fmt.Sprintf("Lok %03d", index), Address: index, Protocol: "DCC128",
		})
	}
	workspace := application.NewDigitalCenterWorkspaceService(
		infrastructure.NewDigitalCenterWorkspaceRepository(db),
		settings,
		ecos,
		nil,
		&apiWorkspaceVehicleReader{},
		auth,
	)
	router := NewRouter(Config{AuthService: auth, DigitalCenterWorkspace: workspace})
	admin := loginRouteTestUser(t, auth, "center-admin", "admin-password")

	summaryResponse := layoutRequest(t, router, admin, http.MethodGet,
		"/api/v1/digital-centers/workspace", nil, true)
	assertStatus(t, summaryResponse, http.StatusOK)
	var summary struct {
		Centers []application.DigitalCenterSummary `json:"centers"`
	}
	decodeResponse(t, summaryResponse, &summary)
	if len(summary.Centers) != 1 || summary.Centers[0].Host != "configured-center.local" {
		t.Fatalf("workspace summary = %#v", summary)
	}

	started := layoutRequest(t, router, admin, http.MethodPost,
		"/api/v1/digital-centers/ecos/read-sessions",
		map[string]any{"host": "attacker.invalid", "port": 1}, true)
	assertStatus(t, started, http.StatusCreated)
	var session application.DigitalCenterReadSession
	decodeResponse(t, started, &session)
	if ecos.input != (application.ECoSConnectionInput{Host: "configured-center.local", Port: 15471}) {
		t.Fatalf("ECoS target = %#v, request body overrode server settings", ecos.input)
	}
	if session.State != application.DigitalCenterSessionReady || session.ID == "" {
		t.Fatalf("started session = %#v", session)
	}

	sessionResponse := layoutRequest(t, router, admin, http.MethodGet,
		"/api/v1/digital-centers/read-sessions/"+session.ID, nil, true)
	assertStatus(t, sessionResponse, http.StatusOK)

	itemsResponse := layoutRequest(t, router, admin, http.MethodGet,
		"/api/v1/digital-centers/read-sessions/"+session.ID+
			"/items?q=Lok&compareStatus=new&page=1&pageSize=1000", nil, true)
	assertStatus(t, itemsResponse, http.StatusOK)
	var page application.DigitalCenterWorkItemPage
	decodeResponse(t, itemsResponse, &page)
	if page.PageSize != 100 || len(page.Items) != 100 || page.Total != 125 || page.TotalPages != 2 {
		t.Fatalf("work item page = %#v", page)
	}

	itemResponse := layoutRequest(t, router, admin, http.MethodGet,
		"/api/v1/digital-centers/read-sessions/"+session.ID+"/items/"+page.Items[0].ID, nil, true)
	assertStatus(t, itemResponse, http.StatusOK)
	var item application.DigitalCenterWorkItem
	decodeResponse(t, itemResponse, &item)
	if item.ID != page.Items[0].ID || item.CompareStatus != application.DigitalCompareNew {
		t.Fatalf("work item detail = %#v", item)
	}
}

func TestDigitalCenterWorkspaceRoutesDenyViewerAndMesse(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	for _, user := range []application.CreateUserInput{
		{Username: "center-viewer", Password: "viewer-password", Roles: []string{"Viewer"}},
		{Username: "center-messe", Password: "messe-password", Roles: []string{"Messe"}},
	} {
		if _, err := auth.CreateUser(t.Context(), "", user); err != nil {
			t.Fatal(err)
		}
	}
	workspace := application.NewDigitalCenterWorkspaceService(
		infrastructure.NewDigitalCenterWorkspaceRepository(db),
		application.NewSettingsService(db),
		&apiWorkspaceECoSReader{},
		nil,
		&apiWorkspaceVehicleReader{},
		auth,
	)
	router := NewRouter(Config{AuthService: auth, DigitalCenterWorkspace: workspace})

	for _, user := range []struct {
		username string
		password string
	}{
		{username: "center-viewer", password: "viewer-password"},
		{username: "center-messe", password: "messe-password"},
	} {
		session := loginRouteTestUser(t, auth, user.username, user.password)
		for _, request := range []struct {
			method string
			path   string
		}{
			{method: http.MethodGet, path: "/api/v1/digital-centers/workspace"},
			{method: http.MethodPost, path: "/api/v1/digital-centers/ecos/read-sessions"},
		} {
			response := layoutRequest(t, router, session, request.method, request.path, nil, true)
			assertProblem(t, response, http.StatusForbidden, "forbidden")
		}
	}
}

func TestDigitalCenterReadSessionRouteRequiresCSRF(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{
		Username: "csrf-admin", Password: "admin-password", Roles: []string{"Admin"},
	}); err != nil {
		t.Fatal(err)
	}
	router := NewRouter(Config{
		AuthService: auth,
		DigitalCenterWorkspace: application.NewDigitalCenterWorkspaceService(
			infrastructure.NewDigitalCenterWorkspaceRepository(db),
			application.NewSettingsService(db),
			&apiWorkspaceECoSReader{},
			nil,
			&apiWorkspaceVehicleReader{},
			auth,
		),
	})
	admin := loginRouteTestUser(t, auth, "csrf-admin", "admin-password")

	response := layoutRequest(t, router, admin, http.MethodPost,
		"/api/v1/digital-centers/ecos/read-sessions", nil, false)
	assertProblem(t, response, http.StatusForbidden, "csrf_required")
}

func TestDigitalCenterWorkItemRouteRejectsUnknownFilter(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{
		Username: "filter-admin", Password: "admin-password", Roles: []string{"Admin"},
	}); err != nil {
		t.Fatal(err)
	}
	repository := infrastructure.NewDigitalCenterWorkspaceRepository(db)
	readSession, err := repository.CreateSession(t.Context(), application.DigitalCenterReadSession{
		Provider: "ecos", State: application.DigitalCenterSessionReady,
	})
	if err != nil {
		t.Fatal(err)
	}
	workspace := application.NewDigitalCenterWorkspaceService(
		repository, application.NewSettingsService(db), &apiWorkspaceECoSReader{}, nil,
		&apiWorkspaceVehicleReader{}, auth,
	)
	router := NewRouter(Config{AuthService: auth, DigitalCenterWorkspace: workspace})
	admin := loginRouteTestUser(t, auth, "filter-admin", "admin-password")

	response := layoutRequest(t, router, admin, http.MethodGet,
		"/api/v1/digital-centers/read-sessions/"+readSession.ID+"/items?compareStatus=unsafe", nil, true)
	assertProblem(t, response, http.StatusBadRequest, "digital_center_filter_invalid")
}

type apiWorkspaceECoSReader struct {
	probe application.ECoSRawProbe
	input application.ECoSConnectionInput
}

func (reader *apiWorkspaceECoSReader) ProbeLocomotiveRaw(
	_ context.Context,
	input application.ECoSConnectionInput,
) (*application.ECoSRawProbe, error) {
	reader.input = input
	probe := reader.probe
	return &probe, nil
}

type apiWorkspaceVehicleReader struct{}

func (*apiWorkspaceVehicleReader) List(context.Context, string) ([]application.Vehicle, error) {
	return []application.Vehicle{}, nil
}
