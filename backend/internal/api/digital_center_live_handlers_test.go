package api

import (
	"context"
	"net/http"
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/infrastructure"
)

func TestDigitalCenterLiveRoutesUseConfiguredTargetExposeMessagesAndRejectUnsupportedProvider(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{
		Username: "live-admin", Password: "admin-password", Roles: []string{"Admin"},
	}); err != nil {
		t.Fatal(err)
	}
	settings := application.NewSettingsService(db)
	if _, err := settings.UpdateDigitalSettings(t.Context(), application.DigitalCenterSettings{
		Provider: "ecos",
		ECoS: application.DigitalProviderSettings{
			Enabled: true, Host: "configured-center.local", Port: "15471",
		},
		Z21: application.DigitalProviderSettings{Enabled: true, Host: "z21.local", Port: "21105"},
	}); err != nil {
		t.Fatal(err)
	}
	repository := infrastructure.NewDigitalCenterWorkspaceRepository(db)
	session, err := repository.CreateSession(t.Context(), application.DigitalCenterReadSession{
		Provider: "ecos", State: application.DigitalCenterSessionReady,
	})
	if err != nil {
		t.Fatal(err)
	}
	ecos := &apiWorkspaceLiveECoS{status: application.ECoSLiveStatus{
		Provider: "ecos", Connected: true, State: application.ECoSLiveRunning,
	}}
	workspace := application.NewDigitalCenterWorkspaceService(
		repository, settings, ecos, nil, &apiWorkspaceVehicleReader{}, auth,
	)
	router := NewRouter(Config{AuthService: auth, DigitalCenterWorkspace: workspace})
	admin := loginRouteTestUser(t, auth, "live-admin", "admin-password")

	started := layoutRequest(t, router, admin, http.MethodPost,
		"/api/v1/digital-centers/ecos/live/start?sessionId="+session.ID,
		map[string]any{"host": "attacker.invalid", "port": 1}, true)
	assertStatus(t, started, http.StatusOK)
	if ecos.startInput != (application.ECoSConnectionInput{Host: "configured-center.local", Port: 15471}) {
		t.Fatalf("live target = %#v, request body overrode server settings", ecos.startInput)
	}

	statusResponse := layoutRequest(t, router, admin, http.MethodGet,
		"/api/v1/digital-centers/ecos/live/status", nil, true)
	assertStatus(t, statusResponse, http.StatusOK)

	messagesResponse := layoutRequest(t, router, admin, http.MethodGet,
		"/api/v1/digital-centers/read-sessions/"+session.ID+"/messages", nil, true)
	assertStatus(t, messagesResponse, http.StatusOK)
	var messages struct {
		Messages []application.DigitalCenterSessionMessage `json:"messages"`
	}
	decodeResponse(t, messagesResponse, &messages)
	if len(messages.Messages) != 1 || messages.Messages[0].Code != application.DigitalCenterMessageLiveStarted {
		t.Fatalf("messages = %#v", messages.Messages)
	}

	stopped := layoutRequest(t, router, admin, http.MethodPost,
		"/api/v1/digital-centers/ecos/live/stop?sessionId="+session.ID, nil, true)
	assertStatus(t, stopped, http.StatusOK)

	unsupported := layoutRequest(t, router, admin, http.MethodPost,
		"/api/v1/digital-centers/z21/live/start?sessionId="+session.ID, nil, true)
	assertProblem(t, unsupported, http.StatusBadRequest, "digital_center_capability_unavailable")
}

func TestDigitalCenterLiveRoutesRequireAdminAndCSRF(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	for _, user := range []application.CreateUserInput{
		{Username: "live-admin-csrf", Password: "admin-password", Roles: []string{"Admin"}},
		{Username: "live-viewer", Password: "viewer-password", Roles: []string{"Viewer"}},
	} {
		if _, err := auth.CreateUser(t.Context(), "", user); err != nil {
			t.Fatal(err)
		}
	}
	workspace := application.NewDigitalCenterWorkspaceService(
		infrastructure.NewDigitalCenterWorkspaceRepository(db), application.NewSettingsService(db),
		&apiWorkspaceLiveECoS{}, nil, &apiWorkspaceVehicleReader{}, auth,
	)
	router := NewRouter(Config{AuthService: auth, DigitalCenterWorkspace: workspace})
	viewer := loginRouteTestUser(t, auth, "live-viewer", "viewer-password")
	admin := loginRouteTestUser(t, auth, "live-admin-csrf", "admin-password")

	denied := layoutRequest(t, router, viewer, http.MethodGet,
		"/api/v1/digital-centers/ecos/live/status", nil, true)
	assertProblem(t, denied, http.StatusForbidden, "forbidden")
	csrf := layoutRequest(t, router, admin, http.MethodPost,
		"/api/v1/digital-centers/ecos/live/start", nil, false)
	assertProblem(t, csrf, http.StatusForbidden, "csrf_required")
}

type apiWorkspaceLiveECoS struct {
	probe      application.ECoSRawProbe
	status     application.ECoSLiveStatus
	startInput application.ECoSConnectionInput
}

func (ecos *apiWorkspaceLiveECoS) ProbeLocomotiveRaw(
	context.Context,
	application.ECoSConnectionInput,
) (*application.ECoSRawProbe, error) {
	probe := ecos.probe
	return &probe, nil
}

func (ecos *apiWorkspaceLiveECoS) StartLive(
	_ context.Context,
	input application.ECoSConnectionInput,
) (*application.ECoSLiveStatus, error) {
	ecos.startInput = input
	status := ecos.status
	return &status, nil
}

func (ecos *apiWorkspaceLiveECoS) StopLive() application.ECoSLiveStatus {
	ecos.status.Connected = false
	ecos.status.State = application.ECoSLiveStopped
	return ecos.status
}

func (ecos *apiWorkspaceLiveECoS) LiveStatus() application.ECoSLiveStatus {
	return ecos.status
}
