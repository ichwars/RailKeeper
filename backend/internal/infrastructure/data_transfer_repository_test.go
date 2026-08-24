package infrastructure_test

import (
	"database/sql"
	"errors"
	"path/filepath"
	"sync"
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/infrastructure"
)

func TestDataTransferRepositoryPersistsJobSnapshot(t *testing.T) {
	db := testDB(t)
	repo := infrastructure.NewDataTransferRepository(db)
	profile, err := repo.CreateProfile(t.Context(), application.DataTransferProfile{
		Name:      "Vollständige Sicherung",
		Direction: application.TransferExport,
		Format:    application.TransferJSON,
		Areas: []application.TransferArea{
			application.TransferVehicles,
			application.TransferAccessories,
			application.TransferExhibitionLists,
		},
		Enabled: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	job, err := repo.CreateJob(t.Context(), application.DataTransferJob{
		ProfileID:   profile.ID,
		ProfileName: profile.Name,
		Direction:   profile.Direction,
		Format:      profile.Format,
		Areas:       profile.Areas,
		State:       application.TransferJobDraft,
	})
	if err != nil {
		t.Fatal(err)
	}
	if job.ProfileName != profile.Name || len(job.Areas) != 3 {
		t.Fatalf("unexpected snapshot: %#v", job)
	}
}

func TestDataTransferImportMutationRollsBackJobAndIssuesTogether(t *testing.T) {
	repo := infrastructure.NewDataTransferRepository(testDB(t))
	job, err := repo.CreateJob(t.Context(), application.DataTransferJob{
		Direction: application.TransferImport, Format: application.TransferCSV,
		Areas: []application.TransferArea{application.TransferVehicles}, State: application.TransferJobDraft,
		Stage: "created", Preview: map[string]any{"marker": "old"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.ReplaceIssues(t.Context(), job.ID, []application.DataTransferIssue{{
		JobID: job.ID, Area: application.TransferVehicles, RecordKey: "old",
		Severity: application.TransferIssueWarning, Code: "old", Message: "old",
	}}); err != nil {
		t.Fatal(err)
	}
	updated := job
	updated.State = application.TransferJobReady
	updated.Stage = "preview"
	updated.Preview = map[string]any{"marker": "new"}
	_, err = repo.CompareAndUpdateImportJob(t.Context(), application.DataTransferImportMutation{
		ExpectedState: job.State, ExpectedRevision: job.Revision, Job: updated, ReplaceIssues: true,
		Issues: []application.DataTransferIssue{{
			JobID: job.ID, Area: application.TransferVehicles, RecordKey: "new",
			Severity: application.TransferIssueSeverity("invalid"), Code: "new", Message: "new",
		}},
	})
	if err == nil {
		t.Fatal("expected issue constraint failure")
	}
	loaded, err := repo.GetJob(t.Context(), job.ID)
	if err != nil {
		t.Fatal(err)
	}
	issues, err := repo.ListIssues(t.Context(), job.ID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.State != application.TransferJobDraft || loaded.Revision != job.Revision ||
		loaded.Preview["marker"] != "old" || len(issues) != 1 || issues[0].RecordKey != "old" {
		t.Fatalf("failed mutation left mixed state: job=%#v issues=%#v", loaded, issues)
	}
}

func TestDataTransferImportMutationConcurrentPreviewsCommitOneConsistentResult(t *testing.T) {
	repo := infrastructure.NewDataTransferRepository(testDB(t))
	job, err := repo.CreateJob(t.Context(), application.DataTransferJob{
		Direction: application.TransferImport, Format: application.TransferCSV,
		Areas: []application.TransferArea{application.TransferVehicles}, State: application.TransferJobDraft,
		Stage: "created", Preview: map[string]any{},
	})
	if err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	errorsByMarker := make(chan struct {
		marker string
		err    error
	}, 2)
	var wait sync.WaitGroup
	for _, marker := range []string{"first", "second"} {
		marker := marker
		wait.Add(1)
		go func() {
			defer wait.Done()
			<-start
			updated := job
			updated.State = application.TransferJobReviewRequired
			updated.Stage = "preview"
			updated.Preview = map[string]any{"marker": marker}
			_, mutationErr := repo.CompareAndUpdateImportJob(t.Context(), application.DataTransferImportMutation{
				ExpectedState: job.State, ExpectedRevision: job.Revision, Job: updated, ReplaceIssues: true,
				Issues: []application.DataTransferIssue{{
					JobID: job.ID, Area: application.TransferVehicles, RecordKey: marker,
					Severity: application.TransferIssueWarning, Code: "duplicate", Message: marker,
				}},
			})
			errorsByMarker <- struct {
				marker string
				err    error
			}{marker: marker, err: mutationErr}
		}()
	}
	close(start)
	wait.Wait()
	close(errorsByMarker)
	successes := 0
	conflicts := 0
	for result := range errorsByMarker {
		switch {
		case result.err == nil:
			successes++
		case errors.Is(result.err, application.ErrDataTransferConflict):
			conflicts++
		default:
			t.Fatalf("%s mutation failed unexpectedly: %v", result.marker, result.err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("concurrent mutations: successes=%d conflicts=%d", successes, conflicts)
	}
	loaded, err := repo.GetJob(t.Context(), job.ID)
	if err != nil {
		t.Fatal(err)
	}
	issues, err := repo.ListIssues(t.Context(), job.ID)
	if err != nil {
		t.Fatal(err)
	}
	marker, _ := loaded.Preview["marker"].(string)
	if marker == "" || len(issues) != 1 || issues[0].RecordKey != marker || loaded.Revision != job.Revision+1 {
		t.Fatalf("preview and issues do not belong to one mutation: job=%#v issues=%#v", loaded, issues)
	}
}

func TestDataTransferImportMutationRejectsStalePreviewAfterCancel(t *testing.T) {
	repo := infrastructure.NewDataTransferRepository(testDB(t))
	job, err := repo.CreateJob(t.Context(), application.DataTransferJob{
		Direction: application.TransferImport, Format: application.TransferCSV,
		Areas: []application.TransferArea{application.TransferVehicles}, State: application.TransferJobDraft,
		Stage: "created", Preview: map[string]any{},
	})
	if err != nil {
		t.Fatal(err)
	}
	cancelled := job
	cancelled.State = application.TransferJobCancelled
	cancelled.Stage = "cancelled"
	if _, err := repo.CompareAndUpdateImportJob(t.Context(), application.DataTransferImportMutation{
		ExpectedState: job.State, ExpectedRevision: job.Revision, Job: cancelled,
	}); err != nil {
		t.Fatal(err)
	}
	stalePreview := job
	stalePreview.State = application.TransferJobReady
	stalePreview.Stage = "preview"
	stalePreview.Preview = map[string]any{"marker": "stale"}
	_, err = repo.CompareAndUpdateImportJob(t.Context(), application.DataTransferImportMutation{
		ExpectedState: job.State, ExpectedRevision: job.Revision, Job: stalePreview, ReplaceIssues: true,
	})
	if !errors.Is(err, application.ErrDataTransferConflict) {
		t.Fatalf("stale preview error = %v, want conflict", err)
	}
	loaded, err := repo.GetJob(t.Context(), job.ID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.State != application.TransferJobCancelled || loaded.Preview["marker"] != nil {
		t.Fatalf("cancelled job was resurrected: %#v", loaded)
	}
}

func TestDataTransferRepositoryPersistsAndUpdatesProfiles(t *testing.T) {
	repo := infrastructure.NewDataTransferRepository(testDB(t))
	created, err := repo.CreateProfile(t.Context(), application.DataTransferProfile{
		Name: "Fahrzeuge CSV", Direction: application.TransferExport, Format: application.TransferCSV,
		Areas:   []application.TransferArea{application.TransferVehicles},
		Options: map[string]any{"includeImages": false}, CreatedByUserID: "user-1", Enabled: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.ID == "" || created.CreatedAt == "" || created.UpdatedAt == "" || !created.Enabled {
		t.Fatalf("unexpected created profile: %#v", created)
	}

	created.Name = "Fahrzeuge CSV ohne Bilder"
	created.Options = map[string]any{"includeImages": false, "delimiter": ";"}
	created.Enabled = false
	created.LastUsedAt = "2026-08-20T10:15:00Z"
	updated, err := repo.UpdateProfile(t.Context(), created)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != created.Name || updated.Enabled || updated.LastUsedAt != created.LastUsedAt ||
		updated.Options["delimiter"] != ";" {
		t.Fatalf("unexpected updated profile: %#v", updated)
	}

	profiles, err := repo.ListProfiles(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(profiles) != 1 || profiles[0].ID != created.ID {
		t.Fatalf("unexpected profiles: %#v", profiles)
	}
	found, err := repo.GetProfile(t.Context(), created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if found.Options == nil || found.Areas == nil || found.CreatedByUserID != "user-1" {
		t.Fatalf("unexpected loaded profile: %#v", found)
	}
}

func TestDataTransferRepositoryRejectsDuplicateActiveProfileNameAtomically(t *testing.T) {
	repo := infrastructure.NewDataTransferRepository(testDB(t))
	first, err := repo.CreateProfile(t.Context(), application.DataTransferProfile{
		Name: "Import Fahrzeuge", Direction: application.TransferImport, Format: application.TransferCSV,
		Areas: []application.TransferArea{application.TransferVehicles}, Enabled: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = repo.CreateProfile(t.Context(), application.DataTransferProfile{
		Name: " import fahrzeuge ", Direction: application.TransferImport, Format: application.TransferJSON,
		Areas: []application.TransferArea{application.TransferVehicles}, Enabled: true,
	})
	if !errors.Is(err, application.ErrDataTransferConflict) {
		t.Fatalf("duplicate create error = %v, want conflict", err)
	}
	disabled, err := repo.CreateProfile(t.Context(), application.DataTransferProfile{
		Name: first.Name, Direction: application.TransferImport, Format: application.TransferCSV,
		Areas: []application.TransferArea{application.TransferVehicles}, Enabled: false,
	})
	if err != nil {
		t.Fatal(err)
	}
	disabled.Enabled = true
	if _, err := repo.UpdateProfile(t.Context(), disabled); !errors.Is(err, application.ErrDataTransferConflict) {
		t.Fatalf("duplicate enable error = %v, want conflict", err)
	}
}

func TestDataTransferRepositorySerializesConcurrentDuplicateProfileCreates(t *testing.T) {
	repo := infrastructure.NewDataTransferRepository(testDB(t))
	start := make(chan struct{})
	results := make(chan error, 2)
	var wait sync.WaitGroup
	for _, name := range []string{"Änderung", "änderung"} {
		name := name
		wait.Add(1)
		go func() {
			defer wait.Done()
			<-start
			_, err := repo.CreateProfile(t.Context(), application.DataTransferProfile{
				Name: name, Direction: application.TransferImport,
				Format: application.TransferCSV, Areas: []application.TransferArea{application.TransferVehicles},
				Enabled: true,
			})
			results <- err
		}()
	}
	close(start)
	wait.Wait()
	close(results)
	successes, conflicts := 0, 0
	for err := range results {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, application.ErrDataTransferConflict):
			conflicts++
		default:
			t.Fatalf("concurrent profile create failed unexpectedly: %v", err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("concurrent creates: successes=%d conflicts=%d", successes, conflicts)
	}
}

func TestDataTransferImportMutationRollsBackJobWhenProfileMappingIsStale(t *testing.T) {
	repo := infrastructure.NewDataTransferRepository(testDB(t))
	profile, err := repo.CreateProfile(t.Context(), application.DataTransferProfile{
		Name: "Import Fahrzeuge", Direction: application.TransferImport, Format: application.TransferCSV,
		Areas: []application.TransferArea{application.TransferVehicles}, Options: map[string]any{"marker": "old"},
		Enabled: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	job, err := repo.CreateJob(t.Context(), application.DataTransferJob{
		ProfileID: profile.ID, ProfileName: profile.Name, Direction: application.TransferImport,
		Format: application.TransferCSV, Areas: profile.Areas, State: application.TransferJobDraft,
		Stage: "created", Preview: map[string]any{"marker": "old"},
	})
	if err != nil {
		t.Fatal(err)
	}
	updated := job
	updated.State = application.TransferJobReady
	updated.Stage = "preview"
	updated.Preview = map[string]any{"marker": "new"}
	_, err = repo.CompareAndUpdateImportJob(t.Context(), application.DataTransferImportMutation{
		ExpectedState: job.State, ExpectedRevision: job.Revision, Job: updated,
		ProfileOptions: &application.DataTransferProfileOptionsMutation{
			ProfileID: profile.ID, ExpectedUpdatedAt: "stale", ExpectedOptions: profile.Options,
			Options: map[string]any{"marker": "new"},
		},
	})
	if !errors.Is(err, application.ErrDataTransferConflict) {
		t.Fatalf("stale profile mapping error = %v, want conflict", err)
	}
	loadedJob, err := repo.GetJob(t.Context(), job.ID)
	if err != nil {
		t.Fatal(err)
	}
	loadedProfile, err := repo.GetProfile(t.Context(), profile.ID)
	if err != nil {
		t.Fatal(err)
	}
	if loadedJob.State != application.TransferJobDraft || loadedJob.Preview["marker"] != "old" ||
		loadedProfile.Options["marker"] != "old" {
		t.Fatalf("stale mapping left partial state: job=%#v profile=%#v", loadedJob, loadedProfile)
	}
	updated = loadedJob
	updated.State = application.TransferJobReady
	updated.Stage = "preview"
	updated.Preview = map[string]any{"marker": "new"}
	if _, err := repo.CompareAndUpdateImportJob(t.Context(), application.DataTransferImportMutation{
		ExpectedState: loadedJob.State, ExpectedRevision: loadedJob.Revision, Job: updated,
		ProfileOptions: &application.DataTransferProfileOptionsMutation{
			ProfileID: loadedProfile.ID, ExpectedUpdatedAt: loadedProfile.UpdatedAt,
			ExpectedOptions: loadedProfile.Options, Options: map[string]any{"marker": "new"},
		},
	}); err != nil {
		t.Fatal(err)
	}
	loadedProfile, err = repo.GetProfile(t.Context(), profile.ID)
	if err != nil {
		t.Fatal(err)
	}
	if loadedProfile.Options["marker"] != "new" {
		t.Fatalf("profile mapping was not committed with preview: %#v", loadedProfile)
	}
}

func TestDataTransferRepositoryFiltersAndUpdatesJobs(t *testing.T) {
	repo := infrastructure.NewDataTransferRepository(testDB(t))
	job, err := repo.CreateJob(t.Context(), application.DataTransferJob{
		ProfileName: "Import Fahrzeuge", Direction: application.TransferImport, Format: application.TransferJSON,
		Areas: []application.TransferArea{application.TransferVehicles}, State: application.TransferJobDraft,
		Stage: "file", SourceName: "fahrzeuge.json", SourceSHA256: "sha-1", PackageVersion: 1,
		Preview: map[string]any{"newRecords": 2}, CreatedByUserID: "editor-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	job.State = application.TransferJobReady
	job.Stage = "review"
	job.TotalRecords = 2
	job.ReadyRecords = 2
	job.ConfirmedByUserID = "editor-1"
	job.ConfirmedAt = "2026-08-20T11:00:00Z"
	updated, err := repo.UpdateJob(t.Context(), job)
	if err != nil {
		t.Fatal(err)
	}
	if updated.State != application.TransferJobReady || updated.ReadyRecords != 2 ||
		updated.Preview["newRecords"] != float64(2) {
		t.Fatalf("unexpected updated job: %#v", updated)
	}

	jobs, err := repo.ListJobs(t.Context(), application.DataTransferJobFilter{
		Direction: application.TransferImport, States: []application.TransferJobState{application.TransferJobReady},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 1 || jobs[0].ID != job.ID {
		t.Fatalf("unexpected jobs: %#v", jobs)
	}
	found, err := repo.GetJob(t.Context(), job.ID)
	if err != nil {
		t.Fatal(err)
	}
	if found.Areas == nil || found.Options == nil || found.Preview == nil || found.ConfirmedAt == "" {
		t.Fatalf("unexpected loaded job: %#v", found)
	}
}

func TestDataTransferRepositoryDeleteJobCascadesRelatedRows(t *testing.T) {
	db := testDB(t)
	repo := infrastructure.NewDataTransferRepository(db)
	job, err := repo.CreateJob(t.Context(), application.DataTransferJob{
		Direction: application.TransferImport, Format: application.TransferCSV,
		Areas: []application.TransferArea{application.TransferVehicles}, State: application.TransferJobCancelled,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.ReplaceIssues(t.Context(), job.ID, []application.DataTransferIssue{{
		JobID: job.ID, Area: application.TransferVehicles, Severity: application.TransferIssueWarning,
		Code: "cancelled", Message: "Cancelled preview issue",
	}}); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.CreateArtifact(t.Context(), application.DataTransferArtifact{
		JobID: job.ID, RelativePath: "exports/cancelled.csv", DisplayName: "cancelled.csv",
		MIMEType: "text/csv", SHA256: "cancelled",
	}); err != nil {
		t.Fatal(err)
	}

	if err := repo.DeleteJob(t.Context(), job.ID); err != nil {
		t.Fatal(err)
	}
	for _, table := range []string{"data_transfer_jobs", "data_transfer_job_issues", "data_transfer_artifacts"} {
		var count int
		if err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Fatalf("%s count = %d, want 0", table, count)
		}
	}
}

func TestDataTransferRepositoryDeleteJobReturnsNotFound(t *testing.T) {
	repo := infrastructure.NewDataTransferRepository(testDB(t))
	if err := repo.DeleteJob(t.Context(), "missing"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("delete missing job error = %v, want sql.ErrNoRows", err)
	}
}

func TestDataTransferRepositoryReplacesIssuesAndMarksArtifactsDeleted(t *testing.T) {
	db := testDB(t)
	repo := infrastructure.NewDataTransferRepository(db)
	job, err := repo.CreateJob(t.Context(), application.DataTransferJob{
		Direction: application.TransferImport, Format: application.TransferCSV,
		Areas: []application.TransferArea{application.TransferAccessories}, State: application.TransferJobDraft,
	})
	if err != nil {
		t.Fatal(err)
	}
	rowNumber := 7
	issues := []application.DataTransferIssue{{
		JobID: job.ID, Area: application.TransferAccessories, RecordKey: "4711", RowNumber: &rowNumber,
		Field: "articleNumber", Severity: application.TransferIssueWarning, Code: "duplicate",
		Message: "Artikel existiert bereits", ProposedResolution: "update", SelectedResolution: "create",
	}}
	if err := repo.ReplaceIssues(t.Context(), job.ID, issues); err != nil {
		t.Fatal(err)
	}
	loadedIssues, err := repo.ListIssues(t.Context(), job.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(loadedIssues) != 1 || loadedIssues[0].RowNumber == nil || *loadedIssues[0].RowNumber != rowNumber ||
		loadedIssues[0].Severity != application.TransferIssueWarning {
		t.Fatalf("unexpected issues: %#v", loadedIssues)
	}

	artifact, err := repo.CreateArtifact(t.Context(), application.DataTransferArtifact{
		JobID: job.ID, RelativePath: "data-transfers/export-1.json", DisplayName: "Export.json",
		MIMEType: "application/json", SizeBytes: 44, SHA256: "digest-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if artifact.ID == "" || artifact.CreatedAt == "" || artifact.DeletedAt != "" {
		t.Fatalf("unexpected artifact: %#v", artifact)
	}
	if err := repo.MarkArtifactDeleted(t.Context(), artifact.ID, "2026-08-20T12:00:00Z"); err != nil {
		t.Fatal(err)
	}
	var deletedAt string
	if err := db.QueryRow(`SELECT COALESCE(deleted_at, '') FROM data_transfer_artifacts WHERE id=?`, artifact.ID).
		Scan(&deletedAt); err != nil {
		t.Fatal(err)
	}
	if deletedAt != "2026-08-20T12:00:00Z" {
		t.Fatalf("artifact deletion timestamp = %q", deletedAt)
	}
}

func TestDataTransferRepositoryAllowsOnlyOneActiveArtifactPerJob(t *testing.T) {
	repo := infrastructure.NewDataTransferRepository(testDB(t))
	job, err := repo.CreateJob(t.Context(), application.DataTransferJob{
		Direction: application.TransferExport, Format: application.TransferCSV,
		Areas: []application.TransferArea{application.TransferVehicles}, State: application.TransferJobRunning,
	})
	if err != nil {
		t.Fatal(err)
	}
	first, err := repo.CreateArtifact(t.Context(), application.DataTransferArtifact{
		JobID: job.ID, RelativePath: "exports/first.csv", DisplayName: "first.csv",
		MIMEType: "text/csv", SHA256: "first",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = repo.CreateArtifact(t.Context(), application.DataTransferArtifact{
		JobID: job.ID, RelativePath: "exports/second.csv", DisplayName: "second.csv",
		MIMEType: "text/csv", SHA256: "second",
	})
	if !errors.Is(err, application.ErrDataTransferConflict) {
		t.Fatalf("second active artifact error = %v, want conflict", err)
	}
	if err := repo.MarkArtifactDeleted(t.Context(), first.ID, "2026-08-20T14:00:00Z"); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.CreateArtifact(t.Context(), application.DataTransferArtifact{
		JobID: job.ID, RelativePath: "exports/retry.csv", DisplayName: "retry.csv",
		MIMEType: "text/csv", SHA256: "retry",
	}); err != nil {
		t.Fatalf("artifact after deletion: %v", err)
	}
}

func testDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := infrastructure.OpenSQLite(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := infrastructure.Migrate(db, filepath.Join("..", "..", "migrations")); err != nil {
		t.Fatal(err)
	}
	return db
}
