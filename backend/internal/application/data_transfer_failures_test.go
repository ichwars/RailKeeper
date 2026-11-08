package application

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

type failingTransferExport struct {
	*dataTransferExportRepositoryStub
	stage                    string
	failure, recoveryFailure error
	marked                   bool
}

func (s *failingTransferExport) GetProfile(ctx context.Context, id string) (DataTransferProfile, error) {
	if s.stage == "profile" {
		return DataTransferProfile{}, s.failure
	}
	return s.dataTransferExportRepositoryStub.GetProfile(ctx, id)
}
func (s *failingTransferExport) GetJob(ctx context.Context, id string) (DataTransferJob, error) {
	if s.stage == "job" {
		return DataTransferJob{}, s.failure
	}
	return s.dataTransferExportRepositoryStub.GetJob(ctx, id)
}
func (s *failingTransferExport) ClaimExportJob(ctx context.Context, id string) (DataTransferJob, error) {
	if s.stage == "claim" {
		return DataTransferJob{}, s.failure
	}
	return s.dataTransferExportRepositoryStub.ClaimExportJob(ctx, id)
}
func (s *failingTransferExport) Snapshot(ctx context.Context, areas []TransferArea) (DataTransferSnapshot, error) {
	if s.stage == "snapshot" {
		return DataTransferSnapshot{}, s.failure
	}
	return s.dataTransferExportRepositoryStub.Snapshot(ctx, areas)
}
func (s *failingTransferExport) CreateArtifact(ctx context.Context, artifact DataTransferArtifact) (DataTransferArtifact, error) {
	if s.stage == "artifact" {
		return DataTransferArtifact{}, s.failure
	}
	return s.dataTransferExportRepositoryStub.CreateArtifact(ctx, artifact)
}
func (s *failingTransferExport) UpdateJob(ctx context.Context, job DataTransferJob) (DataTransferJob, error) {
	if job.State == TransferJobFailed && s.recoveryFailure != nil {
		return DataTransferJob{}, s.recoveryFailure
	}
	return s.dataTransferExportRepositoryStub.UpdateJob(ctx, job)
}
func (s *failingTransferExport) MarkArtifactDeleted(_ context.Context, _, at string) error {
	s.marked = true
	s.artifact.DeletedAt = at
	return nil
}

func TestDataTransferExportFailuresNeverPublishSuccessOrChangeInventory(t *testing.T) {
	for _, stage := range []string{"profile", "job", "claim", "snapshot", "encoding", "directory", "artifact", "recovery"} {
		t.Run(stage, func(t *testing.T) {
			sentinel := errors.New("synthetic transfer failure")
			original := DataTransferSnapshot{Vehicles: []TransferVehicle{{ID: "vehicle", InventoryNumber: "RK-1", Name: "Recorded vehicle", PurchaseDate: "", PurchasePrice: "129,99"}}}
			repository := &failingTransferExport{dataTransferExportRepositoryStub: &dataTransferExportRepositoryStub{profile: DataTransferProfile{ID: "profile", Enabled: true, Direction: TransferExport, Format: TransferCSV, Areas: []TransferArea{TransferVehicles}}, snapshot: original}, failure: sentinel}
			dataDir := t.TempDir()
			service := NewDataTransferService(repository, dataDir)
			if stage == "profile" {
				repository.stage = stage
				if _, err := service.CreateExportJob(t.Context(), "profile", "actor"); !errors.Is(err, sentinel) || repository.jobNumber != 0 {
					t.Fatal("profile read failure created a job", err)
				}
				return
			}
			job, err := service.CreateExportJob(t.Context(), "profile", "actor")
			if err != nil {
				t.Fatal(err)
			}
			repository.stage = stage
			if stage == "encoding" {
				repository.job.Areas = []TransferArea{TransferVehicles, TransferAccessories}
			}
			if stage == "directory" {
				if err := os.WriteFile(filepath.Join(dataDir, "exports"), []byte("must-remain"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			if stage == "recovery" {
				repository.stage = "snapshot"
				repository.recoveryFailure = errors.New("synthetic state persistence failure")
			}
			result, err := service.ExecuteExport(t.Context(), job.ID)
			if err == nil || result.Job.ID != "" || result.Artifact.ID != "" {
				t.Fatal("failed export acknowledged success")
			}
			if stage == "recovery" && !errors.Is(err, repository.recoveryFailure) {
				t.Fatal("lost recovery persistence failure", err)
			}
			if stage != "encoding" && stage != "directory" && !errors.Is(err, sentinel) {
				t.Fatal("lost original transfer failure", err)
			}
			if !reflect.DeepEqual(repository.snapshot, original) || repository.artifact.ID != "" {
				t.Fatal("failed export changed inventory or published an artifact")
			}
			if stage == "snapshot" || stage == "encoding" || stage == "directory" || stage == "artifact" {
				if repository.job.State != TransferJobFailed {
					t.Fatal("failed export remained acknowledged as running")
				}
			}
			if stage == "artifact" {
				files, err := os.ReadDir(filepath.Join(dataDir, "exports"))
				if err != nil || len(files) != 0 {
					t.Fatal("failed metadata write left published or temporary files", err)
				}
			}
			if stage == "directory" {
				content, err := os.ReadFile(filepath.Join(dataDir, "exports"))
				if err != nil || string(content) != "must-remain" {
					t.Fatal("export failure overwrote an existing filesystem entry", err)
				}
			}
		})
	}
}

func TestDataTransferArtifactActionsRejectDirectoriesAndHandleMissingFiles(t *testing.T) {
	for _, mode := range []string{"directory", "missing", "unavailable-job", "traversal", "invalid-id", "deleted"} {
		t.Run(mode, func(t *testing.T) {
			dataDir := t.TempDir()
			if err := os.Mkdir(filepath.Join(dataDir, "exports"), 0o700); err != nil {
				t.Fatal(err)
			}
			relative := "exports"
			if mode != "directory" {
				relative = "exports/missing.csv"
			}
			if mode == "traversal" {
				relative = "exports/../../outside.csv"
			}
			repository := &failingTransferExport{dataTransferExportRepositoryStub: &dataTransferExportRepositoryStub{job: DataTransferJob{ID: "job", Areas: []TransferArea{TransferVehicles}}, artifact: DataTransferArtifact{ID: "artifact", JobID: "job", RelativePath: relative}}, failure: errors.New("synthetic job read failure")}
			if mode == "unavailable-job" {
				repository.stage = "job"
			}
			if mode == "deleted" {
				repository.artifact.DeletedAt = "2026-11-08T12:00:00Z"
			}
			service := NewDataTransferService(repository, dataDir)
			id := "artifact"
			if mode == "invalid-id" {
				id = "  "
			}
			_, file, err := service.OpenArtifact(t.Context(), id)
			if file != nil {
				_ = file.Close()
				t.Fatal("invalid artifact produced readable content")
			}
			if err == nil {
				t.Fatal("invalid artifact opened")
			}
			if mode == "deleted" && !errors.Is(err, ErrDataTransferArtifactDeleted) {
				t.Fatal("deleted artifact did not retain its tombstone", err)
			}
			err = service.DeleteArtifact(t.Context(), id)
			if mode == "missing" {
				if err != nil || !repository.marked {
					t.Fatal("missing-file metadata could not be retired", err)
				}
			} else if err == nil || repository.marked {
				t.Fatal("rejected artifact action changed metadata", err)
			}
			info, err := os.Stat(filepath.Join(dataDir, "exports"))
			if err != nil || !info.IsDir() {
				t.Fatal("artifact action removed its directory", err)
			}
		})
	}
}

type failingTransferQuery struct {
	*dataTransferQueryRepositoryStub
	stage   string
	failure error
}

func (s *failingTransferQuery) GetJob(ctx context.Context, id string) (DataTransferJob, error) {
	if s.stage == "job" {
		return DataTransferJob{}, s.failure
	}
	return s.dataTransferQueryRepositoryStub.GetJob(ctx, id)
}
func (s *failingTransferQuery) ListJobs(ctx context.Context, filter DataTransferJobFilter) ([]DataTransferJob, error) {
	if s.stage == "jobs" {
		return nil, s.failure
	}
	return s.dataTransferQueryRepositoryStub.ListJobs(ctx, filter)
}
func (s *failingTransferQuery) ListArtifacts(ctx context.Context) ([]DataTransferArtifact, error) {
	if s.stage == "artifacts" {
		return nil, s.failure
	}
	return s.dataTransferQueryRepositoryStub.ListArtifacts(ctx)
}
func (s *failingTransferQuery) ListIssues(ctx context.Context, id string) ([]DataTransferIssue, error) {
	if s.stage == "issues" {
		return nil, s.failure
	}
	return s.dataTransferQueryRepositoryStub.ListIssues(ctx, id)
}

func TestDataTransferQueriesDoNotReportPartialResultsOnReadFailure(t *testing.T) {
	for _, stage := range []string{"job", "jobs", "artifacts", "issues"} {
		t.Run(stage, func(t *testing.T) {
			sentinel := errors.New("synthetic transfer read failure")
			repository := &failingTransferQuery{dataTransferQueryRepositoryStub: &dataTransferQueryRepositoryStub{
				jobs: []DataTransferJob{{ID: "job", State: TransferJobFailed, Areas: []TransferArea{TransferVehicles}, TotalRecords: 42}},
			}, stage: stage, failure: sentinel}
			service := NewDataTransferService(repository, t.TempDir())
			if stage == "jobs" || stage == "artifacts" {
				result, err := service.Summary(t.Context())
				if !errors.Is(err, sentinel) || result != (DataTransferSummary{}) {
					t.Fatal("read failure reported a partial summary", result, err)
				}
			}
			if stage != "jobs" {
				result, err := service.GetJobDetails(t.Context(), "job")
				if !errors.Is(err, sentinel) || result.Job.ID != "" || result.Issues != nil || result.Artifacts != nil {
					t.Fatal("read failure reported partial job details", result, err)
				}
			}
			if stage == "job" {
				if _, err := service.RetryJob(t.Context(), "job", "actor"); !errors.Is(err, sentinel) || len(repository.jobs) != 1 {
					t.Fatal("read failure created a replacement job", err)
				}
				if err := service.DeleteJob(t.Context(), "job"); !errors.Is(err, sentinel) || len(repository.deletedJobIDs) != 0 {
					t.Fatal("read failure deleted a job", err)
				}
			}
		})
	}
}

func TestDataTransferFolderDoesNotInvokePlatformOpenerForBlockedPath(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "exports")
	if err := os.WriteFile(path, []byte("existing entry"), 0o600); err != nil {
		t.Fatal(err)
	}
	called := false
	service := NewDataTransferService(nil, directory, WithDataTransferFolderOpener(true,
		func(context.Context, string) error { called = true; return nil }))
	if err := service.OpenArtifactFolder(t.Context()); err == nil || called {
		t.Fatal("blocked export path reached the platform opener", err)
	}
	content, err := os.ReadFile(path)
	if err != nil || string(content) != "existing entry" {
		t.Fatal("folder action changed an existing entry", err)
	}
}

func TestDataTransferJobCreationRequiresActorBeforeStorageAccess(t *testing.T) {
	service := NewDataTransferService(nil, t.TempDir())
	operations := map[string]func() error{
		"export": func() error { _, err := service.CreateExportJob(t.Context(), "profile", "  "); return err },
		"import": func() error { _, err := service.CreateImportJob(t.Context(), "profile", "  "); return err },
		"retry":  func() error { _, err := service.RetryJob(t.Context(), "job", "  "); return err },
	}
	for name, operation := range operations {
		t.Run(name, func(t *testing.T) {
			if err := operation(); !errors.Is(err, ErrDataTransferValidation) {
				t.Fatal("unattributed job reached storage", err)
			}
		})
	}
}
