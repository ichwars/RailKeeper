package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestBrowserProofCookieProtectsBudgetButDoesNotAuthenticate(t *testing.T) {
	db := testRouterDB(t)
	if err := application.NewSetupService(db).CreateAdmin(t.Context(), application.CreateAdminInput{Username: "admin", Email: "admin@example.test", Password: "very-secure-password"}); err != nil {
		t.Fatal(err)
	}
	router := NewRouter(Config{AuthService: application.NewAuthService(db), CookieSecure: true})
	login := func(index int, password string, cookie *http.Cookie) *httptest.ResponseRecorder {
		request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(fmt.Sprintf(`{"username":"admin","password":%q}`, password)))
		request.Header.Set("Content-Type", "application/json")
		request.RemoteAddr = fmt.Sprintf("192.0.2.%d:1234", index)
		if cookie != nil {
			request.AddCookie(cookie)
		}
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		return response
	}
	first := login(1, "very-secure-password", nil)
	if first.Code != http.StatusOK {
		t.Fatal(first.Code, first.Body.String())
	}
	var proof *http.Cookie
	for _, cookie := range first.Result().Cookies() {
		if cookie.Name == loginOriginCookieName("admin") {
			proof = cookie
		}
	}
	if proof == nil || !proof.HttpOnly || !proof.Secure || proof.SameSite != http.SameSiteLaxMode || proof.MaxAge <= 0 || proof.Path != "/" {
		t.Fatal("unsafe browser proof cookie", proof)
	}
	if strings.Contains(first.Body.String(), proof.Value) || strings.Contains(first.Body.String(), "origin") {
		t.Fatal("origin proof exposed to JavaScript response")
	}
	for index := range 10 {
		if response := login(index+2, "wrong", nil); response.Code != http.StatusUnauthorized {
			t.Fatal(response.Code)
		}
	}
	if response := login(20, "very-secure-password", nil); response.Code != http.StatusTooManyRequests {
		t.Fatal("new origin bypassed budget", response.Code)
	}
	if response := login(21, "very-secure-password", proof); response.Code != http.StatusOK {
		t.Fatal("recognized origin locked out", response.Code, response.Body.String())
	}
	if response := login(22, "wrong", proof); response.Code != http.StatusUnauthorized || len(response.Result().Cookies()) != 0 {
		t.Fatal("proof alone authenticated")
	}
	// Same NAT IP is not a recognition signal. An invalid token also cannot use
	// the protected budget even when it is the correct length.
	forged := *proof
	forged.Value = strings.Repeat("a", 43)
	if response := login(21, "very-secure-password", &forged); response.Code != http.StatusTooManyRequests {
		t.Fatal("NAT or forged proof granted trust", response.Code)
	}
}

func TestLoginRecoveryRouteRequiresAdminAndCSRF(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	if err := application.NewSetupService(db).CreateAdmin(t.Context(), application.CreateAdminInput{Username: "admin", Email: "admin@example.test", Password: "very-secure-password"}); err != nil {
		t.Fatal(err)
	}
	var id string
	if err := db.QueryRow(`SELECT id FROM users`).Scan(&id); err != nil {
		t.Fatal(err)
	}
	session := loginRouteTestUser(t, auth, "admin", "very-secure-password")
	router := NewRouter(Config{AuthService: auth})
	for _, csrf := range []string{"", "incorrect", session.CSRFToken} {
		request := httptest.NewRequest(http.MethodPost, "/api/v1/users/"+id+"/login-budget/reset", nil)
		request.AddCookie(&http.Cookie{Name: "rk_session", Value: session.SessionToken})
		request.Header.Set("X-CSRF-Token", csrf)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		want := http.StatusForbidden
		if csrf == session.CSRFToken {
			want = http.StatusNoContent
		}
		if response.Code != want {
			t.Fatal("recovery protection", response.Code, want, response.Body.String())
		}
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM audit_logs WHERE action='LoginBudgetReset'`).Scan(&count); err != nil || count != 1 {
		t.Fatal("unauthorized recovery was audited as successful", count, err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/users/missing/login-budget/reset", nil)
	request.AddCookie(&http.Cookie{Name: "rk_session", Value: session.SessionToken})
	request.Header.Set("X-CSRF-Token", session.CSRFToken)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Fatal("missing target", response.Code)
	}
}
