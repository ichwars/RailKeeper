package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
)

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
