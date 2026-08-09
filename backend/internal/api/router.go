package api

import (
	"log/slog"
	"net/http"
	"strings"

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
	LayoutService               *application.LayoutService
	TrackPlannerService         *application.TrackPlannerService
	AccessoryService            *application.AccessoryService
	AccessoryAllocationService  *application.AccessoryAllocationService
	AccessoryDocumentService    *application.AccessoryDocumentService
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
	layoutService               *application.LayoutService
	trackPlannerService         *application.TrackPlannerService
	accessoryService            *application.AccessoryService
	accessoryAllocationService  *application.AccessoryAllocationService
	accessoryDocumentService    *application.AccessoryDocumentService
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
		layoutService:               config.LayoutService,
		trackPlannerService:         config.TrackPlannerService,
		accessoryService:            config.AccessoryService,
		accessoryAllocationService:  config.AccessoryAllocationService,
		accessoryDocumentService:    config.AccessoryDocumentService,
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
