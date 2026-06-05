package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"railkeeper/backend/internal/api"
	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/infrastructure"
)

const (
	version               = "0.1.14"
	defaultUpdateCheckURL = "https://api.github.com/repos/ichwars/RailKeeper/releases/latest"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		resp, err := http.Get("http://127.0.0.1:8080/health")
		if err != nil || resp.StatusCode > http.StatusOK {
			os.Exit(1)
		}
		return
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	portable := portableMode()
	baseDir := executableDir()
	addrDefault := ":8080"
	dataDirDefault := "./data"
	migrationsDirDefault := "./migrations"
	seedsDirDefault := "./seeds"
	staticDirDefault := "../../frontend/dist"
	if portable {
		addrDefault = "127.0.0.1:8080"
		dataDirDefault = filepath.Join(baseDir, "data")
		migrationsDirDefault = filepath.Join(baseDir, "migrations")
		seedsDirDefault = filepath.Join(baseDir, "seeds")
		staticDirDefault = filepath.Join(baseDir, "web")
	}
	addr := env("RAILKEEPER_ADDR", addrDefault)
	dataDir := env("RAILKEEPER_DATA_DIR", dataDirDefault)
	migrationsDir := env("RAILKEEPER_MIGRATIONS_DIR", migrationsDirDefault)
	seedsDir := env("RAILKEEPER_SEEDS_DIR", seedsDirDefault)
	staticDir := env("RAILKEEPER_STATIC_DIR", staticDirDefault)
	cookieSecure := env("RAILKEEPER_COOKIE_SECURE", "false") == "true"
	maxImageBytes := envMegabytes("RAILKEEPER_MAX_IMAGE_MB", 10)
	maxAttachmentBytes := envMegabytes("RAILKEEPER_MAX_ATTACHMENT_MB", 25)
	allowedAttachmentExtensions := envExtensionSet("RAILKEEPER_ALLOWED_ATTACHMENT_EXTENSIONS")
	updateCheckURL := env("RAILKEEPER_UPDATE_CHECK_URL", defaultUpdateCheckURL)

	listener, appURL, err := listen(addr, portable)
	if err != nil {
		logger.Error("server listen failed", "error", err, "addr", addr)
		os.Exit(1)
	}
	defer func() { _ = listener.Close() }()

	publicURL := env("RAILKEEPER_PUBLIC_URL", "")
	if portable && publicURL == "" {
		publicURL = appURL
	}
	smtpConfig := application.SMTPPasswordResetMailConfig{
		Host:     env("RAILKEEPER_SMTP_HOST", ""),
		Port:     env("RAILKEEPER_SMTP_PORT", "587"),
		Username: env("RAILKEEPER_SMTP_USER", ""),
		Password: env("RAILKEEPER_SMTP_PASSWORD", ""),
		From:     env("RAILKEEPER_SMTP_FROM", ""),
		TLSMode:  env("RAILKEEPER_SMTP_TLS", "starttls"),
	}
	passwordResetMailer, err := application.NewSMTPPasswordResetMailer(smtpConfig)
	if err != nil {
		logger.Error("smtp configuration invalid", "error", err)
		os.Exit(1)
	}

	db, err := infrastructure.OpenSQLite(dataDir)
	if err != nil {
		logger.Error("database open failed", "error", err)
		os.Exit(1)
	}
	defer func() { _ = db.Close() }()

	if err = infrastructure.Migrate(db, migrationsDir); err != nil {
		logger.Error("database migration failed", "error", err)
		os.Exit(1)
	}
	if err = infrastructure.SeedRoles(db); err != nil {
		logger.Error("role seed failed", "error", err)
		os.Exit(1)
	}
	if err = infrastructure.SeedMasterData(db, seedsDir); err != nil {
		logger.Error("master data seed failed", "error", err)
		os.Exit(1)
	}
	fileBlobService := application.NewFileBlobService(db, dataDir)
	if err = fileBlobService.MigrateFilesystemBlobs(context.Background()); err != nil {
		logger.Error("file blob migration failed", "error", err)
		os.Exit(1)
	}

	masterDataService := application.NewMasterDataService(db)
	if err = masterDataService.WarmCache(context.Background()); err != nil {
		logger.Error("master data cache warmup failed", "error", err)
		os.Exit(1)
	}

	handler := api.NewRouter(api.Config{
		Version:                     version,
		UpdateCheckURL:              updateCheckURL,
		StaticDir:                   staticDir,
		DataDir:                     dataDir,
		MaxImageBytes:               maxImageBytes,
		MaxAttachmentBytes:          maxAttachmentBytes,
		AllowedAttachmentExtensions: allowedAttachmentExtensions,
		Logger:                      logger,
		SetupService:                application.NewSetupService(db),
		AuthService:                 application.NewAuthService(db),
		VehicleService:              application.NewVehicleService(db),
		MasterDataService:           masterDataService,
		ArticleSearch:               application.NewArticleSearchService(masterDataService),
		InventoryNumbers:            application.NewInventoryNumberService(db),
		BackupService:               application.NewBackupService(db, dataDir),
		FileBlobService:             fileBlobService,
		DatabaseMaintenance:         application.NewDatabaseMaintenanceService(db, dataDir),
		ExhibitionService:           application.NewExhibitionService(db),
		ECoSService:                 application.NewECoSService(),
		RateLimitService:            application.NewRateLimitService(db),
		SettingsService:             application.NewSettingsService(db),
		PasswordResetMailer:         passwordResetMailer,
		SMTPSettingsService:         application.NewSMTPSettingsService(db, smtpConfig, publicURL),
		PublicURL:                   publicURL,
		CookieSecure:                cookieSecure,
	})

	server := &http.Server{
		Addr:              listener.Addr().String(),
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	logger.Info("railkeeper started", "addr", server.Addr, "url", appURL, "version", version, "portable", portable)
	if portable {
		printPortableStart(appURL, dataDir)
		if env("RAILKEEPER_OPEN_BROWSER", "true") != "false" {
			go openBrowser(logger, appURL)
		}
	}
	if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func portableMode() bool {
	if env("RAILKEEPER_PORTABLE", "false") == "true" {
		return true
	}
	for _, arg := range os.Args[1:] {
		if arg == "--portable" {
			return true
		}
	}
	baseDir := executableDir()
	return exists(filepath.Join(baseDir, "web", "index.html")) &&
		exists(filepath.Join(baseDir, "migrations")) &&
		exists(filepath.Join(baseDir, "seeds"))
}

func executableDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exe)
}

func listen(addr string, allowFallback bool) (net.Listener, string, error) {
	listener, err := net.Listen("tcp", addr)
	if err == nil {
		return listener, browserURL(listener.Addr().String()), nil
	}
	if !allowFallback {
		return nil, "", err
	}
	host, portText, splitErr := net.SplitHostPort(addr)
	if splitErr != nil {
		return nil, "", err
	}
	port, parseErr := strconv.Atoi(portText)
	if parseErr != nil {
		return nil, "", err
	}
	for nextPort := port + 1; nextPort <= port+10; nextPort++ {
		nextAddr := net.JoinHostPort(host, strconv.Itoa(nextPort))
		listener, listenErr := net.Listen("tcp", nextAddr)
		if listenErr == nil {
			return listener, browserURL(listener.Addr().String()), nil
		}
	}
	return nil, "", err
}

func browserURL(addr string) string {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return "http://127.0.0.1:8080"
	}
	if host == "" || host == "::" || host == "0.0.0.0" || host == "[::]" {
		host = "127.0.0.1"
	}
	if strings.Contains(host, ":") && !strings.HasPrefix(host, "[") {
		host = "[" + host + "]"
	}
	return "http://" + host + ":" + port
}

func printPortableStart(appURL, dataDir string) {
	fmt.Println()
	fmt.Println("RailKeeper Portable wurde gestartet.")
	fmt.Println("Adresse: " + appURL)
	fmt.Println("Datenordner: " + dataDir)
	fmt.Println("Dieses Fenster waehrend der Nutzung geoeffnet lassen.")
	fmt.Println()
}

func openBrowser(logger *slog.Logger, appURL string) {
	time.Sleep(700 * time.Millisecond)
	var command string
	var args []string
	switch runtime.GOOS {
	case "windows":
		command = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", appURL}
	case "darwin":
		command = "open"
		args = []string{appURL}
	default:
		command = "xdg-open"
		args = []string{appURL}
	}
	if err := exec.Command(command, args...).Start(); err != nil {
		logger.Warn("browser open failed", "error", err, "url", appURL)
	}
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func env(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func envMegabytes(key string, fallback int64) int64 {
	value := env(key, "")
	if value == "" {
		return fallback * 1024 * 1024
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed <= 0 {
		return fallback * 1024 * 1024
	}
	return parsed * 1024 * 1024
}

func envExtensionSet(key string) map[string]struct{} {
	value := env(key, "")
	if value == "" {
		return nil
	}
	out := map[string]struct{}{}
	for _, part := range strings.Split(value, ",") {
		extension := strings.ToLower(strings.TrimSpace(part))
		if extension == "" {
			continue
		}
		if !strings.HasPrefix(extension, ".") {
			extension = "." + extension
		}
		out[extension] = struct{}{}
	}
	return out
}
