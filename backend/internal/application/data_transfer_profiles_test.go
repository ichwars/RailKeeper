package application

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"testing"
)

func TestDataTransferProfileRejectsInvalidCSVCombination(t *testing.T) {
	service := NewDataTransferService(profileRepositoryStub{}, t.TempDir())
	for _, areas := range [][]TransferArea{
		{TransferVehicles, TransferAccessories},
		{TransferArea("masterData")},
	} {
		_, err := service.CreateProfile(t.Context(), CreateDataTransferProfileInput{
			Name: "Invalid", Direction: TransferExport, Format: TransferCSV, Areas: areas,
		}, "admin-1")
		if !errors.Is(err, ErrDataTransferValidation) {
			t.Fatalf("expected validation error for %#v, got %v", areas, err)
		}
	}
}

func TestDataTransferProfileServiceUpdatesDisablesAndListsProfiles(t *testing.T) {
	repository := &dataTransferProfileRepositoryStub{profiles: []DataTransferProfile{
		{ID: "profile-1", Name: "Vehicle export", Direction: TransferExport, Format: TransferCSV,
			Areas: []TransferArea{TransferVehicles}, Enabled: true, CreatedByUserID: "admin-1"},
	}}
	service := NewDataTransferService(repository, t.TempDir())

	profiles, err := service.ListProfiles(t.Context())
	if err != nil || len(profiles) != 1 || profiles[0].ID != "profile-1" {
		t.Fatalf("unexpected listed profiles: %#v, %v", profiles, err)
	}

	updated, err := service.UpdateProfile(t.Context(), " profile-1 ", UpdateDataTransferProfileInput{
		Name: " Vehicle JSON export ", Direction: TransferExport, Format: TransferJSON,
		Areas: []TransferArea{TransferVehicles, TransferAccessories},
	})
	if err != nil || updated.Name != "Vehicle JSON export" || len(updated.Areas) != 2 || !updated.Enabled {
		t.Fatalf("unexpected updated profile: %#v, %v", updated, err)
	}

	disabled, err := service.DisableProfile(t.Context(), "profile-1")
	if err != nil || disabled.Enabled {
		t.Fatalf("unexpected disabled profile: %#v, %v", disabled, err)
	}
}

func TestDataTransferProfileRejectsDuplicateActiveNameInSameDirection(t *testing.T) {
	repository := &dataTransferProfileRepositoryStub{profiles: []DataTransferProfile{
		{ID: "import-1", Name: "Fahrzeugimport", Direction: TransferImport, Enabled: true},
		{ID: "import-disabled", Name: "Archivimport", Direction: TransferImport, Enabled: false},
		{ID: "export-1", Name: "Fahrzeugliste", Direction: TransferExport, Enabled: true},
	}}
	service := NewDataTransferService(repository, t.TempDir())

	_, err := service.CreateProfile(t.Context(), CreateDataTransferProfileInput{
		Name: " fahrzeugIMPORT ", Direction: TransferImport, Format: TransferCSV,
		Areas: []TransferArea{TransferVehicles},
	}, "admin-1")
	if !errors.Is(err, ErrDataTransferConflict) {
		t.Fatalf("duplicate active import error = %v, want conflict", err)
	}

	created, err := service.CreateProfile(t.Context(), CreateDataTransferProfileInput{
		Name: "Fahrzeugimport", Direction: TransferExport, Format: TransferCSV,
		Areas: []TransferArea{TransferVehicles},
	}, "admin-1")
	if err != nil || created.Direction != TransferExport {
		t.Fatalf("same name in other direction should be allowed: %#v, %v", created, err)
	}

	created, err = service.CreateProfile(t.Context(), CreateDataTransferProfileInput{
		Name: "Archivimport", Direction: TransferImport, Format: TransferCSV,
		Areas: []TransferArea{TransferVehicles},
	}, "admin-1")
	if err != nil || created.Name != "Archivimport" {
		t.Fatalf("name of disabled profile should be reusable: %#v, %v", created, err)
	}
}

func TestDataTransferProfileRejectsDuplicateNameOnUpdate(t *testing.T) {
	repository := &dataTransferProfileRepositoryStub{profiles: []DataTransferProfile{
		{ID: "import-1", Name: "Fahrzeugimport", Direction: TransferImport, Format: TransferCSV,
			Areas: []TransferArea{TransferVehicles}, Enabled: true},
		{ID: "import-2", Name: "Lieferantenimport", Direction: TransferImport, Format: TransferCSV,
			Areas: []TransferArea{TransferVehicles}, Enabled: true},
	}}
	service := NewDataTransferService(repository, t.TempDir())

	_, err := service.UpdateProfile(t.Context(), "import-2", UpdateDataTransferProfileInput{
		Name: "Fahrzeugimport", Direction: TransferImport, Format: TransferCSV,
		Areas: []TransferArea{TransferVehicles},
	})
	if !errors.Is(err, ErrDataTransferConflict) {
		t.Fatalf("duplicate profile update error = %v, want conflict", err)
	}
}

func TestDataTransferSummaryCountsOpenScopedJobsAndActiveArtifacts(t *testing.T) {
	repository := &dataTransferQueryRepositoryStub{
		jobs: []DataTransferJob{
			{ID: "draft", Direction: TransferImport, Areas: []TransferArea{TransferVehicles},
				State: TransferJobDraft, TotalRecords: 2},
			{ID: "review", Direction: TransferImport, Areas: []TransferArea{TransferExhibitionLists},
				State: TransferJobReviewRequired, TotalRecords: 3},
			{ID: "export-old", Direction: TransferExport, Areas: []TransferArea{TransferExhibitionLists},
				State: TransferJobCompleted, CompletedAt: "2026-08-19T10:00:00Z"},
			{ID: "export-new", Direction: TransferExport, Areas: []TransferArea{TransferExhibitionLists},
				State: TransferJobCompletedWithWarnings, CompletedAt: "2026-08-20T11:00:00Z"},
		},
		artifacts: []DataTransferArtifact{
			{ID: "vehicle", JobID: "draft", SizeBytes: 5},
			{ID: "messe", JobID: "export-new", SizeBytes: 7},
			{ID: "deleted", JobID: "export-old", SizeBytes: 11, DeletedAt: "2026-08-20T12:00:00Z"},
		},
	}
	service := NewDataTransferService(repository, t.TempDir())

	summary, err := service.Summary(t.Context(), TransferExhibitionLists)
	if err != nil {
		t.Fatal(err)
	}
	if summary.OpenJobs != 1 || summary.SelectedRecords != 3 ||
		summary.LastExportAt != "2026-08-20T11:00:00Z" ||
		summary.ArtifactCount != 1 || summary.ArtifactBytes != 7 || summary.OpenFolderAvailable {
		t.Fatalf("unexpected scoped summary: %#v", summary)
	}
	if summary.ArtifactDirectory != "RailKeeper/Exporte" {
		t.Fatalf("summary exposed or changed the display directory: %q", summary.ArtifactDirectory)
	}
}

func TestRetryImportClearsApprovalPreviewIssuesAndResult(t *testing.T) {
	original := DataTransferJob{
		ID: "completed-import", ProfileID: "profile-1", ProfileName: "Vehicle import",
		Direction: TransferImport, Format: TransferCSV, Areas: []TransferArea{TransferVehicles},
		Options: map[string]any{"delimiter": ";"}, State: TransferJobCompleted, Stage: "completed",
		SourceName: "vehicles.csv", SourceSHA256: "source-hash", PackageVersion: 7,
		TotalRecords: 2, ReadyRecords: 2, Preview: map[string]any{"records": []any{"one"}},
		ConfirmedByUserID: "editor-old", ConfirmedAt: "2026-08-20T10:00:00Z",
		CompletedAt: "2026-08-20T10:01:00Z", ResultMessage: "Imported.",
	}
	repository := &dataTransferQueryRepositoryStub{
		jobs: []DataTransferJob{original},
		issues: map[string][]DataTransferIssue{
			original.ID: {{ID: "issue-1", JobID: original.ID, Severity: TransferIssueWarning}},
		},
		artifacts: []DataTransferArtifact{{ID: "artifact-1", JobID: original.ID}},
	}
	service := NewDataTransferService(repository, t.TempDir())

	retry, err := service.RetryJob(t.Context(), original.ID, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	if retry.ID == original.ID || retry.ProfileID != "" || retry.ProfileName != original.ProfileName ||
		retry.Direction != original.Direction || retry.Format != original.Format ||
		!reflect.DeepEqual(retry.Areas, original.Areas) || !reflect.DeepEqual(retry.Options, original.Options) {
		t.Fatalf("retry did not copy the immutable selection: %#v", retry)
	}
	if retry.ConfirmedAt != "" || retry.ConfirmedByUserID != "" || retry.State != TransferJobDraft ||
		retry.Preview == nil || len(retry.Preview) != 0 || retry.SourceName != "" || retry.SourceSHA256 != "" ||
		retry.TotalRecords != 0 || retry.ReadyRecords != 0 || retry.CompletedAt != "" || retry.ResultMessage != "" {
		t.Fatalf("retry inherited prior execution state: %#v", retry)
	}
	if retry.CreatedByUserID != "editor-1" || retry.Stage != "created" {
		t.Fatalf("retry actor or stage = %#v", retry)
	}
	if len(repository.issues[retry.ID]) != 0 {
		t.Fatalf("retry inherited issues: %#v", repository.issues[retry.ID])
	}
	for _, artifact := range repository.artifacts {
		if artifact.JobID == retry.ID {
			t.Fatalf("retry inherited artifact: %#v", artifact)
		}
	}
}

func TestRetryRejectsOpenJob(t *testing.T) {
	repository := &dataTransferQueryRepositoryStub{jobs: []DataTransferJob{{
		ID: "ready", Direction: TransferImport, Format: TransferCSV, Areas: []TransferArea{TransferVehicles},
		State: TransferJobReady,
	}}}
	service := NewDataTransferService(repository, t.TempDir())
	if _, err := service.RetryJob(t.Context(), "ready", "editor-1"); !errors.Is(err, ErrDataTransferConflict) {
		t.Fatalf("retry open job error = %v, want conflict", err)
	}
}

type profileRepositoryStub struct {
	DataTransferRepository
}

func (profileRepositoryStub) CreateProfile(
	context.Context,
	DataTransferProfile,
) (DataTransferProfile, error) {
	return DataTransferProfile{}, nil
}

type dataTransferProfileRepositoryStub struct {
	DataTransferRepository
	profiles []DataTransferProfile
}

func (s *dataTransferProfileRepositoryStub) CreateProfile(
	_ context.Context,
	profile DataTransferProfile,
) (DataTransferProfile, error) {
	profile.ID = "created-profile"
	s.profiles = append(s.profiles, profile)
	return profile, nil
}

type dataTransferQueryRepositoryStub struct {
	DataTransferRepository
	jobs      []DataTransferJob
	issues    map[string][]DataTransferIssue
	artifacts []DataTransferArtifact
	nextID    int
}

func (s *dataTransferQueryRepositoryStub) GetJob(_ context.Context, id string) (DataTransferJob, error) {
	for _, job := range s.jobs {
		if job.ID == id {
			return job, nil
		}
	}
	return DataTransferJob{}, sql.ErrNoRows
}

func (s *dataTransferQueryRepositoryStub) ListJobs(
	context.Context,
	DataTransferJobFilter,
) ([]DataTransferJob, error) {
	return s.jobs, nil
}

func (s *dataTransferQueryRepositoryStub) ListArtifacts(context.Context) ([]DataTransferArtifact, error) {
	return s.artifacts, nil
}

func (s *dataTransferQueryRepositoryStub) ListIssues(_ context.Context, jobID string) ([]DataTransferIssue, error) {
	return s.issues[jobID], nil
}

func (s *dataTransferQueryRepositoryStub) CreateJob(
	_ context.Context,
	job DataTransferJob,
) (DataTransferJob, error) {
	s.nextID++
	job.ID = "retry-" + string(rune('0'+s.nextID))
	job.Revision = 1
	s.jobs = append(s.jobs, job)
	return job, nil
}

func (s *dataTransferProfileRepositoryStub) GetProfile(_ context.Context, id string) (DataTransferProfile, error) {
	for _, profile := range s.profiles {
		if profile.ID == id {
			return profile, nil
		}
	}
	return DataTransferProfile{}, ErrDataTransferNotFound
}

func (s *dataTransferProfileRepositoryStub) ListProfiles(context.Context) ([]DataTransferProfile, error) {
	return s.profiles, nil
}

func (s *dataTransferProfileRepositoryStub) UpdateProfile(
	_ context.Context,
	profile DataTransferProfile,
) (DataTransferProfile, error) {
	for index := range s.profiles {
		if s.profiles[index].ID == profile.ID {
			s.profiles[index] = profile
			return profile, nil
		}
	}
	return DataTransferProfile{}, ErrDataTransferNotFound
}
