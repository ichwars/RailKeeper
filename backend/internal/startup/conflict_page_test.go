package startup

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestConflictHandlerServesEscapedBilingualSafetyPageForEveryRoute(t *testing.T) {
	info := LegacyConflictInfo{
		SafePath:   `C:\Users\Ada\AppData\Local\RailKeeper\<safe>&data`,
		LegacyPath: `C:\RailKeeper\<legacy>&data`,
		Reason:     "safe and legacy databases differ",
	}
	handler := ConflictHandler(info)

	for _, path := range []string{"/", "/api/v1/setup/status", "/api/v1/version", "/missing"} {
		t.Run(path, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, path, nil)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)

			if response.Code != http.StatusConflict {
				t.Fatalf("GET %s status = %d, want 409", path, response.Code)
			}
			if response.Header().Get("Cache-Control") != "no-store" {
				t.Fatalf("GET %s Cache-Control = %q", path, response.Header().Get("Cache-Control"))
			}
			if policy := response.Header().Get("Content-Security-Policy"); policy !=
				"default-src 'none'; style-src 'unsafe-inline'; base-uri 'none'; form-action 'none'" {
				t.Fatalf("GET %s Content-Security-Policy = %q", path, policy)
			}
			if contentType := response.Header().Get("Content-Type"); !strings.HasPrefix(contentType, "text/html;") {
				t.Fatalf("GET %s Content-Type = %q", path, contentType)
			}
			body := response.Body.String()
			for _, expected := range []string{
				"RailKeeper Sicherheitsstopp",
				"RailKeeper safety stop",
				`C:\Users\Ada\AppData\Local\RailKeeper\&lt;safe&gt;&amp;data`,
				`C:\RailKeeper\&lt;legacy&gt;&amp;data`,
			} {
				if !strings.Contains(body, expected) {
					t.Fatalf("GET %s body does not contain %q", path, expected)
				}
			}
			if strings.Contains(body, `<safe>`) || strings.Contains(body, `<legacy>`) {
				t.Fatalf("GET %s contains unescaped path HTML", path)
			}
			if strings.HasPrefix(strings.TrimSpace(body), "{") {
				t.Fatalf("GET %s exposed a JSON API response", path)
			}
		})
	}
}

func TestConflictHandlerContainsNoInteractiveOrDestructiveAction(t *testing.T) {
	handler := ConflictHandler(LegacyConflictInfo{SafePath: `C:\safe`, LegacyPath: `C:\legacy`})
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	body := strings.ToLower(response.Body.String())

	for _, forbidden := range []string{
		"<form", "<button", "<input", "<script", "href=", "delete", "merge", "overwrite", "continue",
	} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("conflict page contains forbidden action or control %q", forbidden)
		}
	}
}

func TestConflictHandlerRejectsNonGetRequestsWithoutAnAction(t *testing.T) {
	handler := ConflictHandler(LegacyConflictInfo{SafePath: `C:\safe`, LegacyPath: `C:\legacy`})
	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("ignored"))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST status = %d, want 405", response.Code)
	}
	if allow := response.Header().Get("Allow"); allow != http.MethodGet {
		t.Fatalf("POST Allow = %q, want GET", allow)
	}
}
