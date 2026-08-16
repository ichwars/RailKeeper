package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestVersionInfoWithoutConfiguredUpdateSource(t *testing.T) {
	router := NewRouter(Config{Version: "0.1.0"})
	request := httptest.NewRequest(http.MethodGet, "/api/v1/version?check=true", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	var body versionInfoResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Status != "not_configured" {
		t.Fatalf("expected not_configured status, got %q", body.Status)
	}
	if body.UpdateAvailable {
		t.Fatal("expected no update when no update source is configured")
	}
}

func TestVersionInfoDetectsUpdate(t *testing.T) {
	updateServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"tag_name":"v0.2.0","html_url":"https://example.test/releases/v0.2.0","body":"Release notes","assets":[{"name":"railkeeper.zip","browser_download_url":"https://example.test/railkeeper.zip"}]}`))
	}))
	defer updateServer.Close()

	router := NewRouter(Config{Version: "0.1.0", UpdateCheckURL: updateServer.URL})
	request := httptest.NewRequest(http.MethodGet, "/api/v1/version?check=true", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	var body versionInfoResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Status != "update_available" {
		t.Fatalf("expected update_available status, got %q", body.Status)
	}
	if !body.UpdateAvailable {
		t.Fatal("expected update to be detected")
	}
	if body.LatestVersion != "v0.2.0" {
		t.Fatalf("expected latest version v0.2.0, got %q", body.LatestVersion)
	}
	if body.ReleaseNotes != "Release notes" || body.ReleaseURL != "https://example.test/releases/v0.2.0" {
		t.Fatalf("expected release notes and release URL, got %#v", body)
	}
}

func TestSelectTrustedWindowsPackage(t *testing.T) {
	const (
		updateURL = "https://api.github.com/repos/ichwars/RailKeeper/releases/latest"
		assetName = "RailKeeper-windows-x64-v0.2.0.zip"
		assetURL  = "https://github.com/ichwars/RailKeeper/releases/download/v0.2.0/" + assetName
	)

	tests := []struct {
		name         string
		enabled      bool
		updateURL    string
		assets       []updateReleaseAsset
		wantPackage  bool
		wantAssetURL string
	}{
		{
			name:         "exact standalone package",
			enabled:      true,
			updateURL:    updateURL,
			assets:       []updateReleaseAsset{{Name: assetName, BrowserDownloadURL: assetURL}},
			wantPackage:  true,
			wantAssetURL: assetURL,
		},
		{
			name:      "source archive before exact package",
			enabled:   true,
			updateURL: updateURL,
			assets: []updateReleaseAsset{
				{Name: "Source code (tar.gz)", BrowserDownloadURL: "https://github.com/ichwars/RailKeeper/archive/v0.2.0.tar.gz"},
				{Name: assetName, BrowserDownloadURL: assetURL},
			},
			wantPackage:  true,
			wantAssetURL: assetURL,
		},
		{
			name:      "wrong version",
			enabled:   true,
			updateURL: updateURL,
			assets: []updateReleaseAsset{{
				Name:               "RailKeeper-windows-x64-v0.1.9.zip",
				BrowserDownloadURL: "https://github.com/ichwars/RailKeeper/releases/download/v0.2.0/RailKeeper-windows-x64-v0.1.9.zip",
			}},
		},
		{
			name:      "HTTP asset",
			enabled:   true,
			updateURL: updateURL,
			assets:    []updateReleaseAsset{{Name: assetName, BrowserDownloadURL: "http://github.com/ichwars/RailKeeper/releases/download/v0.2.0/" + assetName}},
		},
		{
			name:      "foreign repository",
			enabled:   true,
			updateURL: updateURL,
			assets:    []updateReleaseAsset{{Name: assetName, BrowserDownloadURL: "https://github.com/other/RailKeeper/releases/download/v0.2.0/" + assetName}},
		},
		{
			name:      "lookalike host",
			enabled:   true,
			updateURL: updateURL,
			assets:    []updateReleaseAsset{{Name: assetName, BrowserDownloadURL: "https://github.com.evil.test/ichwars/RailKeeper/releases/download/v0.2.0/" + assetName}},
		},
		{
			name:      "server mode",
			enabled:   false,
			updateURL: updateURL,
			assets:    []updateReleaseAsset{{Name: assetName, BrowserDownloadURL: assetURL}},
		},
		{
			name:      "configured data mode",
			enabled:   false,
			updateURL: updateURL,
			assets:    []updateReleaseAsset{{Name: assetName, BrowserDownloadURL: assetURL}},
		},
		{
			name:      "untrusted update source",
			enabled:   true,
			updateURL: "https://example.test/releases/latest",
			assets:    []updateReleaseAsset{{Name: assetName, BrowserDownloadURL: assetURL}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := selectTrustedWindowsPackage(
				test.enabled,
				test.updateURL,
				"v0.2.0",
				"v0.2.0",
				test.assets,
			)
			if test.wantPackage && got == nil {
				t.Fatal("expected trusted Windows package")
			}
			if !test.wantPackage && got != nil {
				t.Fatalf("expected no Windows package, got %#v", got)
			}
			if got != nil {
				if got.Version != "v0.2.0" || got.Name != assetName || got.URL != test.wantAssetURL {
					t.Fatalf("unexpected Windows package: %#v", got)
				}
			}
		})
	}
}

func TestVersionInfoGatesTrustedWindowsPackageByRuntime(t *testing.T) {
	const (
		updateURL  = "https://api.github.com/repos/ichwars/RailKeeper/releases/latest"
		releaseURL = "https://github.com/ichwars/RailKeeper/releases/tag/v0.2.0"
		assetName  = "RailKeeper-windows-x64-v0.2.0.zip"
		assetURL   = "https://github.com/ichwars/RailKeeper/releases/download/v0.2.0/" + assetName
	)

	previousClient := http.DefaultClient
	http.DefaultClient = &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.URL.String() != updateURL {
			t.Fatalf("unexpected update request %q", request.URL.String())
		}
		body := `{"tag_name":"v0.2.0","html_url":"` + releaseURL +
			`","assets":[{"name":"` + assetName + `","browser_download_url":"` + assetURL + `"}]}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(body)),
			Request:    request,
		}, nil
	})}
	t.Cleanup(func() { http.DefaultClient = previousClient })

	for _, test := range []struct {
		name        string
		download    bool
		wantPackage bool
	}{
		{name: "Windows standalone", download: true, wantPackage: true},
		{name: "configured or server mode", download: false, wantPackage: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			router := NewRouter(Config{
				Version:                   "0.1.0",
				UpdateCheckURL:            updateURL,
				WindowsStandaloneDownload: test.download,
			})
			request := httptest.NewRequest(http.MethodGet, "/api/v1/version?check=true", nil)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			var body versionInfoResponse
			if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if response.Code != http.StatusOK || body.ReleaseURL != releaseURL || !body.UpdateAvailable {
				t.Fatalf("unexpected update response: status=%d body=%#v", response.Code, body)
			}
			if test.wantPackage && (body.WindowsPackage == nil || body.WindowsPackage.URL != assetURL) {
				t.Fatalf("expected trusted Windows package, got %#v", body.WindowsPackage)
			}
			if !test.wantPackage && body.WindowsPackage != nil {
				t.Fatalf("expected release-page fallback only, got %#v", body.WindowsPackage)
			}
		})
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func TestVersionInfoCanIncludePrereleases(t *testing.T) {
	updateServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{"tag_name":"v0.3.0-beta.1","html_url":"https://example.test/releases/v0.3.0-beta.1","prerelease":true},
			{"tag_name":"v0.2.0","html_url":"https://example.test/releases/v0.2.0","prerelease":false}
		]`))
	}))
	defer updateServer.Close()

	router := NewRouter(Config{Version: "0.2.0", UpdateCheckURL: updateServer.URL})
	request := httptest.NewRequest(http.MethodGet, "/api/v1/version?check=true&prerelease=true", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	var body versionInfoResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.LatestVersion != "v0.3.0-beta.1" {
		t.Fatalf("expected prerelease version, got %q", body.LatestVersion)
	}
	if !body.UpdateAvailable {
		t.Fatal("expected prerelease update to be detected")
	}
}

func TestVersionInfoHandlesMissingGithubRelease(t *testing.T) {
	updateServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer updateServer.Close()

	router := NewRouter(Config{Version: "0.1.0", UpdateCheckURL: updateServer.URL})
	request := httptest.NewRequest(http.MethodGet, "/api/v1/version?check=true", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	var body versionInfoResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Status != "no_release" {
		t.Fatalf("expected no_release status, got %q", body.Status)
	}
	if body.Message != "Keine Release-Information verfügbar." {
		t.Fatalf("unexpected message %q", body.Message)
	}
	if body.UpdateAvailable {
		t.Fatal("expected no update when no release exists")
	}
}

func TestVersionInfoHandlesEmptyReleaseList(t *testing.T) {
	updateServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}))
	defer updateServer.Close()

	router := NewRouter(Config{Version: "0.1.0", UpdateCheckURL: updateServer.URL})
	request := httptest.NewRequest(http.MethodGet, "/api/v1/version?check=true&prerelease=true", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	var body versionInfoResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Status != "no_release" || body.Message != "Keine Release-Information verfügbar." {
		t.Fatalf("expected no-release response, got %#v", body)
	}
}

func TestVersionInfoCachesRepeatedExternalChecks(t *testing.T) {
	var requests int
	updateServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"tag_name":"v0.2.0","html_url":"https://example.test/releases/v0.2.0"}`))
	}))
	defer updateServer.Close()

	router := NewRouter(Config{Version: "0.1.0", UpdateCheckURL: updateServer.URL})
	for range 2 {
		request := httptest.NewRequest(http.MethodGet, "/api/v1/version?check=true", nil)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", response.Code)
		}
	}
	if requests != 1 {
		t.Fatalf("expected one external request for repeated checks, got %d", requests)
	}
}

func TestReleaseListURLConvertsGithubLatestEndpoint(t *testing.T) {
	got := releaseListURL("https://api.github.com/repos/ichwars/RailKeeper/releases/latest")
	want := "https://api.github.com/repos/ichwars/RailKeeper/releases"
	if got != want {
		t.Fatalf("releaseListURL() = %q, want %q", got, want)
	}
}

func TestCompareVersionStringsTreatsReleaseNewerThanSamePrerelease(t *testing.T) {
	if compareVersionStrings("v0.3.0-beta.1", "v0.3.0") >= 0 {
		t.Fatal("expected final release to be newer than matching prerelease")
	}
}
