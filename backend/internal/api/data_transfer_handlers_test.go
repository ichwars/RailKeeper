package api

import (
	"net/http"
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/infrastructure"
)

func TestDataTransferProfileRoutesEnforceRolesAndMesseScope(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	for _, user := range []application.CreateUserInput{
		{Username: "admin", Password: "admin-password", Roles: []string{"Admin"}},
		{Username: "editor", Password: "editor-password", Roles: []string{"Editor"}},
		{Username: "viewer", Password: "viewer-password", Roles: []string{"Viewer"}},
		{Username: "messe", Password: "messe-password", Roles: []string{"Messe"}},
	} {
		if _, err := auth.CreateUser(t.Context(), "", user); err != nil {
			t.Fatal(err)
		}
	}
	service := application.NewDataTransferService(infrastructure.NewDataTransferRepository(db), t.TempDir())
	if _, err := service.CreateProfile(t.Context(), application.CreateDataTransferProfileInput{
		Name: "Vehicles", Direction: application.TransferExport, Format: application.TransferCSV,
		Areas: []application.TransferArea{application.TransferVehicles},
	}, "admin-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.CreateProfile(t.Context(), application.CreateDataTransferProfileInput{
		Name: "Exhibition", Direction: application.TransferExport, Format: application.TransferJSON,
		Areas: []application.TransferArea{application.TransferExhibitionLists},
	}, "admin-1"); err != nil {
		t.Fatal(err)
	}
	router := NewRouter(Config{AuthService: auth, DataTransferService: service})
	sessions := map[string]*application.LoginResult{
		"admin":  loginRouteTestUser(t, auth, "admin", "admin-password"),
		"editor": loginRouteTestUser(t, auth, "editor", "editor-password"),
		"viewer": loginRouteTestUser(t, auth, "viewer", "viewer-password"),
		"messe":  loginRouteTestUser(t, auth, "messe", "messe-password"),
	}

	viewerProfiles := layoutRequest(t, router, sessions["viewer"], http.MethodGet,
		"/api/v1/data-transfer/profiles", nil, true)
	assertStatus(t, viewerProfiles, http.StatusOK)
	var listed []application.DataTransferProfile
	decodeResponse(t, viewerProfiles, &listed)
	if len(listed) != 2 {
		t.Fatalf("viewer profiles = %#v, want all profiles", listed)
	}

	viewerCreate := layoutRequest(t, router, sessions["viewer"], http.MethodPost,
		"/api/v1/data-transfer/profiles", validDataTransferProfileRequest("Viewer"), true)
	assertProblem(t, viewerCreate, http.StatusForbidden, "forbidden")

	editorCreate := layoutRequest(t, router, sessions["editor"], http.MethodPost,
		"/api/v1/data-transfer/profiles", validDataTransferProfileRequest("Editor"), true)
	assertStatus(t, editorCreate, http.StatusCreated)

	messeProfiles := layoutRequest(t, router, sessions["messe"], http.MethodGet,
		"/api/v1/data-transfer/profiles", nil, true)
	assertStatus(t, messeProfiles, http.StatusOK)
	decodeResponse(t, messeProfiles, &listed)
	if len(listed) != 1 || len(listed[0].Areas) != 1 || listed[0].Areas[0] != application.TransferExhibitionLists {
		t.Fatalf("messe profiles = %#v, want only exhibition-list profiles", listed)
	}

	invalid := layoutRequest(t, router, sessions["editor"], http.MethodPost, "/api/v1/data-transfer/profiles",
		map[string]any{"name": "Master data", "direction": "export", "format": "railkeeper-json",
			"areas": []string{"masterData"}}, true)
	assertProblem(t, invalid, http.StatusBadRequest, "data_transfer_validation")
}

func TestDataTransferProfileRoutesUpdateAndDisableProfiles(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{
		Username: "admin", Password: "admin-password", Roles: []string{"Admin"},
	}); err != nil {
		t.Fatal(err)
	}
	service := application.NewDataTransferService(infrastructure.NewDataTransferRepository(db), t.TempDir())
	profile, err := service.CreateProfile(t.Context(), application.CreateDataTransferProfileInput{
		Name: "Vehicles", Direction: application.TransferExport, Format: application.TransferCSV,
		Areas: []application.TransferArea{application.TransferVehicles},
	}, "admin-1")
	if err != nil {
		t.Fatal(err)
	}
	router := NewRouter(Config{AuthService: auth, DataTransferService: service})
	session := loginRouteTestUser(t, auth, "admin", "admin-password")

	updated := layoutRequest(t, router, session, http.MethodPut,
		"/api/v1/data-transfer/profiles/"+profile.ID,
		map[string]any{"name": "Vehicles and accessories", "direction": "export", "format": "railkeeper-json",
			"areas": []string{"vehicles", "accessories"}}, true)
	assertStatus(t, updated, http.StatusOK)

	disabled := layoutRequest(t, router, session, http.MethodDelete,
		"/api/v1/data-transfer/profiles/"+profile.ID, nil, true)
	assertStatus(t, disabled, http.StatusNoContent)
	profiles, err := service.ListProfiles(t.Context())
	if err != nil || len(profiles) != 1 || profiles[0].Enabled {
		t.Fatalf("unexpected disabled profile state: %#v, %v", profiles, err)
	}
}

func validDataTransferProfileRequest(name string) map[string]any {
	return map[string]any{
		"name": name, "direction": "export", "format": "csv", "areas": []string{"vehicles"},
	}
}
