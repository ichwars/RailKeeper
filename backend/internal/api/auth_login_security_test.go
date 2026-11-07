package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"railkeeper/backend/internal/application"
)

func TestLoginRejectsConsumedTOTPWithoutCreatingAnotherSession(t *testing.T) {
	db := testRouterDB(t)
	if err := application.NewSetupService(db).CreateAdmin(t.Context(), application.CreateAdminInput{
		Username: "admin", Email: "admin@example.test", Password: "very-secure-password",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(t.Context(), `UPDATE users SET two_factor_enabled=1,
two_factor_secret='GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ'`); err != nil {
		t.Fatal(err)
	}
	auth := application.NewAuthService(db, application.WithAuthClock(func() time.Time { return time.Unix(59, 0) }))
	router := NewRouter(Config{AuthService: auth})
	for _, want := range []int{http.StatusOK, http.StatusUnauthorized} {
		request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(
			`{"username":"admin","password":"very-secure-password","twoFactorCode":"287082"}`))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != want {
			t.Fatalf("login status=%d want=%d", response.Code, want)
		}
		if want == http.StatusUnauthorized && len(response.Result().Cookies()) != 0 {
			t.Fatal("replayed TOTP received a session cookie")
		}
	}
	var sessions int
	if err := db.QueryRowContext(t.Context(), `SELECT COUNT(*) FROM sessions`).Scan(&sessions); err != nil || sessions != 1 {
		t.Fatalf("replayed TOTP created another session: count=%d err=%v", sessions, err)
	}
}

func TestLoginAccountLimitCannotBeBypassedByDifferentIPs(t *testing.T) {
	db := testRouterDB(t)
	router := NewRouter(Config{AuthService: application.NewAuthService(db)})
	for i := range 11 {
		request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login",
			strings.NewReader(`{"username":"missing","password":"incorrect"}`))
		request.RemoteAddr = fmt.Sprintf("192.0.2.%d:12345", i+1)
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		want := http.StatusUnauthorized
		if i == 10 {
			want = http.StatusTooManyRequests
			if response.Header().Get("Retry-After") == "" {
				t.Fatal("blocked login lacks retry guidance")
			}
		}
		if response.Code != want {
			t.Fatalf("request %d: status=%d want=%d", i, response.Code, want)
		}
	}
}
