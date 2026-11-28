package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/buildinfo"
	"railkeeper/backend/internal/startup"
)

var (
	version = "0.3.0"
)

const defaultUpdateCheckURL = "https://api.github.com/repos/ichwars/RailKeeper/releases/latest"

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "version" || os.Args[1] == "--version") {
		if err := writeVersion(os.Stdout, version, buildinfo.Current()); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
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
	if err := runServer(logger); err != nil {
		logger.Error("RailKeeper stopped", "error", err)
		os.Exit(1)
	}
}

func runServer(logger *slog.Logger) (exitErr error) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	baseDir := executableDir()
	workingDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("working directory resolution failed: %w", err)
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
		return fmt.Errorf("runtime configuration failed: %w", err)
	}
	control, err := openDesktopControl(os.Args[1:], runtimeConfig.WindowsStandalone)
	if err != nil {
		return err
	}
	if control != nil {
		defer func() { _ = control.ready.Close() }()
		var cancel context.CancelFunc
		ctx, cancel = context.WithCancel(ctx)
		defer cancel()
		go control.watch(cancel)
	}
	standalone := runtimeConfig.Standalone
	addr := env("RAILKEEPER_ADDR", runtimeConfig.AddrDefault)
	staticDir := runtimeConfig.StaticDir
	maxImageBytes := envMegabytes("RAILKEEPER_MAX_IMAGE_MB", 10)
	maxAttachmentBytes := envMegabytes("RAILKEEPER_MAX_ATTACHMENT_MB", 25)
	allowedAttachmentExtensions := envExtensionSet("RAILKEEPER_ALLOWED_ATTACHMENT_EXTENSIONS")
	updateCheckURL := env("RAILKEEPER_UPDATE_CHECK_URL", defaultUpdateCheckURL)
	trustedProxyCIDRs := envList("RAILKEEPER_TRUSTED_PROXY_CIDRS")
	bugReports := bugReportOptionsFromLookup(os.LookupEnv)

	loginWork, err := startup.ResolveLoginWorkLimiter(os.LookupEnv)
	if err != nil {
		return fmt.Errorf("login resource configuration failed: %w", err)
	}

	listener, appURL, err := listen(addr, standalone)
	if err != nil {
		return fmt.Errorf("server listen failed: %w", err)
	}
	defer func() { _ = listener.Close() }()
	if control != nil && !addressIsLoopback(listener.Addr()) {
		return errors.New("desktop control requires a loopback listener")
	}

	publicURL := env("RAILKEEPER_PUBLIC_URL", "")
	if standalone && publicURL == "" {
		publicURL = appURL
	}
	cookieSecure, err := startup.ResolveCookieSecurity(env("RAILKEEPER_COOKIE_SECURE", "auto"), publicURL)
	if err != nil {
		return fmt.Errorf("cookie security configuration failed: %w", err)
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
			BugReports:                  bugReports,
			LoginWork:                   loginWork,
		}),
	)
	if err != nil {
		return fmt.Errorf("application startup failed: %w", err)
	}
	if startupResult.Database != nil {
		defer func() {
			if err := startupResult.Database.Close(); err != nil {
				exitErr = errors.Join(exitErr, fmt.Errorf("close database: %w", err))
			}
		}()
	}
	if startupResult.StopBackground != nil {
		defer startupResult.StopBackground()
	}
	dataDir := startupResult.State.Runtime.DataDir
	if startupResult.State.SafetyBackupPath != "" {
		logger.Info("database migration safety copy created", "path", startupResult.State.SafetyBackupPath)
	}
	if startupResult.Conflict != nil {
		if !addressIsLoopback(listener.Addr()) {
			return errors.New("legacy data conflict page requires a loopback listener")
		}
		logger.Warn(
			"legacy data conflict detected",
			"safe_path", startupResult.Conflict.SafePath,
			"legacy_path", startupResult.Conflict.LegacyPath,
		)
	}

	if ctx.Err() != nil {
		return nil
	}
	if control != nil {
		_, err := fmt.Fprintln(control.ready, appURL)
		_ = control.ready.Close()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("report desktop readiness: %w", err)
		}
	}

	server := newHTTPServer(listener.Addr().String(), startupResult.Handler)

	logger.Info("railkeeper started", "build", buildinfo.Current(), "addr", server.Addr, "url", appURL, "version", version, "standalone", standalone)
	if standalone {
		printStandaloneStart(appURL, dataDir)
		if env("RAILKEEPER_OPEN_BROWSER", "true") != "false" {
			go openBrowser(logger, appURL)
		}
	}
	return serveUntilShutdown(ctx, server, listener)
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
