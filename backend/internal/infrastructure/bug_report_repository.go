package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"time"

	"railkeeper/backend/internal/application"
)

var bugReportErrorCategoryPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)

type BugReportRepository struct {
	db *sql.DB
}

func NewBugReportRepository(db *sql.DB) *BugReportRepository {
	return &BugReportRepository{db: db}
}

func (repository *BugReportRepository) Reserve(
	ctx context.Context,
	reportID string,
	actorUserID string,
	admission application.BugReportAdmissionPolicy,
) (created bool, admitted bool, submission application.BugReportSubmission, err error) {
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return false, false, application.BugReportSubmission{}, fmt.Errorf("begin bug report reservation: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	now := time.Now().UTC()
	reservationTimestamp := now.Format(time.RFC3339)
	result, err := tx.ExecContext(ctx, `
INSERT INTO bug_report_submissions(report_id, actor_user_id, state, created_at, updated_at)
VALUES(?, ?, ?, ?, ?)
ON CONFLICT(report_id) DO NOTHING`, reportID, actorUserID, application.BugReportReserved,
		reservationTimestamp, reservationTimestamp)
	if err != nil {
		return false, false, application.BugReportSubmission{}, fmt.Errorf("reserve bug report: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, false, application.BugReportSubmission{}, fmt.Errorf("read bug report reservation: %w", err)
	}
	submission, err = getBugReportSubmission(ctx, tx, reportID)
	if err != nil {
		return false, false, application.BugReportSubmission{}, err
	}
	if rows == 0 {
		if err := tx.Commit(); err != nil {
			return false, false, application.BugReportSubmission{}, fmt.Errorf("commit duplicate bug report: %w", err)
		}
		return false, true, submission, nil
	}

	if admission.Scope != "" && admission.Limit > 0 && admission.Window > 0 {
		cutoff := now.Add(-admission.Window).Format(time.RFC3339Nano)
		if _, err := tx.ExecContext(ctx, `DELETE FROM rate_limit_attempts WHERE attempted_at <= ?`, cutoff); err != nil {
			return false, false, application.BugReportSubmission{}, fmt.Errorf("cleanup bug report admission: %w", err)
		}
		var count int
		if err := tx.QueryRowContext(ctx, `
SELECT COUNT(*) FROM rate_limit_attempts WHERE scope=? AND key=? AND attempted_at > ?`,
			admission.Scope, actorUserID, cutoff).Scan(&count); err != nil {
			return false, false, application.BugReportSubmission{}, fmt.Errorf("count bug report admission: %w", err)
		}
		if count >= admission.Limit {
			if err := tx.Rollback(); err != nil {
				return false, false, application.BugReportSubmission{}, fmt.Errorf("rollback denied bug report: %w", err)
			}
			return false, false, application.BugReportSubmission{}, nil
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO rate_limit_attempts(id, scope, key, attempted_at) VALUES(?, ?, ?, ?)`,
			randomID(), admission.Scope, actorUserID, now.Format(time.RFC3339Nano)); err != nil {
			return false, false, application.BugReportSubmission{}, fmt.Errorf("record bug report admission: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return false, false, application.BugReportSubmission{}, fmt.Errorf("commit bug report reservation: %w", err)
	}
	return true, true, submission, nil
}

func (repository *BugReportRepository) Get(
	ctx context.Context,
	reportID string,
) (application.BugReportSubmission, error) {
	return getBugReportSubmission(ctx, repository.db, reportID)
}

type bugReportQueryer interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func getBugReportSubmission(
	ctx context.Context,
	queryer bugReportQueryer,
	reportID string,
) (application.BugReportSubmission, error) {
	row := queryer.QueryRowContext(ctx, `
SELECT report_id, actor_user_id, state, COALESCE(github_issue_number, 0), github_issue_url,
       asset_url, asset_sha, error_category, created_at, updated_at
FROM bug_report_submissions
WHERE report_id=?`, reportID)
	submission, err := scanBugReportSubmission(row)
	if errors.Is(err, sql.ErrNoRows) {
		return application.BugReportSubmission{}, fmt.Errorf("get bug report submission: %w", sql.ErrNoRows)
	}
	if err != nil {
		return application.BugReportSubmission{}, fmt.Errorf("get bug report submission: %w", err)
	}
	return submission, nil
}

func (repository *BugReportRepository) Update(
	ctx context.Context,
	reportID string,
	update application.BugReportSubmissionUpdate,
) (application.BugReportSubmission, error) {
	if !validBugReportErrorCategory(update.ErrorCategory) {
		return application.BugReportSubmission{}, application.ErrBugReportStateConflict
	}
	current, err := repository.Get(ctx, reportID)
	if err != nil {
		return application.BugReportSubmission{}, err
	}
	if !validBugReportTransition(current.State, update.State) {
		return application.BugReportSubmission{}, application.ErrBugReportStateConflict
	}
	result, err := repository.db.ExecContext(ctx, `
UPDATE bug_report_submissions
SET state=?, github_issue_number=NULLIF(?, 0), github_issue_url=?, asset_url=?, asset_sha=?,
    error_category=?, updated_at=?
WHERE report_id=? AND state=?`, update.State, update.GitHubIssueNumber, update.GitHubIssueURL,
		update.AssetURL, update.AssetSHA, update.ErrorCategory, timestamp(), reportID, current.State)
	if err != nil {
		return application.BugReportSubmission{}, fmt.Errorf("update bug report submission: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return application.BugReportSubmission{}, fmt.Errorf("read bug report update: %w", err)
	}
	if rows != 1 {
		return application.BugReportSubmission{}, application.ErrBugReportStateConflict
	}
	return repository.Get(ctx, reportID)
}

func validBugReportTransition(current application.BugReportSubmissionState, next application.BugReportSubmissionState) bool {
	if current == next {
		return true
	}
	if current == application.BugReportComplete {
		return false
	}
	if next == application.BugReportPartial {
		return true
	}
	rank := map[application.BugReportSubmissionState]int{
		application.BugReportReserved:      0,
		application.BugReportIssueCreated:  1,
		application.BugReportAssetUploaded: 2,
		application.BugReportComplete:      3,
	}
	currentRank, currentKnown := rank[current]
	nextRank, nextKnown := rank[next]
	if current == application.BugReportPartial {
		currentRank = 0
		currentKnown = true
	}
	return currentKnown && nextKnown && nextRank > currentRank
}

func validBugReportErrorCategory(category string) bool {
	return category == "" || bugReportErrorCategoryPattern.MatchString(category)
}

type bugReportScanner interface {
	Scan(dest ...any) error
}

func scanBugReportSubmission(scanner bugReportScanner) (application.BugReportSubmission, error) {
	var submission application.BugReportSubmission
	var createdAt string
	var updatedAt string
	if err := scanner.Scan(&submission.ReportID, &submission.ActorUserID, &submission.State,
		&submission.GitHubIssueNumber, &submission.GitHubIssueURL, &submission.AssetURL,
		&submission.AssetSHA, &submission.ErrorCategory, &createdAt, &updatedAt); err != nil {
		return application.BugReportSubmission{}, err
	}
	var err error
	submission.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return application.BugReportSubmission{}, fmt.Errorf("parse bug report created timestamp: %w", err)
	}
	submission.UpdatedAt, err = time.Parse(time.RFC3339, updatedAt)
	if err != nil {
		return application.BugReportSubmission{}, fmt.Errorf("parse bug report updated timestamp: %w", err)
	}
	return submission, nil
}
