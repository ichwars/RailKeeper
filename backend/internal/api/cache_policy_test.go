package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestCachePolicyForAppAndPrivateAPI(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "assets"), 0o700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"index.html", "favicon.ico", "assets/main-ABC123.js"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("railkeeper"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	router := NewRouter(Config{StaticDir: root, AuthService: application.NewAuthService(testRouterDB(t))})
	for _, test := range []struct {
		path, cache string
		status      int
	}{
		{"/api/v1/auth/session", "no-store", http.StatusUnauthorized},
		{"/api/v1/version", "no-store", http.StatusOK},
		{"/api/v1/vehicles", "no-store", http.StatusUnauthorized},
		{"/api/v1/missing", "no-store", http.StatusNotFound},
		{"/overview", "no-store", http.StatusOK},
		{"/favicon.ico", "no-cache", http.StatusOK},
		{"/assets/main-ABC123.js", "public, max-age=31536000, immutable", http.StatusOK},
	} {
		t.Run(test.path, func(t *testing.T) {
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, test.path, nil))
			if got := response.Header().Get("Cache-Control"); response.Code != test.status || got != test.cache {
				t.Fatalf("HTTP %d with Cache-Control %q, want HTTP %d with %q", response.Code, got, test.status, test.cache)
			}
		})
	}
}
