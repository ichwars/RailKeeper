package infrastructure_test

import (
	"errors"
	"testing"
	"time"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/infrastructure"
)

func TestBugReportRepositoryReservesAndLoadsSubmission(t *testing.T) {
	repository := infrastructure.NewBugReportRepository(testDB(t))

	created, admitted, reserved, err := repository.Reserve(
		t.Context(), "018f6f9e-3c68-7df2-a0a4-1d6c37fbb201", "user-1", application.BugReportAdmissionPolicy{},
	)
	if err != nil {
		t.Fatal(err)
	}
	if !created || !admitted {
		t.Fatalf("first reservation created=%t admitted=%t", created, admitted)
	}
	if reserved.State != application.BugReportReserved || reserved.ActorUserID != "user-1" {
		t.Fatalf("unexpected reservation: %#v", reserved)
	}

	created, admitted, duplicate, err := repository.Reserve(
		t.Context(), reserved.ReportID, "user-2", application.BugReportAdmissionPolicy{},
	)
	if err != nil {
		t.Fatal(err)
	}
	if created || !admitted || duplicate.ActorUserID != "user-1" {
		t.Fatalf("duplicate reservation changed ownership: created=%v admitted=%v submission=%#v",
			created, admitted, duplicate)
	}
}

func TestBugReportRepositoryAtomicallyAdmitsOnlyNewReservedIDs(t *testing.T) {
	db := testDB(t)
	if _, err := db.ExecContext(t.Context(), `
INSERT INTO rate_limit_attempts(id, scope, key, attempted_at) VALUES(?, ?, ?, ?)`,
		"long-window-attempt", "other-scope", "user-1",
		time.Now().UTC().Add(-90*time.Minute).Format(time.RFC3339Nano)); err != nil {
		t.Fatal(err)
	}
	repository := infrastructure.NewBugReportRepository(db)
	policy := application.BugReportAdmissionPolicy{Scope: "bug_report_new", Limit: 1, Window: time.Hour}

	created, admitted, first, err := repository.Reserve(
		t.Context(), "018f6f9e-3c68-7df2-a0a4-1d6c37fbb211", "user-1", policy,
	)
	if err != nil || !created || !admitted {
		t.Fatalf("first reservation created=%t admitted=%t err=%v", created, admitted, err)
	}
	created, admitted, duplicate, err := repository.Reserve(t.Context(), first.ReportID, "user-1", policy)
	if err != nil || created || !admitted || duplicate.ReportID != first.ReportID {
		t.Fatalf("duplicate reservation created=%t admitted=%t submission=%#v err=%v",
			created, admitted, duplicate, err)
	}
	created, admitted, _, err = repository.Reserve(
		t.Context(), "018f6f9e-3c68-7df2-a0a4-1d6c37fbb212", "user-1", policy,
	)
	if err != nil || created || admitted {
		t.Fatalf("denied reservation created=%t admitted=%t err=%v", created, admitted, err)
	}

	var submissions int
	if err := db.QueryRow(`SELECT COUNT(*) FROM bug_report_submissions`).Scan(&submissions); err != nil {
		t.Fatal(err)
	}
	var attempts int
	if err := db.QueryRow(`SELECT COUNT(*) FROM rate_limit_attempts WHERE scope='bug_report_new'`).Scan(&attempts); err != nil {
		t.Fatal(err)
	}
	if submissions != 1 || attempts != 1 {
		t.Fatalf("submissions=%d attempts=%d, want 1 and 1", submissions, attempts)
	}
	var otherAttempts int
	if err := db.QueryRow(`SELECT COUNT(*) FROM rate_limit_attempts WHERE scope='other-scope'`).Scan(&otherAttempts); err != nil {
		t.Fatal(err)
	}
	if otherAttempts != 1 {
		t.Fatal("report admission must not prune counters from another scope")
	}
}

func TestBugReportRepositoryPersistsForwardSubmissionProgress(t *testing.T) {
	repository := infrastructure.NewBugReportRepository(testDB(t))
	const reportID = "018f6f9e-3c68-7df2-a0a4-1d6c37fbb202"
	if _, _, _, err := repository.Reserve(
		t.Context(), reportID, "user-1", application.BugReportAdmissionPolicy{},
	); err != nil {
		t.Fatal(err)
	}

	issue, err := repository.Update(t.Context(), reportID, application.BugReportSubmissionUpdate{
		State:             application.BugReportIssueCreated,
		GitHubIssueNumber: 184,
		GitHubIssueURL:    "https://github.com/ichwars/RailKeeper/issues/184",
	})
	if err != nil {
		t.Fatal(err)
	}
	asset, err := repository.Update(t.Context(), reportID, application.BugReportSubmissionUpdate{
		State:             application.BugReportAssetUploaded,
		GitHubIssueNumber: issue.GitHubIssueNumber,
		GitHubIssueURL:    issue.GitHubIssueURL,
		AssetURL:          "https://raw.githubusercontent.com/assets/sha/reports/2026/08/report.png",
		AssetSHA:          "0123456789abcdef",
	})
	if err != nil {
		t.Fatal(err)
	}
	complete, err := repository.Update(t.Context(), reportID, application.BugReportSubmissionUpdate{
		State:             application.BugReportComplete,
		GitHubIssueNumber: asset.GitHubIssueNumber,
		GitHubIssueURL:    asset.GitHubIssueURL,
		AssetURL:          asset.AssetURL,
		AssetSHA:          asset.AssetSHA,
	})
	if err != nil {
		t.Fatal(err)
	}
	if complete.State != application.BugReportComplete || complete.AssetSHA != "0123456789abcdef" {
		t.Fatalf("unexpected completed submission: %#v", complete)
	}
	if complete.CreatedAt.IsZero() || complete.UpdatedAt.Before(complete.CreatedAt) {
		t.Fatalf("invalid timestamps: %#v", complete)
	}
}

func TestBugReportRepositoryRejectsBackwardStateTransition(t *testing.T) {
	repository := infrastructure.NewBugReportRepository(testDB(t))
	const reportID = "018f6f9e-3c68-7df2-a0a4-1d6c37fbb203"
	if _, _, _, err := repository.Reserve(
		t.Context(), reportID, "user-1", application.BugReportAdmissionPolicy{},
	); err != nil {
		t.Fatal(err)
	}
	if _, err := repository.Update(t.Context(), reportID, application.BugReportSubmissionUpdate{
		State: application.BugReportComplete,
	}); err != nil {
		t.Fatal(err)
	}
	_, err := repository.Update(t.Context(), reportID, application.BugReportSubmissionUpdate{
		State: application.BugReportReserved,
	})
	if !errors.Is(err, application.ErrBugReportStateConflict) {
		t.Fatalf("backward transition error = %v, want state conflict", err)
	}
}

func TestBugReportRepositoryRecordsOnlySafeFailureCategory(t *testing.T) {
	repository := infrastructure.NewBugReportRepository(testDB(t))
	const reportID = "018f6f9e-3c68-7df2-a0a4-1d6c37fbb204"
	if _, _, _, err := repository.Reserve(
		t.Context(), reportID, "user-1", application.BugReportAdmissionPolicy{},
	); err != nil {
		t.Fatal(err)
	}
	partial, err := repository.Update(t.Context(), reportID, application.BugReportSubmissionUpdate{
		State:         application.BugReportPartial,
		ErrorCategory: "github_asset",
	})
	if err != nil {
		t.Fatal(err)
	}
	if partial.ErrorCategory != "github_asset" {
		t.Fatalf("error category = %q", partial.ErrorCategory)
	}
	_, err = repository.Update(t.Context(), reportID, application.BugReportSubmissionUpdate{
		State:         application.BugReportPartial,
		ErrorCategory: "token=secret response body",
	})
	if !errors.Is(err, application.ErrBugReportStateConflict) {
		t.Fatalf("unsafe category error = %v, want state conflict", err)
	}
}
