package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestExhibitionWorkspaceRoutesPersistLifecycleAndExceptions(t *testing.T) {
	db := testRouterDB(t)
	setup := application.NewSetupService(db)
	auth := application.NewAuthService(db)
	exhibition := application.NewExhibitionService(db)
	if err := setup.CreateAdmin(t.Context(), application.CreateAdminInput{
		Username: "admin", Email: "admin@example.test", Password: "very-secure-password",
	}); err != nil {
		t.Fatal(err)
	}
	list, err := exhibition.Create(t.Context(), application.ExhibitionListInput{
		Designation: "Köln", Date: "2026-08-22", EndDate: "2026-08-24",
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"BR 103", "V 200"} {
		if _, err := exhibition.CreateEntry(t.Context(), list.ID, application.ExhibitionEntryInput{
			Owner: name, LocomotiveName: name, DayScope: "day1", DTDecoder: true,
			DecoderNumber: "3", InterfaceName: "ECoS",
		}); err != nil {
			t.Fatal(err)
		}
	}

	router := NewRouter(Config{SetupService: setup, AuthService: auth, ExhibitionService: exhibition})
	session, cookies := loginTestUser(t, router, "admin", "very-secure-password")
	workspaceResponse := exhibitionRequest(t, router, cookies, session.CSRFToken, http.MethodGet,
		"/api/v1/exhibition-lists/"+list.ID+"/workspace", "")
	if workspaceResponse.Code != http.StatusOK {
		t.Fatalf("workspace status = %d: %s", workspaceResponse.Code, workspaceResponse.Body.String())
	}
	var workspace application.ExhibitionWorkspace
	if err := json.NewDecoder(workspaceResponse.Body).Decode(&workspace); err != nil {
		t.Fatal(err)
	}
	if len(workspace.Conflicts) != 1 {
		t.Fatalf("workspace conflicts = %#v", workspace.Conflicts)
	}

	exceptionResponse := exhibitionRequest(t, router, cookies, session.CSRFToken, http.MethodPut,
		"/api/v1/exhibition-lists/"+list.ID+"/conflicts/"+workspace.Conflicts[0].ID+"/exception",
		`{"reason":"Getrennte Boosterbezirke","expectedRevision":`+itoa(workspace.List.Revision)+`}`)
	if exceptionResponse.Code != http.StatusOK {
		t.Fatalf("exception status = %d: %s", exceptionResponse.Code, exceptionResponse.Body.String())
	}
	if err := json.NewDecoder(exceptionResponse.Body).Decode(&workspace); err != nil {
		t.Fatal(err)
	}
	if !workspace.Conflicts[0].Excepted {
		t.Fatalf("exception not persisted: %#v", workspace.Conflicts[0])
	}

	statusResponse := exhibitionRequest(t, router, cookies, session.CSRFToken, http.MethodPut,
		"/api/v1/exhibition-lists/"+list.ID+"/status",
		`{"status":"locked","expectedRevision":`+itoa(workspace.List.Revision)+`,"confirmConflicts":true,"reason":"Freigabe"}`)
	if statusResponse.Code != http.StatusOK {
		t.Fatalf("status update = %d: %s", statusResponse.Code, statusResponse.Body.String())
	}
}

func exhibitionRequest(
	t *testing.T,
	router http.Handler,
	cookies []*http.Cookie,
	csrfToken string,
	method string,
	path string,
	body string,
) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	if method != http.MethodGet {
		request.Header.Set("X-CSRF-Token", csrfToken)
	}
	for _, cookie := range cookies {
		request.AddCookie(cookie)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func itoa(value int) string {
	data, _ := json.Marshal(value)
	return string(data)
}
