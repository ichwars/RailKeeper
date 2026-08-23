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

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/startup"
)

const (
	version               = "0.1.20.1"
	defaultUpdateCheckURL = "https://api.github.com/repos/ichwars/RailKeeper/releases/latest"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://127.0.0.1:8080/health", nil)
		if err != nil {
			os.Exit(1)
		}
		resp, err := http.DefaultClient.Do(request)
		if err != nil {
			os.Exit(1)
		}
		statusCode := resp.StatusCode
		_ = resp.Body.Close()
		if statusCode > http.StatusOK {
			os.Exit(1)
		}
		return
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	baseDir := executableDir()
	workingDir, err := os.Getwd()
	if err != nil {
		logger.Error("working directory resolution failed", "error", err)
		os.Exit(1)
	}
	runtimeConfig, err := startup.ResolveRuntimeConfig(startup.RuntimeInputs{
		GOOS:          runtime.GOOS,
		Args:          os.Args[1:],
		ExecutableDir: baseDir,
		WorkingDir:    workingDir,
		LookupEnv:     os.LookupEnv,
		PathExists:    exists,
		JoinPath:      filepath.Join,
		AbsPath:       filepath.Abs,
	})
	if err != nil {
		logger.Error("runtime configuration failed", "error", err)
		os.Exit(1)
	}
	standalone := runtimeConfig.Standalone
	addr := env("RAILKEEPER_ADDR", runtimeConfig.AddrDefault)
	staticDir := runtimeConfig.StaticDir
	cookieSecure := env("RAILKEEPER_COOKIE_SECURE", "false") == "true"
	maxImageBytes := envMegabytes("RAILKEEPER_MAX_IMAGE_MB", 10)
	maxAttachmentBytes := envMegabytes("RAILKEEPER_MAX_ATTACHMENT_MB", 25)
	allowedAttachmentExtensions := envExtensionSet("RAILKEEPER_ALLOWED_ATTACHMENT_EXTENSIONS")
	updateCheckURL := env("RAILKEEPER_UPDATE_CHECK_URL", defaultUpdateCheckURL)
	trustedProxyCIDRs := envList("RAILKEEPER_TRUSTED_PROXY_CIDRS")

	listener, appURL, err := listen(addr, standalone)
	if err != nil {
		logger.Error("server listen failed", "error", err, "addr", addr)
		os.Exit(1)
	}
	defer func() { _ = listener.Close() }()

	publicURL := env("RAILKEEPER_PUBLIC_URL", "")
	if standalone && publicURL == "" {
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
	startupResult, err := prepareStartup(
		context.Background(),
		runtimeConfig,
		version,
		defaultStartupDependencies(applicationHandlerOptions{
			Version:                     version,
			UpdateCheckURL:              updateCheckURL,
			StaticDir:                   staticDir,
			MaxImageBytes:               maxImageBytes,
			MaxAttachmentBytes:          maxAttachmentBytes,
			AllowedAttachmentExtensions: allowedAttachmentExtensions,
			TrustedProxyCIDRs:           trustedProxyCIDRs,
			Logger:                      logger,
			SMTPConfig:                  smtpConfig,
			PublicURL:                   publicURL,
			CookieSecure:                cookieSecure,
		}),
	)
	if err != nil {
		logger.Error("application startup failed", "error", err)
		os.Exit(1)
	}
	if startupResult.Database != nil {
		defer func() { _ = startupResult.Database.Close() }()
	}
	dataDir := startupResult.State.Runtime.DataDir
	if startupResult.State.SafetyBackupPath != "" {
		logger.Info("database migration safety copy created", "path", startupResult.State.SafetyBackupPath)
	}
	if startupResult.Conflict != nil {
		if !addressIsLoopback(listener.Addr()) {
			logger.Error("legacy data conflict page requires a loopback listener", "addr", listener.Addr().String())
			os.Exit(1)
		}
		logger.Warn(
			"legacy data conflict detected",
			"safe_path", startupResult.Conflict.SafePath,
			"legacy_path", startupResult.Conflict.LegacyPath,
		)
	}

	server := newHTTPServer(listener.Addr().String(), startupResult.Handler)

	logger.Info("railkeeper started", "addr", server.Addr, "url", appURL, "version", version, "standalone", standalone)
	if standalone {
		printStandaloneStart(appURL, dataDir)
		if env("RAILKEEPER_OPEN_BROWSER", "true") != "false" {
			go openBrowser(logger, appURL)
		}
	}
	if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func newHTTPServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

func executableDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exe)
}

func listen(addr string, allowFallback bool) (net.Listener, string, error) {
	listenConfig := net.ListenConfig{}
	listener, err := listenConfig.Listen(context.Background(), "tcp", addr)
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
		listener, listenErr := listenConfig.Listen(context.Background(), "tcp", nextAddr)
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

func addressIsLoopback(addr net.Addr) bool {
	if tcpAddress, ok := addr.(*net.TCPAddr); ok {
		return tcpAddress.IP.IsLoopback()
	}
	host, _, err := net.SplitHostPort(addr.String())
	if err != nil {
		return false
	}
	address := net.ParseIP(strings.Trim(host, "[]"))
	return address != nil && address.IsLoopback()
}

func printStandaloneStart(appURL, dataDir string) {
	fmt.Println()
	fmt.Println("RailKeeper Windows Standalone wurde gestartet.")
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
	if err := exec.CommandContext(context.Background(), command, args...).Start(); err != nil {
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

func envList(key string) []string {
	value := env(key, "")
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
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
