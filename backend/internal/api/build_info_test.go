package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"railkeeper/backend/internal/buildinfo"
)

func TestVersionBuildIdentityIsIndependentOfUpdateChecksAndRequestHeaders(t *testing.T) {
	expected := buildinfo.Info{Commit: strings.Repeat("a", 40), ShortCommit: strings.Repeat("a", 12), Status: "development", GoVersion: "go-test", Platform: "linux/test"}
	supplied := expected
	router := NewRouter(Config{Version: "0.1.22", BuildInfo: &supplied})
	supplied.Status = "release" // Caller-owned configuration must not mutate an active router.
	for _, path := range []string{"/api/v1/version", "/api/v1/version?check=true"} {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		request.Header.Set("X-Build-Commit", strings.Repeat("b", 40))
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != http.StatusOK {
			t.Fatal(response.Code)
		}
		var body versionInfoResponse
		if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if body.Version != "0.1.22" || body.Build != expected {
			t.Fatalf("source identity changed with update state or caller input: %+v", body)
		}
	}
}
