package workflow

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"railkeeper/backend/internal/transfer"
	transferpolicy "railkeeper/backend/internal/transfer/policy"
)

var (
	ErrDataTransferArtifactPath      = transfer.ErrDataTransferArtifactPath
	ErrDataTransferForbidden         = transfer.ErrDataTransferForbidden
	ErrDataTransferArtifactDeleted   = transfer.ErrDataTransferArtifactDeleted
	ErrDataTransferConflict          = transfer.ErrDataTransferConflict
	ErrDataTransferOpenUnavailable   = transfer.ErrDataTransferOpenUnavailable
	ErrDataTransferExportUnavailable = transfer.ErrDataTransferExportUnavailable
)

type DataTransferExportResult struct {
	Job                 DataTransferJob      `json:"job"`
	Artifact            DataTransferArtifact `json:"artifact"`
	OpenFolderAvailable bool                 `json:"openFolderAvailable"`
}

type DataTransferFolderOpener func(context.Context, string) error

type Option func(*Service)

func WithDataTransferFolderOpener(available bool, opener DataTransferFolderOpener) Option {
	return func(service *Service) {
		service.openFolderAvailable = available && opener != nil
		service.openFolder = opener
	}
}

type dataTransferExportRepository interface {
	DataTransferRepository
	Snapshot(context.Context, []TransferArea) (DataTransferSnapshot, error)
	GetArtifact(context.Context, string) (DataTransferArtifact, error)
	ClaimExportJob(context.Context, string) (DataTransferJob, error)
}

func (s *Service) CreateExportJob(
	ctx context.Context,
	profileID string,
	actorUserID string,
	allowedAreas ...TransferArea,
) (DataTransferJob, error) {
	profileID = strings.TrimSpace(profileID)
	actorUserID = strings.TrimSpace(actorUserID)
	if profileID == "" || actorUserID == "" {
		return DataTransferJob{}, ErrDataTransferValidation
	}
	profile, err := s.profile(ctx, profileID)
	if err != nil {
		return DataTransferJob{}, err
	}
	if !profile.Enabled || profile.Direction != TransferExport {
		return DataTransferJob{}, fmt.Errorf("%w: profile is not an enabled export profile", ErrDataTransferValidation)
	}
	if !transferpolicy.DataTransferAreasAllowed(profile.Areas, allowedAreas) {
		return DataTransferJob{}, ErrDataTransferForbidden
	}
	return s.repository.CreateJob(ctx, DataTransferJob{
		ProfileID: profile.ID, ProfileName: profile.Name, Direction: profile.Direction, Format: profile.Format,
		Areas: slices.Clone(profile.Areas), Options: transferpolicy.CloneTransferOptions(profile.Options), State: TransferJobDraft,
		Stage: "created", PackageVersion: DataTransferPackageVersion, Preview: map[string]any{},
		CreatedByUserID: actorUserID,
	})
}

func (s *Service) ExecuteExport(
	ctx context.Context,
	jobID string,
	allowedAreas ...TransferArea,
) (DataTransferExportResult, error) {
	repository, err := s.exportRepository()
	if err != nil {
		return DataTransferExportResult{}, err
	}
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return DataTransferExportResult{}, ErrDataTransferValidation
	}
	job, err := repository.GetJob(ctx, jobID)
	if errors.Is(err, ErrDataTransferNotFound) {
		return DataTransferExportResult{}, ErrDataTransferNotFound
	}
	if err != nil {
		return DataTransferExportResult{}, err
	}
	if job.Direction != TransferExport || !transferpolicy.ValidTransferFormat(job.Format) {
		return DataTransferExportResult{}, fmt.Errorf("%w: export job cannot be executed", ErrDataTransferValidation)
	}
	if job.State != TransferJobDraft {
		return DataTransferExportResult{}, fmt.Errorf("%w: export job is %s", ErrDataTransferConflict, job.State)
	}
	if !transferpolicy.DataTransferAreasAllowed(job.Areas, allowedAreas) {
		return DataTransferExportResult{}, ErrDataTransferForbidden
	}
	job, err = repository.ClaimExportJob(ctx, job.ID)
	if err != nil {
		return DataTransferExportResult{}, err
	}

	snapshot, err := repository.Snapshot(ctx, job.Areas)
	if err != nil {
		return DataTransferExportResult{}, s.failExportJob(ctx, repository, job, err)
	}
	createdAt := time.Now().UTC().Format(time.RFC3339)
	payload, displayName, mimeType, err := transferpolicy.MarshalDataTransferExport(job, snapshot, createdAt)
	if err != nil {
		return DataTransferExportResult{}, s.failExportJob(ctx, repository, job, err)
	}
	artifact, err := s.files.WriteExport(ctx, repository, job.ID, displayName, mimeType, payload)
	if err != nil {
		return DataTransferExportResult{}, s.failExportJob(ctx, repository, job, err)
	}
	job.State = TransferJobCompleted
	job.Stage = "completed"
	job.SourceName = artifact.DisplayName
	job.SourceSHA256 = artifact.SHA256
	job.TotalRecords = transferpolicy.DataTransferRecordCount(snapshot, job.Areas)
	job.ReadyRecords = job.TotalRecords
	job.CompletedAt = time.Now().UTC().Format(time.RFC3339)
	job.ResultMessage = "Export completed."
	job, err = repository.UpdateJob(ctx, job)
	if err != nil {
		return DataTransferExportResult{}, err
	}
	return DataTransferExportResult{
		Job: job, Artifact: artifact, OpenFolderAvailable: s.OpenFolderAvailable(),
	}, nil
}

func (s *Service) OpenArtifact(ctx context.Context, id string,
	allowedAreas ...TransferArea) (DataTransferArtifact, ArtifactReader, error) {
	repository, err := s.exportRepository()
	if err != nil {
		return DataTransferArtifact{}, nil, err
	}
	artifact, err := s.authorizedArtifact(ctx, repository, id, allowedAreas)
	if err != nil {
		return DataTransferArtifact{}, nil, err
	}
	file, err := s.files.OpenArtifact(artifact.RelativePath)
	if err != nil {
		return DataTransferArtifact{}, nil, err
	}
	return artifact, file, nil
}

func (s *Service) DeleteArtifact(ctx context.Context, id string) error {
	repository, err := s.exportRepository()
	if err != nil {
		return err
	}
	artifact, err := s.authorizedArtifact(ctx, repository, id, nil)
	if err != nil {
		return err
	}
	if err := s.files.RemoveArtifact(artifact.RelativePath); err != nil {
		return err
	}
	return repository.MarkArtifactDeleted(ctx, artifact.ID, time.Now().UTC().Format(time.RFC3339))
}

func (s *Service) OpenFolderAvailable() bool {
	return s != nil && s.openFolderAvailable && s.openFolder != nil
}

func (s *Service) OpenArtifactFolder(ctx context.Context) error {
	if !s.OpenFolderAvailable() {
		return ErrDataTransferOpenUnavailable
	}
	directory, err := s.files.ExportDirectory()
	if err != nil {
		return err
	}
	return s.openFolder(ctx, directory)
}

func (s *Service) exportRepository() (dataTransferExportRepository, error) {
	repository, ok := s.repository.(dataTransferExportRepository)
	if !ok || repository == nil {
		return nil, ErrDataTransferExportUnavailable
	}
	return repository, nil
}

func (s *Service) authorizedArtifact(
	ctx context.Context,
	repository dataTransferExportRepository,
	id string,
	allowedAreas []TransferArea,
) (DataTransferArtifact, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return DataTransferArtifact{}, ErrDataTransferValidation
	}
	artifact, err := repository.GetArtifact(ctx, id)
	if errors.Is(err, ErrDataTransferNotFound) {
		return DataTransferArtifact{}, ErrDataTransferNotFound
	}
	if err != nil {
		return DataTransferArtifact{}, err
	}
	if artifact.DeletedAt != "" {
		return DataTransferArtifact{}, ErrDataTransferArtifactDeleted
	}
	job, err := repository.GetJob(ctx, artifact.JobID)
	if err != nil {
		return DataTransferArtifact{}, err
	}
	if !transferpolicy.DataTransferAreasAllowed(job.Areas, allowedAreas) {
		return DataTransferArtifact{}, ErrDataTransferForbidden
	}
	return artifact, nil
}

func (s *Service) failExportJob(
	ctx context.Context,
	repository dataTransferExportRepository,
	job DataTransferJob,
	cause error,
) error {
	job.State = TransferJobFailed
	job.Stage = "failed"
	job.ResultMessage = "Export failed."
	if _, err := repository.UpdateJob(ctx, job); err != nil {
		return errors.Join(cause, fmt.Errorf("record failed transfer export: %w", err))
	}
	return cause
}
