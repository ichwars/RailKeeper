package application

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestDigitalCenterServiceReadCS3LocomotivesPrefersCurrentAPI(t *testing.T) {
	fixture := readCS3Fixture(t, "testdata/cs3_locos_2_6_anonymized.json")
	legacyCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/app/api/locos":
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			_, _ = w.Write(fixture)
		case "/app/api/loks":
			legacyCalls++
			http.NotFound(w, r)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	service, input := newCS3TestService(t, server)
	locomotives, metadata, err := service.ReadCS3Locomotives(
		context.Background(),
		input,
	)
	if err != nil {
		t.Fatal(err)
	}
	if legacyCalls != 0 {
		t.Fatalf("legacy endpoint calls = %d, want 0", legacyCalls)
	}
	if len(locomotives) != 1 || locomotives[0].ObjectID != 0x10 ||
		locomotives[0].Name != "Neue MM-Lok" || locomotives[0].Address != 16 ||
		locomotives[0].Protocol != "MOTOROLA" {
		t.Fatalf("locomotives = %#v", locomotives)
	}
	if metadata.APIPath != "/app/api/locos" || metadata.APIGeneration != "2.6+" ||
		metadata.LocomotiveCount != 1 {
		t.Fatalf("metadata = %#v", metadata)
	}
}

func TestDigitalCenterServiceReadCS3LocomotivesFallsBackToLegacyAPI(t *testing.T) {
	fixture := readCS3Fixture(t, "testdata/cs3_loks_pre_2_6_anonymized.json")
	paths := []string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		if r.URL.Path == "/app/api/locos" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":"Not Found"}`))
			return
		}
		if r.URL.Path == "/app/api/loks" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(fixture)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	service, input := newCS3TestService(t, server)
	locomotives, metadata, err := service.ReadCS3Locomotives(
		context.Background(),
		input,
	)
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(paths) != "[/app/api/locos /app/api/loks]" {
		t.Fatalf("paths = %v", paths)
	}
	if len(locomotives) != 2 || locomotives[0].Protocol != "MFX" || locomotives[1].Protocol != "DCC" {
		t.Fatalf("locomotives = %#v", locomotives)
	}
	if metadata.APIPath != "/app/api/loks" || metadata.APIGeneration != "pre-2.6" {
		t.Fatalf("metadata = %#v", metadata)
	}
}

func TestDigitalCenterServiceTestCS3ConnectionRequiresCompatibleRoster(t *testing.T) {
	tests := []struct {
		name            string
		contentType     string
		status          int
		body            string
		wantErrorKind   string
		wantMessagePart string
	}{
		{name: "web page", contentType: "text/html", status: http.StatusOK,
			body: "<html>not a CS3 API</html>", wantErrorKind: "content_type", wantMessagePart: "JSON"},
		{name: "authentication", contentType: "application/json", status: http.StatusUnauthorized,
			body: `{"error":"unauthorized"}`, wantErrorKind: "authentication", wantMessagePart: "Authentifizierung"},
		{name: "invalid shape", contentType: "application/json", status: http.StatusOK,
			body: `{"name":"website"}`, wantErrorKind: "format", wantMessagePart: "JSON"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", test.contentType)
				w.WriteHeader(test.status)
				_, _ = w.Write([]byte(test.body))
			}))
			defer server.Close()

			service, input := newCS3TestService(t, server)
			result, err := service.TestCS3Connection(
				context.Background(), input,
			)
			if err != nil {
				t.Fatal(err)
			}
			if result.Connected || result.Fields["errorKind"] != test.wantErrorKind ||
				!strings.Contains(result.Message, test.wantMessagePart) {
				t.Fatalf("result = %#v", result)
			}
		})
	}
}

func TestDigitalCenterServiceReadCS3LocomotivesRejectsUnsafeResponses(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		wantKind cs3ErrorKind
	}{
		{name: "duplicate uid", body: `[
			{"name":"A","uid":"0x10","address":1,"dectyp":"mm"},
			{"name":"B","uid":"0x10","address":2,"dectyp":"dcc"}
		]`, wantKind: cs3ErrorDeviceOutput},
		{name: "invalid uid", body: `[{"name":"A","uid":"javascript:1","address":1,"dectyp":"mm"}]`,
			wantKind: cs3ErrorDeviceOutput},
		{name: "invalid address", body: `[{"name":"A","uid":"0x10","address":70000,"dectyp":"mm"}]`,
			wantKind: cs3ErrorDeviceOutput},
		{name: "invalid item", body: `["not an object"]`, wantKind: cs3ErrorFormat},
		{name: "null roster", body: `null`, wantKind: cs3ErrorFormat},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := newCS3JSONServer(t, test.body)
			defer server.Close()
			service, input := newCS3TestService(t, server)
			_, _, err := service.ReadCS3Locomotives(
				context.Background(), input,
			)
			if err == nil || !strings.Contains(err.Error(), "CS3") || cs3ErrorKindOf(err) != test.wantKind {
				t.Fatalf("error = %v, kind = %q, want bounded %q CS3 validation error",
					err, cs3ErrorKindOf(err), test.wantKind)
			}
		})
	}
}

func TestDigitalCenterServiceReadCS3LocomotivesRejectsMalformedUTF8(t *testing.T) {
	body := []byte(`[{"name":"BR `)
	body = append(body, 0xff)
	body = append(body, []byte(`","uid":"0x10","address":1,"dectyp":"dcc"}]`)...)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/app/api/locos" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	defer server.Close()

	service, input := newCS3TestService(t, server)
	_, _, err := service.ReadCS3Locomotives(
		context.Background(), input,
	)
	if err == nil || cs3ErrorKindOf(err) != cs3ErrorDeviceOutput {
		t.Fatalf("error = %v, want malformed UTF-8 device output error", err)
	}
}

func TestDigitalCenterServiceReadCS3LocomotivesRejectsLoopbackTargetBeforeRequest(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	host, port := splitDigitalCenterTestURL(t, server.URL)
	_, _, err := NewDigitalCenterService().ReadCS3Locomotives(
		context.Background(), DigitalCenterConnectionInput{Host: host, Port: port},
	)
	if err == nil || !strings.Contains(err.Error(), "private IP-Adresse") {
		t.Fatalf("error = %v, want unsafe target rejection", err)
	}
	if requests != 0 {
		t.Fatalf("requests = %d, want 0", requests)
	}
}

func TestDigitalCenterServiceReadCS3LocomotivesRejectsUnsafeResolvedTargets(t *testing.T) {
	tests := []struct {
		name      string
		host      string
		addresses []net.IPAddr
	}{
		{name: "public literal", host: "8.8.8.8"},
		{name: "link local literal", host: "169.254.169.254"},
		{name: "mixed DNS response", host: "cs3.example", addresses: []net.IPAddr{
			{IP: net.ParseIP("192.168.10.23")},
			{IP: net.ParseIP("203.0.113.10")},
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := NewDigitalCenterService()
			service.cs3Resolver = &cs3ResolverStub{addresses: test.addresses}
			dialed := false
			service.cs3DialContext = func(context.Context, string, string) (net.Conn, error) {
				dialed = true
				return nil, fmt.Errorf("unexpected dial")
			}

			_, _, err := service.ReadCS3Locomotives(
				context.Background(), DigitalCenterConnectionInput{Host: test.host, Port: 80},
			)
			if err == nil || cs3ErrorKindOf(err) != cs3ErrorUnsafeTarget || dialed {
				t.Fatalf("error = %v, dialed = %t, want unsafe target rejection", err, dialed)
			}
		})
	}
}

func TestDigitalCenterServicePinsResolvedPrivateTargetAcrossFallback(t *testing.T) {
	paths := []string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/app/api/locos" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`[{"uid":"0x2a","name":"BR 218","address":3,"dectyp":"mfx"}]`))
	}))
	defer server.Close()

	resolver := &cs3ResolverStub{addresses: []net.IPAddr{{IP: net.ParseIP("192.168.10.23")}}}
	service := NewDigitalCenterService()
	service.cs3Resolver = resolver
	dialer := &net.Dialer{}
	service.cs3DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		if address != "192.168.10.23:80" {
			t.Fatalf("dial address = %q, want pinned private target", address)
		}
		return dialer.DialContext(ctx, network, server.Listener.Addr().String())
	}

	locomotives, metadata, err := service.ReadCS3Locomotives(
		context.Background(), DigitalCenterConnectionInput{Host: "cs3.local", Port: 80},
	)
	if err != nil {
		t.Fatal(err)
	}
	if resolver.calls != 1 || fmt.Sprint(paths) != "[/app/api/locos /app/api/loks]" ||
		len(locomotives) != 1 || metadata.APIPath != "/app/api/loks" {
		t.Fatalf("resolver calls = %d, paths = %v, locomotives = %#v, metadata = %#v",
			resolver.calls, paths, locomotives, metadata)
	}
}

func TestDigitalCenterServiceReadCS3LocomotivesRejectsRedirectsAndOversizedBodies(t *testing.T) {
	target := newCS3JSONServer(t, `[]`)
	defer target.Close()
	redirect := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL+r.URL.Path, http.StatusFound)
	}))
	defer redirect.Close()
	service, input := newCS3TestService(t, redirect)
	_, _, err := service.ReadCS3Locomotives(
		context.Background(), input,
	)
	if err == nil || !strings.Contains(err.Error(), "Weiterleitung") {
		t.Fatalf("redirect error = %v", err)
	}

	oversized := newCS3JSONServer(t, `[{"name":"`+strings.Repeat("x", maxCS3ResponseBytes)+
		`","uid":"0x10","address":1,"dectyp":"mm"}]`)
	defer oversized.Close()
	service, input = newCS3TestService(t, oversized)
	_, _, err = service.ReadCS3Locomotives(
		context.Background(), input,
	)
	if err == nil || !strings.Contains(err.Error(), "zu groß") {
		t.Fatalf("oversized error = %v", err)
	}
}

func TestDigitalCenterServiceProbeCS3ConnectionReportsAllowlistedDiagnostics(t *testing.T) {
	server := newCS3JSONServer(t, `[{"name":"BR 218","uid":"0xc003","address":3,"dectyp":"dcc","speed":90}]`)
	defer server.Close()
	service, input := newCS3TestService(t, server)

	result, err := service.ProbeCS3Connection(
		context.Background(), input,
	)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Connected || len(result.Commands) != 1 || !result.Commands[0].OK ||
		result.Commands[0].Request != "GET /app/api/locos" {
		t.Fatalf("result = %#v", result)
	}
	if result.Fields["apiGeneration"] != "2.6+" || result.Fields["locomotiveCount"] != "1" ||
		result.Fields["readOnly"] != "true" {
		t.Fatalf("fields = %#v", result.Fields)
	}
	for _, forbidden := range []string{"speed", "direction", "functions"} {
		if _, found := result.Fields[forbidden]; found {
			t.Fatalf("diagnostics leaked %q: %#v", forbidden, result.Fields)
		}
	}
}

func readCS3Fixture(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func newCS3JSONServer(t *testing.T, body string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/app/api/locos" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
}

func newCS3TestService(
	t *testing.T,
	server *httptest.Server,
) (*DigitalCenterService, DigitalCenterConnectionInput) {
	t.Helper()
	service := NewDigitalCenterService()
	dialer := &net.Dialer{}
	serverAddress := server.Listener.Addr().String()
	service.cs3DialContext = func(ctx context.Context, network, _ string) (net.Conn, error) {
		return dialer.DialContext(ctx, network, serverAddress)
	}
	return service, DigitalCenterConnectionInput{Host: "192.168.10.23", Port: 80}
}

type cs3ResolverStub struct {
	addresses []net.IPAddr
	err       error
	calls     int
}

func (resolver *cs3ResolverStub) LookupIPAddr(context.Context, string) ([]net.IPAddr, error) {
	resolver.calls++
	return resolver.addresses, resolver.err
}
