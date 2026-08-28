package api

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestCS3ProbeRouteReturnsReadOnlyHTTPDiagnostics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet || request.URL.Path != "/app/api/locos" {
			t.Errorf("CS3 request = %s %s, want read-only roster GET", request.Method, request.URL.Path)
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`[{"uid":"0x2a","name":"BR 218","address":3,"dectyp":"mfx+"}]`))
	}))
	defer server.Close()
	dialer := &net.Dialer{}
	cs3Service := application.NewDigitalCenterService(application.WithCS3DialContext(
		func(ctx context.Context, network, _ string) (net.Conn, error) {
			return dialer.DialContext(ctx, network, server.Listener.Addr().String())
		},
	))

	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{
		Username: "cs3-admin", Password: "admin-password", Roles: []string{"Admin"},
	}); err != nil {
		t.Fatal(err)
	}
	router := NewRouter(Config{
		AuthService:          auth,
		DigitalCenterService: cs3Service,
	})
	admin := loginRouteTestUser(t, auth, "cs3-admin", "admin-password")

	response := layoutRequest(t, router, admin, http.MethodPost,
		"/api/v1/digital-centers/cs3/probe",
		map[string]any{"host": "192.168.10.23", "port": 80}, true)
	assertStatus(t, response, http.StatusOK)
	var result application.DigitalCenterProbeResult
	decodeResponse(t, response, &result)
	if !result.Connected || result.Provider != "cs3" || len(result.Commands) != 1 {
		t.Fatalf("CS3 probe result = %#v", result)
	}
	command := result.Commands[0]
	if !command.OK || command.Request != "GET /app/api/locos" || command.CommandHex != "" ||
		command.Fields["readOnly"] != "true" || command.Fields["locomotiveCount"] != "1" {
		t.Fatalf("CS3 probe command = %#v, want bounded read-only HTTP diagnostics", command)
	}

	withoutCSRF := layoutRequest(t, router, admin, http.MethodPost,
		"/api/v1/digital-centers/cs3/probe",
		map[string]any{"host": "192.168.10.23", "port": 80}, false)
	assertProblem(t, withoutCSRF, http.StatusForbidden, "csrf_required")
}
