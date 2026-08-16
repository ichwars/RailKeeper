package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os/exec"
	"path/filepath"

	"railkeeper/backend/internal/api"
	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/infrastructure"
	"railkeeper/backend/internal/startup"
)

type StartupState struct {
	Runtime          startup.RuntimeConfig
	Receipt          *startup.MigrationReceipt
	SafetyBackupPath string
}

type applicationDataPaths struct {
	DataDir       string
	DatabasePath  string
	BlobDataDir   string
	BackupDataDir string
	APIDataDir    string
}

type startupResult struct {
	State    StartupState
	Handler  http.Handler
	Database *sql.DB
	Conflict *startup.LegacyConflictInfo
}

type startupDependencies struct {
	ResolveLegacyData func(
		context.Context,
		string,
		string,
		startup.LegacyMigrationOptions,
	) (startup.LegacyMigrationResult, error)
	Preflight     func() error
	PathExists    func(string) bool
	OpenSQLite    func(string) (*sql.DB, error)
	MigrateSafely func(
		context.Context,
		*sql.DB,
		string,
		string,
		infrastructure.MigrationSafetyOptions,
	) (infrastructure.MigrationSafetyResult, error)
	SeedRoles      func(*sql.DB) error
	SeedMasterData func(*sql.DB, string) error
	BuildHandler   func(
		context.Context, *sql.DB, applicationDataPaths, StartupState,
	) (http.Handler, error)
}

type applicationHandlerOptions struct {
	Version                     string
	UpdateCheckURL              string
	StaticDir                   string
	MaxImageBytes               int64
	MaxAttachmentBytes          int64
	AllowedAttachmentExtensions map[string]struct{}
	Logger                      *slog.Logger
	PasswordResetMailer         application.PasswordResetMailer
	SMTPConfig                  application.SMTPPasswordResetMailConfig
	PublicURL                   string
	CookieSecure                bool
}

func prepareStartup(
	ctx context.Context,
	runtimeConfig startup.RuntimeConfig,
	appVersion string,
	dependencies startupDependencies,
) (result startupResult, err error) {
	if err := validateStartupDependencies(dependencies); err != nil {
		return startupResult{}, err
	}
	result.State.Runtime = runtimeConfig
	if runtimeConfig.LegacyDataDir != "" {
		legacyResult, resolveErr := dependencies.ResolveLegacyData(
			ctx,
			runtimeConfig.DataDir,
			runtimeConfig.LegacyDataDir,
			startup.LegacyMigrationOptions{Version: appVersion},
		)
		if resolveErr != nil {
			return startupResult{}, fmt.Errorf("resolve legacy data: %w", resolveErr)
		}
		if legacyResult.DataDir != "" {
			result.State.Runtime.DataDir = legacyResult.DataDir
		}
		result.State.Receipt = legacyResult.Receipt
		switch legacyResult.Status {
		case startup.LegacyConflict:
			if legacyResult.Conflict == nil {
				return startupResult{}, errors.New("legacy data conflict is missing conflict details")
			}
			result.Conflict = legacyResult.Conflict
			result.Handler = startup.ConflictHandler(*legacyResult.Conflict)
			return result, nil
		case startup.LegacyReady, startup.LegacyMigrated:
		default:
			return startupResult{}, fmt.Errorf("unsupported legacy migration status: %q", legacyResult.Status)
		}
	}
	if preflightErr := dependencies.Preflight(); preflightErr != nil {
		return startupResult{}, fmt.Errorf("startup preflight: %w", preflightErr)
	}

	paths := newApplicationDataPaths(result.State.Runtime.DataDir)
	databaseExisted := dependencies.PathExists(paths.DatabasePath)
	database, openErr := dependencies.OpenSQLite(paths.DataDir)
	if openErr != nil {
		return startupResult{}, fmt.Errorf("open database: %w", openErr)
	}
	result.Database = database
	completed := false
	defer func() {
		if !completed && database != nil {
			_ = database.Close()
		}
	}()

	migrationResult, migrateErr := dependencies.MigrateSafely(
		ctx,
		database,
		paths.DataDir,
		result.State.Runtime.MigrationsDir,
		infrastructure.MigrationSafetyOptions{DatabaseExisted: databaseExisted},
	)
	if migrateErr != nil {
		return startupResult{}, fmt.Errorf("migrate database safely: %w", migrateErr)
	}
	result.State.SafetyBackupPath = migrationResult.BackupPath
	if seedErr := dependencies.SeedRoles(database); seedErr != nil {
		return startupResult{}, fmt.Errorf("seed roles: %w", seedErr)
	}
	if seedErr := dependencies.SeedMasterData(database, result.State.Runtime.SeedsDir); seedErr != nil {
		return startupResult{}, fmt.Errorf("seed master data: %w", seedErr)
	}
	handler, buildErr := dependencies.BuildHandler(ctx, database, paths, result.State)
	if buildErr != nil {
		return startupResult{}, fmt.Errorf("build application handler: %w", buildErr)
	}
	result.Handler = handler
	completed = true
	return result, nil
}

func newApplicationDataPaths(dataDir string) applicationDataPaths {
	return applicationDataPaths{
		DataDir:       dataDir,
		DatabasePath:  filepath.Join(dataDir, "railkeeper.db"),
		BlobDataDir:   dataDir,
		BackupDataDir: dataDir,
		APIDataDir:    dataDir,
	}
}

func validateStartupDependencies(dependencies startupDependencies) error {
	if dependencies.ResolveLegacyData == nil || dependencies.Preflight == nil || dependencies.PathExists == nil ||
		dependencies.OpenSQLite == nil || dependencies.MigrateSafely == nil ||
		dependencies.SeedRoles == nil || dependencies.SeedMasterData == nil ||
		dependencies.BuildHandler == nil {
		return errors.New("startup dependencies are required")
	}
	return nil
}

func defaultStartupDependencies(options applicationHandlerOptions) startupDependencies {
	var passwordResetMailer application.PasswordResetMailer
	return startupDependencies{
		ResolveLegacyData: startup.ResolveLegacyData,
		Preflight: func() error {
			mailer, err := application.NewSMTPPasswordResetMailer(options.SMTPConfig)
			if err != nil {
				return fmt.Errorf("validate SMTP configuration: %w", err)
			}
			passwordResetMailer = mailer
			return nil
		},
		PathExists:     exists,
		OpenSQLite:     infrastructure.OpenSQLite,
		MigrateSafely:  infrastructure.MigrateSafely,
		SeedRoles:      infrastructure.SeedRoles,
		SeedMasterData: infrastructure.SeedMasterData,
		BuildHandler: func(
			ctx context.Context,
			database *sql.DB,
			paths applicationDataPaths,
			state StartupState,
		) (http.Handler, error) {
			options.PasswordResetMailer = passwordResetMailer
			return buildApplicationHandler(ctx, database, paths, state, options)
		},
	}
}

func buildApplicationHandler(
	ctx context.Context,
	database *sql.DB,
	paths applicationDataPaths,
	state StartupState,
	options applicationHandlerOptions,
) (http.Handler, error) {
	fileBlobService := application.NewFileBlobService(database, paths.BlobDataDir)
	if err := fileBlobService.MigrateFilesystemBlobs(ctx); err != nil {
		return nil, fmt.Errorf("migrate file blobs: %w", err)
	}
	masterDataService := application.NewMasterDataService(database)
	if err := masterDataService.WarmCache(ctx); err != nil {
		return nil, fmt.Errorf("warm master data cache: %w", err)
	}

	layoutService := application.NewLayoutService(infrastructure.NewLayoutRepository(database))
	trackPlannerRepository := infrastructure.NewTrackPlannerRepository(database)
	accessoryRepository := infrastructure.NewAccessoryRepository(database)

	return api.NewRouter(api.Config{
		Version:                     options.Version,
		UpdateCheckURL:              options.UpdateCheckURL,
		StaticDir:                   options.StaticDir,
		DataDir:                     paths.APIDataDir,
		MaxImageBytes:               options.MaxImageBytes,
		MaxAttachmentBytes:          options.MaxAttachmentBytes,
		AllowedAttachmentExtensions: options.AllowedAttachmentExtensions,
		Logger:                      options.Logger,
		SetupService:                application.NewSetupService(database),
		AuthService:                 application.NewAuthService(database),
		VehicleService:              application.NewVehicleService(database),
		OverviewValuationService:    application.NewOverviewValuationService(database),
		MasterDataService:           masterDataService,
		ArticleSearch:               application.NewArticleSearchService(masterDataService),
		InventoryNumbers:            application.NewInventoryNumberService(database),
		BackupService:               application.NewBackupService(database, paths.BackupDataDir),
		FileBlobService:             fileBlobService,
		DatabaseMaintenance:         application.NewDatabaseMaintenanceService(database, paths.DataDir),
		ExhibitionService:           application.NewExhibitionService(database),
		LayoutService:               layoutService,
		TrackPlannerService:         application.NewTrackPlannerService(trackPlannerRepository),
		TrackLibraryService:         application.NewTrackLibraryService(trackPlannerRepository),
		AccessoryService:            application.NewAccessoryService(accessoryRepository, fileBlobService),
		AccessoryAllocationService:  application.NewAccessoryAllocationService(accessoryRepository),
		AccessoryDocumentService: application.NewAccessoryDocumentService(
			accessoryRepository, fileBlobService,
		),
		ECoSService:               application.NewECoSService(),
		RateLimitService:          application.NewRateLimitService(database),
		SettingsService:           application.NewSettingsService(database),
		PasswordResetMailer:       options.PasswordResetMailer,
		SMTPSettingsService:       application.NewSMTPSettingsService(database, options.SMTPConfig, options.PublicURL),
		PublicURL:                 options.PublicURL,
		CookieSecure:              options.CookieSecure,
		WindowsStandaloneDownload: state.Runtime.WindowsStandalone,
		StorageLocation: api.StorageLocationConfig{
			DataPath:             paths.DataDir,
			Mode:                 state.Runtime.StorageMode,
			OpenFolderAvailable:  state.Runtime.OpenDataFolderSupported,
			MigrationReceipt:     state.Receipt,
			OpenFolder:           openStorageFolder,
			AcknowledgeMigration: startup.AcknowledgeMigrationReceipt,
		},
	}), nil
}

func storageFolderCommand(dataPath string) *exec.Cmd {
	return exec.CommandContext(context.Background(), "explorer.exe", dataPath)
}

func openStorageFolder(_ context.Context, dataPath string) error {
	command := storageFolderCommand(dataPath)
	if err := command.Start(); err != nil {
		return err
	}
	return command.Process.Release()
}
