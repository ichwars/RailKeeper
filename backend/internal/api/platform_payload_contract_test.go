package api

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPrinterPayloadContractsRetainNativeNamesAndDefaults(t *testing.T) {
	for _, input := range []struct {
		body            string
		count           int
		status, printer string
	}{
		{`[{"Name":"Workshop printer","Default":true},{"Name":"Archive","Default":false},{"Name":" "}]`, 2, "available", "Workshop printer"},
		{`{"Name":"Single printer","Default":true}`, 1, "available", "Single printer"},
		{`[]`, 0, "available", ""}, {`invalid`, 0, "unavailable", ""},
	} {
		result := parseWindowsPrinters([]byte(input.body))
		if len(result.Printers) != input.count || result.Status != input.status || result.DefaultPrinter != input.printer {
			t.Fatalf("native printer payload changed: %#v", result)
		}
	}
	for _, input := range []struct{ value, want string }{{"system default destination: Workshop", "Workshop"}, {"", ""}, {"unrecognized", ""}} {
		if got := parseLPStatDefault(input.value); got != input.want {
			t.Fatalf("default printer %q != %q", got, input.want)
		}
	}
	names := printersFromNames([]string{" A ", "", "A", "B"}, "b")
	if len(names) != 2 || !names[1].IsDefault || names[0].ID != "a" || printerID("###") != "printer" {
		t.Fatalf("native printer names lost: %#v", names)
	}
	for _, c := range []struct{ path, key string }{{"railkeeper.db-wal", "database"}, {"uploads/images/a.png", "images"}, {"uploads/thumbs/a.jpg", "thumbnails"}, {"uploads/cv/a.esux", "decoderFiles"}, {"uploads/a.pdf", "attachments"}, {"config.txt", "other"}} {
		if got := storageCategoryKey(c.path); got != c.key {
			t.Fatalf("storage category for %s: %s", c.path, got)
		}
	}
}

func TestAuthenticationHTTPFailuresRemainRecoverableWithoutAcknowledgingWrites(t *testing.T) {
	f := newContractFixture(t)
	path := "/api/v1/auth/password"
	assertProblem(t, f.raw(t, "PUT", path, "{"), 400, "invalid_json")
	assertProblem(t, f.call(t, "PUT", path, map[string]string{"currentPassword": "wrong", "newPassword": "long-enough-synthetic-password"}), 401, "invalid_password")
	assertProblem(t, f.call(t, "PUT", path, map[string]string{"currentPassword": "synthetic-contract-password", "newPassword": "short"}), 400, "weak_password")
	// Missing accounts in a handler's actor context cannot disclose 2FA information.
	app := &App{authService: f.auth, logger: f.config.Logger}
	for _, handler := range []routeHandler{(*App).twoFactorStatus, (*App).setupTwoFactor, (*App).enableTwoFactor, (*App).disableTwoFactor} {
		r := withActorUserID(httptest.NewRequest("POST", "/contract", strings.NewReader(`{"code":"123456"}`)), "missing-account")
		w := httptest.NewRecorder()
		handler(app, w, r)
		assertProblem(t, w, 401, "unauthorized")
	}
	if err := f.db.Close(); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		method, path string
		body         any
		code         string
	}{
		{"POST", "/api/v1/auth/login", map[string]string{"username": "Admin", "password": "synthetic-contract-password"}, "login_failed"},
		{"GET", "/api/v1/auth/session", nil, "session_failed"},
		{"POST", "/api/v1/auth/password-reset", map[string]string{"email": "Admin@example.test"}, "password_reset_failed"},
		{"POST", "/api/v1/auth/password-reset/confirm", map[string]string{"token": strings.Repeat("a", 64), "newPassword": "long-enough-synthetic-password"}, "password_reset_failed"},
	} {
		assertProblem(t, f.call(t, c.method, c.path, c.body), 500, c.code)
	}
}

func TestRemoteAccessoryImagePayloadContractsAreBoundedAndKeepBytes(t *testing.T) {
	png := testPNG(t)
	for _, c := range []struct {
		status      int
		data        []byte
		length, max int64
		want        error
	}{
		{200, png, int64(len(png)), int64(len(png)), nil},
		{200, png, int64(len(png)), 1, errRemoteAccessoryImageTooLarge},
		{200, png, -1, 1, errRemoteAccessoryImageTooLarge},
		{200, []byte("text"), -1, 1024, errRemoteAccessoryImageTypeUnsupported},
	} {
		client := &http.Client{Transport: contractRoundTripper(func(r *http.Request) (*http.Response, error) {
			response := mediaResponse(c.status, "application/octet-stream", c.data)
			response.ContentLength = c.length
			return response, nil
		})}
		data, mime, err := downloadRemoteAccessoryImage(t.Context(), client, "https://example.test/a.png", c.max)
		if c.want == nil {
			if err != nil || !bytes.Equal(data, png) || mime != "image/png" {
				t.Fatalf("remote image bytes changed: %v", err)
			}
		} else if !errors.Is(err, c.want) {
			t.Fatalf("expected bounded rejection %v, got %v", c.want, err)
		}
	}
	for _, bad := range []string{"", "file:///private", "http://", "not a URL"} {
		_, _, err := downloadRemoteAccessoryImage(t.Context(), nil, bad, 1024)
		if !errors.Is(err, errRemoteAccessoryImageURLInvalid) {
			t.Fatalf("unsafe URL accepted: %s", bad)
		}
	}
	for _, body := range [][]byte{nil, []byte("missing")} {
		client := &http.Client{Transport: contractRoundTripper(func(*http.Request) (*http.Response, error) {
			return mediaResponse(map[bool]int{true: 200, false: 404}[len(body) == 0], "text/plain", body), nil
		})}
		if _, _, err := downloadRemoteAccessoryImage(context.Background(), client, "https://example.test/a.png", 1024); err == nil {
			t.Fatal("empty/error image accepted")
		}
	}
	broken := &http.Client{Transport: contractRoundTripper(func(*http.Request) (*http.Response, error) { return nil, io.ErrUnexpectedEOF })}
	if _, _, err := downloadRemoteAccessoryImage(t.Context(), broken, "https://example.test/a.png", 1024); err == nil {
		t.Fatal("transport failure ignored")
	}
}

type contractRoundTripper func(*http.Request) (*http.Response, error)

func (transport contractRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	return transport(request)
}
