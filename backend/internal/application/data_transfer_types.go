package application

import "context"

type TransferDirection string
type TransferFormat string
type TransferArea string
type TransferJobState string
type TransferIssueSeverity string

const (
	TransferImport TransferDirection = "import"
	TransferExport TransferDirection = "export"

	TransferCSV  TransferFormat = "csv"
	TransferJSON TransferFormat = "railkeeper-json"

	TransferVehicles        TransferArea = "vehicles"
	TransferAccessories     TransferArea = "accessories"
	TransferExhibitionLists TransferArea = "exhibitionLists"

	TransferJobDraft                 TransferJobState = "draft"
	TransferJobReading               TransferJobState = "reading"
	TransferJobReviewRequired        TransferJobState = "review_required"
	TransferJobReady                 TransferJobState = "ready"
	TransferJobRunning               TransferJobState = "running"
	TransferJobCompleted             TransferJobState = "completed"
	TransferJobCompletedWithWarnings TransferJobState = "completed_with_warnings"
	TransferJobFailed                TransferJobState = "failed"
	TransferJobCancelled             TransferJobState = "cancelled"

	TransferIssueWarning TransferIssueSeverity = "warning"
	TransferIssueError   TransferIssueSeverity = "error"
)

type DataTransferProfile struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Direction       TransferDirection `json:"direction"`
	Format          TransferFormat    `json:"format"`
	Areas           []TransferArea    `json:"areas"`
	Options         map[string]any    `json:"options"`
	Enabled         bool              `json:"enabled"`
	CreatedByUserID string            `json:"createdByUserId"`
	LastUsedAt      string            `json:"lastUsedAt,omitempty"`
	CreatedAt       string            `json:"createdAt"`
	UpdatedAt       string            `json:"updatedAt"`
}

type DataTransferJob struct {
	ID                string            `json:"id"`
	ProfileID         string            `json:"profileId"`
	ProfileName       string            `json:"profileName"`
	Direction         TransferDirection `json:"direction"`
	Format            TransferFormat    `json:"format"`
	Areas             []TransferArea    `json:"areas"`
	Options           map[string]any    `json:"options"`
	State             TransferJobState  `json:"state"`
	Stage             string            `json:"stage"`
	SourceName        string            `json:"sourceName"`
	SourceSHA256      string            `json:"sourceSha256"`
	PackageVersion    int               `json:"packageVersion"`
	Revision          int               `json:"revision"`
	TotalRecords      int               `json:"totalRecords"`
	ReadyRecords      int               `json:"readyRecords"`
	WarningRecords    int               `json:"warningRecords"`
	ErrorRecords      int               `json:"errorRecords"`
	Preview           map[string]any    `json:"preview"`
	CreatedByUserID   string            `json:"createdByUserId"`
	ConfirmedByUserID string            `json:"confirmedByUserId"`
	ConfirmedAt       string            `json:"confirmedAt"`
	CompletedAt       string            `json:"completedAt"`
	ResultMessage     string            `json:"resultMessage"`
	CreatedAt         string            `json:"createdAt"`
	UpdatedAt         string            `json:"updatedAt"`
}

type DataTransferJobFilter struct {
	ProfileID string             `json:"profileId"`
	Direction TransferDirection  `json:"direction"`
	States    []TransferJobState `json:"states"`
	Limit     int                `json:"limit"`
}

type DataTransferIssue struct {
	ID                 string                `json:"id"`
	JobID              string                `json:"jobId"`
	Area               TransferArea          `json:"area"`
	RecordKey          string                `json:"recordKey"`
	RowNumber          *int                  `json:"rowNumber"`
	Field              string                `json:"field"`
	Severity           TransferIssueSeverity `json:"severity"`
	Code               string                `json:"code"`
	Message            string                `json:"message"`
	ProposedResolution string                `json:"proposedResolution"`
	SelectedResolution string                `json:"selectedResolution"`
	CreatedAt          string                `json:"createdAt"`
	UpdatedAt          string                `json:"updatedAt"`
}

type DataTransferArtifact struct {
	ID           string `json:"id"`
	JobID        string `json:"jobId"`
	RelativePath string `json:"relativePath"`
	DisplayName  string `json:"displayName"`
	MIMEType     string `json:"mimeType"`
	SizeBytes    int64  `json:"sizeBytes"`
	SHA256       string `json:"sha256"`
	DeletedAt    string `json:"deletedAt"`
	CreatedAt    string `json:"createdAt"`
}

type DataTransferImportMutation struct {
	ExpectedState    TransferJobState
	ExpectedRevision int
	Job              DataTransferJob
	Issues           []DataTransferIssue
	ReplaceIssues    bool
	ProfileOptions   *DataTransferProfileOptionsMutation
}

type DataTransferProfileOptionsMutation struct {
	ProfileID         string
	ExpectedUpdatedAt string
	ExpectedOptions   map[string]any
	Options           map[string]any
}

type DataTransferImportPolicy struct {
	CanManageExhibitionLists bool
}

type DataTransferRepository interface {
	CreateProfile(context.Context, DataTransferProfile) (DataTransferProfile, error)
	UpdateProfile(context.Context, DataTransferProfile) (DataTransferProfile, error)
	ListProfiles(context.Context) ([]DataTransferProfile, error)
	GetProfile(context.Context, string) (DataTransferProfile, error)
	CreateJob(context.Context, DataTransferJob) (DataTransferJob, error)
	UpdateJob(context.Context, DataTransferJob) (DataTransferJob, error)
	GetJob(context.Context, string) (DataTransferJob, error)
	ListJobs(context.Context, DataTransferJobFilter) ([]DataTransferJob, error)
	DeleteJob(context.Context, string) error
	ReplaceIssues(context.Context, string, []DataTransferIssue) error
	ListIssues(context.Context, string) ([]DataTransferIssue, error)
	CreateArtifact(context.Context, DataTransferArtifact) (DataTransferArtifact, error)
	ListArtifacts(context.Context) ([]DataTransferArtifact, error)
	MarkArtifactDeleted(context.Context, string, string) error
}
