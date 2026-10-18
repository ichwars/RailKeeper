package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"railkeeper/backend/internal/application"
)

func authedJSONRequest(
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
