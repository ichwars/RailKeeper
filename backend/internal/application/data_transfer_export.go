package application

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
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

type DataTransferServiceOption func(*DataTransferService)

func WithDataTransferFolderOpener(available bool, opener DataTransferFolderOpener) DataTransferServiceOption {
	return func(service *DataTransferService) {
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

func (s *DataTransferService) CreateExportJob(
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

func (s *DataTransferService) ExecuteExport(
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
	if errors.Is(err, sql.ErrNoRows) {
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
	artifact, err := s.writeExportArtifact(ctx, repository, job.ID, displayName, mimeType, payload)
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

func (s *DataTransferService) OpenArtifact(
	ctx context.Context,
	id string,
	allowedAreas ...TransferArea,
) (DataTransferArtifact, *os.File, error) {
	repository, err := s.exportRepository()
	if err != nil {
		return DataTransferArtifact{}, nil, err
	}
	artifact, err := s.authorizedArtifact(ctx, repository, id, allowedAreas)
	if err != nil {
		return DataTransferArtifact{}, nil, err
	}
	if _, err := ensureDataTransferExportDirectory(s.dataDir); err != nil {
		return DataTransferArtifact{}, nil, err
	}
	path, err := resolveDataTransferArtifactPath(s.dataDir, artifact.RelativePath)
	if err != nil {
		return DataTransferArtifact{}, nil, err
	}
	info, err := os.Lstat(path)
	if err != nil {
		return DataTransferArtifact{}, nil, fmt.Errorf("inspect transfer artifact: %w", err)
	}
	if !info.Mode().IsRegular() {
		return DataTransferArtifact{}, nil, ErrDataTransferArtifactPath
	}
	file, err := os.Open(path)
	if err != nil {
		return DataTransferArtifact{}, nil, fmt.Errorf("open transfer artifact: %w", err)
	}
	return artifact, file, nil
}

func (s *DataTransferService) DeleteArtifact(ctx context.Context, id string) error {
	repository, err := s.exportRepository()
	if err != nil {
		return err
	}
	artifact, err := s.authorizedArtifact(ctx, repository, id, nil)
	if err != nil {
		return err
	}
	if _, err := ensureDataTransferExportDirectory(s.dataDir); err != nil {
		return err
	}
	path, err := resolveDataTransferArtifactPath(s.dataDir, artifact.RelativePath)
	if err != nil {
		return err
	}
	info, err := os.Lstat(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect transfer artifact for deletion: %w", err)
	}
	if err == nil {
		if !info.Mode().IsRegular() {
			return ErrDataTransferArtifactPath
		}
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("delete transfer artifact: %w", err)
		}
	}
	return repository.MarkArtifactDeleted(ctx, artifact.ID, time.Now().UTC().Format(time.RFC3339))
}

func (s *DataTransferService) OpenFolderAvailable() bool {
	return s != nil && s.openFolderAvailable && s.openFolder != nil
}

func (s *DataTransferService) OpenArtifactFolder(ctx context.Context) error {
	if !s.OpenFolderAvailable() {
		return ErrDataTransferOpenUnavailable
	}
	directory, err := ensureDataTransferExportDirectory(s.dataDir)
	if err != nil {
		return err
	}
	resolved, err := resolveDataTransferArtifactPath(s.dataDir, "exports")
	if err != nil || resolved != directory {
		return ErrDataTransferArtifactPath
	}
	return s.openFolder(ctx, resolved)
}

func (s *DataTransferService) exportRepository() (dataTransferExportRepository, error) {
	repository, ok := s.repository.(dataTransferExportRepository)
	if !ok || repository == nil {
		return nil, ErrDataTransferExportUnavailable
	}
	return repository, nil
}

func (s *DataTransferService) authorizedArtifact(
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
	if errors.Is(err, sql.ErrNoRows) {
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

func (s *DataTransferService) writeExportArtifact(
	ctx context.Context,
	repository dataTransferExportRepository,
	jobID string,
	displayName string,
	mimeType string,
	payload []byte,
) (artifact DataTransferArtifact, err error) {
	directory, err := ensureDataTransferExportDirectory(s.dataDir)
	if err != nil {
		return DataTransferArtifact{}, err
	}
	temporary, err := os.CreateTemp(directory, ".railkeeper-export-*.tmp")
	if err != nil {
		return DataTransferArtifact{}, fmt.Errorf("create transfer artifact temp file: %w", err)
	}
	temporaryPath := temporary.Name()
	defer func() {
		_ = temporary.Close()
		if err != nil {
			_ = os.Remove(temporaryPath)
		}
	}()
	hash := sha256.New()
	written, err := io.Copy(io.MultiWriter(temporary, hash), bytes.NewReader(payload))
	if err != nil {
		return DataTransferArtifact{}, fmt.Errorf("write transfer artifact: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		return DataTransferArtifact{}, fmt.Errorf("sync transfer artifact: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return DataTransferArtifact{}, fmt.Errorf("close transfer artifact: %w", err)
	}
	fileName := safeDataTransferArtifactName(jobID+"-"+randomID(), filepath.Ext(displayName))
	relativePath := filepath.ToSlash(filepath.Join("exports", fileName))
	finalPath, err := resolveDataTransferArtifactPath(s.dataDir, relativePath)
	if err != nil {
		return DataTransferArtifact{}, err
	}
	if filepath.Dir(finalPath) != directory {
		return DataTransferArtifact{}, ErrDataTransferArtifactPath
	}
	if err := os.Rename(temporaryPath, finalPath); err != nil {
		return DataTransferArtifact{}, fmt.Errorf("publish transfer artifact: %w", err)
	}
	artifact, err = repository.CreateArtifact(ctx, DataTransferArtifact{
		JobID: jobID, RelativePath: relativePath, DisplayName: displayName, MIMEType: mimeType,
		SizeBytes: written, SHA256: hex.EncodeToString(hash.Sum(nil)),
	})
	if err != nil {
		_ = os.Remove(finalPath)
		return DataTransferArtifact{}, err
	}
	return artifact, nil
}

func (s *DataTransferService) failExportJob(
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

func ensureDataTransferExportDirectory(dataDir string) (string, error) {
	directory, err := resolveDataTransferArtifactPath(dataDir, "exports")
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(directory, 0o750); err != nil {
		return "", fmt.Errorf("create transfer export directory: %w", err)
	}
	info, err := os.Lstat(directory)
	if err != nil {
		return "", fmt.Errorf("inspect transfer export directory: %w", err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return "", ErrDataTransferArtifactPath
	}
	return directory, nil
}

func resolveDataTransferArtifactPath(dataDir, relativePath string) (string, error) {
	dataDir = strings.TrimSpace(dataDir)
	relativePath = strings.TrimSpace(relativePath)
	if dataDir == "" || relativePath == "" || filepath.IsAbs(relativePath) {
		return "", ErrDataTransferArtifactPath
	}
	segments := strings.FieldsFunc(relativePath, func(value rune) bool { return value == '/' || value == '\\' })
	if len(segments) == 0 || !strings.EqualFold(segments[0], "exports") {
		return "", ErrDataTransferArtifactPath
	}
	for _, segment := range segments {
		if segment == "" || segment == "." || segment == ".." {
			return "", ErrDataTransferArtifactPath
		}
	}
	base, err := filepath.Abs(dataDir)
	if err != nil {
		return "", fmt.Errorf("resolve transfer data directory: %w", err)
	}
	exportsDirectory := filepath.Join(base, "exports")
	target, err := filepath.Abs(filepath.Join(base, filepath.FromSlash(strings.ReplaceAll(relativePath, "\\", "/"))))
	if err != nil {
		return "", fmt.Errorf("resolve transfer artifact: %w", err)
	}
	if target != exportsDirectory && !strings.HasPrefix(target, exportsDirectory+string(os.PathSeparator)) {
		return "", ErrDataTransferArtifactPath
	}
	return target, nil
}

func safeDataTransferArtifactName(jobID, extension string) string {
	var builder strings.Builder
	for _, value := range strings.ToLower(strings.TrimSpace(jobID)) {
		if value >= 'a' && value <= 'z' || value >= '0' && value <= '9' || value == '-' || value == '_' {
			builder.WriteRune(value)
		}
	}
	if builder.Len() == 0 {
		builder.WriteString("export")
	}
	return "railkeeper-" + builder.String() + extension
}
