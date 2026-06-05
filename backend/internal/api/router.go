package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"
	"log/slog"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "golang.org/x/image/webp"

	"railkeeper/backend/internal/application"
)

type Config struct {
	Version                     string
	UpdateCheckURL              string
	StaticDir                   string
	DataDir                     string
	MaxImageBytes               int64
	MaxAttachmentBytes          int64
	AllowedAttachmentExtensions map[string]struct{}
	Logger                      *slog.Logger
	SetupService                *application.SetupService
	AuthService                 *application.AuthService
	VehicleService              *application.VehicleService
	MasterDataService           *application.MasterDataService
	ArticleSearch               *application.ArticleSearchService
	InventoryNumbers            *application.InventoryNumberService
	BackupService               *application.BackupService
	FileBlobService             *application.FileBlobService
	DatabaseMaintenance         *application.DatabaseMaintenanceService
	ExhibitionService           *application.ExhibitionService
	ECoSService                 *application.ECoSService
	DigitalCenterService        *application.DigitalCenterService
	SettingsService             *application.SettingsService
	RateLimitService            *application.RateLimitService
	PasswordResetMailer         application.PasswordResetMailer
	SMTPSettingsService         *application.SMTPSettingsService
	PublicURL                   string
	CookieSecure                bool
}

type App struct {
	version                     string
	updateCheckURL              string
	staticDir                   string
	dataDir                     string
	maxImageBytes               int64
	maxAttachmentBytes          int64
	allowedAttachmentExtensions map[string]struct{}
	logger                      *slog.Logger
	setupService                *application.SetupService
	authService                 *application.AuthService
	vehicleService              *application.VehicleService
	masterDataService           *application.MasterDataService
	articleSearch               *application.ArticleSearchService
	inventoryNumbers            *application.InventoryNumberService
	backupService               *application.BackupService
	fileBlobs                   *application.FileBlobService
	databaseMaintenance         *application.DatabaseMaintenanceService
	exhibitionService           *application.ExhibitionService
	ecosService                 *application.ECoSService
	digitalCenterService        *application.DigitalCenterService
	settingsService             *application.SettingsService
	passwordResetMailer         application.PasswordResetMailer
	smtpSettingsService         *application.SMTPSettingsService
	publicURL                   string
	cookieSecure                bool
	rateLimits                  rateLimitStore
}

func NewRouter(config Config) http.Handler {
	if config.Logger == nil {
		config.Logger = slog.Default()
	}
	if config.DataDir == "" {
		config.DataDir = "./data"
	}
	app := &App{
		version:                     config.Version,
		updateCheckURL:              config.UpdateCheckURL,
		staticDir:                   config.StaticDir,
		dataDir:                     config.DataDir,
		maxImageBytes:               effectiveLimit(config.MaxImageBytes, defaultMaxImageBytes),
		maxAttachmentBytes:          effectiveLimit(config.MaxAttachmentBytes, defaultMaxAttachmentBytes),
		allowedAttachmentExtensions: effectiveAttachmentExtensions(config.AllowedAttachmentExtensions),
		logger:                      config.Logger,
		setupService:                config.SetupService,
		authService:                 config.AuthService,
		vehicleService:              config.VehicleService,
		masterDataService:           config.MasterDataService,
		articleSearch:               config.ArticleSearch,
		inventoryNumbers:            config.InventoryNumbers,
		backupService:               config.BackupService,
		fileBlobs:                   config.FileBlobService,
		databaseMaintenance:         config.DatabaseMaintenance,
		exhibitionService:           config.ExhibitionService,
		ecosService:                 config.ECoSService,
		digitalCenterService:        config.DigitalCenterService,
		settingsService:             config.SettingsService,
		passwordResetMailer:         config.PasswordResetMailer,
		smtpSettingsService:         config.SMTPSettingsService,
		publicURL:                   strings.TrimRight(strings.TrimSpace(config.PublicURL), "/"),
		cookieSecure:                config.CookieSecure,
		rateLimits:                  config.RateLimitService,
	}
	if app.rateLimits == nil {
		app.rateLimits = newRateLimiter()
	}
	if app.articleSearch == nil {
		app.articleSearch = application.NewArticleSearchService(app.masterDataService)
	}
	if app.backupService == nil {
		app.backupService = application.NewBackupService(nil, app.dataDir)
	}
	if app.ecosService == nil {
		app.ecosService = application.NewECoSService()
	}
	if app.digitalCenterService == nil {
		app.digitalCenterService = application.NewDigitalCenterService()
	}
	if app.vehicleService != nil {
		app.vehicleService.SetImageLocalizer(app.localizeVehicleImages)
	}

	mux := http.NewServeMux()

	app.registerRoutes(mux)
	mux.Handle("/", staticHandler(app.staticDir))

	return securityHeaders(app.csrf(mux))
}

type storageUsageResponse struct {
	TotalBytes int64                  `json:"totalBytes"`
	Categories []storageUsageCategory `json:"categories"`
	UpdatedAt  string                 `json:"updatedAt"`
}

type storageUsageCategory struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Bytes int64  `json:"bytes"`
	Files int    `json:"files"`
}

type systemPrintersResponse struct {
	Status         string          `json:"status"`
	Message        string          `json:"message"`
	DefaultPrinter string          `json:"defaultPrinter,omitempty"`
	Printers       []systemPrinter `json:"printers"`
}

type systemPrinter struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	IsDefault bool   `json:"isDefault"`
}

type versionInfoResponse struct {
	Version         string `json:"version"`
	LatestVersion   string `json:"latestVersion,omitempty"`
	UpdateAvailable bool   `json:"updateAvailable"`
	SourceURL       string `json:"sourceUrl,omitempty"`
	ReleaseURL      string `json:"releaseUrl,omitempty"`
	ReleaseNotes    string `json:"releaseNotes,omitempty"`
	AssetURL        string `json:"assetUrl,omitempty"`
	AssetName       string `json:"assetName,omitempty"`
	CheckedAt       string `json:"checkedAt"`
	Status          string `json:"status"`
	Message         string `json:"message"`
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

type eCoSLocomotiveSyncRequest struct {
	Host      string `json:"host"`
	Port      int    `json:"port"`
	VehicleID string `json:"vehicleId"`
	ObjectID  int    `json:"objectId"`
	DryRun    bool   `json:"dryRun"`
	Confirm   bool   `json:"confirm"`
}

var errNoUpdateRelease = errors.New("no update release available")

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
	release, err := fetchUpdateRelease(r.Context(), updateURL, includePrerelease)
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
	response.AssetName, response.AssetURL = firstReleaseAsset(release.Assets)
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
	} else {
		response.Status = "current"
		response.Message = "RailKeeper ist aktuell."
	}
	respondJSON(w, http.StatusOK, response)
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

func firstReleaseAsset(assets []updateReleaseAsset) (string, string) {
	for _, asset := range assets {
		if trimmed := strings.TrimSpace(asset.BrowserDownloadURL); trimmed != "" {
			return strings.TrimSpace(asset.Name), trimmed
		}
	}
	return "", ""
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

func (a *App) systemStorage(w http.ResponseWriter, r *http.Request) {
	categories := map[string]*storageUsageCategory{
		"database":     {Key: "database", Label: "Datenbank"},
		"images":       {Key: "images", Label: "Bilder"},
		"thumbnails":   {Key: "thumbnails", Label: "Vorschaubilder"},
		"attachments":  {Key: "attachments", Label: "Beilagen"},
		"decoderFiles": {Key: "decoderFiles", Label: "Decoder-Dateien"},
		"other":        {Key: "other", Label: "Sonstiges"},
	}

	var total int64
	err := filepath.WalkDir(a.dataDir, func(filePath string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}

		relativePath, err := filepath.Rel(a.dataDir, filePath)
		if err != nil {
			relativePath = filePath
		}
		key := storageCategoryKey(relativePath)
		category := categories[key]
		category.Bytes += info.Size()
		category.Files += 1
		total += info.Size()
		return nil
	})
	if err != nil {
		a.logger.Error("storage usage scan failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "storage_usage_failed", "Speichernutzung konnte nicht gelesen werden.")
		return
	}

	order := []string{"database", "images", "thumbnails", "attachments", "decoderFiles", "other"}
	result := make([]storageUsageCategory, 0, len(order))
	for _, key := range order {
		category := categories[key]
		if category.Files == 0 && key != "database" {
			continue
		}
		result = append(result, *category)
	}

	respondJSON(w, http.StatusOK, storageUsageResponse{
		TotalBytes: total,
		Categories: result,
		UpdatedAt:  time.Now().UTC().Format(time.RFC3339),
	})
}

func (a *App) optimizeSystemStorage(w http.ResponseWriter, r *http.Request) {
	if a.databaseMaintenance == nil {
		respondProblem(w, http.StatusInternalServerError, "storage_optimize_unavailable", "Datenbankoptimierung ist nicht konfiguriert.")
		return
	}
	result, err := a.databaseMaintenance.Optimize(r.Context())
	if err != nil {
		a.logger.Error("database optimize failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "storage_optimize_failed", "Datenbank konnte nicht optimiert werden.")
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func (a *App) systemPrinters(w http.ResponseWriter, r *http.Request) {
	response := discoverSystemPrinters()
	respondJSON(w, http.StatusOK, response)
}

func discoverSystemPrinters() systemPrintersResponse {
	if response := printersFromEnv(); len(response.Printers) > 0 {
		return response
	}

	switch runtime.GOOS {
	case "linux", "darwin":
		return printersFromLPStat()
	case "windows":
		return printersFromPowerShell()
	default:
		return systemPrintersResponse{
			Status:   "unavailable",
			Message:  "Druckerabfrage ist auf dieser Plattform nicht verfügbar. Der Browser-Systemdialog bleibt aktiv.",
			Printers: []systemPrinter{},
		}
	}
}

func printersFromEnv() systemPrintersResponse {
	configured := strings.TrimSpace(os.Getenv("RAILKEEPER_PRINTERS"))
	if configured == "" {
		return systemPrintersResponse{}
	}
	defaultPrinter := strings.TrimSpace(os.Getenv("RAILKEEPER_DEFAULT_PRINTER"))
	printers := []systemPrinter{}
	seen := map[string]struct{}{}
	for _, part := range strings.Split(configured, ",") {
		name := strings.TrimSpace(part)
		if name == "" {
			continue
		}
		id := printerID(name)
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		printers = append(printers, systemPrinter{
			ID:        id,
			Name:      name,
			IsDefault: defaultPrinter != "" && strings.EqualFold(defaultPrinter, name),
		})
	}
	if defaultPrinter == "" && len(printers) > 0 {
		printers[0].IsDefault = true
		defaultPrinter = printers[0].Name
	}
	return systemPrintersResponse{
		Status:         "configured",
		Message:        "Druckerliste wurde aus der RailKeeper-Konfiguration gelesen.",
		DefaultPrinter: defaultPrinter,
		Printers:       printers,
	}
}

func printersFromLPStat() systemPrintersResponse {
	namesOut, err := exec.Command("lpstat", "-e").Output()
	if err != nil {
		return systemPrintersResponse{
			Status:   "unavailable",
			Message:  "Keine Systemdrucker im Container oder auf dem Host ermittelbar. Der Browser-Systemdialog bleibt aktiv.",
			Printers: []systemPrinter{},
		}
	}
	defaultPrinter := ""
	if defaultOut, err := exec.Command("lpstat", "-d").Output(); err == nil {
		defaultPrinter = parseLPStatDefault(string(defaultOut))
	}
	printers := printersFromNames(strings.Fields(string(namesOut)), defaultPrinter)
	return systemPrintersResponse{
		Status:         "available",
		Message:        "Systemdrucker wurden über CUPS gelesen.",
		DefaultPrinter: defaultPrinter,
		Printers:       printers,
	}
}

func printersFromPowerShell() systemPrintersResponse {
	script := `Get-CimInstance Win32_Printer | Select-Object Name,Default | ConvertTo-Json -Compress`
	output, err := exec.Command("powershell", "-NoProfile", "-Command", script).Output()
	if err != nil {
		return systemPrintersResponse{
			Status:   "unavailable",
			Message:  "Windows-Drucker konnten nicht gelesen werden. Der Browser-Systemdialog bleibt aktiv.",
			Printers: []systemPrinter{},
		}
	}
	type windowsPrinter struct {
		Name    string `json:"Name"`
		Default bool   `json:"Default"`
	}
	var many []windowsPrinter
	if err := json.Unmarshal(output, &many); err != nil {
		var one windowsPrinter
		if err := json.Unmarshal(output, &one); err != nil {
			return systemPrintersResponse{
				Status:   "unavailable",
				Message:  "Windows-Druckerantwort konnte nicht ausgewertet werden.",
				Printers: []systemPrinter{},
			}
		}
		many = []windowsPrinter{one}
	}
	printers := []systemPrinter{}
	defaultPrinter := ""
	for _, printer := range many {
		name := strings.TrimSpace(printer.Name)
		if name == "" {
			continue
		}
		if printer.Default {
			defaultPrinter = name
		}
		printers = append(printers, systemPrinter{
			ID:        printerID(name),
			Name:      name,
			IsDefault: printer.Default,
		})
	}
	return systemPrintersResponse{
		Status:         "available",
		Message:        "Windows-Systemdrucker wurden gelesen.",
		DefaultPrinter: defaultPrinter,
		Printers:       printers,
	}
}

func parseLPStatDefault(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if index := strings.LastIndex(value, ":"); index >= 0 {
		return strings.TrimSpace(value[index+1:])
	}
	return ""
}

func printersFromNames(names []string, defaultPrinter string) []systemPrinter {
	printers := []systemPrinter{}
	seen := map[string]struct{}{}
	for _, rawName := range names {
		name := strings.TrimSpace(rawName)
		if name == "" {
			continue
		}
		id := printerID(name)
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		printers = append(printers, systemPrinter{
			ID:        id,
			Name:      name,
			IsDefault: defaultPrinter != "" && strings.EqualFold(defaultPrinter, name),
		})
	}
	return printers
}

func printerID(name string) string {
	id := strings.ToLower(strings.TrimSpace(name))
	id = regexp.MustCompile(`[^a-z0-9._-]+`).ReplaceAllString(id, "-")
	id = strings.Trim(id, "-")
	if id == "" {
		return "printer"
	}
	return id
}

func storageCategoryKey(relativePath string) string {
	clean := filepath.ToSlash(relativePath)
	name := path.Base(clean)
	if name == "railkeeper.db" || name == "railkeeper.db-wal" || name == "railkeeper.db-shm" {
		return "database"
	}
	parts := strings.Split(clean, "/")
	if pathContains(parts, "thumbs") {
		return "thumbnails"
	}
	if pathContains(parts, "images") {
		return "images"
	}
	if pathContains(parts, "cv") {
		return "decoderFiles"
	}
	if len(parts) > 0 && parts[0] == "uploads" {
		return "attachments"
	}
	return "other"
}

func pathContains(parts []string, needle string) bool {
	for _, part := range parts {
		if part == needle {
			return true
		}
	}
	return false
}

func (a *App) setupStatus(w http.ResponseWriter, r *http.Request) {
	required, err := a.setupService.SetupRequired(r.Context())
	if err != nil {
		a.logger.Error("setup status failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "setup_status_failed", "Could not read setup state.")
		return
	}

	respondJSON(w, http.StatusOK, map[string]bool{"setupRequired": required})
}

func (a *App) createAdmin(w http.ResponseWriter, r *http.Request) {
	allowed, ok := a.allowRequest(w, r, "setup", clientIP(r), 5, 10*time.Minute)
	if !ok {
		return
	}
	if !allowed {
		respondProblem(w, http.StatusTooManyRequests, "rate_limited", "Too many setup attempts. Please try again later.")
		return
	}

	var input application.CreateAdminInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	if err := a.setupService.CreateAdmin(r.Context(), input); err != nil {
		switch {
		case errors.Is(err, application.ErrWeakSetup):
			respondProblem(w, http.StatusBadRequest, "weak_setup", "Username, valid email and password with at least 12 characters are required.")
		case errors.Is(err, application.ErrAlreadySetup):
			respondProblem(w, http.StatusConflict, "already_setup", "Setup has already been completed.")
		default:
			a.logger.Error("admin setup failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "setup_failed", "Could not create admin user.")
		}
		return
	}

	respondJSON(w, http.StatusCreated, map[string]string{"status": "created"})
}

func (a *App) login(w http.ResponseWriter, r *http.Request) {
	allowed, ok := a.allowRequest(w, r, "login", clientIP(r), 10, 5*time.Minute)
	if !ok {
		return
	}
	if !allowed {
		respondProblem(w, http.StatusTooManyRequests, "rate_limited", "Too many login attempts. Please try again later.")
		return
	}

	var input application.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	result, err := a.authService.Login(r.Context(), input)
	if err != nil {
		if errors.Is(err, application.ErrTwoFactorRequired) {
			respondProblem(w, http.StatusUnauthorized, "two_factor_required", "Zwei-Faktor-Code erforderlich.")
			return
		}
		if errors.Is(err, application.ErrInvalidLogin) {
			respondProblem(w, http.StatusUnauthorized, "invalid_login", "Invalid username or password.")
			return
		}
		a.logger.Error("login failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "login_failed", "Could not create session.")
		return
	}

	setCookie(w, "rk_session", result.SessionToken, int(timeUntil(result.ExpiresAt).Seconds()), true, a.cookieSecure)
	setCookie(w, "rk_csrf", result.CSRFToken, int(timeUntil(result.ExpiresAt).Seconds()), false, a.cookieSecure)
	respondJSON(w, http.StatusOK, result.Session)
}

func (a *App) requestPasswordReset(w http.ResponseWriter, r *http.Request) {
	allowed, ok := a.allowRequest(w, r, "password-reset", clientIP(r), 5, 10*time.Minute)
	if !ok {
		return
	}
	if !allowed {
		respondProblem(w, http.StatusTooManyRequests, "rate_limited", "Too many reset attempts. Please try again later.")
		return
	}

	var input application.PasswordResetRequestInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	result, err := a.authService.RequestPasswordReset(r.Context(), input)
	if err != nil {
		if errors.Is(err, application.ErrUserValidation) {
			respondProblem(w, http.StatusBadRequest, "invalid_email", "Bitte eine g?ltige E-Mail-Adresse angeben.")
			return
		}
		a.logger.Error("password reset request failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "password_reset_failed", "Passwort-Zurücksetzung konnte nicht vorbereitet werden.")
		return
	}
	if result.ResetToken != "" {
		resetURL := a.passwordResetURL(r, result.ResetToken)
		mailer := a.passwordResetMailer
		if a.smtpSettingsService != nil {
			settingsMailer, publicURL, err := a.smtpSettingsService.EffectiveMailer(r.Context())
			if err != nil {
				a.logger.Error("smtp settings invalid", "error", err)
			} else if settingsMailer != nil {
				mailer = settingsMailer
				resetURL = a.passwordResetURLWithBase(r, result.ResetToken, publicURL)
			}
		}
		if mailer != nil {
			if err := mailer.SendPasswordReset(r.Context(), input.Email, resetURL, result.ExpiresAt); err != nil {
				a.logger.Error("password reset email failed", "error", err)
			}
		} else {
			a.logger.Warn("password reset email disabled; link is available in server log for local recovery only", "reset_url", resetURL)
		}
	}
	result.ResetToken = ""
	result.ResetURL = ""
	respondJSON(w, http.StatusAccepted, result)
}

func (a *App) confirmPasswordReset(w http.ResponseWriter, r *http.Request) {
	allowed, ok := a.allowRequest(w, r, "password-reset-confirm", clientIP(r), 10, 10*time.Minute)
	if !ok {
		return
	}
	if !allowed {
		respondProblem(w, http.StatusTooManyRequests, "rate_limited", "Too many reset attempts. Please try again later.")
		return
	}

	var input application.PasswordResetConfirmInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	if err := a.authService.ResetPassword(r.Context(), input); err != nil {
		if errors.Is(err, application.ErrUserValidation) {
			respondProblem(w, http.StatusBadRequest, "invalid_password_reset", "Reset-Link und neues Passwort m?ssen g?ltig sein.")
			return
		}
		if errors.Is(err, application.ErrPasswordResetInvalid) {
			respondProblem(w, http.StatusBadRequest, "invalid_reset_token", "Reset-Link ist ung?ltig oder abgelaufen.")
			return
		}
		a.logger.Error("password reset confirmation failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "password_reset_failed", "Passwort konnte nicht zur?ckgesetzt werden.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) passwordResetURL(r *http.Request, token string) string {
	return a.passwordResetURLWithBase(r, token, a.publicURL)
}

func (a *App) passwordResetURLWithBase(r *http.Request, token string, baseURL string) string {
	if strings.TrimSpace(baseURL) != "" {
		u, err := url.Parse(strings.TrimRight(strings.TrimSpace(baseURL), "/"))
		if err == nil {
			u.Path = "/password-reset"
			query := u.Query()
			query.Set("token", token)
			u.RawQuery = query.Encode()
			return u.String()
		}
	}
	scheme := "http"
	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")); forwarded != "" {
		scheme = strings.Split(forwarded, ",")[0]
	} else if r.TLS != nil {
		scheme = "https"
	}
	u := url.URL{
		Scheme: scheme,
		Host:   r.Host,
		Path:   "/password-reset",
	}
	query := u.Query()
	query.Set("token", token)
	u.RawQuery = query.Encode()
	return u.String()
}

type smtpTestRequest struct {
	Recipient string `json:"recipient"`
}

func (a *App) getSMTPSettings(w http.ResponseWriter, r *http.Request) {
	if a.smtpSettingsService == nil {
		respondJSON(w, http.StatusOK, application.SMTPSettings{TLSMode: "starttls", Port: "587"})
		return
	}
	settings, err := a.smtpSettingsService.Get(r.Context())
	if err != nil {
		a.logger.Error("smtp settings load failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "smtp_settings_failed", "SMTP-Einstellungen konnten nicht geladen werden.")
		return
	}
	respondJSON(w, http.StatusOK, settings)
}

func (a *App) updateSMTPSettings(w http.ResponseWriter, r *http.Request) {
	if a.smtpSettingsService == nil {
		respondProblem(w, http.StatusServiceUnavailable, "smtp_settings_unavailable", "SMTP-Einstellungen sind nicht verfügbar.")
		return
	}
	var input application.SMTPSettingsInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	settings, err := a.smtpSettingsService.Update(r.Context(), input)
	if err != nil {
		if errors.Is(err, application.ErrSMTPSettingsValidation) {
			respondProblem(w, http.StatusBadRequest, "smtp_settings_validation", err.Error())
			return
		}
		a.logger.Error("smtp settings update failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "smtp_settings_failed", "SMTP-Einstellungen konnten nicht gespeichert werden.")
		return
	}
	respondJSON(w, http.StatusOK, settings)
}

func (a *App) testSMTPSettings(w http.ResponseWriter, r *http.Request) {
	if a.smtpSettingsService == nil {
		respondProblem(w, http.StatusServiceUnavailable, "smtp_settings_unavailable", "SMTP-Einstellungen sind nicht verfügbar.")
		return
	}
	var input smtpTestRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	mailer, _, err := a.smtpSettingsService.EffectiveMailer(r.Context())
	if err != nil {
		if errors.Is(err, application.ErrSMTPSettingsValidation) {
			respondProblem(w, http.StatusBadRequest, "smtp_settings_validation", err.Error())
			return
		}
		a.logger.Error("smtp settings load failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "smtp_settings_failed", "SMTP-Einstellungen konnten nicht geladen werden.")
		return
	}
	if mailer == nil {
		respondProblem(w, http.StatusBadRequest, "smtp_disabled", "SMTP ist nicht aktiviert oder unvollständig konfiguriert.")
		return
	}
	if err := mailer.SendTest(r.Context(), input.Recipient); err != nil {
		a.logger.Error("smtp test email failed", "error", err)
		respondProblem(w, http.StatusBadGateway, "smtp_test_failed", err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "sent"})
}

func (a *App) logout(w http.ResponseWriter, r *http.Request) {
	sessionToken := cookieValue(r, "rk_session")
	if err := a.authService.Logout(r.Context(), sessionToken); err != nil {
		a.logger.Error("logout failed", "error", err)
	}

	clearCookie(w, "rk_session", true, a.cookieSecure)
	clearCookie(w, "rk_csrf", false, a.cookieSecure)
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) session(w http.ResponseWriter, r *http.Request) {
	sessionToken := cookieValue(r, "rk_session")
	session, err := a.authService.CurrentSession(r.Context(), sessionToken)
	if err != nil {
		if errors.Is(err, application.ErrUnauthorized) {
			respondProblem(w, http.StatusUnauthorized, "unauthorized", "Not logged in.")
			return
		}
		a.logger.Error("session lookup failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "session_failed", "Could not read current session.")
		return
	}

	respondJSON(w, http.StatusOK, session)
}

func (a *App) changePassword(w http.ResponseWriter, r *http.Request) {
	var input application.ChangePasswordInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	if err := a.authService.ChangeOwnPassword(r.Context(), actorUserID(r), cookieValue(r, "rk_session"), input); err != nil {
		switch {
		case errors.Is(err, application.ErrUserValidation):
			respondProblem(w, http.StatusBadRequest, "weak_password", "Das neue Passwort muss mindestens 12 Zeichen lang sein.")
		case errors.Is(err, application.ErrInvalidLogin):
			respondProblem(w, http.StatusUnauthorized, "invalid_password", "Das aktuelle Passwort ist nicht korrekt.")
		case errors.Is(err, application.ErrUserNotFound), errors.Is(err, application.ErrUnauthorized):
			respondProblem(w, http.StatusUnauthorized, "unauthorized", "Not logged in.")
		default:
			a.logger.Error("password change failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "password_change_failed", "Passwort konnte nicht geändert werden.")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *App) listVehicles(w http.ResponseWriter, r *http.Request) {
	vehicles, err := a.vehicleService.List(r.Context(), r.URL.Query().Get("q"))
	if err != nil {
		a.logger.Error("vehicle list failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "vehicle_list_failed", "Could not list vehicles.")
		return
	}

	respondJSON(w, http.StatusOK, vehicles)
}

func (a *App) testECoSConnection(w http.ResponseWriter, r *http.Request) {
	var input application.ECoSConnectionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	result, err := a.ecosService.TestConnection(r.Context(), input)
	if err != nil {
		respondProblem(w, http.StatusBadRequest, "ecos_validation", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, result)
}

func (a *App) probeECoSLocomotiveRaw(w http.ResponseWriter, r *http.Request) {
	var input application.ECoSConnectionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	probe, err := a.ecosService.ProbeLocomotiveRaw(r.Context(), input)
	if err != nil {
		respondProblem(w, http.StatusBadGateway, "ecos_raw_probe_failed", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, probe)
}

func (a *App) countECoSLocomotives(w http.ResponseWriter, r *http.Request) {
	var input application.ECoSConnectionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	summary, err := a.ecosService.CountLocomotives(r.Context(), input)
	if err != nil {
		respondProblem(w, http.StatusBadGateway, "ecos_locomotive_count_failed", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, summary)
}

func (a *App) syncECoSLocomotive(w http.ResponseWriter, r *http.Request) {
	var request eCoSLocomotiveSyncRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	vehicle, err := a.vehicleService.Get(r.Context(), request.VehicleID)
	if err != nil {
		if errors.Is(err, application.ErrVehicleNotFound) {
			respondProblem(w, http.StatusNotFound, "vehicle_not_found", "Vehicle not found.")
			return
		}
		a.logger.Error("vehicle read for ecos sync failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "vehicle_read_failed", "Could not read vehicle.")
		return
	}

	objectID := request.ObjectID
	mapping := vehicleECoSMappingForSync(vehicle, objectID)
	if objectID <= 0 && mapping != nil {
		objectID = parsePositiveIntText(mapping.ExternalID)
	}
	if objectID <= 0 {
		respondProblem(w, http.StatusBadRequest, "ecos_object_required", "ECoS object ID is required.")
		return
	}

	desired := application.ECoSLocomotiveSyncDesired{
		Name:     vehicle.Name,
		Address:  parsePositiveIntText(vehicle.DigitalDecoderNumber),
		Protocol: "",
	}
	if mapping != nil {
		if address := parsePositiveIntText(mapping.ExternalAddress); address > 0 {
			desired.Address = address
		}
		desired.Protocol = mapping.ExternalProtocol
	}

	result, err := a.ecosService.SyncLocomotive(r.Context(), application.ECoSLocomotiveSyncInput{
		Host:     request.Host,
		Port:     request.Port,
		ObjectID: objectID,
		Desired:  desired,
		DryRun:   request.DryRun,
		Confirm:  request.Confirm,
	})
	if err != nil {
		respondProblem(w, http.StatusBadGateway, "ecos_sync_failed", err.Error())
		return
	}

	if (request.Confirm && len(result.Changes) == 0) || result.Applied {
		address := ""
		if desired.Address > 0 {
			address = strconv.Itoa(desired.Address)
		}
		if _, err := a.vehicleService.UpsertExternalMapping(r.Context(), vehicle.ID, application.VehicleExternalMapInput{
			Provider:         "ecos",
			ExternalID:       strconv.Itoa(objectID),
			ExternalName:     desired.Name,
			ExternalAddress:  address,
			ExternalProtocol: desired.Protocol,
			SyncStatus:       "synced",
		}, actorUserID(r)); err != nil {
			a.logger.Warn("ecos sync mapping update failed", "vehicleID", vehicle.ID, "objectID", objectID, "error", err)
		}
	}

	respondJSON(w, http.StatusOK, result)
}

func (a *App) eCoSLiveStatus(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, a.ecosService.LiveStatus())
}

func (a *App) startECoSLive(w http.ResponseWriter, r *http.Request) {
	var input application.ECoSConnectionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	status, err := a.ecosService.StartLive(r.Context(), input)
	if err != nil {
		respondProblem(w, http.StatusBadGateway, "ecos_live_start_failed", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, status)
}

func (a *App) stopECoSLive(w http.ResponseWriter, r *http.Request) {
	status := a.ecosService.StopLive()
	respondJSON(w, http.StatusOK, status)
}

func (a *App) testZ21Connection(w http.ResponseWriter, r *http.Request) {
	var input application.DigitalCenterConnectionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	result, err := a.digitalCenterService.TestZ21Connection(r.Context(), input)
	if err != nil {
		respondProblem(w, http.StatusBadRequest, "digital_center_validation", err.Error())
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func (a *App) probeZ21Connection(w http.ResponseWriter, r *http.Request) {
	var input application.DigitalCenterConnectionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	result, err := a.digitalCenterService.ProbeZ21Connection(r.Context(), input)
	if err != nil {
		respondProblem(w, http.StatusBadRequest, "digital_center_validation", err.Error())
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func (a *App) testIntellibox3Connection(w http.ResponseWriter, r *http.Request) {
	var input application.DigitalCenterConnectionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	result, err := a.digitalCenterService.TestIntellibox3Connection(r.Context(), input)
	if err != nil {
		respondProblem(w, http.StatusBadRequest, "digital_center_validation", err.Error())
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func (a *App) probeIntellibox3Connection(w http.ResponseWriter, r *http.Request) {
	var input application.DigitalCenterConnectionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	result, err := a.digitalCenterService.ProbeIntellibox3Connection(r.Context(), input)
	if err != nil {
		respondProblem(w, http.StatusBadRequest, "digital_center_validation", err.Error())
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func (a *App) testCS3Connection(w http.ResponseWriter, r *http.Request) {
	var input application.DigitalCenterConnectionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	result, err := a.digitalCenterService.TestCS3Connection(r.Context(), input)
	if err != nil {
		respondProblem(w, http.StatusBadRequest, "digital_center_validation", err.Error())
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func (a *App) createVehicle(w http.ResponseWriter, r *http.Request) {
	var input application.CreateVehicleInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	vehicle, err := a.vehicleService.Create(r.Context(), input, actorUserID(r))
	if err != nil {
		switch {
		case errors.Is(err, application.ErrVehicleValidation), errors.Is(err, application.ErrInventoryNumberValidation):
			respondProblem(w, http.StatusBadRequest, "vehicle_validation", "Manufacturer, name, gauge, category and subtype are required.")
		case errors.Is(err, application.ErrInventoryNumberConflict):
			respondProblem(w, http.StatusConflict, "inventory_number_conflict", "Inventory number already exists.")
		case errors.Is(err, application.ErrInventoryNumberNotFound):
			respondProblem(w, http.StatusBadRequest, "inventory_number_scheme_missing", "No active inventory number scheme is available.")
		default:
			a.logger.Error("vehicle create failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "vehicle_create_failed", "Could not create vehicle.")
		}
		return
	}

	respondJSON(w, http.StatusCreated, vehicle)
}

func (a *App) getVehicle(w http.ResponseWriter, r *http.Request) {
	vehicle, err := a.vehicleService.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		if errors.Is(err, application.ErrVehicleNotFound) {
			respondProblem(w, http.StatusNotFound, "vehicle_not_found", "Vehicle not found.")
			return
		}
		a.logger.Error("vehicle get failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "vehicle_get_failed", "Could not read vehicle.")
		return
	}

	respondJSON(w, http.StatusOK, vehicle)
}

func (a *App) updateVehicle(w http.ResponseWriter, r *http.Request) {
	var input application.CreateVehicleInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	vehicle, err := a.vehicleService.Update(r.Context(), r.PathValue("id"), input, actorUserID(r))
	if err != nil {
		switch {
		case errors.Is(err, application.ErrVehicleValidation), errors.Is(err, application.ErrInventoryNumberValidation):
			respondProblem(w, http.StatusBadRequest, "vehicle_validation", "Manufacturer, name, gauge, category and subtype are required.")
		case errors.Is(err, application.ErrInventoryNumberConflict):
			respondProblem(w, http.StatusConflict, "inventory_number_conflict", "Inventory number already exists.")
		case errors.Is(err, application.ErrVehicleNotFound):
			respondProblem(w, http.StatusNotFound, "vehicle_not_found", "Vehicle not found.")
		default:
			a.logger.Error("vehicle update failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "vehicle_update_failed", "Could not update vehicle.")
		}
		return
	}

	respondJSON(w, http.StatusOK, vehicle)
}

func (a *App) upsertVehicleExternalMapping(w http.ResponseWriter, r *http.Request) {
	var input application.VehicleExternalMapInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	mapping, err := a.vehicleService.UpsertExternalMapping(r.Context(), r.PathValue("id"), input, actorUserID(r))
	if err != nil {
		switch {
		case errors.Is(err, application.ErrVehicleValidation):
			respondProblem(w, http.StatusBadRequest, "external_mapping_validation", "Provider and external id are required.")
		case errors.Is(err, application.ErrVehicleNotFound):
			respondProblem(w, http.StatusNotFound, "vehicle_not_found", "Vehicle not found.")
		default:
			a.logger.Error("vehicle external mapping failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "external_mapping_failed", "Could not save external mapping.")
		}
		return
	}

	respondJSON(w, http.StatusOK, mapping)
}

func (a *App) deleteVehicle(w http.ResponseWriter, r *http.Request) {
	if err := a.vehicleService.Delete(r.Context(), r.PathValue("id"), actorUserID(r)); err != nil {
		if errors.Is(err, application.ErrVehicleNotFound) {
			respondProblem(w, http.StatusNotFound, "vehicle_not_found", "Vehicle not found.")
			return
		}
		a.logger.Error("vehicle delete failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "vehicle_delete_failed", "Could not delete vehicle.")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

const (
	defaultMaxAttachmentBytes = 25 * 1024 * 1024
	defaultMaxImageBytes      = 10 * 1024 * 1024
)

func effectiveLimit(value, fallback int64) int64 {
	if value <= 0 {
		return fallback
	}
	return value
}

var safeFileNamePattern = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

var allowedAttachmentExtensions = map[string]struct{}{
	".csv":  {},
	".jpeg": {},
	".jpg":  {},
	".json": {},
	".pdf":  {},
	".png":  {},
	".txt":  {},
	".webp": {},
	".xml":  {},
	".zip":  {},
}

func effectiveAttachmentExtensions(input map[string]struct{}) map[string]struct{} {
	if len(input) == 0 {
		return allowedAttachmentExtensions
	}
	out := map[string]struct{}{}
	for extension := range input {
		extension = strings.ToLower(strings.TrimSpace(extension))
		if extension == "" {
			continue
		}
		if !strings.HasPrefix(extension, ".") {
			extension = "." + extension
		}
		if isBlockedAttachmentName("file" + extension) {
			continue
		}
		out[extension] = struct{}{}
	}
	if len(out) == 0 {
		return allowedAttachmentExtensions
	}
	return out
}

func (a *App) localizeVehicleImages(ctx context.Context, vehicleID string, images []application.VehicleImageInput) ([]application.VehicleImageInput, error) {
	out := make([]application.VehicleImageInput, len(images))
	copy(out, images)
	for index, image := range out {
		if image.StoragePath != "" || image.BlobID != "" || !strings.HasPrefix(strings.ToLower(image.URL), "http") {
			continue
		}
		localized, err := a.localizeVehicleImage(ctx, vehicleID, image)
		if err != nil {
			a.logger.Warn("article image localization skipped", "url", image.URL, "error", err)
			continue
		}
		out[index] = localized
	}
	return out, nil
}

func (a *App) localizeVehicleImage(ctx context.Context, vehicleID string, image application.VehicleImageInput) (application.VehicleImageInput, error) {
	if !isPublicImageURL(ctx, image.URL) {
		return image, fmt.Errorf("image url is not public http(s)")
	}
	requestCtx, cancel := context.WithTimeout(ctx, 6*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(requestCtx, http.MethodGet, image.URL, nil)
	if err != nil {
		return image, err
	}
	req.Header.Set("User-Agent", "RailKeeper/0.1 image-fetch")
	req.Header.Set("Accept", "image/avif,image/webp,image/png,image/jpeg,image/*;q=0.8")
	client := remoteDocumentHTTPClient(ctx)
	client.Timeout = 6 * time.Second
	resp, err := client.Do(req)
	if err != nil {
		return image, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return image, fmt.Errorf("image fetch returned status %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, a.maxImageBytes+1))
	if err != nil || len(data) == 0 || int64(len(data)) > a.maxImageBytes {
		return image, fmt.Errorf("image size invalid")
	}
	mimeType := http.DetectContentType(data)
	if !isAllowedImageMime(mimeType) {
		return image, fmt.Errorf("image type %s is not allowed", mimeType)
	}
	storageName := fmt.Sprintf("%d-%s", time.Now().UTC().UnixNano(), remoteImageFileName(image, mimeType))
	blobID, err := a.storeFileBlob(ctx, data)
	if err != nil {
		return image, err
	}
	thumbnailBlobID, err := a.createVehicleImageThumbnail(ctx, data, storageName)
	if err != nil {
		a.logger.Warn("image thumbnail skipped", "url", image.URL, "error", err)
	}
	if image.SourceURL == "" {
		image.SourceURL = image.URL
	}
	image.FileName = storageName
	image.MimeType = mimeType
	image.BlobID = blobID
	image.ThumbnailBlobID = thumbnailBlobID
	return image, nil
}

func remoteImageFileName(image application.VehicleImageInput, mimeType string) string {
	extension := ".jpg"
	switch mimeType {
	case "image/png":
		extension = ".png"
	case "image/webp":
		extension = ".webp"
	}
	base := strings.TrimSpace(image.Title)
	if base == "" {
		if parsed, err := url.Parse(image.URL); err == nil {
			base = path.Base(parsed.Path)
		}
	}
	base = strings.TrimSuffix(base, filepath.Ext(base))
	if base == "" || base == "." || base == "/" {
		base = "artikelbild"
	}
	return safeAttachmentFileName(base + extension)
}

func remoteAttachmentFileName(input importVehicleAttachmentInput, rawURL, mimeType string) string {
	base := strings.TrimSpace(input.Title)
	if base == "" {
		if parsed, err := url.Parse(rawURL); err == nil {
			base = path.Base(parsed.Path)
		}
	}
	if base == "" || base == "." || base == "/" {
		base = "dokument"
	}
	extension := strings.ToLower(filepath.Ext(base))
	if extension == "" {
		extension = attachmentExtensionForMime(mimeType)
		base += extension
	}
	return safeAttachmentFileName(base)
}

func attachmentExtensionForMime(mimeType string) string {
	switch strings.ToLower(strings.Split(mimeType, ";")[0]) {
	case "application/pdf":
		return ".pdf"
	case "application/json", "text/json":
		return ".json"
	case "application/xml", "text/xml":
		return ".xml"
	case "application/zip", "application/x-zip-compressed":
		return ".zip"
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "text/csv":
		return ".csv"
	default:
		return ".txt"
	}
}

func attachmentCategoryForRemoteDocument(fileName, title string) string {
	lower := strings.ToLower(fileName + " " + title)
	if strings.Contains(lower, "ersatzteil") || strings.Contains(lower, "spare") || strings.Contains(lower, "et-blatt") {
		return "Ersatzteilliste"
	}
	if strings.Contains(lower, "anleitung") || strings.Contains(lower, "manual") || strings.Contains(lower, "bedienung") {
		return "Anleitung"
	}
	return "Dokumentation"
}

func isPublicImageURL(ctx context.Context, value string) bool {
	parsed, err := url.Parse(value)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Hostname() == "" {
		return false
	}
	host := strings.ToLower(parsed.Hostname())
	if host == "localhost" || strings.HasSuffix(host, ".localhost") {
		return false
	}
	if ip := net.ParseIP(host); ip != nil {
		return isPublicIP(ip)
	}
	lookupCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	addresses, err := net.DefaultResolver.LookupIPAddr(lookupCtx, host)
	if err != nil || len(addresses) == 0 {
		return false
	}
	for _, address := range addresses {
		if !isPublicIP(address.IP) {
			return false
		}
	}
	return true
}

func isPublicIP(ip net.IP) bool {
	return ip != nil &&
		!ip.IsLoopback() &&
		!ip.IsPrivate() &&
		!ip.IsLinkLocalUnicast() &&
		!ip.IsLinkLocalMulticast() &&
		!ip.IsMulticast() &&
		!ip.IsUnspecified()
}

func remoteDocumentHTTPClient(ctx context.Context) *http.Client {
	return &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return http.ErrUseLastResponse
			}
			if !isPublicImageURL(ctx, req.URL.String()) {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}
}

func (a *App) uploadVehicleImage(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, a.maxImageBytes+1024*1024)
	if err := r.ParseMultipartForm(a.maxImageBytes); err != nil {
		respondProblem(w, http.StatusBadRequest, "image_upload_invalid", "Bild konnte nicht gelesen werden.")
		return
	}
	if r.MultipartForm != nil {
		defer func() { _ = r.MultipartForm.RemoveAll() }()
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		respondProblem(w, http.StatusBadRequest, "image_missing", "Eine Bilddatei ist erforderlich.")
		return
	}
	defer func() { _ = file.Close() }()
	if header.Size > a.maxImageBytes {
		respondProblem(w, http.StatusBadRequest, "image_too_large", "Das Bild ist zu gro?.")
		return
	}
	data, err := io.ReadAll(io.LimitReader(file, a.maxImageBytes+1))
	if err != nil || int64(len(data)) > a.maxImageBytes {
		respondProblem(w, http.StatusBadRequest, "image_too_large", "Das Bild ist zu gro?.")
		return
	}
	mimeType := http.DetectContentType(data)
	if !isAllowedImageMime(mimeType) {
		respondProblem(w, http.StatusBadRequest, "image_type_blocked", "Erlaubt sind JPG, PNG und WebP.")
		return
	}
	vehicleID := r.PathValue("id")
	storageName := fmt.Sprintf("%d-%s", time.Now().UTC().UnixNano(), safeAttachmentFileName(header.Filename))
	blobID, err := a.storeFileBlob(r.Context(), data)
	if err != nil {
		a.logger.Error("image blob write failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "image_upload_failed", "Bild konnte nicht gespeichert werden.")
		return
	}
	thumbnailBlobID, err := a.createVehicleImageThumbnail(r.Context(), data, storageName)
	if err != nil {
		a.logger.Warn("image thumbnail skipped", "file", header.Filename, "error", err)
	}
	image, err := a.vehicleService.CreateImage(r.Context(), vehicleID, application.VehicleImageInput{
		Title:           r.FormValue("title"),
		SourceURL:       r.FormValue("sourceUrl"),
		FileName:        storageName,
		MimeType:        mimeType,
		BlobID:          blobID,
		ThumbnailBlobID: thumbnailBlobID,
		MaintenanceID:   r.FormValue("maintenanceId"),
		IsPrimary:       strings.EqualFold(r.FormValue("isPrimary"), "true"),
	})
	if err != nil {
		a.deleteFileBlobIfUnreferenced(r.Context(), blobID)
		a.deleteFileBlobIfUnreferenced(r.Context(), thumbnailBlobID)
		if errors.Is(err, application.ErrVehicleNotFound) {
			respondProblem(w, http.StatusNotFound, "vehicle_not_found", "Vehicle not found.")
			return
		}
		a.logger.Error("image metadata create failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "image_upload_failed", "Bild konnte nicht gespeichert werden.")
		return
	}
	respondJSON(w, http.StatusCreated, image)
}

type importVehicleImageInput struct {
	URL           string `json:"url"`
	Title         string `json:"title"`
	SourceURL     string `json:"sourceUrl"`
	MaintenanceID string `json:"maintenanceId"`
	IsPrimary     bool   `json:"isPrimary"`
	SortOrder     int    `json:"sortOrder"`
}

func (a *App) importVehicleImageFromURL(w http.ResponseWriter, r *http.Request) {
	var input importVehicleImageInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	input.URL = strings.TrimSpace(input.URL)
	if input.URL == "" {
		respondProblem(w, http.StatusBadRequest, "image_url_missing", "Eine Bild-URL ist erforderlich.")
		return
	}
	vehicleID := r.PathValue("id")
	localized, err := a.localizeVehicleImage(r.Context(), vehicleID, application.VehicleImageInput{
		URL:           input.URL,
		Title:         input.Title,
		SourceURL:     input.SourceURL,
		MaintenanceID: input.MaintenanceID,
		IsPrimary:     input.IsPrimary,
		SortOrder:     input.SortOrder,
	})
	if err != nil {
		a.logger.Warn("remote image import failed", "url", input.URL, "error", err)
		respondProblem(w, http.StatusBadGateway, "image_import_failed", "Bild konnte nicht heruntergeladen werden.")
		return
	}
	image, err := a.vehicleService.CreateImage(r.Context(), vehicleID, localized)
	if err != nil {
		a.deleteFileBlobIfUnreferenced(r.Context(), localized.BlobID)
		a.deleteFileBlobIfUnreferenced(r.Context(), localized.ThumbnailBlobID)
		if fullPath, pathErr := confinedDataPath(a.dataDir, localized.StoragePath); pathErr == nil {
			_ = os.Remove(fullPath)
		}
		if fullPath, pathErr := confinedDataPath(a.dataDir, localized.ThumbnailPath); pathErr == nil {
			_ = os.Remove(fullPath)
		}
		if errors.Is(err, application.ErrVehicleNotFound) {
			respondProblem(w, http.StatusNotFound, "vehicle_not_found", "Vehicle not found.")
			return
		}
		if errors.Is(err, application.ErrVehicleValidation) {
			respondProblem(w, http.StatusBadRequest, "image_invalid", "Bilddaten sind unvollst?ndig.")
			return
		}
		a.logger.Error("remote image metadata create failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "image_import_failed", "Bild konnte nicht gespeichert werden.")
		return
	}
	respondJSON(w, http.StatusCreated, image)
}

func (a *App) deleteVehicleImage(w http.ResponseWriter, r *http.Request) {
	image, err := a.vehicleService.DeleteImage(r.Context(), r.PathValue("id"), r.PathValue("imageID"))
	if err != nil {
		if errors.Is(err, application.ErrVehicleNotFound) {
			respondProblem(w, http.StatusNotFound, "image_not_found", "Image not found.")
			return
		}
		if errors.Is(err, application.ErrVehicleImageInUse) {
			respondProblem(w, http.StatusConflict, "image_in_use", "Bild ist mit einem Wartungseintrag verknüpft. Bitte zuerst die Verknüpfung entfernen.")
			return
		}
		a.logger.Error("image delete failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "image_delete_failed", "Bild konnte nicht gelöscht werden.")
		return
	}
	a.removeVehicleImageFileIfUnreferenced(r.Context(), image.StoragePath)
	a.removeVehicleImageFileIfUnreferenced(r.Context(), image.ThumbnailPath)
	a.deleteFileBlobIfUnreferenced(r.Context(), image.BlobID)
	a.deleteFileBlobIfUnreferenced(r.Context(), image.ThumbnailBlobID)
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) removeVehicleImageFileIfUnreferenced(ctx context.Context, storagePath string) {
	if storagePath == "" {
		return
	}
	references, err := a.vehicleService.ImageFileReferenceCount(ctx, storagePath)
	if err != nil {
		a.logger.Warn("image file reference check failed", "path", storagePath, "error", err)
		return
	}
	if references > 0 {
		return
	}
	if fullPath, err := confinedDataPath(a.dataDir, storagePath); err == nil {
		_ = os.Remove(fullPath)
	}
}

func (a *App) storeFileBlob(ctx context.Context, data []byte) (string, error) {
	if a.fileBlobs == nil {
		return "", errors.New("file blob service is not configured")
	}
	return a.fileBlobs.Store(ctx, data)
}

func (a *App) loadFileBlob(ctx context.Context, blobID string) ([]byte, error) {
	if a.fileBlobs == nil {
		return nil, errors.New("file blob service is not configured")
	}
	return a.fileBlobs.Load(ctx, blobID)
}

func (a *App) deleteFileBlobIfUnreferenced(ctx context.Context, blobID string) {
	if blobID == "" || a.fileBlobs == nil {
		return
	}
	if err := a.fileBlobs.DeleteIfUnreferenced(ctx, blobID); err != nil {
		a.logger.Warn("file blob cleanup failed", "blobID", blobID, "error", err)
	}
}

func serveFileBytes(w http.ResponseWriter, r *http.Request, data []byte, mimeType, disposition, fileName string) {
	if mimeType != "" {
		w.Header().Set("Content-Type", mimeType)
	}
	if fileName != "" {
		w.Header().Set("Content-Disposition", mime.FormatMediaType(disposition, map[string]string{"filename": cleanOriginalFileName(fileName)}))
	}
	http.ServeContent(w, r, fileName, time.Now().UTC(), bytes.NewReader(data))
}

func (a *App) downloadVehicleImage(w http.ResponseWriter, r *http.Request) {
	image, err := a.vehicleService.GetImage(r.Context(), r.PathValue("id"), r.PathValue("imageID"))
	if err != nil {
		if errors.Is(err, application.ErrVehicleNotFound) {
			respondProblem(w, http.StatusNotFound, "image_not_found", "Image not found.")
			return
		}
		a.logger.Error("image lookup failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "image_download_failed", "Bild konnte nicht geladen werden.")
		return
	}
	if image.BlobID != "" {
		data, err := a.loadFileBlob(r.Context(), image.BlobID)
		if err != nil {
			a.logger.Error("image blob load failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "image_download_failed", "Bild konnte nicht geladen werden.")
			return
		}
		serveFileBytes(w, r, data, image.MimeType, "inline", path.Base(image.FileName))
		return
	}
	if image.StoragePath == "" {
		respondProblem(w, http.StatusNotFound, "image_file_missing", "Bilddatei ist nicht lokal gespeichert.")
		return
	}
	fullPath, err := confinedDataPath(a.dataDir, image.StoragePath)
	if err != nil {
		respondProblem(w, http.StatusInternalServerError, "image_path_invalid", "Bild konnte nicht geladen werden.")
		return
	}
	if image.MimeType != "" {
		w.Header().Set("Content-Type", image.MimeType)
	}
	w.Header().Set("Content-Disposition", mime.FormatMediaType("inline", map[string]string{"filename": path.Base(image.FileName)}))
	http.ServeFile(w, r, fullPath)
}

func (a *App) downloadVehicleImageThumbnail(w http.ResponseWriter, r *http.Request) {
	image, err := a.vehicleService.GetImage(r.Context(), r.PathValue("id"), r.PathValue("imageID"))
	if err != nil {
		if errors.Is(err, application.ErrVehicleNotFound) {
			respondProblem(w, http.StatusNotFound, "image_not_found", "Image not found.")
			return
		}
		a.logger.Error("image thumbnail lookup failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "image_thumbnail_failed", "Bildvorschau konnte nicht geladen werden.")
		return
	}
	if image.ThumbnailBlobID != "" {
		data, err := a.loadFileBlob(r.Context(), image.ThumbnailBlobID)
		if err != nil {
			a.logger.Error("image thumbnail blob load failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "image_thumbnail_failed", "Bildvorschau konnte nicht geladen werden.")
			return
		}
		serveFileBytes(w, r, data, "image/jpeg", "inline", strings.TrimSuffix(path.Base(image.FileName), path.Ext(image.FileName))+"-thumb.jpg")
		return
	}
	if image.ThumbnailPath == "" {
		a.downloadVehicleImage(w, r)
		return
	}
	fullPath, err := confinedDataPath(a.dataDir, image.ThumbnailPath)
	if err != nil {
		respondProblem(w, http.StatusInternalServerError, "image_path_invalid", "Bildvorschau konnte nicht geladen werden.")
		return
	}
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Disposition", mime.FormatMediaType("inline", map[string]string{"filename": strings.TrimSuffix(path.Base(image.FileName), path.Ext(image.FileName)) + "-thumb.jpg"}))
	http.ServeFile(w, r, fullPath)
}

func (a *App) createVehicleImageThumbnail(ctx context.Context, data []byte, storageName string) (string, error) {
	src, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	thumb := scaleImageToFit(src, 360, 240)
	var out bytes.Buffer
	if err = jpeg.Encode(&out, thumb, &jpeg.Options{Quality: 82}); err != nil {
		return "", err
	}
	return a.storeFileBlob(ctx, out.Bytes())
}

func scaleImageToFit(src image.Image, maxWidth, maxHeight int) image.Image {
	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if width <= 0 || height <= 0 || (width <= maxWidth && height <= maxHeight) {
		return src
	}
	ratioW := float64(maxWidth) / float64(width)
	ratioH := float64(maxHeight) / float64(height)
	ratio := ratioW
	if ratioH < ratio {
		ratio = ratioH
	}
	dstWidth := max(1, int(float64(width)*ratio))
	dstHeight := max(1, int(float64(height)*ratio))
	dst := image.NewRGBA(image.Rect(0, 0, dstWidth, dstHeight))
	for y := range dstHeight {
		srcY := bounds.Min.Y + y*height/dstHeight
		for x := range dstWidth {
			srcX := bounds.Min.X + x*width/dstWidth
			dst.Set(x, y, src.At(srcX, srcY))
		}
	}
	return dst
}

type importVehicleAttachmentInput struct {
	URL           string `json:"url"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	Category      string `json:"category"`
	MaintenanceID string `json:"maintenanceId"`
}

func (a *App) uploadVehicleAttachment(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, a.maxAttachmentBytes+1024*1024)
	if err := r.ParseMultipartForm(a.maxAttachmentBytes); err != nil {
		respondProblem(w, http.StatusBadRequest, "attachment_upload_invalid", "Beilage konnte nicht gelesen werden.")
		return
	}
	if r.MultipartForm != nil {
		defer func() { _ = r.MultipartForm.RemoveAll() }()
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		respondProblem(w, http.StatusBadRequest, "attachment_missing", "Eine Datei ist erforderlich.")
		return
	}
	defer func() { _ = file.Close() }()
	originalName := cleanOriginalFileName(header.Filename)
	if header.Size > a.maxAttachmentBytes {
		respondProblem(w, http.StatusBadRequest, "attachment_too_large", "Die Datei ist zu gro?.")
		return
	}
	if isBlockedAttachmentName(originalName) {
		respondProblem(w, http.StatusBadRequest, "attachment_type_blocked", "Ausführbare Dateien sind nicht erlaubt.")
		return
	}
	data, err := io.ReadAll(io.LimitReader(file, a.maxAttachmentBytes+1))
	if err != nil || int64(len(data)) > a.maxAttachmentBytes {
		respondProblem(w, http.StatusBadRequest, "attachment_too_large", "Die Datei ist zu gro?.")
		return
	}
	if len(data) == 0 {
		respondProblem(w, http.StatusBadRequest, "attachment_empty", "Leere Dateien sind nicht erlaubt.")
		return
	}
	mimeType := http.DetectContentType(data)
	if isBlockedAttachmentMime(mimeType) {
		respondProblem(w, http.StatusBadRequest, "attachment_type_blocked", "Ausführbare Dateien sind nicht erlaubt.")
		return
	}
	if !a.isAllowedAttachmentUpload(originalName, mimeType) {
		respondProblem(w, http.StatusBadRequest, "attachment_type_blocked", "Erlaubt sind PDF, TXT, CSV, JSON, XML, ZIP sowie JPG, PNG und WebP.")
		return
	}
	vehicleID := r.PathValue("id")
	storageName := fmt.Sprintf("%d-%s", time.Now().UTC().UnixNano(), safeAttachmentFileName(originalName))
	blobID, err := a.storeFileBlob(r.Context(), data)
	if err != nil {
		a.logger.Error("attachment blob write failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "attachment_upload_failed", "Beilage konnte nicht gespeichert werden.")
		return
	}
	attachment, err := a.vehicleService.CreateAttachment(r.Context(), vehicleID, application.VehicleAttachmentInput{
		FileName:      storageName,
		OriginalName:  originalName,
		Description:   r.FormValue("description"),
		Category:      r.FormValue("category"),
		MimeType:      mimeType,
		SizeBytes:     int64(len(data)),
		BlobID:        blobID,
		MaintenanceID: r.FormValue("maintenanceId"),
	})
	if err != nil {
		a.deleteFileBlobIfUnreferenced(r.Context(), blobID)
		if errors.Is(err, application.ErrVehicleNotFound) {
			respondProblem(w, http.StatusNotFound, "vehicle_not_found", "Vehicle not found.")
			return
		}
		a.logger.Error("attachment metadata create failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "attachment_upload_failed", "Beilage konnte nicht gespeichert werden.")
		return
	}
	respondJSON(w, http.StatusCreated, attachment)
}

func (a *App) importVehicleAttachmentFromURL(w http.ResponseWriter, r *http.Request) {
	var input importVehicleAttachmentInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	input.URL = strings.TrimSpace(input.URL)
	if input.URL == "" || !isPublicImageURL(r.Context(), input.URL) {
		respondProblem(w, http.StatusBadRequest, "attachment_url_invalid", "Dokument-URL ist nicht erreichbar oder nicht erlaubt.")
		return
	}
	requestCtx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(requestCtx, http.MethodGet, input.URL, nil)
	if err != nil {
		respondProblem(w, http.StatusBadRequest, "attachment_url_invalid", "Dokument-URL ist ung?ltig.")
		return
	}
	req.Header.Set("User-Agent", "RailKeeper/0.1 document-fetch")
	req.Header.Set("Accept", "application/pdf,text/plain,application/json,application/xml,text/xml,application/zip,image/*;q=0.8,*/*;q=0.4")
	client := remoteDocumentHTTPClient(r.Context())
	resp, err := client.Do(req)
	if err != nil {
		a.logger.Warn("remote attachment fetch failed", "url", input.URL, "error", err)
		respondProblem(w, http.StatusBadGateway, "attachment_fetch_failed", "Dokument konnte nicht heruntergeladen werden.")
		return
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		respondProblem(w, http.StatusBadGateway, "attachment_fetch_failed", "Dokument konnte nicht heruntergeladen werden.")
		return
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, a.maxAttachmentBytes+1))
	if err != nil || int64(len(data)) > a.maxAttachmentBytes {
		respondProblem(w, http.StatusBadRequest, "attachment_too_large", "Die Datei ist zu gro?.")
		return
	}
	if len(data) == 0 {
		respondProblem(w, http.StatusBadRequest, "attachment_empty", "Leere Dateien sind nicht erlaubt.")
		return
	}
	mimeType := http.DetectContentType(data)
	if headerMime := strings.TrimSpace(resp.Header.Get("Content-Type")); headerMime != "" && (strings.Contains(headerMime, "pdf") || strings.Contains(headerMime, "zip") || strings.Contains(headerMime, "xml")) {
		mimeType = strings.Split(headerMime, ";")[0]
	}
	originalName := remoteAttachmentFileName(input, input.URL, mimeType)
	if isBlockedAttachmentName(originalName) || isBlockedAttachmentMime(mimeType) || !a.isAllowedAttachmentUpload(originalName, mimeType) {
		respondProblem(w, http.StatusBadRequest, "attachment_type_blocked", "Erlaubt sind PDF, TXT, CSV, JSON, XML, ZIP sowie JPG, PNG und WebP.")
		return
	}
	vehicleID := r.PathValue("id")
	storageName := fmt.Sprintf("%d-%s", time.Now().UTC().UnixNano(), safeAttachmentFileName(originalName))
	blobID, err := a.storeFileBlob(r.Context(), data)
	if err != nil {
		a.logger.Error("remote attachment blob write failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "attachment_import_failed", "Dokument konnte nicht gespeichert werden.")
		return
	}
	category := strings.TrimSpace(input.Category)
	if category == "" {
		category = attachmentCategoryForRemoteDocument(originalName, input.Title)
	}
	attachment, err := a.vehicleService.CreateAttachment(r.Context(), vehicleID, application.VehicleAttachmentInput{
		FileName:      storageName,
		OriginalName:  originalName,
		Description:   strings.TrimSpace(input.Description),
		Category:      category,
		MimeType:      mimeType,
		SizeBytes:     int64(len(data)),
		BlobID:        blobID,
		MaintenanceID: strings.TrimSpace(input.MaintenanceID),
	})
	if err != nil {
		a.deleteFileBlobIfUnreferenced(r.Context(), blobID)
		if errors.Is(err, application.ErrVehicleNotFound) {
			respondProblem(w, http.StatusNotFound, "vehicle_not_found", "Vehicle not found.")
			return
		}
		if errors.Is(err, application.ErrVehicleValidation) {
			respondProblem(w, http.StatusBadRequest, "attachment_invalid", "Beilage ist unvollst?ndig.")
			return
		}
		a.logger.Error("remote attachment metadata create failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "attachment_import_failed", "Dokument konnte nicht gespeichert werden.")
		return
	}
	respondJSON(w, http.StatusCreated, attachment)
}

func (a *App) updateVehicleAttachment(w http.ResponseWriter, r *http.Request) {
	var input application.VehicleAttachmentUpdateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	attachment, err := a.vehicleService.UpdateAttachment(r.Context(), r.PathValue("id"), r.PathValue("attachmentID"), input)
	if err != nil {
		if errors.Is(err, application.ErrVehicleNotFound) {
			respondProblem(w, http.StatusNotFound, "attachment_not_found", "Attachment not found.")
			return
		}
		a.logger.Error("attachment update failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "attachment_update_failed", "Beilage konnte nicht aktualisiert werden.")
		return
	}
	respondJSON(w, http.StatusOK, attachment)
}

func (a *App) deleteVehicleAttachment(w http.ResponseWriter, r *http.Request) {
	attachment, err := a.vehicleService.DeleteAttachment(r.Context(), r.PathValue("id"), r.PathValue("attachmentID"))
	if err != nil {
		if errors.Is(err, application.ErrVehicleNotFound) {
			respondProblem(w, http.StatusNotFound, "attachment_not_found", "Attachment not found.")
			return
		}
		a.logger.Error("attachment delete failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "attachment_delete_failed", "Beilage konnte nicht gelöscht werden.")
		return
	}
	if fullPath, err := confinedDataPath(a.dataDir, attachment.StoragePath); err == nil {
		_ = os.Remove(fullPath)
	}
	a.deleteFileBlobIfUnreferenced(r.Context(), attachment.BlobID)
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) downloadVehicleAttachment(w http.ResponseWriter, r *http.Request) {
	attachment, err := a.vehicleService.GetAttachment(r.Context(), r.PathValue("id"), r.PathValue("attachmentID"))
	if err != nil {
		if errors.Is(err, application.ErrVehicleNotFound) {
			respondProblem(w, http.StatusNotFound, "attachment_not_found", "Attachment not found.")
			return
		}
		a.logger.Error("attachment download lookup failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "attachment_download_failed", "Beilage konnte nicht geladen werden.")
		return
	}
	if attachment.MimeType != "" {
		w.Header().Set("Content-Type", attachment.MimeType)
	}
	disposition := "attachment"
	inlinePreview := r.URL.Query().Get("inline") == "true" && canPreviewAttachmentInline(attachment.MimeType, attachment.OriginalName)
	if inlinePreview {
		disposition = "inline"
		if shouldSandboxInlineAttachment(attachment.MimeType, attachment.OriginalName) {
			w.Header().Set("Content-Security-Policy", "sandbox; default-src 'none'; img-src 'self' data: blob:; style-src 'unsafe-inline'")
		}
	}
	w.Header().Set("Content-Disposition", mime.FormatMediaType(disposition, map[string]string{"filename": cleanOriginalFileName(attachment.OriginalName)}))
	if attachment.BlobID != "" {
		data, err := a.loadFileBlob(r.Context(), attachment.BlobID)
		if err != nil {
			a.logger.Error("attachment blob load failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "attachment_download_failed", "Beilage konnte nicht geladen werden.")
			return
		}
		http.ServeContent(w, r, attachment.FileName, time.Now().UTC(), bytes.NewReader(data))
		return
	}
	fullPath, err := confinedDataPath(a.dataDir, attachment.StoragePath)
	if err != nil {
		respondProblem(w, http.StatusInternalServerError, "attachment_path_invalid", "Beilage konnte nicht geladen werden.")
		return
	}
	http.ServeFile(w, r, fullPath)
}

func canPreviewAttachmentInline(mimeType, fileName string) bool {
	mimeType = strings.ToLower(strings.TrimSpace(mimeType))
	fileName = strings.ToLower(strings.TrimSpace(fileName))
	if strings.Contains(mimeType, "pdf") || strings.HasPrefix(mimeType, "image/") || strings.HasPrefix(mimeType, "text/") {
		return true
	}
	if strings.Contains(mimeType, "json") || strings.Contains(mimeType, "xml") {
		return true
	}
	switch path.Ext(fileName) {
	case ".pdf", ".jpg", ".jpeg", ".png", ".webp", ".txt", ".csv", ".json", ".xml", ".html", ".htm":
		return true
	default:
		return false
	}
}

func shouldSandboxInlineAttachment(mimeType, fileName string) bool {
	mimeType = strings.ToLower(strings.TrimSpace(mimeType))
	fileName = strings.ToLower(strings.TrimSpace(fileName))
	if strings.HasPrefix(mimeType, "text/html") || strings.HasSuffix(fileName, ".html") || strings.HasSuffix(fileName, ".htm") {
		return true
	}
	return false
}

func (a *App) readAttachmentData(ctx context.Context, attachment application.VehicleAttachment, maxBytes int64) ([]byte, error) {
	if attachment.BlobID != "" {
		data, err := a.loadFileBlob(ctx, attachment.BlobID)
		if err != nil {
			return nil, err
		}
		if maxBytes > 0 && int64(len(data)) > maxBytes {
			return nil, fmt.Errorf("attachment blob exceeds read limit")
		}
		return data, nil
	}
	fullPath, err := confinedDataPath(a.dataDir, attachment.StoragePath)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()
	data, err := io.ReadAll(io.LimitReader(file, maxBytes+1))
	if err != nil {
		return nil, err
	}
	if maxBytes > 0 && int64(len(data)) > maxBytes {
		return nil, fmt.Errorf("attachment file exceeds read limit")
	}
	return data, nil
}

func (a *App) listVehicleMaintenance(w http.ResponseWriter, r *http.Request) {
	entries, err := a.vehicleService.ListMaintenance(r.Context(), r.PathValue("id"))
	if err != nil {
		if errors.Is(err, application.ErrVehicleNotFound) {
			respondProblem(w, http.StatusNotFound, "vehicle_not_found", "Vehicle not found.")
			return
		}
		a.logger.Error("maintenance list failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "maintenance_list_failed", "Wartungseintraege konnten nicht geladen werden.")
		return
	}
	respondJSON(w, http.StatusOK, entries)
}

func (a *App) createVehicleMaintenance(w http.ResponseWriter, r *http.Request) {
	var input application.VehicleMaintenanceInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	entry, err := a.vehicleService.CreateMaintenance(r.Context(), r.PathValue("id"), input)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrVehicleValidation):
			respondProblem(w, http.StatusBadRequest, "maintenance_invalid", "Wartungseintrag ist unvollst?ndig.")
		case errors.Is(err, application.ErrVehicleNotFound):
			respondProblem(w, http.StatusNotFound, "vehicle_not_found", "Vehicle not found.")
		default:
			a.logger.Error("maintenance create failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "maintenance_create_failed", "Wartungseintrag konnte nicht gespeichert werden.")
		}
		return
	}
	respondJSON(w, http.StatusCreated, entry)
}

func (a *App) updateVehicleMaintenance(w http.ResponseWriter, r *http.Request) {
	var input application.VehicleMaintenanceInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	entry, err := a.vehicleService.UpdateMaintenance(r.Context(), r.PathValue("id"), r.PathValue("maintenanceID"), input)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrVehicleValidation):
			respondProblem(w, http.StatusBadRequest, "maintenance_invalid", "Wartungseintrag ist unvollst?ndig.")
		case errors.Is(err, application.ErrVehicleNotFound):
			respondProblem(w, http.StatusNotFound, "maintenance_not_found", "Maintenance entry not found.")
		default:
			a.logger.Error("maintenance update failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "maintenance_update_failed", "Wartungseintrag konnte nicht aktualisiert werden.")
		}
		return
	}
	respondJSON(w, http.StatusOK, entry)
}

func (a *App) deleteVehicleMaintenance(w http.ResponseWriter, r *http.Request) {
	if _, err := a.vehicleService.DeleteMaintenance(r.Context(), r.PathValue("id"), r.PathValue("maintenanceID")); err != nil {
		if errors.Is(err, application.ErrVehicleNotFound) {
			respondProblem(w, http.StatusNotFound, "maintenance_not_found", "Maintenance entry not found.")
			return
		}
		a.logger.Error("maintenance delete failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "maintenance_delete_failed", "Wartungseintrag konnte nicht gelöscht werden.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) listVehicleSpareParts(w http.ResponseWriter, r *http.Request) {
	entries, err := a.vehicleService.ListSpareParts(r.Context(), r.PathValue("id"))
	if err != nil {
		if errors.Is(err, application.ErrVehicleNotFound) {
			respondProblem(w, http.StatusNotFound, "vehicle_not_found", "Vehicle not found.")
			return
		}
		a.logger.Error("spare part list failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "spare_part_list_failed", "Ersatzteile konnten nicht geladen werden.")
		return
	}
	respondJSON(w, http.StatusOK, entries)
}

func (a *App) suggestVehicleSpareParts(w http.ResponseWriter, r *http.Request) {
	vehicle, err := a.vehicleService.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		if errors.Is(err, application.ErrVehicleNotFound) {
			respondProblem(w, http.StatusNotFound, "vehicle_not_found", "Vehicle not found.")
			return
		}
		a.logger.Error("spare part suggestion vehicle lookup failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "spare_part_suggestions_failed", "Ersatzteilvorschlaege konnten nicht geladen werden.")
		return
	}
	articleNumber := strings.TrimSpace(vehicle.ArticleNumber)
	if articleNumber == "" {
		respondJSON(w, http.StatusOK, []application.ArticleSearchSparePart{})
		return
	}
	seen := map[string]bool{}
	suggestions := []application.ArticleSearchSparePart{}
	attachmentID := strings.TrimSpace(r.URL.Query().Get("attachmentId"))
	for _, attachment := range vehicle.Attachments {
		if attachmentID != "" && attachment.ID != attachmentID {
			continue
		}
		if len(suggestions) >= 80 || !looksLikeSparePartAttachment(attachment) {
			continue
		}
		data, err := a.readAttachmentData(r.Context(), attachment, 12*1024*1024)
		if err != nil || len(data) == 0 {
			continue
		}
		downloadURL := "/api/v1/vehicles/" + url.PathEscape(vehicle.ID) + "/attachments/" + url.PathEscape(attachment.ID) + "/download"
		parts := application.ArticleSparePartsFromDocumentData(data, articleNumber, downloadURL)
		for _, part := range parts {
			part.Source = cleanOriginalFileName(attachment.OriginalName)
			if part.URL == "" {
				part.URL = downloadURL
			}
			key := strings.ToLower(part.ArticleNumber + "|" + part.Description + "|" + part.URL)
			if key == "||" || seen[key] {
				continue
			}
			seen[key] = true
			suggestions = append(suggestions, part)
			if len(suggestions) >= 80 {
				break
			}
		}
	}
	respondJSON(w, http.StatusOK, suggestions)
}

func looksLikeSparePartAttachment(attachment application.VehicleAttachment) bool {
	lower := strings.ToLower(attachment.Category + " " + attachment.OriginalName + " " + attachment.Description + " " + attachment.MimeType)
	return strings.Contains(lower, "ersatzteil") ||
		strings.Contains(lower, "spare") ||
		strings.Contains(lower, "et-blatt") ||
		strings.Contains(lower, "serviceblatt") ||
		strings.Contains(lower, "bedienungsanl") ||
		strings.Contains(lower, "manual") ||
		strings.Contains(lower, "pdf")
}

func (a *App) createVehicleSparePart(w http.ResponseWriter, r *http.Request) {
	var input application.VehicleSparePartInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	entry, err := a.vehicleService.CreateSparePart(r.Context(), r.PathValue("id"), input)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrVehicleValidation):
			respondProblem(w, http.StatusBadRequest, "spare_part_invalid", "Ersatzteil ist unvollst?ndig.")
		case errors.Is(err, application.ErrVehicleNotFound):
			respondProblem(w, http.StatusNotFound, "vehicle_not_found", "Vehicle not found.")
		default:
			a.logger.Error("spare part create failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "spare_part_create_failed", "Ersatzteil konnte nicht gespeichert werden.")
		}
		return
	}
	respondJSON(w, http.StatusCreated, entry)
}

func (a *App) updateVehicleSparePart(w http.ResponseWriter, r *http.Request) {
	var input application.VehicleSparePartInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	entry, err := a.vehicleService.UpdateSparePart(r.Context(), r.PathValue("id"), r.PathValue("sparePartID"), input)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrVehicleValidation):
			respondProblem(w, http.StatusBadRequest, "spare_part_invalid", "Ersatzteil ist unvollst?ndig.")
		case errors.Is(err, application.ErrVehicleNotFound):
			respondProblem(w, http.StatusNotFound, "spare_part_not_found", "Spare part not found.")
		default:
			a.logger.Error("spare part update failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "spare_part_update_failed", "Ersatzteil konnte nicht aktualisiert werden.")
		}
		return
	}
	respondJSON(w, http.StatusOK, entry)
}

func (a *App) deleteVehicleSparePart(w http.ResponseWriter, r *http.Request) {
	if _, err := a.vehicleService.DeleteSparePart(r.Context(), r.PathValue("id"), r.PathValue("sparePartID")); err != nil {
		if errors.Is(err, application.ErrVehicleNotFound) {
			respondProblem(w, http.StatusNotFound, "spare_part_not_found", "Spare part not found.")
			return
		}
		a.logger.Error("spare part delete failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "spare_part_delete_failed", "Ersatzteil konnte nicht gel?scht werden.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) listVehicleFunctions(w http.ResponseWriter, r *http.Request) {
	functions, err := a.vehicleService.ListFunctions(r.Context(), r.PathValue("id"))
	if err != nil {
		if errors.Is(err, application.ErrVehicleNotFound) {
			respondProblem(w, http.StatusNotFound, "vehicle_not_found", "Vehicle not found.")
			return
		}
		a.logger.Error("function list failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "function_list_failed", "Digitalfunktionen konnten nicht geladen werden.")
		return
	}
	respondJSON(w, http.StatusOK, functions)
}

func (a *App) upsertVehicleFunction(w http.ResponseWriter, r *http.Request) {
	var input application.VehicleFunctionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	function, err := a.vehicleService.UpsertFunction(r.Context(), r.PathValue("id"), r.PathValue("functionKey"), input)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrVehicleValidation):
			respondProblem(w, http.StatusBadRequest, "function_invalid", "Digitalfunktion ist ungültig.")
		case errors.Is(err, application.ErrVehicleNotFound):
			respondProblem(w, http.StatusNotFound, "vehicle_not_found", "Vehicle not found.")
		default:
			a.logger.Error("function save failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "function_save_failed", "Digitalfunktion konnte nicht gespeichert werden.")
		}
		return
	}
	respondJSON(w, http.StatusOK, function)
}

func (a *App) deleteVehicleFunction(w http.ResponseWriter, r *http.Request) {
	if _, err := a.vehicleService.DeleteFunction(r.Context(), r.PathValue("id"), r.PathValue("functionKey")); err != nil {
		if errors.Is(err, application.ErrVehicleNotFound) {
			respondProblem(w, http.StatusNotFound, "function_not_found", "Function entry not found.")
			return
		}
		a.logger.Error("function delete failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "function_delete_failed", "Digitalfunktion konnte nicht gelöscht werden.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) listVehicleCVValues(w http.ResponseWriter, r *http.Request) {
	values, err := a.vehicleService.ListCVValues(r.Context(), r.PathValue("id"))
	if err != nil {
		if errors.Is(err, application.ErrVehicleNotFound) {
			respondProblem(w, http.StatusNotFound, "vehicle_not_found", "Vehicle not found.")
			return
		}
		a.logger.Error("cv value list failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "cv_value_list_failed", "CV-Werte konnten nicht geladen werden.")
		return
	}
	respondJSON(w, http.StatusOK, values)
}

func (a *App) createVehicleCVValue(w http.ResponseWriter, r *http.Request) {
	var input application.VehicleCVValueInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	value, err := a.vehicleService.CreateCVValue(r.Context(), r.PathValue("id"), input)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrVehicleValidation):
			respondProblem(w, http.StatusBadRequest, "cv_value_invalid", "CV-Nummer muss 1-1024 und Wert 0-255 sein.")
		case errors.Is(err, application.ErrVehicleNotFound):
			respondProblem(w, http.StatusNotFound, "vehicle_not_found", "Vehicle not found.")
		default:
			a.logger.Error("cv value create failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "cv_value_create_failed", "CV-Wert konnte nicht gespeichert werden.")
		}
		return
	}
	respondJSON(w, http.StatusCreated, value)
}

func (a *App) updateVehicleCVValue(w http.ResponseWriter, r *http.Request) {
	var input application.VehicleCVValueInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	value, err := a.vehicleService.UpdateCVValue(r.Context(), r.PathValue("id"), r.PathValue("cvValueID"), input)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrVehicleValidation):
			respondProblem(w, http.StatusBadRequest, "cv_value_invalid", "CV-Nummer muss 1-1024 und Wert 0-255 sein.")
		case errors.Is(err, application.ErrVehicleNotFound):
			respondProblem(w, http.StatusNotFound, "cv_value_not_found", "CV value not found.")
		default:
			a.logger.Error("cv value update failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "cv_value_update_failed", "CV-Wert konnte nicht aktualisiert werden.")
		}
		return
	}
	respondJSON(w, http.StatusOK, value)
}

func (a *App) deleteVehicleCVValue(w http.ResponseWriter, r *http.Request) {
	if _, err := a.vehicleService.DeleteCVValue(r.Context(), r.PathValue("id"), r.PathValue("cvValueID")); err != nil {
		if errors.Is(err, application.ErrVehicleNotFound) {
			respondProblem(w, http.StatusNotFound, "cv_value_not_found", "CV value not found.")
			return
		}
		a.logger.Error("cv value delete failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "cv_value_delete_failed", "CV-Wert konnte nicht gelöscht werden.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) previewVehicleCVFile(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, a.maxAttachmentBytes+1024*1024)
	if err := r.ParseMultipartForm(a.maxAttachmentBytes); err != nil {
		respondProblem(w, http.StatusBadRequest, "cv_file_preview_invalid", "CV-Datei konnte nicht gelesen werden.")
		return
	}
	if r.MultipartForm != nil {
		defer func() { _ = r.MultipartForm.RemoveAll() }()
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		respondProblem(w, http.StatusBadRequest, "cv_file_missing", "Eine Datei ist erforderlich.")
		return
	}
	defer func() { _ = file.Close() }()
	originalName := cleanOriginalFileName(header.Filename)
	if header.Size > a.maxAttachmentBytes || isBlockedAttachmentName(originalName) {
		respondProblem(w, http.StatusBadRequest, "cv_file_blocked", "Diese CV-Datei ist nicht erlaubt.")
		return
	}
	data, err := io.ReadAll(io.LimitReader(file, a.maxAttachmentBytes+1))
	if err != nil || int64(len(data)) > a.maxAttachmentBytes {
		respondProblem(w, http.StatusBadRequest, "cv_file_too_large", "Die Datei ist zu groß.")
		return
	}
	if len(data) == 0 {
		respondProblem(w, http.StatusBadRequest, "cv_file_empty", "Leere Dateien sind nicht erlaubt.")
		return
	}
	mimeType := http.DetectContentType(data)
	if isBlockedAttachmentMime(mimeType) {
		respondProblem(w, http.StatusBadRequest, "cv_file_blocked", "Diese CV-Datei ist nicht erlaubt.")
		return
	}
	respondJSON(w, http.StatusOK, esuxPreviewResponse(originalName, int64(len(data)), mimeType, data))
}

func (a *App) uploadVehicleCVFile(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, a.maxAttachmentBytes+1024*1024)
	if err := r.ParseMultipartForm(a.maxAttachmentBytes); err != nil {
		respondProblem(w, http.StatusBadRequest, "cv_file_upload_invalid", "CV-Datei konnte nicht gelesen werden.")
		return
	}
	if r.MultipartForm != nil {
		defer func() { _ = r.MultipartForm.RemoveAll() }()
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		respondProblem(w, http.StatusBadRequest, "cv_file_missing", "Eine Datei ist erforderlich.")
		return
	}
	defer func() { _ = file.Close() }()
	originalName := cleanOriginalFileName(header.Filename)
	if header.Size > a.maxAttachmentBytes || isBlockedAttachmentName(originalName) {
		respondProblem(w, http.StatusBadRequest, "cv_file_blocked", "Diese CV-Datei ist nicht erlaubt.")
		return
	}
	data, err := io.ReadAll(io.LimitReader(file, a.maxAttachmentBytes+1))
	if err != nil || int64(len(data)) > a.maxAttachmentBytes {
		respondProblem(w, http.StatusBadRequest, "cv_file_too_large", "Die Datei ist zu gro?.")
		return
	}
	if len(data) == 0 {
		respondProblem(w, http.StatusBadRequest, "cv_file_empty", "Leere Dateien sind nicht erlaubt.")
		return
	}
	mimeType := http.DetectContentType(data)
	if isBlockedAttachmentMime(mimeType) {
		respondProblem(w, http.StatusBadRequest, "cv_file_blocked", "Diese CV-Datei ist nicht erlaubt.")
		return
	}
	vehicleID := r.PathValue("id")
	storageName := fmt.Sprintf("%d-%s", time.Now().UTC().UnixNano(), safeAttachmentFileName(originalName))
	blobID, err := a.storeFileBlob(r.Context(), data)
	if err != nil {
		a.logger.Error("cv file blob write failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "cv_file_upload_failed", "CV-Datei konnte nicht gespeichert werden.")
		return
	}
	decoderProfile, description := applyESUXMetadata(originalName, data, r.FormValue("decoderProfile"), r.FormValue("description"))
	cvFile, err := a.vehicleService.CreateCVFile(r.Context(), vehicleID, application.VehicleCVFileInput{
		FileName:       storageName,
		OriginalName:   originalName,
		Description:    description,
		DecoderProfile: decoderProfile,
		MimeType:       mimeType,
		SizeBytes:      int64(len(data)),
		BlobID:         blobID,
	})
	if err != nil {
		a.deleteFileBlobIfUnreferenced(r.Context(), blobID)
		if errors.Is(err, application.ErrVehicleNotFound) {
			respondProblem(w, http.StatusNotFound, "vehicle_not_found", "Vehicle not found.")
			return
		}
		a.logger.Error("cv file metadata create failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "cv_file_upload_failed", "CV-Datei konnte nicht gespeichert werden.")
		return
	}
	respondJSON(w, http.StatusCreated, cvFile)
}

func (a *App) deleteVehicleCVFile(w http.ResponseWriter, r *http.Request) {
	file, err := a.vehicleService.DeleteCVFile(r.Context(), r.PathValue("id"), r.PathValue("cvFileID"))
	if err != nil {
		if errors.Is(err, application.ErrVehicleNotFound) {
			respondProblem(w, http.StatusNotFound, "cv_file_not_found", "CV file not found.")
			return
		}
		a.logger.Error("cv file delete failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "cv_file_delete_failed", "CV-Datei konnte nicht gelöscht werden.")
		return
	}
	if fullPath, err := confinedDataPath(a.dataDir, file.StoragePath); err == nil {
		_ = os.Remove(fullPath)
	}
	a.deleteFileBlobIfUnreferenced(r.Context(), file.BlobID)
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) downloadVehicleCVFile(w http.ResponseWriter, r *http.Request) {
	file, err := a.vehicleService.GetCVFile(r.Context(), r.PathValue("id"), r.PathValue("cvFileID"))
	if err != nil {
		if errors.Is(err, application.ErrVehicleNotFound) {
			respondProblem(w, http.StatusNotFound, "cv_file_not_found", "CV file not found.")
			return
		}
		a.logger.Error("cv file download lookup failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "cv_file_download_failed", "CV-Datei konnte nicht geladen werden.")
		return
	}
	if file.MimeType != "" {
		w.Header().Set("Content-Type", file.MimeType)
	}
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": cleanOriginalFileName(file.OriginalName)}))
	if file.BlobID != "" {
		data, err := a.loadFileBlob(r.Context(), file.BlobID)
		if err != nil {
			a.logger.Error("cv file blob load failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "cv_file_download_failed", "CV-Datei konnte nicht geladen werden.")
			return
		}
		http.ServeContent(w, r, file.FileName, time.Now().UTC(), bytes.NewReader(data))
		return
	}
	fullPath, err := confinedDataPath(a.dataDir, file.StoragePath)
	if err != nil {
		respondProblem(w, http.StatusInternalServerError, "cv_file_path_invalid", "CV-Datei konnte nicht geladen werden.")
		return
	}
	http.ServeFile(w, r, fullPath)
}

const maxBackupBytes = 250 * 1024 * 1024
const maxMasterDataImportBytes = 25 * 1024 * 1024

func (a *App) exportBackup(w http.ResponseWriter, r *http.Request) {
	backup, err := a.backupService.Export(r.Context())
	if err != nil {
		a.logger.Error("backup export failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "backup_export_failed", "Backup konnte nicht erstellt werden.")
		return
	}

	filename := "railkeeper-backup-" + time.Now().UTC().Format("20060102-150405") + ".json"
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": filename}))
	if err := json.NewEncoder(w).Encode(backup); err != nil {
		a.logger.Error("backup encode failed", "error", err)
	}
}

func (a *App) restoreBackup(w http.ResponseWriter, r *http.Request) {
	backup, ok := a.readBackupUpload(w, r)
	if !ok {
		return
	}

	result, err := a.backupService.Import(r.Context(), backup)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrBackupInvalid), errors.Is(err, application.ErrBackupPath):
			respondProblem(w, http.StatusBadRequest, "backup_restore_invalid", "Backup-Datei ist ungültig.")
		default:
			a.logger.Error("backup restore failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "backup_restore_failed", "Backup konnte nicht wiederhergestellt werden.")
		}
		return
	}
	if a.fileBlobs != nil {
		if err := a.fileBlobs.MigrateFilesystemBlobs(r.Context()); err != nil {
			a.logger.Error("backup restore blob migration failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "backup_restore_failed", "Backup konnte nicht wiederhergestellt werden.")
			return
		}
	}
	if a.masterDataService != nil {
		if err := a.masterDataService.WarmCache(r.Context()); err != nil {
			a.logger.Error("master data cache refresh after backup restore failed", "error", err)
		}
	}
	respondJSON(w, http.StatusOK, result)
}

func (a *App) validateBackup(w http.ResponseWriter, r *http.Request) {
	backup, ok := a.readBackupUpload(w, r)
	if !ok {
		return
	}
	result, err := a.backupService.Validate(r.Context(), backup)
	if err != nil {
		a.logger.Error("backup validation failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "backup_validation_failed", "Backup konnte nicht geprüft werden.")
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func (a *App) readBackupUpload(w http.ResponseWriter, r *http.Request) (*application.BackupDocument, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBackupBytes+1024*1024)
	if err := r.ParseMultipartForm(maxBackupBytes); err != nil {
		respondProblem(w, http.StatusBadRequest, "backup_restore_invalid", "Backup-Datei konnte nicht gelesen werden.")
		return nil, false
	}
	if r.MultipartForm != nil {
		defer func() { _ = r.MultipartForm.RemoveAll() }()
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		respondProblem(w, http.StatusBadRequest, "backup_file_missing", "Eine Backup-Datei ist erforderlich.")
		return nil, false
	}
	defer func() { _ = file.Close() }()
	if header.Size > maxBackupBytes {
		respondProblem(w, http.StatusBadRequest, "backup_file_too_large", "Die Backup-Datei ist zu gro?.")
		return nil, false
	}
	data, err := io.ReadAll(io.LimitReader(file, maxBackupBytes+1))
	if err != nil || int64(len(data)) > maxBackupBytes {
		respondProblem(w, http.StatusBadRequest, "backup_file_too_large", "Die Backup-Datei ist zu gro?.")
		return nil, false
	}

	backup, err := application.DecodeBackup(data)
	if err != nil {
		respondProblem(w, http.StatusBadRequest, "backup_restore_invalid", "Backup-Datei ist ungültig.")
		return nil, false
	}
	return backup, true
}

func (a *App) exportMasterData(w http.ResponseWriter, r *http.Request) {
	doc, err := a.masterDataService.Export(r.Context())
	if err != nil {
		a.logger.Error("master data export failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "master_data_export_failed", "Stammdaten konnten nicht exportiert werden.")
		return
	}
	filename := "railkeeper-stammdaten-" + time.Now().UTC().Format("20060102-150405") + ".json"
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": filename}))
	if err := json.NewEncoder(w).Encode(doc); err != nil {
		a.logger.Error("master data encode failed", "error", err)
	}
}

func (a *App) importMasterData(w http.ResponseWriter, r *http.Request) {
	doc, ok := a.readMasterDataImportUpload(w, r)
	if !ok {
		return
	}
	result, err := a.masterDataService.Import(r.Context(), doc)
	if err != nil {
		if errors.Is(err, application.ErrMasterDataValidation) {
			respondProblem(w, http.StatusBadRequest, "master_data_import_invalid", "Stammdaten-Datei ist ungültig.")
			return
		}
		a.logger.Error("master data import failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "master_data_import_failed", "Stammdaten konnten nicht importiert werden.")
		return
	}
	if err := a.masterDataService.WarmCache(r.Context()); err != nil {
		a.logger.Error("master data cache refresh after import failed", "error", err)
	}
	respondJSON(w, http.StatusOK, result)
}

func (a *App) readMasterDataImportUpload(w http.ResponseWriter, r *http.Request) (*application.MasterDataDocument, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, maxMasterDataImportBytes+1024*1024)
	if err := r.ParseMultipartForm(maxMasterDataImportBytes); err != nil {
		respondProblem(w, http.StatusBadRequest, "master_data_import_invalid", "Stammdaten-Datei konnte nicht gelesen werden.")
		return nil, false
	}
	if r.MultipartForm != nil {
		defer func() { _ = r.MultipartForm.RemoveAll() }()
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		respondProblem(w, http.StatusBadRequest, "master_data_file_missing", "Eine Stammdaten-Datei ist erforderlich.")
		return nil, false
	}
	defer func() { _ = file.Close() }()
	if header.Size > maxMasterDataImportBytes {
		respondProblem(w, http.StatusBadRequest, "master_data_file_too_large", "Die Stammdaten-Datei ist zu gro?.")
		return nil, false
	}
	data, err := io.ReadAll(io.LimitReader(file, maxMasterDataImportBytes+1))
	if err != nil || int64(len(data)) > maxMasterDataImportBytes {
		respondProblem(w, http.StatusBadRequest, "master_data_file_too_large", "Die Stammdaten-Datei ist zu gro?.")
		return nil, false
	}
	var doc application.MasterDataDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		respondProblem(w, http.StatusBadRequest, "master_data_import_invalid", "Stammdaten-Datei ist ungültig.")
		return nil, false
	}
	return &doc, true
}

func cleanOriginalFileName(value string) string {
	value = strings.TrimSpace(strings.ReplaceAll(value, "\\", "/"))
	value = strings.TrimSpace(path.Base(value))
	if value == "" || value == "." || value == "/" {
		return "beilage"
	}
	return value
}

func safeAttachmentFileName(value string) string {
	value = cleanOriginalFileName(value)
	if value == "" {
		return "beilage"
	}
	value = safeFileNamePattern.ReplaceAllString(value, "-")
	value = strings.Trim(value, ".-")
	if value == "" {
		return "beilage"
	}
	return value
}

func safePathSegment(value string) string {
	value = safeFileNamePattern.ReplaceAllString(strings.TrimSpace(value), "-")
	value = strings.Trim(value, ".-")
	if value == "" {
		return "unknown"
	}
	return value
}

func confinedDataPath(dataDir, relativePath string) (string, error) {
	base, err := filepath.Abs(dataDir)
	if err != nil {
		return "", err
	}
	target, err := filepath.Abs(filepath.Join(base, relativePath))
	if err != nil {
		return "", err
	}
	if target != base && !strings.HasPrefix(target, base+string(os.PathSeparator)) {
		return "", errors.New("path escapes data directory")
	}
	return target, nil
}

func isBlockedAttachmentName(value string) bool {
	switch strings.ToLower(filepath.Ext(value)) {
	case ".exe", ".bat", ".cmd", ".com", ".scr", ".msi", ".dll", ".ps1", ".vbs", ".js", ".jar", ".sh":
		return true
	default:
		return false
	}
}

func isBlockedAttachmentMime(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return strings.Contains(value, "x-msdownload") ||
		strings.Contains(value, "x-dosexec") ||
		strings.Contains(value, "x-sh") ||
		strings.Contains(value, "javascript") ||
		strings.Contains(value, "ecmascript") ||
		strings.Contains(value, "x-msdos-program")
}

func isAllowedAttachmentUpload(filename, mimeType string) bool {
	return isAllowedAttachmentUploadWithExtensions(filename, mimeType, allowedAttachmentExtensions)
}

func (a *App) isAllowedAttachmentUpload(filename, mimeType string) bool {
	return isAllowedAttachmentUploadWithExtensions(filename, mimeType, a.allowedAttachmentExtensions)
}

func isAllowedAttachmentUploadWithExtensions(filename, mimeType string, extensions map[string]struct{}) bool {
	if isBlockedAttachmentName(filename) || isBlockedAttachmentMime(mimeType) {
		return false
	}
	extension := strings.ToLower(filepath.Ext(filename))
	if _, ok := extensions[extension]; !ok {
		return false
	}
	mimeType = strings.ToLower(strings.TrimSpace(strings.Split(mimeType, ";")[0]))
	switch extension {
	case ".pdf":
		return mimeType == "application/pdf"
	case ".jpg", ".jpeg":
		return mimeType == "image/jpeg"
	case ".png":
		return mimeType == "image/png"
	case ".webp":
		return mimeType == "image/webp"
	case ".zip":
		return mimeType == "application/zip" || mimeType == "application/x-zip-compressed" || mimeType == "application/octet-stream"
	case ".txt", ".csv", ".json", ".xml":
		return strings.HasPrefix(mimeType, "text/") ||
			mimeType == "application/json" ||
			mimeType == "application/xml"
	default:
		return false
	}
}

func isAllowedImageMime(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "image/jpeg", "image/png", "image/webp":
		return true
	default:
		return false
	}
}

type rateLimitStore interface {
	Allow(ctx context.Context, scope, key string, limit int, window time.Duration) (bool, error)
}

type rateLimiter struct {
	mu       sync.Mutex
	attempts map[string][]time.Time
}

func newRateLimiter() *rateLimiter {
	return &rateLimiter{attempts: map[string][]time.Time{}}
}

func (r *rateLimiter) Allow(ctx context.Context, scope, key string, limit int, window time.Duration) (bool, error) {
	if r == nil {
		return true, nil
	}
	now := time.Now()
	cutoff := now.Add(-window)
	r.mu.Lock()
	defer r.mu.Unlock()

	compoundKey := scope + ":" + key
	current := r.attempts[compoundKey]
	filtered := current[:0]
	for _, attempt := range current {
		if attempt.After(cutoff) {
			filtered = append(filtered, attempt)
		}
	}
	if len(filtered) >= limit {
		r.attempts[compoundKey] = filtered
		return false, nil
	}
	r.attempts[compoundKey] = append(filtered, now)
	return true, nil
}

func (a *App) allowRequest(w http.ResponseWriter, r *http.Request, scope, key string, limit int, window time.Duration) (bool, bool) {
	allowed, err := a.rateLimits.Allow(r.Context(), scope, key, limit, window)
	if err != nil {
		a.logger.Error("rate limit check failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "rate_limit_failed", "Could not verify request rate limit.")
		return false, false
	}
	return allowed, true
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	if r.RemoteAddr == "" {
		return "unknown"
	}
	return r.RemoteAddr
}

func (a *App) searchArticleData(w http.ResponseWriter, r *http.Request) {
	var input application.ArticleSearchInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	result, err := a.articleSearch.Search(r.Context(), input)
	if err != nil {
		if errors.Is(err, application.ErrArticleSearchValidation) {
			respondProblem(w, http.StatusBadRequest, "article_search_validation", "At least one search field is required.")
			return
		}
		a.logger.Error("article search failed", "error", err)
		respondProblem(w, http.StatusGatewayTimeout, "article_search_failed", "Artikeldaten-Websuche konnte nicht abgeschlossen werden.")
		return
	}

	respondJSON(w, http.StatusOK, result)
}

func (a *App) listInventoryNumberSchemes(w http.ResponseWriter, r *http.Request) {
	schemes, err := a.inventoryNumbers.List(r.Context())
	if err != nil {
		a.logger.Error("inventory number scheme list failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "inventory_number_scheme_list_failed", "Could not list inventory number schemes.")
		return
	}

	respondJSON(w, http.StatusOK, schemes)
}

func (a *App) createInventoryNumberScheme(w http.ResponseWriter, r *http.Request) {
	var input application.InventoryNumberSchemeCreateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	scheme, err := a.inventoryNumbers.Create(r.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrInventoryNumberValidation):
			respondProblem(w, http.StatusBadRequest, "inventory_number_validation", "Category, prefix, next number and padding are required.")
		case errors.Is(err, application.ErrInventoryNumberConflict):
			respondProblem(w, http.StatusConflict, "inventory_number_scheme_exists", "Inventory number scheme already exists.")
		default:
			a.logger.Error("inventory number scheme create failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "inventory_number_scheme_create_failed", "Could not create inventory number scheme.")
		}
		return
	}

	respondJSON(w, http.StatusCreated, scheme)
}

func (a *App) updateInventoryNumberScheme(w http.ResponseWriter, r *http.Request) {
	var input application.InventoryNumberSchemeInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	scheme, err := a.inventoryNumbers.Update(r.Context(), r.PathValue("category"), input)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrInventoryNumberValidation):
			respondProblem(w, http.StatusBadRequest, "inventory_number_validation", "Prefix, next number and padding are required.")
		case errors.Is(err, application.ErrInventoryNumberNotFound):
			respondProblem(w, http.StatusNotFound, "inventory_number_scheme_not_found", "Inventory number scheme not found.")
		default:
			a.logger.Error("inventory number scheme update failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "inventory_number_scheme_update_failed", "Could not update inventory number scheme.")
		}
		return
	}

	respondJSON(w, http.StatusOK, scheme)
}

func (a *App) listMasterData(w http.ResponseWriter, r *http.Request) {
	activeOnly := r.URL.Query().Get("active") == "true"
	items, err := a.masterDataService.List(r.Context(), r.PathValue("type"), activeOnly)
	if err != nil {
		if errors.Is(err, application.ErrMasterDataValidation) {
			respondProblem(w, http.StatusBadRequest, "master_data_validation", "Master data type is required.")
			return
		}
		a.logger.Error("master data list failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "master_data_list_failed", "Could not list master data.")
		return
	}

	respondJSON(w, http.StatusOK, items)
}

func (a *App) listAllMasterData(w http.ResponseWriter, r *http.Request) {
	activeOnly := r.URL.Query().Get("active") == "true"
	items, err := a.masterDataService.ListAll(r.Context(), activeOnly)
	if err != nil {
		a.logger.Error("master data list all failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "master_data_list_failed", "Could not list master data.")
		return
	}

	respondJSON(w, http.StatusOK, items)
}

func (a *App) createMasterData(w http.ResponseWriter, r *http.Request) {
	var input application.MasterDataInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	item, err := a.masterDataService.Create(r.Context(), r.PathValue("type"), input)
	if err != nil {
		if errors.Is(err, application.ErrMasterDataValidation) {
			respondProblem(w, http.StatusBadRequest, "master_data_validation", "Label is required.")
			return
		}
		a.logger.Error("master data create failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "master_data_create_failed", "Could not create master data.")
		return
	}

	respondJSON(w, http.StatusCreated, item)
}

func (a *App) updateMasterData(w http.ResponseWriter, r *http.Request) {
	var input application.MasterDataInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	item, err := a.masterDataService.Update(r.Context(), r.PathValue("type"), r.PathValue("key"), input)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrMasterDataValidation):
			respondProblem(w, http.StatusBadRequest, "master_data_validation", "Label is required.")
		case errors.Is(err, application.ErrMasterDataNotFound):
			respondProblem(w, http.StatusNotFound, "master_data_not_found", "Master data entry not found.")
		default:
			a.logger.Error("master data update failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "master_data_update_failed", "Could not update master data.")
		}
		return
	}

	respondJSON(w, http.StatusOK, item)
}

func (a *App) deleteMasterData(w http.ResponseWriter, r *http.Request) {
	if err := a.masterDataService.Delete(r.Context(), r.PathValue("type"), r.PathValue("key")); err != nil {
		if errors.Is(err, application.ErrMasterDataNotFound) {
			respondProblem(w, http.StatusNotFound, "master_data_not_found", "Master data entry not found.")
			return
		}
		a.logger.Error("master data delete failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "master_data_delete_failed", "Could not delete master data.")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *App) listMasterDataRelations(w http.ResponseWriter, r *http.Request) {
	relations, err := a.masterDataService.Relations(
		r.Context(),
		r.URL.Query().Get("parentType"),
		r.URL.Query().Get("childType"),
	)
	if err != nil {
		a.logger.Error("master data relations list failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "master_data_relations_failed", "Could not list master data relations.")
		return
	}

	respondJSON(w, http.StatusOK, relations)
}

func respondJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func respondProblem(w http.ResponseWriter, status int, code, message string) {
	respondJSON(w, status, map[string]string{
		"error":   code,
		"message": message,
	})
}

func vehicleECoSMappingForSync(vehicle *application.Vehicle, objectID int) *application.VehicleExternalMap {
	if vehicle == nil {
		return nil
	}
	for index := range vehicle.ExternalMappings {
		mapping := &vehicle.ExternalMappings[index]
		if mapping.Provider != "ecos" {
			continue
		}
		if objectID <= 0 || mapping.ExternalID == strconv.Itoa(objectID) {
			return mapping
		}
	}
	return nil
}

func parsePositiveIntText(value string) int {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return 0
	}
	return parsed
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "same-origin")
		w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; connect-src 'self'; img-src 'self' data: blob: http: https:; style-src 'self' 'unsafe-inline'; script-src 'self'; object-src 'none'; frame-ancestors 'self'; base-uri 'self'")
		next.ServeHTTP(w, r)
	})
}

func (a *App) csrf(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}
		if r.URL.Path == "/api/v1/setup/admin" ||
			r.URL.Path == "/api/v1/auth/login" ||
			r.URL.Path == "/api/v1/auth/password-reset" ||
			r.URL.Path == "/api/v1/auth/password-reset/confirm" ||
			r.URL.Path == "/api/v1/auth/logout" {
			next.ServeHTTP(w, r)
			return
		}
		if !strings.HasPrefix(r.URL.Path, "/api/") {
			next.ServeHTTP(w, r)
			return
		}

		if err := a.authService.ValidateCSRF(r.Context(), cookieValue(r, "rk_session"), r.Header.Get("X-CSRF-Token")); err != nil {
			respondProblem(w, http.StatusForbidden, "csrf_required", "CSRF token is missing or invalid.")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (a *App) require(role string, next http.HandlerFunc) http.HandlerFunc {
	return a.requireAny([]string{role}, next)
}

func (a *App) requireMasterDataRead(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roles := []string{"Viewer"}
		if r.PathValue("type") == "symbols" {
			roles = append(roles, "Messe")
		}
		a.requireAny(roles, next)(w, r)
	}
}

func (a *App) requireAny(roles []string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := a.authService.RequireAnyRole(r.Context(), cookieValue(r, "rk_session"), roles...)
		if err != nil {
			if errors.Is(err, application.ErrUnauthorized) {
				respondProblem(w, http.StatusUnauthorized, "unauthorized", "Not logged in.")
				return
			}
			if errors.Is(err, application.ErrForbidden) {
				respondProblem(w, http.StatusForbidden, "forbidden", "Insufficient role.")
				return
			}
			a.logger.Error("role check failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "role_check_failed", "Could not verify permissions.")
			return
		}

		next.ServeHTTP(w, withActorUserID(r, userID))
	}
}

func setCookie(w http.ResponseWriter, name, value string, maxAge int, httpOnly, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: httpOnly,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
	})
}

func clearCookie(w http.ResponseWriter, name string, httpOnly, secure bool) {
	setCookie(w, name, "", -1, httpOnly, secure)
}

func cookieValue(r *http.Request, name string) string {
	cookie, err := r.Cookie(name)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func timeUntil(t time.Time) time.Duration {
	duration := time.Until(t)
	if duration < time.Second {
		return time.Second
	}
	return duration
}

type actorUserIDKey struct{}

func withActorUserID(r *http.Request, userID string) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), actorUserIDKey{}, userID))
}

func actorUserID(r *http.Request) string {
	value, _ := r.Context().Value(actorUserIDKey{}).(string)
	return value
}

func staticHandler(staticDir string) http.Handler {
	fileServer := http.FileServer(http.Dir(staticDir))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			respondJSON(w, http.StatusNotFound, map[string]string{
				"error":   "not_found",
				"message": "API route not found",
			})
			return
		}

		path := filepath.Join(staticDir, filepath.Clean(r.URL.Path))
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			if strings.HasPrefix(r.URL.Path, "/assets/") {
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			} else {
				w.Header().Set("Cache-Control", "no-cache")
			}
			fileServer.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Cache-Control", "no-store")
		http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
	})
}
