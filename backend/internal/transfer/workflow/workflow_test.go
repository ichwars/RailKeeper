package workflow

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	transferpolicy "railkeeper/backend/internal/transfer/policy"
)

type workflowRepository struct {
	DataTransferRepository
	job       DataTransferJob
	artifact  DataTransferArtifact
	mutations int
	reads     int
	failure   error
}

func (r *workflowRepository) GetJob(context.Context, string) (DataTransferJob, error) {
	return r.job, nil
}
func (r *workflowRepository) GetArtifact(context.Context, string) (DataTransferArtifact, error) {
	return r.artifact, nil
}
func (r *workflowRepository) Snapshot(context.Context, []TransferArea) (DataTransferSnapshot, error) {
	r.reads++
	return DataTransferSnapshot{}, nil
}
func (r *workflowRepository) ClaimExportJob(context.Context, string) (DataTransferJob, error) {
	r.mutations++
	return r.job, nil
}
func (r *workflowRepository) CompareAndUpdateImportJob(_ context.Context, mutation DataTransferImportMutation) (DataTransferJob, error) {
	if r.failure != nil {
		return DataTransferJob{}, r.failure
	}
	r.mutations++
	r.job = mutation.Job
	return r.job, nil
}
func (r *workflowRepository) ApplyImport(context.Context, DataTransferJob, string) error {
	r.mutations++
	return nil
}
func (r *workflowRepository) ApplyImportWithPolicy(context.Context, DataTransferJob, string, DataTransferImportPolicy) error {
	r.mutations++
	return nil
}
func (r *workflowRepository) ValidateTransferAccessoryReferences(context.Context, TransferAccessory, string) error {
	return nil
}

type workflowUpload struct {
	*strings.Reader
	closes int
}

func (u *workflowUpload) Close() error { u.closes++; return nil }
func (u *workflowUpload) Metadata() UploadMetadata {
	return UploadMetadata{Name: "rows.csv", SHA256: "synthetic-sha", DetectedMIME: "text/plain"}
}

type workflowFiles struct {
	FileStore
	upload Upload
	stages int
	error  error
}

func (s *workflowFiles) StageUpload(context.Context, string, TransferFormat, string, string, io.Reader) (Upload, error) {
	s.stages++
	return s.upload, s.error
}

func TestDeniedTransferAreasHaveNoStorageOrFileEffects(t *testing.T) {
	for _, operation := range []string{"export", "upload", "resolve", "cancel", "confirm", "download"} {
		t.Run(operation, func(t *testing.T) {
			repository := &workflowRepository{job: DataTransferJob{
				ID: "job", Direction: TransferImport, Format: TransferCSV,
				Areas: []TransferArea{TransferVehicles}, State: TransferJobReady, Revision: 2,
				SourceSHA256: "reviewed-sha", Preview: map[string]any{"records": []any{}},
			}, artifact: DataTransferArtifact{ID: "artifact", JobID: "job"}}
			files := &workflowFiles{}
			service := New(repository, files, transferpolicy.AggregateRules{})
			var err error
			switch operation {
			case "export":
				repository.job.Direction = TransferExport
				repository.job.State = TransferJobDraft
				_, err = service.ExecuteExport(t.Context(), "job", TransferExhibitionLists)
			case "upload":
				_, err = service.UploadAndPreview(t.Context(), "job", "rows.csv", []byte("a;b\n"), "actor", TransferExhibitionLists)
			case "resolve":
				_, err = service.ResolveIssue(t.Context(), "job", "issue", "skip", "actor", TransferExhibitionLists)
			case "cancel":
				_, err = service.CancelJob(t.Context(), "job", "actor", TransferExhibitionLists)
			case "confirm":
				_, err = service.ConfirmImport(t.Context(), "job", 2, true, "actor", TransferExhibitionLists)
			case "download":
				_, _, err = service.OpenArtifact(t.Context(), "artifact", TransferExhibitionLists)
			}
			if !errors.Is(err, ErrDataTransferForbidden) || repository.mutations != 0 || repository.reads != 0 || files.stages != 0 {
				t.Fatal("area denial did not precede storage/file effects", err)
			}
		})
	}
}

func TestPreviewClosesUploadOnParseAndPersistenceFailures(t *testing.T) {
	for _, phase := range []string{"parse", "persist", "stage"} {
		t.Run(phase, func(t *testing.T) {
			payload := "Inventarnummer;Hersteller;Bezeichnung;Spurweite;Kategorie;Gattung\nRK-1;Roco;Lok;H0;Lokomotive;Dampflokomotive\n"
			if phase == "parse" {
				payload = "\"unterminated CSV field"
			}
			upload := &workflowUpload{Reader: strings.NewReader(payload)}
			files := &workflowFiles{upload: upload}
			repository := &workflowRepository{job: DataTransferJob{
				ID: "job", Direction: TransferImport, Format: TransferCSV,
				Areas: []TransferArea{TransferVehicles}, State: TransferJobDraft, Revision: 1,
			}}
			sentinel := errors.New("synthetic boundary failure")
			if phase == "persist" {
				repository.failure = sentinel
			}
			if phase == "stage" {
				files.error, files.upload = sentinel, nil
			}
			service := New(repository, files, transferpolicy.AggregateRules{Vehicle: func(TransferVehicle) error { return nil }})
			preview, err := service.UploadAndPreview(t.Context(), "job", "rows.csv", []byte(payload), "actor")
			if err == nil || preview.Job.ID != "" || repository.mutations != 0 || repository.job.State != TransferJobDraft {
				t.Fatal("failed preview changed or acknowledged a job", err)
			}
			if phase != "parse" && !errors.Is(err, sentinel) {
				t.Fatal("lost boundary failure", err)
			}
			wantCloses := 1
			if phase == "stage" {
				wantCloses = 0
			}
			if upload.closes != wantCloses || files.stages != 1 {
				t.Fatal("upload lifetime was not bounded by the preview", upload.closes, files.stages)
			}
		})
	}
}

func TestOptionalRepositoryCapabilitiesRemainExplicit(t *testing.T) {
	// A base port must not acquire unsupported import/export operations through a wrapper.
	service := New(struct{ DataTransferRepository }{}, nil, transferpolicy.AggregateRules{})
	if _, err := service.ExecuteExport(t.Context(), "job"); !errors.Is(err, ErrDataTransferExportUnavailable) {
		t.Fatal("base repository advertised export capability", err)
	}
	if _, err := service.UploadAndPreview(t.Context(), "job", "rows.csv", nil, "actor"); !errors.Is(err, ErrDataTransferExportUnavailable) {
		t.Fatal("base repository advertised import capability", err)
	}
}
