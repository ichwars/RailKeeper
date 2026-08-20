package infrastructure_test

import (
	"database/sql"
	"errors"
	"path/filepath"
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
