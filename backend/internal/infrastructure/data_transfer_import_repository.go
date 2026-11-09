package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"railkeeper/backend/internal/transfer"
)

func (repository *DataTransferRepository) ReplaceIssues(
	ctx context.Context,
	jobID string,
	issues []transfer.DataTransferIssue,
) error {
	now := timestamp()
	err := repository.withTx(ctx, func(tx *sql.Tx) error {
		return replaceDataTransferIssues(ctx, tx, jobID, issues, now)
	})
	if err != nil {
		return fmt.Errorf("replace transfer issues: %w", err)
	}
	return nil
}

func (repository *DataTransferRepository) CompareAndUpdateImportJob(
	ctx context.Context,
	mutation transfer.DataTransferImportMutation,
) (transfer.DataTransferJob, error) {
	job := mutation.Job
	areas, err := encodeTransferAreas(job.Areas)
	if err != nil {
		return transfer.DataTransferJob{}, err
	}
	options, err := encodeTransferOptions(job.Options)
	if err != nil {
		return transfer.DataTransferJob{}, err
	}
	preview, err := encodeTransferOptions(job.Preview)
	if err != nil {
		return transfer.DataTransferJob{}, fmt.Errorf("encode transfer preview: %w", err)
	}
	profileOptions := ""
	expectedProfileOptions := ""
	if mutation.ProfileOptions != nil {
		profileOptions, err = encodeTransferOptions(mutation.ProfileOptions.Options)
		if err != nil {
			return transfer.DataTransferJob{}, fmt.Errorf("encode transfer profile options: %w", err)
		}
		expectedProfileOptions, err = encodeTransferOptions(mutation.ProfileOptions.ExpectedOptions)
		if err != nil {
			return transfer.DataTransferJob{}, fmt.Errorf("encode expected transfer profile options: %w", err)
		}
	}
	now := timestamp()
	err = repository.withTx(ctx, func(tx *sql.Tx) error {
		result, err := tx.ExecContext(ctx, `
UPDATE data_transfer_jobs
SET profile_id=NULLIF(?, ''), profile_name=?, format=?, areas_json=?, options_json=?, state=?, stage=?,
    source_name=?, source_sha256=?, package_version=?, total_records=?, ready_records=?, warning_records=?,
    error_records=?, preview_json=?, confirmed_by_user_id=NULLIF(?, ''), confirmed_at=NULLIF(?, ''),
    completed_at=NULLIF(?, ''), result_message=?, revision=revision+1, updated_at=?
WHERE id=? AND direction=? AND state=? AND revision=?`, job.ProfileID, job.ProfileName, job.Format, areas, options,
			job.State, job.Stage, job.SourceName, job.SourceSHA256, job.PackageVersion, job.TotalRecords,
			job.ReadyRecords, job.WarningRecords, job.ErrorRecords, preview, job.ConfirmedByUserID, job.ConfirmedAt,
			job.CompletedAt, job.ResultMessage, now, job.ID, transfer.TransferImport, mutation.ExpectedState,
			mutation.ExpectedRevision)
		if err != nil {
			return fmt.Errorf("compare and update transfer import job: %w", err)
		}
		updated, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("read transfer import update result: %w", err)
		}
		if updated == 0 {
			var exists int
			if err := tx.QueryRowContext(ctx,
				`SELECT EXISTS(SELECT 1 FROM data_transfer_jobs WHERE id=?)`, job.ID).Scan(&exists); err != nil {
				return fmt.Errorf("check transfer import job: %w", err)
			}
			if exists == 0 {
				return fmt.Errorf("compare and update transfer import job: %w: %w", transfer.ErrDataTransferNotFound, sql.ErrNoRows)
			}
			return transfer.ErrDataTransferConflict
		}
		if mutation.ReplaceIssues {
			if err := replaceDataTransferIssues(ctx, tx, job.ID, mutation.Issues, now); err != nil {
				return err
			}
		}
		if mutation.ProfileOptions != nil {
			result, err := tx.ExecContext(ctx, `
UPDATE data_transfer_profiles SET options_json=?, updated_at=?
WHERE id=? AND updated_at=? AND options_json=?`, profileOptions, now, mutation.ProfileOptions.ProfileID,
				mutation.ProfileOptions.ExpectedUpdatedAt, expectedProfileOptions)
			if err != nil {
				return fmt.Errorf("update transfer profile mapping: %w", err)
			}
			updated, err := result.RowsAffected()
			if err != nil {
				return fmt.Errorf("read transfer profile mapping update result: %w", err)
			}
			if updated == 0 {
				return transfer.ErrDataTransferConflict
			}
		}
		return nil
	})
	if err != nil {
		return transfer.DataTransferJob{}, err
	}
	return repository.GetJob(ctx, job.ID)
}

func replaceDataTransferIssues(
	ctx context.Context,
	tx *sql.Tx,
	jobID string,
	issues []transfer.DataTransferIssue,
	now string,
) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM data_transfer_job_issues WHERE job_id=?`, jobID); err != nil {
		return fmt.Errorf("delete transfer issues: %w", err)
	}
	for _, issue := range issues {
		issueID := issue.ID
		if issueID == "" {
			issueID = randomID()
		}
		createdAt := issue.CreatedAt
		if createdAt == "" {
			createdAt = now
		}
		updatedAt := issue.UpdatedAt
		if updatedAt == "" {
			updatedAt = now
		}
		var rowNumber any
		if issue.RowNumber != nil {
			rowNumber = *issue.RowNumber
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO data_transfer_job_issues(
  id, job_id, area, record_key, row_number, field, severity, code, message, proposed_resolution,
  selected_resolution, created_at, updated_at
) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, issueID, jobID, issue.Area, issue.RecordKey, rowNumber,
			issue.Field, issue.Severity, issue.Code, issue.Message, issue.ProposedResolution,
			issue.SelectedResolution, createdAt, updatedAt); err != nil {
			return fmt.Errorf("insert transfer issue: %w", err)
		}
	}
	return nil
}

func (repository *DataTransferRepository) ListIssues(
	ctx context.Context,
	jobID string,
) ([]transfer.DataTransferIssue, error) {
	rows, err := repository.db.QueryContext(ctx, dataTransferIssueSelect+`
WHERE job_id=?
ORDER BY created_at, id`, jobID)
	if err != nil {
		return nil, fmt.Errorf("list transfer issues: %w", err)
	}
	defer func() { _ = rows.Close() }()
	issues := []transfer.DataTransferIssue{}
	for rows.Next() {
		issue, err := scanDataTransferIssue(rows)
		if err != nil {
			return nil, fmt.Errorf("scan transfer issue: %w", err)
		}
		issues = append(issues, issue)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate transfer issues: %w", err)
	}
	return issues, nil
}
