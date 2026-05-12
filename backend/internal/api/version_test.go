package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVersionInfoWithoutConfiguredUpdateSource(t *testing.T) {
	router := NewRouter(Config{Version: "0.1.0"})
	request := httptest.NewRequest(http.MethodGet, "/api/v1/version?check=true", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	var body versionInfoResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Status != "not_configured" {
		t.Fatalf("expected not_configured status, got %q", body.Status)
	}
	if body.UpdateAvailable {
		t.Fatal("expected no update when no update source is configured")
	}
}

func TestVersionInfoDetectsUpdate(t *testing.T) {
	updateServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"tag_name":"v0.2.0","html_url":"https://example.test/releases/v0.2.0"}`))
	}))
	defer updateServer.Close()

	router := NewRouter(Config{Version: "0.1.0", UpdateCheckURL: updateServer.URL})
	request := httptest.NewRequest(http.MethodGet, "/api/v1/version?check=true", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	var body versionInfoResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Status != "update_available" {
		t.Fatalf("expected update_available status, got %q", body.Status)
	}
	if !body.UpdateAvailable {
		t.Fatal("expected update to be detected")
	}
	if body.LatestVersion != "v0.2.0" {
		t.Fatalf("expected latest version v0.2.0, got %q", body.LatestVersion)
	}
}
