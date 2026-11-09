package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"railkeeper/backend/internal/transfer"
	"strings"
)

type DataTransferRepository struct {
	db *sql.DB
}

func NewDataTransferRepository(db *sql.DB) *DataTransferRepository {
	return &DataTransferRepository{db: db}
}

func (repository *DataTransferRepository) CreateProfile(
	ctx context.Context,
	profile transfer.DataTransferProfile,
) (transfer.DataTransferProfile, error) {
	areas, err := encodeTransferAreas(profile.Areas)
	if err != nil {
		return transfer.DataTransferProfile{}, err
	}
	options, err := encodeTransferOptions(profile.Options)
	if err != nil {
		return transfer.DataTransferProfile{}, err
	}
	now := timestamp()
	nameKey := transfer.DataTransferProfileNameKey(profile.Name)
	profile.ID = randomID()
	profile.CreatedAt = now
	profile.UpdatedAt = now
	enabled := boolToInt(profile.Enabled)
	result, err := repository.db.ExecContext(ctx, `
INSERT INTO data_transfer_profiles(
  id, name, name_key, direction, format, areas_json, options_json, enabled, created_by_user_id, last_used_at,
  created_at, updated_at
) VALUES(?, ?, ?, ?, ?, ?, ?, ?, NULLIF(?, ''), NULLIF(?, ''), ?, ?)`,
		profile.ID, profile.Name, nameKey, profile.Direction, profile.Format, areas, options, enabled,
		profile.CreatedByUserID, profile.LastUsedAt, now, now)
	if err != nil {
		if isDataTransferProfileNameConflict(err) {
			return transfer.DataTransferProfile{}, transfer.ErrDataTransferConflict
		}
		return transfer.DataTransferProfile{}, fmt.Errorf("create transfer profile: %w", err)
	}
	created, err := result.RowsAffected()
	if err != nil {
		return transfer.DataTransferProfile{}, fmt.Errorf("read create transfer profile result: %w", err)
	}
	if created == 0 {
		return transfer.DataTransferProfile{}, transfer.ErrDataTransferConflict
	}
	return repository.GetProfile(ctx, profile.ID)
}

func (repository *DataTransferRepository) UpdateProfile(
	ctx context.Context,
	profile transfer.DataTransferProfile,
) (transfer.DataTransferProfile, error) {
	areas, err := encodeTransferAreas(profile.Areas)
	if err != nil {
		return transfer.DataTransferProfile{}, err
	}
	options, err := encodeTransferOptions(profile.Options)
	if err != nil {
		return transfer.DataTransferProfile{}, err
	}
	enabled := boolToInt(profile.Enabled)
	nameKey := transfer.DataTransferProfileNameKey(profile.Name)
	result, err := repository.db.ExecContext(ctx, `
UPDATE data_transfer_profiles
SET name=?, name_key=?, direction=?, format=?, areas_json=?, options_json=?, enabled=?, last_used_at=NULLIF(?, ''),
    updated_at=?
WHERE id=?`, profile.Name, nameKey, profile.Direction, profile.Format, areas, options, enabled,
		profile.LastUsedAt, timestamp(), profile.ID)
	if err != nil {
		if isDataTransferProfileNameConflict(err) {
			return transfer.DataTransferProfile{}, transfer.ErrDataTransferConflict
		}
		return transfer.DataTransferProfile{}, fmt.Errorf("update transfer profile: %w", err)
	}
	updated, err := result.RowsAffected()
	if err != nil {
		return transfer.DataTransferProfile{}, fmt.Errorf("read update transfer profile result: %w", err)
	}
	if updated == 0 {
		var exists int
		if err := repository.db.QueryRowContext(ctx,
			`SELECT EXISTS(SELECT 1 FROM data_transfer_profiles WHERE id=?)`, profile.ID).Scan(&exists); err != nil {
			return transfer.DataTransferProfile{}, fmt.Errorf("check transfer profile: %w", err)
		}
		if exists == 0 {
			return transfer.DataTransferProfile{}, fmt.Errorf("update transfer profile: %w: %w", transfer.ErrDataTransferNotFound, sql.ErrNoRows)
		}
		return transfer.DataTransferProfile{}, transfer.ErrDataTransferConflict
	}
	return repository.GetProfile(ctx, profile.ID)
}

func isDataTransferProfileNameConflict(err error) bool {
	return err != nil && strings.Contains(err.Error(),
		"UNIQUE constraint failed: data_transfer_profiles.direction, data_transfer_profiles.name_key")
}

func (repository *DataTransferRepository) ListProfiles(
	ctx context.Context,
) ([]transfer.DataTransferProfile, error) {
	rows, err := repository.db.QueryContext(ctx, dataTransferProfileSelect+`
ORDER BY name COLLATE NOCASE, id`)
	if err != nil {
		return nil, fmt.Errorf("list transfer profiles: %w", err)
	}
	defer func() { _ = rows.Close() }()
	profiles := []transfer.DataTransferProfile{}
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
) (transfer.DataTransferProfile, error) {
	profile, err := scanDataTransferProfile(repository.db.QueryRowContext(ctx, dataTransferProfileSelect+` WHERE id=?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return transfer.DataTransferProfile{}, fmt.Errorf("get transfer profile: %w: %w", transfer.ErrDataTransferNotFound, sql.ErrNoRows)
	}
	if err != nil {
		return transfer.DataTransferProfile{}, fmt.Errorf("get transfer profile: %w", err)
	}
	return profile, nil
}

func (repository *DataTransferRepository) CreateJob(
	ctx context.Context,
	job transfer.DataTransferJob,
) (transfer.DataTransferJob, error) {
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
	now := timestamp()
	job.ID = randomID()
	job.CreatedAt = now
	job.UpdatedAt = now
	if _, err := repository.db.ExecContext(ctx, `
INSERT INTO data_transfer_jobs(
  id, profile_id, profile_name, direction, format, areas_json, options_json, state, stage, source_name,
  source_sha256, package_version, revision, total_records, ready_records, warning_records, error_records, preview_json,
  created_by_user_id, confirmed_by_user_id, confirmed_at, completed_at, result_message, created_at, updated_at
) VALUES(?, NULLIF(?, ''), ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NULLIF(?, ''), NULLIF(?, ''),
  NULLIF(?, ''), NULLIF(?, ''), ?, ?, ?)`, job.ID, job.ProfileID, job.ProfileName, job.Direction, job.Format,
		areas, options, job.State, job.Stage, job.SourceName, job.SourceSHA256, job.PackageVersion,
		1, job.TotalRecords, job.ReadyRecords, job.WarningRecords, job.ErrorRecords, preview, job.CreatedByUserID,
		job.ConfirmedByUserID, job.ConfirmedAt, job.CompletedAt, job.ResultMessage, now, now); err != nil {
		return transfer.DataTransferJob{}, fmt.Errorf("create transfer job: %w", err)
	}
	return repository.GetJob(ctx, job.ID)
}

func (repository *DataTransferRepository) UpdateJob(
	ctx context.Context,
	job transfer.DataTransferJob,
) (transfer.DataTransferJob, error) {
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
	result, err := repository.db.ExecContext(ctx, `
UPDATE data_transfer_jobs
SET profile_id=NULLIF(?, ''), profile_name=?, direction=?, format=?, areas_json=?, options_json=?, state=?,
    stage=?, source_name=?, source_sha256=?, package_version=?, total_records=?, ready_records=?,
    warning_records=?, error_records=?, preview_json=?, confirmed_by_user_id=NULLIF(?, ''),
    confirmed_at=NULLIF(?, ''), completed_at=NULLIF(?, ''), result_message=?, revision=revision+1, updated_at=?
WHERE id=?`, job.ProfileID, job.ProfileName, job.Direction, job.Format, areas, options, job.State, job.Stage,
		job.SourceName, job.SourceSHA256, job.PackageVersion, job.TotalRecords, job.ReadyRecords,
		job.WarningRecords, job.ErrorRecords, preview, job.ConfirmedByUserID, job.ConfirmedAt, job.CompletedAt,
		job.ResultMessage, timestamp(), job.ID)
	if err != nil {
		return transfer.DataTransferJob{}, fmt.Errorf("update transfer job: %w", err)
	}
	if err := requireDataTransferUpdate(result, "update transfer job"); err != nil {
		return transfer.DataTransferJob{}, err
	}
	return repository.GetJob(ctx, job.ID)
}

func (repository *DataTransferRepository) GetJob(
	ctx context.Context,
	id string,
) (transfer.DataTransferJob, error) {
	job, err := scanDataTransferJob(repository.db.QueryRowContext(ctx, dataTransferJobSelect+` WHERE id=?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return transfer.DataTransferJob{}, fmt.Errorf("get transfer job: %w: %w", transfer.ErrDataTransferNotFound, sql.ErrNoRows)
	}
	if err != nil {
		return transfer.DataTransferJob{}, fmt.Errorf("get transfer job: %w", err)
	}
	return job, nil
}

func (repository *DataTransferRepository) DeleteJob(ctx context.Context, id string) error {
	result, err := repository.db.ExecContext(ctx, `DELETE FROM data_transfer_jobs WHERE id=?`, id)
	if err != nil {
		return fmt.Errorf("delete transfer job: %w", err)
	}
	return requireDataTransferUpdate(result, "delete transfer job")
}

func (repository *DataTransferRepository) ListJobs(
	ctx context.Context,
	filter transfer.DataTransferJobFilter,
) ([]transfer.DataTransferJob, error) {
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
	jobs := []transfer.DataTransferJob{}
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

var _ transfer.DataTransferRepository = (*DataTransferRepository)(nil)
