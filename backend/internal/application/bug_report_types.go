package application

import (
	"context"
	"errors"
	"time"
)

type BugReportSubmissionState string

const (
	BugReportReserved      BugReportSubmissionState = "reserved"
	BugReportIssueCreated  BugReportSubmissionState = "issue_created"
	BugReportAssetUploaded BugReportSubmissionState = "asset_uploaded"
	BugReportComplete      BugReportSubmissionState = "complete"
	BugReportPartial       BugReportSubmissionState = "partial"
)

var ErrBugReportStateConflict = errors.New("bug report state conflict")

var ErrBugReportValidation = errors.New("bug report validation failed")

var (
	ErrBugReportingUnavailable     = errors.New("bug reporting is unavailable")
	ErrBugReportGitHubUnavailable  = errors.New("GitHub bug reporting is unavailable")
	ErrBugReportScreenshotRequired = errors.New("bug report screenshot is required to continue")
	ErrBugReportRateLimited        = errors.New("bug report rate limit exceeded")
)

type BugReportAdmissionPolicy struct {
	Scope  string
	Limit  int
	Window time.Duration
}

type BugReportArea string

const (
	BugReportAreaOverview       BugReportArea = "overview"
	BugReportAreaVehicles       BugReportArea = "vehicles"
	BugReportAreaAccessories    BugReportArea = "accessories"
	BugReportAreaExhibition     BugReportArea = "exhibition"
	BugReportAreaImportExport   BugReportArea = "importExport"
	BugReportAreaDigitalCenters BugReportArea = "digitalCenters"
	BugReportAreaSettings       BugReportArea = "settings"
	BugReportAreaOther          BugReportArea = "other"
)

type BugReportClientError struct {
	Timestamp   string `json:"timestamp"`
	Category    string `json:"category"`
	ErrorClass  string `json:"errorClass"`
	ProblemCode string `json:"problemCode"`
	Message     string `json:"message"`
	Method      string `json:"method"`
	Route       string `json:"route"`
	Status      int    `json:"status,omitempty"`
}

type BugReportClientDiagnostics struct {
	BrowserFamily  string                 `json:"browserFamily"`
	BrowserMajor   string                 `json:"browserMajor"`
	OSFamily       string                 `json:"osFamily"`
	OSMajor        string                 `json:"osMajor"`
	ViewportWidth  int                    `json:"viewportWidth"`
	ViewportHeight int                    `json:"viewportHeight"`
	Language       string                 `json:"language"`
	TimeZone       string                 `json:"timeZone"`
	Theme          string                 `json:"theme"`
	Pathname       string                 `json:"pathname"`
	Area           string                 `json:"area"`
	Roles          []string               `json:"roles"`
	Errors         []BugReportClientError `json:"errors"`
}

type BugReportServerDiagnostics struct {
	CollectedAt                 string   `json:"collectedAt,omitempty"`
	UptimeSeconds               int64    `json:"uptimeSeconds,omitempty"`
	CPUCount                    int      `json:"cpuCount,omitempty"`
	Goroutines                  int      `json:"goroutines,omitempty"`
	HeapBytes                   uint64   `json:"heapBytes,omitempty"`
	RuntimeBytes                uint64   `json:"runtimeBytes,omitempty"`
	SQLiteVersion               string   `json:"sqliteVersion,omitempty"`
	DatabaseBytes               int64    `json:"databaseBytes,omitempty"`
	JournalMode                 string   `json:"journalMode,omitempty"`
	RailKeeperVersion           string   `json:"railkeeperVersion"`
	DeploymentType              string   `json:"deploymentType"`
	GoVersion                   string   `json:"goVersion"`
	OperatingSystem             string   `json:"operatingSystem"`
	Architecture                string   `json:"architecture"`
	SchemaVersion               int      `json:"schemaVersion"`
	EnabledDigitalProviderTypes []string `json:"enabledDigitalProviderTypes"`
}

type BugReportInput struct {
	OmitDiagnostics   bool
	ReportID          string
	Area              BugReportArea
	Description       string
	ExpectedBehavior  string
	ReproductionSteps string
	Email             string
	ClientDiagnostics BugReportClientDiagnostics
	ScreenshotConsent bool
}

type ValidatedBugReport struct {
	BugReportInput
	AreaLabel string
}

type BugReportIssueInput struct {
	Report            ValidatedBugReport
	ServerDiagnostics BugReportServerDiagnostics
	EncryptedContact  string
	BaseLabels        []string
	HasScreenshot     bool
}

type BugReportIssue struct {
	Title  string
	Body   string
	Labels []string
}

type GitHubBugReportIssue struct {
	Number int64
	URL    string
	Body   string
}

type GitHubBugReportComment struct {
	ID   int64
	Body string
}

type GitHubBugReportAsset struct {
	BlobSHA   string
	CommitSHA string
	URL       string
}

type GitHubBugReportIssueClient interface {
	CreateIssue(ctx context.Context, issue BugReportIssue) (GitHubBugReportIssue, error)
	ListRecentIssues(ctx context.Context, since time.Time, limit int) ([]GitHubBugReportIssue, error)
	ListIssueComments(ctx context.Context, issueNumber int64, limit int) ([]GitHubBugReportComment, error)
	CreateIssueComment(ctx context.Context, issueNumber int64, body string) (GitHubBugReportComment, error)
}

type GitHubBugReportAssetClient interface {
	CreateFile(ctx context.Context, path string, message string, content []byte) (GitHubBugReportAsset, error)
	FindFile(ctx context.Context, path string) (bool, GitHubBugReportAsset, error)
}

type BugReportContactEncryptor interface {
	Encrypt(value string) (string, error)
}

type BugReportServiceConfig struct {
	CollectDiagnostics func(context.Context) (BugReportServerDiagnostics, error)
	Repository         BugReportSubmissionRepository
	IssueClient        GitHubBugReportIssueClient
	AssetClient        GitHubBugReportAssetClient
	ContactEncryptor   BugReportContactEncryptor
	ServerDiagnostics  BugReportServerDiagnostics
	BaseLabels         []string
}

type BugReportSubmissionRequest struct {
	Input               BugReportInput
	ScreenshotBytes     []byte
	ScreenshotMediaType string
}

type BugReportSubmissionResult struct {
	ReportID    string                   `json:"reportId"`
	IssueNumber int64                    `json:"issueNumber"`
	IssueURL    string                   `json:"issueUrl"`
	Status      BugReportSubmissionState `json:"status"`
	Warning     string                   `json:"warning"`
	Created     bool                     `json:"-"`
}

type BugReportSubmission struct {
	ReportID          string
	ActorUserID       string
	State             BugReportSubmissionState
	GitHubIssueNumber int64
	GitHubIssueURL    string
	AssetURL          string
	AssetSHA          string
	ErrorCategory     string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type BugReportSubmissionUpdate struct {
	State             BugReportSubmissionState
	GitHubIssueNumber int64
	GitHubIssueURL    string
	AssetURL          string
	AssetSHA          string
	ErrorCategory     string
}

type BugReportSubmissionRepository interface {
	Reserve(
		ctx context.Context,
		reportID string,
		actorUserID string,
		admission BugReportAdmissionPolicy,
	) (created bool, admitted bool, submission BugReportSubmission, err error)
	Get(ctx context.Context, reportID string) (BugReportSubmission, error)
	Update(ctx context.Context, reportID string, update BugReportSubmissionUpdate) (BugReportSubmission, error)
}
