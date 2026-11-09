package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVersionHTTPContractNeverOffersAnUnqualifiedOrOlderUpdate(t *testing.T) {
	for _, contract := range []struct {
		name, body, want string
		status           int
	}{
		{"invalid JSON", "not JSON", "unavailable", 200},
		{"missing version", `{}`, "unavailable", 200},
		{"missing release", `{}`, "no_release", 404},
		{"older release", `{"tag_name":"v0.1.22"}`, "current", 200},
	} {
		t.Run(contract.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(contract.status)
				_, _ = w.Write([]byte(contract.body))
			}))
			defer server.Close()
			router := NewRouter(Config{Version: "0.2.0", UpdateCheckURL: server.URL})
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest("GET", "/api/v1/version?check=true", nil))
			assertStatus(t, response, 200)
			var result versionInfoResponse
			decodeResponse(t, response, &result)
			if result.Status != contract.want || result.UpdateAvailable || result.WindowsPackage != nil || result.Version != "0.2.0" {
				t.Fatalf("unsafe update result: %#v", result)
			}
		})
	}
	router := NewRouter(Config{Version: "0.2.0", UpdateCheckURL: "file:///private/release.json"})
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest("GET", "/api/v1/version?check=true", nil))
	assertStatus(t, response, 200)
	var result versionInfoResponse
	decodeResponse(t, response, &result)
	if result.Status != "unavailable" || result.UpdateAvailable {
		t.Fatalf("unsafe update source accepted: %#v", result)
	}
}
