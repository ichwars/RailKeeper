package application

import (
	"context"
	"fmt"
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

	host, port := splitDigitalCenterTestURL(t, server.URL)
	service := NewDigitalCenterService()
	locomotives, metadata, err := service.ReadCS3Locomotives(
		context.Background(),
		DigitalCenterConnectionInput{Host: host, Port: port},
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

	host, port := splitDigitalCenterTestURL(t, server.URL)
	locomotives, metadata, err := NewDigitalCenterService().ReadCS3Locomotives(
		context.Background(),
		DigitalCenterConnectionInput{Host: host, Port: port},
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

			host, port := splitDigitalCenterTestURL(t, server.URL)
			result, err := NewDigitalCenterService().TestCS3Connection(
				context.Background(), DigitalCenterConnectionInput{Host: host, Port: port},
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
		name string
		body string
	}{
		{name: "duplicate uid", body: `[
			{"name":"A","uid":"0x10","address":1,"dectyp":"mm"},
			{"name":"B","uid":"0x10","address":2,"dectyp":"dcc"}
		]`},
		{name: "invalid uid", body: `[{"name":"A","uid":"javascript:1","address":1,"dectyp":"mm"}]`},
		{name: "invalid address", body: `[{"name":"A","uid":"0x10","address":70000,"dectyp":"mm"}]`},
		{name: "invalid item", body: `["not an object"]`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := newCS3JSONServer(t, test.body)
			defer server.Close()
			host, port := splitDigitalCenterTestURL(t, server.URL)
			_, _, err := NewDigitalCenterService().ReadCS3Locomotives(
				context.Background(), DigitalCenterConnectionInput{Host: host, Port: port},
			)
			if err == nil || !strings.Contains(err.Error(), "CS3") {
				t.Fatalf("error = %v, want bounded CS3 validation error", err)
			}
		})
	}
}

func TestDigitalCenterServiceReadCS3LocomotivesRejectsRedirectsAndOversizedBodies(t *testing.T) {
	target := newCS3JSONServer(t, `[]`)
	defer target.Close()
	redirect := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL+r.URL.Path, http.StatusFound)
	}))
	defer redirect.Close()
	host, port := splitDigitalCenterTestURL(t, redirect.URL)
	_, _, err := NewDigitalCenterService().ReadCS3Locomotives(
		context.Background(), DigitalCenterConnectionInput{Host: host, Port: port},
	)
	if err == nil || !strings.Contains(err.Error(), "Weiterleitung") {
		t.Fatalf("redirect error = %v", err)
	}

	oversized := newCS3JSONServer(t, `[{"name":"`+strings.Repeat("x", maxCS3ResponseBytes)+
		`","uid":"0x10","address":1,"dectyp":"mm"}]`)
	defer oversized.Close()
	host, port = splitDigitalCenterTestURL(t, oversized.URL)
	_, _, err = NewDigitalCenterService().ReadCS3Locomotives(
		context.Background(), DigitalCenterConnectionInput{Host: host, Port: port},
	)
	if err == nil || !strings.Contains(err.Error(), "zu groß") {
		t.Fatalf("oversized error = %v", err)
	}
}

func TestDigitalCenterServiceProbeCS3ConnectionReportsAllowlistedDiagnostics(t *testing.T) {
	server := newCS3JSONServer(t, `[{"name":"BR 218","uid":"0xc003","address":3,"dectyp":"dcc","speed":90}]`)
	defer server.Close()
	host, port := splitDigitalCenterTestURL(t, server.URL)

	result, err := NewDigitalCenterService().ProbeCS3Connection(
		context.Background(), DigitalCenterConnectionInput{Host: host, Port: port},
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
