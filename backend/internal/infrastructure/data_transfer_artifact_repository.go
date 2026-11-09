package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"railkeeper/backend/internal/transfer"
)

func (repository *DataTransferRepository) CreateArtifact(
	ctx context.Context,
	artifact transfer.DataTransferArtifact,
) (transfer.DataTransferArtifact, error) {
	artifact.ID = randomID()
	artifact.CreatedAt = timestamp()
	result, err := repository.db.ExecContext(ctx, `
INSERT INTO data_transfer_artifacts(
  id, job_id, relative_path, display_name, mime_type, size_bytes, sha256, deleted_at, created_at
) SELECT ?, ?, ?, ?, ?, ?, ?, NULLIF(?, ''), ?
WHERE NOT EXISTS(
  SELECT 1 FROM data_transfer_artifacts WHERE job_id=? AND deleted_at IS NULL
)`, artifact.ID, artifact.JobID, artifact.RelativePath,
		artifact.DisplayName, artifact.MIMEType, artifact.SizeBytes, artifact.SHA256, artifact.DeletedAt,
		artifact.CreatedAt, artifact.JobID)
	if err != nil {
		return transfer.DataTransferArtifact{}, fmt.Errorf("create transfer artifact: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return transfer.DataTransferArtifact{}, fmt.Errorf("read transfer artifact creation result: %w", err)
	}
	if affected != 1 {
		return transfer.DataTransferArtifact{}, transfer.ErrDataTransferConflict
	}
	return artifact, nil
}

func (repository *DataTransferRepository) ListArtifacts(
	ctx context.Context,
) ([]transfer.DataTransferArtifact, error) {
	rows, err := repository.db.QueryContext(ctx, `
SELECT id, job_id, relative_path, display_name, mime_type, size_bytes, sha256,
       COALESCE(deleted_at, ''), created_at
FROM data_transfer_artifacts
ORDER BY created_at DESC, id DESC`)
	if err != nil {
		return nil, fmt.Errorf("list transfer artifacts: %w", err)
	}
	defer func() { _ = rows.Close() }()
	artifacts := []transfer.DataTransferArtifact{}
	for rows.Next() {
		artifact := transfer.DataTransferArtifact{}
		if err := rows.Scan(
			&artifact.ID, &artifact.JobID, &artifact.RelativePath, &artifact.DisplayName, &artifact.MIMEType,
			&artifact.SizeBytes, &artifact.SHA256, &artifact.DeletedAt, &artifact.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan transfer artifact: %w", err)
		}
		artifacts = append(artifacts, artifact)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate transfer artifacts: %w", err)
	}
	return artifacts, nil
}

func (repository *DataTransferRepository) ClaimExportJob(
	ctx context.Context,
	id string,
) (transfer.DataTransferJob, error) {
	result, err := repository.db.ExecContext(ctx, `
UPDATE data_transfer_jobs
SET state=?, stage='snapshot', revision=revision+1, updated_at=?
WHERE id=? AND direction=? AND state=?`, transfer.TransferJobRunning, timestamp(), id,
		transfer.TransferExport, transfer.TransferJobDraft)
	if err != nil {
		return transfer.DataTransferJob{}, fmt.Errorf("claim transfer export job: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return transfer.DataTransferJob{}, fmt.Errorf("read transfer export claim result: %w", err)
	}
	if affected == 0 {
		var exists int
		if err := repository.db.QueryRowContext(ctx,
			`SELECT EXISTS(SELECT 1 FROM data_transfer_jobs WHERE id=?)`, id).Scan(&exists); err != nil {
			return transfer.DataTransferJob{}, fmt.Errorf("check transfer export job: %w", err)
		}
		if exists == 0 {
			return transfer.DataTransferJob{}, fmt.Errorf("claim transfer export job: %w: %w", transfer.ErrDataTransferNotFound, sql.ErrNoRows)
		}
		return transfer.DataTransferJob{}, transfer.ErrDataTransferConflict
	}
	return repository.GetJob(ctx, id)
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
