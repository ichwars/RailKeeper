package workflow

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"

	"railkeeper/backend/internal/transfer"
	transferpolicy "railkeeper/backend/internal/transfer/policy"
)

const DataTransferMaxUploadBytes int64 = 25 * 1024 * 1024

const DataTransferTargetFingerprintVersion = transfer.DataTransferTargetFingerprintVersion

var (
	ErrDataTransferUploadTooLarge = transfer.ErrDataTransferUploadTooLarge
	ErrDataTransferUploadPath     = transfer.ErrDataTransferUploadPath
)

type DataTransferPreview = transferpolicy.DataTransferPreview

type DataTransferPreviewRecord = transfer.DataTransferPreviewRecord

func DataTransferTargetFingerprint(value any) string {
	return transfer.DataTransferTargetFingerprint(value)
}

func DataTransferTargetFingerprintForVersion(value any, version int) string {
	return transfer.DataTransferTargetFingerprintForVersion(value, version)
}

type dataTransferImportRepository interface {
	DataTransferRepository
	Snapshot(context.Context, []TransferArea) (DataTransferSnapshot, error)
	CompareAndUpdateImportJob(context.Context, DataTransferImportMutation) (DataTransferJob, error)
	ApplyImport(context.Context, DataTransferJob, string) error
	ApplyImportWithPolicy(context.Context, DataTransferJob, string, DataTransferImportPolicy) error
	ValidateTransferAccessoryReferences(context.Context, TransferAccessory, string) error
}

func (s *Service) CreateImportJob(
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
	if !profile.Enabled || profile.Direction != TransferImport {
		return DataTransferJob{}, fmt.Errorf("%w: profile is not an enabled import profile", ErrDataTransferValidation)
	}
	if !transferpolicy.DataTransferAreasAllowed(profile.Areas, allowedAreas) {
		return DataTransferJob{}, ErrDataTransferForbidden
	}
	return s.repository.CreateJob(ctx, DataTransferJob{
		ProfileID: profile.ID, ProfileName: profile.Name, Direction: TransferImport, Format: profile.Format,
		Areas: slices.Clone(profile.Areas), Options: transferpolicy.CloneTransferOptions(profile.Options), State: TransferJobDraft,
		Stage: "created", PackageVersion: DataTransferPackageVersion, Preview: map[string]any{},
		CreatedByUserID: actorUserID,
	})
}

func (s *Service) UploadAndPreview(
	ctx context.Context,
	jobID string,
	sourceName string,
	payload []byte,
	actorUserID string,
	allowedAreas ...TransferArea,
) (DataTransferPreview, error) {
	return s.UploadAndPreviewWithMapping(ctx, jobID, sourceName, payload, actorUserID, nil, allowedAreas...)
}

func (s *Service) UploadAndPreviewWithMapping(
	ctx context.Context,
	jobID string,
	sourceName string,
	payload []byte,
	actorUserID string,
	mapping *DataTransferCSVMappingInput,
	allowedAreas ...TransferArea,
) (DataTransferPreview, error) {
	return s.UploadAndPreviewReader(
		ctx, jobID, sourceName, "", bytes.NewReader(payload), actorUserID, mapping, allowedAreas...,
	)
}

func (s *Service) UploadAndPreviewReader(
	ctx context.Context,
	jobID string,
	sourceName string,
	declaredMIMEType string,
	source io.Reader,
	actorUserID string,
	mapping *DataTransferCSVMappingInput,
	allowedAreas ...TransferArea,
) (DataTransferPreview, error) {
	repository, err := s.importRepository()
	if err != nil {
		return DataTransferPreview{}, err
	}
	jobID = strings.TrimSpace(jobID)
	actorUserID = strings.TrimSpace(actorUserID)
	if jobID == "" || actorUserID == "" || source == nil {
		return DataTransferPreview{}, ErrDataTransferValidation
	}
	job, err := importJob(ctx, repository, jobID)
	if err != nil {
		return DataTransferPreview{}, err
	}
	if job.Direction != TransferImport || !transferpolicy.ValidTransferFormat(job.Format) {
		return DataTransferPreview{}, fmt.Errorf("%w: job is not an import", ErrDataTransferValidation)
	}
	if mapping != nil && (job.Format != TransferCSV || !slices.Contains(job.Areas, TransferVehicles)) {
		return DataTransferPreview{}, fmt.Errorf("%w: manual CSV mapping is only supported for vehicles",
			ErrDataTransferValidation)
	}
	if job.State != TransferJobDraft && job.State != TransferJobReviewRequired && job.State != TransferJobReady {
		return DataTransferPreview{}, fmt.Errorf("%w: import job is %s", ErrDataTransferConflict, job.State)
	}
	if !transferpolicy.DataTransferAreasAllowed(job.Areas, allowedAreas) {
		return DataTransferPreview{}, ErrDataTransferForbidden
	}
	expectedState := job.State
	expectedRevision := job.Revision

	upload, err := s.files.StageUpload(ctx, job.ID, job.Format, sourceName, declaredMIMEType, source)
	if err != nil {
		return DataTransferPreview{}, err
	}
	defer func() { _ = upload.Close() }()
	metadata := upload.Metadata()
	incoming, packageVersion, rowNumbers, csvMapping, err := transferpolicy.ParseDataTransferUpload(job, upload, mapping)
	if err != nil {
		return DataTransferPreview{}, err
	}

	snapshotAreas := slices.Clone(job.Areas)
	if slices.Contains(job.Areas, TransferExhibitionLists) && !slices.Contains(snapshotAreas, TransferVehicles) {
		snapshotAreas = append(snapshotAreas, TransferVehicles)
	}
	current, err := repository.Snapshot(ctx, snapshotAreas)
	if err != nil {
		return DataTransferPreview{}, fmt.Errorf("load current transfer rows: %w", err)
	}
	records, issues := transferpolicy.ClassifyDataTransferImport(job.ID, incoming, current, rowNumbers, s.rules)
	vehicleSets, vehicleSetIssues := transferpolicy.ClassifyDataTransferVehicleSets(job.ID, incoming, current, records)
	issues = append(issues, vehicleSetIssues...)
	records, issues = transferpolicy.SuppressDataTransferVehicleSetMemberConflicts(records, issues, vehicleSets)
	records, issues, err = validateTransferAccessoryPreviewReferences(ctx, repository, job.ID, records, issues)
	if err != nil {
		return DataTransferPreview{}, err
	}
	ready, warnings, failures := transferpolicy.CountDataTransferPreviewRecords(records)
	job.SourceName = metadata.Name
	job.SourceSHA256 = metadata.SHA256
	job.PackageVersion = packageVersion
	job.TotalRecords = len(records)
	job.ReadyRecords = ready
	job.WarningRecords = warnings
	job.ErrorRecords = failures
	job.Stage = "preview"
	if len(csvMapping) > 0 {
		job.Options = transferpolicy.DataTransferOptionsWithCSVMapping(job.Options, csvMapping)
	}
	if len(issues) == 0 {
		job.State = TransferJobReady
	} else {
		job.State = TransferJobReviewRequired
	}
	vehicleFields := []VehicleTransferField(nil)
	if job.Format == TransferCSV && slices.Contains(job.Areas, TransferVehicles) {
		vehicleFields = transferpolicy.VehicleCSVTransferFields()
	}
	job.Preview, err = transferpolicy.DataTransferPreviewMap(metadata.SHA256, records, vehicleSets, csvMapping, vehicleFields)
	if err != nil {
		return DataTransferPreview{}, err
	}
	mutation := DataTransferImportMutation{
		ExpectedState: expectedState, ExpectedRevision: expectedRevision, Job: job,
		Issues: issues, ReplaceIssues: true,
	}
	if mapping != nil && mapping.SaveToProfile && job.ProfileID != "" {
		profile, profileErr := s.profile(ctx, job.ProfileID)
		if profileErr != nil {
			return DataTransferPreview{}, profileErr
		}
		mutation.ProfileOptions = &DataTransferProfileOptionsMutation{
			ProfileID: profile.ID, ExpectedUpdatedAt: profile.UpdatedAt,
			ExpectedOptions: transferpolicy.CloneTransferOptions(profile.Options),
			Options:         transferpolicy.DataTransferOptionsWithCSVMapping(profile.Options, csvMapping),
		}
	}
	job, err = compareAndUpdateDataTransferImport(ctx, repository, mutation)
	if err != nil {
		return DataTransferPreview{}, err
	}
	issues, err = repository.ListIssues(ctx, job.ID)
	if err != nil {
		return DataTransferPreview{}, err
	}
	return transferpolicy.NewDataTransferPreview(job, records, issues, vehicleSets, csvMapping, vehicleFields), nil
}

func validateTransferAccessoryPreviewReferences(
	ctx context.Context,
	repository dataTransferImportRepository,
	jobID string,
	records []DataTransferPreviewRecord,
	issues []DataTransferIssue,
) ([]DataTransferPreviewRecord, []DataTransferIssue, error) {
	for index := range records {
		record := &records[index]
		if record.Area != TransferAccessories || transferpolicy.TransferRecordHasIssue(issues, *record, "invalid_accessory") {
			continue
		}
		accessory := TransferAccessory{}
		if err := json.Unmarshal(record.Data, &accessory); err != nil {
			return nil, nil, fmt.Errorf("%w: invalid accessory preview data", ErrDataTransferValidation)
		}
		if err := repository.ValidateTransferAccessoryReferences(ctx, accessory, record.TargetID); err != nil {
			if !errors.Is(err, ErrDataTransferValidation) {
				return nil, nil, err
			}
			issues = append(issues, transferpolicy.NewTransferIssue(jobID, TransferAccessories, record.RecordKey, record.RowNumber,
				"record", TransferIssueError, "invalid_accessory",
				"Accessory article type or subtype is not active.", "skip"))
			transferpolicy.FinalizeDataTransferRecord(record, transferpolicy.TransferIssuesForPreviewRecord(issues, *record))
		}
	}
	return records, issues, nil
}

func (s *Service) ResolveIssue(
	ctx context.Context,
	jobID string,
	issueID string,
	resolution string,
	actorUserID string,
	allowedAreas ...TransferArea,
) (DataTransferJob, error) {
	repository, err := s.importRepository()
	if err != nil {
		return DataTransferJob{}, err
	}
	if strings.TrimSpace(actorUserID) == "" || strings.TrimSpace(issueID) == "" {
		return DataTransferJob{}, ErrDataTransferValidation
	}
	job, err := importJob(ctx, repository, strings.TrimSpace(jobID))
	if err != nil {
		return DataTransferJob{}, err
	}
	if !transferpolicy.DataTransferAreasAllowed(job.Areas, allowedAreas) {
		return DataTransferJob{}, ErrDataTransferForbidden
	}
	if job.State != TransferJobReviewRequired && job.State != TransferJobReady {
		return DataTransferJob{}, fmt.Errorf("%w: import issues cannot be resolved in state %s", ErrDataTransferConflict, job.State)
	}
	issues, err := repository.ListIssues(ctx, job.ID)
	if err != nil {
		return DataTransferJob{}, err
	}
	found := false
	resolution = strings.TrimSpace(resolution)
	for index := range issues {
		if issues[index].ID != issueID {
			continue
		}
		if !transferpolicy.ValidDataTransferResolution(issues[index], resolution) {
			return DataTransferJob{}, fmt.Errorf("%w: invalid issue resolution", ErrDataTransferValidation)
		}
		issues[index].SelectedResolution = resolution
		issues[index].UpdatedAt = ""
		found = true
		break
	}
	if !found {
		return DataTransferJob{}, ErrDataTransferNotFound
	}
	expectedState := job.State
	expectedRevision := job.Revision
	job.State = TransferJobReady
	for _, issue := range issues {
		if issue.SelectedResolution == "" {
			job.State = TransferJobReviewRequired
			break
		}
	}
	job.Stage = "review"
	return compareAndUpdateDataTransferImport(ctx, repository, DataTransferImportMutation{
		ExpectedState: expectedState, ExpectedRevision: expectedRevision, Job: job,
		Issues: issues, ReplaceIssues: true,
	})
}

func (s *Service) CancelJob(
	ctx context.Context,
	jobID string,
	actorUserID string,
	allowedAreas ...TransferArea,
) (DataTransferJob, error) {
	repository, err := s.importRepository()
	if err != nil {
		return DataTransferJob{}, err
	}
	if strings.TrimSpace(actorUserID) == "" {
		return DataTransferJob{}, ErrDataTransferValidation
	}
	job, err := importJob(ctx, repository, strings.TrimSpace(jobID))
	if err != nil {
		return DataTransferJob{}, err
	}
	if !transferpolicy.DataTransferAreasAllowed(job.Areas, allowedAreas) {
		return DataTransferJob{}, ErrDataTransferForbidden
	}
	if job.State != TransferJobDraft && job.State != TransferJobReviewRequired && job.State != TransferJobReady {
		return DataTransferJob{}, fmt.Errorf("%w: import job cannot be cancelled in state %s", ErrDataTransferConflict, job.State)
	}
	expectedState := job.State
	expectedRevision := job.Revision
	job.State = TransferJobCancelled
	job.Stage = "cancelled"
	job.ResultMessage = "Import cancelled."
	return compareAndUpdateDataTransferImport(ctx, repository, DataTransferImportMutation{
		ExpectedState: expectedState, ExpectedRevision: expectedRevision, Job: job,
	})
}

func (s *Service) ConfirmImport(
	ctx context.Context,
	jobID string,
	expectedRevision int,
	confirm bool,
	actorUserID string,
	allowedAreas ...TransferArea,
) (DataTransferJob, error) {
	return s.ConfirmImportWithPolicy(ctx, jobID, expectedRevision, confirm, actorUserID,
		DataTransferImportPolicy{CanManageExhibitionLists: true}, allowedAreas...)
}

func (s *Service) ConfirmImportWithPolicy(
	ctx context.Context,
	jobID string,
	expectedRevision int,
	confirm bool,
	actorUserID string,
	policy DataTransferImportPolicy,
	allowedAreas ...TransferArea,
) (DataTransferJob, error) {
	repository, err := s.importRepository()
	if err != nil {
		return DataTransferJob{}, err
	}
	actorUserID = strings.TrimSpace(actorUserID)
	if expectedRevision < 1 || !confirm || actorUserID == "" {
		return DataTransferJob{}, fmt.Errorf("%w: explicit import confirmation is required", ErrDataTransferValidation)
	}
	job, err := importJob(ctx, repository, strings.TrimSpace(jobID))
	if err != nil {
		return DataTransferJob{}, err
	}
	if job.Direction != TransferImport {
		return DataTransferJob{}, fmt.Errorf("%w: job is not an import", ErrDataTransferValidation)
	}
	if !transferpolicy.DataTransferAreasAllowed(job.Areas, allowedAreas) {
		return DataTransferJob{}, ErrDataTransferForbidden
	}
	if job.Revision != expectedRevision {
		return DataTransferJob{}, fmt.Errorf("%w: reviewed import revision changed", ErrDataTransferConflict)
	}
	if job.State != TransferJobReady {
		return DataTransferJob{}, fmt.Errorf("%w: import job is not ready", ErrDataTransferConflict)
	}
	if strings.TrimSpace(job.SourceSHA256) == "" || len(job.Preview) == 0 {
		return DataTransferJob{}, fmt.Errorf("%w: import preview is incomplete", ErrDataTransferConflict)
	}
	issues, err := repository.ListIssues(ctx, job.ID)
	if err != nil {
		return DataTransferJob{}, err
	}
	for _, issue := range issues {
		if strings.TrimSpace(issue.SelectedResolution) == "" {
			return DataTransferJob{}, fmt.Errorf("%w: import has unresolved conflicts", ErrDataTransferConflict)
		}
	}
	if err := repository.ApplyImportWithPolicy(ctx, job, actorUserID, policy); err != nil {
		return DataTransferJob{}, err
	}
	return importJob(ctx, repository, job.ID)
}

func compareAndUpdateDataTransferImport(
	ctx context.Context,
	repository dataTransferImportRepository,
	mutation DataTransferImportMutation,
) (DataTransferJob, error) {
	job, err := repository.CompareAndUpdateImportJob(ctx, mutation)
	if errors.Is(err, ErrDataTransferNotFound) {
		return DataTransferJob{}, ErrDataTransferNotFound
	}
	return job, err
}

func (s *Service) importRepository() (dataTransferImportRepository, error) {
	repository, ok := s.repository.(dataTransferImportRepository)
	if !ok || repository == nil {
		return nil, ErrDataTransferExportUnavailable
	}
	return repository, nil
}

func importJob(ctx context.Context, repository DataTransferRepository, id string) (DataTransferJob, error) {
	if id == "" {
		return DataTransferJob{}, ErrDataTransferValidation
	}
	job, err := repository.GetJob(ctx, id)
	if errors.Is(err, ErrDataTransferNotFound) {
		return DataTransferJob{}, ErrDataTransferNotFound
	}
	return job, err
}
