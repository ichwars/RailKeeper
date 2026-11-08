package workflow

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"railkeeper/backend/internal/transfer"
	transferpolicy "railkeeper/backend/internal/transfer/policy"
)

var (
	ErrDataTransferValidation = transfer.ErrDataTransferValidation
	ErrDataTransferNotFound   = transfer.ErrDataTransferNotFound
)

type CreateDataTransferProfileInput struct {
	Name      string            `json:"name"`
	Direction TransferDirection `json:"direction"`
	Format    TransferFormat    `json:"format"`
	Areas     []TransferArea    `json:"areas"`
	Options   map[string]any    `json:"options"`
}

type UpdateDataTransferProfileInput struct {
	Name      string            `json:"name"`
	Direction TransferDirection `json:"direction"`
	Format    TransferFormat    `json:"format"`
	Areas     []TransferArea    `json:"areas"`
	Options   map[string]any    `json:"options"`
}

type DataTransferSummary struct {
	OpenJobs            int    `json:"openJobs"`
	SelectedRecords     int    `json:"selectedRecords"`
	LastExportAt        string `json:"lastExportAt"`
	ArtifactCount       int    `json:"artifactCount"`
	ArtifactBytes       int64  `json:"artifactBytes"`
	OpenFolderAvailable bool   `json:"openFolderAvailable"`
	ArtifactDirectory   string `json:"artifactDirectory"`
}

type DataTransferJobDetails struct {
	Job       DataTransferJob        `json:"job"`
	Issues    []DataTransferIssue    `json:"issues"`
	Artifacts []DataTransferArtifact `json:"artifacts"`
}

type Service struct {
	repository          DataTransferRepository
	files               FileStore
	rules               transferpolicy.AggregateRules
	openFolderAvailable bool
	openFolder          DataTransferFolderOpener
}

func New(repository DataTransferRepository, files FileStore,
	rules transferpolicy.AggregateRules, options ...Option) *Service {
	service := &Service{repository: repository, files: files, rules: rules}
	for _, option := range options {
		if option != nil {
			option(service)
		}
	}
	return service
}

func (s *Service) CreateProfile(
	ctx context.Context,
	input CreateDataTransferProfileInput,
	actorUserID string,
) (DataTransferProfile, error) {
	input.Name = strings.TrimSpace(input.Name)
	actorUserID = strings.TrimSpace(actorUserID)
	if input.Name == "" || actorUserID == "" || !transferpolicy.ValidTransferDirection(input.Direction) ||
		!transferpolicy.ValidTransferFormat(input.Format) {
		return DataTransferProfile{}, ErrDataTransferValidation
	}
	if err := transferpolicy.ValidateTransferSelection(input.Format, input.Areas); err != nil {
		return DataTransferProfile{}, err
	}
	if err := s.ensureUniqueActiveProfileName(ctx, "", input.Name, input.Direction); err != nil {
		return DataTransferProfile{}, err
	}
	return s.repository.CreateProfile(ctx, DataTransferProfile{
		Name: input.Name, Direction: input.Direction, Format: input.Format, Areas: input.Areas, Options: input.Options,
		Enabled: true, CreatedByUserID: actorUserID,
	})
}

func (s *Service) ListProfiles(ctx context.Context) ([]DataTransferProfile, error) {
	return s.repository.ListProfiles(ctx)
}

func (s *Service) Summary(
	ctx context.Context,
	allowedAreas ...TransferArea,
) (DataTransferSummary, error) {
	jobs, err := s.ListJobs(ctx, DataTransferJobFilter{}, allowedAreas...)
	if err != nil {
		return DataTransferSummary{}, err
	}
	artifacts, err := s.repository.ListArtifacts(ctx)
	if err != nil {
		return DataTransferSummary{}, err
	}
	summary := DataTransferSummary{
		OpenFolderAvailable: s.OpenFolderAvailable(),
		ArtifactDirectory:   "RailKeeper/Exporte",
	}
	visibleJobs := make(map[string]DataTransferJob, len(jobs))
	for _, job := range jobs {
		visibleJobs[job.ID] = job
		if transferpolicy.DataTransferJobOpen(job.State) {
			summary.OpenJobs++
			summary.SelectedRecords += job.TotalRecords
		}
		if job.Direction == TransferExport && job.CompletedAt > summary.LastExportAt {
			summary.LastExportAt = job.CompletedAt
		}
	}
	for _, artifact := range artifacts {
		if _, visible := visibleJobs[artifact.JobID]; !visible || artifact.DeletedAt != "" {
			continue
		}
		summary.ArtifactCount++
		summary.ArtifactBytes += artifact.SizeBytes
	}
	return summary, nil
}

func (s *Service) ListJobs(
	ctx context.Context,
	filter DataTransferJobFilter,
	allowedAreas ...TransferArea,
) ([]DataTransferJob, error) {
	if err := transferpolicy.ValidateDataTransferJobFilter(filter); err != nil {
		return nil, err
	}
	repositoryFilter := filter
	if len(allowedAreas) > 0 {
		repositoryFilter.Limit = 0
	}
	jobs, err := s.repository.ListJobs(ctx, repositoryFilter)
	if err != nil {
		return nil, err
	}
	filtered := make([]DataTransferJob, 0, len(jobs))
	for _, job := range jobs {
		if transferpolicy.DataTransferAreasAllowed(job.Areas, allowedAreas) {
			filtered = append(filtered, job)
			if filter.Limit > 0 && len(filtered) == filter.Limit {
				break
			}
		}
	}
	return filtered, nil
}

func (s *Service) GetJobDetails(
	ctx context.Context,
	id string,
	allowedAreas ...TransferArea,
) (DataTransferJobDetails, error) {
	job, err := s.getJob(ctx, id, allowedAreas)
	if err != nil {
		return DataTransferJobDetails{}, err
	}
	issues, err := s.repository.ListIssues(ctx, job.ID)
	if err != nil {
		return DataTransferJobDetails{}, err
	}
	artifacts, err := s.repository.ListArtifacts(ctx)
	if err != nil {
		return DataTransferJobDetails{}, err
	}
	jobArtifacts := make([]DataTransferArtifact, 0, 1)
	for _, artifact := range artifacts {
		if artifact.JobID == job.ID {
			jobArtifacts = append(jobArtifacts, artifact)
		}
	}
	return DataTransferJobDetails{Job: job, Issues: issues, Artifacts: jobArtifacts}, nil
}

func (s *Service) RetryJob(
	ctx context.Context,
	id string,
	actorUserID string,
	allowedAreas ...TransferArea,
) (DataTransferJob, error) {
	actorUserID = strings.TrimSpace(actorUserID)
	if actorUserID == "" {
		return DataTransferJob{}, ErrDataTransferValidation
	}
	job, err := s.getJob(ctx, id, allowedAreas)
	if err != nil {
		return DataTransferJob{}, err
	}
	if !transferpolicy.DataTransferJobRetryable(job.State) {
		return DataTransferJob{}, fmt.Errorf("%w: job is %s", ErrDataTransferConflict, job.State)
	}
	return s.repository.CreateJob(ctx, DataTransferJob{
		ProfileName:     job.ProfileName,
		Direction:       job.Direction,
		Format:          job.Format,
		Areas:           slices.Clone(job.Areas),
		Options:         transferpolicy.CloneTransferOptions(job.Options),
		State:           TransferJobDraft,
		Stage:           "created",
		PackageVersion:  DataTransferPackageVersion,
		Preview:         map[string]any{},
		CreatedByUserID: actorUserID,
	})
}

func (s *Service) DeleteJob(
	ctx context.Context,
	id string,
	allowedAreas ...TransferArea,
) error {
	job, err := s.getJob(ctx, id, allowedAreas)
	if err != nil {
		return err
	}
	if job.State != TransferJobCancelled {
		return fmt.Errorf("%w: job is %s", ErrDataTransferConflict, job.State)
	}
	return s.repository.DeleteJob(ctx, job.ID)
}

func (s *Service) getJob(
	ctx context.Context,
	id string,
	allowedAreas []TransferArea,
) (DataTransferJob, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return DataTransferJob{}, ErrDataTransferValidation
	}
	job, err := s.repository.GetJob(ctx, id)
	if errors.Is(err, ErrDataTransferNotFound) {
		return DataTransferJob{}, ErrDataTransferNotFound
	}
	if err != nil {
		return DataTransferJob{}, err
	}
	if !transferpolicy.DataTransferAreasAllowed(job.Areas, allowedAreas) {
		return DataTransferJob{}, ErrDataTransferForbidden
	}
	return job, nil
}

func (s *Service) UpdateProfile(
	ctx context.Context,
	id string,
	input UpdateDataTransferProfileInput,
) (DataTransferProfile, error) {
	id = strings.TrimSpace(id)
	input.Name = strings.TrimSpace(input.Name)
	if id == "" || input.Name == "" || !transferpolicy.ValidTransferDirection(input.Direction) || !transferpolicy.ValidTransferFormat(input.Format) {
		return DataTransferProfile{}, ErrDataTransferValidation
	}
	if err := transferpolicy.ValidateTransferSelection(input.Format, input.Areas); err != nil {
		return DataTransferProfile{}, err
	}
	profile, err := s.profile(ctx, id)
	if err != nil {
		return DataTransferProfile{}, err
	}
	if profile.Enabled {
		if err := s.ensureUniqueActiveProfileName(ctx, profile.ID, input.Name, input.Direction); err != nil {
			return DataTransferProfile{}, err
		}
	}
	profile.Name = input.Name
	profile.Direction = input.Direction
	profile.Format = input.Format
	profile.Areas = input.Areas
	profile.Options = input.Options
	return s.repository.UpdateProfile(ctx, profile)
}

func (s *Service) ensureUniqueActiveProfileName(
	ctx context.Context,
	excludeID string,
	name string,
	direction TransferDirection,
) error {
	profiles, err := s.repository.ListProfiles(ctx)
	if err != nil {
		return fmt.Errorf("list data transfer profiles: %w", err)
	}
	for _, profile := range profiles {
		if profile.ID == excludeID || !profile.Enabled || profile.Direction != direction {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(profile.Name), strings.TrimSpace(name)) {
			return fmt.Errorf("%w: active %s profile named %q already exists",
				ErrDataTransferConflict, direction, strings.TrimSpace(name))
		}
	}
	return nil
}

func (s *Service) DisableProfile(ctx context.Context, id string) (DataTransferProfile, error) {
	profile, err := s.profile(ctx, strings.TrimSpace(id))
	if err != nil {
		return DataTransferProfile{}, err
	}
	profile.Enabled = false
	return s.repository.UpdateProfile(ctx, profile)
}

func (s *Service) profile(ctx context.Context, id string) (DataTransferProfile, error) {
	if id == "" {
		return DataTransferProfile{}, ErrDataTransferValidation
	}
	profile, err := s.repository.GetProfile(ctx, id)
	if errors.Is(err, ErrDataTransferNotFound) {
		return DataTransferProfile{}, ErrDataTransferNotFound
	}
	return profile, err
}
