package application

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
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

func (s *DataTransferService) CreateImportJob(
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

func (s *DataTransferService) UploadAndPreview(
	ctx context.Context,
	jobID string,
	sourceName string,
	payload []byte,
	actorUserID string,
	allowedAreas ...TransferArea,
) (DataTransferPreview, error) {
	return s.UploadAndPreviewWithMapping(ctx, jobID, sourceName, payload, actorUserID, nil, allowedAreas...)
}

func (s *DataTransferService) UploadAndPreviewWithMapping(
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

func (s *DataTransferService) UploadAndPreviewReader(
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

	cleanName, err := validateDataTransferUploadName(job.Format, sourceName)
	if err != nil {
		return DataTransferPreview{}, err
	}
	temporaryFile, digest, detectedMIMEType, err := s.storeDataTransferUpload(job.ID, source)
	if err != nil {
		return DataTransferPreview{}, err
	}
	defer func() { _ = os.Remove(temporaryFile) }()
	if err := validateDataTransferUploadMIME(job.Format, declaredMIMEType, detectedMIMEType); err != nil {
		return DataTransferPreview{}, err
	}

	file, err := os.Open(temporaryFile)
	if err != nil {
		return DataTransferPreview{}, fmt.Errorf("open transfer upload: %w", err)
	}
	defer func() { _ = file.Close() }()
	incoming, packageVersion, rowNumbers, csvMapping, err := transferpolicy.ParseDataTransferUpload(job, file, mapping)
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
	records, issues := classifyDataTransferImport(job.ID, incoming, current, rowNumbers)
	vehicleSets, vehicleSetIssues := transferpolicy.ClassifyDataTransferVehicleSets(job.ID, incoming, current, records)
	issues = append(issues, vehicleSetIssues...)
	records, issues = transferpolicy.SuppressDataTransferVehicleSetMemberConflicts(records, issues, vehicleSets)
	records, issues, err = validateTransferAccessoryPreviewReferences(ctx, repository, job.ID, records, issues)
	if err != nil {
		return DataTransferPreview{}, err
	}
	ready, warnings, failures := transferpolicy.CountDataTransferPreviewRecords(records)
	job.SourceName = cleanName
	job.SourceSHA256 = digest
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
		vehicleFields = VehicleCSVTransferFields()
	}
	job.Preview, err = transferpolicy.DataTransferPreviewMap(digest, records, vehicleSets, csvMapping, vehicleFields)
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
			if !errors.Is(err, ErrAccessoryValidation) {
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

func (s *DataTransferService) ResolveIssue(
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

func (s *DataTransferService) CancelJob(
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

func (s *DataTransferService) ConfirmImport(
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

func (s *DataTransferService) ConfirmImportWithPolicy(
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
	if errors.Is(err, sql.ErrNoRows) {
		return DataTransferJob{}, ErrDataTransferNotFound
	}
	return job, err
}

func (s *DataTransferService) importRepository() (dataTransferImportRepository, error) {
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
	if errors.Is(err, sql.ErrNoRows) {
		return DataTransferJob{}, ErrDataTransferNotFound
	}
	return job, err
}

func (s *DataTransferService) storeDataTransferUpload(
	jobID string,
	source io.Reader,
) (path string, digest string, mimeType string, err error) {
	directory, err := ensureDataTransferImportDirectory(s.dataDir)
	if err != nil {
		return "", "", "", err
	}
	file, err := os.CreateTemp(directory, safeDataTransferImportName(jobID)+"-*.upload")
	if err != nil {
		return "", "", "", fmt.Errorf("create transfer upload temp file: %w", err)
	}
	path = file.Name()
	remove := true
	defer func() {
		if closeErr := file.Close(); err == nil && closeErr != nil {
			err = fmt.Errorf("close transfer upload: %w", closeErr)
		}
		if remove || err != nil {
			_ = os.Remove(path)
		}
	}()
	hash := sha256.New()
	written, err := io.Copy(io.MultiWriter(file, hash), io.LimitReader(source, DataTransferMaxUploadBytes+1))
	if err != nil {
		return "", "", "", fmt.Errorf("store transfer upload: %w", err)
	}
	if written > DataTransferMaxUploadBytes {
		return "", "", "", ErrDataTransferUploadTooLarge
	}
	if err := file.Sync(); err != nil {
		return "", "", "", fmt.Errorf("sync transfer upload: %w", err)
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", "", "", fmt.Errorf("rewind transfer upload: %w", err)
	}
	prefix := make([]byte, 512)
	read, readErr := file.Read(prefix)
	if readErr != nil && !errors.Is(readErr, io.EOF) {
		return "", "", "", fmt.Errorf("inspect transfer upload: %w", readErr)
	}
	remove = false
	return path, hex.EncodeToString(hash.Sum(nil)), http.DetectContentType(prefix[:read]), nil
}

func ensureDataTransferImportDirectory(dataDir string) (string, error) {
	base, err := filepath.Abs(strings.TrimSpace(dataDir))
	if err != nil || strings.TrimSpace(dataDir) == "" {
		return "", ErrDataTransferUploadPath
	}
	directory := filepath.Join(base, "imports")
	if err := os.MkdirAll(directory, 0o750); err != nil {
		return "", fmt.Errorf("create transfer import directory: %w", err)
	}
	info, err := os.Lstat(directory)
	if err != nil {
		return "", fmt.Errorf("inspect transfer import directory: %w", err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return "", ErrDataTransferUploadPath
	}
	return directory, nil
}

func safeDataTransferImportName(jobID string) string {
	name := strings.TrimSuffix(safeDataTransferArtifactName(jobID, ""), "")
	return strings.TrimPrefix(name, "railkeeper-")
}

func validateDataTransferUploadName(format TransferFormat, sourceName string) (string, error) {
	sourceName = strings.TrimSpace(strings.ReplaceAll(sourceName, "\\", "/"))
	sourceName = filepath.Base(sourceName)
	if sourceName == "" || sourceName == "." {
		return "", fmt.Errorf("%w: upload filename is required", ErrDataTransferValidation)
	}
	extension := strings.ToLower(filepath.Ext(sourceName))
	if format == TransferCSV && extension != ".csv" || format == TransferJSON && extension != ".json" {
		return "", fmt.Errorf("%w: upload extension does not match job format", ErrDataTransferValidation)
	}
	return sourceName, nil
}

func validateDataTransferUploadMIME(format TransferFormat, declared, detected string) error {
	declared = normalizeTransferMIME(declared)
	detected = normalizeTransferMIME(detected)
	allowedDeclared := map[TransferFormat]map[string]bool{
		TransferCSV: {
			"": true, "application/octet-stream": true, "text/csv": true, "text/plain": true,
			"application/vnd.ms-excel": true,
		},
		TransferJSON: {"": true, "application/octet-stream": true, "application/json": true, "text/plain": true},
	}
	allowedDetected := map[TransferFormat]map[string]bool{
		TransferCSV:  {"text/plain": true, "text/csv": true, "application/octet-stream": true},
		TransferJSON: {"application/json": true, "text/plain": true, "application/octet-stream": true},
	}
	if !allowedDeclared[format][declared] || !allowedDetected[format][detected] {
		return fmt.Errorf("%w: upload MIME type does not match job format", ErrDataTransferValidation)
	}
	return nil
}

func normalizeTransferMIME(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if separator := strings.IndexByte(value, ';'); separator >= 0 {
		value = strings.TrimSpace(value[:separator])
	}
	return value
}
