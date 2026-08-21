package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strings"
)

var (
	ErrDataTransferValidation = errors.New("data transfer validation failed")
	ErrDataTransferNotFound   = errors.New("data transfer profile not found")
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

type DataTransferService struct {
	repository          DataTransferRepository
	dataDir             string
	openFolderAvailable bool
	openFolder          DataTransferFolderOpener
}

func NewDataTransferService(
	repository DataTransferRepository,
	dataDir string,
	options ...DataTransferServiceOption,
) *DataTransferService {
	service := &DataTransferService{repository: repository, dataDir: strings.TrimSpace(dataDir)}
	for _, option := range options {
		if option != nil {
			option(service)
		}
	}
	return service
}

func (s *DataTransferService) CreateProfile(
	ctx context.Context,
	input CreateDataTransferProfileInput,
	actorUserID string,
) (DataTransferProfile, error) {
	input.Name = strings.TrimSpace(input.Name)
	actorUserID = strings.TrimSpace(actorUserID)
	if input.Name == "" || actorUserID == "" || !validTransferDirection(input.Direction) ||
		!validTransferFormat(input.Format) {
		return DataTransferProfile{}, ErrDataTransferValidation
	}
	if err := validateTransferSelection(input.Format, input.Areas); err != nil {
		return DataTransferProfile{}, err
	}
	return s.repository.CreateProfile(ctx, DataTransferProfile{
		Name: input.Name, Direction: input.Direction, Format: input.Format, Areas: input.Areas, Options: input.Options,
		Enabled: true, CreatedByUserID: actorUserID,
	})
}

func (s *DataTransferService) ListProfiles(ctx context.Context) ([]DataTransferProfile, error) {
	return s.repository.ListProfiles(ctx)
}

func (s *DataTransferService) Summary(
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
		if dataTransferJobOpen(job.State) {
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

func (s *DataTransferService) ListJobs(
	ctx context.Context,
	filter DataTransferJobFilter,
	allowedAreas ...TransferArea,
) ([]DataTransferJob, error) {
	if err := validateDataTransferJobFilter(filter); err != nil {
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
		if dataTransferAreasAllowed(job.Areas, allowedAreas) {
			filtered = append(filtered, job)
			if filter.Limit > 0 && len(filtered) == filter.Limit {
				break
			}
		}
	}
	return filtered, nil
}

func (s *DataTransferService) GetJobDetails(
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

func (s *DataTransferService) RetryJob(
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
	if !dataTransferJobRetryable(job.State) {
		return DataTransferJob{}, fmt.Errorf("%w: job is %s", ErrDataTransferConflict, job.State)
	}
	return s.repository.CreateJob(ctx, DataTransferJob{
		ProfileName:     job.ProfileName,
		Direction:       job.Direction,
		Format:          job.Format,
		Areas:           slices.Clone(job.Areas),
		Options:         cloneTransferOptions(job.Options),
		State:           TransferJobDraft,
		Stage:           "created",
		PackageVersion:  DataTransferPackageVersion,
		Preview:         map[string]any{},
		CreatedByUserID: actorUserID,
	})
}

func (s *DataTransferService) getJob(
	ctx context.Context,
	id string,
	allowedAreas []TransferArea,
) (DataTransferJob, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return DataTransferJob{}, ErrDataTransferValidation
	}
	job, err := s.repository.GetJob(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return DataTransferJob{}, ErrDataTransferNotFound
	}
	if err != nil {
		return DataTransferJob{}, err
	}
	if !dataTransferAreasAllowed(job.Areas, allowedAreas) {
		return DataTransferJob{}, ErrDataTransferForbidden
	}
	return job, nil
}

func validateDataTransferJobFilter(filter DataTransferJobFilter) error {
	if filter.Limit < 0 || filter.Limit > 200 {
		return ErrDataTransferValidation
	}
	if filter.Direction != "" && !validTransferDirection(filter.Direction) {
		return ErrDataTransferValidation
	}
	for _, state := range filter.States {
		if !validDataTransferJobState(state) {
			return ErrDataTransferValidation
		}
	}
	return nil
}

func validDataTransferJobState(state TransferJobState) bool {
	return slices.Contains([]TransferJobState{
		TransferJobDraft, TransferJobReading, TransferJobReviewRequired, TransferJobReady, TransferJobRunning,
		TransferJobCompleted, TransferJobCompletedWithWarnings, TransferJobFailed, TransferJobCancelled,
	}, state)
}

func dataTransferJobOpen(state TransferJobState) bool {
	return state == TransferJobDraft || state == TransferJobReading || state == TransferJobReviewRequired ||
		state == TransferJobReady || state == TransferJobRunning
}

func dataTransferJobRetryable(state TransferJobState) bool {
	return state == TransferJobCompleted || state == TransferJobCompletedWithWarnings ||
		state == TransferJobFailed || state == TransferJobCancelled
}

func (s *DataTransferService) UpdateProfile(
	ctx context.Context,
	id string,
	input UpdateDataTransferProfileInput,
) (DataTransferProfile, error) {
	id = strings.TrimSpace(id)
	input.Name = strings.TrimSpace(input.Name)
	if id == "" || input.Name == "" || !validTransferDirection(input.Direction) || !validTransferFormat(input.Format) {
		return DataTransferProfile{}, ErrDataTransferValidation
	}
	if err := validateTransferSelection(input.Format, input.Areas); err != nil {
		return DataTransferProfile{}, err
	}
	profile, err := s.profile(ctx, id)
	if err != nil {
		return DataTransferProfile{}, err
	}
	profile.Name = input.Name
	profile.Direction = input.Direction
	profile.Format = input.Format
	profile.Areas = input.Areas
	profile.Options = input.Options
	return s.repository.UpdateProfile(ctx, profile)
}

func (s *DataTransferService) DisableProfile(ctx context.Context, id string) (DataTransferProfile, error) {
	profile, err := s.profile(ctx, strings.TrimSpace(id))
	if err != nil {
		return DataTransferProfile{}, err
	}
	profile.Enabled = false
	return s.repository.UpdateProfile(ctx, profile)
}

func (s *DataTransferService) profile(ctx context.Context, id string) (DataTransferProfile, error) {
	if id == "" {
		return DataTransferProfile{}, ErrDataTransferValidation
	}
	profile, err := s.repository.GetProfile(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return DataTransferProfile{}, ErrDataTransferNotFound
	}
	return profile, err
}

func validateTransferSelection(format TransferFormat, areas []TransferArea) error {
	allowed := map[TransferArea]bool{
		TransferVehicles: true, TransferAccessories: true, TransferExhibitionLists: true,
	}
	if len(areas) == 0 {
		return fmt.Errorf("%w: select at least one area", ErrDataTransferValidation)
	}
	seen := map[TransferArea]bool{}
	for _, area := range areas {
		if !allowed[area] || seen[area] {
			return fmt.Errorf("%w: unsupported or repeated area %q", ErrDataTransferValidation, area)
		}
		seen[area] = true
	}
	if format == TransferCSV && len(areas) != 1 {
		return fmt.Errorf("%w: csv requires exactly one area", ErrDataTransferValidation)
	}
	if format == TransferCSV && areas[0] == TransferExhibitionLists {
		return fmt.Errorf("%w: exhibition lists require railkeeper-json", ErrDataTransferValidation)
	}
	return nil
}

func validTransferDirection(direction TransferDirection) bool {
	return direction == TransferImport || direction == TransferExport
}

func validTransferFormat(format TransferFormat) bool {
	return format == TransferCSV || format == TransferJSON
}
