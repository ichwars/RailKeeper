package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"railkeeper/backend/internal/api"
	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/infrastructure"
)

var bugReportProcessStarted = time.Now()

type bugReportOptions struct {
	IssueToken          string
	IssueRepository     string
	AssetToken          string
	AssetRepository     string
	ContactAgeRecipient string
	Labels              []string
	APIBaseURL          string
}

func bugReportOptionsFromLookup(lookup func(string) (string, bool)) bugReportOptions {
	read := func(key string) string {
		value, _ := lookup(key)
		return strings.TrimSpace(value)
	}
	labelsValue := read("RAILKEEPER_BUG_REPORT_LABELS")
	if labelsValue == "" {
		labelsValue = "bug,source:in-app"
	}
	seen := map[string]bool{}
	labels := []string{}
	for _, label := range strings.Split(labelsValue, ",") {
		label = strings.TrimSpace(label)
		key := strings.ToLower(label)
		if label == "" || seen[key] {
			continue
		}
		seen[key] = true
		labels = append(labels, label)
	}
	return bugReportOptions{
		IssueToken:          read("RAILKEEPER_GITHUB_ISSUE_TOKEN"),
		IssueRepository:     read("RAILKEEPER_GITHUB_ISSUE_REPOSITORY"),
		AssetToken:          read("RAILKEEPER_GITHUB_ASSET_TOKEN"),
		AssetRepository:     read("RAILKEEPER_GITHUB_ASSET_REPOSITORY"),
		ContactAgeRecipient: read("RAILKEEPER_BUG_REPORT_CONTACT_AGE_RECIPIENT"),
		Labels:              labels,
	}
}

type bugReportFeature struct {
	Service    *application.BugReportService
	Capability api.BugReportCapability
	Provider   func(context.Context) api.BugReportCapability
	Invalidate func()
}

func buildBugReportFeature(
	ctx context.Context,
	database *sql.DB,
	options bugReportOptions,
	diagnostics application.BugReportServerDiagnostics,
	logger *slog.Logger,
) bugReportFeature {
	capability := api.BugReportCapability{ServerDiagnostics: &diagnostics}
	if options.IssueToken == "" || options.IssueRepository == "" {
		return bugReportFeature{Capability: capability}
	}
	if options.AssetToken != "" || options.AssetRepository != "" {
		if err := infrastructure.ValidateGitHubBugReportRepositoryTargets(
			options.IssueRepository, options.AssetRepository,
		); err != nil {
			logger.Warn("bug reporting configuration rejected", "error_category", "repository_targets")
			return bugReportFeature{Capability: capability}
		}
	}
	issueClient, err := infrastructure.NewGitHubBugReportIssueClient(infrastructure.GitHubBugReportClientConfig{
		APIBaseURL: options.APIBaseURL, Token: options.IssueToken, Repository: options.IssueRepository,
	})
	if err != nil {
		logger.Warn("bug reporting configuration rejected", "error_category", "issue_client")
		return bugReportFeature{Capability: capability}
	}
	var assetClient *infrastructure.GitHubBugReportAssetClient
	if options.AssetToken != "" && options.AssetRepository != "" {
		assetClient, err = infrastructure.NewGitHubBugReportAssetClient(infrastructure.GitHubBugReportClientConfig{
			APIBaseURL: options.APIBaseURL, Token: options.AssetToken, Repository: options.AssetRepository,
		})
		if err != nil {
			logger.Warn("bug reporting configuration rejected", "error_category", "asset_client")
			return bugReportFeature{Capability: capability}
		}
		capability.ScreenshotAvailable = true
	}
	requiredLabels := append(append([]string{}, options.Labels...),
		"area:overview", "area:vehicles", "area:accessories", "area:exhibition",
		"area:import-export", "area:digital-centers", "area:settings", "area:other")
	var encryptor application.BugReportContactEncryptor
	if options.ContactAgeRecipient != "" {
		encryptor, err = application.NewAgeContactEncryptor(options.ContactAgeRecipient)
		if err != nil {
			logger.Warn("bug reporting contact encryption rejected", "error_category", "age_recipient")
			return bugReportFeature{Capability: capability}
		}
		capability.ContactAvailable = true
	}
	var assetService application.GitHubBugReportAssetClient
	if assetClient != nil {
		assetService = assetClient
	}
	service := application.NewBugReportService(application.BugReportServiceConfig{
		Repository: infrastructure.NewBugReportRepository(database), IssueClient: issueClient,
		AssetClient: assetService, ContactEncryptor: encryptor, ServerDiagnostics: diagnostics,
		BaseLabels: options.Labels,
		CollectDiagnostics: func(ctx context.Context) (application.BugReportServerDiagnostics, error) {
			return collectBugReportServerDiagnostics(ctx, database, application.NewSettingsService(database),
				StartupState{}, diagnostics.RailKeeperVersion, diagnostics.DeploymentType)
		},
	})
	cache := newBugReportCapabilityCache(capability, func(validationContext context.Context) error {
		return infrastructure.ValidateGitHubBugReportCapability(
			validationContext, issueClient, assetClient, requiredLabels,
		)
	})
	current := cache.Current(ctx)
	if !current.Enabled {
		logger.Warn("bug reporting capability unavailable", "error_category", "github_capability")
	}
	return bugReportFeature{
		Service: service, Capability: current, Provider: cache.Current, Invalidate: cache.Invalidate,
	}
}

func collectBugReportServerDiagnostics(
	ctx context.Context,
	database *sql.DB,
	settingsService *application.SettingsService,
	_ StartupState,
	appVersion string,
	deploymentType string,
) (application.BugReportServerDiagnostics, error) {
	diagnostics := application.BugReportServerDiagnostics{
		RailKeeperVersion: appVersion, DeploymentType: deploymentType, GoVersion: runtime.Version(),
		OperatingSystem: runtime.GOOS, Architecture: runtime.GOARCH,
	}
	diagnostics.CollectedAt = time.Now().UTC().Format(time.RFC3339)
	diagnostics.UptimeSeconds = int64(time.Since(bugReportProcessStarted).Seconds())
	diagnostics.CPUCount = runtime.NumCPU()
	diagnostics.Goroutines = runtime.NumGoroutine()
	var memory runtime.MemStats
	runtime.ReadMemStats(&memory)
	diagnostics.HeapBytes, diagnostics.RuntimeBytes = memory.HeapAlloc, memory.Sys
	if err := database.QueryRowContext(ctx, `SELECT sqlite_version()`).Scan(&diagnostics.SQLiteVersion); err != nil {
		return diagnostics, fmt.Errorf("read SQLite version: %w", err)
	}
	var pageSize, pageCount int64
	if err := database.QueryRowContext(ctx, `PRAGMA page_size`).Scan(&pageSize); err != nil {
		return diagnostics, err
	}
	if err := database.QueryRowContext(ctx, `PRAGMA page_count`).Scan(&pageCount); err != nil {
		return diagnostics, err
	}
	diagnostics.DatabaseBytes = pageSize * pageCount
	if err := database.QueryRowContext(ctx, `PRAGMA journal_mode`).Scan(&diagnostics.JournalMode); err != nil {
		return diagnostics, err
	}
	var migration string
	if err := database.QueryRowContext(ctx, `SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1`).Scan(
		&migration,
	); err != nil {
		return diagnostics, fmt.Errorf("read schema version: %w", err)
	}
	prefix := migration
	if index := strings.IndexByte(prefix, '_'); index >= 0 {
		prefix = prefix[:index]
	}
	diagnostics.SchemaVersion, _ = strconv.Atoi(prefix)
	settings, err := settingsService.DigitalSettings(ctx)
	if err != nil {
		return diagnostics, fmt.Errorf("read enabled digital providers: %w", err)
	}
	if settings.ECoS.Enabled {
		diagnostics.EnabledDigitalProviderTypes = append(diagnostics.EnabledDigitalProviderTypes, "ecos")
	}
	if settings.Z21.Enabled {
		diagnostics.EnabledDigitalProviderTypes = append(diagnostics.EnabledDigitalProviderTypes, "z21")
	}
	if settings.Intellibox3.Enabled {
		diagnostics.EnabledDigitalProviderTypes = append(diagnostics.EnabledDigitalProviderTypes, "intellibox3")
	}
	if settings.CS3.Enabled {
		diagnostics.EnabledDigitalProviderTypes = append(diagnostics.EnabledDigitalProviderTypes, "cs3")
	}
	sort.Strings(diagnostics.EnabledDigitalProviderTypes)
	return diagnostics, nil
}

func bugReportDeploymentType(state StartupState) string {
	if state.Runtime.WindowsStandalone {
		return "windows-portable"
	}
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return "container"
	}
	return "native"
}
