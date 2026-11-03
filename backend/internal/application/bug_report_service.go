package application

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"hash/fnv"
	"strings"
	"sync"
	"time"
)

const bugReportReconciliationLimit = 1_000

type BugReportService struct {
	collectDiagnostics func(context.Context) (BugReportServerDiagnostics, error)
	repository         BugReportSubmissionRepository
	issueClient        GitHubBugReportIssueClient
	assetClient        GitHubBugReportAssetClient
	contactEncryptor   BugReportContactEncryptor
	serverDiagnostics  BugReportServerDiagnostics
	baseLabels         []string
	locks              [64]sync.Mutex
}

func NewBugReportService(config BugReportServiceConfig) *BugReportService {
	baseLabels := append([]string{}, config.BaseLabels...)
	if len(baseLabels) == 0 {
		baseLabels = []string{"bug", "source:in-app"}
	}
	return &BugReportService{
		repository: config.Repository, issueClient: config.IssueClient, assetClient: config.AssetClient,
		collectDiagnostics: config.CollectDiagnostics,
		contactEncryptor:   config.ContactEncryptor, serverDiagnostics: config.ServerDiagnostics,
		baseLabels: baseLabels,
	}
}

func (service *BugReportService) IsExistingSubmission(ctx context.Context, reportID string) (bool, error) {
	if service == nil || service.repository == nil {
		return false, ErrBugReportingUnavailable
	}
	_, err := service.repository.Get(ctx, strings.TrimSpace(reportID))
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("check bug report submission: %w", err)
	}
	return true, nil
}

func (service *BugReportService) Submit(
	ctx context.Context,
	actorUserID string,
	request BugReportSubmissionRequest,
	admission ...BugReportAdmissionPolicy,
) (BugReportSubmissionResult, error) {
	if service == nil || service.repository == nil || service.issueClient == nil {
		return BugReportSubmissionResult{}, ErrBugReportingUnavailable
	}
	request.Input.ReportID = strings.TrimSpace(request.Input.ReportID)
	lock := service.lockForReport(request.Input.ReportID)
	lock.Lock()
	defer lock.Unlock()

	hasScreenshot := len(request.ScreenshotBytes) > 0
	validated, err := ValidateBugReportInput(request.Input, hasScreenshot)
	if err != nil {
		return BugReportSubmissionResult{}, err
	}
	var screenshot BugReportScreenshot
	if hasScreenshot {
		screenshot, err = SanitizeBugReportScreenshot(
			bytes.NewReader(request.ScreenshotBytes), request.ScreenshotMediaType,
		)
		if err != nil {
			return BugReportSubmissionResult{}, err
		}
	}

	encryptedContact := ""
	if validated.Email != "" {
		if service.contactEncryptor == nil {
			return BugReportSubmissionResult{}, ErrBugReportingUnavailable
		}
		encryptedContact, err = service.contactEncryptor.Encrypt(validated.Email)
		if err != nil {
			return BugReportSubmissionResult{}, fmt.Errorf("%w: contact encryption", ErrBugReportingUnavailable)
		}
	}

	policy := BugReportAdmissionPolicy{}
	if len(admission) > 0 {
		policy = admission[0]
	}
	created, admitted, submission, err := service.repository.Reserve(
		ctx, validated.ReportID, actorUserID, policy,
	)
	if err != nil {
		return BugReportSubmissionResult{}, fmt.Errorf("reserve bug report: %w", err)
	}
	if !admitted {
		return BugReportSubmissionResult{}, ErrBugReportRateLimited
	}
	if submission.ActorUserID != actorUserID {
		return BugReportSubmissionResult{}, ErrBugReportStateConflict
	}
	if submission.State == BugReportComplete {
		return bugReportResult(submission, false, ""), nil
	}
	diagnostics := BugReportServerDiagnostics{}
	if !validated.OmitDiagnostics {
		diagnostics, err = service.Diagnostics(ctx)
		if err != nil {
			return BugReportSubmissionResult{}, fmt.Errorf("%w: diagnostics unavailable", ErrBugReportingUnavailable)
		}
	}
	issueDraft, err := BuildBugReportIssue(BugReportIssueInput{
		Report: validated, ServerDiagnostics: diagnostics,
		EncryptedContact: encryptedContact, BaseLabels: service.baseLabels, HasScreenshot: hasScreenshot,
	})
	if err != nil {
		return BugReportSubmissionResult{}, err
	}
	submission, err = service.ensureIssue(ctx, submission, issueDraft)
	if err != nil {
		return BugReportSubmissionResult{}, err
	}

	if !hasScreenshot {
		if submission.State == BugReportPartial &&
			(submission.ErrorCategory == "github_asset" || submission.ErrorCategory == "github_comment") {
			return bugReportResult(submission, created, "Screenshot muss für den erneuten Versuch erneut gesendet werden."), nil
		}
		submission, err = service.updateSubmission(ctx, submission, BugReportComplete, "")
		if err != nil {
			return BugReportSubmissionResult{}, err
		}
		return bugReportResult(submission, created, ""), nil
	}

	if service.assetClient == nil {
		submission = service.recordPartial(ctx, submission, "github_asset")
		return bugReportResult(submission, created, "Screenshot-Upload ist derzeit nicht verfügbar."), nil
	}
	assetPath := bugReportAssetPath(submission.CreatedAt, validated.ReportID, screenshot.Extension)
	submission, err = service.ensureAsset(ctx, submission, assetPath, screenshot.Bytes)
	if err != nil {
		return bugReportResult(submission, created, "Der Screenshot konnte noch nicht hochgeladen werden."), nil
	}
	submission, err = service.ensureScreenshotComment(ctx, submission)
	if err != nil {
		return bugReportResult(submission, created, "Der Screenshot-Link konnte noch nicht ergänzt werden."), nil
	}
	submission, err = service.updateSubmission(ctx, submission, BugReportComplete, "")
	if err != nil {
		return BugReportSubmissionResult{}, err
	}
	return bugReportResult(submission, created, ""), nil
}

func (service *BugReportService) ensureIssue(
	ctx context.Context,
	submission BugReportSubmission,
	draft BugReportIssue,
) (BugReportSubmission, error) {
	if submission.GitHubIssueNumber > 0 && submission.GitHubIssueURL != "" {
		return submission, nil
	}
	marker := bugReportIssueMarker(submission.ReportID)
	issues, err := service.issueClient.ListRecentIssues(
		ctx, submission.CreatedAt.Add(-time.Minute), bugReportReconciliationLimit,
	)
	if err != nil {
		return BugReportSubmission{}, ErrBugReportGitHubUnavailable
	}
	for _, issue := range issues {
		if strings.Contains(issue.Body, marker) {
			return service.updateIssueCreated(ctx, submission, issue)
		}
	}
	issue, err := service.issueClient.CreateIssue(ctx, draft)
	if err != nil {
		_ = service.recordPartial(ctx, submission, "github_issue")
		return BugReportSubmission{}, ErrBugReportGitHubUnavailable
	}
	return service.updateIssueCreated(ctx, submission, issue)
}

func (service *BugReportService) updateIssueCreated(
	ctx context.Context,
	submission BugReportSubmission,
	issue GitHubBugReportIssue,
) (BugReportSubmission, error) {
	return service.repository.Update(ctx, submission.ReportID, BugReportSubmissionUpdate{
		State: BugReportIssueCreated, GitHubIssueNumber: issue.Number, GitHubIssueURL: issue.URL,
		AssetURL: submission.AssetURL, AssetSHA: submission.AssetSHA,
	})
}

func (service *BugReportService) ensureAsset(
	ctx context.Context,
	submission BugReportSubmission,
	assetPath string,
	content []byte,
) (BugReportSubmission, error) {
	if submission.AssetURL != "" && submission.AssetSHA != "" {
		return submission, nil
	}
	found, asset, err := service.assetClient.FindFile(ctx, assetPath)
	if err != nil {
		return service.recordPartial(ctx, submission, "github_asset"), err
	}
	if !found {
		message := "Add RailKeeper bug report screenshot " + submission.ReportID +
			" for issue #" + fmt.Sprint(submission.GitHubIssueNumber)
		asset, err = service.assetClient.CreateFile(ctx, assetPath, message, content)
		if err != nil {
			return service.recordPartial(ctx, submission, "github_asset"), err
		}
	}
	updated, err := service.repository.Update(ctx, submission.ReportID, BugReportSubmissionUpdate{
		State: BugReportAssetUploaded, GitHubIssueNumber: submission.GitHubIssueNumber,
		GitHubIssueURL: submission.GitHubIssueURL, AssetURL: asset.URL, AssetSHA: asset.CommitSHA,
	})
	if err != nil {
		return submission, err
	}
	return updated, nil
}

func (service *BugReportService) ensureScreenshotComment(
	ctx context.Context,
	submission BugReportSubmission,
) (BugReportSubmission, error) {
	marker := bugReportScreenshotMarker(submission.ReportID)
	comments, err := service.issueClient.ListIssueComments(
		ctx, submission.GitHubIssueNumber, bugReportReconciliationLimit,
	)
	if err != nil {
		return service.recordPartial(ctx, submission, "github_comment"), err
	}
	for _, comment := range comments {
		if strings.Contains(comment.Body, marker) {
			return submission, nil
		}
	}
	body := "Der folgende Screenshot ist öffentlich sichtbar. / The following screenshot is publicly visible.\n\n" +
		"![RailKeeper bug report screenshot](" + submission.AssetURL + ")\n\n" + marker
	if _, err := service.issueClient.CreateIssueComment(ctx, submission.GitHubIssueNumber, body); err != nil {
		return service.recordPartial(ctx, submission, "github_comment"), err
	}
	return submission, nil
}

func (service *BugReportService) updateSubmission(
	ctx context.Context,
	submission BugReportSubmission,
	state BugReportSubmissionState,
	errorCategory string,
) (BugReportSubmission, error) {
	return service.repository.Update(ctx, submission.ReportID, BugReportSubmissionUpdate{
		State: state, GitHubIssueNumber: submission.GitHubIssueNumber, GitHubIssueURL: submission.GitHubIssueURL,
		AssetURL: submission.AssetURL, AssetSHA: submission.AssetSHA, ErrorCategory: errorCategory,
	})
}

func (service *BugReportService) recordPartial(
	ctx context.Context,
	submission BugReportSubmission,
	category string,
) BugReportSubmission {
	updated, err := service.updateSubmission(ctx, submission, BugReportPartial, category)
	if err != nil {
		return submission
	}
	return updated
}

func (service *BugReportService) lockForReport(reportID string) *sync.Mutex {
	hasher := fnv.New32a()
	_, _ = hasher.Write([]byte(reportID))
	return &service.locks[hasher.Sum32()%uint32(len(service.locks))]
}

func bugReportAssetPath(createdAt time.Time, reportID string, extension string) string {
	createdAt = createdAt.UTC()
	return fmt.Sprintf("reports/%04d/%02d/%s.%s", createdAt.Year(), createdAt.Month(), reportID, extension)
}

func bugReportIssueMarker(reportID string) string {
	return "<!-- railkeeper-bug-report:v1:" + reportID + " -->"
}

func bugReportScreenshotMarker(reportID string) string {
	return "<!-- railkeeper-screenshot:v1:" + reportID + " -->"
}

func bugReportResult(submission BugReportSubmission, created bool, warning string) BugReportSubmissionResult {
	status := submission.State
	if status != BugReportComplete {
		status = BugReportPartial
	}
	return BugReportSubmissionResult{
		ReportID: submission.ReportID, IssueNumber: submission.GitHubIssueNumber,
		IssueURL: submission.GitHubIssueURL, Status: status, Warning: warning, Created: created,
	}
}

func (service *BugReportService) Diagnostics(ctx context.Context) (BugReportServerDiagnostics, error) {
	if service.collectDiagnostics != nil {
		return service.collectDiagnostics(ctx)
	}
	return service.serverDiagnostics, nil
}
