package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/startup"
)

func TestSetupDoesNotExposeCodeAndRequiresItsAuthorization(t *testing.T) {
	db := testRouterDB(t)
	setup := application.NewSetupService(db)
	const code = "synthetic-local-setup-code-for-test"
	if _, err := setup.PrepareAuthorization(t.Context(), code); err != nil {
		t.Fatal(err)
	}
	router := NewRouter(Config{SetupService: setup})
	status := httptest.NewRecorder()
	router.ServeHTTP(status, httptest.NewRequest(http.MethodGet, "/api/v1/setup/status", nil))
	var state map[string]any
	if err := json.Unmarshal(status.Body.Bytes(), &state); err != nil || len(state) != 2 || state["setupRequired"] != true || state["cookieSecure"] != false {
		t.Fatalf("unexpected public setup status: %v", err)
	}
	for _, attempt := range []string{"", code} {
		body, err := json.Marshal(map[string]string{"username": "admin", "email": "admin@example.test",
			"password": "very-secure-password", "setupCode": attempt})
		if err != nil {
			t.Fatal(err)
		}
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/v1/setup/admin", strings.NewReader(string(body))))
		want := http.StatusForbidden
		if attempt == code {
			want = http.StatusCreated
		}
		if response.Code != want || strings.Contains(response.Body.String(), code) {
			t.Fatalf("setup status=%d want=%d or authorization leaked", response.Code, want)
		}
	}
}

func TestConfiguredCookieSecurityControlsActualSessionCookie(t *testing.T) {
	for _, test := range []struct {
		mode, url string
		secure    bool
	}{
		{"auto", "", false}, {"auto", "http://127.0.0.1:8080", false},
		{"auto", "https://railkeeper.example.test", true},
		{"true", "http://127.0.0.1:8080", true}, {"false", "https://railkeeper.example.test", false},
	} {
		t.Run(test.mode+"/"+test.url, func(t *testing.T) {
			db := testRouterDB(t)
			if err := application.NewSetupService(db).CreateAdmin(t.Context(), application.CreateAdminInput{
				Username: "admin", Email: "admin@example.test", Password: "very-secure-password",
			}); err != nil {
				t.Fatal(err)
			}
			secure, err := startup.ResolveCookieSecurity(test.mode, test.url)
			if err != nil {
				t.Fatal(err)
			}
			router := NewRouter(Config{AuthService: application.NewAuthService(db), CookieSecure: secure})
			request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(
				`{"username":"admin","password":"very-secure-password"}`))
			request.Header.Set("X-Forwarded-Proto", "https")
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			if response.Code != http.StatusOK {
				t.Fatalf("login status=%d", response.Code)
			}
			cookies := response.Result().Cookies()
			if len(cookies) != 3 {
				t.Fatal("login must set session, CSRF and browser-proof cookies")
			}
			for _, cookie := range cookies {
				if (cookie.Name != "rk_session" && cookie.Name != "rk_csrf" && cookie.Name != loginOriginCookieName("admin")) || cookie.Secure != test.secure ||
					cookie.HttpOnly != (cookie.Name != "rk_csrf") || cookie.SameSite != http.SameSiteLaxMode {
					t.Fatal("actual cookie attributes do not match the configured policy")
				}
			}
		})
	}
}
