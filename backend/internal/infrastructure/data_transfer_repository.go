package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"railkeeper/backend/internal/application"
)

type DataTransferRepository struct {
	db *sql.DB
}

func NewDataTransferRepository(db *sql.DB) *DataTransferRepository {
	return &DataTransferRepository{db: db}
}

func (repository *DataTransferRepository) CreateProfile(
	ctx context.Context,
	profile application.DataTransferProfile,
) (application.DataTransferProfile, error) {
	areas, err := encodeTransferAreas(profile.Areas)
	if err != nil {
		return application.DataTransferProfile{}, err
	}
	options, err := encodeTransferOptions(profile.Options)
	if err != nil {
		return application.DataTransferProfile{}, err
	}
	now := timestamp()
	profile.ID = randomID()
	profile.CreatedAt = now
	profile.UpdatedAt = now
	if _, err := repository.db.ExecContext(ctx, `
INSERT INTO data_transfer_profiles(
  id, name, direction, format, areas_json, options_json, enabled, created_by_user_id, last_used_at,
  created_at, updated_at
) VALUES(?, ?, ?, ?, ?, ?, ?, NULLIF(?, ''), NULLIF(?, ''), ?, ?)`, profile.ID, profile.Name,
		profile.Direction, profile.Format, areas, options, boolToInt(profile.Enabled), profile.CreatedByUserID,
		profile.LastUsedAt, now, now); err != nil {
		return application.DataTransferProfile{}, fmt.Errorf("create transfer profile: %w", err)
	}
	return repository.GetProfile(ctx, profile.ID)
}

func (repository *DataTransferRepository) UpdateProfile(
	ctx context.Context,
	profile application.DataTransferProfile,
) (application.DataTransferProfile, error) {
	areas, err := encodeTransferAreas(profile.Areas)
	if err != nil {
		return application.DataTransferProfile{}, err
	}
	options, err := encodeTransferOptions(profile.Options)
	if err != nil {
		return application.DataTransferProfile{}, err
	}
	result, err := repository.db.ExecContext(ctx, `
UPDATE data_transfer_profiles
SET name=?, direction=?, format=?, areas_json=?, options_json=?, enabled=?, last_used_at=NULLIF(?, ''),
    updated_at=?
WHERE id=?`, profile.Name, profile.Direction, profile.Format, areas, options, boolToInt(profile.Enabled),
		profile.LastUsedAt, timestamp(), profile.ID)
	if err != nil {
		return application.DataTransferProfile{}, fmt.Errorf("update transfer profile: %w", err)
	}
	if err := requireDataTransferUpdate(result, "update transfer profile"); err != nil {
		return application.DataTransferProfile{}, err
	}
	return repository.GetProfile(ctx, profile.ID)
}

func (repository *DataTransferRepository) ListProfiles(
	ctx context.Context,
) ([]application.DataTransferProfile, error) {
	rows, err := repository.db.QueryContext(ctx, dataTransferProfileSelect+`
ORDER BY name COLLATE NOCASE, id`)
	if err != nil {
		return nil, fmt.Errorf("list transfer profiles: %w", err)
	}
	defer func() { _ = rows.Close() }()
	profiles := []application.DataTransferProfile{}
	for rows.Next() {
		profile, err := scanDataTransferProfile(rows)
		if err != nil {
			return nil, fmt.Errorf("scan transfer profile: %w", err)
		}
		profiles = append(profiles, profile)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate transfer profiles: %w", err)
	}
	return profiles, nil
}

func (repository *DataTransferRepository) GetProfile(
	ctx context.Context,
	id string,
) (application.DataTransferProfile, error) {
	profile, err := scanDataTransferProfile(repository.db.QueryRowContext(ctx, dataTransferProfileSelect+` WHERE id=?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return application.DataTransferProfile{}, fmt.Errorf("get transfer profile: %w", sql.ErrNoRows)
	}
	if err != nil {
		return application.DataTransferProfile{}, fmt.Errorf("get transfer profile: %w", err)
	}
	return profile, nil
}

func (repository *DataTransferRepository) CreateJob(
	ctx context.Context,
	job application.DataTransferJob,
) (application.DataTransferJob, error) {
	areas, err := encodeTransferAreas(job.Areas)
	if err != nil {
		return application.DataTransferJob{}, err
	}
	options, err := encodeTransferOptions(job.Options)
	if err != nil {
		return application.DataTransferJob{}, err
	}
	preview, err := encodeTransferOptions(job.Preview)
	if err != nil {
		return application.DataTransferJob{}, fmt.Errorf("encode transfer preview: %w", err)
	}
	now := timestamp()
	job.ID = randomID()
	job.CreatedAt = now
	job.UpdatedAt = now
	if _, err := repository.db.ExecContext(ctx, `
INSERT INTO data_transfer_jobs(
  id, profile_id, profile_name, direction, format, areas_json, options_json, state, stage, source_name,
  source_sha256, package_version, total_records, ready_records, warning_records, error_records, preview_json,
  created_by_user_id, confirmed_by_user_id, confirmed_at, completed_at, result_message, created_at, updated_at
) VALUES(?, NULLIF(?, ''), ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NULLIF(?, ''), NULLIF(?, ''),
  NULLIF(?, ''), NULLIF(?, ''), ?, ?, ?)`, job.ID, job.ProfileID, job.ProfileName, job.Direction, job.Format,
		areas, options, job.State, job.Stage, job.SourceName, job.SourceSHA256, job.PackageVersion,
		job.TotalRecords, job.ReadyRecords, job.WarningRecords, job.ErrorRecords, preview, job.CreatedByUserID,
		job.ConfirmedByUserID, job.ConfirmedAt, job.CompletedAt, job.ResultMessage, now, now); err != nil {
		return application.DataTransferJob{}, fmt.Errorf("create transfer job: %w", err)
	}
	return repository.GetJob(ctx, job.ID)
}

func (repository *DataTransferRepository) UpdateJob(
	ctx context.Context,
	job application.DataTransferJob,
) (application.DataTransferJob, error) {
	areas, err := encodeTransferAreas(job.Areas)
	if err != nil {
		return application.DataTransferJob{}, err
	}
	options, err := encodeTransferOptions(job.Options)
	if err != nil {
		return application.DataTransferJob{}, err
	}
	preview, err := encodeTransferOptions(job.Preview)
	if err != nil {
		return application.DataTransferJob{}, fmt.Errorf("encode transfer preview: %w", err)
	}
	result, err := repository.db.ExecContext(ctx, `
UPDATE data_transfer_jobs
SET profile_id=NULLIF(?, ''), profile_name=?, direction=?, format=?, areas_json=?, options_json=?, state=?,
    stage=?, source_name=?, source_sha256=?, package_version=?, total_records=?, ready_records=?,
    warning_records=?, error_records=?, preview_json=?, confirmed_by_user_id=NULLIF(?, ''),
    confirmed_at=NULLIF(?, ''), completed_at=NULLIF(?, ''), result_message=?, updated_at=?
WHERE id=?`, job.ProfileID, job.ProfileName, job.Direction, job.Format, areas, options, job.State, job.Stage,
		job.SourceName, job.SourceSHA256, job.PackageVersion, job.TotalRecords, job.ReadyRecords,
		job.WarningRecords, job.ErrorRecords, preview, job.ConfirmedByUserID, job.ConfirmedAt, job.CompletedAt,
		job.ResultMessage, timestamp(), job.ID)
	if err != nil {
		return application.DataTransferJob{}, fmt.Errorf("update transfer job: %w", err)
	}
	if err := requireDataTransferUpdate(result, "update transfer job"); err != nil {
		return application.DataTransferJob{}, err
	}
	return repository.GetJob(ctx, job.ID)
}

func (repository *DataTransferRepository) GetJob(
	ctx context.Context,
	id string,
) (application.DataTransferJob, error) {
	job, err := scanDataTransferJob(repository.db.QueryRowContext(ctx, dataTransferJobSelect+` WHERE id=?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return application.DataTransferJob{}, fmt.Errorf("get transfer job: %w", sql.ErrNoRows)
	}
	if err != nil {
		return application.DataTransferJob{}, fmt.Errorf("get transfer job: %w", err)
	}
	return job, nil
}

func (repository *DataTransferRepository) ListJobs(
	ctx context.Context,
	filter application.DataTransferJobFilter,
) ([]application.DataTransferJob, error) {
	conditions := []string{"1=1"}
	arguments := []any{}
	if filter.ProfileID != "" {
		conditions = append(conditions, "profile_id=?")
		arguments = append(arguments, filter.ProfileID)
	}
	if filter.Direction != "" {
		conditions = append(conditions, "direction=?")
		arguments = append(arguments, filter.Direction)
	}
	if len(filter.States) > 0 {
		placeholders := make([]string, len(filter.States))
		for index, state := range filter.States {
			placeholders[index] = "?"
			arguments = append(arguments, state)
		}
		conditions = append(conditions, "state IN ("+strings.Join(placeholders, ", ")+")")
	}
	query := dataTransferJobSelect + " WHERE " + strings.Join(conditions, " AND ") +
		" ORDER BY created_at DESC, id DESC"
	if filter.Limit > 0 {
		query += " LIMIT ?"
		arguments = append(arguments, filter.Limit)
	}
	rows, err := repository.db.QueryContext(ctx, query, arguments...)
	if err != nil {
		return nil, fmt.Errorf("list transfer jobs: %w", err)
	}
	defer func() { _ = rows.Close() }()
	jobs := []application.DataTransferJob{}
	for rows.Next() {
		job, err := scanDataTransferJob(rows)
		if err != nil {
			return nil, fmt.Errorf("scan transfer job: %w", err)
		}
		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate transfer jobs: %w", err)
	}
	return jobs, nil
}

func (repository *DataTransferRepository) ReplaceIssues(
	ctx context.Context,
	jobID string,
	issues []application.DataTransferIssue,
) error {
	now := timestamp()
	err := repository.withTx(ctx, func(tx *sql.Tx) error {
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
	})
	if err != nil {
		return fmt.Errorf("replace transfer issues: %w", err)
	}
	return nil
}

func (repository *DataTransferRepository) ListIssues(
	ctx context.Context,
	jobID string,
) ([]application.DataTransferIssue, error) {
	rows, err := repository.db.QueryContext(ctx, dataTransferIssueSelect+`
WHERE job_id=?
ORDER BY created_at, id`, jobID)
	if err != nil {
		return nil, fmt.Errorf("list transfer issues: %w", err)
	}
	defer func() { _ = rows.Close() }()
	issues := []application.DataTransferIssue{}
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

func (repository *DataTransferRepository) CreateArtifact(
	ctx context.Context,
	artifact application.DataTransferArtifact,
) (application.DataTransferArtifact, error) {
	artifact.ID = randomID()
	artifact.CreatedAt = timestamp()
	if _, err := repository.db.ExecContext(ctx, `
INSERT INTO data_transfer_artifacts(
  id, job_id, relative_path, display_name, mime_type, size_bytes, sha256, deleted_at, created_at
) VALUES(?, ?, ?, ?, ?, ?, ?, NULLIF(?, ''), ?)`, artifact.ID, artifact.JobID, artifact.RelativePath,
		artifact.DisplayName, artifact.MIMEType, artifact.SizeBytes, artifact.SHA256, artifact.DeletedAt,
		artifact.CreatedAt); err != nil {
		return application.DataTransferArtifact{}, fmt.Errorf("create transfer artifact: %w", err)
	}
	return artifact, nil
}

func (repository *DataTransferRepository) MarkArtifactDeleted(
	ctx context.Context,
	id string,
	deletedAt string,
) error {
	if deletedAt == "" {
		deletedAt = timestamp()
	}
	result, err := repository.db.ExecContext(ctx,
		`UPDATE data_transfer_artifacts SET deleted_at=? WHERE id=?`, deletedAt, id)
	if err != nil {
		return fmt.Errorf("mark transfer artifact deleted: %w", err)
	}
	return requireDataTransferUpdate(result, "mark transfer artifact deleted")
}

const dataTransferProfileSelect = `
SELECT id, name, direction, format, areas_json, options_json, enabled, COALESCE(created_by_user_id, ''),
       COALESCE(last_used_at, ''), created_at, updated_at
FROM data_transfer_profiles`

const dataTransferJobSelect = `
SELECT id, COALESCE(profile_id, ''), profile_name, direction, format, areas_json, options_json, state, stage,
       source_name, source_sha256, package_version, total_records, ready_records, warning_records, error_records,
       preview_json, COALESCE(created_by_user_id, ''), COALESCE(confirmed_by_user_id, ''),
       COALESCE(confirmed_at, ''), COALESCE(completed_at, ''), result_message, created_at, updated_at
FROM data_transfer_jobs`

const dataTransferIssueSelect = `
SELECT id, job_id, area, record_key, row_number, field, severity, code, message, proposed_resolution,
       selected_resolution, created_at, updated_at
FROM data_transfer_job_issues`

type dataTransferRowScanner interface {
	Scan(...any) error
}

func scanDataTransferProfile(scanner dataTransferRowScanner) (application.DataTransferProfile, error) {
	profile := application.DataTransferProfile{}
	var areas, options string
	var enabled int
	if err := scanner.Scan(&profile.ID, &profile.Name, &profile.Direction, &profile.Format, &areas, &options,
		&enabled, &profile.CreatedByUserID, &profile.LastUsedAt, &profile.CreatedAt, &profile.UpdatedAt); err != nil {
		return application.DataTransferProfile{}, err
	}
	decodedAreas, err := decodeTransferAreas(areas)
	if err != nil {
		return application.DataTransferProfile{}, err
	}
	decodedOptions, err := decodeTransferOptions(options)
	if err != nil {
		return application.DataTransferProfile{}, err
	}
	profile.Areas = decodedAreas
	profile.Options = decodedOptions
	profile.Enabled = enabled != 0
	return profile, nil
}

func scanDataTransferJob(scanner dataTransferRowScanner) (application.DataTransferJob, error) {
	job := application.DataTransferJob{}
	var areas, options, preview string
	if err := scanner.Scan(&job.ID, &job.ProfileID, &job.ProfileName, &job.Direction, &job.Format, &areas, &options,
		&job.State, &job.Stage, &job.SourceName, &job.SourceSHA256, &job.PackageVersion, &job.TotalRecords,
		&job.ReadyRecords, &job.WarningRecords, &job.ErrorRecords, &preview, &job.CreatedByUserID,
		&job.ConfirmedByUserID, &job.ConfirmedAt, &job.CompletedAt, &job.ResultMessage, &job.CreatedAt,
		&job.UpdatedAt); err != nil {
		return application.DataTransferJob{}, err
	}
	decodedAreas, err := decodeTransferAreas(areas)
	if err != nil {
		return application.DataTransferJob{}, err
	}
	decodedOptions, err := decodeTransferOptions(options)
	if err != nil {
		return application.DataTransferJob{}, err
	}
	decodedPreview, err := decodeTransferOptions(preview)
	if err != nil {
		return application.DataTransferJob{}, fmt.Errorf("decode transfer preview: %w", err)
	}
	job.Areas = decodedAreas
	job.Options = decodedOptions
	job.Preview = decodedPreview
	return job, nil
}

func scanDataTransferIssue(scanner dataTransferRowScanner) (application.DataTransferIssue, error) {
	issue := application.DataTransferIssue{}
	var rowNumber sql.NullInt64
	if err := scanner.Scan(&issue.ID, &issue.JobID, &issue.Area, &issue.RecordKey, &rowNumber, &issue.Field,
		&issue.Severity, &issue.Code, &issue.Message, &issue.ProposedResolution, &issue.SelectedResolution,
		&issue.CreatedAt, &issue.UpdatedAt); err != nil {
		return application.DataTransferIssue{}, err
	}
	if rowNumber.Valid {
		value := int(rowNumber.Int64)
		issue.RowNumber = &value
	}
	return issue, nil
}

func encodeTransferAreas(areas []application.TransferArea) (string, error) {
	if areas == nil {
		areas = []application.TransferArea{}
	}
	payload, err := json.Marshal(areas)
	if err != nil {
		return "", fmt.Errorf("encode transfer areas: %w", err)
	}
	return string(payload), nil
}

func decodeTransferAreas(payload string) ([]application.TransferArea, error) {
	areas := []application.TransferArea{}
	if err := json.Unmarshal([]byte(payload), &areas); err != nil {
		return nil, fmt.Errorf("decode transfer areas: %w", err)
	}
	if areas == nil {
		return []application.TransferArea{}, nil
	}
	return areas, nil
}

func encodeTransferOptions(options map[string]any) (string, error) {
	if options == nil {
		options = map[string]any{}
	}
	payload, err := json.Marshal(options)
	if err != nil {
		return "", fmt.Errorf("encode transfer options: %w", err)
	}
	return string(payload), nil
}

func decodeTransferOptions(payload string) (map[string]any, error) {
	options := map[string]any{}
	if err := json.Unmarshal([]byte(payload), &options); err != nil {
		return nil, fmt.Errorf("decode transfer options: %w", err)
	}
	if options == nil {
		return map[string]any{}, nil
	}
	return options, nil
}

func requireDataTransferUpdate(result sql.Result, operation string) error {
	updated, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s rows affected: %w", operation, err)
	}
	if updated == 0 {
		return fmt.Errorf("%s: %w", operation, sql.ErrNoRows)
	}
	return nil
}

func (repository *DataTransferRepository) withTx(ctx context.Context, work func(*sql.Tx) error) error {
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin data transfer transaction: %w", err)
	}
	if err := work(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit data transfer transaction: %w", err)
	}
	return nil
}

var _ application.DataTransferRepository = (*DataTransferRepository)(nil)
