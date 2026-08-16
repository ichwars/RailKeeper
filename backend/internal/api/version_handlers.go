package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type versionInfoResponse struct {
	Version         string                  `json:"version"`
	LatestVersion   string                  `json:"latestVersion,omitempty"`
	UpdateAvailable bool                    `json:"updateAvailable"`
	SourceURL       string                  `json:"sourceUrl,omitempty"`
	ReleaseURL      string                  `json:"releaseUrl,omitempty"`
	ReleaseNotes    string                  `json:"releaseNotes,omitempty"`
	WindowsPackage  *windowsPackageResponse `json:"windowsPackage,omitempty"`
	CheckedAt       string                  `json:"checkedAt"`
	Status          string                  `json:"status"`
	Message         string                  `json:"message"`
}

type windowsPackageResponse struct {
	Version string `json:"version"`
	Name    string `json:"name"`
	URL     string `json:"url"`
}

type updateReleaseResponse struct {
	Version    string               `json:"version"`
	TagName    string               `json:"tag_name"`
	Name       string               `json:"name"`
	Body       string               `json:"body"`
	HTMLURL    string               `json:"html_url"`
	Assets     []updateReleaseAsset `json:"assets"`
	Prerelease bool                 `json:"prerelease"`
	Draft      bool                 `json:"draft"`
}

type updateReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

var errNoUpdateRelease = errors.New("no update release available")

const updateCheckCacheDuration = 5 * time.Minute

type cachedUpdateCheck struct {
	release   *updateReleaseResponse
	err       error
	checkedAt time.Time
}

func (a *App) versionInfo(w http.ResponseWriter, r *http.Request) {
	response := versionInfoResponse{
		Version:   a.version,
		CheckedAt: time.Now().UTC().Format(time.RFC3339),
		Status:    "local",
		Message:   "Lokale RailKeeper-Version gelesen.",
	}

	if r.URL.Query().Get("check") != "true" {
		respondJSON(w, http.StatusOK, response)
		return
	}

	updateURL := strings.TrimSpace(a.updateCheckURL)
	if updateURL == "" {
		response.Status = "not_configured"
		response.Message = "Keine externe Updatequelle konfiguriert."
		respondJSON(w, http.StatusOK, response)
		return
	}
	if !isAllowedUpdateURL(updateURL) {
		response.Status = "unavailable"
		response.SourceURL = updateURL
		response.Message = "Updatequelle ist nicht erlaubt. Bitte eine HTTP- oder HTTPS-URL konfigurieren."
		respondJSON(w, http.StatusOK, response)
		return
	}

	includePrerelease := r.URL.Query().Get("prerelease") == "true"
	release, err := a.fetchCachedUpdateRelease(r.Context(), updateURL, includePrerelease)
	response.SourceURL = updateURL
	if err != nil {
		if errors.Is(err, errNoUpdateRelease) {
			response.Status = "no_release"
			response.Message = "Keine Release-Information verfügbar."
			respondJSON(w, http.StatusOK, response)
			return
		}
		a.logger.Warn("update check failed", "url", updateURL, "error", err)
		response.Status = "unavailable"
		response.Message = "Updatequelle konnte nicht erreicht werden."
		respondJSON(w, http.StatusOK, response)
		return
	}

	response.LatestVersion = firstUpdateVersion(release.Version, release.TagName, release.Name)
	response.ReleaseURL = release.HTMLURL
	response.ReleaseNotes = strings.TrimSpace(release.Body)
	if response.LatestVersion == "" {
		response.Status = "unavailable"
		response.Message = "Updatequelle enthielt keine auswertbare Version."
		respondJSON(w, http.StatusOK, response)
		return
	}

	if compareVersionStrings(response.LatestVersion, a.version) > 0 {
		response.UpdateAvailable = true
		response.Status = "update_available"
		response.Message = "Eine neuere RailKeeper-Version ist verfügbar."
		response.WindowsPackage = selectTrustedWindowsPackage(
			a.windowsStandaloneDownload,
			updateURL,
			response.LatestVersion,
			release.TagName,
			release.Assets,
		)
	} else {
		response.Status = "current"
		response.Message = "RailKeeper ist aktuell."
	}
	respondJSON(w, http.StatusOK, response)
}

func (a *App) fetchCachedUpdateRelease(
	ctx context.Context,
	updateURL string,
	includePrerelease bool,
) (*updateReleaseResponse, error) {
	a.versionCheckMu.Lock()
	defer a.versionCheckMu.Unlock()

	if cached, ok := a.versionCheckCache[includePrerelease]; ok && time.Since(cached.checkedAt) < updateCheckCacheDuration {
		return cached.release, cached.err
	}
	release, err := fetchUpdateRelease(ctx, updateURL, includePrerelease)
	a.versionCheckCache[includePrerelease] = cachedUpdateCheck{
		release:   release,
		err:       err,
		checkedAt: time.Now(),
	}
	return release, err
}

func isAllowedUpdateURL(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return parsed.Scheme == "https" || parsed.Scheme == "http"
}

func fetchUpdateRelease(ctx context.Context, updateURL string, includePrerelease bool) (*updateReleaseResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if includePrerelease {
		updateURL = releaseListURL(updateURL)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, updateURL, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "RailKeeper")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer func() { _ = response.Body.Close() }()

	if response.StatusCode == http.StatusNotFound {
		return nil, errNoUpdateRelease
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected update status %d", response.StatusCode)
	}

	data, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if release := decodeReleaseList(data, includePrerelease); release != nil {
		return release, nil
	}
	if isReleaseList(data) {
		return nil, errNoUpdateRelease
	}

	var release updateReleaseResponse
	if err := json.Unmarshal(data, &release); err != nil {
		return nil, err
	}
	return &release, nil
}

func releaseListURL(updateURL string) string {
	if strings.HasSuffix(updateURL, "/releases/latest") {
		return strings.TrimSuffix(updateURL, "/latest")
	}
	return updateURL
}

func decodeReleaseList(data []byte, includePrerelease bool) *updateReleaseResponse {
	var releases []updateReleaseResponse
	if err := json.Unmarshal(data, &releases); err != nil {
		return nil
	}
	for index := range releases {
		release := releases[index]
		if release.Draft || (!includePrerelease && release.Prerelease) {
			continue
		}
		if firstUpdateVersion(release.Version, release.TagName, release.Name) == "" {
			continue
		}
		return &release
	}
	return nil
}

func isReleaseList(data []byte) bool {
	var releases []updateReleaseResponse
	return json.Unmarshal(data, &releases) == nil
}

func firstUpdateVersion(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func selectTrustedWindowsPackage(
	enabled bool,
	updateURL string,
	version string,
	tagName string,
	assets []updateReleaseAsset,
) *windowsPackageResponse {
	if !enabled || !isTrustedRailKeeperReleaseAPI(updateURL) {
		return nil
	}

	normalizedVersion := strings.TrimPrefix(strings.TrimSpace(version), "v")
	tagName = strings.TrimSpace(tagName)
	if normalizedVersion == "" || tagName == "" {
		return nil
	}
	expectedName := fmt.Sprintf("RailKeeper-windows-x64-v%s.zip", normalizedVersion)
	for _, asset := range assets {
		if strings.TrimSpace(asset.Name) != expectedName {
			continue
		}
		assetURL := strings.TrimSpace(asset.BrowserDownloadURL)
		if !isTrustedRailKeeperAssetURL(assetURL, tagName, expectedName) {
			continue
		}
		return &windowsPackageResponse{
			Version: strings.TrimSpace(version),
			Name:    expectedName,
			URL:     assetURL,
		}
	}
	return nil
}

func isTrustedRailKeeperReleaseAPI(rawURL string) bool {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsed.Scheme != "https" || parsed.Host != "api.github.com" {
		return false
	}
	path := strings.TrimSuffix(parsed.EscapedPath(), "/")
	const releasesPath = "/repos/ichwars/RailKeeper/releases"
	return path == releasesPath || strings.HasPrefix(path, releasesPath+"/")
}

func isTrustedRailKeeperAssetURL(rawURL, tagName, expectedName string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme != "https" || parsed.Host != "github.com" {
		return false
	}
	segments := strings.Split(strings.Trim(parsed.EscapedPath(), "/"), "/")
	if len(segments) != 6 || segments[0] != "ichwars" || segments[1] != "RailKeeper" ||
		segments[2] != "releases" || segments[3] != "download" {
		return false
	}
	assetTag, tagErr := url.PathUnescape(segments[4])
	assetName, nameErr := url.PathUnescape(segments[5])
	return tagErr == nil && nameErr == nil && assetTag == tagName && assetName == expectedName
}

func compareVersionStrings(latest, current string) int {
	latestParts := numericVersionParts(latest)
	currentParts := numericVersionParts(current)
	if len(latestParts) == 0 || len(currentParts) == 0 {
		return strings.Compare(strings.TrimSpace(latest), strings.TrimSpace(current))
	}

	length := len(latestParts)
	if len(currentParts) > length {
		length = len(currentParts)
	}
	for i := 0; i < length; i++ {
		var left, right int
		if i < len(latestParts) {
			left = latestParts[i]
		}
		if i < len(currentParts) {
			right = currentParts[i]
		}
		if left > right {
			return 1
		}
		if left < right {
			return -1
		}
	}
	latestPrerelease := versionPrerelease(latest)
	currentPrerelease := versionPrerelease(current)
	if latestPrerelease != "" && currentPrerelease == "" {
		return -1
	}
	if latestPrerelease == "" && currentPrerelease != "" {
		return 1
	}
	if latestPrerelease != "" || currentPrerelease != "" {
		return strings.Compare(latestPrerelease, currentPrerelease)
	}
	return 0
}

func numericVersionParts(value string) []int {
	cleaned := versionCore(value)
	matches := regexp.MustCompile(`\d+`).FindAllString(cleaned, -1)
	parts := make([]int, 0, len(matches))
	for _, match := range matches {
		var part int
		if _, err := fmt.Sscanf(match, "%d", &part); err == nil {
			parts = append(parts, part)
		}
	}
	return parts
}

func versionCore(value string) string {
	cleaned := strings.TrimPrefix(strings.TrimSpace(value), "v")
	if index := strings.IndexAny(cleaned, "-+"); index >= 0 {
		return cleaned[:index]
	}
	return cleaned
}

func versionPrerelease(value string) string {
	cleaned := strings.TrimPrefix(strings.TrimSpace(value), "v")
	start := strings.Index(cleaned, "-")
	if start < 0 {
		return ""
	}
	prerelease := cleaned[start+1:]
	if end := strings.Index(prerelease, "+"); end >= 0 {
		prerelease = prerelease[:end]
	}
	return prerelease
}
