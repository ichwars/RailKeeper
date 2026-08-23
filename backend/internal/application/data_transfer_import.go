package application

import (
	"bufio"
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
)

const DataTransferMaxUploadBytes int64 = 25 * 1024 * 1024

var (
	ErrDataTransferUploadTooLarge = errors.New("data transfer upload exceeds the size limit")
	ErrDataTransferUploadPath     = errors.New("data transfer upload path is not confined")
)

type DataTransferPreview struct {
	Job            DataTransferJob                `json:"job"`
	Records        []DataTransferPreviewRecord    `json:"records"`
	Issues         []DataTransferIssue            `json:"issues"`
	TotalRecords   int                            `json:"totalRecords"`
	ReadyRecords   int                            `json:"readyRecords"`
	WarningRecords int                            `json:"warningRecords"`
	ErrorRecords   int                            `json:"errorRecords"`
	CSVMapping     []DataTransferCSVColumnMapping `json:"csvMapping,omitempty"`
	VehicleFields  []VehicleTransferField         `json:"vehicleFields,omitempty"`
}

type DataTransferPreviewRecord struct {
	Area              TransferArea    `json:"area"`
	RecordKey         string          `json:"recordKey"`
	RowNumber         *int            `json:"rowNumber,omitempty"`
	Classification    string          `json:"classification"`
	ProposedAction    string          `json:"proposedAction"`
	TargetID          string          `json:"targetId,omitempty"`
	TargetUpdatedAt   string          `json:"targetUpdatedAt,omitempty"`
	TargetFingerprint string          `json:"targetFingerprint,omitempty"`
	Data              json.RawMessage `json:"data"`
}

func DataTransferTargetFingerprint(value any) string {
	fingerprintValue := value
	switch accessory := value.(type) {
	case TransferAccessory:
		fingerprintValue = struct {
			TransferAccessory
			FingerprintState TransferAccessoryFingerprintState `json:"fingerprintState"`
		}{TransferAccessory: accessory, FingerprintState: accessory.FingerprintState}
	case *TransferAccessory:
		if accessory != nil {
			fingerprintValue = struct {
				TransferAccessory
				FingerprintState TransferAccessoryFingerprintState `json:"fingerprintState"`
			}{TransferAccessory: *accessory, FingerprintState: accessory.FingerprintState}
		}
	}
	payload, err := json.Marshal(fingerprintValue)
	if err != nil {
		return ""
	}
	digest := sha256.Sum256(payload)
	return hex.EncodeToString(digest[:])
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
	if !dataTransferAreasAllowed(profile.Areas, allowedAreas) {
		return DataTransferJob{}, ErrDataTransferForbidden
	}
	return s.repository.CreateJob(ctx, DataTransferJob{
		ProfileID: profile.ID, ProfileName: profile.Name, Direction: TransferImport, Format: profile.Format,
		Areas: slices.Clone(profile.Areas), Options: cloneTransferOptions(profile.Options), State: TransferJobDraft,
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
	if job.Direction != TransferImport || !validTransferFormat(job.Format) {
		return DataTransferPreview{}, fmt.Errorf("%w: job is not an import", ErrDataTransferValidation)
	}
	if mapping != nil && (job.Format != TransferCSV || !slices.Contains(job.Areas, TransferVehicles)) {
		return DataTransferPreview{}, fmt.Errorf("%w: manual CSV mapping is only supported for vehicles",
			ErrDataTransferValidation)
	}
	if job.State != TransferJobDraft && job.State != TransferJobReviewRequired && job.State != TransferJobReady {
		return DataTransferPreview{}, fmt.Errorf("%w: import job is %s", ErrDataTransferConflict, job.State)
	}
	if !dataTransferAreasAllowed(job.Areas, allowedAreas) {
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
	incoming, packageVersion, rowNumbers, csvMapping, err := parseDataTransferUpload(job, file, mapping)
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
	records, issues, err = validateTransferAccessoryPreviewReferences(ctx, repository, job.ID, records, issues)
	if err != nil {
		return DataTransferPreview{}, err
	}
	ready, warnings, failures := countDataTransferPreviewRecords(records)
	job.SourceName = cleanName
	job.SourceSHA256 = digest
	job.PackageVersion = packageVersion
	job.TotalRecords = len(records)
	job.ReadyRecords = ready
	job.WarningRecords = warnings
	job.ErrorRecords = failures
	job.Stage = "preview"
	if len(csvMapping) > 0 {
		job.Options = dataTransferOptionsWithCSVMapping(job.Options, csvMapping)
	}
	if len(issues) == 0 {
		job.State = TransferJobReady
	} else {
		job.State = TransferJobReviewRequired
	}
	vehicleFields := []VehicleTransferField(nil)
	if job.Format == TransferCSV && slices.Contains(job.Areas, TransferVehicles) {
		vehicleFields = VehicleTransferFields()
	}
	job.Preview, err = dataTransferPreviewMap(digest, records, csvMapping, vehicleFields)
	if err != nil {
		return DataTransferPreview{}, err
	}
	job, err = compareAndUpdateDataTransferImport(ctx, repository, DataTransferImportMutation{
		ExpectedState: expectedState, ExpectedRevision: expectedRevision, Job: job,
		Issues: issues, ReplaceIssues: true,
	})
	if err != nil {
		return DataTransferPreview{}, err
	}
	if mapping != nil && mapping.SaveToProfile && job.ProfileID != "" {
		profile, profileErr := s.profile(ctx, job.ProfileID)
		if profileErr != nil {
			return DataTransferPreview{}, profileErr
		}
		profile.Options = dataTransferOptionsWithCSVMapping(profile.Options, csvMapping)
		if _, profileErr = repository.UpdateProfile(ctx, profile); profileErr != nil {
			return DataTransferPreview{}, fmt.Errorf("save CSV mapping to profile: %w", profileErr)
		}
	}
	issues, err = repository.ListIssues(ctx, job.ID)
	if err != nil {
		return DataTransferPreview{}, err
	}
	return newDataTransferPreview(job, records, issues, csvMapping, vehicleFields), nil
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
		if record.Area != TransferAccessories || transferRecordHasIssue(issues, *record, "invalid_accessory") {
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
			issues = append(issues, newTransferIssue(jobID, TransferAccessories, record.RecordKey, record.RowNumber,
				"record", TransferIssueError, "invalid_accessory",
				"Accessory article type or subtype is not active.", "skip"))
			finalizeDataTransferRecord(record, transferIssuesForPreviewRecord(issues, *record))
		}
	}
	return records, issues, nil
}

func transferRecordHasIssue(
	issues []DataTransferIssue,
	record DataTransferPreviewRecord,
	code string,
) bool {
	for _, issue := range issues {
		if issue.Area == record.Area && issue.RecordKey == record.RecordKey && issue.Code == code &&
			transferPreviewRowsEqual(issue.RowNumber, record.RowNumber) {
			return true
		}
	}
	return false
}

func transferIssuesForPreviewRecord(
	issues []DataTransferIssue,
	record DataTransferPreviewRecord,
) []DataTransferIssue {
	matched := make([]DataTransferIssue, 0)
	for _, issue := range issues {
		if issue.Area == record.Area && issue.RecordKey == record.RecordKey &&
			transferPreviewRowsEqual(issue.RowNumber, record.RowNumber) {
			matched = append(matched, issue)
		}
	}
	return matched
}

func transferPreviewRowsEqual(left, right *int) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
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
	if !dataTransferAreasAllowed(job.Areas, allowedAreas) {
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
		if !validDataTransferResolution(issues[index], resolution) {
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
	if !dataTransferAreasAllowed(job.Areas, allowedAreas) {
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
	if !dataTransferAreasAllowed(job.Areas, allowedAreas) {
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

func parseDataTransferUpload(
	job DataTransferJob,
	reader io.Reader,
	mapping *DataTransferCSVMappingInput,
) (DataTransferSnapshot, int, map[TransferArea][]int, []DataTransferCSVColumnMapping, error) {
	switch job.Format {
	case TransferJSON:
		packageDocument, err := decodeDataTransferPackage(reader)
		if err != nil {
			return DataTransferSnapshot{}, 0, nil, nil, err
		}
		if err := validateDataTransferPackageSelection(packageDocument.Areas, job.Areas); err != nil {
			return DataTransferSnapshot{}, 0, nil, nil, err
		}
		return DataTransferSnapshot{
			Vehicles: packageDocument.Areas.Vehicles, Accessories: packageDocument.Areas.Accessories,
			ExhibitionLists: packageDocument.Areas.ExhibitionLists,
		}, packageDocument.Version, nil, nil, nil
	case TransferCSV:
		if len(job.Areas) != 1 {
			return DataTransferSnapshot{}, 0, nil, nil, ErrDataTransferValidation
		}
		snapshot, rows, effectiveMapping, err := parseDataTransferCSVWithMapping(
			job.Areas[0], reader, mapping, dataTransferCSVMappingDefaults(job.Options),
		)
		return snapshot, DataTransferPackageVersion, map[TransferArea][]int{job.Areas[0]: rows}, effectiveMapping, err
	default:
		return DataTransferSnapshot{}, 0, nil, nil, ErrDataTransferValidation
	}
}

func decodeDataTransferPackage(reader io.Reader) (DataTransferPackage, error) {
	decoder := json.NewDecoder(bufio.NewReader(reader))
	decoder.DisallowUnknownFields()
	document := DataTransferPackage{}
	if err := decoder.Decode(&document); err != nil {
		return DataTransferPackage{}, fmt.Errorf("%w: invalid transfer package: %v", ErrDataTransferValidation, err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return DataTransferPackage{}, fmt.Errorf("%w: transfer package has trailing data", ErrDataTransferValidation)
	}
	if document.Format != DataTransferPackageFormat || document.Version != DataTransferPackageVersion {
		return DataTransferPackage{}, fmt.Errorf("%w: unsupported transfer package format or version", ErrDataTransferValidation)
	}
	return document, nil
}

func validateDataTransferPackageSelection(areas DataTransferPackageAreas, selected []TransferArea) error {
	present := []TransferArea{}
	if areas.Vehicles != nil {
		present = append(present, TransferVehicles)
	}
	if areas.Accessories != nil {
		present = append(present, TransferAccessories)
	}
	if areas.ExhibitionLists != nil {
		present = append(present, TransferExhibitionLists)
	}
	if len(present) == 0 || len(present) != len(selected) {
		return fmt.Errorf("%w: package areas do not match the import profile", ErrDataTransferValidation)
	}
	for _, area := range selected {
		if !slices.Contains(present, area) {
			return fmt.Errorf("%w: package areas do not match the import profile", ErrDataTransferValidation)
		}
	}
	return nil
}

func dataTransferPreviewMap(
	sourceSHA256 string,
	records []DataTransferPreviewRecord,
	csvMapping []DataTransferCSVColumnMapping,
	vehicleFields []VehicleTransferField,
) (map[string]any, error) {
	payload, err := json.Marshal(struct {
		SourceSHA256  string                         `json:"sourceSha256"`
		Records       []DataTransferPreviewRecord    `json:"records"`
		CSVMapping    []DataTransferCSVColumnMapping `json:"csvMapping,omitempty"`
		VehicleFields []VehicleTransferField         `json:"vehicleFields,omitempty"`
	}{SourceSHA256: sourceSHA256, Records: records, CSVMapping: csvMapping, VehicleFields: vehicleFields})
	if err != nil {
		return nil, fmt.Errorf("encode transfer preview: %w", err)
	}
	preview := map[string]any{}
	if err := json.Unmarshal(payload, &preview); err != nil {
		return nil, fmt.Errorf("persist transfer preview: %w", err)
	}
	return preview, nil
}

func newDataTransferPreview(
	job DataTransferJob,
	records []DataTransferPreviewRecord,
	issues []DataTransferIssue,
	csvMapping []DataTransferCSVColumnMapping,
	vehicleFields []VehicleTransferField,
) DataTransferPreview {
	return DataTransferPreview{
		Job: job, Records: records, Issues: issues, TotalRecords: job.TotalRecords,
		ReadyRecords: job.ReadyRecords, WarningRecords: job.WarningRecords, ErrorRecords: job.ErrorRecords,
		CSVMapping: csvMapping, VehicleFields: vehicleFields,
	}
}

func validDataTransferResolution(issue DataTransferIssue, resolution string) bool {
	allowed := map[string]map[string]bool{
		"missing_inventory_number":             {"skip": true},
		"missing_manufacturer":                 {"skip": true},
		"missing_name":                         {"skip": true},
		"missing_gauge":                        {"skip": true},
		"missing_category":                     {"skip": true},
		"missing_gattung":                      {"skip": true},
		"invalid_vehicle":                      {"skip": true},
		"invalid_accessory":                    {"skip": true},
		"duplicate_inventory_number":           {"replace": true, "copy": true, "skip": true},
		"matching_manufacturer_article_number": {"use_existing": true, "create": true, "skip": true},
		"duplicate_exhibition_list":            {"replace": true, "merge": true, "copy": true, "skip": true},
		"locked_exhibition_list":               {"copy": true, "skip": true},
		"exhibition_vehicle_reference":         {"link": true, "skip": true},
		"missing_vehicle_reference":            {"skip": true},
		"duplicate_input_inventory_number":     {"skip": true},
	}
	return resolution != "" && allowed[issue.Code][resolution]
}
