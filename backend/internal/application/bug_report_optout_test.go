package application_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"railkeeper/backend/internal/application"
)

func TestBugReportOptOutSkipsCollectionAndDiscardsSuppliedDiagnostics(t *testing.T) {
	issues := &fakeBugReportIssueClient{}
	service := application.NewBugReportService(application.BugReportServiceConfig{
		Repository: newMemoryBugReportRepository(), IssueClient: issues,
		CollectDiagnostics: func(context.Context) (application.BugReportServerDiagnostics, error) {
			t.Fatal("diagnostics collected after opt-out")
			return application.BugReportServerDiagnostics{}, nil
		},
	})
	input := validBugReportInput()
	input.OmitDiagnostics = true
	input.ClientDiagnostics.BrowserFamily = "secret-browser-value"
	if _, err := service.Submit(t.Context(), "user-1", application.BugReportSubmissionRequest{Input: input}); err != nil {
		t.Fatal(err)
	}
	body := issues.issues[0].Body
	for _, forbidden := range []string{"secret-browser-value", "browserFamily", "goVersion", "heapBytes", "### Server"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("opted-out issue contains %q", forbidden)
		}
	}
	if !strings.Contains(body, "Diagnostics were disabled") || !strings.Contains(body, input.Description) {
		t.Fatal("report must retain the description and explain omitted diagnostics")
	}
}

func TestBugReportIncludesFreshServerDiagnosticsWhenEnabled(t *testing.T) {
	issues := &fakeBugReportIssueClient{}
	calls := 0
	service := application.NewBugReportService(application.BugReportServiceConfig{
		Repository: newMemoryBugReportRepository(), IssueClient: issues,
		CollectDiagnostics: func(context.Context) (application.BugReportServerDiagnostics, error) {
			calls++
			return application.BugReportServerDiagnostics{SQLiteVersion: "test-sqlite", HeapBytes: 1234}, nil
		},
	})
	if _, err := service.Submit(t.Context(), "user-1", application.BugReportSubmissionRequest{Input: validBugReportInput()}); err != nil {
		t.Fatal(err)
	}
	if calls != 1 || !strings.Contains(issues.issues[0].Body, "test-sqlite") ||
		!strings.Contains(issues.issues[0].Body, `"heapBytes": 1234`) {
		t.Fatal("enabled reports must include fresh server diagnostics")
	}
}

type unavailableReconciliationClient struct{ *fakeBugReportIssueClient }

func (client unavailableReconciliationClient) ListRecentIssues(context.Context, time.Time, int) ([]application.GitHubBugReportIssue, error) {
	return nil, errors.New("unavailable")
}

func TestBugReportDoesNotCreateIssueWhenReconciliationIsUnavailable(t *testing.T) {
	issues := &fakeBugReportIssueClient{}
	service := application.NewBugReportService(application.BugReportServiceConfig{
		Repository: newMemoryBugReportRepository(), IssueClient: unavailableReconciliationClient{issues},
	})
	_, err := service.Submit(t.Context(), "user-1", application.BugReportSubmissionRequest{Input: validBugReportInput()})
	if !errors.Is(err, application.ErrBugReportGitHubUnavailable) || issues.createCalls != 0 {
		t.Fatalf("uncertain reconciliation must prevent duplicate issues: calls=%d err=%v", issues.createCalls, err)
	}
}
