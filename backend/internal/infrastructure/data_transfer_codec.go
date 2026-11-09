package infrastructure

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"railkeeper/backend/internal/transfer"
)

const dataTransferProfileSelect = `
SELECT id, name, direction, format, areas_json, options_json, enabled, COALESCE(created_by_user_id, ''),
       COALESCE(last_used_at, ''), created_at, updated_at
FROM data_transfer_profiles`

const dataTransferJobSelect = `
SELECT id, COALESCE(profile_id, ''), profile_name, direction, format, areas_json, options_json, state, stage,
       source_name, source_sha256, package_version, revision, total_records, ready_records, warning_records, error_records,
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

func scanDataTransferProfile(scanner dataTransferRowScanner) (transfer.DataTransferProfile, error) {
	profile := transfer.DataTransferProfile{}
	var areas, options string
	var enabled int
	if err := scanner.Scan(&profile.ID, &profile.Name, &profile.Direction, &profile.Format, &areas, &options,
		&enabled, &profile.CreatedByUserID, &profile.LastUsedAt, &profile.CreatedAt, &profile.UpdatedAt); err != nil {
		return transfer.DataTransferProfile{}, err
	}
	decodedAreas, err := decodeTransferAreas(areas)
	if err != nil {
		return transfer.DataTransferProfile{}, err
	}
	decodedOptions, err := decodeTransferOptions(options)
	if err != nil {
		return transfer.DataTransferProfile{}, err
	}
	profile.Areas = decodedAreas
	profile.Options = decodedOptions
	profile.Enabled = enabled != 0
	return profile, nil
}

func scanDataTransferJob(scanner dataTransferRowScanner) (transfer.DataTransferJob, error) {
	job := transfer.DataTransferJob{}
	var areas, options, preview string
	if err := scanner.Scan(&job.ID, &job.ProfileID, &job.ProfileName, &job.Direction, &job.Format, &areas, &options,
		&job.State, &job.Stage, &job.SourceName, &job.SourceSHA256, &job.PackageVersion, &job.Revision, &job.TotalRecords,
		&job.ReadyRecords, &job.WarningRecords, &job.ErrorRecords, &preview, &job.CreatedByUserID,
		&job.ConfirmedByUserID, &job.ConfirmedAt, &job.CompletedAt, &job.ResultMessage, &job.CreatedAt,
		&job.UpdatedAt); err != nil {
		return transfer.DataTransferJob{}, err
	}
	decodedAreas, err := decodeTransferAreas(areas)
	if err != nil {
		return transfer.DataTransferJob{}, err
	}
	decodedOptions, err := decodeTransferOptions(options)
	if err != nil {
		return transfer.DataTransferJob{}, err
	}
	decodedPreview, err := decodeTransferOptions(preview)
	if err != nil {
		return transfer.DataTransferJob{}, fmt.Errorf("decode transfer preview: %w", err)
	}
	job.Areas = decodedAreas
	job.Options = decodedOptions
	job.Preview = decodedPreview
	return job, nil
}

func scanDataTransferIssue(scanner dataTransferRowScanner) (transfer.DataTransferIssue, error) {
	issue := transfer.DataTransferIssue{}
	var rowNumber sql.NullInt64
	if err := scanner.Scan(&issue.ID, &issue.JobID, &issue.Area, &issue.RecordKey, &rowNumber, &issue.Field,
		&issue.Severity, &issue.Code, &issue.Message, &issue.ProposedResolution, &issue.SelectedResolution,
		&issue.CreatedAt, &issue.UpdatedAt); err != nil {
		return transfer.DataTransferIssue{}, err
	}
	if rowNumber.Valid {
		value := int(rowNumber.Int64)
		issue.RowNumber = &value
	}
	return issue, nil
}

func encodeTransferAreas(areas []transfer.TransferArea) (string, error) {
	if areas == nil {
		areas = []transfer.TransferArea{}
	}
	payload, err := json.Marshal(areas)
	if err != nil {
		return "", fmt.Errorf("encode transfer areas: %w", err)
	}
	return string(payload), nil
}

func decodeTransferAreas(payload string) ([]transfer.TransferArea, error) {
	areas := []transfer.TransferArea{}
	if err := json.Unmarshal([]byte(payload), &areas); err != nil {
		return nil, fmt.Errorf("decode transfer areas: %w", err)
	}
	if areas == nil {
		return []transfer.TransferArea{}, nil
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
		return fmt.Errorf("%s: %w: %w", operation, transfer.ErrDataTransferNotFound, sql.ErrNoRows)
	}
	return nil
}
